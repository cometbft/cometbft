package privval

import (
	"fmt"

	cryptoproto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	pvproto "github.com/cometbft/cometbft/api/cometbft/privval/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/types"
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
			res = mustWrapMsg(&pvproto.PubKeyResponse{
				PubKey: cryptoproto.PublicKey{}, Error: &pvproto.RemoteSignerError{
					Code: 0, Description: "unable to provide pubkey",
				},
			})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.PubKeyRequest.GetChainId(), chainID)
		}

		var pubKey crypto.PubKey
		pubKey, err = privVal.GetPubKey()
		if err != nil {
			return res, err
		}
		pk, err := cryptoenc.PubKeyToProto(pubKey)
		if err != nil {
			return res, err
		}

		if err != nil {
			res = mustWrapMsg(&pvproto.PubKeyResponse{
				PubKey: cryptoproto.PublicKey{}, Error: &pvproto.RemoteSignerError{Code: 0, Description: err.Error()},
			})
		} else {
			res = mustWrapMsg(&pvproto.PubKeyResponse{PubKey: pk, Error: nil})
		}

	case *pvproto.Message_SignVoteRequest:
		if r.SignVoteRequest.ChainId != chainID {
			res = mustWrapMsg(&pvproto.SignedVoteResponse{
				Vote: cmtproto.Vote{}, Error: &pvproto.RemoteSignerError{
					Code: 0, Description: "unable to sign vote",
				},
			})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.SignVoteRequest.GetChainId(), chainID)
		}

		vote := r.SignVoteRequest.Vote

		err = privVal.SignVote(chainID, vote)
		if err != nil {
			res = mustWrapMsg(&pvproto.SignedVoteResponse{
				Vote: cmtproto.Vote{}, Error: &pvproto.RemoteSignerError{Code: 0, Description: err.Error()},
			})
		} else {
			res = mustWrapMsg(&pvproto.SignedVoteResponse{Vote: *vote, Error: nil})
		}

	case *pvproto.Message_SignProposalRequest:
		if r.SignProposalRequest.GetChainId() != chainID {
			res = mustWrapMsg(&pvproto.SignedProposalResponse{
				Proposal: cmtproto.Proposal{}, Error: &pvproto.RemoteSignerError{
					Code:        0,
					Description: "unable to sign proposal",
				},
			})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.SignProposalRequest.GetChainId(), chainID)
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
	case *pvproto.Message_PingRequest:
		err, res = nil, mustWrapMsg(&pvproto.PingResponse{})

	default:
		err = fmt.Errorf("unknown msg: %v", r)
	}

	return res, err
}
