# ADR 104: Out-of-band state sync support

## Changelog

- 2023-05-05: Initial Draft (@cason)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

> 1. What is state sync?

From CometBFT v0.34, synchronizing a fresh node with the rest of the network
can benefit from [state sync][state-sync], a protocol for discovering,
fetching, and installing application snapshots.
State sync is able to rapidly bootstrap a node by installing a relatively
recent state machine snapshot, instead of replaying all historical blocks.

> 2. What is the problem?

With the widespread adoption of state sync to bootstrap nodes, however,
what should be one of its strengths - the ability to discover and fetch
application snapshots from peers in the p2p network - has turned out to be one
of its weakness.
In fact, while downloading recent snapshots is very convenient for new nodes
(clients of the protocol), providing snapshots to multiple peers (as servers of
the protocol) is _bandwidth-consuming_.
As a result, the number of nodes in production CometBFT networks offering the
state sync service (i.e., servers offering snapshots) has been limited, which
has rendered the service _fragile_ (from the client's point of view).

> 3. High level idea behind the solution

This ADR stems from the observation that state sync is more than a protocol to
discover and fetch snapshots from peers in the network.
In fact, in addition to installing a snapshot, state sync also checks the
consistency of the installed application state with the state (`appHash`)
recorded on the blockchain at the same height, and bootstraps CometBFT's block
store and state store accordingly.
As a result, once state sync is completed the node can safely switch to Block
sync and/or Consensus to catch up with the latest state of the network.

The purpose of this ADR is to provide node operators with more flexibility in
defining how or where state sync should look for application snapshots.
More precisely, it enables state sync to support the bootstrap of nodes from
application snapshots obtained _out-of-band_ by operators.
The goal is to provide an alternative to the mechanism currently adopted by
state sync, discovering and fetching application snapshots from peers in the
network, in order to address the above mentioned limitations, while preserving
most of state sync's operation.

## Alternative Approaches

> This section contains information around alternative options that are considered
> before making a decision. It should contain a explanation on why the alternative
> approach(es) were not chosen.

### Strengthen p2p-based state sync

This ADR proposes a workaround for some limitations of the existing state sync
protocol, briefly summarized above. A more comprehensive and sustainable
long-term solution would require addressing such limitations.

From a protocol design point of view, probably the major limitation of state
sync is the lack of mechanisms to incentivize and reward established nodes that
provide "good" (updated and consistent) snapshots to peers.
The existing protocol is essentially altruistic: it assumes that established
nodes will support the bootstrap of fresh nodes without receiving anything in
return, other than having new peers joining their network.

From an operational point of view, the design of the p2p layer should take into
consideration the fact that not every node in the network is a "good" state
sync server. Fresh nodes, which need support for bootstrapping, should then be
able to discover peers in the network that are able or willing to provide
application snapshots.
The implementation of this feature would require a more complex content-based
peer discovery mechanism.

However, while relevant, strengthening p2p-based state sync should be seen as
_orthogonal_ to the solution proposed in this ADR, which should have relevant use
cases even for an enhanced version of state sync.


### Manual application state transfer

The proposed [ADR 083][adr083]/[ADR 103][adr103] inspired the discussions that
led to this ADR. At first glance, their proposal might look identical to the
solution discussed here. The main distinction is that their proposal does not
involve installing application snapshots, but relies on node operators to
manually synchronize the application state.

More precisely, their proposal is to
"_allow applications to start with a bootstrapped state_
(i.e., the application has already a state before the node is started)
_alongside an empty CometBFT instance_
(i.e., the node's block store is empty and the state store is at genesis state)
_using a subsystem of the state sync protocol_".

Starting from an "empty CometBFT instance" is a requirement for running state
sync, which is maintained in the current proposal.
The mentioned "subsystem of the state sync protocol" is the light client,
responsible for verifying the application state and for bootstrapping
CometBFT's block store and state store accordingly.

The distinction, therefore, lies in the way which the state of the application
from a running node is transferred to the new node to be bootstrapped.
Instead of relying on application snapshots, produced by the application but
handled by state sync, [ADR 083][adr083]/[ADR 103][adr103] assumes that node
operators are able to manually synchronize the application state from a running
node (it might be necessary to stop it) to a not-yet-started fresh node.

The main limitation of the approach in [ADR 083][adr083]/[ADR 103][adr103] is that it relies on the ability of node
operators to properly synchronize the application state between two nodes.
While experienced node operators are likely able to perform this operation in a
proper way, we have to consider a broader set of users and emphasize that it is
an operation susceptible to errors.
Furthermore, it is an operation that is, by definition, application-specific:
applications are free to manage their state as they see fit, and this includes
how and where the state is persisted.
Node operators would therefore need to know the specifics of each application
in order to adopt this solution.

## Decision

> This section records the decision that was made.
> It is best to record as much info as possible from the discussion that happened.
> This aids in not having to go back to the Pull Request to get the needed information.

State sync should support the bootstrap of new nodes from application snapshots
obtained out-of-band by node operators.

In other words, the following operation should be, in general terms, possible:

1. When configuring a new node, operators should be able to define where (e.g.,
   a local file system location) the client side of state sync should look for
   application snapshots to install or restore;
1. Operators should be able to instruct the server side of state sync on a
   running node to export application snapshots, when they are available,
   to a given location (e.g., at the local file system);
1. It is up to node operators to transfer application snapshots produced by the
   running node (server side) to the new node (client side) to be bootstrapped
   using this new state sync feature.


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

- This is a non-breaking change: it provides an alternative and complementary
  implementation for state sync
- Node operators will not need to use workarounds to make state sync to download
  application snapshots from specific nodes in the network, in particular from
  nodes that are controlled by the same operators

### Negative

### Neutral

- Additional complexity, with additional parameters for state sync's
  configuration and the bootstrap of a node
- Additional complexity, with the possible addition of a CLI command to save
  application snapshots to the file system

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- [State sync][state-sync] description, as part of ABCI++ spec
- Original issue on Tendermint Core repository: [statesync: bootstrap node with state obtained out-of-band #4642][original-issue]
- Original solution on Tendermint Core repository: [ADR 083: Supporting out of band state sync #9651][adr083]
- Original proposal ported to CometBFT repository: [ADR 103: Local State Sync Support][adr103]

[state-sync]: https://github.com/cometbft/cometbft/blob/main/spec/abci/abci%2B%2B_app_requirements.md#state-sync
[original-issue]: https://github.com/tendermint/tendermint/issues/4642
[adr083]: https://github.com/tendermint/tendermint/pull/9651
[adr103]: https://github.com/cometbft/cometbft/pull/729
