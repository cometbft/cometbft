package consensus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
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
		ic := ts.MakeIngestCandidate()

		// That was already ingested
		err := ts.IngestVerifiedBlock(ic)
		require.NoError(t, err)

		// ACT
		// Ingest it again
		err = ts.IngestVerifiedBlock(ic)

		// ASSERT
		require.ErrorIs(t, err, ErrAlreadyIncluded)
	})

	t.Run("heightGap", func(t *testing.T) {
		// ARRANGE
		ts := newIngestTestSuite(t)

		// Given block that is not the next height
		ic := ts.MakeIngestCandidate()
		ic.block.Height++

		// ACT
		err := ts.IngestVerifiedBlock(ic)

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

func TestIngestCandidate(t *testing.T) {
	t.Run("ValidateBasic", func(t *testing.T) {
		ts := newIngestTestSuite(t)

		for _, tt := range []struct {
			name        string
			mutate      func(ic *IngestCandidate)
			errContains string
		}{
			{
				name:   "valid candidate",
				mutate: nil,
			},
			{
				name: "nil block",
				mutate: func(ic *IngestCandidate) {
					ic.block = nil
				},
				errContains: "block is nil",
			},
			{
				name: "nil part set",
				mutate: func(ic *IngestCandidate) {
					ic.blockParts = nil
				},
				errContains: "part set is nil",
			},
			{
				name: "nil commit",
				mutate: func(ic *IngestCandidate) {
					ic.commit = nil
				},
				errContains: "commit is nil",
			},
			{
				name: "commit height mismatch",
				mutate: func(ic *IngestCandidate) {
					ic.extCommit = nil
					ic.commit.Height = ic.block.Height + 1
				},
				errContains: "commit height mismatch",
			},
			{
				name: "commit blockID mismatch",
				mutate: func(ic *IngestCandidate) {
					ic.extCommit = nil
					ic.commit.BlockID = types.BlockID{}
				},
				errContains: "commit blockID mismatch",
			},
			{
				name: "extended commit height mismatch",
				mutate: func(ic *IngestCandidate) {
					ic.extCommit = &types.ExtendedCommit{
						Height: ic.block.Height + 1,
					}
				},
				errContains: "extCommit height mismatch",
			},
			{
				name: "extended commit blockID mismatch",
				mutate: func(ic *IngestCandidate) {
					ic.extCommit = &types.ExtendedCommit{
						Height:  ic.block.Height,
						BlockID: types.BlockID{},
					}
				},
				errContains: "extended commit blockID mismatch",
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// ARRANGE
				ic := ts.MakeIngestCandidate()

				if tt.mutate != nil {
					tt.mutate(&ic)
				}

				// ACT
				err := ic.ValidateBasic()

				// ASSERT
				if tt.errContains == "" {
					require.NoError(t, err)
					return
				}

				require.ErrorIs(t, err, ErrValidation)
				require.ErrorContains(t, err, tt.errContains)
			})
		}
	})

	t.Run("Verify", func(t *testing.T) {
		for _, tt := range []struct {
			name           string
			voteExtensions bool
			mutate         func(ic *IngestCandidate, st *sm.State)
			errContains    string
		}{
			{
				name:           "valid candidate",
				voteExtensions: false,
				mutate:         nil,
			},
			{
				name:           "valid candidate with vote extensions",
				voteExtensions: true,
				mutate:         nil,
			},
			{
				name:           "extensions invariant mismatch",
				voteExtensions: true,
				mutate: func(ic *IngestCandidate, st *sm.State) {
					st.ConsensusParams.ABCI.VoteExtensionsEnableHeight = 0
				},
				errContains: "invalid ext commit state",
			},
			{
				name:           "invalid block",
				voteExtensions: false,
				mutate: func(ic *IngestCandidate, st *sm.State) {
					validBlock := ic.block
					ic.block = &types.Block{
						Header:     validBlock.Header,
						Data:       validBlock.Data,
						Evidence:   validBlock.Evidence,
						LastCommit: nil,
					}
				},
				errContains: "validate block",
			},
			{
				name:           "commit verification fails",
				voteExtensions: false,
				mutate: func(ic *IngestCandidate, st *sm.State) {
					ic.commit.Signatures[0].Signature = nil
				},
				errContains: "verify commit",
			},
			{
				name:           "extended commit missing extension signature",
				voteExtensions: true,
				mutate: func(ic *IngestCandidate, st *sm.State) {
					ic.extCommit.ExtendedSignatures[0].ExtensionSignature = nil
				},
				errContains: "ensure extensions",
			},
			{
				name:           "extended commit verification fails",
				voteExtensions: true,
				mutate: func(ic *IngestCandidate, st *sm.State) {
					ic.extCommit.ExtendedSignatures[0].Signature = nil
				},
				errContains: "verify extended commit",
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// ARRANGE
				ts := newIngestTestSuite(t)

				if tt.voteExtensions {
					ts.cs.state.ConsensusParams.ABCI.VoteExtensionsEnableHeight = 1
				} else {
					ts.cs.state.ConsensusParams.ABCI.VoteExtensionsEnableHeight = 0
				}

				// Given a valid ingest candidate
				ic := ts.MakeIngestCandidate()
				if tt.voteExtensions {
					require.NotNil(t, ic.extCommit)
				} else {
					require.Nil(t, ic.extCommit)
				}

				// with verification disabled
				ic.verified = false
				verifyState := ts.cs.state

				if tt.mutate != nil {
					tt.mutate(&ic, &verifyState)
				}

				// ACT
				err := ic.Verify(verifyState)

				// ASSERT
				if tt.errContains == "" {
					require.NoError(t, err)
					require.True(t, ic.verified)
					return
				}

				require.ErrorContains(t, err, tt.errContains)
				require.False(t, ic.verified)
			})
		}
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

func (ts *ingestTestSuite) IngestVerifiedBlock(ic IngestCandidate) error {
	ts.t.Helper()

	ts.cs.mtx.Lock()
	defer ts.cs.mtx.Unlock()

	return ts.cs.ingestBlock(ic)
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
