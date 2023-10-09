package grammar

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
)

type Test struct {
	name         string
	abciCalls    []*abci.Request
	isCleanStart bool
	result       bool
}

var (
	initChain       = &abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}}
	finalizeBlock   = &abci.Request{Value: &abci.Request_FinalizeBlock{FinalizeBlock: &abci.RequestFinalizeBlock{}}}
	commit          = &abci.Request{Value: &abci.Request_Commit{Commit: &abci.RequestCommit{}}}
	offerSnapshot   = &abci.Request{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &abci.RequestOfferSnapshot{}}}
	applyChunk      = &abci.Request{Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &abci.RequestApplySnapshotChunk{}}}
	prepareProposal = &abci.Request{Value: &abci.Request_PrepareProposal{PrepareProposal: &abci.RequestPrepareProposal{}}}
	processProposal = &abci.Request{Value: &abci.Request_ProcessProposal{ProcessProposal: &abci.RequestProcessProposal{}}}
)

const (
	CLEAN_START = true
	PASS        = true
	FAIL        = false
)

var tests = []Test{
	// start = clean-start
	// clean-start = init-chain consensus-exec
	// consensus-height = finalizeBlock commit
	{"empty-block-1", []*abci.Request{initChain, finalizeBlock, commit}, CLEAN_START, PASS},
	{"consensus-exec-missing", []*abci.Request{initChain}, CLEAN_START, FAIL},
	{"finalize-block-missing-1", []*abci.Request{initChain, commit}, CLEAN_START, FAIL},
	{"commit-missing-1", []*abci.Request{initChain, finalizeBlock}, CLEAN_START, FAIL},
	// consensus-height = *consensus-round finalizeBlock commit
	{"proposer-round-1", []*abci.Request{initChain, prepareProposal, processProposal, finalizeBlock, commit}, CLEAN_START, PASS},
	{"proposer-round-2", []*abci.Request{initChain, prepareProposal, finalizeBlock, commit}, CLEAN_START, PASS},
	{"non-proposer-round-1", []*abci.Request{initChain, processProposal, finalizeBlock, commit}, CLEAN_START, PASS},
	{"multiple-rounds-1", []*abci.Request{initChain, prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, CLEAN_START, PASS},

	// clean-start = init-chain state-sync consensus-exec
	// state-sync = success-sync
	{"one-apply-chunk-1", []*abci.Request{initChain, offerSnapshot, applyChunk, finalizeBlock, commit}, CLEAN_START, PASS},
	{"multiple-apply-chunks-1", []*abci.Request{initChain, offerSnapshot, applyChunk, applyChunk, finalizeBlock, commit}, CLEAN_START, PASS},
	{"offer-snapshot-missing-1", []*abci.Request{initChain, applyChunk, finalizeBlock, commit}, CLEAN_START, FAIL},
	{"apply-chunk-missing", []*abci.Request{initChain, offerSnapshot, finalizeBlock, commit}, CLEAN_START, FAIL},
	// state-sync = *state-sync-attempt success-sync
	{"one-apply-chunk-2", []*abci.Request{initChain, offerSnapshot, applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CLEAN_START, PASS},
	{"mutliple-apply-chunks-2", []*abci.Request{initChain, offerSnapshot, applyChunk, applyChunk, applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CLEAN_START, PASS},
	{"offer-snapshot-missing-2", []*abci.Request{initChain, applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CLEAN_START, FAIL},
	{"no-apply-chunk", []*abci.Request{initChain, offerSnapshot, offerSnapshot, applyChunk, finalizeBlock, commit}, CLEAN_START, PASS},

	// start = recovery
	// recovery = consensus-exec
	// consensus-height = finalizeBlock commit
	{"empty-block-2", []*abci.Request{finalizeBlock, commit}, !CLEAN_START, PASS},
	{"finalize-block-missing-2", []*abci.Request{commit}, !CLEAN_START, FAIL},
	{"commit-missing-2", []*abci.Request{finalizeBlock}, !CLEAN_START, FAIL},
	// consensus-height = *consensus-round finalizeBlock commit
	{"proposer-round-3", []*abci.Request{prepareProposal, processProposal, finalizeBlock, commit}, !CLEAN_START, PASS},
	{"proposer-round-4", []*abci.Request{prepareProposal, finalizeBlock, commit}, !CLEAN_START, PASS},
	{"non-proposer-round-2", []*abci.Request{processProposal, finalizeBlock, commit}, !CLEAN_START, PASS},
	{"multiple-rounds-2", []*abci.Request{prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, !CLEAN_START, PASS},

	// corner cases
	{"empty execution", nil, CLEAN_START, FAIL},
	{"empty execution", nil, !CLEAN_START, FAIL},
}

func TestVerify(t *testing.T) {
	for _, test := range tests {
		checker := NewGrammarChecker(DefaultConfig())
		result, err := checker.Verify(test.abciCalls, test.isCleanStart)
		if result == test.result {
			continue
		}
		if err == nil {
			err = fmt.Errorf("Grammar parsed an incorrect execution: %v\n", checker.getExecutionString(test.abciCalls))
		}
		t.Errorf("Test %v returned %v, expected %v\n%v\n", test.name, result, test.result, err)
	}
}

func TestFilterLastHeight(t *testing.T) {
	reqs := []*abci.Request{initChain, finalizeBlock, commit}
	checker := NewGrammarChecker(DefaultConfig())
	rr, n := checker.filterLastHeight(reqs)
	require.Equal(t, len(reqs),len(rr), "FilterLastHeight check failed with filtered ABCI calls")
	require.Zero(t, n, "Check failed with filtered ABCI calls")

	reqs = append(reqs, finalizeBlock)
	rrr, n := checker.filterLastHeight(reqs)
	if len(rr) != len(rrr) || n != 1 {
		t.Errorf("FilterLastHeight filtered %v ABCI calls, expected %v\n", n, 1)
	}
}
