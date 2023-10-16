//nolint:revive,stylecheck
package privval

import (
	v1beta1 "github.com/cometbft/cometbft/api/cometbft/privval/v1beta1"
	v1beta2 "github.com/cometbft/cometbft/api/cometbft/privval/v1beta2"
)

type (
	Message                        = v1beta2.Message
	Message_PingRequest            = v1beta2.Message_PingRequest
	Message_PingResponse           = v1beta2.Message_PingResponse
	Message_PubKeyRequest          = v1beta2.Message_PubKeyRequest
	Message_PubKeyResponse         = v1beta2.Message_PubKeyResponse
	Message_SignProposalRequest    = v1beta2.Message_SignProposalRequest
	Message_SignVoteRequest        = v1beta2.Message_SignVoteRequest
	Message_SignedProposalResponse = v1beta2.Message_SignedProposalResponse
	Message_SignedVoteResponse     = v1beta2.Message_SignedVoteResponse
)

type (
	PingRequest            = v1beta1.PingRequest
	PingResponse           = v1beta1.PingResponse
	PubKeyRequest          = v1beta1.PubKeyRequest
	PubKeyResponse         = v1beta1.PubKeyResponse
	RemoteSignerError      = v1beta1.RemoteSignerError
	SignProposalRequest    = v1beta1.SignProposalRequest
	SignedProposalResponse = v1beta1.SignedProposalResponse
	SignVoteRequest        = v1beta2.SignVoteRequest
	SignedVoteResponse     = v1beta2.SignedVoteResponse
)
