package e2e_test

import (
	"testing"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar"
)

func TestABCIGrammar(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	testNode(t, func(t *testing.T, node e2e.Node) {
		reqs := m[node.Name]
		checker := grammar.NewGrammarChecker(grammar.DefaultConfig())
		_, err := checker.Verify(reqs)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestNodeNameExtracting(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	testNode(t, func(t *testing.T, node e2e.Node) {
		_, ok := m[node.Name]
		if !ok {
			t.Errorf("Node %v is not in map.\n", node.Name)
		}
	})
}
