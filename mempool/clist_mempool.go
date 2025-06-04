package mempool

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	abcicli "github.com/cometbft/cometbft/v2/abci/client"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/config"
	"github.com/cometbft/cometbft/v2/internal/clist"
	"github.com/cometbft/cometbft/v2/libs/log"
	cmtmath "github.com/cometbft/cometbft/v2/libs/math"
	cmtsync "github.com/cometbft/cometbft/v2/libs/sync"
	"github.com/cometbft/cometbft/v2/p2p"
	"github.com/cometbft/cometbft/v2/proxy"
	"github.com/cometbft/cometbft/v2/types"
	cmttime "github.com/cometbft/cometbft/v2/types/time"
)

const (
	noSender    = ""
	defaultLane = "default"
)

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
	txsMtx    cmtsync.RWMutex
	lanes     map[LaneID]*clist.CList         // each lane is a linked-list of (valid) txs
	txsMap    map[types.TxKey]*clist.CElement // for quick access to the mempool entry of a given tx
	laneBytes map[LaneID]int64                // number of bytes per lane (for metrics)
	txsBytes  int64                           // total size of mempool, in bytes
	numTxs    int64                           // total number of txs in the mempool

	addTxChMtx    cmtsync.RWMutex  // Protects the fields below
	addTxCh       chan struct{}    // Blocks until the next TX is added
	addTxSeq      int64            // Helps detect is new TXs have been added to a given lane
	addTxLaneSeqs map[LaneID]int64 // Sequence of the last TX added to a given lane

	// Immutable fields, only set during initialization.
	defaultLane LaneID
	sortedLanes []lane // lanes sorted by priority, in descending order

	// Keep a cache of already-seen txs.
	// This reduces the pressure on the proxyApp.
	cache TxCache

	logger  log.Logger
	metrics *Metrics
}

var _ Mempool = &CListMempool{}

// CListMempoolOption sets an optional parameter on the mempool.
type CListMempoolOption func(*CListMempool)

// A LaneID is a string that uniquely identifies a lane.
// Multiple lanes can have the same priority.
type LaneID string

// LanePriority represents the priority of a lane.
type LanePriority uint32

// lane corresponds to a transaction class as defined by the application.
// A lane is identified by a unique string name (LaneID) and has a priority level (LanePriority).
// Different lanes can have the same priority.
type lane struct {
	id       LaneID
	priority LanePriority
}

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
		laneBytes:     make(map[LaneID]int64),
		logger:        log.NewNopLogger(),
		metrics:       NopMetrics(),
		addTxCh:       make(chan struct{}),
		addTxLaneSeqs: make(map[LaneID]int64),
	}
	mp.height.Store(height)

	// Initialize lanes
	if lanesInfo == nil || len(lanesInfo.lanes) == 0 {
		// The only lane will be "default" with priority 1.
		lanesInfo = &LanesInfo{lanes: map[LaneID]LanePriority{defaultLane: 1}, defaultLane: defaultLane}
	}
	numLanes := len(lanesInfo.lanes)
	mp.lanes = make(map[LaneID]*clist.CList, numLanes)
	mp.defaultLane = lanesInfo.defaultLane
	mp.sortedLanes = make([]lane, 0, numLanes)
	for id, priority := range lanesInfo.lanes {
		mp.lanes[id] = clist.New()
		mp.sortedLanes = append(mp.sortedLanes, lane{id: id, priority: priority})
	}
	slices.SortStableFunc(mp.sortedLanes, func(i, j lane) int {
		if i.priority > j.priority {
			return -1
		}
		if i.priority < j.priority {
			return 1
		}
		return 0
	})

	mp.recheck = newRecheck(mp)

	if cfg.CacheSize > 0 {
		mp.cache = NewLRUTxCache(cfg.CacheSize)
	} else {
		mp.cache = NopTxCache{}
	}

	for _, option := range options {
		option(mp)
	}

	return mp
}

func (mem *CListMempool) GetSenders(txKey types.TxKey) ([]p2p.ID, error) {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	elem, ok := mem.txsMap[txKey]
	if !ok {
		return nil, ErrTxNotFound
	}
	memTx := elem.Value.(*mempoolTx)
	return memTx.Senders(), nil
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

func (mem *CListMempool) removeAllTxs(lane LaneID) {
	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	for e := mem.lanes[lane].Front(); e != nil; e = e.Next() {
		mem.lanes[lane].Remove(e)
		e.DetachPrev()
	}
	mem.txsMap = make(map[types.TxKey]*clist.CElement)
	delete(mem.laneBytes, lane)
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

// Lock acquires the exclusive lock for mempool updates.
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Lock() {
	mem.updateMtx.Lock()
}

// Unlock releases the exclusive lock for mempool updates.
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Unlock() {
	mem.updateMtx.Unlock()
}

// PreUpdate sets the recheckFull flag and logs if its state changes.
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

// LaneSizes returns, the number of transactions in the given lane and the total
// number of bytes used by all transactions in the lane.
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) LaneSizes(lane LaneID) (numTxs int, bytes int64) {
	mem.txsMtx.RLock()
	defer mem.txsMtx.RUnlock()

	bytes = mem.laneBytes[lane]

	txs, ok := mem.lanes[lane]
	if !ok {
		panic(ErrLaneNotFound{laneID: lane})
	}
	return txs.Len(), bytes
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
		mem.metrics.RejectedTxs.Add(1)
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
			mem.logger.Error("Could not add sender to tx", "tx", log.NewLazyHash(tx), "sender", sender, "err", err)
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
		panic(fmt.Errorf("CheckTx request for tx %s failed: %w", tx.Hash(), err))
	}
	reqRes.SetCallback(mem.handleCheckTxResponse(tx, sender))

	return reqRes, nil
}

// handleCheckTxResponse handles CheckTx responses for transactions validated for the first time.
//
//   - sender optionally holds the ID of the peer that sent the transaction, if any.
func (mem *CListMempool) handleCheckTxResponse(tx types.Tx, sender p2p.ID) func(res *abci.Response) error {
	return func(r *abci.Response) error {
		res := r.GetCheckTx()
		if res == nil {
			panic(fmt.Sprintf("unexpected response value %v not of type CheckTx", r))
		}

		// Check that rechecking txs is not in process.
		if !mem.recheck.done() {
			panic(fmt.Sprint("rechecking has not finished; cannot check new tx ", tx.Hash()))
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
				"tx", log.NewLazyHash(tx),
				"res", res,
				"err", postCheckErr,
			)
			mem.metrics.FailedTxs.Add(1)

			if postCheckErr != nil {
				return postCheckErr
			}
			return ErrInvalidTx{Code: res.Code, Data: res.Data, Log: res.Log, Codespace: res.Codespace, Hash: tx.Hash()}
		}

		// If the app returned a non-empty lane, use it; otherwise use the default lane.
		lane := mem.defaultLane
		if res.LaneId != "" {
			if _, ok := mem.lanes[lane]; !ok {
				panic(ErrLaneNotFound{laneID: lane})
			}
			lane = LaneID(res.LaneId)
		}

		if err := mem.isLaneFull(len(tx), lane); err != nil {
			mem.forceRemoveFromCache(tx) // lane might have space later
			// use debug level to avoid spamming logs when traffic is high
			mem.logger.Debug(err.Error())
			mem.metrics.RejectedTxs.Add(1)
			return err
		}

		// Check that tx is not already in the mempool. This can happen when the
		// cache overflows. See https://github.com/cometbft/cometbft/v2/pull/890.
		txKey := tx.Key()
		if mem.Contains(txKey) {
			mem.metrics.RejectedTxs.Add(1)
			if err := mem.addSender(txKey, sender); err != nil {
				mem.logger.Error("Could not add sender to tx", "tx", tx.Hash(), "sender", sender, "err", err)
			}
			mem.logger.Debug("Reject tx", "tx", log.NewLazyHash(tx), "height", mem.height.Load(), "err", ErrTxInMempool)
			return ErrTxInMempool
		}

		// Add tx to mempool and notify that new txs are available.
		mem.addTx(tx, res.GasWanted, sender, lane)
		mem.notifyTxsAvailable()

		if mem.onNewTx != nil {
			mem.onNewTx(tx)
		}

		mem.updateSizeMetrics(lane)

		return nil
	}
}

// Called from:
//   - handleCheckTxResponse (lock not held) if tx is valid
func (mem *CListMempool) addTx(tx types.Tx, gasWanted int64, sender p2p.ID, lane LaneID) {
	mem.txsMtx.Lock()
	defer mem.txsMtx.Unlock()

	// Get lane's clist.
	txs, ok := mem.lanes[lane]
	if !ok {
		panic(ErrLaneNotFound{laneID: lane})
	}

	// Increase sequence number.
	mem.addTxChMtx.Lock()
	defer mem.addTxChMtx.Unlock()
	mem.addTxSeq++
	mem.addTxLaneSeqs[lane] = mem.addTxSeq

	// Add new transaction.
	memTx := &mempoolTx{
		tx:        tx,
		height:    mem.height.Load(),
		gasWanted: gasWanted,
		lane:      lane,
		seq:       mem.addTxSeq,
	}
	_ = memTx.addSender(sender)
	e := txs.PushBack(memTx)

	// Update auxiliary variables.
	mem.txsMap[tx.Key()] = e
	mem.txsBytes += int64(len(tx))
	mem.numTxs++
	mem.laneBytes[lane] += int64(len(tx))

	// Notify iterators there's a new transaction.
	close(mem.addTxCh)
	mem.addTxCh = make(chan struct{})

	// Update metrics.
	mem.metrics.TxSizeBytes.Observe(float64(len(tx)))

	mem.logger.Debug(
		"Added transaction",
		"tx", log.NewLazyHash(tx),
		"lane", lane,
		"height", mem.height.Load(),
		"total", mem.numTxs,
	)
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

	label := string(memTx.lane)
	mem.metrics.TxLifeSpan.With("lane", label).Observe(float64(memTx.timestamp.Sub(time.Now().UTC())))

	// Remove tx from lane.
	mem.lanes[memTx.lane].Remove(elem)
	elem.DetachPrev()

	// Update auxiliary variables.
	delete(mem.txsMap, txKey)
	mem.txsBytes -= int64(len(memTx.tx))
	mem.numTxs--
	mem.laneBytes[memTx.lane] -= int64(len(memTx.tx))

	mem.logger.Debug(
		"Removed transaction",
		"tx", log.NewLazyHash(memTx.tx),
		"lane", memTx.lane,
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

func (mem *CListMempool) isLaneFull(txSize int, lane LaneID) error {
	laneTxs, laneBytes := mem.LaneSizes(lane)

	// The mempool is partitioned evenly across all lanes.
	laneTxsCapacity := mem.config.Size / len(mem.sortedLanes)
	laneBytesCapacity := mem.config.MaxTxsBytes / int64(len(mem.sortedLanes))

	if laneTxs > laneTxsCapacity || int64(txSize)+laneBytes > laneBytesCapacity {
		return ErrLaneIsFull{
			Lane:     lane,
			NumTxs:   laneTxs,
			MaxTxs:   laneTxsCapacity,
			Bytes:    laneBytes,
			MaxBytes: laneBytesCapacity,
		}
	}

	if mem.recheck.consideredFull() {
		return ErrRecheckFull
	}

	return nil
}

// handleRecheckTxResponse handles CheckTx responses for transactions in the mempool that need to be
// revalidated after a mempool update.
func (mem *CListMempool) handleRecheckTxResponse(tx types.Tx) func(res *abci.Response) error {
	return func(r *abci.Response) error {
		res := r.GetCheckTx()
		if res == nil {
			panic(fmt.Sprintf("unexpected response value %v not of type CheckTx", r))
		}

		// Check whether the rechecking process has finished.
		if mem.recheck.done() {
			mem.logger.Error("Failed to recheck tx", "tx", log.NewLazyHash(tx), "err", ErrLateRecheckResponse)
			return ErrLateRecheckResponse
		}
		mem.metrics.RecheckTimes.Add(1)

		// Check whether tx is still in the list of transactions that can be rechecked.
		if !mem.recheck.findNextEntryMatching(&tx) {
			// Reached the end of the list and didn't find a matching tx; rechecking has finished.
			return nil
		}

		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx, res)
		}

		// If tx is invalid, remove it from the mempool and the cache.
		if (res.Code != abci.CodeTypeOK) || postCheckErr != nil {
			// Tx became invalidated due to newly committed block.
			mem.logger.Debug("Tx is no longer valid", "tx", log.NewLazyHash(tx), "res", res, "postCheckErr", postCheckErr)
			if err := mem.RemoveTxByKey(tx.Key()); err != nil {
				mem.logger.Debug("Transaction could not be removed from mempool", "err", err)
				return err
			}

			// update metrics
			mem.metrics.EvictedTxs.Add(1)
			if elem, ok := mem.txsMap[tx.Key()]; ok {
				mem.updateSizeMetrics(elem.Value.(*mempoolTx).lane)
			} else {
				mem.logger.Error("Cannot update metrics", "err", ErrTxNotFound)
			}

			mem.tryRemoveFromCache(tx)
			if postCheckErr != nil {
				return postCheckErr
			}
			return ErrInvalidTx{Code: res.Code, Data: res.Data, Log: res.Log, Codespace: res.Codespace, Hash: tx.Hash()}
		}

		return nil
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
	iter := NewNonBlockingIterator(mem)
	for {
		memTx := iter.Next()
		if memTx == nil {
			break
		}
		txs = append(txs, memTx.Tx())

		dataSize := types.ComputeProtoSizeForTxs([]types.Tx{memTx.Tx()})

		// Check total size requirement
		if maxBytes > -1 && runningSize+dataSize > maxBytes {
			return txs[:len(txs)-1]
		}

		runningSize += dataSize

		// Check total gas requirement.
		// If maxGas is negative, skip this check.
		// Since newTotalGas < masGas, which
		// must be non-negative, it follows that this won't overflow.
		newTotalGas := totalGas + memTx.GasWanted()
		if maxGas > -1 && newTotalGas > maxGas {
			return txs[:len(txs)-1]
		}
		totalGas = newTotalGas
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
	iter := NewNonBlockingIterator(mem)
	for len(txs) <= max {
		memTx := iter.Next()
		if memTx == nil {
			break
		}
		txs = append(txs, memTx.Tx())
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
				"tx", log.NewLazyHash(tx),
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
	for lane := range mem.lanes {
		mem.updateSizeMetrics(lane)
	}

	return nil
}

// updateSizeMetrics updates the size-related metrics of a given lane.
func (mem *CListMempool) updateSizeMetrics(laneID LaneID) {
	laneTxs, laneBytes := mem.LaneSizes(laneID)
	label := string(laneID)
	mem.metrics.LaneSize.With("lane", label).Set(float64(laneTxs))
	mem.metrics.LaneBytes.With("lane", label).Set(float64(laneBytes))
	mem.metrics.Size.Set(float64(mem.Size()))
	mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))
}

// recheckTxs sends all transactions in the mempool to the app for re-validation. When the function
// returns, all recheck responses from the app have been processed.
func (mem *CListMempool) recheckTxs() {
	mem.logger.Debug("Recheck txs", "height", mem.height.Load(), "num-txs", mem.Size())

	if mem.Size() <= 0 {
		return
	}

	defer func(start time.Time) {
		mem.metrics.RecheckDurationSeconds.Set(cmttime.Since(start).Seconds())
	}(cmttime.Now())

	mem.recheck.init()

	iter := NewNonBlockingIterator(mem)
	for {
		memTx := iter.Next()
		if memTx == nil {
			break
		}

		// NOTE: handleCheckTxResponse may be called concurrently, but CheckTx cannot be executed concurrently
		// because this function has the lock (via Update and Lock).
		mem.recheck.numPendingTxs.Add(1)

		// Send CheckTx request to the app to re-validate transaction.
		resReq, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.CheckTxRequest{
			Tx:   memTx.Tx(),
			Type: abci.CHECK_TX_TYPE_RECHECK,
		})
		if err != nil {
			panic(fmt.Errorf("(re-)CheckTx request for tx %s failed: %w", memTx.Tx().Hash(), err))
		}
		resReq.SetCallback(mem.handleRecheckTxResponse(memTx.Tx()))
	}

	// Flush any pending asynchronous recheck requests to process.
	mem.proxyAppConn.Flush(context.TODO())

	// Give some time to finish processing the responses; then finish the rechecking process, even
	// if not all txs were rechecked.
	select {
	case <-time.After(mem.config.RecheckTimeout):
		mem.recheck.setDone()
		mem.logger.Error("Timed out waiting for recheck responses")
	case <-mem.recheck.doneRechecking():
	}

	if n := mem.recheck.numPendingTxs.Load(); n > 0 {
		mem.logger.Error("Not all txs were rechecked", "not-rechecked", n)
	}

	mem.logger.Debug("Done rechecking", "height", mem.height.Load(), "num-txs", mem.Size())
}

// When a recheck response for a transaction is received, cursor will point to
// the entry in the mempool corresponding to that transaction, advancing the
// cursor, thus narrowing the list of transactions to recheck. In case there are
// entries between the previous and the current positions of cursor, they will
// be ignored for rechecking. This is to guarantee that recheck responses are
// processed in the same sequential order as they appear in the mempool.
type recheck struct {
	iter          *NonBlockingIterator
	cursor        Entry         // next expected recheck response
	doneCh        chan struct{} // to signal that rechecking has finished successfully (for async app connections)
	numPendingTxs atomic.Int32  // number of transactions still pending to recheck
	isRechecking  atomic.Bool   // true iff the rechecking process has begun and is not yet finished
	recheckFull   atomic.Bool   // whether rechecking TXs cannot be completed before a new block is decided
	mem           *CListMempool
}

func newRecheck(mp *CListMempool) *recheck {
	r := recheck{}
	r.iter = NewNonBlockingIterator(mp)
	r.mem = mp
	return &r
}

func (rc *recheck) init() {
	if !rc.done() {
		panic("Having more than one rechecking process at a time is not possible.")
	}
	rc.numPendingTxs.Store(0)
	rc.iter = NewNonBlockingIterator(rc.mem)

	rc.cursor = rc.iter.Next()
	rc.doneCh = make(chan struct{})
	if rc.cursor == nil {
		rc.setDone()
		return
	}
	rc.isRechecking.Store(true)
	rc.recheckFull.Store(false)
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
	close(rc.doneCh) // notify channel that recheck has finished
}

// findNextEntryMatching searches for the next transaction matching the given transaction, which
// corresponds to the recheck response to be processed next. Then it checks if it has reached the
// end of the list, so it can set recheck as finished.
//
// The goal is to guarantee that transactions are rechecked in the order in which they are in the
// mempool. Transactions whose recheck response arrive late or don't arrive at all are skipped and
// not rechecked.
func (rc *recheck) findNextEntryMatching(tx *types.Tx) (found bool) {
	for rc.cursor != nil {
		expectedTx := rc.cursor.Tx()
		rc.cursor = rc.iter.Next()
		if bytes.Equal(*tx, expectedTx) {
			// Found an entry in the list of txs to recheck that matches tx.
			found = true
			rc.numPendingTxs.Add(-1)
			break
		}
	}

	if rc.cursor == nil { // reached end of list
		rc.setDone()
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
