package abci

import (
	"fmt"
	"strconv"
	"strings"
)

type ABCICallType int32

const (
	Info               ABCICallType = 0
	InitChain          ABCICallType = 1
	CheckTx            ABCICallType = 2
	BeginBlock         ABCICallType = 3
	DeliverTx          ABCICallType = 4
	EndBlock           ABCICallType = 5
	Commit             ABCICallType = 6
	Query              ABCICallType = 7
	ListSnapshots      ABCICallType = 8
	LoadSnapshotChunk  ABCICallType = 9
	OfferSnapshot      ABCICallType = 10
	ApplySnapshotChunk ABCICallType = 11
	Rollback           ABCICallType = 12
	PrepareProposal    ABCICallType = 13
	ProcessProposal    ABCICallType = 14
)

const ABCICallIdentifier = "ABCI-Call"

type ABCICall struct {
	Type ABCICallType
}

func NewInfoABCICall() *ABCICall {
	return &ABCICall{
		Type: Info,
	}
}

func NewInitChainABCICall() *ABCICall {
	return &ABCICall{
		Type: InitChain,
	}
}

func NewCheckTxABCICall() *ABCICall {
	return &ABCICall{
		Type: CheckTx,
	}
}

func NewBeginBlockABCICall() *ABCICall {
	return &ABCICall{
		Type: BeginBlock,
	}
}

func NewDeliverTxABCICall() *ABCICall {
	return &ABCICall{
		Type: DeliverTx,
	}
}

func NewCommitABCICall() *ABCICall {
	return &ABCICall{
		Type: Commit,
	}
}

func NewQueryABCICall() *ABCICall {
	return &ABCICall{
		Type: Query,
	}
}

func NewListSnapshotsABCICall() *ABCICall {
	return &ABCICall{
		Type: ListSnapshots,
	}
}

func NewLoadSnapshotChunkABCICall() *ABCICall {
	return &ABCICall{
		Type: LoadSnapshotChunk,
	}
}

func NewOfferSnapshotABCICall() *ABCICall {
	return &ABCICall{
		Type: OfferSnapshot,
	}
}

func NewApplySnapshotChunkABCICall() *ABCICall {
	return &ABCICall{
		Type: ApplySnapshotChunk,
	}
}

func NewRollbackABCICall() *ABCICall {
	return &ABCICall{
		Type: Rollback,
	}
}

func NewPrepareProposalABCICall() *ABCICall {
	return &ABCICall{
		Type: PrepareProposal,
	}
}

func NewProcessProposalABCICall() *ABCICall {
	return &ABCICall{
		Type: ProcessProposal,
	}
}

func NewEndBlockABCICall() *ABCICall {
	return &ABCICall{
		Type: EndBlock,
	}
}

func (a *ABCICall) ToString() string {
	s := fmt.Sprintf("%v:%v", ABCICallIdentifier, a.Type)
	return s
}

func (a *ABCICall) FromString(s string) {
	parts := strings.Split(s, ":")
	t, _ := strconv.Atoi(parts[1])
	a.Type = ABCICallType(t)
}
