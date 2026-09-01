package store

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// TestDeleteLatestBlockEvictsExtendedCommitCache pins that after a rollback
// via DeleteLatestBlock and a resync that writes a *different* block at the
// same height, LoadBlockExtendedCommit returns the fresh, post-resync
// extended commit rather than a stale value served straight from the LRU
// commit caches DeleteLatestBlock never evicted.
func TestDeleteLatestBlockEvictsExtendedCommitCache(t *testing.T) {
	state, bs, cleanup := makeStateAndBlockStore()
	defer cleanup()

	h := bs.Height() + 1
	block, err := state.MakeBlock(h, test.MakeNTxs(h, 1), new(types.Commit), nil, state.Validators.GetProposer().Address)
	require.NoError(t, err)
	originalExtCommit := makeTestExtCommit(h, cmttime.Now())
	ps, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	bs.SaveBlockWithExtendedCommit(block, ps, originalExtCommit)

	// Populate the LRU cache with the pre-rollback extended commit.
	loaded := bs.LoadBlockExtendedCommit(h)
	require.Equal(t, originalExtCommit, loaded)

	require.NoError(t, bs.DeleteLatestBlock())
	require.Equal(t, h-1, bs.Height())

	// Resync: a fresh (necessarily different, since makeTestExtCommit uses
	// random bytes for hash/signatures) block lands at the same height.
	block2, err := state.MakeBlock(h, test.MakeNTxs(h, 1), new(types.Commit), nil, state.Validators.GetProposer().Address)
	require.NoError(t, err)
	freshExtCommit := makeTestExtCommit(h, cmttime.Now())
	require.NotEqual(t, originalExtCommit, freshExtCommit, "test fixture must produce a distinguishable extended commit")
	ps2, err := block2.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	bs.SaveBlockWithExtendedCommit(block2, ps2, freshExtCommit)

	got := bs.LoadBlockExtendedCommit(h)
	require.Equal(t, freshExtCommit, got,
		"LoadBlockExtendedCommit must return the post-resync extended commit, not a stale pre-rollback value served from cache")
}
