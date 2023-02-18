# ADR 82: E2E tests for CometBFT's behaviour in respect to ABCI++

## Context

We want to test whether CommetBFT respects the ABCI++ grammar. To be able to do this we need to enhance the e2e tests infrastructure. Specifically, we 
plan to do three things:

- Log every CometBFT's ABCI++ call.
- Parse the logs and extract all ABCI++ calls.
- Check if the set of observed calls respect the ABCI++ grammar.

Issue: [353](https://github.com/cometbft/cometbft/issues/353).


## Decision

### 1) ABCI++ calls logging
We plan to do this at the Application side. Every time the App receives a call it generates an object ABCI++_CALL and log it. We need a concise way of printing this information. 
### 2) Go program for parsing the logs
This can be a separated Go program that is receiving the logs and filter parameters as input, and it gives back a set of observed ABCI++ calls as output.

### 3) ABCI++ grammar checker
This can also be a separate program that is receiving a set of ABCI++ calls as input and is returning whether they respect the grammar or not. Ideally, we should use some library here that will help us. Specifically, the library should receive the grammar and set of events and gives whether the set of events are possible within this grammar. 

## Status

Not implemented.

## Consequences

### Positive

### Negative

### Neutral

