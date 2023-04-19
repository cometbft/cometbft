package e2e

import (
	"fmt"
	"testing"
)

type Test struct {
	name      string
	execution string
	result    bool
}

var tests = []Test{
	// start = clean-start
	// clean-start = init-chain consensus-exec
	// consensus-height = decide commit
	{"consensus-exec-missing", "1", false},
	{"empty-block-1", "1 2 4 5", true},
	{"begin-block-missing-1", "1 4 5", false},
	{"end-block-missing-1", "1 2 5", false},
	{"commit-missing-1", "1 2 4", false},
	{"one-tx-block-1", "1 2 3 4 5", true},
	{"multiple-tx-block-1", "1 2 3 3 4 5", true},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-1", "1 8 9 2 4 5", true},
	{"process-proposal-missing-1", "1 8 2 4 5", false},
	{"non-proposer-round-1", "1 9 2 4 5", true},
	{"multiple-rounds-1", "1 8 9 9 8 9 9 9 2 4 5", true},

	// clean-start = init-chain state-sync consensus-exec
	// state-sync = success-sync
	{"one-apply-chunk-1", "1 6 7 2 4 5", true},
	{"multiple-apply-chunks-1", "1 6 7 7 2 4 5", true},
	{"offer-snapshot-missing-1", "1 7 2 4 5", false},
	{"apply-chunk-missing", "1 6 2 4 5", false},
	// state-sync = *state-sync-attempt success-sync
	{"one-apply-chunk-2", "1 6 7 6 7 2 4 5", true},
	{"mutliple-apply-chunks-2", "1 6 7 7 7 6 7 2 4 5", true},
	{"offer-snapshot-missing-2", "1 7 6 7 2 4 5", false},
	{"no-apply-chunk", "1 6 6 7 2 4 5", true},

	// start = recovery
	// recovery = consensus-exec
	// consensus-height = decide commit
	{"empty-block-2", "2 4 5", true},
	{"begin-block-missing-2", "4 5", false},
	{"end-block-missing-2", "2 5", false},
	{"commit-missing-2", "2 4", false},
	{"one-tx-block-2", "2 3 4 5", true},
	{"multiple-tx-block-2", "2 3 3 4 5", true},
	// consensus-height = *consensus-round decide commit
	{"proposer-round-2", "8 9 2 4 5", true},
	{"process-proposal-missing-2", "8 2 4 5", false},
	{"non-proposer-round-2", "9 2 4 5", true},
	{"multiple-rounds-2", "8 9 9 8 9 9 9 2 4 5", true},

	// corner cases
	{"empty execution", "", false},
}

func TestVerify(t *testing.T) {
	for _, test := range tests {
		result, err := Verify(test.execution)
		if result == test.result {
			continue
		}
		if err == nil {
			err = fmt.Errorf("Grammar parsed an incorrect execution: %v\n", test.execution)
		}
		t.Errorf("Test %v returned %v, expected %v\n Error: %v\n", test.name, result, test.result, err)
	}
}
