# ADR 116: Modular crypto

## Changelog

 - June 12 2024: Created by @melekes

## Status

**Proposed**

## Context

Currently there's no way to add a new cryptographic curve (used for consensus
signing) to CometBFT without forking the codebase. Maintaining a fork is a
significant burden and makes it hard to keep up with the latest changes in the
main codebase.

On the other hand, **adding a new curve to the main codebase is a significant
burden as well** because it brings in new dependencies and requires to maintain
the code for the new curve. For example, the [recent BLS12-381 curve
addition](https://github.com/cometbft/cometbft/pull/2765).

The goal of this ADR is to make it possible to add new cryptographic curves to
the CometBFT codebase without forking it, but at the same time avoid the burden
of maintaining the code for all possible curves in the main codebase.

## Proposal



## Alternative Approaches

1. Do nothing. Keep adding new curves to the main codebase as needed =>
   burdensome, time consuming, and error-prone.

## Detailed Design

## Consequences

### Positive

### Neutral

### Negative

## References
