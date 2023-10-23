//nolint:revive,stylecheck
package types

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"

	"github.com/cometbft/cometbft/api/cometbft/abci/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/abci/v1beta2"
	"github.com/cometbft/cometbft/api/cometbft/abci/v1beta3"
	"github.com/cometbft/cometbft/api/cometbft/abci/v1beta4"
)

type (
	Request                    = v1beta4.Request
	EchoRequest                = v1beta4.EchoRequest
	FlushRequest               = v1beta4.FlushRequest
	InfoRequest                = v1beta4.InfoRequest
	InitChainRequest           = v1beta4.InitChainRequest
	QueryRequest               = v1beta4.QueryRequest
	CheckTxRequest             = v1beta4.CheckTxRequest
	CommitRequest              = v1beta4.CommitRequest
	ListSnapshotsRequest       = v1beta4.ListSnapshotsRequest
	OfferSnapshotRequest       = v1beta4.OfferSnapshotRequest
	LoadSnapshotChunkRequest   = v1beta4.LoadSnapshotChunkRequest
	ApplySnapshotChunkRequest  = v1beta4.ApplySnapshotChunkRequest
	PrepareProposalRequest     = v1beta4.PrepareProposalRequest
	ProcessProposalRequest     = v1beta4.ProcessProposalRequest
	ExtendVoteRequest          = v1beta4.ExtendVoteRequest
	VerifyVoteExtensionRequest = v1beta4.VerifyVoteExtensionRequest
	FinalizeBlockRequest       = v1beta4.FinalizeBlockRequest
)

// Discriminated Request variants are defined in the latest proto package.
type (
	Request_Echo                = v1beta4.Request_Echo
	Request_Flush               = v1beta4.Request_Flush
	Request_Info                = v1beta4.Request_Info
	Request_InitChain           = v1beta4.Request_InitChain
	Request_Query               = v1beta4.Request_Query
	Request_CheckTx             = v1beta4.Request_CheckTx
	Request_Commit              = v1beta4.Request_Commit
	Request_ListSnapshots       = v1beta4.Request_ListSnapshots
	Request_OfferSnapshot       = v1beta4.Request_OfferSnapshot
	Request_LoadSnapshotChunk   = v1beta4.Request_LoadSnapshotChunk
	Request_ApplySnapshotChunk  = v1beta4.Request_ApplySnapshotChunk
	Request_PrepareProposal     = v1beta4.Request_PrepareProposal
	Request_ProcessProposal     = v1beta4.Request_ProcessProposal
	Request_ExtendVote          = v1beta4.Request_ExtendVote
	Request_VerifyVoteExtension = v1beta4.Request_VerifyVoteExtension
	Request_FinalizeBlock       = v1beta4.Request_FinalizeBlock
)

type (
	Response                    = v1beta4.Response
	ExceptionResponse           = v1beta4.ExceptionResponse
	EchoResponse                = v1beta4.EchoResponse
	FlushResponse               = v1beta4.FlushResponse
	InfoResponse                = v1beta4.InfoResponse
	InitChainResponse           = v1beta4.InitChainResponse
	QueryResponse               = v1beta4.QueryResponse
	CheckTxResponse             = v1beta4.CheckTxResponse
	CommitResponse              = v1beta4.CommitResponse
	ListSnapshotsResponse       = v1beta4.ListSnapshotsResponse
	OfferSnapshotResponse       = v1beta4.OfferSnapshotResponse
	LoadSnapshotChunkResponse   = v1beta4.LoadSnapshotChunkResponse
	ApplySnapshotChunkResponse  = v1beta4.ApplySnapshotChunkResponse
	PrepareProposalResponse     = v1beta4.PrepareProposalResponse
	ProcessProposalResponse     = v1beta4.ProcessProposalResponse
	ExtendVoteResponse          = v1beta4.ExtendVoteResponse
	VerifyVoteExtensionResponse = v1beta4.VerifyVoteExtensionResponse
	FinalizeBlockResponse       = v1beta4.FinalizeBlockResponse
)

// Discriminated Response variants are defined in the latest proto package.
type (
	Response_Exception           = v1beta4.Response_Exception
	Response_Echo                = v1beta4.Response_Echo
	Response_Flush               = v1beta4.Response_Flush
	Response_Info                = v1beta4.Response_Info
	Response_InitChain           = v1beta4.Response_InitChain
	Response_Query               = v1beta4.Response_Query
	Response_CheckTx             = v1beta4.Response_CheckTx
	Response_Commit              = v1beta4.Response_Commit
	Response_ListSnapshots       = v1beta4.Response_ListSnapshots
	Response_OfferSnapshot       = v1beta4.Response_OfferSnapshot
	Response_LoadSnapshotChunk   = v1beta4.Response_LoadSnapshotChunk
	Response_ApplySnapshotChunk  = v1beta4.Response_ApplySnapshotChunk
	Response_PrepareProposal     = v1beta4.Response_PrepareProposal
	Response_ProcessProposal     = v1beta4.Response_ProcessProposal
	Response_ExtendVote          = v1beta4.Response_ExtendVote
	Response_VerifyVoteExtension = v1beta4.Response_VerifyVoteExtension
	Response_FinalizeBlock       = v1beta4.Response_FinalizeBlock
)

type (
	CommitInfo         = v1beta3.CommitInfo
	ExecTxResult       = v1beta3.ExecTxResult
	ExtendedCommitInfo = v1beta3.ExtendedCommitInfo
	ExtendedVoteInfo   = v1beta3.ExtendedVoteInfo
	Event              = v1beta2.Event
	EventAttribute     = v1beta2.EventAttribute
	Misbehavior        = v1beta4.Misbehavior
	Snapshot           = v1beta1.Snapshot
	TxResult           = v1beta3.TxResult
	Validator          = v1beta1.Validator
	ValidatorUpdate    = v1beta1.ValidatorUpdate
	VoteInfo           = v1beta3.VoteInfo
)

type (
	ABCIServiceClient = v1beta4.ABCIServiceClient
	ABCIServiceServer = v1beta4.ABCIServiceServer
)

func NewABCIClient(cc grpc.ClientConn) ABCIServiceClient {
	return v1beta4.NewABCIServiceClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServiceServer) {
	v1beta4.RegisterABCIServiceServer(s, srv)
}

type CheckTxType = v1beta4.CheckTxType

const (
	CHECK_TX_TYPE_UNKNOWN CheckTxType = v1beta4.CHECK_TX_TYPE_UNKNOWN
	CHECK_TX_TYPE_CHECK   CheckTxType = v1beta4.CHECK_TX_TYPE_CHECK
	CHECK_TX_TYPE_RECHECK CheckTxType = v1beta4.CHECK_TX_TYPE_RECHECK
)

type MisbehaviorType = v1beta4.MisbehaviorType

const (
	MISBEHAVIOR_TYPE_UNKNOWN             MisbehaviorType = v1beta4.MISBEHAVIOR_TYPE_UNKNOWN
	MISBEHAVIOR_TYPE_DUPLICATE_VOTE      MisbehaviorType = v1beta4.MISBEHAVIOR_TYPE_DUPLICATE_VOTE
	MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK MisbehaviorType = v1beta4.MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK
)

type ApplySnapshotChunkResult = v1beta4.ApplySnapshotChunkResult

const (
	APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN         ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN
	APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT          ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT
	APPLY_SNAPSHOT_CHUNK_RESULT_ABORT           ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY           ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT  ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT
	APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT ApplySnapshotChunkResult = v1beta4.APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT
)

type OfferSnapshotResult = v1beta4.OfferSnapshotResult

const (
	OFFER_SNAPSHOT_RESULT_UNKNOWN       OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_UNKNOWN
	OFFER_SNAPSHOT_RESULT_ACCEPT        OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_ACCEPT
	OFFER_SNAPSHOT_RESULT_ABORT         OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_ABORT
	OFFER_SNAPSHOT_RESULT_REJECT        OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_REJECT
	OFFER_SNAPSHOT_RESULT_REJECT_FORMAT OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT
	OFFER_SNAPSHOT_RESULT_REJECT_SENDER OfferSnapshotResult = v1beta4.OFFER_SNAPSHOT_RESULT_REJECT_SENDER
)

type ProcessProposalStatus = v1beta4.ProcessProposalStatus

const (
	PROCESS_PROPOSAL_STATUS_UNKNOWN ProcessProposalStatus = v1beta4.PROCESS_PROPOSAL_STATUS_UNKNOWN
	PROCESS_PROPOSAL_STATUS_ACCEPT  ProcessProposalStatus = v1beta4.PROCESS_PROPOSAL_STATUS_ACCEPT
	PROCESS_PROPOSAL_STATUS_REJECT  ProcessProposalStatus = v1beta4.PROCESS_PROPOSAL_STATUS_REJECT
)

type VerifyVoteExtensionStatus = v1beta4.VerifyVoteExtensionStatus

const (
	VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN VerifyVoteExtensionStatus = v1beta4.VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN
	VERIFY_VOTE_EXTENSION_STATUS_ACCEPT  VerifyVoteExtensionStatus = v1beta4.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT
	VERIFY_VOTE_EXTENSION_STATUS_REJECT  VerifyVoteExtensionStatus = v1beta4.VERIFY_VOTE_EXTENSION_STATUS_REJECT
)

const (
	CodeTypeOK uint32 = 0
)

// Some compile time assertions to ensure we don't
// have accidental runtime surprises later on.

// jsonEncodingRoundTripper ensures that asserted
// interfaces implement both MarshalJSON and UnmarshalJSON
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

// constructs a copy of response that omits
// non-deterministic fields. The input response is not modified.
func DeterministicExecTxResult(response *ExecTxResult) *ExecTxResult {
	return &ExecTxResult{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

// MarshalTxResults encodes the the TxResults as a list of byte
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
