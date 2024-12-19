//nolint:stylecheck,revive
package types

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"

	v2 "github.com/cometbft/cometbft/api/cometbft/abci/v2"
)

type (
	Request                    = v2.Request
	EchoRequest                = v2.EchoRequest
	FlushRequest               = v2.FlushRequest
	InfoRequest                = v2.InfoRequest
	InitChainRequest           = v2.InitChainRequest
	QueryRequest               = v2.QueryRequest
	CheckTxRequest             = v2.CheckTxRequest
	CommitRequest              = v2.CommitRequest
	ListSnapshotsRequest       = v2.ListSnapshotsRequest
	OfferSnapshotRequest       = v2.OfferSnapshotRequest
	LoadSnapshotChunkRequest   = v2.LoadSnapshotChunkRequest
	ApplySnapshotChunkRequest  = v2.ApplySnapshotChunkRequest
	PrepareProposalRequest     = v2.PrepareProposalRequest
	ProcessProposalRequest     = v2.ProcessProposalRequest
	ExtendVoteRequest          = v2.ExtendVoteRequest
	VerifyVoteExtensionRequest = v2.VerifyVoteExtensionRequest
	FinalizeBlockRequest       = v2.FinalizeBlockRequest
)

// Discriminated Request variants are defined in the latest proto package.
type (
	Request_Echo                = v2.Request_Echo
	Request_Flush               = v2.Request_Flush
	Request_Info                = v2.Request_Info
	Request_InitChain           = v2.Request_InitChain
	Request_Query               = v2.Request_Query
	Request_CheckTx             = v2.Request_CheckTx
	Request_Commit              = v2.Request_Commit
	Request_ListSnapshots       = v2.Request_ListSnapshots
	Request_OfferSnapshot       = v2.Request_OfferSnapshot
	Request_LoadSnapshotChunk   = v2.Request_LoadSnapshotChunk
	Request_ApplySnapshotChunk  = v2.Request_ApplySnapshotChunk
	Request_PrepareProposal     = v2.Request_PrepareProposal
	Request_ProcessProposal     = v2.Request_ProcessProposal
	Request_ExtendVote          = v2.Request_ExtendVote
	Request_VerifyVoteExtension = v2.Request_VerifyVoteExtension
	Request_FinalizeBlock       = v2.Request_FinalizeBlock
)

type (
	Response                    = v2.Response
	ExceptionResponse           = v2.ExceptionResponse
	EchoResponse                = v2.EchoResponse
	FlushResponse               = v2.FlushResponse
	InfoResponse                = v2.InfoResponse
	InitChainResponse           = v2.InitChainResponse
	QueryResponse               = v2.QueryResponse
	CheckTxResponse             = v2.CheckTxResponse
	CommitResponse              = v2.CommitResponse
	ListSnapshotsResponse       = v2.ListSnapshotsResponse
	OfferSnapshotResponse       = v2.OfferSnapshotResponse
	LoadSnapshotChunkResponse   = v2.LoadSnapshotChunkResponse
	ApplySnapshotChunkResponse  = v2.ApplySnapshotChunkResponse
	PrepareProposalResponse     = v2.PrepareProposalResponse
	ProcessProposalResponse     = v2.ProcessProposalResponse
	ExtendVoteResponse          = v2.ExtendVoteResponse
	VerifyVoteExtensionResponse = v2.VerifyVoteExtensionResponse
	FinalizeBlockResponse       = v2.FinalizeBlockResponse
)

// Discriminated Response variants are defined in the latest proto package.
type (
	Response_Exception           = v2.Response_Exception
	Response_Echo                = v2.Response_Echo
	Response_Flush               = v2.Response_Flush
	Response_Info                = v2.Response_Info
	Response_InitChain           = v2.Response_InitChain
	Response_Query               = v2.Response_Query
	Response_CheckTx             = v2.Response_CheckTx
	Response_Commit              = v2.Response_Commit
	Response_ListSnapshots       = v2.Response_ListSnapshots
	Response_OfferSnapshot       = v2.Response_OfferSnapshot
	Response_LoadSnapshotChunk   = v2.Response_LoadSnapshotChunk
	Response_ApplySnapshotChunk  = v2.Response_ApplySnapshotChunk
	Response_PrepareProposal     = v2.Response_PrepareProposal
	Response_ProcessProposal     = v2.Response_ProcessProposal
	Response_ExtendVote          = v2.Response_ExtendVote
	Response_VerifyVoteExtension = v2.Response_VerifyVoteExtension
	Response_FinalizeBlock       = v2.Response_FinalizeBlock
)

type (
	CommitInfo         = v2.CommitInfo
	ExecTxResult       = v2.ExecTxResult
	ExtendedCommitInfo = v2.ExtendedCommitInfo
	ExtendedVoteInfo   = v2.ExtendedVoteInfo
	Event              = v2.Event
	EventAttribute     = v2.EventAttribute
	Misbehavior        = v2.Misbehavior
	Snapshot           = v2.Snapshot
	TxResult           = v2.TxResult
	Validator          = v2.Validator
	ValidatorUpdate    = v2.ValidatorUpdate
	VoteInfo           = v2.VoteInfo
)

type (
	ABCIServiceClient = v2.ABCIServiceClient
	ABCIServiceServer = v2.ABCIServiceServer
)

func NewABCIClient(cc grpc.ClientConn) ABCIServiceClient {
	return v2.NewABCIServiceClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServiceServer) {
	v2.RegisterABCIServiceServer(s, srv)
}

type CheckTxType = v2.CheckTxType

const (
	CHECK_TX_TYPE_UNKNOWN CheckTxType = v2.CHECK_TX_TYPE_UNKNOWN
	CHECK_TX_TYPE_CHECK   CheckTxType = v2.CHECK_TX_TYPE_CHECK
	CHECK_TX_TYPE_RECHECK CheckTxType = v2.CHECK_TX_TYPE_RECHECK
)

type MisbehaviorType = v2.MisbehaviorType

const (
	MISBEHAVIOR_TYPE_UNKNOWN             MisbehaviorType = v2.MISBEHAVIOR_TYPE_UNKNOWN
	MISBEHAVIOR_TYPE_DUPLICATE_VOTE      MisbehaviorType = v2.MISBEHAVIOR_TYPE_DUPLICATE_VOTE
	MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK MisbehaviorType = v2.MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK
)

type ApplySnapshotChunkResult = v2.ApplySnapshotChunkResult

const (
	APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN         ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN
	APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT          ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT
	APPLY_SNAPSHOT_CHUNK_RESULT_ABORT           ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY           ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT  ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT
	APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT ApplySnapshotChunkResult = v2.APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT
)

type OfferSnapshotResult = v2.OfferSnapshotResult

const (
	OFFER_SNAPSHOT_RESULT_UNKNOWN       OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_UNKNOWN
	OFFER_SNAPSHOT_RESULT_ACCEPT        OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_ACCEPT
	OFFER_SNAPSHOT_RESULT_ABORT         OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_ABORT
	OFFER_SNAPSHOT_RESULT_REJECT        OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_REJECT
	OFFER_SNAPSHOT_RESULT_REJECT_FORMAT OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT
	OFFER_SNAPSHOT_RESULT_REJECT_SENDER OfferSnapshotResult = v2.OFFER_SNAPSHOT_RESULT_REJECT_SENDER
)

type ProcessProposalStatus = v2.ProcessProposalStatus

const (
	PROCESS_PROPOSAL_STATUS_UNKNOWN ProcessProposalStatus = v2.PROCESS_PROPOSAL_STATUS_UNKNOWN
	PROCESS_PROPOSAL_STATUS_ACCEPT  ProcessProposalStatus = v2.PROCESS_PROPOSAL_STATUS_ACCEPT
	PROCESS_PROPOSAL_STATUS_REJECT  ProcessProposalStatus = v2.PROCESS_PROPOSAL_STATUS_REJECT
)

type VerifyVoteExtensionStatus = v2.VerifyVoteExtensionStatus

const (
	VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN VerifyVoteExtensionStatus = v2.VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN
	VERIFY_VOTE_EXTENSION_STATUS_ACCEPT  VerifyVoteExtensionStatus = v2.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT
	VERIFY_VOTE_EXTENSION_STATUS_REJECT  VerifyVoteExtensionStatus = v2.VERIFY_VOTE_EXTENSION_STATUS_REJECT
)

const (
	CodeTypeOK uint32 = 0
)

// Some compile time assertions to ensure we don't
// have accidental runtime surprises later on.

// jsonRoundTripper ensures that asserted
// interfaces implement both MarshalJSON and UnmarshalJSON.
type jsonRoundTripper interface {
	json.Marshaler
	json.Unmarshaler
}

var (
	_ jsonRoundTripper = (*CommitResponse)(nil)
	_ jsonRoundTripper = (*QueryResponse)(nil)
	_ jsonRoundTripper = (*ExecTxResult)(nil)
	_ jsonRoundTripper = (*CheckTxResponse)(nil)
)

var _ jsonRoundTripper = (*EventAttribute)(nil)

// DeterministicExecTxResult constructs a copy of the ExecTxResult response that omits
// non-deterministic fields. The input response is not modified.
func DeterministicExecTxResult(response *ExecTxResult) *ExecTxResult {
	return &ExecTxResult{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

// MarshalTxResults encodes the TxResults as a list of byte
// slices. It strips off the non-deterministic pieces of the TxResults
// so that the resulting data can be used for hash comparisons and used
// in Merkle proofs.
func MarshalTxResults(r []*ExecTxResult) ([][]byte, error) {
	s := make([][]byte, len(r))
	for i, e := range r {
		d := DeterministicExecTxResult(e)
		b, err := d.Marshal()
		if err != nil {
			return nil, err
		}
		s[i] = b
	}
	return s, nil
}

// -----------------------------------------------
// construct Result data
