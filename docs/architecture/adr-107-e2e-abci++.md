# ADR 107: E2E tests for CometBFT's behaviour in respect to ABCI++.

## Context

We want to test whether CommetBFT respects the ABCI++ grammar. To do this, we need to enhance the e2e tests infrastructure. Specifically, we plan to do three things:
- Log every CometBFT's ABCI++ request during the execution.
- Parse the logs post-mortem and extract all ABCI++ requests.
- Check if the set of observed requests respects the ABCI++ grammar.


Issue: [353](https://github.com/cometbft/cometbft/issues/353).

Current version does not support vote extensions. However, this is the next step. 


## Decision

### 1) ABCI++ requests logging
The idea was to do this at the Application side. Every time the Application 
receives a request, it logs it.

**Implementation**

The rationale behind this part of the implementation was to log the request concisely and use the existing structures as much as possible. 

Whenever an ABCI request is made, the application will create `abci.Request` (`abci` stands for `"github.com/cometbft/cometbft/abci/types"`) and log it.  The example is below.  

```go
func (app *Application) InitChain(_ context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {

	r := &abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}}
	err := app.logAbciRequest(r)
	if err != nil {
		return nil, err
	}
    
	...
}
```
Notice here that we create an empty `abci.RequestInitChain` object while we can also use the one passed to the `InitChain` function. The reason behind this is that, at the moment, we do not need specific fields of the request; we just need to be able to extract the information about the request type. For this, an empty object of a particular type is enough. 

`app.logAbciRequest(r)` function is a new function implemented in the same file (`test/e2e/app/app.go`). Its implementation is the following: 

```go
func (app *Application) logAbciRequest(req *abci.Request) error {
	s, err := GetABCIRequestString(req)
	if err != nil {
		return err
	}
	app.logger.Debug(s)
	return nil
}
```

`GetABCIRequestString(req)` is a new method that receives a request and returns its string representation. The implementation and tests for this function and the opposite function `GetABCIRequestFromString(req)`
that returns `abci.Request` from the string are provided in files `test/e2e/app/log.go` and `test/e2e/app/log_test.go`, respectively. To create a string representation of a request, we first marshal the request via `proto.Marshal()` method and then convert received bytes in the string using `base64.StdEncoding.EncodeToString()` method. The code of this method is below: 

```go
func GetABCIRequestString(req *abci.Request) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}
	reqStr := base64.StdEncoding.EncodeToString(b)
	s := ABCI_REQ + reqStr + ABCI_REQ
	return s, nil
}
```
In addition, we surround the new string with `abci-req` constants so that we can find lines with ABCI++ request more easily.
If in the future we want to log another ABCI++ request type, we just need to do the same thing: 
create a corresponding `abci.Request` and log it via 
`app.logAbciRequest(r)`. 

### 2) Parsing the logs
We need a code that will take the logs from all nodes and collect the ABCI++ requests that were logged by the application. 

**Implementation**

This logic is implemented inside the `fetchABCIRequestsByNodeName()` function that resides in `test/e2e/tests/e2e_test.go` file. This function does three things:
- Takes the output of all nodes in the testnet from the moment we launched the testnet until the function is called. It uses the `docker-compose logs` command. 
- Parses the logs line by line and extracts the node name and the  `abci.Request`, if one exists. The node name is extracted manually and `abci.Request` is received by forwarding each line to the `app.GetABCIRequestFromString(req)` method.
- Returns the map where the key is the node name, and the value is the list of all `abci.Request` logged on that node. 
We can now use `[]*abci.Request` to store ABCI++ requests of any type, which is why we logged them in the previously described way. 

 

### 3) ABCI++ grammar checker
The idea here was to find a library that automatically verifies whether a specific execution respects the prescribed grammar. 

**Implementation**

We found the following library - https://github.com/goccmack/gogll. It generates a GLL or LR(1) parser and an FSA-based lexer for any context-free grammar. What we needed to do is to rewrite ABCI++ grammar ([CometBFT's expected behaviour](../../spec/abci/abci%2B%2B_comet_expected_behavior.md#valid-method-call-sequences))
using the syntax that the library understands. We should emphasise here that both grammars, the original and the new one, represent the expected behaviour
from the perspective of one node. This is why, later, when we verify if the specific execution respects the grammar, we need to check the logs of each node separately. 
The new grammar is below and can be found inside `test/e2e/pkg/grammar/abci_grammar.md` file.

```abnf

Start : CleanStart | Recovery ;

CleanStart : InitChain StateSync ConsensusExec | InitChain ConsensusExec ;
StateSync : StateSyncAttempts SuccessSync |  SuccessSync ; 
StateSyncAttempts : StateSyncAttempt | StateSyncAttempt StateSyncAttempts ;
StateSyncAttempt : OfferSnapshot ApplyChunks | OfferSnapshot ;
SuccessSync : OfferSnapshot ApplyChunks ; 
ApplyChunks : ApplyChunk | ApplyChunk ApplyChunks ;  

Recovery :  ConsensusExec ;

ConsensusExec : ConsensusHeights ;
ConsensusHeights : ConsensusHeight | ConsensusHeight ConsensusHeights ;
ConsensusHeight : ConsensusRounds Decide Commit | Decide Commit ;
ConsensusRounds : ConsensusRound | ConsensusRound ConsensusRounds ;
ConsensusRound : Proposer | NonProposer ; 

Proposer : PrepareProposal | PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;


InitChain : "init_chain" ;
Decide : "finalize_block" ; 
Commit : "commit" ;
OfferSnapshot : "offer_snapshot" ;
ApplyChunk : "apply_snapshot_chunk" ; 
PrepareProposal : "prepare_proposal" ; 
ProcessProposal : "process_proposal" ;
 
 ```

If you compare this grammar with the original one, you will notice that, in addition to vote extensions,  
`Info` is removed. The reason is that, as explained in the section [CometBFT's expected behaviour](../../spec/abci/abci%2B%2B_comet_expected_behavior.md#valid-method-call-sequences), one of the 
purposes of the `Info` method is being part of the RPC handling from an external 
client. Since this can happen at any time, it complicates the 
grammar.  
This is not true in other cases, but since the Application does 
not know why the `Info` is called, we removed 
it totally from the new grammar. The Application is still logging the `Info` 
call, but a specific test would need to be written to check whether it happens
in the right moment. 

The `gogll` library receives the file with the grammar as input, and it generates the corresponding parser and lexer. Specifically, we need to run 
`gogll pkg/grammar/abci_grammar.md` from `test/e2e/` directory.
The resulting code is inside the following directories: 
- `test/e2e/pkg/grammar/lexer`,
- `test/e2e/pkg/grammar/parser`,
- `test/e2e/pkg/grammar/sppf`,
- `test/e2e/pkg/grammar/token`.

Apart from this auto-generated code, we implemented `GrammarChecker` abstraction
which knows how to use the generated parser and lexer to verify whether a
specific execution (list of ABCI++ calls logged by the Application while the
testnet was running) respects the ABCI++ grammar. The implementation and tests 
for it are inside `test/e2e/pkg/grammar/checker.go` and 
`test/e2e/pkg/grammar/checker_test.go`, respectively. 

How the `GrammarChecker` works is demonstrated with the test `TestABCIGrammar`
implemented in `test/e2e/tests/abci_test.go` file. 

```go
func TestABCIGrammar(t *testing.T) {
	m := fetchABCIRequestsByNodeName(t)
	checker := grammar.NewGrammarChecker(grammar.DefaultConfig())
	testNode(t, func(t *testing.T, node e2e.Node) {
		reqs := m[node.Name]
		_, err := checker.Verify(reqs)
		if err != nil {
			t.Error(fmt.Errorf("ABCI grammar verification failed: %w", err))
		}
	})
}
```

Specifically, the test first fetches all ABCI++ requests and creates a `GrammarChecker` object. Then for each
node in the testnet, it checks if a specific set of requests, logged by this node, respects the ABCI++ 
grammar by calling `checker.Verify(reqs)` method. If this method returns an error, the specific execution does not respect the grammar. Again, we do
this for each node individually because the grammar describes the correct behaviour from the perspective of one node, and this is what the `checker.Verify(reqs)` inspects. 

The `Verify()` method is shown below. 
```go
func (g *GrammarChecker) Verify(reqs []*abci.Request) (bool, error) {
	var r []*abci.Request
	r, _ = g.filterLastHeight(reqs)
	s := g.GetExecutionString(r)
	return g.VerifyExecution(s)
}
```

It takes a list of requests and does the following things.
- Filter the last height. Basically, it removes all ABCI++ requests after the 
last `Commit`. The function `fetchABCIRequestsByNodeName()` can be called in the middle of the height. As a result, the last height may be incomplete, and 
the parser may return an error. The simple example here is that the last 
request fetched via `fetchABCIRequestsByNodeName()` is `Decide`; however, `Commit` happens after 
`fetchABCIRequestsByNodeName()` was invoked. Consequently, the parser
will return an error that `Commit` is missing, even though the `Commit` 
will happen after.  
- Generates an execution string by replacing `abci.Request` with the 
corresponding terminal from the grammar. This logic is implemented in
`GetExecutionString()` function. This function receives a list of `abci.Request` and generates a string where every request the grammar covers 
will be replaced with a corresponding terminal. For example, request `r` of type `abci.Request_PrepareProposal` is replaced with the string `prepare_proposal`, the first part of `r`'s string representation. If the grammar does not cover the request, it will be ignored. 
- Checks if the resulting string with terminals respects the grammar. This 
logic is implemented inside the `VerifyExecution` function. 

```go
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
```
This function is the only function that uses auto-generated parser and 
lexer. It returns true if the execution is valid. Otherwise, it returns an 
error composed of parser errors and some additional information 
we added. In addition, if the `ShowFullExecution` is set to `true`, it prints the whole execution. An example of an error produced by `VerifyExecution`
is the following:

```
ABCI grammar verification failed: Parser failed, number of errors is 2
            ---Error 0---
            Height: 0
            ABCI requests: offer_snapshot apply_snapshot_chunk finalize_block commit
            Unexpected request: offer_snapshot
            Expected one of: [init_chain,process_proposal,finalize_block,prepare_proposal]
            -------------
            Full execution:
            0: offer_snapshot apply_snapshot_chunk finalize_block commit
            1: finalize_block commit
            2: finalize_block commit
            3: finalize_block commit
            4: finalize_block commit
            5: process_proposal finalize_block commit
			...
```
The parse error shown above represents an error that happened at height 0. The `ABCI requests` part represents 
requests observed in this height, while `Unexpected request` and `Expected one of` represent the request which 
should not happen, and the requests that should have happened instead, respectively. 
Lastly, after the errors the full execution, one height per line, is printed. This can be turned off with the config flag. 
Notice here that the parser can return many errors because the parser returns an error at every point at which the parser fails to parse
a grammar production. Usually, the error of interest is the one that has 
parsed the largest number of tokens. This is why, by default, we are printing only the last error; however, this is also part of the configuration and can be changed. 

### Suporting additional ABCI++ requests

Here we present all the steps we need to do if we want to support other 
ABCI++ requests in the future: 

- The application needs to log the new request in the same way as we do now.
- We should include the new request to the grammar and generate a new parser and lexer.  
- We should add new requests to the list of supported requests. Namely, we should modify the function `isSupportedByGrammar()` in `test/e2e/pkg/grammar/checker.go` to return `true` for the new type of requests.


## Status

Implemented.

To-do list:
- integrating the generation of parser/lexer into the codebase.
## Consequences

### Positive
- We should be able to check whether CommetBFT respects ABCI++ grammar. 
### Negative

### Neutral

