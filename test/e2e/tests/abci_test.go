package e2e_test

import (
	"testing"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

func TestABCIGrammar(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	testNode(t, func(t *testing.T, node e2e.Node) {
		reqs := m[node.Name]
		e := e2e.GetExecutionString(reqs)
		_, err := e2e.Verify(e)
		if err != nil {
			t.Error(err)
		}
	})
}
