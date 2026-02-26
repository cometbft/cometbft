package consensus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
)

func TestStateIngestVerifiedBlock(t *testing.T) {
	t.Run("ingestedBlock", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block
		ic := ts.MakeIngestCandidate()

		// ACT
		err := ts.IngestVerifiedBlock(ic)

		// ASSERT
		require.NoError(t, err)

		assert.Equal(t, ic.Height(), ts.cs.GetLastHeight())
		assert.NotNil(t, ts.cs.blockStore.LoadBlock(ic.Height()))
	})

	t.Run("alreadyIncluded", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block
		vb := ts.MakeIngestCandidate()

		// That was already ingested
		err := ts.IngestVerifiedBlock(vb)
		require.NoError(t, err)

		// ACT
		// Ingest it again
		err = ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.ErrorIs(t, err, ErrAlreadyIncluded)
	})

	t.Run("heightGap", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given block that is not the next height
		vb := ts.MakeIngestCandidate()
		vb.block.Height++

		// ACT
		err := ts.IngestVerifiedBlock(vb)

		// ASSERT
		require.ErrorIs(t, err, ErrHeightGap)
	})

	t.Run("invalidVerifiedBlock", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given a verified block with an invalid block
		ic := ts.MakeIngestCandidate()
		validBlock := ic.block

		// copy block to drop hash cache and trigger validation
		ic.verified = false
		ic.block = &types.Block{
			Header:     validBlock.Header,
			Data:       validBlock.Data,
			Evidence:   validBlock.Evidence,
			LastCommit: nil, // invalid last commit
		}

		// ACT
		err := ts.IngestVerifiedBlock(ic)

		// ASSERT
		require.ErrorIs(t, err, ErrValidation)
		require.ErrorContains(t, err, "unverified ingest candidate")
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

func (ts *ingestTestSuite) IngestVerifiedBlock(vb IngestCandidate) error {
	ts.t.Helper()

	ts.cs.mtx.Lock()
	defer ts.cs.mtx.Unlock()

	return ts.cs.handleIngestVerifiedBlock(vb)
}

func (ts *ingestTestSuite) MakeIngestCandidate() IngestCandidate {
	ts.t.Helper()

	block, err := ts.cs.createProposalBlock(context.Background())
	require.NoError(ts.t, err)

	blockParts, err := block.MakePartSet(types.BlockPartSizeBytes)
	require.NoError(ts.t, err)

	privVals := make([]types.PrivValidator, len(ts.validators))
	for i, vs := range ts.validators {
		privVals[i] = vs.PrivValidator
	}

	var (
		extensionsEnabled = ts.cs.state.ConsensusParams.ABCI.VoteExtensionsEnabled(block.Height)
		chainID           = ts.cs.state.ChainID
		blockHeight       = block.Height
		blockID           = types.BlockID{
			Hash:          block.Hash(),
			PartSetHeader: blockParts.Header(),
		}
	)

	var voteSet *types.VoteSet
	if extensionsEnabled {
		voteSet = types.NewExtendedVoteSet(chainID, blockHeight, 0, cmtproto.PrecommitType, ts.cs.Validators)
	} else {
		voteSet = types.NewVoteSet(chainID, blockHeight, 0, cmtproto.PrecommitType, ts.cs.Validators)
	}

	for i := 0; i < len(privVals); i++ {
		ts.validators[i].Height = blockHeight
		ts.validators[i].Round = 0
		vote := signVote(ts.validators[i], cmtproto.PrecommitType, blockID.Hash, blockID.PartSetHeader, extensionsEnabled)
		added, err := voteSet.AddVote(vote)
		require.NoError(ts.t, err)
		require.True(ts.t, added)
	}

	extCommit := voteSet.MakeExtendedCommit(ts.cs.state.ConsensusParams.ABCI)
	commit := extCommit.ToCommit()
	if !extensionsEnabled {
		extCommit = nil
	}

	ic, err := NewIngestCandidate(block, blockParts, commit, extCommit)
	require.NoError(ts.t, err, "failed to create ingest candidate")
	require.NoError(ts.t, ic.Verify(ts.cs.state), "failed to verify ingest candidate")

	return ic
}
