package e2e_test

import (
	"fmt"
	"testing"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar"
)

func TestABCIGrammar(t *testing.T) {
	checker := grammar.NewGrammarChecker(grammar.DefaultConfig())
	testNode(t, func(t *testing.T, node e2e.Node) {
		if !node.Testnet.ABCITestsEnabled {
			return
		}
		reqs, err := fetchABCIRequests(t, node.Name)
		require.NoError(t, err)
		for i, r := range reqs {
			isCleanStart := i == 0
			_, err := checker.Verify(r, isCleanStart)
			require.NoError(t, err)
		}
	})
}

func TestNodeNameExtracting(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		if !node.Testnet.ABCITestsEnabled {
			return
		}
		reqs, err := fetchABCIRequests(t, node.Name)
		if err != nil {
			t.Error(fmt.Errorf("Collecting of ABCI requests failed: %w", err))
		}
		if len(reqs) == 0 {
			t.Errorf("No ABCI requests on node %v", node.Name)
		}
	})
}
