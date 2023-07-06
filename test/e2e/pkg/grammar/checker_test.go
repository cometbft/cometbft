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

// abciGrammar terminals from "/pkg/grammar/abci_grammar.md" file.
const (
	InitChain       = "init_chain"
	Decide          = "finalize_block"
	Commit          = "commit"
	OfferSnapshot   = "offer_snapshot"
	ApplyChunk      = "apply_snapshot_chunk"
	PrepareProposal = "prepare_proposal"
	ProcessProposal = "process_proposal"
)

var tests = []Test{
	// start = clean-start
	// clean-start = init-chain consensus-exec
	// consensus-height = decide commit
	{"consensus-exec-missing", []string{InitChain}, false},
	{"empty-block-1", []string{InitChain, Decide, Commit}, true},
	{"finalize-block-missing-1", []string{InitChain, Commit}, false},
	{"commit-missing-1", []string{InitChain, Decide}, false},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-1", []string{InitChain, PrepareProposal, ProcessProposal, Decide, Commit}, true},
	{"proposer-round-2", []string{InitChain, PrepareProposal, Decide, Commit}, true},
	{"non-proposer-round-1", []string{InitChain, ProcessProposal, Decide, Commit}, true},
	{"multiple-rounds-1", []string{InitChain, PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, Decide, Commit}, true},

	// clean-start = init-chain state-sync consensus-exec
	// state-sync = success-sync
	{"one-apply-chunk-1", []string{InitChain, OfferSnapshot, ApplyChunk, Decide, Commit}, true},
	{"multiple-apply-chunks-1", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, Decide, Commit}, true},
	{"offer-snapshot-missing-1", []string{InitChain, ApplyChunk, Decide, Commit}, false},
	{"apply-chunk-missing", []string{InitChain, OfferSnapshot, Decide, Commit}, false},
	// state-sync = *state-sync-attempt success-sync
	{"one-apply-chunk-2", []string{InitChain, OfferSnapshot, ApplyChunk, OfferSnapshot, ApplyChunk, Decide, Commit}, true},
	{"mutliple-apply-chunks-2", []string{InitChain, OfferSnapshot, ApplyChunk, ApplyChunk, ApplyChunk, OfferSnapshot, ApplyChunk, Decide, Commit}, true},
	{"offer-snapshot-missing-2", []string{InitChain, ApplyChunk, OfferSnapshot, ApplyChunk, Decide, Commit}, false},
	{"no-apply-chunk", []string{InitChain, OfferSnapshot, OfferSnapshot, ApplyChunk, Decide, Commit}, true},

	// start = recovery
	// recovery = consensus-exec
	// consensus-height = decide commit
	{"empty-block-2", []string{Decide, Commit}, true},
	{"finalize-block-missing-2", []string{Commit}, false},
	{"commit-missing-2", []string{Decide}, false},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-3", []string{PrepareProposal, ProcessProposal, Decide, Commit}, true},
	{"proposer-round-4", []string{PrepareProposal, Decide, Commit}, true},
	{"non-proposer-round-2", []string{ProcessProposal, Decide, Commit}, true},
	{"multiple-rounds-2", []string{PrepareProposal, ProcessProposal, ProcessProposal, PrepareProposal, ProcessProposal, ProcessProposal, ProcessProposal, Decide, Commit}, true},

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
