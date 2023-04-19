# ADR 82: E2E tests for CometBFT's behaviour in respect to ABCI++.

## Context

We want to test whether CommetBFT respects the ABCI++ grammar. To be able to do this we need to enhance the e2e tests infrastructure. Specifically, we 
plan to do three things:

- Log every CometBFT's ABCI++ request.
- Parse the logs and extract all ABCI++ requests.
- Check if the set of observed requests respect the ABCI++ grammar.

Issue: [353](https://github.com/cometbft/cometbft/issues/353).


## Decision

### 1) ABCI++ requests logging
We plan to do this at the Application side. Every time the App receives a request, it logs it.

<strong>Implementation</strong>

The key idea behind this part of the implementation was to log the request concisely and use the existing structures as much as possible. 

Every time an ABCI request is made, the application will create `abci.Request` (`abci` stands for `"github.com/cometbft/cometbft/abci/types"`) and log it. The example is below.  

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
that returns `abci.Request` from the string are provided in files `test/e2e/app/log.go` and `test/e2e/app/log_test.go`, respectively. 


### 2) Parsing the logs
We need a code that will take the logs from all nodes and parse the ABCI requests that were logged by the application. 

<strong>Implementation</strong>

This logic is implemented inside the `fetchABCIRequestsByNodeName()` function that resides in `test/e2e/tests/e2e_test.go` file. This function does three things:
- Takes the output of all nodes in the testnet from the moment we launched the testnet until the function is called. It uses the `docker-compose logs` command. 
- Parses the logs line by line and extracts the node name and the abci request, if one exists.
- Returns the map where the key is the node name and the value is the list of all abci requests logged on that node. 
 

### 3) ABCI++ grammar checker
This can also be a separate program that is receiving a set of ABCI++ calls as input and is returning whether they respect the grammar or not. Ideally, we should use some library here that will help us. Specifically, the library should receive the grammar and set of events and gives whether the set of events are possible within this grammar. 

#### Implementation

## Status

Partially implemented.

## Consequences

### Positive
- We will be able to check whether CommetBFT respects ABCI++ grammar. 
### Negative

### Neutral

