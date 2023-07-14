package e2e_test

import (
	"fmt"
	"testing"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar"
)

func TestABCIGrammar(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	checker := grammar.NewGrammarChecker(grammar.DefaultConfig())
	testNode(t, func(t *testing.T, node e2e.Node) {
		if !node.Testnet.ABCITestsEnabled {
			return
		}
		reqs := m[node.Name]
		_, err := checker.Verify(reqs)
		if err != nil {
			t.Error(fmt.Errorf("ABCI grammar verification failed: %w", err))
		}
	})
}

func TestNodeNameExtracting(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	testNode(t, func(t *testing.T, node e2e.Node) {
		if !node.Testnet.ABCITestsEnabled {
			return
		}
		_, ok := m[node.Name]
		if !ok {
			t.Errorf("Node %v is not in map.\n", node.Name)
		}
	})
}
