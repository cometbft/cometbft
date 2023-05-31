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

// abciGrammar terminals from "/pkg/grammar/abci_grammar.md" file.
const (
	InitChain       = "<InitChain>"
	BeginBlock      = "<BeginBlock>"
	DeliverTx       = "<DeliverTx>"
	EndBlock        = "<EndBlock>"
	Commit          = "<Commit>"
	OfferSnapshot   = "<OfferSnapshot>"
	ApplyChunk      = "<ApplyChunk>"
	PrepareProposal = "<PrepareProposal>"
	ProcessProposal = "<ProcessProposal>"
)

// GrammarChecker is a checker that can verify whether a specific set of abci calls
// respect the abci grammar.
type GrammarChecker struct {
	logger log.Logger
	cfg    *Config
}

// Config allows for setting some parameters mostly about errors logging.
type Config struct {
	// Tell if we should filter the last height.
	FilterLastHeight bool
	// Maximum number of errors grammar outputs.
	MaxNumberOfErrorsToShow int
	// How many abci calls grammar checker prints in the same line.
	MaxTerminalsInLine int
	// Number of lines with abci calls, before the line with an error, grammar checker prints.
	NumberOfLinesToShowBeforeError int
	// Number of lines with abci calls, after the line with an error, grammar checker prints.
	NumberOfLinesToShowAfterError int
}

func DefaultConfig() *Config {
	return &Config{
		FilterLastHeight:               true,
		MaxNumberOfErrorsToShow:        10,
		MaxTerminalsInLine:             5,
		NumberOfLinesToShowBeforeError: 5,
		NumberOfLinesToShowAfterError:  5,
	}
}

// NewGrammarChecker returns a grammar checher object.
func NewGrammarChecker(cfg *Config) *GrammarChecker {
	return &GrammarChecker{
		cfg:    cfg,
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// getRequestTerminal returns a value of a corresponding terminal in the abci grammar for a specific request.
func (g *GrammarChecker) getRequestTerminal(req *abci.Request) string {
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

// GetExecutionString returns all requests that grammar understand as string of terminal symbols
// and number of requests included in it of all requests in parser readable format.
func (g *GrammarChecker) GetExecutionString(reqs []*abci.Request) (string, int) {
	s := ""
	n := 0
	for _, r := range reqs {
		t := g.getRequestTerminal(r)
		if t == "" {
			continue
		}
		// Every fifth terminal I put in new line. This is needed
		// so we can find the parsing error, if there is one, more
		// easily.
		if n != 0 && n%g.cfg.MaxTerminalsInLine == 0 {
			s += "\n" + t
		} else {
			s += " " + t
		}
		n++

	}
	return s, n
}

// Verify verifies whether a list of request satisfy abci grammar.
func (g *GrammarChecker) Verify(reqs []*abci.Request) (bool, error) {
	var r []*abci.Request
	var n int
	if g.cfg.FilterLastHeight {
		r, n = g.filterLastHeight(reqs)
		if n != 0 {
			debugMsg := fmt.Sprintf("Last height filtered, removed last %v abci calls out of %v.\n", n, len(reqs))
			g.logger.Debug(debugMsg)
		}
	}
	s, _ := g.GetExecutionString(r)
	return g.VerifyExecution(s)
}

// Verify checks if "execution string" respect abci grammar.
// Only this method is using auto-generated code by gogll.
func (g *GrammarChecker) VerifyExecution(execution string) (bool, error) {
	lexer := lexer.New([]rune(execution))
	_, errs := parser.Parse(lexer)
	if len(errs) > 0 {
		err := g.combineParseErrors(execution, errs, g.cfg.MaxNumberOfErrorsToShow)
		return false, err
	}
	return true, nil
}

// filterLastHeight removes abci calls from the last height if <Commit> has not been called.
func (g *GrammarChecker) filterLastHeight(reqs []*abci.Request) ([]*abci.Request, int) {
	pos := len(reqs) - 1
	r := reqs[pos]
	cnt := 0
	// Find the last commit.
	for g.getRequestTerminal(r) != Commit && pos >= 0 {
		pos--
		r = reqs[pos]
		cnt++
	}
	return reqs[:pos+1], cnt
}

// combineParseErrors combines all parse errors in one.
func (g *GrammarChecker) combineParseErrors(execution string, errs []*parser.Error, n int) error {
	s := fmt.Sprintf("Parser failed, number of errors is %v\n", len(errs))
	for i, e := range errs {
		if i == n {
			break
		}
		lines := g.getStringOfLinesAroundError(execution, e)
		err := fmt.Errorf("***Error %v***\n%v\nExecution:\n%v", i, e, lines)
		s = fmt.Sprintf("%v%v\n", s, err)
	}
	return fmt.Errorf("%v\n", s)
}

func (g *GrammarChecker) getStringOfLinesAroundError(execution string, err *parser.Error) string {
	lineNumber := err.Line
	lines := strings.Split(execution, "\n")
	s := ""
	for i, l := range lines {
		// parser returns line numbers starting from 1.
		index := i + 1
		firstLine := lineNumber - g.cfg.NumberOfLinesToShowBeforeError
		lastLine := lineNumber + g.cfg.NumberOfLinesToShowAfterError
		if index >= firstLine && index <= lastLine {
			s = fmt.Sprintf("%v%v:%v\n", s, i+1, l)
		}
	}
	return s
}
