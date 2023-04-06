package privval

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/privval/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/privval/v2"
)

type Message = v2.Message
type Message_PingRequest = v2.Message_PingRequest
type Message_PingResponse = v2.Message_PingResponse
type Message_PubKeyRequest = v2.Message_PubKeyRequest
type Message_PubKeyResponse = v2.Message_PubKeyResponse
type Message_SignProposalRequest = v2.Message_SignProposalRequest
type Message_SignVoteRequest = v2.Message_SignVoteRequest
type Message_SignedProposalResponse = v2.Message_SignedProposalResponse
type Message_SignedVoteResponse = v2.Message_SignedVoteResponse

type PingRequest = v1.PingRequest
type PingResponse = v1.PingResponse
type PubKeyRequest = v1.PubKeyRequest
type PubKeyResponse = v1.PubKeyResponse
type RemoteSignerError = v1.RemoteSignerError
type SignProposalRequest = v1.SignProposalRequest
type SignedProposalResponse = v1.SignedProposalResponse
type SignVoteRequest = v2.SignVoteRequest
type SignedVoteResponse = v2.SignedVoteResponse
