package e2e

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/lexer"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/parser"
)

// abciGrammar terminals from "/pkg/grammar/abci_grammar.md" file.
const (
	InitChain       = "1"
	BeginBlock      = "2"
	DeliverTx       = "3"
	EndBlock        = "4"
	Commit          = "5"
	OfferSnapshot   = "6"
	ApplyChunk      = "7"
	PrepareProposal = "8"
	ProcessProposal = "9"
)

// getRequestTerminal returns a value of a corresponding terminal in the abci grammar for a specific request.
func getRequestTerminal(req *abci.Request) string {
	switch req.Value.(type) {
	case *abci.Request_InitChain:
		return InitChain
	case *abci.Request_BeginBlock:
		return BeginBlock
	case *abci.Request_DeliverTx:
		return DeliverTx
	case *abci.Request_EndBlock:
		return EndBlock
	case *abci.Request_Commit:
		return Commit
	case *abci.Request_OfferSnapshot:
		return OfferSnapshot
	case *abci.Request_ApplySnapshotChunk:
		return ApplyChunk
	case *abci.Request_PrepareProposal:
		return PrepareProposal
	case *abci.Request_ProcessProposal:
		return ProcessProposal
	default:
		return ""

	}
}

func GetExecutionString(reqs []*abci.Request) string {
	s := ""
	for _, r := range reqs {
		t := getRequestTerminal(r)
		s += " " + t
	}
	return s
}

func Verify(execution string) (bool, error) {
	lexer := lexer.New([]rune(execution))
	_, errs := parser.Parse(lexer)
	if len(errs) > 0 {
		err := combineParseErrors(execution, errs)
		return false, err
	}
	return true, nil

}

func combineParseErrors(execution string, errs []*parser.Error) error {
	s := fmt.Sprintf("Parser failed\nExecution:%v\n", execution)
	for _, e := range errs {
		s = fmt.Sprintf("%v%v\n", s, e)
	}
	return fmt.Errorf("%v\n", s)
}
