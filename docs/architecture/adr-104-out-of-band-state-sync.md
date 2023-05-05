# ADR 104: Out-of-band state sync support

## Changelog

- {date}: {changelog}

## Status

> An architecture decision is considered "proposed" when a PR containing the ADR
> is submitted. When merged, an ADR must have a status associated with it, which
> must be one of: "Accepted", "Rejected", "Deprecated" or "Superseded".
>
> An accepted ADR's implementation status must be tracked via a tracking issue,
> milestone or project board (only one of these is necessary). For example:
>
>     Accepted
>
>     [Tracking issue](https://github.com/cometbft/cometbft/issues/123)
>     [Milestone](https://github.com/cometbft/cometbft/milestones/123)
>     [Project board](https://github.com/orgs/cometbft/projects/123)
>
> Rejected ADRs are captured as a record of recommendations that we specifically
> do not (and possibly never) want to implement. The ADR itself must, for
> posterity, include reasoning as to why it was rejected.
>
> If an ADR is deprecated, simply write "Deprecated" in this section. If an ADR
> is superseded by one or more other ADRs, provide local a reference to those
> ADRs, e.g.:
>
>     Superseded by [ADR 123](./adr-123.md)

Accepted | Rejected | Deprecated | Superseded by

## Context

> 1. What is state sync?

From CometBFT v0.34, synchronizing a fresh node with the rest of the network
can benefit from [state sync][state-sync], a protocol for discovering,
fetching, and installing application snapshots.
State sync is able to rapidly bootstrap a node by installing a relatively
recent state machine snapshot, instead of replaying all historical blocks.

> 2. What is the problem?

With the widespread adoption of State sync to bootstrap nodes, however,
what should be one of its strengths - the ability to discover and fetch
application snapshots from peers in the p2p network - has turned out to be one
of its weakness.
In fact, while downloading recent snapshots is very convenient for new nodes
(clients of the protocol), providing snapshots to multiple peers (as servers of
the protocol) is _bandwidth-consuming_.
As a result, the number of nodes in production CometBFT networks offering the
State sync service (i.e., servers offering snapshots) has been limited, which
has rendered the service _fragile_ (from the client's point of view).

> 3. High level idea behind the solution

This ADR stems from the observation that State sync is more than a protocol to
discover and fetch snapshots from peers in the network.
In fact, in addition to installing a snapshot, State sync also checks the
consistency of the installed application state with the state (`appHash`)
recorded on the blockchain at the same height, and bootstraps CometBFT's block
store and state store accordingly.
As a result, once State sync is completed the node can safely switch to Block
sync and/or Consensus to catch up with the latest state of the network.

The purpose of this ADR is to provide node operators with more flexibility in
defining how or where State sync should look for application snapshots.
The goal is to provide an alternative to the synchronization mechanism
currently adopted by State sync, discovering and fetching application snapshots
from peers in the network, in order to address the above mentioned limitations,
while preserving the core of State sync's operation.

## Alternative Approaches

> This section contains information around alternative options that are considered
> before making a decision. It should contain a explanation on why the alternative
> approach(es) were not chosen.

## Decision

> This section records the decision that was made.
> It is best to record as much info as possible from the discussion that happened.
> This aids in not having to go back to the Pull Request to get the needed information.

State sync should support the bootstrap of new nodes from application snapshots
obtained out-of-band by node operators.
In other words, the following operation should be, in general terms, possible:

1. When configuring a new node, operators should be able to define where (e.g.,
   a local file system location) the client side of State sync should look for
   application snapshots to install or restore;
1. Operators should be able to instruct the server side of State sync on a
   running node to export application snapshots, when they are available,
   to a given location (e.g., at the local file system);
1. It is up to node operators to transfer application snapshots produced by the
   running node (server side) to the new node (client side) to be bootstrapped
   using this new State sync feature.


## Detailed Design

> This section does not need to be filled in at the start of the ADR, but must
> be completed prior to the merging of the implementation.
>
> Here are some common questions that get answered as part of the detailed design:
>
> - What are the user requirements?
>
> - What systems will be affected?
>
> - What new data structures are needed, what data structures will be changed?
>
> - What new APIs will be needed, what APIs will be changed?
>
> - What are the efficiency considerations (time/space)?
>
> - What are the expected access patterns (load/throughput)?
>
> - Are there any logging, monitoring or observability needs?
>
> - Are there any security considerations?
>
> - Are there any privacy considerations?
>
> - How will the changes be tested?
>
> - If the change is large, how will the changes be broken up for ease of review?
>
> - Will these changes require a breaking (major) release?
>
> - Does this change require coordination with the SDK or other?

## Consequences

> This section describes the consequences, after applying the decision. All
> consequences should be summarized here, not just the "positive" ones.

### Positive

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- {reference link}

[state-sync]: https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_app_requirements.md#state-sync
