package mempool

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/clist"
	"github.com/cometbft/cometbft/libs/log"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

const noSender = p2p.ID("")

// CListMempool is an ordered in-memory pool for transactions before they are
// proposed in a consensus round. Transaction validity is checked using the
// CheckTx abci message before the transaction is added to the pool. The
// mempool uses a concurrent list structure for storing transactions that can
// be efficiently accessed by multiple concurrent readers.
type CListMempool struct {
	height atomic.Int64 // the last block Update()'d to

	// notify listeners (ie. consensus) when txs are available
	notifiedTxsAvailable atomic.Bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty
	onNewTx              func(types.Tx)

	config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of
	// CheckTx or ReapMaxBytesMaxGas(ReapMaxTxs) methods.
	updateMtx cmtsync.RWMutex
	preCheck  PreCheckFunc
	postCheck PostCheckFunc

	proxyAppConn proxy.AppConnMempool

	// Keeps track of the rechecking process.
	recheck *recheck

	// Data in the following variables must to be kept in sync and updated atomically.
	txsMtx   cmtsync.RWMutex
	lanes    map[types.Lane]*clist.CList     // each lane is a linked-list of (valid) txs
	txsMap   map[types.TxKey]*clist.CElement // for quick access to the mempool entry of a given tx
	txsBytes int64                           // total size of mempool, in bytes
	numTxs   int64                           // total number of txs in the mempool

	addTxChMtx    cmtsync.RWMutex // Protects the fields below
	addTxCh       chan struct{}   // Blocks until the next TX is added
	addTxSeq      int64
	addTxLaneSeqs map[types.Lane]int64

	// Immutable fields, only set during initialization.
	defaultLane types.Lane
	sortedLanes []types.Lane // lanes sorted by priority

	// Keep a cache of already-seen txs.
	// This reduces the pressure on the proxyApp.
	cache TxCache

	logger  log.Logger
	metrics *Metrics
}

var _ Mempool = &CListMempool{}

// CListMempoolOption sets an optional parameter on the mempool.
type CListMempoolOption func(*CListMempool)

// NewCListMempool returns a new mempool with the given configuration and
// connection to an application.
func NewCListMempool(
	cfg *config.MempoolConfig,
	proxyAppConn proxy.AppConnMempool,
	lanesInfo *LanesInfo,
	height int64,
	options ...CListMempoolOption,
) *CListMempool {
	mp := &CListMempool{
		config:        cfg,
		proxyAppConn:  proxyAppConn,
		txsMap:        make(map[types.TxKey]*clist.CElement),
		recheck:       &recheck{},
		logger:        log.NewNopLogger(),
		metrics:       NopMetrics(),
		addTxCh:       make(chan struct{}),
		addTxLaneSeqs: make(map[types.Lane]int64),
	}
	mp.height.Store(height)

	// Initialize lanes
	if lanesInfo == nil || len(lanesInfo.lanes) == 0 {
		// Lane 1 will be the only lane.
		mp.lanes = make(map[types.Lane]*clist.CList, 1)
		mp.defaultLane = types.Lane(1)
		mp.lanes[mp.defaultLane] = clist.New()
		mp.sortedLanes = []types.Lane{mp.defaultLane}
	} else {
		numLanes := len(lanesInfo.lanes)
		mp.lanes = make(map[types.Lane]*clist.CList, numLanes)
		mp.defaultLane = lanesInfo.defaultLane
		mp.sortedLanes = make([]types.Lane, numLanes)
		for i, lane := range lanesInfo.lanes {
			mp.lanes[lane] = clist.New()
			mp.sortedLanes[i] = lane
		}
		slices.Sort(mp.sortedLanes)
		slices.Reverse(mp.sortedLanes)
	}

	if cfg.CacheSize > 0 {
		mp.cache = NewLRUTxCache(cfg.CacheSize)
	} else {
		mp.cache = NopTxCache{}
	}

	for _, option := range options {
		option(mp)
	}

	mp.logger.Info("CListMempool created", "defaultLane", mp.defaultLane, "lanes", mp.sortedLanes)
	return mp
}

func (mem *CListMempool) addToCache(tx types.Tx) bool {
	return mem.cache.Push(tx)
}

func (mem *CListMempool) forceRemoveFromCache(tx types.Tx) {
	mem.cache.Remove(tx)
}

// tryRemoveFromCache removes a transaction from the cache in case it can be
// added to the mempool at a later stage (probably when the transaction becomes
// valid).
func (mem *CListMempool) tryRemoveFromCache(tx types.Tx) {
	if !mem.config.KeepInvalidTxsInCache {
		mem.forceRemoveFromCache(tx)
	}
}

func (mem *CListMempool) removeAllTxs(lane types.Lane) {
	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	for e := mem.lanes[lane].Front(); e != nil; e = e.Next() {
		mem.lanes[lane].Remove(e)
		e.DetachPrev()
	}
	mem.txsMap = make(map[types.TxKey]*clist.CElement)
	mem.txsBytes = 0
}

// addSender adds a peer ID to the list of senders on the entry corresponding to
// tx, identified by its key.
func (mem *CListMempool) addSender(txKey types.TxKey, sender p2p.ID) error {
	if sender == noSender {
		return nil
	}

	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	elem, ok := mem.txsMap[txKey]
	if !ok {
		return ErrTxNotFound
	}

	memTx := elem.Value.(*mempoolTx)
	if found := memTx.addSender(sender); found {
		// It should not be possible to receive twice a tx from the same sender.
		return ErrTxAlreadyReceivedFromSender
	}
	return nil
}

// NOTE: not thread safe - should only be called once, on startup.
func (mem *CListMempool) EnableTxsAvailable() {
	mem.txsAvailable = make(chan struct{}, 1)
}

// SetLogger sets the Logger.
func (mem *CListMempool) SetLogger(l log.Logger) {
	mem.logger = l
}

// WithPreCheck sets a filter for the mempool to reject a tx if f(tx) returns
// false. This is ran before CheckTx. Only applies to the first created block.
// After that, Update overwrites the existing value.
func WithPreCheck(f PreCheckFunc) CListMempoolOption {
	return func(mem *CListMempool) { mem.preCheck = f }
}

// WithPostCheck sets a filter for the mempool to reject a tx if f(tx) returns
// false. This is ran after CheckTx. Only applies to the first created block.
// After that, Update overwrites the existing value.
func WithPostCheck(f PostCheckFunc) CListMempoolOption {
	return func(mem *CListMempool) { mem.postCheck = f }
}

// WithMetrics sets the metrics.
func WithMetrics(metrics *Metrics) CListMempoolOption {
	return func(mem *CListMempool) { mem.metrics = metrics }
}

// WithNewTxCallback sets a callback function to be executed when a new transaction is added to the mempool.
// The callback function will receive the newly added transaction as a parameter.
func WithNewTxCallback(cb func(types.Tx)) CListMempoolOption {
	return func(mem *CListMempool) { mem.onNewTx = cb }
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Lock() {
	mem.updateMtx.Lock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Unlock() {
	mem.updateMtx.Unlock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) PreUpdate() {
	if mem.recheck.setRecheckFull() {
		mem.logger.Debug("The state of recheckFull has flipped")
	}
}

// Size returns the total number of transactions in the mempool (that is, all lanes).
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Size() int {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	return int(mem.numTxs)
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) SizeBytes() int64 {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	return mem.txsBytes
}

// Lock() must be help by the caller during execution.
func (mem *CListMempool) FlushAppConn() error {
	err := mem.proxyAppConn.Flush(context.TODO())
	if err != nil {
		return ErrFlushAppConn{Err: err}
	}

	return nil
}

// XXX: Unsafe! Calling Flush may leave mempool in inconsistent state.
func (mem *CListMempool) Flush() {
	mem.updateMtx.Lock()
	defer mem.updateMtx.Unlock()

	mem.txsBytes = 0
	mem.numTxs = 0
	mem.cache.Reset()

	for lane := range mem.lanes {
		mem.removeAllTxs(lane)
	}
}

func (mem *CListMempool) Contains(txKey types.TxKey) bool {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	_, ok := mem.txsMap[txKey]
	return ok
}

// It blocks if we're waiting on Update() or Reap().
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) CheckTx(tx types.Tx, sender p2p.ID) (*abcicli.ReqRes, error) {
	mem.updateMtx.RLock()
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.RUnlock()

	txSize := len(tx)

	if err := mem.isFull(txSize); err != nil {
		return nil, err
	}

	if txSize > mem.config.MaxTxBytes {
		return nil, ErrTxTooLarge{
			Max:    mem.config.MaxTxBytes,
			Actual: txSize,
		}
	}

	if mem.preCheck != nil {
		if err := mem.preCheck(tx); err != nil {
			return nil, ErrPreCheck{Err: err}
		}
	}

	// NOTE: proxyAppConn may error if tx buffer is full
	if err := mem.proxyAppConn.Error(); err != nil {
		return nil, ErrAppConnMempool{Err: err}
	}

	if added := mem.addToCache(tx); !added {
		mem.metrics.AlreadyReceivedTxs.Add(1)
		// Record a new sender for a tx we've already seen.
		// Note it's possible a tx is still in the cache but no longer in the mempool
		// (eg. after committing a block, txs are removed from mempool but not cache),
		// so we only record the sender for txs still in the mempool.
		if err := mem.addSender(tx.Key(), sender); err != nil {
			mem.logger.Error("Could not add sender to tx", "tx", tx.Hash(), "sender", sender, "err", err)
		}
		// TODO: consider punishing peer for dups,
		// its non-trivial since invalid txs can become valid,
		// but they can spam the same tx with little cost to them atm.
		return nil, ErrTxInCache
	}

	reqRes, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.CheckTxRequest{
		Tx:   tx,
		Type: abci.CHECK_TX_TYPE_CHECK,
	})
	if err != nil {
		panic(fmt.Errorf("CheckTx request for tx %s failed: %w", log.NewLazySprintf("%X", tx.Hash()), err))
	}
	reqRes.SetCallback(mem.handleCheckTxResponse(tx, sender))

	return reqRes, nil
}

// handleCheckTxResponse handles CheckTx responses for transactions validated for the first time.
//
//   - sender optionally holds the ID of the peer that sent the transaction, if any.
func (mem *CListMempool) handleCheckTxResponse(tx types.Tx, sender p2p.ID) func(res *abci.Response) {
	return func(r *abci.Response) {
		res := r.GetCheckTx()
		if res == nil {
			panic(fmt.Sprintf("unexpected response value %v not of type CheckTx", r))
		}

		// Check that rechecking txs is not in process.
		if !mem.recheck.done() {
			panic(log.NewLazySprintf("rechecking has not finished; cannot check new tx %X", tx.Hash()))
		}

		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx, res)
		}

		// If tx is invalid, remove it from the cache.
		if res.Code != abci.CodeTypeOK || postCheckErr != nil {
			mem.tryRemoveFromCache(tx)
			mem.logger.Debug(
				"Rejected invalid transaction",
				"tx", tx.Hash(),
				"res", res,
				"err", postCheckErr,
			)
			mem.metrics.FailedTxs.Add(1)
			return
		}

		// Check again that mempool isn't full, to reduce the chance of exceeding the limits.
		if err := mem.isFull(len(tx)); err != nil {
			mem.forceRemoveFromCache(tx) // mempool might have space later
			mem.logger.Error(err.Error())
			return
		}

		// If the app returned a (non-zero) lane, use it; otherwise use the default lane.
		lane := mem.defaultLane
		if l := types.Lane(res.Lane); l != 0 {
			lane = l
		}

		// Check that tx is not already in the mempool. This can happen when the
		// cache overflows. See https://github.com/cometbft/cometbft/pull/890.
		txKey := tx.Key()
		if mem.Contains(txKey) {
			if err := mem.addSender(txKey, sender); err != nil {
				mem.logger.Error("Could not add sender to tx", "tx", tx.Hash(), "sender", sender, "err", err)
			}
			mem.logger.Debug(
				"Transaction already in mempool, not adding it again",
				"tx", tx.Hash(),
				"lane", lane,
				"height", mem.height.Load(),
				"total", mem.Size(),
			)
			return
		}

		// Add tx to mempool and notify that new txs are available.
		memTx := mempoolTx{
			height:    mem.height.Load(),
			gasWanted: res.GasWanted,
			tx:        tx,
		}
		if mem.addTx(&memTx, sender, lane) {
			mem.notifyTxsAvailable()

			if mem.onNewTx != nil {
				mem.onNewTx(tx)
			}

			// update metrics
			mem.metrics.Size.Set(float64(mem.Size()))
			mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))
		}
	}
}

// Called from:
//   - handleCheckTxResponse (lock not held) if tx is valid
func (mem *CListMempool) addTx(memTx *mempoolTx, sender p2p.ID, lane types.Lane) bool {
	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	tx := memTx.tx
	txKey := tx.Key()

	// Get lane's clist.
	txs, ok := mem.lanes[lane]
	if !ok {
		mem.logger.Error("Lane does not exist, not adding TX", "tx", log.NewLazySprintf("%v", tx.Hash()), "lane", lane)
		return false
	}

	mem.addTxChMtx.Lock()
	defer mem.addTxChMtx.Unlock()
	mem.addTxSeq++
	memTx.seq = mem.addTxSeq

	// Add new transaction.
	_ = memTx.addSender(sender)
	memTx.lane = lane
	e := txs.PushBack(memTx)
	mem.addTxLaneSeqs[lane] = mem.addTxSeq

	// Update auxiliary variables.
	mem.txsMap[txKey] = e

	// Update size variables.
	mem.txsBytes += int64(len(tx))
	mem.numTxs++

	close(mem.addTxCh)
	mem.addTxCh = make(chan struct{})

	// Update metrics.
	mem.metrics.TxSizeBytes.Observe(float64(len(tx)))

	mem.logger.Debug(
		"Added transaction",
		"tx", tx.Hash(),
		"lane", lane,
		"lane size", mem.lanes[lane].Len(),
		"height", mem.height.Load(),
		"total", mem.numTxs,
	)

	return true
}

// RemoveTxByKey removes a transaction from the mempool by its TxKey index.
// Called from:
//   - Update (updateMtx held) if tx was committed
//   - handleRecheckTxResponse (updateMtx not held) if tx was invalidated
func (mem *CListMempool) RemoveTxByKey(txKey types.TxKey) error {
	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	elem, ok := mem.txsMap[txKey]
	if !ok {
		return ErrTxNotFound
	}

	memTx := elem.Value.(*mempoolTx)

	// Remove tx from lane.
	mem.lanes[memTx.lane].Remove(elem)
	elem.DetachPrev()

	// Update auxiliary variables.
	delete(mem.txsMap, txKey)

	// Update size variables.
	mem.txsBytes -= int64(len(memTx.tx))
	mem.numTxs--

	mem.logger.Debug(
		"Removed transaction",
		"tx", memTx.tx.Hash(),
		"lane", memTx.lane,
		"lane size", mem.lanes[memTx.lane].Len(),
		"height", mem.height.Load(),
		"total", mem.numTxs,
	)
	return nil
}

func (mem *CListMempool) isFull(txSize int) error {
	memSize := mem.Size()
	txsBytes := mem.SizeBytes()
	if memSize >= mem.config.Size || uint64(txSize)+uint64(txsBytes) > uint64(mem.config.MaxTxsBytes) {
		return ErrMempoolIsFull{
			NumTxs:      memSize,
			MaxTxs:      mem.config.Size,
			TxsBytes:    txsBytes,
			MaxTxsBytes: mem.config.MaxTxsBytes,
		}
	}

	if mem.recheck.consideredFull() {
		return ErrRecheckFull
	}

	return nil
}

// handleRecheckTxResponse handles CheckTx responses for transactions in the mempool that need to be
// revalidated after a mempool update.
func (mem *CListMempool) handleRecheckTxResponse(tx types.Tx) func(res *abci.Response) {
	return func(r *abci.Response) {
		res := r.GetCheckTx()
		if res == nil {
			panic(fmt.Sprintf("unexpected response value %v not of type CheckTx", r))
		}

		// Check whether the rechecking process has finished.
		if mem.recheck.done() {
			mem.logger.Error("Rechecking has finished; discard late recheck response",
				"tx", log.NewLazySprintf("%X", tx.Hash()))
			return
		}
		mem.metrics.RecheckTimes.Add(1)

		// Check whether tx is still in the list of transactions that can be rechecked.
		if !mem.recheck.findNextEntryMatching(&tx) {
			// Reached the end of the list and didn't find a matching tx; rechecking has finished.
			return
		}

		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx, res)
		}

		// If tx is invalid, remove it from the mempool and the cache.
		if (res.Code != abci.CodeTypeOK) || postCheckErr != nil {
			// Tx became invalidated due to newly committed block.
			mem.logger.Debug("Tx is no longer valid", "tx", tx.Hash(), "res", res, "postCheckErr", postCheckErr)
			if err := mem.RemoveTxByKey(tx.Key()); err != nil {
				mem.logger.Debug("Transaction could not be removed from mempool", "err", err)
			} else {
				// update metrics
				mem.metrics.Size.Set(float64(mem.Size()))
				mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))
			}
			mem.tryRemoveFromCache(tx)
		}
	}
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsAvailable() <-chan struct{} {
	return mem.txsAvailable
}

func (mem *CListMempool) notifyTxsAvailable() {
	if mem.Size() == 0 {
		panic("notified txs available but mempool is empty!")
	}
	if mem.txsAvailable != nil && mem.notifiedTxsAvailable.CompareAndSwap(false, true) {
		// channel cap is 1, so this will send once
		select {
		case mem.txsAvailable <- struct{}{}:
		default:
		}
	}
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	var (
		totalGas    int64
		runningSize int64
	)

	// TODO: we will get a performance boost if we have a good estimate of avg
	// size per tx, and set the initial capacity based off of that.
	// txs := make([]types.Tx, 0, cmtmath.MinInt(mem.Size(), max/mem.avgTxSize))
	txs := make([]types.Tx, 0, mem.Size())
	for _, lane := range mem.sortedLanes {
		for e := mem.lanes[lane].Front(); e != nil; e = e.Next() {
			memTx := e.Value.(*mempoolTx)

			txs = append(txs, memTx.tx)

			dataSize := types.ComputeProtoSizeForTxs([]types.Tx{memTx.tx})

			// Check total size requirement
			if maxBytes > -1 && runningSize+dataSize > maxBytes {
				return txs[:len(txs)-1]
			}

			runningSize += dataSize

			// Check total gas requirement.
			// If maxGas is negative, skip this check.
			// Since newTotalGas < masGas, which
			// must be non-negative, it follows that this won't overflow.
			newTotalGas := totalGas + memTx.gasWanted
			if maxGas > -1 && newTotalGas > maxGas {
				return txs[:len(txs)-1]
			}
			totalGas = newTotalGas
		}
	}
	return txs
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) ReapMaxTxs(max int) types.Txs {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	if max < 0 {
		max = mem.Size()
	}

	txs := make([]types.Tx, 0, cmtmath.MinInt(mem.Size(), max))
	for _, lane := range mem.sortedLanes {
		for e := mem.lanes[lane].Front(); e != nil && len(txs) <= max; e = e.Next() {
			memTx := e.Value.(*mempoolTx)
			txs = append(txs, memTx.tx)
		}
	}
	return txs
}

// GetTxByHash returns the types.Tx with the given hash if found in the mempool, otherwise returns nil.
func (mem *CListMempool) GetTxByHash(hash []byte) types.Tx {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	if elem, ok := mem.txsMap[types.TxKey(hash)]; ok {
		return elem.Value.(*mempoolTx).tx
	}
	return nil
}

// Lock() must be help by the caller during execution.
// TODO: this function always returns nil; remove the return value.
func (mem *CListMempool) Update(
	height int64,
	txs types.Txs,
	txResults []*abci.ExecTxResult,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) error {
	mem.logger.Debug("Update", "height", height, "len(txs)", len(txs))

	// Set height
	mem.height.Store(height)
	mem.notifiedTxsAvailable.Store(false)

	if preCheck != nil {
		mem.preCheck = preCheck
	}
	if postCheck != nil {
		mem.postCheck = postCheck
	}

	for i, tx := range txs {
		if txResults[i].Code == abci.CodeTypeOK {
			// Add valid committed tx to the cache (if missing).
			_ = mem.addToCache(tx)
		} else {
			mem.tryRemoveFromCache(tx)
		}

		// Remove committed tx from the mempool.
		//
		// Note an evil proposer can drop valid txs!
		// Mempool before:
		//   100 -> 101 -> 102
		// Block, proposed by an evil proposer:
		//   101 -> 102
		// Mempool after:
		//   100
		// https://github.com/tendermint/tendermint/issues/3322.
		if err := mem.RemoveTxByKey(tx.Key()); err != nil {
			mem.logger.Debug("Committed transaction not in local mempool (not an error)",
				"tx", tx.Hash(),
				"error", err.Error())
		}
	}

	// Recheck txs left in the mempool to remove them if they became invalid in the new state.
	if mem.config.Recheck {
		mem.recheckTxs()
	}

	// Notify if there are still txs left in the mempool.
	if mem.Size() > 0 {
		mem.notifyTxsAvailable()
	}

	// Update metrics
	mem.metrics.Size.Set(float64(mem.Size()))
	mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))

	return nil
}

// recheckTxs sends all transactions in the mempool to the app for re-validation. When the function
// returns, all recheck responses from the app have been processed.
func (mem *CListMempool) recheckTxs() {
	mem.logger.Debug("Recheck txs", "height", mem.height.Load(), "num-txs", mem.Size())

	if mem.Size() <= 0 {
		return
	}

	var lastElement *clist.CElement
	var firstElement *clist.CElement
	for _, lane := range mem.sortedLanes {
		if mem.lanes[lane].Len() != 0 {
			lastElement = mem.lanes[lane].Back()
		}
		if firstElement == nil {
			firstElement = mem.lanes[lane].Front()
		}
	}
	mem.recheck.init(firstElement, lastElement)
	// Recheck all transactions in each lane, sequentially.
	// TODO: parallelize rechecking on lanes?
LANE_FOR_LOOP:
	for _, lane := range mem.sortedLanes {
		mem.logger.Debug("recheck lane", "height", mem.height.Load(), "lane", lane, "num-txs", mem.lanes[lane].Len())

		// NOTE: handleCheckTxResponse may be called concurrently, but CheckTx cannot be executed concurrently
		// because this function has the lock (via Update and Lock).
		for e := mem.lanes[lane].Front(); e != nil; e = e.Next() {
			tx := e.Value.(*mempoolTx).tx
			mem.recheck.numPendingTxs.Add(1)

			// Send CheckTx request to the app to re-validate transaction.
			resReq, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.CheckTxRequest{
				Tx:   tx,
				Type: abci.CHECK_TX_TYPE_RECHECK,
			})
			if err != nil {
				panic(fmt.Errorf("(re-)CheckTx request for tx %s failed: %w", log.NewLazySprintf("%v", tx.Hash()), err))
			}
			resReq.SetCallback(mem.handleRecheckTxResponse(tx))
		}

		// Flush any pending asynchronous recheck requests to process.
		mem.proxyAppConn.Flush(context.TODO())
		numlanes := len(mem.sortedLanes)
		select {
		case <-time.After(mem.config.RecheckTimeout / time.Duration(numlanes+1)):
			// mem.recheck.setDone()
			continue LANE_FOR_LOOP
			// mem.logger.Error("timed out waiting for recheck responses")
		case <-mem.recheck.doneRechecking():
		}

		if n := mem.recheck.numPendingTxs.Load(); n > 0 {
			mem.logger.Error("not all txs were rechecked", "not-rechecked", n)
		}
		mem.logger.Debug("done rechecking lane", "height", mem.height.Load(), "lane", lane)
	}
	// Give some time to finish processing the responses; then finish the rechecking process, even
	// if not all txs were rechecked.
	numlanes := len(mem.sortedLanes)
	select {
	case <-time.After(mem.config.RecheckTimeout / time.Duration(numlanes+1)):
		mem.recheck.setDone()
		mem.logger.Error("Timed out waiting for recheck responses")
	case <-mem.recheck.doneRechecking():
	}

	if n := mem.recheck.numPendingTxs.Load(); n > 0 {
		mem.logger.Error("Not all txs were rechecked", "not-rechecked", n)
	}

	mem.logger.Debug("Done rechecking", "height", mem.height.Load(), "num-txs", mem.Size())
}

// The cursor and end pointers define a dynamic list of transactions that could be rechecked. The
// end pointer is fixed. When a recheck response for a transaction is received, cursor will point to
// the entry in the mempool corresponding to that transaction, thus narrowing the list. Transactions
// corresponding to entries between the old and current positions of cursor will be ignored for
// rechecking. This is to guarantee that recheck responses are processed in the same sequential
// order as they appear in the mempool.
type recheck struct {
	cursor        *clist.CElement // next expected recheck response
	end           *clist.CElement // last entry in the mempool to recheck
	doneCh        chan struct{}   // to signal that rechecking has finished successfully (for async app connections)
	numPendingTxs atomic.Int32    // number of transactions still pending to recheck
	isRechecking  atomic.Bool     // true iff the rechecking process has begun and is not yet finished
	recheckFull   atomic.Bool     // whether rechecking TXs cannot be completed before a new block is decided
}

func (rc *recheck) init(first, last *clist.CElement) {
	if !rc.done() {
		panic("Having more than one rechecking process at a time is not possible.")
	}
	rc.cursor = first
	rc.end = last
	rc.doneCh = make(chan struct{})
	rc.numPendingTxs.Store(0)
	rc.isRechecking.Store(true)
	// rc.recheckFull.Store(false)

	// rc.tryFinish()
}

// done returns true when there is no recheck response to process.
// Safe for concurrent use by multiple goroutines.
func (rc *recheck) done() bool {
	return !rc.isRechecking.Load()
}

// setDone registers that rechecking has finished.
func (rc *recheck) setDone() {
	rc.cursor = nil
	rc.recheckFull.Store(false)
	rc.isRechecking.Store(false)
}

// tryFinish will check if the cursor is at the end of the list and notify the channel that
// rechecking has finished. It returns true iff it's done rechecking.
func (rc *recheck) tryFinish() bool {
	if rc.cursor == nil || rc.cursor == rc.end {
		// Reached end of the list without finding a matching tx.
		rc.setDone()
	}
	if rc.done() {
		close(rc.doneCh) // notify channel that recheck has finished
		return true
	}
	return false
}

// findNextEntryMatching searches for the next transaction matching the given transaction, which
// corresponds to the recheck response to be processed next. Then it checks if it has reached the
// end of the list, so it can set recheck as finished.
//
// The goal is to guarantee that transactions are rechecked in the order in which they are in the
// mempool. Transactions whose recheck response arrive late or don't arrive at all are skipped and
// not rechecked.
func (rc *recheck) findNextEntryMatching(tx *types.Tx) bool {
	found := false
	for ; !rc.done() && rc.cursor != nil; rc.cursor = rc.cursor.Next() { // when cursor is the last one, Next returns nil
		expectedTx := rc.cursor.Value.(*mempoolTx).tx
		if bytes.Equal(*tx, expectedTx) {
			// Found an entry in the list of txs to recheck that matches tx.
			found = true
			rc.numPendingTxs.Add(-1)
			break
		}
	}

	if !rc.tryFinish() {
		// Not finished yet; set the cursor for processing the next recheck response.
		rc.cursor = rc.cursor.Next()
	}
	return found
}

// doneRechecking returns the channel used to signal that rechecking has finished.
func (rc *recheck) doneRechecking() <-chan struct{} {
	return rc.doneCh
}

// setRecheckFull sets recheckFull to true if rechecking is still in progress. It returns true iff
// the value of recheckFull has changed.
func (rc *recheck) setRecheckFull() bool {
	rechecking := !rc.done()
	recheckFull := rc.recheckFull.Swap(rechecking)
	return rechecking != recheckFull
}

// consideredFull returns true iff the mempool should be considered as full while rechecking is in
// progress.
func (rc *recheck) consideredFull() bool {
	return rc.recheckFull.Load()
}

// CListIterator implements an Iterator that traverses the lanes with the classical Weighted Round
// Robin (WRR) algorithm.
type CListIterator struct {
	mp               *CListMempool                  // to wait on and retrieve the first entry
	currentLaneIndex int                            // current lane being iterated; index on mp.sortedLanes
	counters         map[types.Lane]uint32          // counters of accessed entries for WRR algorithm
	cursors          map[types.Lane]*clist.CElement // last accessed entries on each lane
}

func (mem *CListMempool) NewIterator() Iterator {
	return &CListIterator{
		mp:       mem,
		counters: make(map[types.Lane]uint32, len(mem.sortedLanes)),
		cursors:  make(map[types.Lane]*clist.CElement, len(mem.sortedLanes)),
	}
}

// WaitNextCh returns a channel to wait for the next available entry. The channel will be explicitly
// closed when the entry gets removed before it is added to the channel, or when reaching the end of
// the list.
//
// Unsafe for concurrent use by multiple goroutines.
func (iter *CListIterator) WaitNextCh() <-chan Entry {
	ch := make(chan Entry)
	go func() {
		// Add the next entry to the channel if not nil.
		if entry := iter.Next(); entry != nil {
			ch <- entry.Value.(Entry)
			close(ch)
		} else {
			// Unblock the receiver (it will receive nil).
			close(ch)
		}
	}()
	return ch
}

// PickLane returns a _valid_ lane on which to iterate, according to the WRR algorithm. A lane is
// valid if it is not empty or the number of accessed entries in the lane has not yet reached its
// priority value.
func (iter *CListIterator) PickLane() types.Lane {
	// Loop until finding a valid lanes
	// If the current lane is not valid, continue with the next lane with lower priority, in a
	// round robin fashion.
	lane := iter.mp.sortedLanes[iter.currentLaneIndex]

	iter.mp.addTxChMtx.RLock()
	defer iter.mp.addTxChMtx.RUnlock()

	nIter := 0
	for {
		if iter.mp.lanes[lane].Len() == 0 ||
			(iter.cursors[lane] != nil &&
				iter.cursors[lane].Value.(*mempoolTx).seq == iter.mp.addTxLaneSeqs[lane]) {
			prevLane := lane
			iter.currentLaneIndex = (iter.currentLaneIndex + 1) % len(iter.mp.sortedLanes)
			lane = iter.mp.sortedLanes[iter.currentLaneIndex]
			nIter++
			if nIter >= len(iter.mp.sortedLanes) {
				ch := iter.mp.addTxCh
				iter.mp.addTxChMtx.RUnlock()
				iter.mp.logger.Info("YYY PickLane, bef block", "lane", lane, "prevLane", prevLane)
				<-ch
				iter.mp.logger.Info("YYY PickLane, aft block", "lane", lane, "prevLane", prevLane)
				iter.mp.addTxChMtx.RLock()
				nIter = 0
			}
			iter.mp.logger.Info("YYY PickLane, skipped lane 1", "prevLane", prevLane, "lane", lane)
			continue
		}

		if iter.counters[lane] >= uint32(lane) {
			// Reset the counter only when the limit on the lane was reached.
			iter.counters[lane] = 0
			iter.currentLaneIndex = (iter.currentLaneIndex + 1) % len(iter.mp.sortedLanes)
			prevLane := lane
			lane = iter.mp.sortedLanes[iter.currentLaneIndex]
			nIter = 0
			iter.mp.logger.Info("YYY PickLane, skipped lane 2", "lane", prevLane, "new Lane ", lane)
			continue
		}
		// TODO: if we detect that a higher-priority lane now has entries, do we preempt access to the current lane?
		iter.mp.logger.Info("YYY PickLane, returned lane", "lane", lane)
		return lane
	}
}

// Next implements the classical Weighted Round Robin (WRR) algorithm.
//
// In classical WRR, the iterator cycles over the lanes. When a lane is selected, Next returns an
// entry from the selected lane. On subsequent calls, Next will return the next entries from the
// same lane until `lane` entries are accessed or the lane is empty, where `lane` is the priority.
// The next time, Next will select the successive lane with lower priority.
//
// TODO: Note that this code does not block waiting for an available entry on a CList or a CElement, as
// was the case on the original code. Is this the best way to do it?
func (iter *CListIterator) Next() *clist.CElement {
	lane := iter.PickLane()
	// Load the last accessed entry in the lane and set the next one.
	var next *clist.CElement
	if cursor := iter.cursors[lane]; cursor != nil {
		// If the current entry is the last one or was removed, Next will return nil.
		// Note we don't need to wait until the next entry is available (with <-cursor.NextWaitChan()).
		next = cursor.Next()
	} else {
		// We are at the beginning of the iteration or the saved entry got removed. Pick the first
		// entry in the lane if it's available (don't wait for it); if not, Front will return nil.
		next = iter.mp.lanes[lane].Front()
	}

	// Update auxiliary variables.
	if next != nil {
		// Save entry and increase the number of accessed transactions for this lane.
		iter.cursors[lane] = next
		iter.counters[lane]++
	} else {
		// The entry got removed or it was the last one in the lane.
		// At the moment this should not happen - the loop in PickLane will loop forever until there
		// is data in at least one lane
		delete(iter.cursors, lane)
	}

	return next
}
