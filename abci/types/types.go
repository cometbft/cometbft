package types

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"

	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/abci/v2"
	v3 "github.com/cometbft/cometbft/api/cometbft/abci/v3"
)

type Request = v3.Request
type RequestEcho = v1.RequestEcho
type RequestFlush = v1.RequestFlush
type RequestInfo = v2.RequestInfo
type RequestInitChain = v3.RequestInitChain
type RequestQuery = v1.RequestQuery
type RequestCheckTx = v1.RequestCheckTx
type RequestCommit = v1.RequestCommit
type RequestListSnapshots = v1.RequestListSnapshots
type RequestOfferSnapshot = v1.RequestOfferSnapshot
type RequestLoadSnapshotChunk = v1.RequestLoadSnapshotChunk
type RequestApplySnapshotChunk = v1.RequestApplySnapshotChunk
type RequestPrepareProposal = v3.RequestPrepareProposal
type RequestProcessProposal = v3.RequestProcessProposal
type RequestExtendVote = v3.RequestExtendVote
type RequestVerifyVoteExtension = v3.RequestVerifyVoteExtension
type RequestFinalizeBlock = v3.RequestFinalizeBlock

// Discriminated Request variants are defined in the latest proto package.
type Request_Echo = v3.Request_Echo
type Request_Flush = v3.Request_Flush
type Request_Info = v3.Request_Info
type Request_InitChain = v3.Request_InitChain
type Request_Query = v3.Request_Query
type Request_CheckTx = v3.Request_CheckTx
type Request_Commit = v3.Request_Commit
type Request_ListSnapshots = v3.Request_ListSnapshots
type Request_OfferSnapshot = v3.Request_OfferSnapshot
type Request_LoadSnapshotChunk = v3.Request_LoadSnapshotChunk
type Request_ApplySnapshotChunk = v3.Request_ApplySnapshotChunk
type Request_PrepareProposal = v3.Request_PrepareProposal
type Request_ProcessProposal = v3.Request_ProcessProposal
type Request_ExtendVote = v3.Request_ExtendVote
type Request_VerifyVoteExtension = v3.Request_VerifyVoteExtension
type Request_FinalizeBlock = v3.Request_FinalizeBlock

type Response = v3.Response
type ResponseException = v1.ResponseException
type ResponseEcho = v1.ResponseEcho
type ResponseFlush = v1.ResponseFlush
type ResponseInfo = v1.ResponseInfo
type ResponseInitChain = v3.ResponseInitChain
type ResponseQuery = v1.ResponseQuery
type ResponseCheckTx = v3.ResponseCheckTx
type ResponseCommit = v3.ResponseCommit
type ResponseListSnapshots = v1.ResponseListSnapshots
type ResponseOfferSnapshot = v1.ResponseOfferSnapshot
type ResponseLoadSnapshotChunk = v1.ResponseLoadSnapshotChunk
type ResponseApplySnapshotChunk = v1.ResponseApplySnapshotChunk
type ResponsePrepareProposal = v2.ResponsePrepareProposal
type ResponseProcessProposal = v2.ResponseProcessProposal
type ResponseExtendVote = v3.ResponseExtendVote
type ResponseVerifyVoteExtension = v3.ResponseVerifyVoteExtension
type ResponseFinalizeBlock = v3.ResponseFinalizeBlock

// Discriminated Response variants are defined in the latest proto package.
type Response_Exception = v3.Response_Exception
type Response_Echo = v3.Response_Echo
type Response_Flush = v3.Response_Flush
type Response_Info = v3.Response_Info
type Response_InitChain = v3.Response_InitChain
type Response_Query = v3.Response_Query
type Response_CheckTx = v3.Response_CheckTx
type Response_Commit = v3.Response_Commit
type Response_ListSnapshots = v3.Response_ListSnapshots
type Response_OfferSnapshot = v3.Response_OfferSnapshot
type Response_LoadSnapshotChunk = v3.Response_LoadSnapshotChunk
type Response_ApplySnapshotChunk = v3.Response_ApplySnapshotChunk
type Response_PrepareProposal = v3.Response_PrepareProposal
type Response_ProcessProposal = v3.Response_ProcessProposal
type Response_ExtendVote = v3.Response_ExtendVote
type Response_VerifyVoteExtension = v3.Response_VerifyVoteExtension
type Response_FinalizeBlock = v3.Response_FinalizeBlock

type CommitInfo = v3.CommitInfo
type ExecTxResult = v3.ExecTxResult
type ExtendedCommitInfo = v3.ExtendedCommitInfo
type ExtendedVoteInfo = v3.ExtendedVoteInfo
type Event = v2.Event
type EventAttribute = v2.EventAttribute
type Misbehavior = v2.Misbehavior
type Snapshot = v1.Snapshot
type TxResult = v3.TxResult
type Validator = v1.Validator
type ValidatorUpdate = v1.ValidatorUpdate
type VoteInfo = v3.VoteInfo

type ABCIClient = v3.ABCIClient
type ABCIServer = v3.ABCIServer

func NewABCIClient(cc grpc.ClientConn) ABCIClient {
	return v3.NewABCIClient(cc)
}

func RegisterABCIServer(s grpc.Server, srv ABCIServer) {
	v3.RegisterABCIServer(s, srv)
}

type CheckTxType = v1.CheckTxType

const (
	CheckTxType_New     CheckTxType = v1.CheckTxType_New
	CheckTxType_Recheck CheckTxType = v1.CheckTxType_Recheck
)

type MisbehaviorType = v2.MisbehaviorType

const (
	MisbehaviorType_UNKNOWN             MisbehaviorType = v2.MisbehaviorType_UNKNOWN
	MisbehaviorType_DUPLICATE_VOTE      MisbehaviorType = v2.MisbehaviorType_DUPLICATE_VOTE
	MisbehaviorType_LIGHT_CLIENT_ATTACK MisbehaviorType = v2.MisbehaviorType_LIGHT_CLIENT_ATTACK
)

type ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_Result

const (
	ResponseApplySnapshotChunk_UNKNOWN         ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_UNKNOWN
	ResponseApplySnapshotChunk_ACCEPT          ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_ACCEPT
	ResponseApplySnapshotChunk_ABORT           ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_ABORT
	ResponseApplySnapshotChunk_RETRY           ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_RETRY
	ResponseApplySnapshotChunk_RETRY_SNAPSHOT  ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_RETRY_SNAPSHOT
	ResponseApplySnapshotChunk_REJECT_SNAPSHOT ResponseApplySnapshotChunk_Result = v1.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
)

type ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_Result

const (
	ResponseOfferSnapshot_UNKNOWN       ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_UNKNOWN
	ResponseOfferSnapshot_ACCEPT        ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_ACCEPT
	ResponseOfferSnapshot_ABORT         ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_ABORT
	ResponseOfferSnapshot_REJECT        ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_REJECT
	ResponseOfferSnapshot_REJECT_FORMAT ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_REJECT_FORMAT
	ResponseOfferSnapshot_REJECT_SENDER ResponseOfferSnapshot_Result = v1.ResponseOfferSnapshot_REJECT_SENDER
)

type ResponseProcessProposal_ProposalStatus = v2.ResponseProcessProposal_ProposalStatus

const (
	ResponseProcessProposal_UNKNOWN ResponseProcessProposal_ProposalStatus = v2.ResponseProcessProposal_UNKNOWN
	ResponseProcessProposal_ACCEPT  ResponseProcessProposal_ProposalStatus = v2.ResponseProcessProposal_ACCEPT
	ResponseProcessProposal_REJECT  ResponseProcessProposal_ProposalStatus = v2.ResponseProcessProposal_REJECT
)

type ResponseVerifyVoteExtension_VerifyStatus = v3.ResponseVerifyVoteExtension_VerifyStatus

const (
	ResponseVerifyVoteExtension_UNKNOWN ResponseVerifyVoteExtension_VerifyStatus = v3.ResponseVerifyVoteExtension_UNKNOWN
	ResponseVerifyVoteExtension_ACCEPT  ResponseVerifyVoteExtension_VerifyStatus = v3.ResponseVerifyVoteExtension_ACCEPT
	ResponseVerifyVoteExtension_REJECT  ResponseVerifyVoteExtension_VerifyStatus = v3.ResponseVerifyVoteExtension_REJECT
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
