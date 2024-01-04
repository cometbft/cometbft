package mempool

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/clist"
	cmtsync "github.com/cometbft/cometbft/internal/sync"
	"github.com/cometbft/cometbft/libs/log"
	cmtmath "github.com/cometbft/cometbft/libs/math"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

// CListMempool is an ordered in-memory pool for transactions before they are
// proposed in a consensus round. Transaction validity is checked using the
// CheckTx abci message before the transaction is added to the pool. The
// mempool uses a concurrent list structure for storing transactions that can
// be efficiently accessed by multiple concurrent readers.
type CListMempool struct {
	// Atomic integers
	height   int64 // the last block Update()'d to
	txsBytes int64 // total size of mempool, in bytes

	// notify listeners (ie. consensus) when txs are available
	notifiedTxsAvailable bool
	txsAvailable         chan struct{} // fires once for each height, when the mempool is not empty

	// Function set by the reactor to be called when a transaction is removed
	// from the mempool.
	removeTxOnReactorCb func(txKey types.TxKey)

	config *config.MempoolConfig

	// Exclusive mutex for Update method to prevent concurrent execution of
	// CheckTx or ReapMaxBytesMaxGas(ReapMaxTxs) methods.
	updateMtx cmtsync.RWMutex
	preCheck  PreCheckFunc
	postCheck PostCheckFunc

	proxyAppConn proxy.AppConnMempool

	// Track whether we're rechecking txs.
	// These are not protected by a mutex and are expected to be mutated in
	// serial (ie. by abci responses which are called in serial).
	recheckCursor *clist.CElement // next expected response
	recheckEnd    *clist.CElement // re-checking stops here

	// Concurrent linked-list of valid txs.
	// `txsMap`: txKey -> CElement is for quick access to txs.
	// Transactions in both `txs` and `txsMap` must to be kept in sync.
	txs    *clist.CList
	txsMap sync.Map

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
	height int64,
	options ...CListMempoolOption,
) *CListMempool {
	mp := &CListMempool{
		config:        cfg,
		proxyAppConn:  proxyAppConn,
		txs:           clist.New(),
		height:        height,
		recheckCursor: nil,
		recheckEnd:    nil,
		logger:        log.NewNopLogger(),
		metrics:       NopMetrics(),
	}

	if cfg.CacheSize > 0 {
		mp.cache = NewLRUTxCache(cfg.CacheSize)
	} else {
		mp.cache = NopTxCache{}
	}

	proxyAppConn.SetResponseCallback(mp.globalCb)

	for _, option := range options {
		option(mp)
	}

	return mp
}

func (mem *CListMempool) getCElement(txKey types.TxKey) (*clist.CElement, bool) {
	if e, ok := mem.txsMap.Load(txKey); ok {
		return e.(*clist.CElement), true
	}
	return nil, false
}

func (mem *CListMempool) InMempool(txKey types.TxKey) bool {
	_, ok := mem.getCElement(txKey)
	return ok
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

func (mem *CListMempool) removeAllTxs() {
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		mem.txs.Remove(e)
		e.DetachPrev()
	}

	mem.txsMap.Range(func(key, _ interface{}) bool {
		mem.txsMap.Delete(key)
		mem.invokeRemoveTxOnReactor(key.(types.TxKey))
		return true
	})
}

// NOTE: not thread safe - should only be called once, on startup.
func (mem *CListMempool) EnableTxsAvailable() {
	mem.txsAvailable = make(chan struct{}, 1)
}

func (mem *CListMempool) SetTxRemovedCallback(cb func(txKey types.TxKey)) {
	mem.removeTxOnReactorCb = cb
}

func (mem *CListMempool) invokeRemoveTxOnReactor(txKey types.TxKey) {
	// Note that the callback is nil in the unit tests, where there are no
	// reactors.
	if mem.removeTxOnReactorCb != nil {
		mem.removeTxOnReactorCb(txKey)
	}
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

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Lock() {
	mem.updateMtx.Lock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Unlock() {
	mem.updateMtx.Unlock()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) Size() int {
	return mem.txs.Len()
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) SizeBytes() int64 {
	return atomic.LoadInt64(&mem.txsBytes)
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
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	_ = atomic.SwapInt64(&mem.txsBytes, 0)
	mem.cache.Reset()

	mem.removeAllTxs()
}

// TxsFront returns the first transaction in the ordered list for peer
// goroutines to call .NextWait() on.
// FIXME: leaking implementation details!
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsFront() *clist.CElement {
	return mem.txs.Front()
}

// TxsWaitChan returns a channel to wait on transactions. It will be closed
// once the mempool is not empty (ie. the internal `mem.txs` has at least one
// element)
//
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) TxsWaitChan() <-chan struct{} {
	return mem.txs.WaitChan()
}

// It blocks if we're waiting on Update() or Reap().
// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) CheckTx(tx types.Tx) (*abcicli.ReqRes, error) {
	mem.updateMtx.RLock()
	mem.logger.Debug("Locked updateMtx for read", "tx", tx)
	// use defer to unlock mutex because application (*local client*) might panic
	defer mem.updateMtx.RUnlock()
	defer mem.logger.Debug("Released updateMtx for read", "tx", tx)

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
		mem.logger.Debug("Not cached", "tx", tx)
		mem.metrics.AlreadyReceivedTxs.Add(1)
		// TODO: consider punishing peer for dups,
		// its non-trivial since invalid txs can become valid,
		// but they can spam the same tx with little cost to them atm.
		return nil, ErrTxInCache
	}
	mem.logger.Debug("Cached", "tx", tx)

	reqRes, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.CheckTxRequest{
		Tx:   tx,
		Type: abci.CHECK_TX_TYPE_CHECK,
	})
	if err != nil {
		mem.logger.Error("RequestCheckTx", "err", err)
		return nil, ErrCheckTxAsync{Err: err}
	}

	return reqRes, nil
}

// Global callback that will be called after every ABCI response.
func (mem *CListMempool) globalCb(req *abci.Request, res *abci.Response) {
	switch res.Value.(type) {
	case *abci.Response_CheckTx:
		checkType := req.GetCheckTx().GetType()
		switch checkType {
		case abci.CHECK_TX_TYPE_CHECK:
			if mem.recheckCursor != nil {
				// this should never happen
				panic("recheck cursor is not nil before resCbFirstTime")
			}
			mem.resCbFirstTime(req.GetCheckTx().Tx, res)

		case abci.CHECK_TX_TYPE_RECHECK:
			if mem.recheckCursor == nil {
				return
			}
			mem.metrics.RecheckTimes.Add(1)
			mem.resCbRecheck(req, res)

		default:
			panic(fmt.Sprintf("unexpected value %d of RequestCheckTx.type", checkType))
		}

		// update metrics
		mem.metrics.Size.Set(float64(mem.Size()))
		mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))

	default:
		// ignore other messages
	}
}

// Called from:
//   - resCbFirstTime (lock not held) if tx is valid
func (mem *CListMempool) addTx(memTx *mempoolTx) {
	e := mem.txs.PushBack(memTx)
	mem.txsMap.Store(memTx.tx.Key(), e)
	atomic.AddInt64(&mem.txsBytes, int64(len(memTx.tx)))
	mem.metrics.TxSizeBytes.Observe(float64(len(memTx.tx)))
	mem.logger.Debug("Clisted", "tx", memTx.tx)
}

// RemoveTxByKey removes a transaction from the mempool by its TxKey index.
// Called from:
//   - Update (lock held) if tx was committed
//   - resCbRecheck (lock not held) if tx was invalidated
func (mem *CListMempool) RemoveTxByKey(txKey types.TxKey) error {
	// The transaction should be removed from the reactor, even if it cannot be
	// found in the mempool.
	mem.invokeRemoveTxOnReactor(txKey)
	if elem, ok := mem.getCElement(txKey); ok {
		mem.txs.Remove(elem)
		elem.DetachPrev()
		mem.txsMap.Delete(txKey)
		tx := elem.Value.(*mempoolTx).tx
		atomic.AddInt64(&mem.txsBytes, int64(-len(tx)))
		return nil
	}
	return ErrTxNotFound
}

func (mem *CListMempool) isFull(txSize int) error {
	var (
		memSize  = mem.Size()
		txsBytes = mem.SizeBytes()
	)

	if memSize >= mem.config.Size || int64(txSize)+txsBytes > mem.config.MaxTxsBytes {
		return ErrMempoolIsFull{
			NumTxs:      memSize,
			MaxTxs:      mem.config.Size,
			TxsBytes:    txsBytes,
			MaxTxsBytes: mem.config.MaxTxsBytes,
		}
	}

	return nil
}

// callback, which is called after the app checked the tx for the first time.
//
// The case where the app checks the tx for the second and subsequent times is
// handled by the resCbRecheck callback.
func (mem *CListMempool) resCbFirstTime(
	tx []byte,
	res *abci.Response,
) {
	switch r := res.Value.(type) {
	case *abci.Response_CheckTx:
		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx, r.CheckTx)
		}
		txKey := types.Tx(tx).Key()
		if (r.CheckTx.Code == abci.CodeTypeOK) && postCheckErr == nil {
			// Check mempool isn't full again to reduce the chance of exceeding the
			// limits.
			if err := mem.isFull(len(tx)); err != nil {
				mem.forceRemoveFromCache(tx) // mempool might have space later
				mem.logger.Error(err.Error())
				return
			}

			// Check transaction not already in the mempool
			if mem.InMempool(txKey) {
				mem.logger.Debug(
					"transaction already there, not adding it again",
					"tx", types.Tx(tx).Hash(),
					"res", r,
					"height", mem.height,
					"total", mem.Size(),
				)
				return
			}

			mem.addTx(&mempoolTx{
				height:    mem.height,
				gasWanted: r.CheckTx.GasWanted,
				tx:        tx,
			})
			mem.logger.Debug(
				"added valid transaction",
				"tx", types.Tx(tx).Hash(),
				"res", r,
				"height", mem.height,
				"total", mem.Size(),
			)
			mem.notifyTxsAvailable()
		} else {
			mem.tryRemoveFromCache(tx)
			mem.logger.Debug(
				"rejected invalid transaction",
				"tx", types.Tx(tx).Hash(),
				"res", r,
				"err", postCheckErr,
			)
			mem.metrics.FailedTxs.Add(1)
		}

	default:
		// ignore other messages
	}
}

// callback, which is called after the app rechecked the tx.
//
// The case where the app checks the tx for the first time is handled by the
// resCbFirstTime callback.
func (mem *CListMempool) resCbRecheck(req *abci.Request, res *abci.Response) {
	switch r := res.Value.(type) {
	case *abci.Response_CheckTx:
		tx := req.GetCheckTx().Tx
		memTx := mem.recheckCursor.Value.(*mempoolTx)

		// Search through the remaining list of tx to recheck for a transaction that matches
		// the one we received from the ABCI application.
		for {
			if bytes.Equal(tx, memTx.tx) {
				// We've found a tx in the recheck list that matches the tx that we
				// received from the ABCI application.
				// Break, and use this transaction for further checks.
				break
			}

			mem.logger.Error(
				"re-CheckTx transaction mismatch",
				"got", types.Tx(tx),
				"expected", memTx.tx,
			)

			if mem.recheckCursor == mem.recheckEnd {
				// we reached the end of the recheckTx list without finding a tx
				// matching the one we received from the ABCI application.
				// Return without processing any tx.
				mem.recheckCursor = nil
				return
			}

			mem.recheckCursor = mem.recheckCursor.Next()
			memTx = mem.recheckCursor.Value.(*mempoolTx)
		}

		var postCheckErr error
		if mem.postCheck != nil {
			postCheckErr = mem.postCheck(tx, r.CheckTx)
		}

		if (r.CheckTx.Code != abci.CodeTypeOK) || postCheckErr != nil {
			// Tx became invalidated due to newly committed block.
			mem.logger.Debug("tx is no longer valid", "tx", types.Tx(tx).Hash(), "res", r, "err", postCheckErr)
			if err := mem.RemoveTxByKey(memTx.tx.Key()); err != nil {
				mem.logger.Debug("Transaction could not be removed from mempool", "err", err)
			}
			mem.tryRemoveFromCache(tx)
		}
		if mem.recheckCursor == mem.recheckEnd {
			mem.recheckCursor = nil
		} else {
			mem.recheckCursor = mem.recheckCursor.Next()
		}
		if mem.recheckCursor == nil {
			// Done!
			mem.logger.Debug("done rechecking txs")

			// in case the recheck removed all txs
			if mem.Size() > 0 {
				mem.notifyTxsAvailable()
			}
		}
	default:
		// ignore other messages
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
	if mem.txsAvailable != nil && !mem.notifiedTxsAvailable {
		// channel cap is 1, so this will send once
		mem.notifiedTxsAvailable = true
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
	// txs := make([]types.Tx, 0, cmtmath.MinInt(mem.txs.Len(), max/mem.avgTxSize))
	txs := make([]types.Tx, 0, mem.txs.Len())
	for e := mem.txs.Front(); e != nil; e = e.Next() {
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
	return txs
}

// Safe for concurrent use by multiple goroutines.
func (mem *CListMempool) ReapMaxTxs(max int) types.Txs {
	mem.updateMtx.RLock()
	defer mem.updateMtx.RUnlock()

	if max < 0 {
		max = mem.txs.Len()
	}

	txs := make([]types.Tx, 0, cmtmath.MinInt(mem.txs.Len(), max))
	for e := mem.txs.Front(); e != nil && len(txs) <= max; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		txs = append(txs, memTx.tx)
	}
	return txs
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
	// Set height
	mem.height = height
	mem.notifiedTxsAvailable = false

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
				"key", tx.Key(),
				"error", err.Error())
		}
	}

	// Either recheck non-committed txs to see if they became invalid
	// or just notify there're some txs left.
	if mem.Size() > 0 {
		if mem.config.Recheck {
			mem.logger.Debug("recheck txs", "numtxs", mem.Size(), "height", height)
			mem.recheckTxs()
			// At this point, mem.txs are being rechecked.
			// mem.recheckCursor re-scans mem.txs and possibly removes some txs.
			// Before mem.Reap(), we should wait for mem.recheckCursor to be nil.
		} else {
			mem.notifyTxsAvailable()
		}
	}

	// Update metrics
	mem.metrics.Size.Set(float64(mem.Size()))
	mem.metrics.SizeBytes.Set(float64(mem.SizeBytes()))

	return nil
}

func (mem *CListMempool) recheckTxs() {
	if mem.Size() == 0 {
		panic("recheckTxs is called, but the mempool is empty")
	}

	mem.recheckCursor = mem.txs.Front()
	mem.recheckEnd = mem.txs.Back()

	// Push txs to proxyAppConn
	// NOTE: globalCb may be called concurrently.
	for e := mem.txs.Front(); e != nil; e = e.Next() {
		memTx := e.Value.(*mempoolTx)
		_, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.CheckTxRequest{
			Tx:   memTx.tx,
			Type: abci.CHECK_TX_TYPE_RECHECK,
		})
		if err != nil {
			mem.logger.Error("recheckTx", err, "err")
			return
		}
	}

	// In <v0.37 we would call FlushAsync at the end of recheckTx forcing the buffer to flush
	// all pending messages to the app. There doesn't seem to be any need here as the buffer
	// will get flushed regularly or when filled.
}
