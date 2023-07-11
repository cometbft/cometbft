package grammar

import (
	"fmt"
	"os"
	"strings"

	"github.com/cometbft/cometbft/libs/log"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/lexer"
	"github.com/cometbft/cometbft/test/e2e/pkg/grammar/parser"
)

// GrammarChecker is a checker that can verify whether a specific set of abci calls
// respect the abci grammar.
type GrammarChecker struct {
	logger log.Logger
	cfg    *Config
}

// Config allows for setting some parameters mostly about errors logging.
type Config struct {
	// Number of errors checker outputs.
	NumberOfErrorsToShow int
	// Show full execution.
	ShowFullExecution bool
}

func DefaultConfig() *Config {
	return &Config{
		NumberOfErrorsToShow: 1,
		ShowFullExecution:    true,
	}
}

// NewGrammarChecker returns a grammar checker object.
func NewGrammarChecker(cfg *Config) *GrammarChecker {
	return &GrammarChecker{
		cfg:    cfg,
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// isSupportedByGrammar returns true for all requests supported by the current grammar in "/pkg/grammar/abci_grammar.md" file.
func (g *GrammarChecker) isSupportedByGrammar(req *abci.Request) bool {
	switch req.Value.(type) {
	case *abci.Request_InitChain, *abci.Request_FinalizeBlock, *abci.Request_Commit,
		*abci.Request_OfferSnapshot, *abci.Request_ApplySnapshotChunk, *abci.Request_PrepareProposal,
		*abci.Request_ProcessProposal:
		return true
	default:
		return false
	}
}

// getRequestTerminal returns a value of a corresponding terminal in the abci grammar for a specific request.
func (g *GrammarChecker) getRequestTerminal(req *abci.Request) string {
	if g.isSupportedByGrammar(req) {
		// req.String() produces an output like this "init_chain:<time:<seconds:-62135596800 > >"
		// we take just the part before the ":" (init_chain, in previous example) for each request
		s := req.String()
		t := strings.Split(s, ":")[0]
		return t
	}
	return ""

}

// GetExecutionString returns all requests that grammar understand as string of terminal symbols
// in parser readable format.
func (g *GrammarChecker) GetExecutionString(reqs []*abci.Request) string {
	s := ""
	for _, r := range reqs {
		t := g.getRequestTerminal(r)
		if t == "" {
			continue
		}
		// We ensure to have one height per line for readability.
		if t == "commit" {
			s += t + "\n"
		} else {
			s += t + " "
		}
	}
	return s
}

// Verify verifies whether a list of request satisfy abci grammar.
func (g *GrammarChecker) Verify(reqs []*abci.Request) (bool, error) {
	var r []*abci.Request
	r, _ = g.filterLastHeight(reqs)
	s := g.GetExecutionString(r)
	return g.VerifyExecution(s)
}

// filterLastHeight removes abci calls from the last height if "commit" has not been called.
func (g *GrammarChecker) filterLastHeight(reqs []*abci.Request) ([]*abci.Request, int) {
	pos := len(reqs) - 1
	r := reqs[pos]
	cnt := 0
	// Find the last commit.
	for g.getRequestTerminal(r) != "commit" && pos >= 0 {
		pos--
		r = reqs[pos]
		cnt++
	}
	return reqs[:pos+1], cnt
}

// Verify checks if "execution string" respect abci grammar.
// Only this method is using auto-generated code by gogll.
func (g *GrammarChecker) VerifyExecution(execution string) (bool, error) {
	lexer := lexer.New([]rune(execution))
	_, errs := parser.Parse(lexer)
	if len(errs) > 0 {
		err := g.combineParseErrors(execution, errs, g.cfg.NumberOfErrorsToShow)
		if g.cfg.ShowFullExecution {
			e := g.addHeightNumbersToTheExecution(execution)
			err = fmt.Errorf("%v\nFull execution:\n%v", err, e)
		}
		return false, err
	}
	return true, nil
}

// combineParseErrors combines all parse errors in one.
func (g *GrammarChecker) combineParseErrors(execution string, errs []*parser.Error, n int) error {
	s := fmt.Sprintf("Parser failed, number of errors is %v\n", len(errs))
	heights := strings.Split(execution, "\n")
	for i, e := range errs {
		if i == n {
			break
		}
		// e.Line-1 because the parser returns line numbers starting from 1
		h := e.Line - 1
		heightWithError := heights[h]
		exp := []string{}
		for _, ex := range e.Expected {
			exp = append(exp, ex)
		}
		err := fmt.Errorf("---Error %v---\nHeight: %v\nABCI requests: %v\nUnexpected request: %v\nExpected one of: [%v]", i, h, heightWithError, e.Token.TypeID(), strings.Join(exp, ","))
		s = fmt.Sprintf("%v%v\n", s, err)
	}
	return fmt.Errorf("%v-------------", s)
}

func (g *GrammarChecker) addHeightNumbersToTheExecution(execution string) string {
	heights := strings.Split(execution, "\n")
	s := ""
	for i, l := range heights {
		s = fmt.Sprintf("%v%v: %v\n", s, i, l)
	}
	return s
}
