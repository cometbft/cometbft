package state

import (
	"fmt"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)

// generateDummyTxs generates a slice of dummy transactions of a given size.
func generateDummyTxs(numTxs, txSize int) (txs []*abcitypes.ExecTxResult) {
	for i := 0; i < numTxs; i++ {
		tx := make([]byte, txSize)
		txs = append(txs, &abcitypes.ExecTxResult{Data: tx})
	}
	return txs
}

// BenchmarkSaveFinalizeBlockResponse benchmarks the SaveFinalizeBlockResponse function with different transaction sizes.
func BenchmarkSaveFinalizeBlockResponse(b *testing.B) {
	db := dbm.NewMemDB() // Using an in-memory DB for benchmarking
	store := NewDBStore(db)

	// Define different transaction sizes to test
	txSizes := []int{100, 1000, 10000, 100000, 1000000, 10000000} // in bytes
	numTxs := 10000                                               // Number of transactions to include in the block

	for _, size := range txSizes {
		b.Run(
			fmt.Sprintf("TxSize%d", size),
			func(b *testing.B) {
				txs := generateDummyTxs(numTxs, size)
				resp := &abcitypes.FinalizeBlockResponse{TxResults: txs}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := store.SaveFinalizeBlockResponse(int64(i), resp); err != nil {
						b.Fatalf("Failed to save finalize block response: %v", err)
					}
				}
			},
		)
	}
}
