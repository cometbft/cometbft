# ADR 113: Mempool Lanes

## Changelog

- 2024-04-12: Initial notes (@hvanz)
- 2024-04-12: Comments on the notes (@sergio-mena)
- 2024-04-17: Discussions (@sergio-mena @hvanz)
- 2024-04-18: Preliminary structure (@hvanz)
- 2024-05-01: Add Context and Properties (@hvanz)
- 2024-05-21: Add more Properties + priority mempool (@sergio-mena)
- 2024-06-13: Technical design (@hvanz)

## Status

In writing (we're not sure yet whether this will be an ADR or a spec)

## Context

In the current implementation, the only property that the mempool tries to enforce when processing
and disseminating transactions is maintaining the order in which transactions arrive to the nodes,
that is, a FIFO ordering. However, ensuring a strict transmission order over the network proves
challenging due to inherent characteristics of the underlying communication protocols that causes
message delays and potential reordering. Consequently, while many Tendermint Core and CometBFT
applications have always assumed this ordering always holds, the FIFO-ness of transactions is not
guaranteed and is offered only as a best effort.

Beyond the apparent FIFO sequencing, transactions in the mempool are treated equally, meaning that
they are not discriminated as to which are disseminated first or which transactions the mempool
offers to the proposer when creating the next block. In practice, however, not all transactions have
the same importance for the application logic, especially when it comes to latency requirements.
Depending on the application, we may think of countless categories of transactions based on their
importance and requirements, spanning from IBC messages to transactions for exchanges, for smart
contract execution, for smart contract deployment, grouped by SDK modules, and so on. Even
transactions prioritized by economic incentives could be given a preferential treatment. Or big
transactions, regardless of their nature, could be categorized as low priority, to mitigate
potential attacks on the mempool.

The goal of this document is thus to propose a mechanism enabling the mempool to prioritize
transactions by *classes*, for processing and dissemination, directly impacting block creation and transaction latency. In
IP networking terminology, this is known as Quality of Service (QoS). By providing certain QoS
guarantees, developers will be able to more easily estimate when transactions will be disseminated, processed and
included in a block.

In practical terms, we envision an implementation of the transaction class abstraction as *mempool
lanes*. The application will be allowed to split the mempool transaction space into a hierarchy of
lanes, with each lane operating as an independent mempool. At the same time, all of them need to be
coordinated to ensure the delivery of the desired levels of QoS.

Note that improving the dissemination protocol to reduce bandwidth and/or latency is a separate
concern and falls outside the scope of this proposal. Likewise, graceful degradation under high load
is an orthogonal problem to transaction classification, although the latter may help improve the former.

## Properties

Before jumping into the design of the proposal, we define more formally the properties supported by
the current implementation of the mempool. Then we state what properties the new mempool should
offer to guarantee the desired QoS.

The following definition is common to all properties.

:memo: _Definition_: Given any two different transactions `tx1` and `tx2`, we say that `tx1` is
*processed and disseminated before* `tx2` in a given node, when:
- `tx1` is reaped from the mempool to form a block proposal before `tx2`,
- `tx1` is rechecked before `tx2`, and
- `tx1` is sent to a given peer before `tx2`.

Note that in the current implementation there is one dissemination routine per peer, so it could
happen that `tx2` is sent to a peer before `tx1` is sent to a different peer.

### Current mempool

As stated above, the current mempool offers a best-effort FIFO ordering of transactions. We state
this property as follows.

:parking: _Property_ **FIFO ordering of transactions**: We say that the mempool makes a best effort
in maintaining the FIFO ordering of transactions when transactions are validated, processed, and
disseminated in the same order in which the mempool has received them.

More formally, given any two different transactions `tx1` and `tx2`, if a node's mempool receives `tx1`
before receiving `tx2`, then:
- `tx1` will be validated against the application (via `CheckTx`) before `tx2`, and
- `tx1` will be processed and disseminated before `tx2` (as defined above).

Note that a node's mempool can receive a transaction either from a `broadcast_tx_*` RPC endpoint or
from a peer.

This property guarantees the FIFO ordering at any given node, but it cannot be generalised to all
the nodes in the network because the property does not hold at the network level. Hence, FIFO
ordering on the whole system is best effort.

### Mempool with QoS

The main goal of QoS is to guarantee that certain transactions have lower latency than others.
Before stating this property, we need to make some definitions.

:memo: _Definition_: a *transaction class* is a disjoint set of transactions having some common
characteristics as defined by the application.

A transaction may only have one class. If it is not assigned any specific class, it will be assigned
a *default class*, which is a special class always present in any set of classes. This is analogous
to the _native VLAN_ for untagged traffic in an 802.1Q network. Because no transaction can belong to
two or more classes, transaction classes form disjoint sets, that is, the intersection between
classes is empty. Also, all transactions in the mempool are the union of the transactions in all
classes.

:memo: _Definition_: Each class has a *priority* and two classes cannot have the same priority.
Therefore all classes can be ordered by priority.

Given these definitions, we want the proposed QoS mechanism to offer the following property:

#### Basic properties

:parking: _Property_ **Priorities between classes**: Transactions belonging to a certain class will
be processed and disseminated before transactions belonging to another class with lower priority.

Formally, given two transaction classes `c1` and `c2`, with `c1` having more priority than `c2`, if
the application assigns the classes `c1` and `c2` respectively to transactions `tx1` and `tx2`, then
`tx1` will be processed and disseminated before `tx2`.

More importantly, as a direct consequence of this property, `tx1` will be disseminated faster and it
will be included in a block before `tx2`, thus `tx1` will have a lower latency than `tx2`.
Currently, it is not possible to guarantee this kind of property.

:memo: _Definition_: The *latency of a transaction* is the difference between the time at which a
user or client submits the transaction for the first time to any node in the network, and the
timestamp of the block in which the transaction finally was included.

We want also to keep the FIFO ordering within each class (for the time being):

:parking: _Property_ **FIFO ordering per class**: For transactions within the same class, the
mempool will maintain FIFO order at a node when they are validated, processed,
and disseminated.

Given any two different transactions `tx1` and `tx2` belonging to the same class, if the mempool
receives `tx1` before receiving `tx2`, then:
- `tx1` will be validated against the application (via `CheckTx`) before `tx2`, and
- `tx1` will be processed and disseminated before `tx2`.

As a consequence, given that classes of transactions have a sequential ordering, and that classes do
not have elements in common, we can state the following property:

:parking: _Property_ **Partial ordering of all transactions**: The set of all the transactions in
the mempool, regardless of their classes, will have a *partial order*.

This means that some pairs of
transactions are comparable and, thus, have and order, while others not.

#### Network-wide consistency

The properties presented so far may be interpreted as per-node properties.
However, we need to define some network-wide properties in order for a mempool QoS implementation
to be useful and predictable for the whole appchain network.
These properties are expressed in terms of consistency of the information, configuration and behaviour
across nodes in the network.

> TODO: Define "for the first time" here: received from RPC, peer, or `Update`
> (try to define it in as modular a way as possible)

:parking: _Property_ **Consistent transaction classes**: For any transaction `tx`,
and any two correct nodes $p$ and $q$ that receive `tx` *for the first time*,
$p$ and $q$ MUST have the same set of transaction classes and their relative priority and configuration.

The property is only required to hold for on-the-fly transactions:
if a node receives a (late) transaction that has already been decided, this property does not enforce anything.
The same goes for duplicate transactions.
Notice that, if this property does not hold, it is not possible to guarantee any property across the network,
such as transaction latency as defined above.

:parking: _Property_ **Consistent transaction classification**: For any transaction `tx`
and any two correct nodes $p$ and $q$ that receive `tx` *for the first time*,
$p$'s application MUST classify `tx` into the same transaction class as $q$'s application.

This property only makes sense when the previous property (_consistent transaction classes_) defined above holds.
Even if we ensure consistent transaction classes, if this property does not hold, a given transaction
may not receive the same classification across the network and it will thus be impossible to reason
about any network-wide guarantees we want to provide that transaction with.

Additionally, it is important to note that these two properties also constrain the way transaction
classes and transaction classification logic can evolve in an existing implementation.
If either transaction classes or classification logic are not modified in a coordinated manner in a working system,
there will be at least a period where the these two properties may not hold for all transactions.

> TODO: Need to find somewhere in the text to say: "ReCheckTx" doesn't classify, its mempool information is disregarded"

## Alternative Approaches

### CometBFT Priority Mempool

CometBFT used to have a `v1` mempool, specified in Tendermint Core [ADR067][adr067] and deprecated as of `v0.37.x`,
which supported per-transaction priority assignment.
The key point of the priority mempool's design was that `CheckTxResponse` was extended with a few fields,
one of which being an `int64` that the application could use to provide a priority to the transaction being checked.

This design can be seen as partially addressing the specification of a Mempool with QoS
presented in the previous section. Every possible value of the `int64` priority field returned by the application
can be understood as a _different_ traffic class.
Let us examine whether the properties specified above are fulfilled by the priority mempool design
as described in [ADR067][adr067]:

1. Partial ordering of all transactions is maintained because the design still keeps a FIFO queue for gossiping transactions.
  Also, transactions are reaped according to non-decreasing priority first, and then in FIFO order
  for transactions with equal priority (see this `ReapMaxBytesMaxGas`'s [docstring][reapmaxbytesmaxgas]).
1. Since the priority mempool uses FIFO for transactions of equal priority, it also fulfills "FIFO ordering per class".
  The problem here is that, since every value of the priority `int64` field is considered a different transaction class,
  there are virtually unlimited traffic classes.
  So it is too easy for an application to end up using hundreds, if not thousands of transactions classes at a given time.
  In this situation, "FIFO ordering per class", while fulfilled, becomes a corner case and thus does not add much value.
1. The consistent transaction classes property is trivially fulfilled, as the set of transaction classes never changes:
  it is the set of all possible values of an `int64`.
1. Finally, the priority mempool design does not make any provisions on how the application is to evolve its prioritization
  (i.e., transaction classification) logic.
  Therefore, the design does not guarantee the fulfillment of the consistent transaction classification property.

The main hindrance for the wide adoption of the priority mempool was
the dramatic reduction of the _observable_ FIFO guarantees for transactions (as explained in point 2 above)
with respect to the `v0` mempool.

Besides, the lack of provisions for evolving the prioritization logic (point 4 above) could have also got
in the way of adoption.


### Solana

#### Introduction to Gulf Stream and Comparison with CometBFT's Mempool

A core part of Solana's design is [Gulf Stream][gulf-stream],
which is marketed as a "mempool-less" way of processing in-flight transactions.
Similarly of a CometBFT- based chain, the sequence of leaders (nodes that produce blocks) is known in advance.
However, unlike CometBFT, Solana keeps the same leader for a whole epoch, whole typical length is approx. 2 days
(what if the leader fails in the middle of an epoch?).
According to the Gulf Stream design, rather than maintaining a mempool at all nodes to ensure transactions
will reach _any_ leader/validator, transactions are directly sent to the current leader and the next,
according to the sequence of leaders calculated locally (known as _leader schedule_).
As a result, Gulf Stream does not use gossip-based primitives to disseminate transactions,
but UDP packets sent directly to the current (and next) leader's IP address.
One of the main points of adopting gossip protocols by Tendermint Core and CometBFT (coming from Bitcoin and Ethereum)
is censorship resistance. It is not clear how Gulf Stream deals with an adversary controlling a part of the network
that stands on the way of those UDP packets containing submitted transactions.

#### Transaction Priority Design

In Solana, transaction priority is controlled by fees: they introduce the concept of [_priority fees_][solana-prio-fees].
The priority fee is an optional configuration parameter when submitting a transaction,
which allows the submitter to increase the likelihood of their transaction making it to a block.
The priority fee is provided in terms of _price per Unit of Computation_ (UC), priced in [micro-lamports per CU][prio-fee-price].
A UC is the equivalent of Cosmos's _gas_, and so, the priority fee is analogous (in concept)
to the Cosmos SDK's `--gas-prices` [flag][sdk-gas-prices].
The main difference if that the SDK (currently) uses `--gas-prices`
to set up a per-node threshold of acceptability in gas prices,
whereas Solana uses the (default or user-configured) priority fee as the transaction's _actual_ priority.

This is very similar to the way CometBFT's priority mempool in `v0.37.x` was supposed to be used by applications,
but in a monolithic manner: there is no "priority" abstraction in Solana as there is nothing similar to ABCI.
In short, the fees _are_ the priority.
Thus, if we were to check the properties specified [above](#mempool-with-qos),
with the caveat that Solana does not have a built-in mempool,
we would reach the same conclusions as with the CometBFT's `v0.37.x` [priority mempool](#cometbft-priority-mempool).
Namely, a _degradation_ in observable FIFO guarantees (affecting applications that depend on it for performance),
and a lack a provisions of evolving priority classification in a consistent manner.
The latter may appear less important as transactions are directly sent to the current leader,
but it is not clear how retried transactions in periods of high load can be receive a consistent priority treatment.

### Ethereum pre-confirmations

TODO

### Skip's Block-SDK Lanes

As of version `v0.47.x`, the Cosmos SDK offers application developers the possibility to use an [Application-Side Mempool][sdk-app-mempool].
It is a mempool structure maintained by the SDK application and populated with valid transactions received via `CheckTx`.
An application maintaining such a mempool is free to define the way transactions are ordered, reaped for a block, aggregated, removed, etc.
Typically, upon `PrepareProposal`, the SDK application disregards the transactions proposed by CometBFT,
and rather proposes transactions reaped from its own mempool, and according to its mempool's rules.

The Skip team have released an extension of the Application-Side mempool, called [Block-SDK][skip-block-sdk],
that introduces the concept of _lanes_, turning the mempool "into a *highway* consisting of individual *lanes*".
The concept of lanes, introduced in Skip's Block-SDK, is pretty aligned with Mempool QoS as specified above.
Indeed, we use the same term, _lanes_, in the [Detailed Design](#detailed-design) section below,
which describes a minimum viable product (MVP) implementation of the concept of transaction classes.

The main difference between Skip's Block-SDK's lanes and the design we present below is that
the Block-SDK implements mempool lanes at the application level, whereas this ADR proposes a specification and a design at CometBFT level,
thus including provisions for **transaction gossiping** as an integral part of it.
As a result, the Block-SDK's lanes can be used to implement the Mempool QoS specification in everything that relates to block production,
but not at the network gossip level.

Importantly, both designs, Skip's Block SDK and the one described [below](#detailed-design), are complementary.
An application using Skip's Block-SDK lanes already contains transaction classification logic, and so,
it can easily be extended to provide `CheckTx` with the information needed by an implementation of CometBFT mempool QoS
(such as the design we propose below) to also achieve a more predictable transaction latency,
depending on the lane/class a transaction belongs to.

## Decision

TODO

## Detailed Design

This sections describes the architectural changes needed to implement an MVP of lanes in the
mempool. The following is a summary of the key design decisions:
- [[1](#lanes-definition-and-initialization)] The list of lanes and their corresponding priorities
  will be hardcoded in the application logic.
- [[2](#internal-data-structures)] There will be one CList per lane.
- [[3](#configuration)] All lanes will share the same mempool configuration.
- [[4](#adding-transactions-to-the-mempool)] When validating a transaction via CheckTx, the
  application will optionally return a lane for the transaction.
- [[5](#reaping-transactions-for-creating-blocks)] Transactions will be reaped from higher-priority
  lanes first, preserving the FIFO ordering.
- [[6](#transaction-dissemination)] We will continue to use the current P2P channel for
   disseminating transactions, and we will implement in the mempool the logic for selecting the
   order in which to send transactions.

### Lanes definition and initialization

The list of lanes and their corresponding priorities will be hardcoded in the application logic. A
priority is a value of type `uint32`, with 0 being a reserved value (see below). The application
also needs to define which of the lanes it defines is the default lane. 

To obtain the lane information from the application, we need to extend the ABCI `Info` response to
include the following fields. These fields need to be filled by the application only in case it
wants to implement lanes.
```protobuf
message InfoResponse {
  ...
  repeated uint32 lanes = 6;
  uint32 default_lane = 7;
}
```
Internally, the application may use `string`s to name lanes, and then map those names to priorities.
The mempool does not care about the names, only about the priorities. That is why the lane
information returned by the application only contains priorities.

Currently, querying the app for `Info` happens during the handshake between CometBFT and the app,
during the node initialization, and only when state sync is not enabled. The `Handshaker` first
sends an `Info` request to fetch the app information, and then replays any stored block needed to
sync CometBFT with the app. The lane information is needed regardless of whether state sync is
enabled, so one option is to query the app information outside of the `Handshaker`.

The highest priority a lane may have is 1. The value 0 is reserved for two cases: when there are no
priorities and for `CheckTx` responses of invalid transactions.

On receiving the information from the app, CometBFT will validate that:
- `lanes` has no duplicates (values in `lanes` don't need to be sorted),
- `default_value` is in `lanes` (the default lane is not necessarily the lane with the lowest
  priority), and
- the list `lanes` is empty if and only if `default_value` is 0.

Different nodes also need to agree on the lanes they use. When a node connects to a peer, they both
perform a handshake to agree on some basic information (see `DefaultNodeInfo`), including the
version of the application they are executing. Since the application includes the lane definitions,
both nodes will also agree on the lanes they implement. Therefore we don't need to modify the P2P
handshake.

### Internal data structures

In the mempool, a lane is defined by its priority:
```golang
type Lane uint32
```

Currently, the `CListMempool` data structure has two fields to store and access transactions:
```golang
txs    *clist.CList // Concurrent list of mempool entries.
txsMap sync.Map     // Map of type TxKey -> *clist.CElement, for quick access to elements in txs.
```

With the introduction of lanes, the main change will be to divide the `CList` data structure into
$N$ `Clist`s, one per lane. `CListMempool` will have the following fields:
```golang
lanes   map[Lane]*clist.CList
txsMap  sync.Map // Map of type TxKey -> *clist.CElement, for quick access to elements in lanes.
txLanes sync.Map // Map of type TxKey -> Lane, for quick access to the lane corresponding to a tx.

// Fixed variables set during initialization.
defaultLane Lane
sortedLanes []Lane // Lanes sorted by priority
```
The auxiliary fields `txsMap` and `txLanes` are, respectively, for direct access to the mempool
entries, and for direct access to the lane of a given transaction.

If the application does not implement lanes (that is, it responds with empty values in
`InfoResponse`), then `defaultLane` will be set to 1, and `lanes` will have only one entry for the
default lane.

`CListMempool` also contains the cache, which is only needed before transactions have a lane
assigned. Since the cache is independent of the lanes, we do not need to modify it.

### Configuration

For an MVP, all lanes can share the same mempool configuration. In this scenario, all lanes will be
capped by the total mempool capacities as currently defined in the configuration. Namely, these are:
- `Size`, the total number of transactions allowed in the mempool, 
- `MaxTxsBytes`, the maximum total number of bytes of the mempool, and
- `MaxTxBytes`, the maximum size in bytes of a single transaction accepted into the mempool.

With lanes, the total size of the mempool will be the sum of the sizes of all lanes.

Additionally, the `Recheck` and `Broadcast` flags will apply to all lanes or to none. Remember that,
with `PrepareProposal`, it becomes _always_ mandatory to recheck remaining transactions in the
mempool, so there is no point in disabling `Recheck` per lane.

### Adding transactions to the mempool

When validating a transaction received for the first time with `CheckTx`, the application will
optionally return its lane in the response.
```protobuf
message CheckTxResponse {
  ...
  uint32 lane = 12;
}
```
The callback that handles the first-time CheckTx response will append the new mempool entry to the
corresponding `CList`, namely `lanes[lane]`, and update the other auxiliary variables accordingly.
If `lane` is 0, it means that the application did not set any lane in the response message, so the
transaction will be assigned to the default lane.

### Removing transactions from the mempool

A transaction may be removed in two scenarios: when updating a list of committed transactions, and
during rechecking if the transaction is reassessed as invalid. In either case, the first step is to
identify the lane the transaction belongs to by accessing the `txLanes` map. Then, we remove the
entry from the CList corresponding to its lane and update the auxiliary variables accordingly.

Since the broadcast goroutines are constantly reading the list of transactions to disseminate them,
it's important to prioritize the removal of transactions from high-priority lanes.

When updating the mempool, there is potential for a slight optimization by removing transactions
from different lanes in parallel. To achieve this, we would first need to preprocess the list of
transactions to determine the lane of each transaction. However, this optimization has minimal
impact if the committed block contains few transactions. Therefore, we decided to exclude it from
the MVP.

### Reaping transactions for creating blocks

Currently, the function `ReapMaxBytesMaxGas(maxBytes, maxGas)` collects transactions in FIFO order
from the CList until either reaching `maxBytes` or `maxGas` (both of these values are consensus
parameters).

With multiple CLists, we need to collect transactions from higher-priority lanes first, also in FIFO
order, continuing with successive lanes in the `sortedLanes` array, that is, in decreasing priority
order, and breaking the iteration when reaching `maxBytes` or `maxGas`.

The mempool is locked during `ReapMaxBytesMaxGas`, so no transaction will be added or removed from
the mempool during reaping.

### Transaction dissemination

The current dissemination algorithm works as follows. The broadcast routine in the mempool reactor
has a variable, `next`, pointing to the entry in the CList of transactions to be sent next to a
peer. Once a transaction is sent, `next` is updated to `next.Next()`, continuing the traversal of
the CList to send subsequent transactions.

Initially, `next` is nil, causing the dissemination algorithm to wait on the `mempool.TxsWaitChan()`
channel for a signal that the CList is not empty. Upon receinving this signal, the algorithm sets
`next` to `mempool.TxsFront()`, the the first entry in the list. If the end of the list is reached,
`next.Next()` returns nil, and the routine will block again on `mempool.TxsWaitChan()` until a new
entry is available. Additionally, `next` may become nil when the rechecking process removes the
entries positioned at or after the entry currently being broadcasted, potentially leaving the CList
fragmented.

For broadcasting transactions from multiple lanes, we see two possible options:
1. Reserve $N$ p2p channels for use by the mempool. P2P channels have priorities that we can reuse
   as lane priorities. There are a maximum of 256 P2P channels, thus limiting the number of lanes.
2. Continue using the current P2P channel for disseminating transactions and implement logic within
   the mempool to select the order of transactions to put in the channel. This option theoretically
   allows for an unlimited number of lanes, constrained only by the nodes’ capacity to store the
   lane data structures.

We choose the second option for its flexibility, allowing us to start with a simple scheduling
algorithm that can be refined over time. The desired algorithm must satisfy the properties
"Priorities between classes" and "FIFO ordering per class". This requires supporting selection by
_weight_, ensuring each lane gets a fraction of the P2P channel capacity proportional to its
priority. Moreover, it should be _fair_, to prevent starvation of low-priority lanes. Given the
extensive research on scheduling algorithms in operating systems and networking, we propose
implementing a variant of the [Weighted Round Robin (WRR)](wrr) algorithm, which meets these
requirements and is straightforward to implement.

### Checking received transactions

A malicious node may decide to send lower-priority transactions before higher-priority ones. The
receiving node can easily check the priority of a transaction when it calls `CheckTx`. Still, when
using one P2P channel, it is not possible to detect when a peer sends transactions out of order,
unless for example, when all received transactions are of low priority.

  * Note: if we punish peers (e.g., close connection to them) that send messages on wrong lanes, we will need a property that requires all node to produce consistent lane info in `CheckTx`

## Alternative designs

### One CList for all lanes

We briefly considered sharing one CList for all lanes, changing the internal logic of CList to
accommodate the lanes requirements, but this design clearly makes the code more complex, in
particular the transaction dissemination logic.

### One P2P channel per lane

P2P channels already have priorities implemented.
P2P channel could be different for different lanes.
A disadvantage is that nodes need to agree on the channels during the P2P handshake.

* A channel ID is a byte. The current channel distribution among all reactors goes up to channel ID `0x61`.
* Proposal: reserve a channel range for the mempool, for instance, channel ID `0x80` and all
  channels above (so, all channels whose most-significant byte is 1).
  * max of 128 lanes, which seems big enough for most users.
* Currently, the mempool's p2p channel ID is `0x30`, which would be used as the default lane.

### Duality lane/priority

The duality lane/priority could introduce a powerful indirection. The app could just define the lane
of a transaction in `CheckTx`, but the priority of the lane itself could be configured (and
fine-tuned) elsewhere. For example, by the app itself or by node operators. The proposed design does
not support this pattern.

### Custom configuration per lane

This could be a simple, future improvement to the MVP.

  * lane/mempool capacity (# of txs, total # of bytes, bytes per tx) **can be different** easily,
  with little extra complexity on config.
  * bytes per tx: ok to share one value for MVP... not so sure
  * \# of txs: probably not all use cases properly addressed if we share a common limit
  * total # of bytes: if the other two are per-lane, this can just be a "generous" value
    (and we'd fine tune the other two)

### Where to define lanes and priorities

There are two alternative options for _where_ to configure lanes and priorities:
1. `config.toml` / `app.toml`
  * It does not make sense for different nodes to have different lane configurations, so definitely
    we don't want 1.
2. `ConsensusParams`
  * If we can change lane info via `ConsensusParams`, CometBFT's mempool needs logic for changing
    lanes dynamically (complex, not really appealing for MVP).
  * The process of updating lane info would be very complex and cumbersome:
      * To update lanes, you'd need to pass _two_ governance proposals
        1. Upgrade the app, because the lane classification logic (so, the app's code) needs to know the lane config beforehand.
        2. Then upgrade the lanes via `ConsensusParams`.
      * Also, not clear in which order: community should be careful not to break performance between the passing of both proposals.
      * The `gov` module may allow the two things to be shipped in the same gov proposal,
        but, if we need to do it that way, what's the point in having lanes in `ConsensusParams`?
  1. Hardcoded in the application. Info passed to CometBFT during handshake.
    * Simple lane update cycle: one software update gov proposal.
    * CometBFT doesn't need to deal with dynamic lane changes: it just needs to set up lanes when starting up (whether afresh, or recovery).
    * Currently, one of the conditions for the handshake to succeed is that there must exist an intersection of p2p channels.

 * How to deal with nodes that are late? So, lane info at mempool level (and thus p2p channels to establish) does not match.
    * Two cases:
      1. A node that is up to date and falls behind (e.g., network instability, etc.)
        * Not a problem. For lanes to change we **require** a coordinated upgrade.
          * Reason: if we allow upgrade changing lane info **not** to be a coordinated upgrade,
            then we won't have consistent lane info across nodes (each node would be running a particular version of the software)
      2. A node that is blocksyncing from far away in the past (software upgrades in the middle)
        * Normal channels (including `0x30`): same as before
        * Lane channels (>=`0x80`), not declared. The node can't send/receive mempool-related info. Not a problem, since it's blocksyncing
          * If we're not at the latest version (e.g., Cosmovisor-drive blocksync)
            * channel info likely to be wrong
            * but Cosmovisor will kill the node before switching to consensus
    * We will need to extend the channel-related handshake between nodes to be able to add channels after initial handshake.
    * TODO: this needs a bit more thought/discussion
    * IDEA: use versions:
      * bump up the p2p version. Any late legacy node will have to upgrade to latest P2P version to acquire nodes
      * two ways of upgrading
        * Ethereum-like
          * easy: if `version` in `LaneInfo` don't match --> break down the connection (same treatment as `p2p` version)
          * we cannot afford this, as we need to support Cosmovisor-like apps (e.g., SDK apps)
        * Cosmovisor-like
          * harder: below
      * add a field `version` to `LaneInfo`
        * in the new p2p version, p2p handshake exchanges that version
        * if `version` in `LaneInfo` match
          * life is good!
        * if `version` in `LaneInfo` don't match
          * we still proceed with the handshake
          * but we don't announce any mempool channel info
          * so the two nodes can't exchange mempool-related messages,
            * they can exchange info in other channels (e.g., BlockSync or StateSync)
          * panic upon `SwitchToConsensus`
        * all this is to support the Cosmovisor way of things

## Consequences

TODO

### Positive

- Application developers will be able to better predict when transactions will be disseminated and
  included in a block.
- Lanes will preserve the "FIFO ordering of transactions" property within the same class (as best effort).

### Negative

- Total FIFO ordering of transaction dissemination and block creation will not be guaranteed.

### Neutral

- Lanes are optional. Current applications do not need to make any change to their code. Future
  applications will not be forced to use the lanes feature.
- The dissemination algorithm is fair, so low-priority transactions will not get stuck in the
  mempool forever.

## References

- [ADR067][adr067], Priority mempool
- [Docstring][reapmaxbytesmaxgas] of `ReapMaxBytesMaxGas`
- Solana's [Gulf Stream][gulf-stream]
- Solana's [Priority Fees][solana-prio-fees]
- Solana's [priority fee pricing][prio-fee-price]
- Cosmos SDK's [gas prices][sdk-gas-prices]
- Cosmos SDK's [application-side mempool][sdk-app-mempool]
- Skip's [Block SDK][skip-block-sdk]

[adr067]: ./tendermint-core/adr-067-mempool-refactor.md
[reapmaxbytesmaxgas]: https://github.com/cometbft/cometbft/blob/v0.37.6/mempool/v1/mempool.go#L315-L324
[gulf-stream]: https://medium.com/solana-labs/gulf-stream-solanas-mempool-less-transaction-forwarding-protocol-d342e72186ad
[solana-prio-fees]: https://solana.com/developers/guides/advanced/how-to-use-priority-fees
[prio-fee-price]: https://solana.com/developers/guides/advanced/how-to-use-priority-fees
[sdk-gas-prices]: https://docs.cosmos.network/v0.50/learn/beginner/tx-lifecycle#gas-and-fees
[sdk-app-mempool]: https://docs.cosmos.network/v0.47/build/building-apps/app-mempool
[skip-block-sdk]: https://github.com/skip-mev/block-sdk/blob/v2.1.3/README.md
[wrr]: https://en.wikipedia.org/wiki/Weighted_round_robin