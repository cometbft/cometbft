package store

import (
	"errors"
	"fmt"
	"os"
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

	// failOnBatchN, if > 0, makes the failOnBatchN'th batch write
	// (WriteSync or Write, 1-indexed across the whole DB) return an error
	// instead of committing, simulating PruneBlocks erroring out partway
	// through, after some checkpoint flushes already landed but before the
	// final flush.
	failOnBatchN int
	batchCount   int
}

func (a *auditingDB) NewBatch() dbm.Batch {
	return &auditingBatch{Batch: a.DB.NewBatch(), audit: a}
}

type auditingBatch struct {
	dbm.Batch
	audit *auditingDB
}

func (b *auditingBatch) shouldFail() bool {
	if b.audit.failOnBatchN <= 0 {
		return false
	}
	b.audit.batchCount++
	return b.audit.batchCount == b.audit.failOnBatchN
}

func (b *auditingBatch) Write() error {
	if b.shouldFail() {
		return errors.New("injected write failure")
	}
	if err := b.Batch.Write(); err != nil {
		return err
	}
	if b.audit.onWrite != nil {
		b.audit.onWrite(b.audit.DB)
	}
	return nil
}

func (b *auditingBatch) WriteSync() error {
	if b.shouldFail() {
		return errors.New("injected write failure")
	}
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
	t.Cleanup(func() { os.RemoveAll(config.RootDir) })

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

// TestPruneBlocksCheckpointBaseSurvivesLaterError proves the "becomes
// permanent" half of the same bug: if PruneBlocks errors after an
// intermediate checkpoint flush has already landed but before the final
// flush runs, whatever base the checkpoint persisted is what the store is
// left with. On unfixed code that's a base pointing at a just-deleted block;
// with the fix it's the correct next-surviving-block height.
func TestPruneBlocksCheckpointBaseSurvivesLaterError(t *testing.T) {
	config := test.ResetTestRoot("blockchain_reactor_test")
	t.Cleanup(func() { os.RemoveAll(config.RootDir) })

	stateStore := sm.NewStore(dbm.NewMemDB(), sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisFile(config.GenesisFile())
	require.NoError(t, err)

	underlying := dbm.NewMemDB()
	adb := &auditingDB{DB: underlying}
	bs := NewBlockStore(adb)

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

	// Reset the batch counter now, so it only counts writes made during the
	// upcoming PruneBlocks call: batch #1 is the first (h=1000) checkpoint
	// flush, batch #2 would be the final flush. Failing on #2 simulates an
	// error landing after the first checkpoint durably persisted, before the
	// prune as a whole completes.
	adb.batchCount = 0
	adb.failOnBatchN = 2

	const pruneTo = int64(1200)
	_, _, err = bs.PruneBlocks(pruneTo, state)
	require.Error(t, err, "expected the injected failure on the final flush to surface")

	// The first checkpoint's base is now permanent: PruneBlocks won't be
	// retried by this test, and no later flush ever corrects it.
	persistedBase := LoadBlockStoreState(underlying).Base
	require.EqualValues(t, 1001, persistedBase, "expected the checkpoint's flush (fired while pruning height 1000) to have durably persisted base=1001, matching the final flush's own h+1-style semantics")

	metaBytes, err := underlying.Get(calcBlockMetaKey(persistedBase))
	require.NoError(t, err)
	assert.NotEmpty(t, metaBytes,
		"persisted Base=%d points at a block whose meta was already deleted -- this is the permanent corruption the fix prevents", persistedBase)
}
