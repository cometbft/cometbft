package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cometbft/cometbft/internal/test"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// auditingDB wraps a dbm.DB and, after every batch Write(), records the
// durable BlockStoreState.Base and whether the block meta for that base
// height is still present. This lets a test observe intermediate durable
// state during a single PruneBlocks call, not just the state after it
// returns.
type auditingDB struct {
	dbm.DB
	onWrite func(db dbm.DB)
}

func (a *auditingDB) NewBatch() dbm.Batch {
	return &auditingBatch{Batch: a.DB.NewBatch(), audit: a}
}

type auditingBatch struct {
	dbm.Batch
	audit *auditingDB
}

func (b *auditingBatch) Write() error {
	if err := b.Batch.Write(); err != nil {
		return err
	}
	if b.audit.onWrite != nil {
		b.audit.onWrite(b.audit.DB)
	}
	return nil
}

func (b *auditingBatch) WriteSync() error {
	if err := b.Batch.WriteSync(); err != nil {
		return err
	}
	if b.audit.onWrite != nil {
		b.audit.onWrite(b.audit.DB)
	}
	return nil
}

// TestPruneBlocksIntermediateBaseNeverPointsAtDeletedBlock is a regression
// test for a bug where PruneBlocks' periodic checkpoint flush persisted
// BlockStoreState.Base as the height JUST deleted in that iteration, instead
// of one past it. On unfixed code this durably records a Base whose block
// meta has already been deleted from the same batch, which corrupts
// LoadBaseMeta/Size and can become permanent if PruneBlocks errors before
// its final flush.
func TestPruneBlocksIntermediateBaseNeverPointsAtDeletedBlock(t *testing.T) {
	config := test.ResetTestRoot("blockchain_reactor_test")
	t.Cleanup(func() { _ = config.RootDir })

	stateStore := sm.NewStore(dbm.NewMemDB(), sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisFile(config.GenesisFile())
	require.NoError(t, err)

	underlying := dbm.NewMemDB()
	var violations []string
	adb := &auditingDB{DB: underlying}
	adb.onWrite = func(db dbm.DB) {
		bsState := LoadBlockStoreState(db)
		base := bsState.Base
		if base <= 0 {
			return
		}
		metaBytes, err := db.Get(calcBlockMetaKey(base))
		require.NoError(t, err)
		if len(metaBytes) == 0 {
			violations = append(violations, fmt.Sprintf(
				"durable Base=%d but block meta for that height was already deleted", base))
		}
	}

	bs := NewBlockStore(adb)

	// Mirror TestPruneBlocks' proven setup: with this timestamp/evidence-age
	// configuration, evidenceRetainHeight lands at 1100, so blocks below it
	// (including the pruned%1000==0 checkpoint at h=1000) really do have
	// their meta deleted, letting this test actually exercise the
	// intermediate-flush bug rather than vacuously passing because nothing
	// was deleted yet.
	const total = int64(1500)
	for h := int64(1); h <= total; h++ {
		block, err := state.MakeBlock(h, test.MakeNTxs(h, 2), new(types.Commit), nil, state.Validators.GetProposer().Address)
		require.NoError(t, err)
		partSet, err := block.MakePartSet(types.BlockPartSizeBytes)
		require.NoError(t, err)
		seenCommit := makeTestExtCommit(h, cmttime.Now())
		bs.SaveBlockWithExtendedCommit(block, partSet, seenCommit)
	}

	state.LastBlockTime = time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)
	state.LastBlockHeight = total
	state.ConsensusParams.Evidence.MaxAgeNumBlocks = 400
	state.ConsensusParams.Evidence.MaxAgeDuration = 1 * time.Second

	const pruneTo = int64(1200)
	pruned, evidenceRetainHeight, err := bs.PruneBlocks(pruneTo, state)
	require.NoError(t, err)
	require.EqualValues(t, pruneTo-1, pruned)
	// Sanity: the checkpoint at h=1000 must fall below the evidence retain
	// point, or its meta was never deleted and this test would pass
	// vacuously.
	require.Greater(t, evidenceRetainHeight, int64(1000))

	require.Empty(t, violations, "intermediate checkpoint(s) recorded a durable Base pointing at an already-deleted block:\n%v", violations)

	// Sanity: final state is still correct either way.
	assert.EqualValues(t, pruneTo, bs.Base())
}
