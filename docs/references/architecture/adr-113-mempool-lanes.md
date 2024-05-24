# ADR 113: Mempool Lanes

## Changelog

- 2024-04-12: Initial notes (@hvanz)
- 2024-04-12: Comments on the notes (@sergio-mena)
- 2024-04-17: Discussions (@sergio-mena @hvanz)
- 2024-04-18: Preliminary structure (@hvanz)
- 2024-05-01: Add Context and Properties (@hvanz)
- 2024-05-21: Add more Properties + priority mempool (@sergio-mena)

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
importance and requirements, spanning from IBC messages, transactions for exchanges, for smart
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
a *default class*, which is a special class always present in any set of classes. Because no
transaction can belong to two or more classes, transaction classes form disjoint sets, that is, the
intersection between classes is empty. Also, all transactions in the mempool are the union of the
transactions in all classes.

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

:parking: _Property_ **Consistent transaction classes**: For any transaction `tx`,
and any two correct nodes $p$ and $q$ that receive `tx` for the first time,
$p$ and $q$ MUST have the same set of transaction classes and their relative priority and configuration,
as long as `tx` has not been included in a block.

> TODO: Do we really need to require that the `tx` hasn't been decided? Isn't it adding complexity for nothing?
> Besides, it break modularity, as we use a consensus concept in the mempool.

The property is only required to hold for on-the-fly transactions:
if a node receives a (late) transaction that has already been decided, this property does not enforce anything.
The same goes for duplicate transactions.
Notice that, if this property does not hold, it is not possible to guarantee any property across the network,
such as transaction latency as defined above.

:parking: _Property_ **Consistent transaction classification**: For any transaction `tx`
and any two correct nodes $p$ and $q$ that receive `tx` for the first time,
$p$'s application MUST classify `tx` into the same transaction class as $q$'s application,
as long as `tx` has not been included in a block.

This property only makes sense when the previous property (_consistent transaction classes_) defined above holds.
Even if we ensure consistent transaction classes, if this property does not hold, a given transaction
may not receive the same classification across the network and it will thus be impossible to reason
about any network-wide guarantees we want to provide that transaction with.

Additionally, it is important to note that these two properties also constrain the way transaction
classes and transaction classification logic can evolve in an existing implementation.
If either transaction classes or classification logic are not modified in a coordinated manner in a working system,
there will be at least a period where the these two properties may not hold for all transactions.

## Alternative Approaches

### Priority Mempool

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
1. Since the priority mempool uses FIFO for transactions of equal priority, it also fulfills FIFO ordering per class.
  The problem here is that, since every value of the priority `int64` field is considered a different transaction class,
  there are virtually unlimited traffic classes.
  So it is too easy for an application to end up using hundreds, if not thousands of transactions classes at a given time.
  In this situation, FIFO ordering per class, while fulfilled, becomes a corner case and thus does not add much value.
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

TODO

Other alternatives:
* Consider looking into mempool discussions in Solana
* Ethereum pre-confirmations
* Timebox this: 1-2 days each


## Decision

TODO

## Detailed Design

TO DECIDE:

* bytes per tx, \# of txs, total # of bytes: shared?
  * Current view: **can be different** easily
  * MVP is viable without it, but use cases less customizable
  * Update: How about Injective's use case?

* Two options for introducing several TX lists:
  1. Multiple `Clist`s inside `CListMempool`, the cache stays inside `CListMempool`.
  2. Multiple `CListMempool` instances, the cache needs to be outside `CListMempool` (likely, directly controlled by the reactor)

* After discussion with Osmosis and Injective, there's a dilemma.
  * For Osmosis, a two-lane solution ("native" and "priority") would probably be enough.
    * This means that we could defer the lane and priority configuration, and not tackle it as part of the MVP
  * For Injective, their use case seems to be more demanding: many lanes (4?, 9?),
    with one lane even having less prio than the native lane
    * If we want to address that use case fully, our MVP _does need_ to tackle the design of lane & prio configuration
    * so, a more ambitious MVP, implying more energy, time, and resources. We must be careful with what we promise for when

* Byzantine nodes may gossip transactions on a high-priority lane to get preferential treatment.
  * we may consider banning peers that send transactions with incorrect lane information.
  * Decision was not to ship lane info, but we're still using different channels for different lanes
    * **What shall we do when we see a peer misusing a p2p channel?**
      * not clear what to do, will heavily depend on use case. So
        * let's wait for someone to hit this so we have a use case to solve
        * and less things to do for MVP
      * Updated discussion, looks like the way to go is:
        * Do "stop peer for error" in any case
        * Defer decision about banning policy until a use case pops up


### Decisions on Minimum Viable Product (MVP)

A lane can be viewed as a mempool on its own, with its own config, capacity, cache,
gossip protocol, p2p channel, p2p bandwidth capacity, etc.
For an MVP, probably some (most?) of these components can be common to all lanes.
In further versions, we can see which of these need to be per-lane.
What is clear is that each lane will need its own `CList` data structure, right from the beginning.
* `CList` **need to be different**
  * if using a common `CList`, the broadcast routine becomes overly complicated,
    whereas each lane having its own `CList`, the broadcast routine is very simple
    (advance the highest priority, non-empty lane)
* cache **can be shared** in MVP
  * lock contention: we think it won't be a big deal performance-wise for an MVP
  * cache eviction of entries: lanes with a lot of traffic may force eviction on TXs
    on lanes with little traffic.
    * For an SDK application this is irrelevant: SDK implements replay protection (nonce)
  * it actually doesn't make sense to split the cache per lane, and it would complicate the code significantly.
* p2p channel **need to be different** in MVP for different lanes
  * p2p channels already have priorities implemented.
  * we can implement priorities of lanes by assigning them to different p2p channels
    as a first approach
  * currently mempool uses one channel: we would reserve a _channel range_ for the lanes
* p2p bandwidth capacity **can be shared** in MVP
  * bandwidth capacity per lane is nice to have, but if the MVP doesn't implement it,
    it's still very usable
* gossip protocol: each lane could select its own gossip protocol.
  For the future, when we have more than one gossip protocol to choose from.
* lane/mempool capacity (# of txs, total # of bytes, bytes per tx) **can be different** easily,
  with little extra complexity on config.
  * bytes per tx: ok to share one value for MVP... not so sure
  * \# of txs: probably not all use cases properly addressed if we share a common limit
  * total # of bytes: if the other two are per-lane, this can just be a "generous" value
    (and we'd fine tune the other two)
  * IDEA: Leave this for the end: MVP is viable without it, but use cases less customizable
* re-check **can be shared** in MVP
  * with `PrepareProposal`, it becomes _always_ mandatory, so no point in making it per-lane

ADR: This is good info, should probably go at the beginning,
"These are our decisions on what to implement and what not to.
The rest of the document only tackles the items reported as "to implement" here.

### Lane coordination

There should be some kind of Lane Coordinator that receives uncategorised txs and assigns them a lane.
This could be implemented in the reactor or inside `CListMempool`.
* For an MVP, probably it's easier to put the coordination logic in `CListMempool`.
* There will be a _native_ or _default_ lane, similar to the _native VLAN_ in an 802.1Q network.
* `CheckTx` will (optionally) return lane information, and only then it can be put in a lane.
  (So a tx in a lane has been validated at least once by some node.)
  * If `CheckTx` provides no lane information: native lane
  * Note: if we punish peers (e.g., close connection to them) that send messages on wrong lanes, we will need a property that requires all node to produce consistent lane info in `CheckTx`
* No difference in treatment between TXs coming from RPC, or TXs coming from peers (p2p)
  * Peers aren't trusted
  * In any case, the TXs need to be validated first via `CheckTx`
* The duality lane/priority introduces a powerful indirection
  * The app can just define the lane of a transaction in `CheckTx`, but the priority of the lane
    itself can configured (and fine-tuned) elsewhere (also the app?, operators?)
* Two options for introducing several TX lists:
  1. Multiple `Clist`s inside `CListMempool`, the cache stays inside `CListMempool`.
  2. Multiple `CListMempool` instances, the cache needs to be outside `CListMempool` (likely, directly controlled by the reactor)
* We need to investigate the details of these two options in terms of complexity and risks in the code

ADR: Probably part of [Data flow](#data-flow) section:
1. some bullets are clearly "Entry flow"
2. other are general considerations of all flows, or architectural.
   These are likely to go in a first, introductory subsection of [Data flow](#data-flow) section

### Data flow

* Entry flow:
  * [Rough description]: the steps when a Tx enters the node are unmodified, except
    * `CheckTx` (optionally) returns lane information (if TX valid)
    * The lane information is used to decide to which transaction list the Tx will be added
* Broadcast flow:
  * Higher-priority lanes get to be gossiped first, and gossip within a lane is still in FIFO order.
  * We will need a warning about channels sharing the send queue at p2p level (will need data early on to see if this is a problem for lanes in practice
  * [Rough description]: we need to extend the loop in the broadcast goroutine. Several channels now, not just one
    * To decide
      * if we're going for (simpler) solution where with max 2 lanes,
        the broadcast goroutine itself could manage the priority of lanes within the existing `select` at the end of the loop
      * if we're going for (future-proof) solution where max lanes can be configurable,
        we need to refactor method `TxWaitChan` (or the code that sends the data to that channel) to manage the lane priority there.
        In this case, the broadcast goroutine will be mainly unmodified.
* Reap flow
  * TXs are reaped from higher-priority lanes first, in FIFO order.
  * We go from the reap loop (currently just one), to probably a nested one
    * the outer loop goes through all TX lists (remember, one per lane), in decreasing priority order
    * the current `break` statements in the inner loop (limit of bytes or gas reached), should also break from the outer loop
* Exit flow (Update, Re-Check). Update: unconditionally; Re-Check: only if App says TX is now invalid
  * Transactions in higher-priority lanes are processed (updated) first.
  * Depends on how the multiple TX lists are implemented (see discussion above)
  * Update contains Re-check at the end
  * Remember Update is done while holding the mempool lock (no new CheckTx calls... they have to wait)
  * Changes:
    * We could update the different TX lists in parallel
    * `Update` method (implementing `Update` method in `Mempool` interface) is implemented by `CListMempool` so
      * If we go for several `CListMempool`s, then we'll have to call `Update` on each of them with all the TXs every time
        * or move `Update` method out of `CListMempool` (where?)
      * Also, the way we will do `ReCheck` would be complicated if we had N `Update` calls
        * We will do `ReCheck` lane-by-lane, in FIFO order within a lane
      * So, this seems to make us lean toward: one `CListMempool` containing N `txs` lists
* Each lane informs consensus when there are txs available.
  *  `TxsAvailable` is in our laundry list

ADR: Core part of detailed design. Bullets will likely become subsections

### Who defines lanes and priorities

The list of lanes and their priorities: are consensus parameters? are defined by the app?
* Discussion on _where_ the lane configuration is set up. Three options
  1. `config.toml` / `app.toml`
  2. ConsensusParams
  3. Hardcoded in the application. Info passed to CometBFT during handshake
* Outcome of discussion
  * 1. vs. 2.
    * It does not make sense for different nodes to have different lane configuration,
      so definitely we don't want 1
  * 2. vs. 3.
    * If we can change lane info via `ConsensusParams`
      * CometBFT's mempool needs logic for changing lanes dynamically (complex, not really appealing for MVP)
      * The process of updating lane info very complex and cumbersome:
        * to update lanes you'd need to pass _two_ governance proposals
          * upgrade the app, because the lane classification logic (so, the app's code) needs to know the lane config beforehand
          * then upgrade the lanes via `ConsensusParams`
        * also, not clear in which order: community should be careful not to break performance between the passing of both proposals
        * the `gov` module may allow the two things to be shipped in the same gov proposal,
          but, if we need to do it that way, what's the point in having lanes in `ConsensusParams` ?
  * Conclusion: 3.
    * lane info is "hardcoded" in the app's logic
    * lane info is passed to CometBFT during handshake
  * Advantages:
    * Simple lane update cycle: one software update gov proposal
    * CometBFT doesn't need to deal with dynamic lane changes: it just needs to set up lanes when starting up (whether afresh, or recovery)
* What does lane info (passed to CometBFT) look like?
  * Current state  of discussions.
    * Draft of data structure
        ```protobuf
        message LaneInfo {
          repeated Lane lanes;
        }
        message Lane {
          byte id;
          string name;
        }
        ```
  * The `Lane` list MUST NOT have duplicate lane IDs
  * The order of the `Lane` elements in the `lanes` field defines their priority
  * Lane ID 0 is the native lane.
    * It MAY be present in the list
    * If it is absent, it is equivalent to having it as the last element (lowest priority)
  * Channel ID is a byte
    * Current channel distribution (among reactors) goes up to `0x61`
    * Proposal: reserve for mempool lanes channel ID `0x80` and all channels above (so, all channels whose MSB is 1)
      * max of 128 lanes. Big enough?
    * currently, mempool is p2p channel ID is `0x30`, which would be a special case: native lane.
  * How to deal with nodes that are late? So, lane info at mempool level (and thus p2p channels to establish) does not match.
    * Two cases
    * A node that is up to date and falls behind (e.g., network instability, etc.)
      * Not a problem. For lanes to change we **require** a coordinated upgrade.
        * Reason: if we allow upgrade changing lane info **not** to be a coordinated upgrage,
          then we won't have consistent lane info across nodes (each node would be running a particular version of the software)
    * A node that is blocksyncing from far away in the past (software upgrades in the middle)
      * Normal channels (including `0x30`): same as before
      * Lane channels (>=`0x80`), not declared. The node can't send/receive mempool-related info. Not a problem, since it's blocksyncing
        * If we're not at the latest version (e.g., Cosmovisor-drive blocksync)
          * channel info likely to be wrong
          * but we Cosmovisor will kill the node before switching to consensus
    * We will need to extend the channel-related handshake between nodes to be able to add channels after initial handshake
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

ADR: This deserves its own section of the detailed design, likely after the one describing the transaction flows

### Tmp notes (likely to be deleted)

* The `Tx` message needs an extra `lane` field? If empty, the transaction doesn't have a category (and the message is compatible with the current format).
  * yes, _native lane_ (see above)
  * Update: we decided to
    * use different p2p channels for different lanes (we will need to _reserve_ a channel range for this in p2p)
    * not trust info from peers, since we anyway need to run `CheckTx` to check the TX's validity
      * so not trusting lane info from peers is aligned with our current model on checking validity of TXs received from peers
  * So, at least in a first version, the lane info won't be shipped with TXs when broadcasting them
  * TODO: we leave this bullet point here ATM, but we'll probably just deleted it
* Before implementing lanes, is it better to first modularise the current mempool?
  * we currently think improving modularisation is not gating for a mempool lanes MVP
    * BUT, some things in the
      _[modularity laundry list](https://www.notion.so/informalsystems/8a887f27e40b45dead689f2d1762b778?v=30a6fd81673b483fb16f4b7a9fc311fb)_
      will be gating (e.g., `TxsAvailable`)

## Consequences

TODO

### Positive

### Negative

### Neutral

## References

- [ADR067][adr067], Priority mempool
- [Docstring][reapmaxbytesmaxgas] of `ReapMaxBytesMaxGas`

[adr067]: ./tendermint-core/adr-067-mempool-refactor.md
[reapmaxbytesmaxgas]: https://github.com/cometbft/cometbft/blob/v0.37.6/mempool/v1/mempool.go#L315-L324