package grammar

import (
	"fmt"
	"os"
	"strings"

	"github.com/cometbft/cometbft/libs/log"

	abci "github.com/cometbft/cometbft/abci/types"
	clean_start_lexer "github.com/cometbft/cometbft/test/e2e/pkg/grammar/clean-start/lexer"
	clean_start_parser "github.com/cometbft/cometbft/test/e2e/pkg/grammar/clean-start/parser"
	recovery_lexer "github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/lexer"
	recovery_parser "github.com/cometbft/cometbft/test/e2e/pkg/grammar/recovery/parser"
)

// GrammarChecker is a checker that can verify whether a specific set of ABCI calls
// respects the ABCI grammar.
type GrammarChecker struct {
	logger log.Logger
	cfg    *Config
}

// Config allows for setting some parameters, mostly about error logging.
type Config struct {
	// The number of errors checker outputs.
	NumberOfErrorsToShow int
}

// DefaultConfig returns a default config for GrammarChecker.
func DefaultConfig() *Config {
	return &Config{
		NumberOfErrorsToShow: 1,
	}
}

// Error represents the error type checker returns.
type Error struct {
	description string
	// height maps 1-to-1 to the consensus height only in case of clean-start. If it is recovery, it corresponds to the height number after recovery: height 0 is
	// first consensus height after recovery.
	height int
}

// String returns string representation of an error.
func (e *Error) String() string {
	s := fmt.Sprintf("The error: %q has occurred at height %v.", e.description, e.height)
	return s
}

// NewGrammarChecker returns a grammar checker object.
func NewGrammarChecker(cfg *Config) *GrammarChecker {
	return &GrammarChecker{
		cfg:    cfg,
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// isSupportedByGrammar returns true for all requests supported by the current grammar ("/pkg/grammar/clean-start/abci_grammar_clean_start.md" and "/pkg/grammar/recovery/abci_grammar_recovery.md").
// This method needs to be modified if we add another ABCI call.
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

// filterRequests returns requests supported by grammar and remove the last height.
func (g *GrammarChecker) filterRequests(reqs []*abci.Request) []*abci.Request {
	var r []*abci.Request
	for _, req := range reqs {
		if g.isSupportedByGrammar(req) {
			r = append(r, req)
		}
	}
	r, _ = g.filterLastHeight(r)
	return r
}

// filterLastHeight removes ABCI requests from the last height if "commit" has not been called
// and returns the tuple (remaining(non-filtered) requests, # of filtered requests).
func (g *GrammarChecker) filterLastHeight(reqs []*abci.Request) ([]*abci.Request, int) {
	if len(reqs) == 0 {
		return nil, 0
	}
	pos := len(reqs) - 1
	cnt := 0
	// Find the last commit.
	for pos > 0 && g.getRequestTerminal(reqs[pos]) != "commit" {
		pos--
		cnt++
	}
	return reqs[:pos+1], cnt
}

// getRequestTerminal returns a value of a corresponding terminal in the ABCI grammar for a specific request.
func (g *GrammarChecker) getRequestTerminal(req *abci.Request) string {
	// req.String() produces an output like this "init_chain:<time:<seconds:-62135596800 > >"
	// we take just the part before the ":" (init_chain, in previous example) for each request
	s := req.String()
	t := strings.Split(s, ":")[0]
	return t
}

// getExecutionString returns a string of terminal symbols in parser readable format.
func (g *GrammarChecker) getExecutionString(reqs []*abci.Request) string {
	s := ""
	for _, r := range reqs {
		t := g.getRequestTerminal(r)
		// We ensure to have one height per line for readability.
		if t == "commit" {
			s += t + "\n"
		} else {
			s += t + " "
		}
	}
	return s
}

// Verify verifies whether a list of request satisfy ABCI grammar.
func (g *GrammarChecker) Verify(reqs []*abci.Request, isCleanStart bool) (bool, error) {
	if len(reqs) == 0 {
		return false, fmt.Errorf("execution with no ABCI calls.")
	}
	r := g.filterRequests(reqs)
	// Check if the execution is incomplete.
	if len(r) == 0 {
		return true, nil
	}
	var errors []*Error
	execution := g.getExecutionString(r)
	if isCleanStart {
		errors = g.verifyCleanStart(execution)
	} else {
		errors = g.verifyRecovery(execution)
	}
	if errors == nil {
		return true, nil
	}
	return false, fmt.Errorf("%v\nFull execution:\n%v", g.combineErrors(errors, g.cfg.NumberOfErrorsToShow), g.addHeightNumbersToTheExecution(execution))
}

// verifyCleanStart verifies if a specific execution is a valid clean-start execution.
func (g *GrammarChecker) verifyCleanStart(execution string) []*Error {
	var errors []*Error
	lexer := clean_start_lexer.New([]rune(execution))
	_, errs := clean_start_parser.Parse(lexer)
	for _, err := range errs {
		exp := []string{}
		for _, ex := range err.Expected {
			exp = append(exp, ex)
		}
		expectedTokens := strings.Join(exp, ",")
		unexpectedToken := err.Token.TypeID()
		e := &Error{
			description: fmt.Sprintf("Invalid clean-start execution: parser was expecting one of [%v], got [%v] instead.", expectedTokens, unexpectedToken),
			height:      err.Line - 1,
		}
		errors = append(errors, e)
	}
	return errors
}

// verifyRecovery verifies if a specific execution is a valid recovery execution.
func (g *GrammarChecker) verifyRecovery(execution string) []*Error {
	var errors []*Error
	lexer := recovery_lexer.New([]rune(execution))
	_, errs := recovery_parser.Parse(lexer)
	for _, err := range errs {
		exp := []string{}
		for _, ex := range err.Expected {
			exp = append(exp, ex)
		}
		expectedTokens := strings.Join(exp, ",")
		unexpectedToken := err.Token.TypeID()
		e := &Error{
			description: fmt.Sprintf("Invalid recovery execution: parser was expecting one of [%v], got [%v] instead.", expectedTokens, unexpectedToken),
			height:      err.Line - 1,
		}
		errors = append(errors, e)
	}
	return errors
}

// addHeightNumbersToTheExecution adds height numbers to the execution. This is used just when printing the execution so we can find the height with error more easily.
func (g *GrammarChecker) addHeightNumbersToTheExecution(execution string) string {
	heights := strings.Split(execution, "\n")
	s := ""
	for i, l := range heights {
		if l != "" {
			s = fmt.Sprintf("%v%v: %v\n", s, i, l)
		}
	}
	return s
}

// combineErrors combines at most n errors in one.
func (g *GrammarChecker) combineErrors(errors []*Error, n int) error {
	s := ""
	for i, e := range errors {
		if i == n {
			break
		}
		s = fmt.Sprintf("%v%v\n", s, e)
	}
	return fmt.Errorf("%v", s)
}
