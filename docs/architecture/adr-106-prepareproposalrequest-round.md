# ADR 106: PrepareProposalRequest Round

## Changelog

- 2023-05-23: First draft (@brendanchou)

## Status

In Review

## Context

The `PrepareProposal` request does not contain a `Round` field which may give
proposers additional desired information as to whether to add some additional
trasactions to the block.

For example, the proposer may want to have different logic for `Round=0`
than for other rounds.

Tracking issue: [\#882]

See also: https://docs.cometbft.com/v0.37/spec/abci/abci++_methods.html#prepareproposal

## Alternative Approaches

Add a new `PartialHeader` field to the `Request` instead that encapsulates information
such as the `Height`, `Time`, and `Round`. However, this would be a breaking change.

## Decision

TODO

## Detailed Design

To the `PrepareProposal` `Request`, add a `Round` field to indicate the round
of the block.


## Consequences

### Positive

- Additional information for the proposer to make decisions
- Non-breaking change

### Negative

- Code change required in CometBFT
- Like `Height` and `Time`, the `Round` field is just added to the top-level of
  the request making the request fields more noisy (i.e. in the interest of
  compatibility, they are not well-contained within a `PartialHeader` field).

### Neutral

## References
