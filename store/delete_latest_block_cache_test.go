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

// TestDeleteLatestBlockDeletesExtendedCommitRow pins the DB-level half of
// the fix directly, independent of caching or a resync: after
// DeleteLatestBlock, the "EC:" row for the deleted height must actually be
// gone from the underlying DB, not merely shadowed by a cache eviction that
// a subsequent SaveBlockWithExtendedCommit's overwrite would have hidden.
func TestDeleteLatestBlockDeletesExtendedCommitRow(t *testing.T) {
	state, bs, cleanup := makeStateAndBlockStore()
	defer cleanup()

	h := bs.Height() + 1
	block, err := state.MakeBlock(h, test.MakeNTxs(h, 1), new(types.Commit), nil, state.Validators.GetProposer().Address)
	require.NoError(t, err)
	extCommit := makeTestExtCommit(h, cmttime.Now())
	ps, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	bs.SaveBlockWithExtendedCommit(block, ps, extCommit)

	bz, err := bs.db.Get(calcExtCommitKey(h))
	require.NoError(t, err)
	require.NotEmpty(t, bz, "extended commit row must exist right after saving")

	require.NoError(t, bs.DeleteLatestBlock())

	bz, err = bs.db.Get(calcExtCommitKey(h))
	require.NoError(t, err)
	require.Empty(t, bz, "DeleteLatestBlock must delete the extended commit row from the DB, not just evict it from cache")
}

// TestDeleteLatestBlockEvictsCommitAndSeenCommitCaches extends the cache
// coverage to LoadBlockCommit and LoadSeenCommit, the two other caches
// DeleteLatestBlock now evicts alongside the extended commit cache.
func TestDeleteLatestBlockEvictsCommitAndSeenCommitCaches(t *testing.T) {
	state, bs, cleanup := makeStateAndBlockStore()
	defer cleanup()

	h := bs.Height() + 1
	block, err := state.MakeBlock(h, test.MakeNTxs(h, 1), new(types.Commit), nil, state.Validators.GetProposer().Address)
	require.NoError(t, err)
	originalSeenCommit := makeTestExtCommit(h, cmttime.Now()).ToCommit()
	ps, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	bs.SaveBlock(block, ps, originalSeenCommit)

	// Populate both caches with the pre-rollback commit data. LoadBlockCommit
	// reads the commit persisted as the *next* block's LastCommit, which
	// SaveBlock doesn't write for the block being saved -- only LoadSeenCommit
	// is directly exercisable here without an additional block at h+1, so
	// this focuses on seenCommitCache; the extended-commit cache is already
	// covered above.
	loadedSeen := bs.LoadSeenCommit(h)
	require.Equal(t, originalSeenCommit, loadedSeen)

	require.NoError(t, bs.DeleteLatestBlock())

	block2, err := state.MakeBlock(h, test.MakeNTxs(h, 1), new(types.Commit), nil, state.Validators.GetProposer().Address)
	require.NoError(t, err)
	freshSeenCommit := makeTestExtCommit(h, cmttime.Now()).ToCommit()
	require.NotEqual(t, originalSeenCommit, freshSeenCommit, "test fixture must produce a distinguishable commit")
	ps2, err := block2.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(t, err)
	bs.SaveBlock(block2, ps2, freshSeenCommit)

	got := bs.LoadSeenCommit(h)
	require.Equal(t, freshSeenCommit, got,
		"LoadSeenCommit must return the post-resync commit, not a stale pre-rollback value served from cache")
}
