package mempool

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	abcimock "github.com/cometbft/cometbft/abci/client/mocks"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAppMempool(t *testing.T) {
	tx := func(v string) types.Tx { return types.Tx(v) }

	t.Run("InsertTx", func(t *testing.T) {
		// ARRANGE
		added := atomic.Uint64{}

		// Given app
		app := newMockAppMempoolClient(t)
		app.
			On("InsertTx", mock.Anything, mock.Anything).
			Return(func(_ context.Context, req *abci.RequestInsertTx) (*abci.ResponseInsertTx, error) {
				if string(req.Tx) == "fail" {
					t.Logf("returning retryable error")
					return &abci.ResponseInsertTx{Code: abci.CodeTypeRetry}, nil
				}

				added.Add(1)
				return &abci.ResponseInsertTx{Code: abci.CodeTypeOK}, nil
			})
		app.
			On("CheckTx", mock.Anything, mock.Anything).
			Return(func(_ context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
				if string(req.Tx) == "fail" {
					return &abci.ResponseCheckTx{Code: abci.CodeTypeRetry}, nil
				}
				return &abci.ResponseCheckTx{Code: abci.CodeTypeOK}, nil
			})

		// Given mempool
		m := NewAppMempool(config.DefaultMempoolConfig(), app)

		// Given txs
		txs := []types.Tx{tx("tx1"), tx("tx2"), tx(""), tx("fail")}

		// ACT
		err1 := m.InsertTx(txs[0])
		err2 := m.InsertTx(txs[1])
		err3 := m.InsertTx(txs[0]) // seen tx
		err4 := m.InsertTx(txs[2]) // empty tx
		err5 := m.InsertTx(txs[3]) // retryable error

		// ASSERT
		require.NoError(t, err1)
		require.NoError(t, err2)

		require.ErrorIs(t, err3, ErrSeenTx)
		require.ErrorIs(t, err4, ErrEmptyTx)

		require.ErrorContains(t, err5, "invalid code: (code=32000)")
		require.False(t, m.seen.Has(txs[3]), "should be removed from seen cache")

		require.Equal(t, uint64(2), added.Load())

		t.Run("CheckTx", func(t *testing.T) {
			for _, tt := range []struct {
				name        string
				tx          types.Tx
				errContains string
				noCallback  bool
				assert      func(t *testing.T, res *abci.ResponseCheckTx)
			}{
				{
					name:        "seen",
					tx:          tx("tx1"),
					errContains: "already seen",
				},
				{
					name: "fail",
					tx:   tx("fail"),
					assert: func(t *testing.T, res *abci.ResponseCheckTx) {
						require.Equal(t, abci.CodeTypeRetry, res.Code)
					},
				},
				{
					name: "ok",
					tx:   tx("ok"),
					assert: func(t *testing.T, res *abci.ResponseCheckTx) {
						require.Equal(t, abci.CodeTypeOK, res.Code)
					},
				},
				{
					name: "ok-no-callback",
					tx:   tx("ok2"),
				},
			} {
				t.Run(tt.name, func(t *testing.T) {
					// ARRANGE
					var (
						result       = atomic.Pointer[abci.ResponseCheckTx]{}
						callback     = func(res *abci.ResponseCheckTx) { result.Store(res) }
						ensureResult = func() bool { return result.Load() != nil }
					)

					if tt.noCallback {
						callback = nil
					}

					// ACT
					err := m.CheckTx(tt.tx, callback, TxInfo{})

					// ASSERT
					if tt.errContains != "" {
						require.ErrorContains(t, err, tt.errContains)
						return
					}

					require.NoError(t, err)
					require.Eventually(t, ensureResult, time.Second, time.Millisecond*50)

					if tt.assert != nil {
						tt.assert(t, result.Load())
					}
				})
			}
		})
	})

	t.Run("TxStream", func(t *testing.T) {
		// ARRANGE
		const amount = 100
		const callsToCancel = 4

		// Given context
		ctx, cancel := context.WithCancel(context.Background())
		calls := atomic.Uint64{}

		// Given app
		allMempoolTxs := [][]byte{}

		app := newMockAppMempoolClient(t)
		app.
			On("ReapTxs", mock.Anything, mock.Anything).
			Return(func(_ context.Context, _ *abci.RequestReapTxs) (*abci.ResponseReapTxs, error) {
				txs := make([][]byte, 0, amount)
				for i := 0; i < amount; i++ {
					txs = append(txs, []byte(fmt.Sprintf("tx-%d", i)))
				}

				allMempoolTxs = append(allMempoolTxs, txs...)

				calls.Add(1)
				if calls.Load() == callsToCancel {
					cancel()
				}

				return &abci.ResponseReapTxs{Txs: txs}, nil
			})

		// Given mempool
		m := NewAppMempool(config.DefaultMempoolConfig(), app)

		// ACT
		// stream txs from app
		sink := [][]byte{}
		ch := m.TxStream(ctx)

		for txs := range ch {
			sink = append(sink, txs.ToSliceOfBytes()...)
		}

		require.Subset(t, allMempoolTxs, sink)
	})
}

func TestAppMempool_UsesConfigValues(t *testing.T) {
	t.Run("ReapTxs receives MaxBytes and MaxGas from config", func(t *testing.T) {
		cfg := config.TestMempoolConfig()
		cfg.Type = config.MempoolTypeApp
		cfg.ReapMaxBytes = 12345
		cfg.ReapMaxGas = 67890
		cfg.ReapInterval = 10 * time.Millisecond // short for fast test

		var receivedReq *abci.RequestReapTxs
		app := newMockAppMempoolClient(t)
		app.On("ReapTxs", mock.Anything, mock.MatchedBy(func(req *abci.RequestReapTxs) bool {
			receivedReq = req
			return true
		})).Return(&abci.ResponseReapTxs{Txs: [][]byte{[]byte("tx1")}}, nil).Once()

		m := NewAppMempool(cfg, app)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ch := m.TxStream(ctx)
		// Consume one batch to trigger ReapTxs call
		batch := <-ch
		require.Len(t, batch, 1)
		cancel()

		require.NotNil(t, receivedReq)
		require.Equal(t, uint64(12345), receivedReq.MaxBytes)
		require.Equal(t, uint64(67890), receivedReq.MaxGas)
	})

	t.Run("SeenCacheSize limits LRU cache", func(t *testing.T) {
		cfg := config.TestMempoolConfig()
		cfg.Type = config.MempoolTypeApp
		cfg.SeenCacheSize = 3 // small cache to test eviction

		app := newMockAppMempoolClient(t)
		app.On("InsertTx", mock.Anything, mock.Anything).
			Return(&abci.ResponseInsertTx{Code: abci.CodeTypeOK}, nil)
		app.On("CheckTx", mock.Anything, mock.Anything).
			Return(&abci.ResponseCheckTx{Code: abci.CodeTypeOK}, nil).Maybe() // not called by InsertTx

		m := NewAppMempool(cfg, app)

		// Fill cache with 3 txs
		require.NoError(t, m.InsertTx(types.Tx("tx1")))
		require.NoError(t, m.InsertTx(types.Tx("tx2")))
		require.NoError(t, m.InsertTx(types.Tx("tx3")))

		// tx1 is seen - reject
		require.ErrorIs(t, m.InsertTx(types.Tx("tx1")), ErrSeenTx)

		// Insert 3 more to evict tx1, tx2, tx3 from LRU
		require.NoError(t, m.InsertTx(types.Tx("tx4")))
		require.NoError(t, m.InsertTx(types.Tx("tx5")))
		require.NoError(t, m.InsertTx(types.Tx("tx6")))

		// tx1 was evicted - should succeed now
		require.NoError(t, m.InsertTx(types.Tx("tx1")))
	})
}

// mockAppMempoolClient wraps abcimock.Client to implement AppMempoolClient
type mockAppMempoolClient struct {
	*abcimock.Client
}

func (m *mockAppMempoolClient) CheckTxUnlocked(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return m.CheckTx(ctx, req)
}

func newMockAppMempoolClient(t *testing.T) *mockAppMempoolClient {
	return &mockAppMempoolClient{Client: abcimock.NewClient(t)}
}
