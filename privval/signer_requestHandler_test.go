package privval

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	privvalproto "github.com/cometbft/cometbft/proto/tendermint/privval"
	cmterrors "github.com/cometbft/cometbft/types/errors"
)

func TestDefaultValidationRequestHandlerMissingVote(t *testing.T) {
	dir := t.TempDir()
	privVal := GenFilePV(filepath.Join(dir, "key.json"), filepath.Join(dir, "state.json"))
	req := mustWrapMsg(&privvalproto.SignVoteRequest{ChainId: "test-chain"})
	res, err := DefaultValidationRequestHandler(privVal, req, "test-chain")
	require.ErrorIs(t, err, cmterrors.ErrRequiredField{Field: "vote"})

	response := res.GetSignedVoteResponse()
	require.NotNil(t, response)
	require.Equal(t, &privvalproto.RemoteSignerError{Code: 0, Description: err.Error()}, response.Error)
	require.Empty(t, response.Vote)
}

func TestDefaultValidationRequestHandlerMissingProposal(t *testing.T) {
	dir := t.TempDir()
	privVal := GenFilePV(filepath.Join(dir, "key.json"), filepath.Join(dir, "state.json"))
	req := mustWrapMsg(&privvalproto.SignProposalRequest{ChainId: "test-chain"})
	res, err := DefaultValidationRequestHandler(privVal, req, "test-chain")
	require.ErrorIs(t, err, cmterrors.ErrRequiredField{Field: "proposal"})

	response := res.GetSignedProposalResponse()
	require.NotNil(t, response)
	require.Equal(t, &privvalproto.RemoteSignerError{Code: 0, Description: err.Error()}, response.Error)
	require.Empty(t, response.Proposal)
}
