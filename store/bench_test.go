package store

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// BenchmarkRepeatedLoadSeenCommitSameBlock benchmarks the performance of repeatedly 
// loading the same seen commit for a block.
func BenchmarkRepeatedLoadSeenCommitSameBlock(b *testing.B) {
	state, bs, _, _, cleanup, _ := makeStateAndBlockStoreAndIndexers()
	defer cleanup()
	h := bs.Height() + 1
	block := state.MakeBlock(h, test.MakeNTxs(h, 10), new(types.Commit), nil, state.Validators.GetProposer().Address)
	seenCommit := makeTestExtCommitWithNumSigs(block.Header.Height, cmttime.Now(), 100).ToCommit()
	ps, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(b, err)
	bs.SaveBlock(block, ps, seenCommit)

	// sanity check
	res := bs.LoadSeenCommit(block.Height)
	require.Equal(b, seenCommit, res)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res := bs.LoadSeenCommit(block.Height)
		require.NotNil(b, res)
	}
}
