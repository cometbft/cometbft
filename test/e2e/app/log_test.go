package app

import (
	"reflect"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
)

// Tests for logging each type of requests.
func TestLogging(t *testing.T) {
	var reqs = []*abci.Request{
		{Value: &abci.Request_Echo{Echo: &abci.RequestEcho{}}},
		{Value: &abci.Request_Flush{Flush: &abci.RequestFlush{}}},
		{Value: &abci.Request_Info{Info: &abci.RequestInfo{}}},
		{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}},
		{Value: &abci.Request_Query{Query: &abci.RequestQuery{}}},
		{Value: &abci.Request_BeginBlock{BeginBlock: &abci.RequestBeginBlock{}}},
		{Value: &abci.Request_CheckTx{CheckTx: &abci.RequestCheckTx{}}},
		{Value: &abci.Request_DeliverTx{DeliverTx: &abci.RequestDeliverTx{}}},
		{Value: &abci.Request_Commit{Commit: &abci.RequestCommit{}}},
		{Value: &abci.Request_EndBlock{EndBlock: &abci.RequestEndBlock{}}},
		{Value: &abci.Request_ListSnapshots{ListSnapshots: &abci.RequestListSnapshots{}}},
		{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &abci.RequestOfferSnapshot{}}},
		{Value: &abci.Request_LoadSnapshotChunk{LoadSnapshotChunk: &abci.RequestLoadSnapshotChunk{}}},
		{Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &abci.RequestApplySnapshotChunk{}}},
		{Value: &abci.Request_PrepareProposal{PrepareProposal: &abci.RequestPrepareProposal{}}},
		{Value: &abci.Request_ProcessProposal{ProcessProposal: &abci.RequestProcessProposal{}}},
	}
	for _, r := range reqs {
		s, err := GetABCIRequestString(r)
		if err != nil {
			t.Error(err)
		}
		rr, err := GetABCIRequestFromString(s)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(r, rr) {
			t.Errorf("Logging unsuccessful: got %v expected %v\n", rr, r)
		}

	}
}
