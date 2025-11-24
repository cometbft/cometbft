package mempool

import (
	"context"
	"sync/atomic"
	"testing"

	abcimock "github.com/cometbft/cometbft/abci/client/mocks"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatelessMempool(t *testing.T) {
	tx := func(v string) types.Tx { return types.Tx(v) }

	t.Run("InsertTx", func(t *testing.T) {
		// ARRANGE
		added := atomic.Uint64{}

		app := abcimock.NewClient(t)
		app.
			On("InsertTx", mock.Anything, mock.Anything).
			Return(func(_ context.Context, _ *abci.RequestInsertTx) (*abci.ResponseInsertTx, error) {
				added.Add(1)
				return &abci.ResponseInsertTx{Code: abci.CodeTypeOK}, nil
			})

		m := NewStatelessMempool(config.DefaultMempoolConfig(), app)

		tx1, tx2, tx3 := tx("tx1"), tx("tx2"), tx("")

		// ACT
		err1 := m.InsertTx(tx1)
		err2 := m.InsertTx(tx2)
		err3 := m.InsertTx(tx1) // seen tx
		err4 := m.InsertTx(tx3) // empty tx

		// ASSERT
		require.NoError(t, err1)
		require.NoError(t, err2)
		require.ErrorIs(t, err3, ErrSeenTx)
		require.ErrorIs(t, err4, ErrEmptyTx)
		require.Equal(t, uint64(2), added.Load())
	})
}
