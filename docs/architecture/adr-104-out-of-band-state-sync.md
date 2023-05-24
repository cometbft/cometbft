# ADR 104: Out-of-band state sync support

## Changelog

- 2023-05-05: Initial Draft (@cason)
- 2023-24-05: Added description of SDK-based & a CometBFT based solution (@jmalicevic)

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
the protocol) is _bandwidth-consuming_, especially without a clear incentive for
node operators to provide this service. 
As a result, the number of nodes in production CometBFT networks offering the
state sync service (i.e., servers offering snapshots) has been limited, which
has rendered the service _fragile_ (from the client's point of view). It is very
hard to find a node with _good_ snapshots, leading nodes to often block during sync up.

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

The ADR presents two solutions:

> 1. The first solution is implemented by the application and was proposed and 
implemented by the SDK (PR #16067 and #16061). This ADR describes the solution 
in a general way, proividing guidlines to non-SDK based applications if they wish
to implement their own local state sync. 
> 2. The second part of the ADR proposes a more general solution, that uses ABCI
to achieve the same behaviour achieved by the SDK.


## Alternative Approaches

> This section contains information around alternative options that were considered
> before making a decision. It should contain an explanation on why the alternative
> approach(es) were not chosen.


### Strengthen p2p-based state sync

This ADR proposes a workaround for some limitations of the existing state sync
protocol, briefly summarized above. A more comprehensive and sustainable
long-term solution would require addressing such limitations, by improving the
current p2p based state sync solution.

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

As the SDK implemented a solution for all applications using it, and we do not have
users requesting local state sync that are running in production, we will not implement the generally applicable
solution. 

This ADR will outline both the solution implemented by the SDK, as well as a design proposal of a 
generally applicable solution, that can be later on picked up and implemented in CometBFT.

## Detailed Design

### Application-implemented local state sync

This section describes the solution to local state sync implemented by the SDK. An application can 
chose to implement local state sync differently, or implement only a subset of the functionalities
implemented by the SDK.

This solution exposes a command line interface enabling a node to manipulate the snapshots
including dumping existing snapshots to an exportable format, loading and deleting exported snapshots,
as well as bootstraping a node's state based on a local snapshot. 

The SDK exposes the following commands for snapshot manipulation:

```script

delete      Delete a local snapshot
dump        Dump the snapshot as portable archive format
export      Export app state to snapshot store
list        List local snapshots
load        Load a snapshot archive file into snapshot store
restore     Restore app state from local snapshot

```

and the following command to bootstrap the state of CometBFT upon a completed state sync:

``` script

comet bootstrap-state          Bootstrap the state of CometBFT's state and block store

```

These commands enable the implementation of both the client and server side of statesync. 
Namely, a statesync server can use `dump` to create a portable archieve format out existing snapshots,
or trigger snapshot creation using `export`. 

The client side, restores the application state from a local snapshot that was previously exported, using
`restore`. Upon successful completion of the previous sequence of commands, the state of CometBFT is bootstrapped
using `bootstrap-state` and CometBFT can be launched. 


There are three prerequisites for this solution to work when a node is syncing:
1. The application has access to the snapshot database
2. CometBFTs state and block stores are empty or reset
3. CometBFT is not running while the node is state syncing 

The server side of state sync (snapshot generation and dumping), can be performed while the node is running.
The application has to be careful not to interfere with normal node operations, and to use a snapshot store
and dumping mechanism that will mitigate the risk of requesting snapshots while they are being dumped to an archive format. 

In order to be able to dump or export the snapshots, the application must have access to the snapshot store. 

We describe the main interface expected from the snapshot database and used by the above mentioned CLI commands
for snapshot manipulation. 
The interface was derived from the SDK's implementation of local state sync.  

```golang

// Delete a snapshot for a certain height
func (s *Store) Delete(height uint64, format uint32) error

// Retrieves a snapshot for a certain height
func (s *Store) Get(height uint64, format uint32) (*types.Snapshot, error) 

// List recent snapshots, in reverse order (newest first)
func (s *Store) List() ([]*types.Snapshot, error)

// Loads a snapshot (both metadata and binary chunks). The chunks must be consumed and closed.
// Returns nil if the snapshot does not exist.
func (s *Store) Load(height uint64, format uint32) (*types.Snapshot, <-chan io.ReadCloser, error)

// LoadChunk loads a chunk from disk, or returns nil if it does not exist. The caller must call
// Close() on it when done.
func (s *Store) LoadChunk(height uint64, format, chunk uint32) (io.ReadCloser, error)

// Save saves a snapshot to disk, returning it.
func (s *Store) Save(height uint64, format uint32, chunks <-chan io.ReadCloser) (*types.Snapshot, error) 

// PathChunk generates a snapshot chunk path.
func (s *Store) PathChunk(height uint64, format, chunk uint32) string

```

In order to dump a snapshot, an application needs to retrieve all the chunks stored at a certain path. 

**CometBFT state bootstrap**

In addition to managing snapshots, it is neccessary to bootstrap (setup) the state and block store of CometBFT
in order for CometBFT to continue with block sync and consensus after state sync. 
At the moment of writing this ADR, there is no command line in CometBFT that supports this.
Thus application developers have to, within their bootstraping command:

- Create a state and block store
- Launch a light client to verify the snapshot headers
- Use the light client's `State` and `Commit` functions to 
retrieve the proper state, verify the `AppHash` and retrieve the `LastCommit`.
- Save the retrieved values into the state and block stores. 

This code is essentially what the existing implementation of the statesync reactor does, once state syncing is complete 
minus the functions that apply the snapshot chunks (as this has already been done offline). 


## CometBFT based local state sync

Given that snapshot manipulation is entirely application defined, and to avoid pulling this functionality into
CometBFT, we propose a solution using ABCI, that mimics the behaviour described in the previous section.

Local state sync can be performed online, while CometBFT is running, by reading the snapshots from a 
predefined location. For this to be possible, this ADR proposes the following additions to the configuration file:

```bash
[statesync]
# State syncing from a local snapshot
local_sync=false 
# Path to snapshot, will be ignored if local_sync=true
snapshot_load_path=""
# Periodically dump snapshots into archive format (optional)
auto_snapshot_dump=false
# If dumping nodes into archive format, set the path to where the file is dumped
# and the file format
# This can be changed when using a CLI command
snapshot_dump_path=""
snapshot_dump_format=""
```

Operators can instruct the application to periodically dump the existing snapshots into an exportable arhive format.
They can also configure the path and format of the file to use when exporting snapshots.

CometBFT exposes a CLI command to instruct the application to dump the existing snapshots into an exportable format:

```bash
dump      Dump existing snapshots to exportable file format.
```

We can expand the CLI interface to allow for additional operations, similar to the above provided CLI. However,
as snapshot generation and dumping would be built in into CometBFT, while the node is running, we do not need
to rely on a CLI for this functionality.

Applications can themselves be in charge of dumping the snapshots into a given file format. An alternative solution is 
that CometBFT, itself, after recieveing snapshots obtained via `ListSnapshots`, generates a file and dumps it to 
a predefined location. 

On startup, if `local_sync` is set to `true`, CometBFT will try to look for an existing snapshot at the path
given by the operator. The snapshots will be loaded and state restored as if they came from a peer in the current implementation. 

The `dump` command can be implemented in two ways:
- Rely on the existing ABCI functions `ListSnapshots` and `LoadChunks` to retrieve the snapshots and chunks from a peer. 
This approach requires no change to the current API and is easy to implement. It however involves more than one ABCI call and 
data transafer. 
- Extend `RequestListSnapshots` with a flag to indicate that we want an exportable snapshot format and extend `ResponseListSnapshot` to return a 
path and format of the exported snapshots.
- Add path parameters to the command, and include the path into `RequestListSnapshots` instructing the application to generate a snapshot export
at a given location. 

The same caveat of making sure the data exported is not corrupted by concurrent reads and writes applies here as well. If we decide to simply export
snapshots into a compressed file, we need to make sure that we do not return files that are not already written. 
An alternative is for CometBFT to provide a Snapshot export store with snapshot isolation. This store would essentially be pruned as soon as a new snapshot is written,
thus not taking too much more space.


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
