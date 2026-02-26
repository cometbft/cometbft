package consensus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtinternaltest "github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/types"
)

func TestStateIngestVerifiedBlock(t *testing.T) {
	t.Run("ingestedBlock", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block
		vb := ts.MakeVerifiedBlock()

		// ACT
		err, malicious := ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.NoError(t, err)

		assert.False(t, malicious)
		assert.Equal(t, vb.Block.Height, ts.cs.GetLastHeight())
		assert.NotNil(t, ts.cs.blockStore.LoadBlock(vb.Block.Height))
	})

	t.Run("alreadyIncluded", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block
		vb := ts.MakeVerifiedBlock()

		// That was already ingested
		err, _ := ts.IngestVerifiedBlock(vb)
		require.NoError(t, err)

		// ACT
		// Ingest it again
		err, malicious := ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.ErrorIs(t, err, ErrAlreadyIncluded)
		require.False(t, malicious)
	})

	t.Run("heightGap", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given block that is not the next height
		vb := ts.MakeVerifiedBlock()
		vb.Block.Height++

		// ACT
		err, malicious := ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.ErrorIs(t, err, ErrHeightGap)
		require.False(t, malicious)
	})

	t.Run("invalidVerifiedBlock", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block with an invalid block
		vb := ts.MakeVerifiedBlock()
		validBlock := vb.Block

		// copy block to drop hash cache and trigger validation
		vb.Block = &types.Block{
			Header:     validBlock.Header,
			Data:       validBlock.Data,
			Evidence:   validBlock.Evidence,
			LastCommit: nil, // invalid last commit
		}

		// ACT
		err, malicious := ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.ErrorContains(t, err, "failed to validate block: nil LastCommit")
		require.True(t, malicious)
	})
}

type ingestTestSuite struct {
	t          *testing.T
	cs         *State
	validators []*validatorStub
}

func newIngestTestSuite(t *testing.T) *ingestTestSuite {
	cs, validators := randState(4)

	return &ingestTestSuite{
		t:          t,
		cs:         cs,
		validators: validators,
	}
}

func (ts *ingestTestSuite) IngestVerifiedBlock(vb VerifiedBlock) (error, bool) {
	ts.t.Helper()

	ts.cs.mtx.Lock()
	defer ts.cs.mtx.Unlock()

	return ts.cs.handleIngestVerifiedBlock(vb)
}

func (ts *ingestTestSuite) MakeVerifiedBlock() VerifiedBlock {
	ts.t.Helper()

	block, err := ts.cs.createProposalBlock(context.Background())
	require.NoError(ts.t, err)

	blockParts, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(ts.t, err)

	privVals := make([]types.PrivValidator, len(ts.validators))
	for i, vs := range ts.validators {
		privVals[i] = vs.PrivValidator
	}

	blockID := types.BlockID{
		Hash:          block.Hash(),
		PartSetHeader: blockParts.Header(),
	}

	commit, err := cmtinternaltest.MakeCommit(
		blockID,
		block.Height,
		0,
		ts.cs.Validators,
		privVals,
		ts.cs.state.ChainID,
		block.Time,
	)
	require.NoError(ts.t, err)

	return VerifiedBlock{
		Block:      block,
		BlockParts: blockParts,
		Commit:     commit,
	}
}
