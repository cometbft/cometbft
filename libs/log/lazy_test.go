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
		hash := txs[i].Hash()
		lazyHash := log.NewLazyHash(txs[i])
		if lazyHash.String() != hash.String() {
			t.Fatalf("expected %s, got %s", hash.String(), lazyHash.String())
		}
		if len(hash) <= 0 {
			t.Fatalf("expected non-empty hash, got empty hash")
		}
	}
}
