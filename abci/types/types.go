//nolint:revive,stylecheck
package types

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"

	v1beta1 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta1"
	v1beta2 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta2"
	v1beta3 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta3"
	v1beta4 "github.com/cometbft/cometbft/api/cometbft/abci/v1beta4"
)

type (
	Request                    = v1beta4.Request
	RequestEcho                = v1beta1.RequestEcho
	RequestFlush               = v1beta1.RequestFlush
	RequestInfo                = v1beta2.RequestInfo
	RequestInitChain           = v1beta3.RequestInitChain
	RequestQuery               = v1beta1.RequestQuery
	RequestCheckTx             = v1beta4.RequestCheckTx
	RequestCommit              = v1beta1.RequestCommit
	RequestListSnapshots       = v1beta1.RequestListSnapshots
	RequestOfferSnapshot       = v1beta1.RequestOfferSnapshot
	RequestLoadSnapshotChunk   = v1beta1.RequestLoadSnapshotChunk
	RequestApplySnapshotChunk  = v1beta1.RequestApplySnapshotChunk
	RequestPrepareProposal     = v1beta4.RequestPrepareProposal
	RequestProcessProposal     = v1beta4.RequestProcessProposal
	RequestExtendVote          = v1beta4.RequestExtendVote
	RequestVerifyVoteExtension = v1beta3.RequestVerifyVoteExtension
	RequestFinalizeBlock       = v1beta4.RequestFinalizeBlock
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
	ResponseException           = v1beta1.ResponseException
	ResponseEcho                = v1beta1.ResponseEcho
	ResponseFlush               = v1beta1.ResponseFlush
	ResponseInfo                = v1beta1.ResponseInfo
	ResponseInitChain           = v1beta3.ResponseInitChain
	ResponseQuery               = v1beta1.ResponseQuery
	ResponseCheckTx             = v1beta3.ResponseCheckTx
	ResponseCommit              = v1beta3.ResponseCommit
	ResponseListSnapshots       = v1beta1.ResponseListSnapshots
	ResponseOfferSnapshot       = v1beta4.ResponseOfferSnapshot
	ResponseLoadSnapshotChunk   = v1beta1.ResponseLoadSnapshotChunk
	ResponseApplySnapshotChunk  = v1beta4.ResponseApplySnapshotChunk
	ResponsePrepareProposal     = v1beta2.ResponsePrepareProposal
	ResponseProcessProposal     = v1beta4.ResponseProcessProposal
	ResponseExtendVote          = v1beta3.ResponseExtendVote
	ResponseVerifyVoteExtension = v1beta4.ResponseVerifyVoteExtension
	ResponseFinalizeBlock       = v1beta3.ResponseFinalizeBlock
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
	ABCIClient = v1beta4.ABCIClient
	ABCIServer = v1beta4.ABCIServer
)

func NewABCIClient(cc grpc.ClientConn) ABCIClient {
	return v1beta4.NewABCIClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServer) {
	v1beta4.RegisterABCIServer(s, srv)
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
