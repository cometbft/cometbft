# ADR 82: E2E tests for CometBFT's behaviour in respect to ABCI++.

## Context

We want to test whether CommetBFT respects the ABCI++ grammar. To do this, we need to enhance the e2e tests infrastructure. Specifically, we plan to do three things:
- Log every CometBFT's ABCI++ request during the execution.
- Parse the logs post-mortem and extract all ABCI++ requests.
- Check if the set of observed requests respects the ABCI++ grammar.


Issue: [353](https://github.com/cometbft/cometbft/issues/353).


## Decision

### 1) ABCI++ requests logging
We plan to do this at the Application side. Every time the App receives a request, it logs it.

<strong>Implementation</strong>

The key idea behind this part of the implementation was to log the request concisely and use the existing structures as much as possible. 

Whenever an ABCI request is made, the application will create `abci.Request` (`abci` stands for `"github.com/cometbft/cometbft/abci/types"`) and log it.  The example is below.  

```go
func (app *Application) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {

	r := &abci.Request{Value: &abci.Request_InitChain{InitChain: &abci.RequestInitChain{}}}
	app.logRequest(r)

	...
}
```
Notice here that we create an empty `abci.RequestInitChain` object while we can also use the one passed to the `InitChain` function. The reason behind this is that, at the moment, we do not need specific fields of the request; we just need to be able to extract the information about the request type. For this, an empty object of a particular type is enough. 

`app.logRequest(r)` function is a new function implemented in the same file (`test/e2e/app/app.go`). Its implementation is the following: 

```go
func (app *Application) logRequest(req *abci.Request) {
	s, err := GetABCIRequestString(req)
	if err != nil {
		panic(err)
	}
	app.logger.Debug(s)
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
In addition, we surround the new string with `abci-call` constants so that we can find lines with ABCI++ request more easily.

Lastly, if in the future we want to log another ABCI++ request type, we just need to do the same thing: 
create a corresponding `abci.Request` and log it via 
`app.logRequest(r)`. 

### 2) Parsing the logs
We need a code that will take the logs from all nodes and collect the ABCI++ requests that were logged by the application. 

<strong>Implementation</strong>

This logic is implemented inside the `fetchABCIRequestsByNodeName()` function that resides in `test/e2e/tests/e2e_test.go` file. This function does three things:
- Takes the output of all nodes in the testnet from the moment we launched the testnet until the function is called. It uses the `docker-compose logs` command. 
- Parses the logs line by line and extracts the node name and the  `abci.Request`, if one exists. The node name is extracted manually and `abci.Request` is received by forwarding each line to the `app.GetABCIRequestFromString(req)` method.
- Returns the map where the key is the node name, and the value is the list of all `abci.Request` logged on that node. 
We can now use the list of `abci.Request` to refer to ABCI++ requests of any type, which is why we logged them in the previously described way. 

 

### 3) ABCI++ grammar checker
The idea here was to find a library that automatically verifies whether a specific execution respects the prescribed grammar. 

<strong>Implementation</strong>
We found the following library - https://github.com/goccmack/gogll. It generates a GLL or LR(1) parser and FSA-based lexer for any context-free grammar. What we needed to do is to write ABCI++ grammar (ref to the grammar)
using the synthax that the library understand. 
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

Proposer : PrepareProposal ProcessProposal ; 
NonProposer: ProcessProposal ;
Decide : BeginBlock DeliverTxs EndBlock | BeginBlock EndBlock ; 
DeliverTxs : DeliverTx | DeliverTx DeliverTxs ; 


InitChain : "<InitChain>" ;
BeginBlock : "<BeginBlock>" ; 
DeliverTx : "<DeliverTx>" ;
EndBlock : "<EndBlock>" ;
Commit : "<Commit>" ;
OfferSnapshot : "<OfferSnapshot>" ;
ApplyChunk : "<ApplyChunk>" ; 
PrepareProposal : "<PrepareProposal>" ; 
ProcessProposal : "<ProcessProposal>" ;
 
 ```

If you compare this grammar with the original one, you will notice that method
`Info` is removed. The reason is that, as explained in the section [CometBFT's expected behaviour](../../spec/abci/abci%2B%2B_tmint_expected_behavior.md#valid-method-call-sequences), one of the 
purposes of the `Info` method is part of the RPC handling from an external 
client, which can happen at any time, and as such, cannot be expressed with 
grammar.  
This is not the case with the other two purposes, but since the Application does 
not distinguish between different cases of why the `Info` is called, we removed 
it totally from the new grammar. The Application is still logging the `Info` 
call, but a specific test would need to be written to check whether it happens
in the right moment. 

The `gogll` library receives the file with the grammar as input, and it generates the corresponding parser and lexer. The code that 
this library generates is inside the following directories: 
- `test/e2e/pkg/grammar/lexer`,
- `test/e2e/pkg/grammar/parser`,
- `test/e2e/pkg/grammar/sppf`,
- `test/e2e/pkg/grammar/token`.

Apart from this auto-generated code, we implemented `GrammarChecker` abstraction
which knows how to use the generated parser and lexer to verify whether a
specific execution (set of ABCI++ calls logged by the Application while the
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
			t.Error(err)
		}
	})
}
```

Specifically, it first fetches all ABCI++ requests and creates a `GrammarChecker` object. Then for each
node in the testnet it checks if a specific set of requests respects the ABCI++ 
grammar by calling `checker.Verify(reqs)` method. If this method returns an error, the specific execution does not respect the grammar. 

The `Verify()` method is shown below. It takes a list of requests and does the following things:
- filter the last height. Basically, it removes all ABCI++ requests after the 
last `Commit`. This is needed because when we collect the requests, we collect 
all requests from the start until we call `fetchABCIRequestsByNodeName()`. As a result the last height may be incomplete, and 
the parser may return an error. The simple example here is that the last 
request is `BeginBlock`; however `EndBlock` still did not happen, and the parser
will return an error that `EndBlock` is missing, even though the `EndBlock` may happen but after the moment when the `fetchABCIRequestsByNodeName()` was invoked. 




## Status

Partially implemented.
To-do list:
- integrating the generation of parser/lexer into the codebase.
## Consequences

### Positive
- We should be able to check whether CommetBFT respects ABCI++ grammar. 
### Negative

### Neutral

