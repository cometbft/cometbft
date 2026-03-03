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
		app := abcimock.NewClient(t)
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

	t.Run("CheckTx", func(t *testing.T) {
		for _, tt := range []struct {
			name         string
			tx           types.Tx
			mockResponse *abci.ResponseCheckTx
			mockErr      error
			shouldRemove bool
		}{
			{
				name:         "happy path - tx stays in seen",
				tx:           tx("happy"),
				mockResponse: &abci.ResponseCheckTx{Code: abci.CodeTypeOK},
				shouldRemove: false,
			},
			{
				name:         "error - tx removed from seen",
				tx:           tx("error"),
				mockErr:      fmt.Errorf("some error"),
				shouldRemove: true,
			},
			{
				name:         "non-ok code - tx removed from seen",
				tx:           tx("badcode"),
				mockResponse: &abci.ResponseCheckTx{Code: 1},
				shouldRemove: true,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				mockDone := make(chan struct{})
				callbackDone := make(chan struct{})

				app := abcimock.NewClient(t)
				app.On("CheckTx", mock.Anything, mock.Anything).
					Return(func(context.Context, *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
						defer close(mockDone)
						return tt.mockResponse, tt.mockErr
					}).Once()

				m := NewAppMempool(config.DefaultMempoolConfig(), app)

				callback := func(*abci.ResponseCheckTx) {
					close(callbackDone)
				}

				err := m.CheckTx(tt.tx, callback, TxInfo{})
				require.NoError(t, err)
				require.True(t, m.seen.Has(tt.tx))
				if tt.mockErr != nil {
					// error case: callback not called
					// after mock returns, seen.Remove runs sync
					// verify removal by attempting CheckTx again. should be able to run without error.
					<-mockDone
					retryDone := make(chan struct{})
					app.On("CheckTx", mock.Anything, mock.Anything).
						Return(&abci.ResponseCheckTx{Code: abci.CodeTypeOK}, nil).Once()
					err = m.CheckTx(tt.tx, func(*abci.ResponseCheckTx) { close(retryDone) }, TxInfo{})
					require.NoError(t, err, "tx should not be seen after error removal")
					<-retryDone
				} else {
					// no error cases: callback is called after any removal
					<-callbackDone
					if tt.shouldRemove {
						require.False(t, m.seen.Has(tt.tx), "tx should be removed from seen cache")
					} else {
						require.True(t, m.seen.Has(tt.tx), "tx should stay in seen cache")
					}
				}
			})
		}
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

		app := abcimock.NewClient(t)
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
