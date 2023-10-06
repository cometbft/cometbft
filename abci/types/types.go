//nolint:revive,stylecheck
package types

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"

	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/abci/v2"
	v3 "github.com/cometbft/cometbft/api/cometbft/abci/v3"
	v4 "github.com/cometbft/cometbft/api/cometbft/abci/v4"
)

type (
	Request                    = v4.Request
	RequestEcho                = v1.RequestEcho
	RequestFlush               = v1.RequestFlush
	RequestInfo                = v2.RequestInfo
	RequestInitChain           = v3.RequestInitChain
	RequestQuery               = v1.RequestQuery
	RequestCheckTx             = v4.RequestCheckTx
	RequestCommit              = v1.RequestCommit
	RequestListSnapshots       = v1.RequestListSnapshots
	RequestOfferSnapshot       = v1.RequestOfferSnapshot
	RequestLoadSnapshotChunk   = v1.RequestLoadSnapshotChunk
	RequestApplySnapshotChunk  = v1.RequestApplySnapshotChunk
	RequestPrepareProposal     = v4.RequestPrepareProposal
	RequestProcessProposal     = v4.RequestProcessProposal
	RequestExtendVote          = v4.RequestExtendVote
	RequestVerifyVoteExtension = v3.RequestVerifyVoteExtension
	RequestFinalizeBlock       = v4.RequestFinalizeBlock
)

// Discriminated Request variants are defined in the latest proto package.
type (
	Request_Echo                = v4.Request_Echo
	Request_Flush               = v4.Request_Flush
	Request_Info                = v4.Request_Info
	Request_InitChain           = v4.Request_InitChain
	Request_Query               = v4.Request_Query
	Request_CheckTx             = v4.Request_CheckTx
	Request_Commit              = v4.Request_Commit
	Request_ListSnapshots       = v4.Request_ListSnapshots
	Request_OfferSnapshot       = v4.Request_OfferSnapshot
	Request_LoadSnapshotChunk   = v4.Request_LoadSnapshotChunk
	Request_ApplySnapshotChunk  = v4.Request_ApplySnapshotChunk
	Request_PrepareProposal     = v4.Request_PrepareProposal
	Request_ProcessProposal     = v4.Request_ProcessProposal
	Request_ExtendVote          = v4.Request_ExtendVote
	Request_VerifyVoteExtension = v4.Request_VerifyVoteExtension
	Request_FinalizeBlock       = v4.Request_FinalizeBlock
)

type (
	Response                    = v4.Response
	ResponseException           = v1.ResponseException
	ResponseEcho                = v1.ResponseEcho
	ResponseFlush               = v1.ResponseFlush
	ResponseInfo                = v1.ResponseInfo
	ResponseInitChain           = v3.ResponseInitChain
	ResponseQuery               = v1.ResponseQuery
	ResponseCheckTx             = v3.ResponseCheckTx
	ResponseCommit              = v3.ResponseCommit
	ResponseListSnapshots       = v1.ResponseListSnapshots
	ResponseOfferSnapshot       = v4.ResponseOfferSnapshot
	ResponseLoadSnapshotChunk   = v1.ResponseLoadSnapshotChunk
	ResponseApplySnapshotChunk  = v4.ResponseApplySnapshotChunk
	ResponsePrepareProposal     = v2.ResponsePrepareProposal
	ResponseProcessProposal     = v4.ResponseProcessProposal
	ResponseExtendVote          = v3.ResponseExtendVote
	ResponseVerifyVoteExtension = v4.ResponseVerifyVoteExtension
	ResponseFinalizeBlock       = v3.ResponseFinalizeBlock
)

// Discriminated Response variants are defined in the latest proto package.
type (
	Response_Exception           = v4.Response_Exception
	Response_Echo                = v4.Response_Echo
	Response_Flush               = v4.Response_Flush
	Response_Info                = v4.Response_Info
	Response_InitChain           = v4.Response_InitChain
	Response_Query               = v4.Response_Query
	Response_CheckTx             = v4.Response_CheckTx
	Response_Commit              = v4.Response_Commit
	Response_ListSnapshots       = v4.Response_ListSnapshots
	Response_OfferSnapshot       = v4.Response_OfferSnapshot
	Response_LoadSnapshotChunk   = v4.Response_LoadSnapshotChunk
	Response_ApplySnapshotChunk  = v4.Response_ApplySnapshotChunk
	Response_PrepareProposal     = v4.Response_PrepareProposal
	Response_ProcessProposal     = v4.Response_ProcessProposal
	Response_ExtendVote          = v4.Response_ExtendVote
	Response_VerifyVoteExtension = v4.Response_VerifyVoteExtension
	Response_FinalizeBlock       = v4.Response_FinalizeBlock
)

type (
	CommitInfo         = v3.CommitInfo
	ExecTxResult       = v3.ExecTxResult
	ExtendedCommitInfo = v3.ExtendedCommitInfo
	ExtendedVoteInfo   = v3.ExtendedVoteInfo
	Event              = v2.Event
	EventAttribute     = v2.EventAttribute
	Misbehavior        = v4.Misbehavior
	Snapshot           = v1.Snapshot
	TxResult           = v3.TxResult
	Validator          = v1.Validator
	ValidatorUpdate    = v1.ValidatorUpdate
	VoteInfo           = v3.VoteInfo
)

type (
	ABCIClient = v4.ABCIClient
	ABCIServer = v4.ABCIServer
)

func NewABCIClient(cc grpc.ClientConn) ABCIClient {
	return v4.NewABCIClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServer) {
	v4.RegisterABCIServer(s, srv)
}

type CheckTxType = v4.CheckTxType

const (
	CHECK_TX_TYPE_UNKNOWN CheckTxType = v4.CHECK_TX_TYPE_UNKNOWN
	CHECK_TX_TYPE_CHECK   CheckTxType = v4.CHECK_TX_TYPE_CHECK
	CHECK_TX_TYPE_RECHECK CheckTxType = v4.CHECK_TX_TYPE_RECHECK
)

type MisbehaviorType = v4.MisbehaviorType

const (
	MISBEHAVIOR_TYPE_UNKNOWN             MisbehaviorType = v4.MISBEHAVIOR_TYPE_UNKNOWN
	MISBEHAVIOR_TYPE_DUPLICATE_VOTE      MisbehaviorType = v4.MISBEHAVIOR_TYPE_DUPLICATE_VOTE
	MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK MisbehaviorType = v4.MISBEHAVIOR_TYPE_LIGHT_CLIENT_ATTACK
)

type ApplySnapshotChunkResult = v4.ApplySnapshotChunkResult

const (
	APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN         ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_UNKNOWN
	APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT          ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT
	APPLY_SNAPSHOT_CHUNK_RESULT_ABORT           ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY           ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY
	APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT  ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT
	APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT ApplySnapshotChunkResult = v4.APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT
)

type OfferSnapshotResult = v4.OfferSnapshotResult

const (
	OFFER_SNAPSHOT_RESULT_UNKNOWN       OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_UNKNOWN
	OFFER_SNAPSHOT_RESULT_ACCEPT        OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_ACCEPT
	OFFER_SNAPSHOT_RESULT_ABORT         OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_ABORT
	OFFER_SNAPSHOT_RESULT_REJECT        OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_REJECT
	OFFER_SNAPSHOT_RESULT_REJECT_FORMAT OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT
	OFFER_SNAPSHOT_RESULT_REJECT_SENDER OfferSnapshotResult = v4.OFFER_SNAPSHOT_RESULT_REJECT_SENDER
)

type ProcessProposalStatus = v4.ProcessProposalStatus

const (
	PROCESS_PROPOSAL_STATUS_UNKNOWN ProcessProposalStatus = v4.PROCESS_PROPOSAL_STATUS_UNKNOWN
	PROCESS_PROPOSAL_STATUS_ACCEPT  ProcessProposalStatus = v4.PROCESS_PROPOSAL_STATUS_ACCEPT
	PROCESS_PROPOSAL_STATUS_REJECT  ProcessProposalStatus = v4.PROCESS_PROPOSAL_STATUS_REJECT
)

type VerifyVoteExtensionStatus = v4.VerifyVoteExtensionStatus

const (
	VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN VerifyVoteExtensionStatus = v4.VERIFY_VOTE_EXTENSION_STATUS_UNKNOWN
	VERIFY_VOTE_EXTENSION_STATUS_ACCEPT  VerifyVoteExtensionStatus = v4.VERIFY_VOTE_EXTENSION_STATUS_ACCEPT
	VERIFY_VOTE_EXTENSION_STATUS_REJECT  VerifyVoteExtensionStatus = v4.VERIFY_VOTE_EXTENSION_STATUS_REJECT
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
	_ jsonRoundTripper = (*ResponseCommit)(nil)
	_ jsonRoundTripper = (*ResponseQuery)(nil)
	_ jsonRoundTripper = (*ExecTxResult)(nil)
	_ jsonRoundTripper = (*ResponseCheckTx)(nil)
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
