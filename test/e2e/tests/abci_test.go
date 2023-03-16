package e2e_test

import (
	"fmt"
	"strings"
	"testing"

	abci_call "github.com/cometbft/cometbft/test/e2e/abci"
	"github.com/cometbft/cometbft/test/e2e/abci/lexer"
	"github.com/cometbft/cometbft/test/e2e/abci/parser"
)

func TestABCIGrammar(t *testing.T) {
	testnet := loadTestnet(t)
	abciCalls := loadAbciCalls(testnet.Dir)
	for _, node := range testnet.Nodes {
		s := ""
		for _, a := range abciCalls[node.Name] {
			parts := strings.Split(a.ToString(), ":")
			s = fmt.Sprintf("%v %v", s, parts[1])
		}
		lexer := lexer.New([]rune(s))

		_, errs := parser.Parse(lexer)
		if len(errs) > 0 {
			t.Errorf("Execution on node %v does not respect grammar\n", node.Name)
		}
	}
}

func TestABCISpecific(t *testing.T) {
	testnet := loadTestnet(t)
	abciCalls := loadAbciCalls(testnet.Dir)
	for _, node := range testnet.Nodes {
		cnt := 0
		for _, a := range abciCalls[node.Name] {
			if a.Type == abci_call.BeginBlock {
				cnt++
			}
			if a.Type == abci_call.EndBlock {
				cnt--
			}
			if cnt != 1 && cnt != 0 {
				t.Errorf("BeginBlock and EndBlock are not called in pair!")
			}
		}
	}
}

/*
func TestABCI_II(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		cnt := 0
		for _, a := range node.ABCICalls {
			if a.Type == e2e.BeginBlock {
				cnt++
			}
			if a.Type == e2e.EndBlock {
				cnt--
			}
			if cnt != 1 && cnt != 0 {
				t.Errorf("BeginBlock and EndBlock are not called in pair!")
			}
		}
	})
}
*/
