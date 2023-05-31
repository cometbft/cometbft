package grammar

import (
	"fmt"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
)

type Test struct {
	name      string
	abciCalls []string
	result    bool
}

var tests = []Test{
	// start = clean-start
	// clean-start = init-chain consensus-exec
	// consensus-height = decide commit
	{"consensus-exec-missing", []string{InitChain}, false},
	{"empty-block-1", []string{InitChain, BeginBlock, EndBlock, Commit}, true},
	{"begin-block-missing-1", []string{InitChain, EndBlock, Commit}, false},
	{"end-block-missing-1", []string{InitChain, BeginBlock, Commit}, false},
	{"commit-missing-1", []string{InitChain, BeginBlock, EndBlock}, false},
	{"one-tx-block-1", []string{InitChain, BeginBlock, DeliverTx, EndBlock, Commit}, true},
	{"multiple-tx-block-1", []string{InitChain, BeginBlock, DeliverTx, DeliverTx, EndBlock, Commit}, true},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-1", []string{InitChain, PrepareProposal, ProcessProposal, BeginBlock, EndBlock, Commit}, true},
	{"process-proposal-missing-1", []string{InitChain, PrepareProposal, BeginBlock, EndBlock, Commit}, false},
	{"non-proposer-round-1", []string{InitChain, ProcessProposal, BeginBlock, EndBlock, Commit}, true},
	{"multiple-rounds-1", []string{InitChain, PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, BeginBlock, EndBlock, Commit}, true},

	// clean-start = init-chain state-sync consensus-exec
	// state-sync = success-sync
	{"one-apply-chunk-1", []string{InitChain, OfferSnapshot, ApplyChunk, BeginBlock, EndBlock, Commit}, true},
	{"multiple-apply-chunks-1", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, BeginBlock, EndBlock, Commit}, true},
	{"offer-snapshot-missing-1", []string{InitChain, ApplyChunk, BeginBlock, EndBlock, Commit}, false},
	{"apply-chunk-missing", []string{InitChain, OfferSnapshot, BeginBlock, EndBlock, Commit}, false},
	// state-sync = *state-sync-attempt success-sync
	{"one-apply-chunk-2", []string{InitChain, OfferSnapshot, ApplyChunk, OfferSnapshot, ApplyChunk, BeginBlock, EndBlock, Commit}, true},
	{"mutliple-apply-chunks-2", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, ApplyChunk, OfferSnapshot, ApplyChunk, BeginBlock, EndBlock, Commit}, true},
	{"offer-snapshot-missing-2", []string{InitChain, ApplyChunk, OfferSnapshot, ApplyChunk, BeginBlock, EndBlock, Commit}, false},
	{"no-apply-chunk", []string{InitChain, OfferSnapshot, OfferSnapshot, ApplyChunk, BeginBlock, EndBlock, Commit}, true},

	// start = recovery
	// recovery = consensus-exec
	// consensus-height = decide commit
	{"empty-block-2", []string{BeginBlock, EndBlock, Commit}, true},
	{"begin-block-missing-2", []string{EndBlock, Commit}, false},
	{"end-block-missing-2", []string{BeginBlock, Commit}, false},
	{"commit-missing-2", []string{BeginBlock, EndBlock}, false},
	{"one-tx-block-2", []string{BeginBlock, DeliverTx, EndBlock, Commit}, true},
	{"multiple-tx-block-2", []string{BeginBlock, DeliverTx, DeliverTx, EndBlock, Commit}, true},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-2", []string{PrepareProposal, ProcessProposal, BeginBlock, EndBlock, Commit}, true},
	{"process-proposal-missing-2", []string{PrepareProposal, BeginBlock, EndBlock, Commit}, false},
	{"non-proposer-round-2", []string{ProcessProposal, BeginBlock, EndBlock, Commit}, true},
	{"multiple-rounds-2", []string{PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, BeginBlock, EndBlock, Commit}, true},

	// corner cases
	{"empty execution", []string{""}, false},
}

func TestVerify(t *testing.T) {
	for _, test := range tests {
		checker := NewGrammarChecker(DefaultConfig())
		execution := strings.Join(test.abciCalls, " ")
		result, err := checker.VerifyExecution(execution)
		if result == test.result {
			continue
		}
		if err == nil {
			err = fmt.Errorf("Grammar parsed an incorrect execution: %v\n", execution)
		}
		t.Errorf("Test %v returned %v, expected %v\n Error: %v\n", test.name, result, test.result, err)
	}
}

func TestVerifySpecific(t *testing.T) {
	//calls := []string{PrepareProposal, BeginBlock, EndBlock, Commit}
	//execution := strings.Join(calls, " ")
	checker := NewGrammarChecker(DefaultConfig())
	//execution := InitChain + " " + BeginBlock + "" + EndBlock
	execution := InitChain + " " + BeginBlock + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + "\n" + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + "\n" + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + "\n" + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + Commit + " " + BeginBlock + "\n" + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + DeliverTx + "\n" + DeliverTx + " " + DeliverTx + " " + DeliverTx + " " + EndBlock + " " + Commit
	_, err := checker.VerifyExecution(execution)
	if err != nil {
		t.Error(err)
	}
}

func TestFilterLastHeight(t *testing.T) {
	reqs := []*abci.Request{
		&abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}},
		&abci.Request{Value: &abci.Request_BeginBlock{BeginBlock: &abci.RequestBeginBlock{}}},
		&abci.Request{Value: &abci.Request_EndBlock{EndBlock: &abci.RequestEndBlock{}}},
		&abci.Request{Value: &abci.Request_Commit{Commit: &abci.RequestCommit{}}},
	}
	checker := NewGrammarChecker(DefaultConfig())
	rr, n := checker.filterLastHeight(reqs)
	if len(reqs) != len(rr) || n != 0 {
		t.Errorf("FilterLastHeight filtered %v abci calls, expected %v\n", n, 0)
	}

	reqs = append(reqs, &abci.Request{Value: &abci.Request_BeginBlock{BeginBlock: &abci.RequestBeginBlock{}}})
	rrr, n := checker.filterLastHeight(reqs)
	if len(rr) != len(rrr) || n != 1 {
		t.Errorf("FilterLastHeight filtered %v abci calls, expected %v\n", n, 1)
	}
}
