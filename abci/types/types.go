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

type Request = v4.Request
type RequestEcho = v1.RequestEcho
type RequestFlush = v1.RequestFlush
type RequestInfo = v2.RequestInfo
type RequestInitChain = v3.RequestInitChain
type RequestQuery = v1.RequestQuery
type RequestCheckTx = v4.RequestCheckTx
type RequestCommit = v1.RequestCommit
type RequestListSnapshots = v1.RequestListSnapshots
type RequestOfferSnapshot = v1.RequestOfferSnapshot
type RequestLoadSnapshotChunk = v1.RequestLoadSnapshotChunk
type RequestApplySnapshotChunk = v1.RequestApplySnapshotChunk
type RequestPrepareProposal = v4.RequestPrepareProposal
type RequestProcessProposal = v4.RequestProcessProposal
type RequestExtendVote = v3.RequestExtendVote
type RequestVerifyVoteExtension = v3.RequestVerifyVoteExtension
type RequestFinalizeBlock = v4.RequestFinalizeBlock

// Discriminated Request variants are defined in the latest proto package.
type Request_Echo = v4.Request_Echo
type Request_Flush = v4.Request_Flush
type Request_Info = v4.Request_Info
type Request_InitChain = v4.Request_InitChain
type Request_Query = v4.Request_Query
type Request_CheckTx = v4.Request_CheckTx
type Request_Commit = v4.Request_Commit
type Request_ListSnapshots = v4.Request_ListSnapshots
type Request_OfferSnapshot = v4.Request_OfferSnapshot
type Request_LoadSnapshotChunk = v4.Request_LoadSnapshotChunk
type Request_ApplySnapshotChunk = v4.Request_ApplySnapshotChunk
type Request_PrepareProposal = v4.Request_PrepareProposal
type Request_ProcessProposal = v4.Request_ProcessProposal
type Request_ExtendVote = v4.Request_ExtendVote
type Request_VerifyVoteExtension = v4.Request_VerifyVoteExtension
type Request_FinalizeBlock = v4.Request_FinalizeBlock

type Response = v4.Response
type ResponseException = v1.ResponseException
type ResponseEcho = v1.ResponseEcho
type ResponseFlush = v1.ResponseFlush
type ResponseInfo = v1.ResponseInfo
type ResponseInitChain = v3.ResponseInitChain
type ResponseQuery = v1.ResponseQuery
type ResponseCheckTx = v3.ResponseCheckTx
type ResponseCommit = v3.ResponseCommit
type ResponseListSnapshots = v1.ResponseListSnapshots
type ResponseOfferSnapshot = v4.ResponseOfferSnapshot
type ResponseLoadSnapshotChunk = v1.ResponseLoadSnapshotChunk
type ResponseApplySnapshotChunk = v4.ResponseApplySnapshotChunk
type ResponsePrepareProposal = v2.ResponsePrepareProposal
type ResponseProcessProposal = v4.ResponseProcessProposal
type ResponseExtendVote = v3.ResponseExtendVote
type ResponseVerifyVoteExtension = v4.ResponseVerifyVoteExtension
type ResponseFinalizeBlock = v3.ResponseFinalizeBlock

// Discriminated Response variants are defined in the latest proto package.
type Response_Exception = v4.Response_Exception
type Response_Echo = v4.Response_Echo
type Response_Flush = v4.Response_Flush
type Response_Info = v4.Response_Info
type Response_InitChain = v4.Response_InitChain
type Response_Query = v4.Response_Query
type Response_CheckTx = v4.Response_CheckTx
type Response_Commit = v4.Response_Commit
type Response_ListSnapshots = v4.Response_ListSnapshots
type Response_OfferSnapshot = v4.Response_OfferSnapshot
type Response_LoadSnapshotChunk = v4.Response_LoadSnapshotChunk
type Response_ApplySnapshotChunk = v4.Response_ApplySnapshotChunk
type Response_PrepareProposal = v4.Response_PrepareProposal
type Response_ProcessProposal = v4.Response_ProcessProposal
type Response_ExtendVote = v4.Response_ExtendVote
type Response_VerifyVoteExtension = v4.Response_VerifyVoteExtension
type Response_FinalizeBlock = v4.Response_FinalizeBlock

type CommitInfo = v3.CommitInfo
type ExecTxResult = v3.ExecTxResult
type ExtendedCommitInfo = v3.ExtendedCommitInfo
type ExtendedVoteInfo = v3.ExtendedVoteInfo
type Event = v2.Event
type EventAttribute = v2.EventAttribute
type Misbehavior = v4.Misbehavior
type Snapshot = v1.Snapshot
type TxResult = v3.TxResult
type Validator = v1.Validator
type ValidatorUpdate = v1.ValidatorUpdate
type VoteInfo = v3.VoteInfo

type ABCIClient = v4.ABCIClient
type ABCIServer = v4.ABCIServer

func NewABCIClient(cc grpc.ClientConn) ABCIClient {
	return v4.NewABCIClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServer) {
	v4.RegisterABCIServer(s, srv)
}

type CheckTxType = v4.CheckTxType

const (
	CHECK_TX_TYPE_UNKNOWN CheckTxType = v4.CHECK_TX_TYPE_UNKNOWN
	CHECK_TX_TYPE_NEW     CheckTxType = v4.CHECK_TX_TYPE_NEW
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

var _ jsonRoundTripper = (*ResponseCommit)(nil)
var _ jsonRoundTripper = (*ResponseQuery)(nil)
var _ jsonRoundTripper = (*ExecTxResult)(nil)
var _ jsonRoundTripper = (*ResponseCheckTx)(nil)

var _ jsonRoundTripper = (*EventAttribute)(nil)

// deterministicExecTxResult constructs a copy of response that omits
// non-deterministic fields. The input response is not modified.
func deterministicExecTxResult(response *ExecTxResult) *ExecTxResult {
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
		d := deterministicExecTxResult(e)
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
