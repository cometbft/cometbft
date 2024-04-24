# ADR 113: Mempool Lanes

## Changelog

- 2024-04-12: Initial notes (@hvanz)
- 2024-04-12: Comments on the notes (@hvanz)
- 2024-04-17: Discussions (@sergio-mena @hvanz)
- 2024-04-18: Preliminary structure (@hvanz)

## Status

In writing (we're not sure yet whether this will be an ADR or a spec)

## Context

TODO

Try using "QoS" term. Because that's what it actually is.

We need problem definition.


Definitions
Tx1, Tx2
Traffic classes
Traffic class order (partial?)
Latency(tx) --- important

Properties / Requirements (similar to abci++ requirement)

[Examples]
* Property 1: tx1 --> tc1, tx2 --> tc1 ---> FIFO

* Property 2: tx1 --> tc1, tx2 --> tc2; tc1 < tc2 --> tx1 first, then tx2

* Property 3: tx1 --> tc1, tx2 --> tc2; tc1 < tc2 --> latency(tx1)<=latency(tx2)

Try to come up with properties:
* Currently, these properties not guaranteed
* The rest of this ADR, proposal to guarantee these properties

Note: Consider explaining that the problem we are trying to solve is **not** general bandwidth or latency improvements: that is an orthogonal problem.

## Alternative Approaches

TODO

Prio mempool: why is it not a good fit?

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

TODO

- Reference list
