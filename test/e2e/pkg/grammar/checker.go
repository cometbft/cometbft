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
	s := fmt.Sprintf("The error: \"%v\" has occured at height %v.", e.description, e.height)
	return s
}

// NewGrammarChecker returns a grammar checker object.
func NewGrammarChecker(cfg *Config) *GrammarChecker {
	return &GrammarChecker{
		cfg:    cfg,
		logger: log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
	}
}

// isSupportedByGrammar returns true for all requests supported by the current grammar in "/pkg/grammar/abci_grammar.md" file.
// this method needs to be modified if we add another ABCI call.
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

// filterLastHeight removes ABCI requests from the last height if "commit" has not been called.
func (g *GrammarChecker) filterLastHeight(reqs []*abci.Request) ([]*abci.Request, int) {
	if len(reqs) == 0 {
		return nil, 0
	}
	pos := len(reqs) - 1
	r := reqs[pos]
	cnt := 0
	// Find the last commit.
	for g.getRequestTerminal(r) != "commit" && pos > 0 {
		pos--
		r = reqs[pos]
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

// Verify verifies whether a list of request satisfy abci grammar.
func (g *GrammarChecker) Verify(reqs []*abci.Request, isCleanStart bool) (bool, error) {
	r := g.filterRequests(reqs)
	// This should not happen in our tests.
	if len(reqs) == 0 {
		return false, fmt.Errorf("Execution with no ABCI calls.")
	}
	execution := g.getExecutionString(r)
	_, err := g.verifySpecific(r, isCleanStart)
	if err != nil {
		return false, fmt.Errorf("%v\nExecution:\n%v", err, g.addHeightNumbersToTheExecution(execution))
	}
	_, errs := g.verifyGeneric(execution)
	if errs != nil {
		return false, fmt.Errorf("%v\nExecution:\n%v", g.combineErrors(errs, g.cfg.NumberOfErrorsToShow), g.addHeightNumbersToTheExecution(execution))
	}
	return true, nil
}

// verifySpecific should do all specific checks for catching differencies between clean-start and recovery. This is because verifyGeneric cannot distinguish
// if it should check whether the specific execution should respect clean-start or recovery, it returns true if any of the two is respected.
func (g *GrammarChecker) verifySpecific(reqs []*abci.Request, isCleanStart bool) (bool, *Error) {
	firstReq := g.getRequestTerminal(reqs[0])
	if isCleanStart {
		if firstReq != "init_chain" {
			err := &Error{
				description: fmt.Sprintf("Clean-start starts with %v", firstReq),
				height:      0,
			}
			return false, err
		}
	} else {
		if firstReq != "finalize_block" && firstReq != "prepare_proposal" && firstReq != "process_proposal" {
			err := &Error{
				description: fmt.Sprintf("Recovery starts with %v", firstReq),
				height:      0,
			}
			return false, err
		}
	}
	return true, nil
}

// verifyGeneric checks the whole execution by using the gogll generated lexer and parser. It does not distinguish between clean-start and recovery. If
// the execution respect any of the two it will return true.
func (g *GrammarChecker) verifyGeneric(execution string) (bool, []*Error) {
	var errors []*Error
	lexer := lexer.New([]rune(execution))
	_, errs := parser.Parse(lexer)
	if len(errs) == 0 {
		return true, nil
	}
	for _, err := range errs {
		exp := []string{}
		for _, ex := range err.Expected {
			exp = append(exp, ex)
		}
		expectedTokens := strings.Join(exp, ",")
		unexpectedToken := err.Token.TypeID()
		e := &Error{
			description: fmt.Sprintf("Parser was expecting one of [%v], got [%v] instead.", expectedTokens, unexpectedToken),
			height:      err.Line - 1,
		}
		errors = append(errors, e)
	}
	return false, errors
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
