package log_test

import (
	"testing"

	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
)

func TestLazyHash_Txs(t *testing.T) {
	const height = 2
	const numTxs = 5
	txs := test.MakeNTxs(height, numTxs)

	for i := 0; i < numTxs; i++ {
		lazyHash := log.NewLazyHash(txs[i])
		if lazyHash.String() != txs[i].Hash().String() {
			t.Fatalf("expected %s, got %s", txs[i].Hash().String(), lazyHash.String())
		}
		if len(lazyHash.String()) <= 0 {
			t.Fatalf("expected non-empty hash, got empty hash")
		}
	}
}
