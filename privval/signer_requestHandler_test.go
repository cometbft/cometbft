package privval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	privvalproto "github.com/cometbft/cometbft/proto/tendermint/privval"
	cmterrors "github.com/cometbft/cometbft/types/errors"
)

// A SignVoteRequest without a vote must be rejected with an error instead of
// panicking in the signer (see #6093).
func TestDefaultValidationRequestHandler_MissingVote(t *testing.T) {
	const chainID = "mychainid"
	privVal, _, _ := newTestFilePV(t)

	req := mustWrapMsg(&privvalproto.SignVoteRequest{ChainId: chainID, Vote: nil})

	res, err := DefaultValidationRequestHandler(privVal, req, chainID)
	require.Error(t, err)
	assert.ErrorIs(t, err, cmterrors.ErrRequiredField{Field: "vote"})

	resp := res.GetSignedVoteResponse()
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error)
	assert.Equal(t, cmterrors.ErrRequiredField{Field: "vote"}.Error(), resp.Error.Description)
}

// A SignProposalRequest without a proposal must be rejected with an error
// instead of panicking in the signer (see #6093).
func TestDefaultValidationRequestHandler_MissingProposal(t *testing.T) {
	const chainID = "mychainid"
	privVal, _, _ := newTestFilePV(t)

	req := mustWrapMsg(&privvalproto.SignProposalRequest{ChainId: chainID, Proposal: nil})

	res, err := DefaultValidationRequestHandler(privVal, req, chainID)
	require.Error(t, err)
	assert.ErrorIs(t, err, cmterrors.ErrRequiredField{Field: "proposal"})

	resp := res.GetSignedProposalResponse()
	require.NotNil(t, resp)
	require.NotNil(t, resp.Error)
	assert.Equal(t, cmterrors.ErrRequiredField{Field: "proposal"}.Error(), resp.Error.Description)
}
