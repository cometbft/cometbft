package core

import (
	"fmt"
	"net/http/httptest"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/types"
)

type txSearchReadDB struct {
	dbm.DB
	reads int
}

func (db *txSearchReadDB) Get(key []byte) ([]byte, error) {
	db.reads++
	return db.DB.Get(key)
}

func TestTxSearchRepeatedEvents(t *testing.T) {
	const descending = "desc"
	for _, n := range []int{1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			store := &txSearchReadDB{DB: dbm.NewMemDB()}
			t.Cleanup(func() { require.NoError(t, store.Close()) })
			indexer := kv.NewTxIndex(store)
			batch := txindex.NewBatch(2)
			txs := []types.Tx{[]byte("first transfer batch"), []byte("second transfer batch")}
			for i, tx := range txs {
				events := make([]abci.Event, n)
				for j := range events {
					events[j] = abci.Event{Type: "transfer", Attributes: []abci.EventAttribute{
						{Key: "recipient", Value: "alice", Index: true},
					}}
				}
				require.NoError(t, batch.Add(&abci.TxResult{
					Height: 1, Index: uint32(i), Tx: tx, Result: abci.ExecTxResult{Events: events},
				}))
			}
			require.NoError(t, indexer.AddBatch(batch))
			env := &Environment{TxIndexer: indexer}
			ctx := &rpctypes.Context{HTTPReq: httptest.NewRequest("GET", "/tx_search", nil)}
			perPage := 1
			for _, order := range []string{"asc", descending} {
				for page := 1; page <= 2; page++ {
					store.reads = 0
					result, err := env.TxSearch(ctx, "transfer.recipient = 'alice'", false, &page, &perPage, order)
					require.NoError(t, err)
					require.Equal(t, 2, result.TotalCount)
					require.Len(t, result.Txs, 1)
					index := page - 1
					if order == descending {
						index = 2 - page
					}
					require.Equal(t, txs[index], result.Txs[0].Tx)
					require.Equal(t, txs[index].Hash(), []byte(result.Txs[0].Hash))
					require.Len(t, result.Txs[0].TxResult.Events, n)
					require.Equal(t, 2, store.reads, "each distinct transaction is read once before pagination")
				}
			}
		})
	}
}
