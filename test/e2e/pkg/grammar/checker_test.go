package grammar

import (
	"fmt"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	initChain       = &abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.InitChainRequest{}}}
	finalizeBlock   = &abci.Request{Value: &abci.Request_FinalizeBlock{FinalizeBlock: &abci.FinalizeBlockRequest{}}}
	commit          = &abci.Request{Value: &abci.Request_Commit{Commit: &abci.CommitRequest{}}}
	offerSnapshot   = &abci.Request{Value: &abci.Request_OfferSnapshot{OfferSnapshot: &abci.OfferSnapshotRequest{}}}
	applyChunk      = &abci.Request{Value: &abci.Request_ApplySnapshotChunk{ApplySnapshotChunk: &abci.ApplySnapshotChunkRequest{}}}
	prepareProposal = &abci.Request{Value: &abci.Request_PrepareProposal{PrepareProposal: &abci.PrepareProposalRequest{}}}
	processProposal = &abci.Request{Value: &abci.Request_ProcessProposal{ProcessProposal: &abci.ProcessProposalRequest{}}}
)

const (
	CleanStart = true
	Pass       = true
	Fail       = false
)

func TestVerify(t *testing.T) {
	tests := []struct {
		name         string
		abciCalls    []*abci.Request
		isCleanStart bool
		result       bool
	}{
		// start = clean-start
		// clean-start = init-chain consensus-exec
		// consensus-height = finalizeBlock commit
		{"empty-block-1", []*abci.Request{initChain, finalizeBlock, commit}, CleanStart, Pass},
		{"consensus-exec-missing", []*abci.Request{initChain}, CleanStart, Fail},
		{"finalize-block-missing-1", []*abci.Request{initChain, commit}, CleanStart, Fail},
		{"commit-missing-1", []*abci.Request{initChain, finalizeBlock}, CleanStart, Fail},
		// consensus-height = *consensus-round finalizeBlock commit
		{"proposer-round-1", []*abci.Request{initChain, prepareProposal, processProposal, finalizeBlock, commit}, CleanStart, Pass},
		{"proposer-round-2", []*abci.Request{initChain, prepareProposal, finalizeBlock, commit}, CleanStart, Pass},
		{"non-proposer-round-1", []*abci.Request{initChain, processProposal, finalizeBlock, commit}, CleanStart, Pass},
		{"multiple-rounds-1", []*abci.Request{initChain, prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, CleanStart, Pass},

		// clean-start = state-sync consensus-exec
		// state-sync = success-sync
		{"one-apply-chunk-1", []*abci.Request{offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"multiple-apply-chunks-1", []*abci.Request{offerSnapshot, applyChunk, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"offer-snapshot-missing-1", []*abci.Request{applyChunk, finalizeBlock, commit}, CleanStart, Fail},
		{"apply-chunk-missing", []*abci.Request{offerSnapshot, finalizeBlock, commit}, CleanStart, Fail},
		// state-sync = *state-sync-attempt success-sync
		{"one-apply-chunk-2", []*abci.Request{offerSnapshot, applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"multiple-apply-chunks-2", []*abci.Request{offerSnapshot, applyChunk, applyChunk, applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"offer-snapshot-missing-2", []*abci.Request{applyChunk, offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Fail},
		{"no-apply-chunk", []*abci.Request{offerSnapshot, offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Pass},

		{"init-chain+state-sync", []*abci.Request{initChain, offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Fail},
		{"no-init-chain+state-sync", []*abci.Request{finalizeBlock, commit}, CleanStart, Fail},

		// start = recovery

		// recovery = init-chain consensus-exec
		// consensus-height = finalizeBlock commit
		{"empty-block-2", []*abci.Request{initChain, finalizeBlock, commit}, !CleanStart, Pass},
		{"finalize-block-missing-2", []*abci.Request{initChain, commit}, !CleanStart, Fail},
		{"commit-missing-2", []*abci.Request{initChain, finalizeBlock}, !CleanStart, Fail},
		// consensus-height = *consensus-round finalizeBlock commit
		{"proposer-round-3", []*abci.Request{initChain, prepareProposal, processProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"proposer-round-4", []*abci.Request{initChain, prepareProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"non-proposer-round-2", []*abci.Request{initChain, processProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"multiple-rounds-2", []*abci.Request{initChain, prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, !CleanStart, Pass},

		// recovery = consensus-exec
		// consensus-height = finalizeBlock commit
		{"empty-block-3", []*abci.Request{finalizeBlock, commit}, !CleanStart, Pass},
		{"finalize-block-missing-3", []*abci.Request{commit}, !CleanStart, Fail},
		{"commit-missing-3", []*abci.Request{finalizeBlock}, !CleanStart, Fail},
		// consensus-height = *consensus-round finalizeBlock commit
		{"proposer-round-4", []*abci.Request{prepareProposal, processProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"proposer-round-5", []*abci.Request{prepareProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"non-proposer-round-3", []*abci.Request{processProposal, finalizeBlock, commit}, !CleanStart, Pass},
		{"multiple-rounds-3", []*abci.Request{prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, !CleanStart, Pass},

		// corner cases
		{"empty execution", nil, CleanStart, Fail},
		{"empty execution", nil, !CleanStart, Fail},
	}

	for _, test := range tests {
		checker := NewGrammarChecker(DefaultConfig())
		result, err := checker.Verify(test.abciCalls, test.isCleanStart)
		if result == test.result {
			continue
		}
		if err == nil {
			err = fmt.Errorf("grammar parsed an incorrect execution: %v", checker.getExecutionString(test.abciCalls))
		}
		t.Errorf("Test %v returned %v, expected %v\n%v\n", test.name, result, test.result, err)
	}
}

func TestFilterLastHeight(t *testing.T) {
	reqs := []*abci.Request{initChain, finalizeBlock, commit}
	checker := NewGrammarChecker(DefaultConfig())
	rr, n := checker.filterLastHeight(reqs)
	require.Equal(t, len(reqs), len(rr))
	require.Zero(t, n)

	reqs = append(reqs, finalizeBlock)
	rrr, n := checker.filterLastHeight(reqs)
	require.Equal(t, len(rr), len(rrr))
	require.Equal(t, n, 1)
}
