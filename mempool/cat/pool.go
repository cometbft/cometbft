package cat

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/creachadair/taskgroup"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
)

// enforce compile-time satisfaction of the Mempool interface
var _ mempool.Mempool = (*TxPool)(nil)

var (
	ErrTxInMempool       = errors.New("tx already exists in mempool")
	ErrTxAlreadyRejected = errors.New("tx was previously rejected")
)

// TxPoolOption sets an optional parameter on the TxPool.
type TxPoolOption func(*TxPool)

// TxPool implemements the Mempool interface and allows the application to
// set priority values on transactions in the CheckTx response. When selecting
// transactions to include in a block, higher-priority transactions are chosen
// first.  When evicting transactions from the mempool for size constraints,
// lower-priority transactions are evicted first. Transactions themselves are
// unordered (A map is used). They can be broadcast in an order different from
// the order to which transactions are entered. There is no guarantee when CheckTx
// passes that a transaction has been successfully broadcast to any of its peers.
//
// A TTL can be set to remove transactions after a period of time or a number
// of heights.
//
// A cache of rejectedTxs can be set in the mempool config. Transactions that
// are rejected because of `CheckTx` or other validity checks will be instantly
// rejected if they are seen again. Committed transactions are also added to
// this cache. This serves somewhat as replay protection but applications should
// implement something more comprehensive
type TxPool struct {
	// Immutable fields
	logger       log.Logger
	config       *config.MempoolConfig
	proxyAppConn proxy.AppConnMempool
	metrics      *mempool.Metrics

	// these values are modified once per height
	updateMtx            sync.Mutex
	notifiedTxsAvailable bool
	txsAvailable         chan struct{} // one value sent per height when mempool is not empty
	preCheckFn           mempool.PreCheckFunc
	postCheckFn          mempool.PostCheckFunc
	height               int64     // the latest height passed to Update
	lastPurgeTime        time.Time // the last time we attempted to purge transactions via the TTL

	// Thread-safe cache of rejected transactions for quick look-up
	rejectedTxCache *LRUTxCache
	// Thread-safe list of transactions peers have seen that we have not yet seen
	seenByPeersSet *SeenTxSet

	// Store of wrapped transactions
	store *store

	// broadcastCh is an unbuffered channel of new transactions that need to
	// be broadcasted to peers. Only populated if `broadcast` in the config is enabled
	broadcastCh      chan *wrappedTx
	broadcastMtx     sync.Mutex
	txsToBeBroadcast []types.TxKey
}

// NewTxPool constructs a new, empty content addressable txpool at the specified
// initial height and using the given config and options.
func NewTxPool(
	logger log.Logger,
	cfg *config.MempoolConfig,
	proxyAppConn proxy.AppConnMempool,
	height int64,
	options ...TxPoolOption,
) *TxPool {
	txmp := &TxPool{
		logger:           logger,
		config:           cfg,
		proxyAppConn:     proxyAppConn,
		metrics:          mempool.NopMetrics(),
		rejectedTxCache:  NewLRUTxCache(cfg.CacheSize),
		seenByPeersSet:   NewSeenTxSet(),
		height:           height,
		preCheckFn:       func(_ types.Tx) error { return nil },
		postCheckFn:      func(_ types.Tx, _ *abci.ResponseCheckTx) error { return nil },
		store:            newStore(),
		broadcastCh:      make(chan *wrappedTx),
		txsToBeBroadcast: make([]types.TxKey, 0),
	}

	for _, opt := range options {
		opt(txmp)
	}

	return txmp
}

// WithPreCheck sets a filter for the mempool to reject a transaction if f(tx)
// returns an error. This is executed before CheckTx. It only applies to the
// first created block. After that, Update() overwrites the existing value.
func WithPreCheck(f mempool.PreCheckFunc) TxPoolOption {
	return func(txmp *TxPool) { txmp.preCheckFn = f }
}

// WithPostCheck sets a filter for the mempool to reject a transaction if
// f(tx, resp) returns an error. This is executed after CheckTx. It only applies
// to the first created block. After that, Update overwrites the existing value.
func WithPostCheck(f mempool.PostCheckFunc) TxPoolOption {
	return func(txmp *TxPool) { txmp.postCheckFn = f }
}

// WithMetrics sets the mempool's metrics collector.
func WithMetrics(metrics *mempool.Metrics) TxPoolOption {
	return func(txmp *TxPool) { txmp.metrics = metrics }
}

// Lock is a noop as ABCI calls are serialized
func (txmp *TxPool) Lock() {}

// Unlock is a noop as ABCI calls are serialized
func (txmp *TxPool) Unlock() {}

// Size returns the number of valid transactions in the mempool. It is
// thread-safe.
func (txmp *TxPool) Size() int { return txmp.store.size() }

// SizeBytes returns the total sum in bytes of all the valid transactions in the
// mempool. It is thread-safe.
func (txmp *TxPool) SizeBytes() int64 { return txmp.store.totalBytes() }

// FlushAppConn executes FlushSync on the mempool's proxyAppConn.
//
// The caller must hold an exclusive mempool lock (by calling txmp.Lock) before
// calling FlushAppConn.
func (txmp *TxPool) FlushAppConn() error {
	return txmp.proxyAppConn.FlushSync()
}

// EnableTxsAvailable enables the mempool to trigger events when transactions
// are available on a block by block basis.
func (txmp *TxPool) EnableTxsAvailable() {
	txmp.txsAvailable = make(chan struct{}, 1)
}

// TxsAvailable returns a channel which fires once for every height, and only
// when transactions are available in the mempool. It is thread-safe.
func (txmp *TxPool) TxsAvailable() <-chan struct{} { return txmp.txsAvailable }

// Height returns the latest height that the mempool is at
func (txmp *TxPool) Height() int64 {
	txmp.updateMtx.Lock()
	defer txmp.updateMtx.Unlock()
	return txmp.height
}

// Has returns true if the transaction is currently in the mempool
func (txmp *TxPool) Has(txKey types.TxKey) bool {
	return txmp.store.has(txKey)
}

// Get retrieves a transaction based on the key. It returns a bool
// if the transaction exists or not
func (txmp *TxPool) Get(txKey types.TxKey) (types.Tx, bool) {
	wtx := txmp.store.get(txKey)
	if wtx != nil {
		return wtx.tx, true
	}
	return types.Tx{}, false
}

// IsRejectedTx returns true if the transaction was recently rejected and is
// currently within the cache
func (txmp *TxPool) IsRejectedTx(txKey types.TxKey) bool {
	return txmp.rejectedTxCache.Has(txKey)
}

// CheckToPurgeExpiredTxs checks if there has been adequate time since the last time
// the txpool looped through all transactions and if so, performs a purge of any transaction
// that has expired according to the TTLDuration. This is thread safe.
func (txmp *TxPool) CheckToPurgeExpiredTxs() {
	txmp.updateMtx.Lock()
	defer txmp.updateMtx.Unlock()
	if txmp.config.TTLDuration > 0 && time.Since(txmp.lastPurgeTime) > txmp.config.TTLDuration {
		expirationAge := time.Now().Add(-txmp.config.TTLDuration)
		// a height of 0 means no transactions will be removed because of height
		// (in other words, no transaction has a height less than 0)
		numExpired := txmp.store.purgeExpiredTxs(0, expirationAge)
		txmp.metrics.EvictedTxs.Add(float64(numExpired))
		txmp.lastPurgeTime = time.Now()
	}
}

// CheckTx adds the given transaction to the mempool if it fits and passes the
// application's ABCI CheckTx method. This should be viewed as the entry method for new transactions
// into the network. In practice this happens via an RPC endpoint
func (txmp *TxPool) CheckTx(tx types.Tx, cb func(*abci.Response), txInfo mempool.TxInfo) error {
	// Reject transactions in excess of the configured maximum transaction size.
	if len(tx) > txmp.config.MaxTxBytes {
		return mempool.ErrTxTooLarge{Max: txmp.config.MaxTxBytes, Actual: len(tx)}
	}

	// This is a new transaction that we haven't seen before. Verify it against the app and attempt
	// to add it to the transaction pool.
	key := tx.Key()
	rsp, err := txmp.TryAddNewTx(tx, key, txInfo)
	if err != nil {
		return err
	}
	defer func() {
		// call the callback if it is set
		if cb != nil {
			cb(&abci.Response{Value: &abci.Response_CheckTx{CheckTx: rsp}})
		}
	}()

	// push to the broadcast queue that a new transaction is ready
	txmp.markToBeBroadcast(key)
	return nil
}

// next is used by the reactor to get the next transaction to broadcast
// to all other peers.
func (txmp *TxPool) next() <-chan *wrappedTx {
	txmp.broadcastMtx.Lock()
	defer txmp.broadcastMtx.Unlock()
	for len(txmp.txsToBeBroadcast) != 0 {
		ch := make(chan *wrappedTx, 1)
		key := txmp.txsToBeBroadcast[0]
		txmp.txsToBeBroadcast = txmp.txsToBeBroadcast[1:]
		wtx := txmp.store.get(key)
		if wtx == nil {
			continue
		}
		ch <- wtx
		return ch
	}

	return txmp.broadcastCh
}

// markToBeBroadcast marks a transaction to be broadcasted to peers.
// This should never block so we use a map to create an unbounded queue
// of transactions that need to be gossiped.
func (txmp *TxPool) markToBeBroadcast(key types.TxKey) {
	if !txmp.config.Broadcast {
		return
	}

	wtx := txmp.store.get(key)
	if wtx == nil {
		return
	}

	select {
	case txmp.broadcastCh <- wtx:
	default:
		txmp.broadcastMtx.Lock()
		defer txmp.broadcastMtx.Unlock()
		txmp.txsToBeBroadcast = append(txmp.txsToBeBroadcast, key)
	}
}

// TryAddNewTx attempts to add a tx that has not already been seen before. It first marks it as seen
// to avoid races with the same tx. It then call `CheckTx` so that the application can validate it.
// If it passes `CheckTx`, the new transaction is added to the mempool as long as it has
// sufficient priority and space else if evicted it will return an error
func (txmp *TxPool) TryAddNewTx(tx types.Tx, key types.TxKey, txInfo mempool.TxInfo) (*abci.ResponseCheckTx, error) {
	// First check any of the caches to see if we can conclude early. We may have already seen and processed
	// the transaction if:
	// - We are connected to nodes running v0 or v1 which simply flood the network
	// - If a client submits a transaction to multiple nodes (via RPC)
	// - We send multiple requests and the first peer eventually responds after the second peer has already provided the tx
	if txmp.IsRejectedTx(key) {
		// The peer has sent us a transaction that we have previously marked as invalid. Since `CheckTx` can
		// be non-deterministic, we don't punish the peer but instead just ignore the tx
		return nil, ErrTxAlreadyRejected
	}

	if txmp.Has(key) {
		txmp.metrics.AlreadySeenTxs.Add(1)
		// The peer has sent us a transaction that we have already seen
		return nil, ErrTxInMempool
	}

	// reserve the key
	if !txmp.store.reserve(key) {
		txmp.logger.Debug("mempool already attempting to verify and add transaction", "txKey", fmt.Sprintf("%X", key))
		txmp.PeerHasTx(txInfo.SenderID, key)
		return nil, ErrTxInMempool
	}
	defer txmp.store.release(key)

	// If a precheck hook is defined, call it before invoking the application.
	if err := txmp.preCheck(tx); err != nil {
		txmp.metrics.FailedTxs.Add(1)
		return nil, mempool.ErrPreCheck{Reason: err}
	}

	// Early exit if the proxy connection has an error.
	if err := txmp.proxyAppConn.Error(); err != nil {
		return nil, err
	}

	// Invoke an ABCI CheckTx for this transaction.
	rsp, err := txmp.proxyAppConn.CheckTxSync(abci.RequestCheckTx{Tx: tx})
	if err != nil {
		return rsp, err
	}
	if rsp.Code != abci.CodeTypeOK {
		if txmp.config.KeepInvalidTxsInCache {
			txmp.rejectedTxCache.Push(key)
		}
		txmp.metrics.FailedTxs.Add(1)
		return rsp, fmt.Errorf("application rejected transaction with code %d (Log: %s)", rsp.Code, rsp.Log)
	}

	// Create wrapped tx
	wtx := newWrappedTx(
		tx, key, txmp.Height(), rsp.GasWanted, rsp.Priority, rsp.Sender,
	)

	// Perform the post check
	err = txmp.postCheck(wtx.tx, rsp)
	if err != nil {
		if txmp.config.KeepInvalidTxsInCache {
			txmp.rejectedTxCache.Push(key)
		}
		txmp.metrics.FailedTxs.Add(1)
		return rsp, fmt.Errorf("rejected bad transaction after post check: %w", err)
	}

	// Now we consider the transaction to be valid. Once a transaction is valid, it
	// can only become invalid if recheckTx is enabled and RecheckTx returns a non zero code
	if err := txmp.addNewTransaction(wtx, rsp); err != nil {
		return nil, err
	}
	return rsp, nil
}

// RemoveTxByKey removes the transaction with the specified key from the
// mempool. It adds it to the rejectedTxCache so it will not be added again
func (txmp *TxPool) RemoveTxByKey(txKey types.TxKey) error {
	txmp.removeTxByKey(txKey)
	txmp.metrics.EvictedTxs.Add(1)
	return nil
}

func (txmp *TxPool) removeTxByKey(txKey types.TxKey) {
	txmp.rejectedTxCache.Push(txKey)
	_ = txmp.store.remove(txKey)
	txmp.seenByPeersSet.RemoveKey(txKey)
}

// Flush purges the contents of the mempool and the cache, leaving both empty.
// The current height is not modified by this operation.
func (txmp *TxPool) Flush() {
	// Remove all the transactions in the list explicitly, so that the sizes
	// and indexes get updated properly.
	size := txmp.Size()
	txmp.store.reset()
	txmp.seenByPeersSet.Reset()
	txmp.rejectedTxCache.Reset()
	txmp.metrics.EvictedTxs.Add(float64(size))
	txmp.broadcastMtx.Lock()
	defer txmp.broadcastMtx.Unlock()
	txmp.txsToBeBroadcast = make([]types.TxKey, 0)
}

// PeerHasTx marks that the transaction has been seen by a peer.
func (txmp *TxPool) PeerHasTx(peer uint16, txKey types.TxKey) {
	txmp.logger.Debug("peer has tx", "peer", peer, "txKey", fmt.Sprintf("%X", txKey))
	txmp.seenByPeersSet.Add(txKey, peer)
}

// allEntriesSorted returns a slice of all the transactions currently in the
// mempool, sorted in nonincreasing order by priority with ties broken by
// increasing order of arrival time.
func (txmp *TxPool) allEntriesSorted() []*wrappedTx {
	txs := txmp.store.getAllTxs()
	sort.Slice(txs, func(i, j int) bool {
		if txs[i].priority == txs[j].priority {
			return txs[i].timestamp.Before(txs[j].timestamp)
		}
		return txs[i].priority > txs[j].priority // N.B. higher priorities first
	})
	return txs
}

// ReapMaxBytesMaxGas returns a slice of valid transactions that fit within the
// size and gas constraints. The results are ordered by nonincreasing priority,
// with ties broken by increasing order of arrival. Reaping transactions does
// not remove them from the mempool
//
// If maxBytes < 0, no limit is set on the total size in bytes.
// If maxGas < 0, no limit is set on the total gas cost.
//
// If the mempool is empty or has no transactions fitting within the given
// constraints, the result will also be empty.
func (txmp *TxPool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs {
	var totalGas, totalBytes int64

	var keep []types.Tx //nolint:prealloc
	for _, w := range txmp.allEntriesSorted() {
		// N.B. When computing byte size, we need to include the overhead for
		// encoding as protobuf to send to the application.
		totalGas += w.gasWanted
		totalBytes += types.ComputeProtoSizeForTxs([]types.Tx{w.tx})
		if (maxGas >= 0 && totalGas > maxGas) || (maxBytes >= 0 && totalBytes > maxBytes) {
			break
		}
		keep = append(keep, w.tx)
	}
	return keep
}

// ReapMaxTxs returns up to max transactions from the mempool. The results are
// ordered by nonincreasing priority with ties broken by increasing order of
// arrival. Reaping transactions does not remove them from the mempool.
//
// If max < 0, all transactions in the mempool are reaped.
//
// The result may have fewer than max elements (possibly zero) if the mempool
// does not have that many transactions available.
func (txmp *TxPool) ReapMaxTxs(max int) types.Txs {
	var keep []types.Tx //nolint:prealloc

	for _, w := range txmp.allEntriesSorted() {
		if max >= 0 && len(keep) >= max {
			break
		}
		keep = append(keep, w.tx)
	}
	return keep
}

// Update removes all the given transactions from the mempool and the cache,
// and updates the current block height. The blockTxs and deliverTxResponses
// must have the same length with each response corresponding to the tx at the
// same offset.
//
// If the configuration enables recheck, Update sends each remaining
// transaction after removing blockTxs to the ABCI CheckTx method.  Any
// transactions marked as invalid during recheck are also removed.
//
// The caller must hold an exclusive mempool lock (by calling txmp.Lock) before
// calling Update.
func (txmp *TxPool) Update(
	blockHeight int64,
	blockTxs types.Txs,
	deliverTxResponses []*abci.ResponseDeliverTx,
	newPreFn mempool.PreCheckFunc,
	newPostFn mempool.PostCheckFunc,
) error {
	// Safety check: Transactions and responses must match in number.
	if len(blockTxs) != len(deliverTxResponses) {
		panic(fmt.Sprintf("mempool: got %d transactions but %d DeliverTx responses",
			len(blockTxs), len(deliverTxResponses)))
	}
	txmp.logger.Debug("updating mempool", "height", blockHeight, "txs", len(blockTxs))

	txmp.updateMtx.Lock()
	txmp.height = blockHeight
	txmp.notifiedTxsAvailable = false

	if newPreFn != nil {
		txmp.preCheckFn = newPreFn
	}
	if newPostFn != nil {
		txmp.postCheckFn = newPostFn
	}
	txmp.lastPurgeTime = time.Now()
	txmp.updateMtx.Unlock()

	txmp.metrics.SuccessfulTxs.Add(float64(len(blockTxs)))
	for _, tx := range blockTxs {
		// Regardless of success, remove the transaction from the mempool.
		txmp.removeTxByKey(tx.Key())
	}

	txmp.purgeExpiredTxs(blockHeight)

	// If there any uncommitted transactions left in the mempool, we either
	// initiate re-CheckTx per remaining transaction or notify that remaining
	// transactions are left.
	size := txmp.Size()
	txmp.metrics.Size.Set(float64(size))
	if size > 0 {
		if txmp.config.Recheck {
			txmp.recheckTransactions()
		} else {
			txmp.notifyTxsAvailable()
		}
	}
	return nil
}

// addNewTransaction handles the ABCI CheckTx response for the first time a
// transaction is added to the mempool.  A recheck after a block is committed
// goes to handleRecheckResult.
//
// If either the application rejected the transaction or a post-check hook is
// defined and rejects the transaction, it is discarded.
//
// Otherwise, if the mempool is full, check for lower-priority transactions
// that can be evicted to make room for the new one. If no such transactions
// exist, this transaction is logged and dropped; otherwise the selected
// transactions are evicted.
//
// Finally, the new transaction is added and size stats updated.
func (txmp *TxPool) addNewTransaction(wtx *wrappedTx, checkTxRes *abci.ResponseCheckTx) error {
	// At this point the application has ruled the transaction valid, but the
	// mempool might be full. If so, find the lowest-priority items with lower
	// priority than the application assigned to this new one, and evict as many
	// of them as necessary to make room for tx. If no such items exist, we
	// discard tx.
	if !txmp.canAddTx(wtx.size()) {
		victims, victimBytes := txmp.store.getTxsBelowPriority(wtx.priority)

		// If there are no suitable eviction candidates, or the total size of
		// those candidates is not enough to make room for the new transaction,
		// drop the new one.
		if len(victims) == 0 || victimBytes < wtx.size() {
			txmp.metrics.EvictedTxs.Add(1)
			checkTxRes.MempoolError = fmt.Sprintf("rejected valid incoming transaction; mempool is full (%X)",
				wtx.key)
			return fmt.Errorf("rejected valid incoming transaction; mempool is full (%X). Size: (%d:%d)",
				wtx.key.String(), txmp.Size(), txmp.SizeBytes())
		}

		txmp.logger.Debug("evicting lower-priority transactions",
			"new_tx", wtx.key.String(),
			"new_priority", wtx.priority,
		)

		// Sort lowest priority items first so they will be evicted first.  Break
		// ties in favor of newer items (to maintain FIFO semantics in a group).
		sort.Slice(victims, func(i, j int) bool {
			iw := victims[i]
			jw := victims[j]
			if iw.priority == jw.priority {
				return iw.timestamp.After(jw.timestamp)
			}
			return iw.priority < jw.priority
		})

		// Evict as many of the victims as necessary to make room.
		availableBytes := txmp.availableBytes()
		for _, tx := range victims {
			txmp.evictTx(tx)

			// We may not need to evict all the eligible transactions.  Bail out
			// early if we have made enough room.
			availableBytes += tx.size()
			if availableBytes >= wtx.size() {
				break
			}
		}
	}

	txmp.store.set(wtx)

	txmp.metrics.TxSizeBytes.Observe(float64(wtx.size()))
	txmp.metrics.Size.Set(float64(txmp.Size()))
	txmp.logger.Debug(
		"inserted new valid transaction",
		"priority", wtx.priority,
		"tx", fmt.Sprintf("%X", wtx.key),
		"height", wtx.height,
		"num_txs", txmp.Size(),
	)
	txmp.notifyTxsAvailable()
	return nil
}

func (txmp *TxPool) evictTx(wtx *wrappedTx) {
	txmp.store.remove(wtx.key)
	txmp.metrics.EvictedTxs.Add(1)
	txmp.logger.Debug(
		"evicted valid existing transaction; mempool full",
		"old_tx", fmt.Sprintf("%X", wtx.key),
		"old_priority", wtx.priority,
	)
}

// handleRecheckResult handles the responses from ABCI CheckTx calls issued
// during the recheck phase of a block Update.  It removes any transactions
// invalidated by the application.
//
// This method is NOT executed for the initial CheckTx on a new transaction;
// that case is handled by addNewTransaction instead.
func (txmp *TxPool) handleRecheckResult(wtx *wrappedTx, checkTxRes *abci.ResponseCheckTx) {
	txmp.metrics.RecheckTimes.Add(1)

	// If a postcheck hook is defined, call it before checking the result.
	err := txmp.postCheck(wtx.tx, checkTxRes)

	if checkTxRes.Code == abci.CodeTypeOK && err == nil {
		// Note that we do not update the transaction with any of the values returned in
		// recheck tx
		return // N.B. Size of mempool did not change
	}

	txmp.logger.Debug(
		"existing transaction no longer valid; failed re-CheckTx callback",
		"priority", wtx.priority,
		"tx", fmt.Sprintf("%X", wtx.key),
		"err", err,
		"code", checkTxRes.Code,
	)
	txmp.store.remove(wtx.key)
	if txmp.config.KeepInvalidTxsInCache {
		txmp.rejectedTxCache.Push(wtx.key)
	}
	txmp.metrics.FailedTxs.Add(1)
	txmp.metrics.Size.Set(float64(txmp.Size()))
}

// recheckTransactions initiates re-CheckTx ABCI calls for all the transactions
// currently in the mempool. It reports the number of recheck calls that were
// successfully initiated.
//
// Precondition: The mempool is not empty.
// The caller must hold txmp.mtx exclusively.
func (txmp *TxPool) recheckTransactions() {
	if txmp.Size() == 0 {
		panic("mempool: cannot run recheck on an empty mempool")
	}
	txmp.logger.Debug(
		"executing re-CheckTx for all remaining transactions",
		"num_txs", txmp.Size(),
		"height", txmp.Height(),
	)

	// Collect transactions currently in the mempool requiring recheck.
	wtxs := txmp.store.getAllTxs()

	// Issue CheckTx calls for each remaining transaction, and when all the
	// rechecks are complete signal watchers that transactions may be available.
	go func() {
		g, start := taskgroup.New(nil).Limit(2 * runtime.NumCPU())

		for _, wtx := range wtxs {
			wtx := wtx
			start(func() error {
				// The response for this CheckTx is handled by the default recheckTxCallback.
				rsp, err := txmp.proxyAppConn.CheckTxSync(abci.RequestCheckTx{
					Tx:   wtx.tx,
					Type: abci.CheckTxType_Recheck,
				})
				if err != nil {
					txmp.logger.Error("failed to execute CheckTx during recheck",
						"err", err, "key", fmt.Sprintf("%x", wtx.key))
				} else {
					txmp.handleRecheckResult(wtx, rsp)
				}
				return nil
			})
		}
		_ = txmp.proxyAppConn.FlushAsync()

		// When recheck is complete, trigger a notification for more transactions.
		_ = g.Wait()
		txmp.notifyTxsAvailable()
	}()
}

// availableBytes returns the number of bytes available in the mempool.
func (txmp *TxPool) availableBytes() int64 {
	return txmp.config.MaxTxsBytes - txmp.SizeBytes()
}

// canAddTx returns an error if we cannot insert the provided *wrappedTx into
// the mempool due to mempool configured constraints. Otherwise, nil is
// returned and the transaction can be inserted into the mempool.
func (txmp *TxPool) canAddTx(size int64) bool {
	numTxs := txmp.Size()
	txBytes := txmp.SizeBytes()

	if numTxs > txmp.config.Size || size+txBytes > txmp.config.MaxTxsBytes {
		return false
	}

	return true
}

// purgeExpiredTxs removes all transactions from the mempool that have exceeded
// their respective height or time-based limits as of the given blockHeight.
// Transactions removed by this operation are not removed from the rejectedTxCache.
func (txmp *TxPool) purgeExpiredTxs(blockHeight int64) {
	if txmp.config.TTLNumBlocks == 0 && txmp.config.TTLDuration == 0 {
		return // nothing to do
	}

	expirationHeight := blockHeight - txmp.config.TTLNumBlocks
	if txmp.config.TTLNumBlocks == 0 {
		expirationHeight = 0
	}

	now := time.Now()
	expirationAge := now.Add(-txmp.config.TTLDuration)
	if txmp.config.TTLDuration == 0 {
		expirationAge = time.Time{}
	}

	numExpired := txmp.store.purgeExpiredTxs(expirationHeight, expirationAge)
	txmp.metrics.EvictedTxs.Add(float64(numExpired))

	// purge old evicted and seen transactions
	if txmp.config.TTLDuration == 0 {
		// ensure that seenByPeersSet are eventually pruned
		expirationAge = now.Add(-time.Hour)
	}
	txmp.seenByPeersSet.Prune(expirationAge)
}

func (txmp *TxPool) notifyTxsAvailable() {
	if txmp.Size() == 0 {
		return // nothing to do
	}

	if txmp.txsAvailable != nil && !txmp.notifiedTxsAvailable {
		// channel cap is 1, so this will send once
		txmp.notifiedTxsAvailable = true

		select {
		case txmp.txsAvailable <- struct{}{}:
		default:
		}
	}
}

func (txmp *TxPool) preCheck(tx types.Tx) error {
	txmp.updateMtx.Lock()
	defer txmp.updateMtx.Unlock()
	if txmp.preCheckFn != nil {
		return txmp.preCheckFn(tx)
	}
	return nil
}

func (txmp *TxPool) postCheck(tx types.Tx, res *abci.ResponseCheckTx) error {
	txmp.updateMtx.Lock()
	defer txmp.updateMtx.Unlock()
	if txmp.postCheckFn != nil {
		return txmp.postCheckFn(tx, res)
	}
	return nil
}
