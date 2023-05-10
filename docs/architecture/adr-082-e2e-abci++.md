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
In addition, we surround the new string with `abci-call` constants so that I can find lines with ABCI++ request more easily.

Lastly, if in the future we want to log another ABCI++ request type, we just need to do the same thing: 
create a corresponding `abci.Request` and log it via 
`app.logRequest(r)`. 

### 2) Parsing the logs
We need a code that will take the logs from all nodes and parse the ABCI++ requests that were logged by the application. 

<strong>Implementation</strong>

This logic is implemented inside the `fetchABCIRequestsByNodeName()` function that resides in `test/e2e/tests/e2e_test.go` file. This function does three things:
- Takes the output of all nodes in the testnet from the moment we launched the testnet until the function is called. It uses the `docker-compose logs` command. 
- Parses the logs line by line and extracts the node name and the  `abci.Request`, if one exists. The node name is extracted manually and `abci.Request` is received by forwarding each line to the `app.GetABCIRequestFromString(req)` method.
- Returns the map where the key is the node name, and the value is the list of all `abci.Request` logged on that node. 
We can now use the list of `abci.Request` to refer to ABCI++ requests of any type, which is why we logged them in the previously described way. 

 

### 3) ABCI++ grammar checker


#### Implementation

## Status

Partially implemented.

## Consequences

### Positive
- We will be able to check whether CommetBFT respects ABCI++ grammar. 
### Negative

### Neutral

