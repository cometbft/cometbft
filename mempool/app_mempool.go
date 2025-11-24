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

func WithSMMetics(metrics *Metrics) AppMempoolOpt {
	return func(m *AppMempool) { m.metrics = metrics }
}

func WithSMLogger(logger log.Logger) AppMempoolOpt {
	return func(m *AppMempool) { m.logger = logger }
}

func NewAppMempool(
	config *config.MempoolConfig,
	app AppMempoolClient,
	opts ...AppMempoolOpt,
) *AppMempool {
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
	txSize := len(tx)

	if txSize == 0 {
		return ErrEmptyTx
	}

	if m.config.MaxBatchBytes > 0 && txSize > m.config.MaxBatchBytes {
		return ErrTxTooLarge{
			Max:    m.config.MaxBatchBytes,
			Actual: txSize,
		}
	}

	pushed := m.seen.Push(tx)
	if !pushed {
		m.metrics.AlreadyReceivedTxs.Add(1)
		return ErrSeenTx
	}

	code, err := m.insertTx(tx)

	// todo (@swift1337): should we define more codes and handle them respectively?
	// todo: remove tx from seen is app returns "is full" code --> retry later
	switch {
	case err != nil:
		m.metrics.FailedTxs.Add(1)
		return errors.Wrapf(err, "unable to insert tx (tx_hash=%x, code=%d)", tx.Hash(), code)
	case code != abci.CodeTypeOK:
		m.metrics.RejectedTxs.Add(1)
		return fmt.Errorf("invalid code: %d", code)
	default:
		m.metrics.TxSizeBytes.Observe(float64(txSize))
		return nil
	}
}

func (m *AppMempool) insertTx(tx types.Tx) (uint32, error) {
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
func (m *AppMempool) TxStream(ctx context.Context, capacity int) <-chan types.Tx {
	ch := make(chan types.Tx, capacity)

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

func (m *AppMempool) reapTxs(ctx context.Context, channel chan<- types.Tx) {
	req := &abci.RequestReapTxs{
		MaxBytes: reapMaxBytes,
		MaxGas:   reapMaxGas,
	}

	ticker := time.NewTicker(reapInterval)
	defer ticker.Stop()

	for range ticker.C {
		res, err := m.app.ReapTxs(ctx, req)
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			m.logger.Debug("AppMempool.reapTxs: context done while reaping txs")
			return
		case err != nil:
			m.logger.Error("AppMempool.reapTxs: error reaping txs", "error", err)
			continue
		}

		for _, tx := range res.Txs {
			select {
			case <-ctx.Done():
				m.logger.Debug("AppMempool.reapTxs: context done while streaming txs")
				return
			case channel <- tx:
				// all good
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

// CheckTx returns an error on purpose for explicitness
func (m *AppMempool) CheckTx(_ types.Tx, _ func(*abci.ResponseCheckTx), _ TxInfo) error {
	return ErrNotImplemented
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
