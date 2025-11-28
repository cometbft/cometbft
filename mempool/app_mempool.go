package mempool

import (
	"context"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/types"
	"github.com/pkg/errors"
)

// AppMempool represents a mempool that's implemented completely on the app-side via ABCI methods in opposite to
// concurrent-list mempool that stores transactions on comet's side. AppMempool only proxies requests to the app
// and broadcasts transactions to peers. Expectations are:
// - The app is expected to handle PreCheck, PostCheck, and Recheck by itself;
// - The mempool always returns 0 txs for ReapMaxBytesMaxGas as the app is expected to build the block;
// - It doesn't block other reactors for ABCI methods --> the app is expected to handle the mempool concurrently;
type AppMempool struct {
	ctx     context.Context
	config  *config.MempoolConfig
	metrics *Metrics
	app     AppMempoolClient
	seen    TxCache
	logger  log.Logger
}

// AppMempoolClient is the interface for the app-side mempool.
type AppMempoolClient interface {
	// InsertTx inserts a tx into app-side mempool
	InsertTx(ctx context.Context, req *abci.RequestInsertTx) (*abci.ResponseInsertTx, error)

	// ReapTxs reaps txs from app-side mempool
	ReapTxs(ctx context.Context, req *abci.RequestReapTxs) (*abci.ResponseReapTxs, error)

	// Flush app's connection
	Flush(context.Context) error
}

// AppMempoolOpt is the option for AppMempool
type AppMempoolOpt func(*AppMempool)

// todo STACK-1851: move to config
const (
	seenCacheSize = 100_000
	reapMaxBytes  = 0
	reapMaxGas    = 0
	reapInterval  = 500 * time.Millisecond
)

var _ Mempool = &AppMempool{}

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrEmptyTx        = errors.New("tx is empty")
	ErrSeenTx         = errors.New("tx already seen")
)

func WithAMMetrics(metrics *Metrics) AppMempoolOpt {
	return func(m *AppMempool) { m.metrics = metrics }
}

func WithAMLogger(logger log.Logger) AppMempoolOpt {
	return func(m *AppMempool) { m.logger = logger }
}

func NewAppMempool(
	config *config.MempoolConfig,
	app AppMempoolClient,
	opts ...AppMempoolOpt,
) *AppMempool {
	// cache to avoid receiving the same txs from other peers.
	// we should add TTL w/ eviction policy.
	seen := NewLRUTxCache(seenCacheSize)

	m := &AppMempool{
		ctx:     context.Background(),
		config:  config,
		app:     app,
		seen:    seen,
		metrics: NopMetrics(),
		logger:  log.NewNopLogger(),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// InsertTx inserts a tx into app-side mempool. The call is blocking, but thread-safe.
// Concurrent calls are expected and are caller's responsibility to handle.
func (m *AppMempool) InsertTx(tx types.Tx) error {
	if err := m.guardTx(tx); err != nil {
		return err
	}

	code, err := m.insertTx(tx)

	// todo (@swift1337): should we define more codes and handle them respectively?
	switch {
	case err != nil:
		m.metrics.FailedTxs.Add(1)
		return wrapErrCode("unable to insert tx", code, err)
	case codeRetry(code):
		// drop tx from seen cache (to retry later), but still return the error
		m.seen.Remove(tx)
		fallthrough
	case code != abci.CodeTypeOK:
		m.metrics.RejectedTxs.Add(1)
		return wrapErrCode("invalid code", code, err)
	default:
		m.metrics.TxSizeBytes.Observe(float64(len(tx)))
		return nil
	}
}

// guardTx guards the tx against empty and too large errors, and adds it to the seen cache.
func (m *AppMempool) guardTx(tx types.Tx) error {
	txSize := len(tx)

	if txSize == 0 {
		return ErrEmptyTx
	}

	if m.config.MaxTxBytes > 0 && txSize > m.config.MaxTxBytes {
		return &ErrTxTooLarge{
			Max:    m.config.MaxTxBytes,
			Actual: txSize,
		}
	}

	pushed := m.seen.Push(tx)
	if !pushed {
		m.metrics.AlreadyReceivedTxs.Add(1)
		return ErrSeenTx
	}

	return nil
}

func (m *AppMempool) insertTx(tx types.Tx) (uint32, error) {
	// note: other ABCI methods panic if err is not nil
	resp, err := m.app.InsertTx(m.ctx, &abci.RequestInsertTx{Tx: tx})
	if err != nil {
		if resp != nil {
			return resp.Code, err
		}
		return 0, err
	}

	return resp.Code, nil
}

// TxStream spins up a channel that streams valid transactions from app-side mempool.
// The expectation is that the caller would share it with other peers to gossip transactions.
// chan type is a list of txs, it is guaranteed to be non-empty.
func (m *AppMempool) TxStream(ctx context.Context) <-chan types.Txs {
	ch := make(chan types.Txs, 1)

	go func() {
		defer func() {
			close(ch)

			if p := recover(); p != nil {
				m.logger.Error("panic in AppMempool.reapTxs", "panic", p)
			}
		}()

		m.reapTxs(ctx, ch)
	}()

	return ch
}

func (m *AppMempool) reapTxs(ctx context.Context, channel chan<- types.Txs) {
	req := &abci.RequestReapTxs{
		MaxBytes: reapMaxBytes,
		MaxGas:   reapMaxGas,
	}

	for {
		select {
		case <-ctx.Done():
			m.logger.Debug("AppMempool.reapTxs: context is done")
			return
		case <-time.After(reapInterval):
			// note that time.After GC mem leak was fixed in go 1.23
			res, err := m.app.ReapTxs(ctx, req)
			switch {
			case isErrCtx(err):
				m.logger.Debug("AppMempool.reapTxs: context done while reaping txs")
				return
			case err != nil:
				m.logger.Error("AppMempool.reapTxs: error reaping txs", "error", err)
				continue
			case len(res.Txs) == 0:
				// no txs to send
				continue
			}

			txs := types.ToTxs(res.Txs)

			select {
			case <-ctx.Done():
				m.logger.Debug("AppMempool.reapTxs: context done while streaming txs")
				return
			case channel <- txs:
				// all good
			}

			// avoid receiving these txs again from other peers.
			for _, tx := range txs {
				if m.seen.Has(tx) {
					continue
				}

				m.seen.Push(tx)
			}
		}
	}
}

// FlushAppConn flushes app client (copied from CListMempool)
func (m *AppMempool) FlushAppConn() error {
	err := m.app.Flush(m.ctx)
	if err != nil {
		return ErrFlushAppConn{Err: err}
	}

	return nil
}

// CheckTx mimics the behavior of the CListMempool's CheckTx method,
// but actually only calls the app's InsertTx method. This is required for RPC compatibility.
// @see: BroadcastTxSync, BroadcastTxAsync, and BroadcastTxCommit.
func (m *AppMempool) CheckTx(tx types.Tx, callback func(res *abci.ResponseCheckTx), _ TxInfo) error {
	if err := m.guardTx(tx); err != nil {
		return err
	}

	go func() {
		defer func() {
			if p := recover(); p != nil {
				m.logger.Error("panic in AppMempool.CheckTx", "panic", p, "tx", txHash(tx))
			}
		}()

		code, err := m.insertTx(tx)
		if err != nil {
			// note that other ABCI methods panic if err is not nil
			m.logger.Error("AppMempool.CheckTx: error inserting tx", "error", err, "tx", txHash(tx))
			return
		}

		// App mempool doesn't execute the tx, so we ALWAYS return an empty response here.
		// This will most likely break many clients. Clients should rely on app-specific
		// broadcasting endpoints (think of eth_sendRawTransaction, etc...).
		if callback != nil {
			callback(&abci.ResponseCheckTx{
				Code:      code,
				Data:      []byte{},
				Log:       "",
				Info:      "",
				GasWanted: 0,
				GasUsed:   0,
				Events:    []abci.Event{},
				Codespace: "",
			})
		}
	}()

	return nil
}

// Update does nothing for an app mempool
func (m *AppMempool) Update(_ int64, _ types.Txs, _ []*abci.ExecTxResult, _ PreCheckFunc, _ PostCheckFunc) error {
	return nil
}

// reading from this channel will block forever, which is fine for an app mempool
func (m *AppMempool) TxsAvailable() <-chan struct{} { return nil }
func (m *AppMempool) EnableTxsAvailable()           {}

func (m *AppMempool) Size() int        { return 0 }
func (m *AppMempool) SizeBytes() int64 { return 0 }

func (m *AppMempool) ReapMaxBytesMaxGas(_, _ int64) types.Txs { return nil }
func (m *AppMempool) ReapMaxTxs(_ int) types.Txs              { return nil }
func (m *AppMempool) RemoveTxByKey(_ types.TxKey) error       { return nil }
func (m *AppMempool) Flush()                                  {}

func (m *AppMempool) Lock()   {}
func (m *AppMempool) Unlock() {}

func isErrCtx(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func codeRetry(code uint32) bool {
	return code >= abci.CodeTypeRetry
}

func wrapErrCode(msg string, code uint32, err error) error {
	if err == nil {
		return fmt.Errorf("%s: (code=%d)", msg, code)
	}

	return errors.Wrapf(err, "%s: (code=%d)", msg, code)
}
