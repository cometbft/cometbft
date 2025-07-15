# ADR 118: Mempool QoS

## Changelog

- 2024-04-12: Initial notes (@hvanz)
- 2024-04-12: Comments on the notes (@sergio-mena)
- 2024-04-17: Discussions (@sergio-mena @hvanz)
- 2024-04-18: Preliminary structure (@hvanz)
- 2024-05-01: Add Context and Properties (@hvanz)
- 2024-05-21: Add more Properties + priority mempool (@sergio-mena)
- 2024-06-13: Technical design (@hvanz)
- 2024-07-02: Updates based on reviewer's comments (@hvanz, @sergio-mena)
- 2024-07-09: Updates based on reviewer's comments (@hvanz)
- 2024-09-13: Added pre-confirmations section (@sergio-mena)
- 2024-09-27: Allow lanes to have same priority + lane capacities (@hvanz)

## Status

Accepted. Tracking issue: [#2803][tracking-issue].

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
transactions by *classes*, for processing and dissemination, directly impacting block creation and
transaction latency. In IP networking terminology, this is known as Quality of Service (QoS). By
providing certain QoS guarantees, developers will be able to more easily estimate when transactions
will be disseminated and reaped from the mempool to be included in a block.

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
offer to guarantee the desired QoS. The following definitions are common to all properties.

When attempting to add an incoming transaction to the mempool, the node first checks that it is not
already in the cache before checking its validity with the application. 

:memo: _Definition_: We say that a node receives a transaction `tx` _for the first time_ when the
node receives `tx` and `tx` is not in the cache.

By this definition, it is possible that a node receives a transaction "for the first time", then
gets the transaction evicted from the cache, and at a later time receives it "for the first time"
again. The cache implements a Least-Recently Used (LRU) policy for removing entries when the
cache is full.

:memo: _Definition_: Given any two different transactions `tx1` and `tx2`, in a given node, we say that:
1. `tx1` is *validated before* `tx2`, when `tx1` and `tx2` are received for the first time, and `tx1`
is validated against the application (via `CheckTx`) before `tx2`,
1. `tx1` is *rechecked before* `tx2`, when `tx1` and `tx2` are in the mempool and `tx1` is
re-validated (rechecked via `CheckTx`) before `tx2`,
1. `tx1` is *reaped before* `tx2`, when `tx1` is reaped from the mempool to be included in a block
  proposal before `tx2`,
1. `tx1` is *disseminated before* `tx2`, when `tx1` is sent to a given peer before `tx2`.

In 2, both transactions are rechecked at the same height, because both are in the mempool.

In 4, note that in the current implementation there is one dissemination routine per peer, so it
could happen that `tx2` is sent to a peer before `tx1` is sent to a different peer.
Hence the importance of expression "to a given peer" in that definition.

### Current mempool

As stated above, the current mempool offers a best-effort FIFO ordering of transactions. We state
this property as follows.

:parking: _Property_ **FIFO ordering of transactions**: We say that the mempool makes a best effort
in maintaining the FIFO ordering of transactions when transactions are validated, rechecked, reaped,
and disseminated in the same order in which the mempool has received them.

More formally, given any two different transactions `tx1` and `tx2`, if a node's mempool receives
`tx1` before receiving `tx2`, then `tx1` will be validated, rechecked, reaped, and disseminated
before `tx2` (as defined above).

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
two or more classes, transaction classes form disjoint sets, that is, the intersection between any two
classes is empty. Also, all transactions in the mempool are the union of the transactions in all
classes.

:memo: _Definition_: Each class has a *priority* and two classes cannot have the same priority.
Therefore all classes can be ordered by priority.

When a transaction is received for the first time and validated via `CheckTx`, the application MAY
return the class that it assigns to the transaction. If it actually returns a class, the mempool
MUST use it to prioritize the transaction. When transactions are rechecked, applications MAY return
a class, but the mempool will discard it.

Given these definitions, we want the proposed QoS mechanism to offer the following property:

#### Basic properties

:parking: _Property_ **Priorities between classes**: Transactions belonging to a certain class will
be reaped and disseminated before transactions belonging to another class with lower priority.

Formally, given two transaction classes `c1` and `c2`, with `c1` having more priority than `c2`, if
the application assigns the classes `c1` and `c2` respectively to transactions `tx1` and `tx2`, then
`tx1` will be reaped and disseminated before `tx2`.

More importantly, as a direct consequence of this property, `tx1` will be disseminated faster and it
will be included in a block before `tx2`. Thus, `tx1` will have a lower latency than `tx2`.
Currently, it is not possible to guarantee this kind of property.

:memo: _Definition_: The *latency of a transaction* is the difference between the time at which a
user or client submits the transaction for the first time to any node in the network, and the
timestamp of the block in which the transaction finally was included.

We want also to keep the FIFO ordering within each class (for the time being):

:parking: _Property_ **FIFO ordering per class**: For transactions within the same class, the
mempool will maintain a FIFO order within the class when transactions are validated, rechecked,
reaped, and disseminated.

Given any two different transactions `tx1` and `tx2` belonging to the same class, if the mempool
receives `tx1` before receiving `tx2`, then:
- `tx1` will be validated and recheck against the application (via `CheckTx`) before `tx2`, and
- `tx1` will be reaped and disseminated before `tx2`.

As a consequence, given that classes of transactions have a sequential ordering, and that classes do
not have elements in common, we can state the following property:

:parking: _Property_ **Partial ordering of all transactions**: The set of all the transactions in
the mempool, regardless of their classes, will have a *partial order*.

This means that some pairs of transactions are comparable and, thus, have an order, while others
not.

#### Network-wide consistency

The properties presented so far may be interpreted as per-node properties.
However, we need to define some network-wide properties in order for a mempool QoS implementation
to be useful and predictable for the whole appchain network.
These properties are expressed in terms of consistency of the information, configuration and behaviour
across nodes in the network.

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

This property only makes sense when the previous property ("Consistent transaction classes") defined above holds.
Even if we ensure consistent transaction classes, if this property does not hold, a given transaction
may not receive the same classification across the network and it will thus be impossible to reason
about any network-wide guarantees we want to provide that transaction with.

Additionally, it is important to note that these two properties also constrain the way transaction
classes and transaction classification logic can evolve in an existing implementation.
If either transaction classes or classification logic are not modified in a coordinated manner in a working system,
there will be at least a period where these two properties may not hold for all transactions.

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
1. Since the priority mempool uses FIFO for transactions of equal priority, it also fulfills the "FIFO ordering per class" property.
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

### Ethereum Pre-confirmations

#### Brief Explanation

Ethereum pre-confirmations are a mechanism designed to reduce transaction latency. Justin Drake's [proposal][based-preconfs]
for _based pre-confirmations_ has gained attention in recent months in the Ethereum research community,
though similar ideas date back to Bitcoin's [Oconfs][Oconfs].

Pre-confirmations occur in the context of _fast games_, techniques applied between consecutive Layer-1 blocks
to improve certain performance guarantees and help manage _MEV_ (Maximal Extractable Value).

The process is straightforward. A user submits a transaction and requests a _preconfer_ (a validator) to guarantee specific handling
of that transaction, typically for a fee, called _tip_.
In exchange, the preconfer signs a _promise_ &mdash; most often guaranteeing transaction inclusion in the next block.
The preconfer can only claim the tip if the promise is fulfilled, and validators opting in to become preconfers
accept new slashing conditions related to _liveness_ (failure to propose a block) and _safety_ (failure to meet the promise).

This design enables various implementations of pre-confirmations, and it's still early to determine which form will dominate in Ethereum.

#### Comparison to Mempool QoS

Unlike Mempool QoS &mdash; the design described [below](#detailed-design) &mdash; which prioritizes transactions based
on network resource availability,
pre-confirmations focus on individual user guarantees about transaction treatment and certainty of inclusion.
While the connection to MEV is not fully understood yet, pre-confirmations may provide some mitigation against MEV-related risks.

Pre-confirmations can also coexist with Mempool QoS in CometBFT-based blockchains.
For instance, particular Mempool QoS configurations, such as a starving, FIFO, high-priority lane,
could be part of an implementation of pre-confirmations in a CometBFT-based chain.

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

Implement an MVP following the design in the next section.

## Detailed Design

This section describes the architectural changes needed to implement an MVP of lanes in the
mempool. The following is a summary of the key design decisions:
- [[Lanes definition](#lanes-definition)] The list of lanes and their corresponding priorities
  will be hardcoded in the application logic.
- [[Initialization](#initialization)] How the application configures lanes on CometBFT.
- [[Internal data structures](#internal-data-structures)] There will be one concurrent list (CList) data structure per
  lane.
- [[Configuration](#configuration)] All lanes will share the same mempool configuration.
- [[Adding transactions to the mempool](#adding-transactions-to-the-mempool)] When validating a transaction via CheckTx, the
  application will optionally return a lane for the transaction.
- [[Transaction dissemination](#transaction-dissemination)] We will continue to use the current P2P
   channel for disseminating transactions, and we will implement in the mempool the logic for
   selecting the order in which to send transactions.
- [[Reaping transactions for creating blocks](#reaping-transactions-for-creating-blocks)]
  Transactions will be reaped from higher-priority lanes first, preserving intra-lane FIFO ordering.
- [[Prioritization logic](#prioritization-logic)] For disseminating and reaping transactions, the
   scheduling algorithm should be prevent starvation of low-priority lanes.

### Lanes definition

The list of lanes and their associated priorities will be hardcoded in the application logic. A lane
is identified by a **name** of type `string` and assigned a **priority** of type `uint32`. The
application also needs to define which of the lanes is the **default lane**, which is not
necessarily the lane with the lowest priority.

To obtain the lane information from the application, we need to extend the ABCI `Info` response to
include the following fields. These fields need to be filled by the application only in case it
wants to implement lanes.
```protobuf
message InfoResponse {
  ...
  map<string, uint32> lane_priorities = 6;
  uint32 default_lane = 7;
}
```
The field `lane_priorities` is a map from lane identifiers to priorities. Different lanes may have
the same priority. On the mempool side, lane identifiers will mainly be used for user interfacing
(logging, metric labels).

The lowest priority a lane may have is 1. Higher values correspond to higher priorities. The value 0
is reserved for when the application does not have a lane to assign, so it leaves the `lane_id`
field empty in the `CheckTx` response (see [below](#adding-transactions-to-the-mempool)). This
happens either when the application does not classify transactions, or when the transaction is
invalid.

On receiving the information from the app, CometBFT will validate that:
- `default_lane` is a key in `lane_priorities`, and
- `lane_priorities` is empty if and only if `default_lane` is empty.

### Initialization

Currently, querying the app for `Info` happens during the handshake between CometBFT and the app,
during the node initialization, and only when state sync is not enabled. The `Handshaker` first
sends an `Info` request to fetch the app information, and then replays any stored block needed to
sync CometBFT with the app. The lane information is needed regardless of whether state sync is
enabled, so one option is to query the app information outside of the `Handshaker`.

In this proposed approach, updating the lane definitions will require a single governance proposal
for updating the software. CometBFT will not need to deal with dynamic lane changes: it will just
need to set up the lanes when starting up (whether afresh or in recovery mode).

Different nodes also need to agree on the lanes they use. When a node connects to a peer, they both
perform a handshake to agree on some basic information (see `DefaultNodeInfo`). Since the
application includes the lane definitions, it suffices to ensure that both nodes agree on the
version of the application. Although the application version is included in `DefaultNodeInfo`, there
is currently no check for compatibility between the versions. To address this, we would need to
modify the P2P handshake process to validate that the application versions are compatible.

Finally, this approach is compatible with applications that need to swap binaries when
catching up or upgrading, such as SDK applications using [Cosmovisor][cosmovisor].
When a node is catching up (i.e., state or block syncing), its peers will detect
that the node is late and will not send it any transactions until it is caught up.
So, the particular lane configuration of the node is irrelevant while catching up.
When going through a Cosmovisor-driven upgrade, all nodes will swap binaries at the same
height (which is specified by the corresponding Software Upgrade gov proposal).
If the new version of the software contains modified lane configuration
(and therefore new transaction classification logic), those changes will kick in
in a coordinated manner thanks to the regular Cosmovisor workflow.

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
default lane. In this case, the new mempool's behaviour will be equivalent to that of the current mempool.

`CListMempool` also contains the cache, which is only needed before transactions have a lane
assigned. Since the cache is independent of the lanes, we do not need to modify it.

### Configuration

For an MVP, we do not need to have a customized configuration for each lane. The current mempool
configuration will continue to apply to the mempool, which now is a union of lanes. The total size
of the mempool will be the sum of the sizes of all lanes. Therefore, the mempool capacities as
currently defined in the configuration will put an upper limit on the union of all lanes. These
configurations are:
- `Size`, the total number of transactions allowed in the mempool, 
- `MaxTxsBytes`, the maximum total number of bytes of the mempool, and
- `MaxTxBytes`, the maximum size in bytes of a single transaction accepted into the mempool.

However, we still need to enforce limits on each lane's capacity. Without such limits, a
low-priority lane could end up occupying all the mempool space. Since we want to avoid introducing
new configuration options unless absolutely necessary, we propose two simple approaches for
partitioning the mempool space.

1. Proportionally to lane priorities: This approach could lead to under-utilization of the mempool if
   there are significant discrepancies between priority values, as it would allocate space unevenly.
2. Evenly across all lanes: Assuming high-priority transactions are smaller in size than
   low-priority transactions, this approach would still allow for more high-priority transactions to
   fit in the mempool compared to lower-priority ones.

Note that each lane's capacity will be limited both by the number of transactions and their total
size in bytes.

For the MVP, we've chosen the second approach. If users find that the lane capacity is insufficient,
they still have the option of increasing the total mempool size, which will proportionally increase
the capacity of all lanes. In future iterations, we may introduce more granular control over lane
capacities if needed.

Additionally, the `Recheck` and `Broadcast` flags will apply to all lanes or to none. Remember that,
if `PrepareProposal`'s app logic can ever add a new transaction, it becomes _always_ mandatory to
recheck remaining transactions in the mempool, so there is no point in disabling `Recheck` per lane.

### Adding transactions to the mempool

When validating a transaction received for the first time with `CheckTx`, the application will
optionally return its lane identifier in the response.
```protobuf
message CheckTxResponse {
  ...
  string lane_id = 12;
}
```
The callback that handles the first-time CheckTx response will append the new mempool entry to the
corresponding `CList`, namely `lanes[lane_id]`, and update the other auxiliary variables accordingly.
If `lane_id` is an empty string, it means that the application did not set any lane in the response
message, so the transaction will be assigned to the default lane.

### Removing transactions from the mempool

A transaction may be removed in two scenarios: when updating the mempool with a list of committed
transactions, or during rechecking if the transaction is reassessed as invalid. In either case, the
first step is to identify the lane the transaction belongs to by accessing the `txLanes` map. Then,
we remove the entry from the CList corresponding to its lane and update the auxiliary variables
accordingly.

As an optimization, we could prioritize the removal of transactions from high-priority lanes first.
The broadcast goroutines are constantly reading the list of transactions to disseminate them, though
there is no guarantee that they will not send transactions that are about to be removed.

When updating the mempool, there is potential for a slight optimization by removing transactions
from different lanes in parallel. To achieve this, we would first need to preprocess the list of
transactions to determine the lane of each transaction. However, this optimization has minimal
impact if the committed block contains few transactions. Therefore, we decided to exclude it from
the MVP.

### Transaction dissemination

For broadcasting transactions from multiple lanes, we have considered two possible approaches:
1. Reserve $N$ p2p channels for use by the mempool. P2P channels have priorities that we can reuse
   as lane priorities. There are a maximum of 256 P2P channels, thus limiting the number of lanes.
2. Continue using the current P2P channel for disseminating transactions and implement logic within
   the mempool to select the order of transactions to put in the channel. This option theoretically
   allows for an unlimited number of lanes, constrained only by the nodes’ capacity to store the
   lane data structures.

We choose the second approach for its flexibility, allowing us to start with a simple scheduling
algorithm that can be refined over time (see below). Another reason is that on the first option we
would need to initialize channels dynamically (currently there is a fixed list of channels passed as
node info) and assign lanes to channels.

Before modifying the dissemination logic, we need to refactor the current implementation and the
`Mempool` interface to clearly separate the broadcast goroutine in the mempool reactor from
`CListMempool` that includes the mempool data structures. `CListMempool` provides two methods used
by the broadcast code, `TxsWaitChan() <-chan struct{}` and `TxsFront() *clist.CElement`, which are
just wrappers around the methods `WaitChan` and `Front` of the `CList` implementation. In
particular, `TxsFront` is leaking implementation details outside the `Mempool` interface.

### Reaping transactions for block creation

In the current single-lane mempool, the function `ReapMaxBytesMaxGas(maxBytes, maxGas)` collects
transactions in FIFO order from the CList until either reaching `maxBytes` or `maxGas` (both of
these values are consensus parameters).

With multiple CLists, we need to collect transactions from higher-priority lanes first, also in FIFO
order, continuing with successive lanes in the `sortedLanes` array, that is, in decreasing priority
order, and breaking the iteration when reaching `maxBytes` or `maxGas`. Note that the mempool is
locked during `ReapMaxBytesMaxGas`, so no transaction will be added or removed from the mempool
during reaping.

This simple algorithm, though good enough for an MVP, does not guarantee that low-priority lanes
will not starve. That is why we prefer to implement one that is starvation-free, as explained in the
next section. It could be the same algorithm or similar to the one used for transaction
dissemination.

### Prioritization logic

For transaction dissemination and for reaping transactions for creating blocks we want a scheduling
algorithm that satisfies the properties "Priorities between classes" and "FIFO ordering per class".
This means that it must support selection by _weight_, ensuring each lane gets a fraction of the P2P
channel capacity proportional to its priority. Additionally, we want the algorithm to be _fair_ to
prevent starvation of low-priority lanes. 

A first option that meets these criteria is the current prioritization algorithm on the P2P reactor,
which we could easily reimplement in the mempool. It works as follows:
- On each P2P channel, the variable `recentlySent` keeps track of how many bytes were recently sent
  over the channel. Every time data is sent, increase `recentlySent` with the number of bytes
  written to the channel. Every 2 seconds, decrease `recentlySent` by 20% on all channels (these
  values are fixed).
- When sending the next message, [pick the channel][selectChannelToGossipOn] whose ratio
  `recentlySent/Priority` is the least.

From the extensive research in operating systems and networking, we can pick for the MVP an existing
scheduling algorithm that meets these requirements and is straightforward to implement, such as a
variant of the [Weighted Round Robin][wrr] (WRR) algorithm. We choose this option at it gives us
more flexibility for improving the logic in the future, for example, by adding a mechanism for
congestion control or by allowing some lanes to have customized, non-FIFO scheduling algorithms.

### Validating lanes of received transactions

Transactions are transmitted without lane information because peers cannot be trusted to send the
correct data. A node may take advantage of the network by sending lower-priority transactions before
higher-priority ones. Although the receiving node could easily verify the priority of a transaction
when it calls `CheckTx`, it cannot detect if a peer is sending transactions out of order over a
single P2P channel. For the moment, we leave out of the MVP any mechanism for detecting and possibly
penalizing nodes for this kind of behaviour.

## Alternative designs

### Identify lanes by their priorities

In the initial prototype we identified lanes by their priorities, meaning each priority could only
be assigned to a single lane. This simplified approach proved too restrictive for applications. To
address this, now we identify lanes by `string` names, decoupling lane identifiers from their
priorities.

### One CList for all lanes

We briefly considered sharing one CList for all lanes, changing the internal logic of CList to
accommodate lane requirements. However, this design significantly increases code complexity,
particularly in the transaction dissemination logic.

### One P2P channel per lane

Since P2P channels already have a built-in priority mechanism, they present a reasonable option to
implement transaction dissemination from lanes. By assigning a P2P channel to each lane, we could
simply append new transactions to their respective channels and allow the P2P layer to manage the
order of transmission. We decided against this option mainly because the prioritization logic cannot
be easily modified without altering the P2P code, potentially affecting other non-mempool channels.

Another drawback is that this option imposes a limit to the number of P2P channels. Channels use a
byte as an ID, and the current distribution among all reactors goes up to channel `0x61`. For
example, the current mempool’s P2P channel ID is `0x30`, which would serve as the default lane. We
could reserve a range of channels for the mempool, such as starting from channel ID `0x80` and above
(all channels with the most significant byte set to 1). This would provide a maximum of 128 lanes,
which should suffice for most users.

Nodes would also need to agree on the channel assignments during the P2P handshake. Currently, one
of the conditions for the handshake to succeed is that there must exist an intersection of P2P
channels. Since lanes are defined in the application logic, the nodes only need to agree on the
application version, as it already happens in the current implementation.

### Duality lane/priority

The duality lane/priority could introduce a powerful indirection. The app could just define the lane
of a transaction in `CheckTx`, but the priority of the lane itself could be configured (and
fine-tuned) elsewhere. For example, by the app itself or by node operators. The proposed design for
the MVP does not support this pattern.

### Custom configuration per lane

A straightforward, future improvement that we leave for after the MVP is to allow customized
configuration of the lanes instead of sharing the current mempool configuration among lanes. The
application would need to define new configuration values per lane and pass them to CometBFT during
the handshake.

### Where to define lanes and priorities

We have considered two alternative approaches for _where_ to configure lanes and priorities:
1. In `config.toml` or `app.toml`. We have discarded this option as it does not make sense for
   different nodes to have different lane configurations. The properties defined in the
   specification above are end-to-end, and so, the lane configuration has to be consistent across
   the network.
1. In `ConsensusParams`. There are several disadvantages with this approach. If we allow changing
   lane information via `ConsensusParams`, the mempool would need to update lanes dynamically. The
   updating process would be very complex and cumbersome, and not really appealing for an MVP. Two
   governance proposals would be required to pass to update the lane definitions. A first proposal
   would be required for upgrading the application, because the lane classification logic (thus the
   application's code) needs to know the lane configuration beforehand. And a second proposal would
   be needed for upgrading the lanes via `ConsensusParams`. While it is true that SDK applications
   could pass a governance proposal with both elements together, it would be something to _always_
   do, and it is not clear what the situation would be for non-SDK applications.
  
   Also, it is not clear in which order the proposals should apply. The community should be careful
   not to break performance between the passing of both proposals. The `gov` module could be
   modified to allow the two changes to be shipped in the same gov proposal, but this does not seem
   a feasible solution.

Moreover, these two alternatives have a common problem which is how to deal with nodes that are
late, possibly having lane definitions that do not match with those of nodes at the latest heights.

## Consequences

### Positive

- Application developers will be able to better predict when transactions will be disseminated and
  reaped from the mempool to be included in a block. This has direct impact on block creation and
  transaction latency.
- The mempool will be able to offer Quality of Service (QoS) guarantees, which does not exist in the
  current implementation. This MVP will serve as a base to further extend QoS in future iterations
  of lanes.
- Applications that are unaware of this feature, and therefore not classifying transactions in
  `CheckTx`, will observe the same behavior from the mempool as the current implementation.  

### Negative

- The best-effort FIFO ordering that currently applies to all transactions may be broken when using
  multiple lanes, which will apply FIFO ordering per lane. Since FIFO ordering is important within
  the same class of transactions, we expect this will not be a real problem.
- Increased complexity in the logic of `CheckTx` (ante handlers) in order to classify transactions,
  with a possibility of introducing bugs in the classification logic.

### Neutral

- Lanes are optional. Current applications do not need to make any change to their code. Future
  applications will not be forced to use the lanes feature.
- Lanes will preserve the "FIFO ordering of transactions" property within the same class (with a
  best effort approach, as the current implementation).
- The proposed prioritization algorithm (WRR) for transaction dissemination and block creation is
  fair, so low-priority transactions will not get stuck in the mempool for long periods of time, and
  will get included in blocks proportionally to their priorities.

## References

- [ADR067][adr067], Priority mempool
- [Docstring][reapmaxbytesmaxgas] of `ReapMaxBytesMaxGas`
- Solana's [Gulf Stream][gulf-stream]
- Solana's [Priority Fees][solana-prio-fees]
- Solana's [priority fee pricing][prio-fee-price]
- Cosmos SDK's [gas prices][sdk-gas-prices]
- Cosmos SDK's [application-side mempool][sdk-app-mempool]
- Skip's [Block SDK][skip-block-sdk]
- P2P's [selectChannelToGossipOn][selectChannelToGossipOn] function
- [Weighted Round Robin][wrr]
- [Cosmovisor][cosmovisor]
- [Mempool's cache][cache]

[cache]: https://github.com/cometbft/cometbft/blob/main/mempool/cache.go
[adr067]: ./tendermint-core/adr-067-mempool-refactor.md
[reapmaxbytesmaxgas]: https://github.com/cometbft/cometbft/blob/v0.37.6/mempool/v1/mempool.go#L315-L324
[gulf-stream]: https://medium.com/solana-labs/gulf-stream-solanas-mempool-less-transaction-forwarding-protocol-d342e72186ad
[solana-prio-fees]: https://solana.com/developers/guides/advanced/how-to-use-priority-fees
[prio-fee-price]: https://solana.com/developers/guides/advanced/how-to-use-priority-fees
[sdk-gas-prices]: https://docs.cosmos.network/v0.50/learn/beginner/tx-lifecycle#gas-and-fees
[sdk-app-mempool]: https://docs.cosmos.network/v0.47/build/building-apps/app-mempool
[skip-block-sdk]: https://github.com/skip-mev/block-sdk/blob/v2.1.3/README.md
[cosmovisor]: https://docs.cosmos.network/v0.50/build/tooling/cosmovisor
[selectChannelToGossipOn]: https://github.com/cometbft/cometbft/blob/6d3ff343c2d5a06e7522344d1a4e17d24ce982ad/p2p/conn/connection.go#L542-L563
[wrr]: https://en.wikipedia.org/wiki/Weighted_round_robin
[based-preconfs]: https://ethresear.ch/t/based-preconfirmations/17353
[Oconfs]: https://www.reddit.com/r/btc/comments/vxr3qf/explaining_0_conf_transactions/
[tracking-issue]: https://github.com/cometbft/cometbft/issues/2803
