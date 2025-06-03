package privval

import (
	"fmt"

	pvproto "github.com/cometbft/cometbft/api/cometbft/privval/v2"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/cometbft/cometbft/v2/crypto"
	"github.com/cometbft/cometbft/v2/types"
)

func DefaultValidationRequestHandler(
	privVal types.PrivValidator,
	req pvproto.Message,
	chainID string,
) (pvproto.Message, error) {
	var (
		res pvproto.Message
		err error
	)

	switch r := req.Sum.(type) {
	case *pvproto.Message_PubKeyRequest:
		if r.PubKeyRequest.GetChainId() != chainID {
			return chainIDMismatchError(r.PubKeyRequest.GetChainId(), chainID)
		}

		var pubKey crypto.PubKey

		pubKey, err = privVal.GetPubKey()
		if err != nil {
			res = mustWrapMsg(&pvproto.PubKeyResponse{
				PubKeyType: "", PubKeyBytes: []byte{}, Error: &pvproto.RemoteSignerError{
					Code: 0, Description: err.Error(),
				},
			})
		} else {
			res = mustWrapMsg(&pvproto.PubKeyResponse{PubKeyType: pubKey.Type(), PubKeyBytes: pubKey.Bytes(), Error: nil})
		}
	case *pvproto.Message_SignVoteRequest:
		if r.SignVoteRequest.ChainId != chainID {
			return chainIDMismatchError(r.SignVoteRequest.GetChainId(), chainID)
		}

		vote := r.SignVoteRequest.Vote

		err = privVal.SignVote(chainID, vote, !r.SignVoteRequest.SkipExtensionSigning)
		if err != nil {
			res = mustWrapMsg(&pvproto.SignedVoteResponse{
				Vote: cmtproto.Vote{}, Error: &pvproto.RemoteSignerError{Code: 0, Description: err.Error()},
			})
		} else {
			res = mustWrapMsg(&pvproto.SignedVoteResponse{Vote: *vote, Error: nil})
		}
	case *pvproto.Message_SignProposalRequest:
		if r.SignProposalRequest.GetChainId() != chainID {
			return chainIDMismatchError(r.SignProposalRequest.GetChainId(), chainID)
		}

		proposal := r.SignProposalRequest.Proposal

		err = privVal.SignProposal(chainID, proposal)
		if err != nil {
			res = mustWrapMsg(&pvproto.SignedProposalResponse{
				Proposal: cmtproto.Proposal{}, Error: &pvproto.RemoteSignerError{Code: 0, Description: err.Error()},
			})
		} else {
			res = mustWrapMsg(&pvproto.SignedProposalResponse{Proposal: *proposal, Error: nil})
		}
	case *pvproto.Message_SignBytesRequest:
		var signature []byte

		signature, err = privVal.SignBytes(r.SignBytesRequest.Value)
		if err != nil {
			res = mustWrapMsg(&pvproto.SignBytesResponse{
				Signature: nil, Error: &pvproto.RemoteSignerError{Code: 0, Description: err.Error()},
			})
		} else {
			res = mustWrapMsg(&pvproto.SignBytesResponse{Signature: signature, Error: nil})
		}
	case *pvproto.Message_PingRequest:
		err, res = nil, mustWrapMsg(&pvproto.PingResponse{})
	default:
		err = fmt.Errorf("unknown msg: %v", r)
	}

	return res, err
}

func chainIDMismatchError(want, got string) (pvproto.Message, error) {
	res := mustWrapMsg(&pvproto.PubKeyResponse{
		PubKeyType: "", PubKeyBytes: []byte{}, Error: &pvproto.RemoteSignerError{
			Code: 0, Description: "unable to serve request",
		},
	})
	return res, fmt.Errorf("want chainID: %s, got chainID: %s", want, got)
}
