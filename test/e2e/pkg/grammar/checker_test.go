package grammar

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
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

<<<<<<< HEAD
func TestVerify(t *testing.T) {
	tests := []struct {
		name         string
		abciCalls    []*abci.Request
		isCleanStart bool
		result       bool
	}{
=======
type ABCIExecution struct {
	abciCalls []*abci.Request
	isValid   bool
}

// consensus-exec part of executions
// consensus-exec = (inf)consensus-height
// it is part of each executions.
var consExecPart = []ABCIExecution{
	// consensus-height = finalizeBlock commit
	{[]*abci.Request{finalizeBlock, commit}, true},
	{[]*abci.Request{commit}, false},
	// consensus-height = *consensus-round finalizeBlock commit
	// consensus-height = *consensus-round finalizeBlock commit
	// consensus-round = proposer
	// proposer = *gotVote
	{[]*abci.Request{gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, finalizeBlock, commit}, true},
	// proposer = [prepare-proposal [process-proposal]]
	{[]*abci.Request{prepareProposal, processProposal, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, finalizeBlock, commit}, true},
	// proposer = [extend]
	{[]*abci.Request{extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{extendVote, gotVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, extendVote, gotVote, gotVote, finalizeBlock, commit}, true},
	// proposer = *gotVote [prepare-proposal [process-proposal]]
	{[]*abci.Request{gotVote, prepareProposal, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, prepareProposal, processProposal, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, processProposal, finalizeBlock, commit}, true},
	// proposer = *gotVote [extend]
	// same as just [extend]
	// proposer = [prepare-proposal [process-proposal]] [extend]
	{[]*abci.Request{prepareProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, processProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, processProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, processProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{prepareProposal, processProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},
	// proposer = *gotVote [prepare-proposal [process-proposal]] [extend]
	{[]*abci.Request{gotVote, prepareProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, prepareProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, prepareProposal, processProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, processProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, prepareProposal, processProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, prepareProposal, processProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},

	// consensus-round = non-proposer
	// non-proposer = *gotVote
	// same as for proposer

	// non-proposer = [process-proposal]
	{[]*abci.Request{processProposal, finalizeBlock, commit}, true},
	// non-proposer = [extend]
	// same as for proposer

	// non-proposer = *gotVote [process-proposal]
	{[]*abci.Request{gotVote, processProposal, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, processProposal, finalizeBlock, commit}, true},
	// non-proposer = *gotVote [extend]
	// same as just [extend]

	// non-proposer = [process-proposal] [extend]
	{[]*abci.Request{processProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{processProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{processProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{processProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},

	// non-proposer = *gotVote [prepare-proposal [process-proposal]] [extend]
	{[]*abci.Request{gotVote, processProposal, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, processProposal, gotVote, extendVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, processProposal, extendVote, gotVote, finalizeBlock, commit}, true},
	{[]*abci.Request{gotVote, gotVote, processProposal, gotVote, extendVote, gotVote, finalizeBlock, commit}, true},

	{[]*abci.Request{prepareProposal, processProposal, processProposal, prepareProposal, processProposal, processProposal, processProposal, finalizeBlock, commit}, true},
}

func TestVerifyCleanStart(t *testing.T) {
	// Parts of executions specific for clean-start execution
	specificCleanStartPart := []ABCIExecution{
>>>>>>> 5ee75a0a5 (fix(e2e): Fixing the bug in ABCI e2e tests (#2468))
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
<<<<<<< HEAD
		{"one-apply-chunk-1", []*abci.Request{offerSnapshot, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"multiple-apply-chunks-1", []*abci.Request{offerSnapshot, applyChunk, applyChunk, finalizeBlock, commit}, CleanStart, Pass},
		{"offer-snapshot-missing-1", []*abci.Request{applyChunk, finalizeBlock, commit}, CleanStart, Fail},
		{"apply-chunk-missing", []*abci.Request{offerSnapshot, finalizeBlock, commit}, CleanStart, Fail},
=======
		{[]*abci.Request{offerSnapshot, applyChunk}, true},
		{[]*abci.Request{offerSnapshot, applyChunk, applyChunk}, true},
		{[]*abci.Request{applyChunk}, false},
		{[]*abci.Request{offerSnapshot}, false},
>>>>>>> 5ee75a0a5 (fix(e2e): Fixing the bug in ABCI e2e tests (#2468))
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
	reqs := []*abci.Request{initChain, finalizeBlock}
	checker := NewGrammarChecker(DefaultConfig())
	r, n := checker.filterLastHeight(reqs)
	require.Equal(t, len(r), 0)
	require.Equal(t, n, 2)
	reqs = append(reqs, commit)
	r, n = checker.filterLastHeight(reqs)
	require.Equal(t, len(r), len(reqs))
	require.Zero(t, n)
<<<<<<< HEAD

	reqs = append(reqs, finalizeBlock)
	rrr, n := checker.filterLastHeight(reqs)
	require.Equal(t, len(rr), len(rrr))
	require.Equal(t, 1, n)
=======
	reqs = append(reqs, []*abci.Request{prepareProposal, processProposal}...)
	r, n = checker.filterLastHeight(reqs)
	require.Equal(t, len(r), 3)
	require.Equal(t, n, 2)
>>>>>>> 5ee75a0a5 (fix(e2e): Fixing the bug in ABCI e2e tests (#2468))
}
