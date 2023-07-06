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
	{"empty-block-1", []string{InitChain, FinalizeBlock, Commit}, true},
	{"finalize-block-missing-1", []string{InitChain, Commit}, false},
	{"commit-missing-1", []string{InitChain, FinalizeBlock}, false},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-1", []string{InitChain, PrepareProposal, ProcessProposal, FinalizeBlock, Commit}, true},
	{"proposer-round-2", []string{InitChain, PrepareProposal, FinalizeBlock, Commit}, true},
	{"non-proposer-round-1", []string{InitChain, ProcessProposal, FinalizeBlock, Commit}, true},
	{"multiple-rounds-1", []string{InitChain, PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, FinalizeBlock, Commit}, true},

	// clean-start = init-chain state-sync consensus-exec
	// state-sync = success-sync
	{"one-apply-chunk-1", []string{InitChain, OfferSnapshot, ApplyChunk, FinalizeBlock, Commit}, true},
	{"multiple-apply-chunks-1", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, FinalizeBlock, Commit}, true},
	{"offer-snapshot-missing-1", []string{InitChain, ApplyChunk, FinalizeBlock, Commit}, false},
	{"apply-chunk-missing", []string{InitChain, OfferSnapshot, FinalizeBlock, Commit}, false},
	// state-sync = *state-sync-attempt success-sync
	{"one-apply-chunk-2", []string{InitChain, OfferSnapshot, ApplyChunk, OfferSnapshot, ApplyChunk, FinalizeBlock, Commit}, true},
	{"mutliple-apply-chunks-2", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, ApplyChunk, OfferSnapshot, ApplyChunk, FinalizeBlock, Commit}, true},
	{"offer-snapshot-missing-2", []string{InitChain, ApplyChunk, OfferSnapshot, ApplyChunk, FinalizeBlock, Commit}, false},
	{"no-apply-chunk", []string{InitChain, OfferSnapshot, OfferSnapshot, ApplyChunk, FinalizeBlock, Commit}, true},

	// start = recovery
	// recovery = consensus-exec
	// consensus-height = decide commit
	{"empty-block-2", []string{FinalizeBlock, Commit}, true},
	{"finalize-block-missing-2", []string{Commit}, false},
	{"commit-missing-2", []string{FinalizeBlock}, false},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-3", []string{PrepareProposal, ProcessProposal, FinalizeBlock, Commit}, true},
	{"proposer-round-4", []string{PrepareProposal, FinalizeBlock, Commit}, true},
	{"non-proposer-round-2", []string{ProcessProposal, FinalizeBlock, Commit}, true},
	{"multiple-rounds-2", []string{PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, FinalizeBlock, Commit}, true},

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
	execution := InitChain + " " + FinalizeBlock + "" + Commit
	_, err := checker.VerifyExecution(execution)
	if err != nil {
		t.Error(err)
	}
}

func TestFilterLastHeight(t *testing.T) {
	reqs := []*abci.Request{
		&abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}},
		&abci.Request{Value: &abci.Request_FinalizeBlock{FinalizeBlock: &abci.RequestFinalizeBlock{}}},
		&abci.Request{Value: &abci.Request_Commit{Commit: &abci.RequestCommit{}}},
	}
	checker := NewGrammarChecker(DefaultConfig())
	rr, n := checker.filterLastHeight(reqs)
	if len(reqs) != len(rr) || n != 0 {
		t.Errorf("FilterLastHeight filtered %v abci calls, expected %v\n", n, 0)
	}

	reqs = append(reqs, &abci.Request{Value: &abci.Request_FinalizeBlock{FinalizeBlock: &abci.RequestFinalizeBlock{}}})
	rrr, n := checker.filterLastHeight(reqs)
	if len(rr) != len(rrr) || n != 1 {
		t.Errorf("FilterLastHeight filtered %v abci calls, expected %v\n", n, 1)
	}
}
