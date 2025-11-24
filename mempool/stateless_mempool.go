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

// StatelessMempool represents a mempool that is implemented completely on the app-side via ABCI methods.
// This entity is only responsible for receiving and broadcasting transactions between peers.
//
// Expectations:
// - The app is expected to handle PreCheck, PostCheck, and Recheck by itself;
// - The mempool always returns 0 txs for ReapMaxBytesMaxGas as the app is expected to build the block;
// - It doesn't block other reactors for ABCI methods as the app is expected to handle the mempool concurrently;
type StatelessMempool struct {
	ctx     context.Context
	config  *config.MempoolConfig
	metrics *Metrics
	app     AppMempool
	seen    TxCache
	logger  log.Logger
}

// AppMempool is the interface for the app-side mempool.
type AppMempool interface {
	// InsertTx inserts a tx into app-side mempool
	InsertTx(ctx context.Context, req *abci.RequestInsertTx) (*abci.ResponseInsertTx, error)

	// ReapTxs reaps txs from app-side mempool
	ReapTxs(ctx context.Context, req *abci.RequestReapTxs) (*abci.ResponseReapTxs, error)

	// Flush app's connection
	Flush(context.Context) error
}

// StatelessMempoolOpt is the option for StatelessMempool
type StatelessMempoolOpt func(*StatelessMempool)

// todo STACK-1851: move to config
const (
	seenCacheSize = 100_000
	reapMaxBytes  = 0
	reapMaxGas    = 0
	reapInterval  = 500 * time.Millisecond
)

var _ Mempool = &StatelessMempool{}

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrEmptyTx        = errors.New("tx is empty")
	ErrSeenTx         = errors.New("tx already seen")
)

func WithSMMetics(metrics *Metrics) StatelessMempoolOpt {
	return func(m *StatelessMempool) { m.metrics = metrics }
}

func WithSMLogger(logger log.Logger) StatelessMempoolOpt {
	return func(m *StatelessMempool) { m.logger = logger }
}

func NewStatelessMempool(
	config *config.MempoolConfig,
	app AppMempool,
	opts ...StatelessMempoolOpt,
) *StatelessMempool {
	seen := NewLRUTxCache(seenCacheSize)

	m := &StatelessMempool{
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
func (m *StatelessMempool) InsertTx(tx types.Tx) error {
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

func (m *StatelessMempool) insertTx(tx types.Tx) (uint32, error) {
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
func (m *StatelessMempool) TxStream(ctx context.Context, capacity int) <-chan types.Tx {
	ch := make(chan types.Tx, capacity)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				m.logger.Error("panic in StatelessMempool.reapTxs", "panic", p)
			}
		}()

		m.reapTxs(ctx, ch)
		close(ch)
	}()

	return ch
}

func (m *StatelessMempool) reapTxs(ctx context.Context, channel chan<- types.Tx) {
	req := &abci.RequestReapTxs{
		MaxBytes: reapMaxBytes,
		MaxGas:   reapMaxGas,
	}

	ticker := time.NewTicker(reapInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Debug("StatelessMempool.reapTxs: context done")
			return
		case <-ticker.C:
			// no "already reaped" logic here as the app is expected to handle it
			res, err := m.app.ReapTxs(ctx, req)
			if err != nil {
				m.logger.Error("StatelessMempool.reapTxs: error reaping txs", "error", err)
				continue
			}

			for _, tx := range res.Txs {
				channel <- tx
			}
		}
	}
}

// FlushAppConn flushes app client (copied from CListMempool)
func (m *StatelessMempool) FlushAppConn() error {
	err := m.app.Flush(m.ctx)
	if err != nil {
		return ErrFlushAppConn{Err: err}
	}

	return nil
}

// CheckTx returns an error on purpose for explicitness
func (m *StatelessMempool) CheckTx(_ types.Tx, _ func(*abci.ResponseCheckTx), _ TxInfo) error {
	return ErrNotImplemented
}

// Update does nothing for a stateless mempool
func (m *StatelessMempool) Update(_ int64, _ types.Txs, _ []*abci.ExecTxResult, _ PreCheckFunc, _ PostCheckFunc) error {
	return nil
}

// reading from this channel will block forever, which is fine for a stateless mempool
func (m *StatelessMempool) TxsAvailable() <-chan struct{} { return nil }
func (m *StatelessMempool) EnableTxsAvailable()           {}

func (m *StatelessMempool) Size() int        { return 0 }
func (m *StatelessMempool) SizeBytes() int64 { return 0 }

func (m *StatelessMempool) ReapMaxBytesMaxGas(_, _ int64) types.Txs { return nil }
func (m *StatelessMempool) ReapMaxTxs(_ int) types.Txs              { return nil }
func (m *StatelessMempool) RemoveTxByKey(_ types.TxKey) error       { return nil }
func (m *StatelessMempool) Flush()                                  {}

func (m *StatelessMempool) Lock()   {}
func (m *StatelessMempool) Unlock() {}
