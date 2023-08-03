# ADR 108: E2E tests for CometBFT's behaviour in respect to ABCI++.

## Context

We want to be able to test whether CommetBFT respects the ABCI++ grammar. To do this, we need to enhance the e2e tests infrastructure. Specifically, we plan to do three things:
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

The `app.logABCIRequest(r)` function is a new function implemented in the same file (`test/e2e/app/app.go`). If the `ABCIRequestsLoggingEnabled` flag is set to `true`, set automatically when ABCI tests are enabled, it logs received requests. The full implementation is the following: 

```go
func (app *Application) logABCIRequest(req *abci.Request) error {
	if !app.cfg.ABCIRequestsLoggingEnabled {
		return nil
	}
	s, err := GetABCIRequestString(req)
	if err != nil {
		return err
	}
	app.logger.Info(s)
	return nil
}
```

`GetABCIRequestString(req)` is a new method that receives a request and returns its string representation. The implementation and tests for this function and the opposite function `GetABCIRequestFromString(req)`
that returns `abci.Request` from the string are provided in files `test/e2e/app/log.go` and `test/e2e/app/log_test.go`, respectively. To create a string representation of a request, we first marshal the request via `proto.Marshal()` method and then convert received bytes in the string using `base64.StdEncoding.EncodeToString()` method. In addition, we surround the new string with `abci-req` constants so that we can find lines with ABCI++ request more easily. The code of the method is below: 

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

*Note:* At the moment, we are not compressing the marshalled request before converting it to `base64` `string` because we are logging the empty requests that take at most 24B. However, if we decide to log the actual requests in the future, we might want to compress them. Based on a few tests, we observed that the size of a request can go up to 7KB.  

If in the future we want to log another ABCI++ request type, we just need to do the same thing: 
create a corresponding `abci.Request` and log it via 
`app.logABCIRequest(r)`. 

### 2) Parsing the logs
We need a code that will take the logs from all nodes and collect the ABCI++ requests that were logged by the application. 

**Implementation**

This logic is implemented inside the `fetchABCIRequests(t *testing.T, nodeName string)` function that resides in `test/e2e/tests/e2e_test.go` file. This function does three things:
- Takes the output of a specific node in the testnet from the moment we launched the testnet until the function is called. The node name is passed as a function parameter. It uses the `docker-compose logs` and `grep nodeName` commands. 
- Parses the logs line by line and extracts the  `abci.Request`, if one exists. The request is received by forwarding each line to the `app.GetABCIRequestFromString(req)` method.
- Returns the array of slices where each slice contains the set of `abci.Request`s logged on that node. Every time the crash happens, a new array element (new slice `[]*abci.Request`) will be created. We know a crash has happened because we log "Application started" every time the application starts. Specifically, we added this log inside `NewApplication()` function in `test/e2e/app/app.go` file. In the end, the function will return just one slice if the node did not experience any crashes and $n+1$ slices if there were crashes, $n$ being the number of crashes. The benefit of logging the requests in the previously described way is that now we can use `[]*abci.Request` to store ABCI++ requests of any type.

 

### 3) ABCI++ grammar checker
The idea here was to find a library that automatically verifies whether a specific execution respects the prescribed grammar. 

**Implementation**

We found the following library - https://github.com/goccmack/gogll. It generates a GLL or LR(1) parser and an FSA-based lexer for any context-free grammar. What we needed to do is to rewrite ABCI++ grammar ([CometBFT's expected behaviour](../../spec/abci/abci%2B%2B_comet_expected_behavior.md#valid-method-call-sequences))
using the syntax that the library understands. 
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
*Important note:* Both grammars, the original and the new one, represent the expected behaviour
from the perspective of one node from the fresh beginning (`CleanStart`) or after the crash (`Recovery`). This is why, later, when we verify if the specific execution respects the grammar, we need to check each node's logs separately and distinguish the fresh start from recovery. 

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

The `gogll` library receives the file with the grammar as input, and it generates the corresponding parser and lexer. The actual command is integrated into `test/e2e/Makefile` and is executed when `make grammar` is invoked. 
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
	checker := grammar.NewGrammarChecker(grammar.DefaultConfig())
	testNode(t, func(t *testing.T, node e2e.Node) {
		if !node.Testnet.ABCITestsEnabled {
			return
		}
		reqs, err := fetchABCIRequests(t, node.Name)
		if err != nil {
			t.Error(fmt.Errorf("Collecting of ABCI requests failed: %w", err))
		}
		for i, r := range reqs {
			_, err := checker.Verify(r, i == 0)
			if err != nil {
				t.Error(fmt.Errorf("ABCI grammar verification failed: %w", err))
			}
		}
	})
}
```

Specifically, the test first creates a `GrammarChecker` object. Then for each node in the testnet, it collects all requests logged by this node. Remember here that `fetchABCIRequests()` returns an array of slices(`[]*abci.Request`) where the slice with index 0 corresponds to the node's `CleanStart` execution, and each additional slice corresponds to the `Recovery` execution after a specific crash. Each node must have one `CleanStart` execution and the same number of `Recovery` executions as the number of crashes that happened on this node. If collecting was successful, the test checks whether each execution respects the ABCI++ 
grammar by calling `checker.Verify(r, i == 0)` method. The second parameter (`i == 0`) indicates whether the set of requests `r` represents a `CleanStart` or a `Recovery` execution. If `Verify` returns an error, the specific execution does not respect the grammar, and the test will fail. 

The tests are executed only if `ABCITestsEnabled` is set to `true`. This is done through the manifest file. Namely, if we want to test whether CometBFT respects ABCI++ grammar, we would need to enable these tests by adding `abci_tests_enabled = true` in the manifest file of a particular testnet (e.g. `networks/ci.toml`). This will automatically activate logging on the application side. 

The `Verify()` method is shown below. 
```go
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
```

The method `Verify()` first calls method `filterRequests()` that is going to remove all the requests from the set that are not supported by the current version of the grammar. In addition, it will filter the last height by removing all ABCI++ requests after the 
last `Commit`. The function `fetchABCIRequests()` can be called in the middle of the height. As a result, the last height may be incomplete, and 
classified as invalid even if that is not the reality. The simple example here is that the last 
request fetched via `fetchABCIRequests()` is `Decide`; however, `Commit` happens after 
`fetchABCIRequests()` was invoked. Consequently, the execution
will be considered as faulty because `Commit` is missing, even though the `Commit` 
will happen after. 
After filtering the requests `Verify()` checks if the remaining set of requests respect the grammar. It does that by calling two methods: `verifySpecific()` and `verifyGeneric()`. 
Former should always be called first and is responsible of doing some specific checks that the parser cannot do. For example, at the moment it is checking if 

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

**Changing the grammar**

Any modification to the grammar (`test/e2e/pkg/grammar/abci_grammar.md`) requires generating a new parser and lexer. This is done by 
going to the `test/e2e/` directory and running:

```bash 
make grammar
``` 
Notice here that you need to have `gogll` installed 
on your machine to run the make successfully. If this is not the case, you can install it with the following command: 

```bash 
go get github.com/goccmack/gogll/v3
```  

### Suporting additional ABCI++ requests

Here we present all the steps we need to do if we want to support other 
ABCI++ requests in the future: 

- The application needs to log the new request in the same way as we do now.
- We should include the new request to the grammar and generate a new parser and lexer.  
- We should add new requests to the list of supported requests. Namely, we should modify the function `isSupportedByGrammar()` in `test/e2e/pkg/grammar/checker.go` to return `true` for the new type of requests.


## Status

Implemented.

To-do list:
- adding the CI workflow to check if make grammar is executed. 
## Consequences

### Positive
- We should be able to check whether CommetBFT respects ABCI++ grammar. 
### Negative

### Neutral

