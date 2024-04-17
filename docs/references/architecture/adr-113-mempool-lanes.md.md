# ADR 113: Mempool Lanes

## Changelog

- 2024-04-12: Initial notes (@hvanz)
- 2024-04-12: Comments on the notes (@hvanz)

## Status

In writing (we're not sure yet whether this will be an ADR or a spec)

## Context

TODO

## Alternative Approaches

TODO

## Decision

TODO

## Detailed Design

Initial notes:
* Each lane is a mempool on its own, with its own config, capacity, cache, gossip protocol, p2p channel, p2p bandwidth capacity, etc.
  * for an MVP, probably some (most?) of those can be common to all lanes.
    In further versions, we can see which of those above need to be per-lane
  * it seems clear to me that each lane will need its own Clist, right from the beginning
    * cache **can be shared** in MVP
      * lock contention: we think it won't be a big deal performance-wise for an MVP
      * replacement of entries: lanes with a lot of traffic may force replacement on TXs
        on lanes with little traffic.
        * For an SDK application this is irrelevant: SDK implements replay protection (nonce)
    * p2p channel **need to be different** in MVP for different lanes
      * we can implement priorities of lanes by assigning them to different p2p channels
        as a first approach
      * currently mempool uses one channel: we would assign it a _channel range_
    * gossip protocol: each lane can select its own gossip proto.
      For the future, when we have more than 1 gossip proto to choose from.
    * clist **need to be different**
      * if using a common clist, the broadcast routine becomes overly complicated,
        whereas each lane having its own clist, the broadcast routine is very simple
        (advance the highest priority, non-empty lane)
    * p2p bandwidth capacity **can be shared** in MVP
      * bandwidth capacity per lane is nice to have, but if the MVP doesn't implement it,
        it's still very usable
    * lane/mempool capacity (# of txs, total # of bytes, bytes per tx) **can be different** easily,
      with little extra complexity on config.
      * bytes per tx: ok to share one value for MVP... not so sure
      * \# of txs: probably not all uses cases not properly addressed if we share a common limit
      * total # of bytes: if the other two are per-lane, this can just be a "generous" value
        (and we'd fine tune the other two)
      * IDEA: Leave this for the end: MVP is viable without it, but use cases less customizable
    * re-check **can be shared** in MVP
      * with PrepareProposal, it becomes _always_ mandatory, so no point in making it per-lane
* Additionally, there should be a Lane Coordinator, with a cache for uncategorised txs. These are transactions received from RPC endpoints.
  * After doing CheckTx, the transaction will have a category, and only then it can be put in a lane. (So a tx in a lane has been validated at least once by some node.)
  * Results of discussion:
    * No difference in treatment between TXs coming from RPC, or TXs coming from peers (p2p)
      * Peers aren't trusted
      * In any case, The TXs needs to be validated via CheckTx
    * Reminder (from above):
      * Cache will be shared (in MVP)
      * TX lists will be different, one per lane
    * Not sure yet how to introduce several TX lists. Options:
      * multiple `Clist` inside `CListMempool`, the cache stays inside `CListMempool`
      * multiple `CListMempool` instances, the cache needs to be outside `CListMempool` (likely, directly controlled by the reactor)
    * Entry flow:
      * [Rough description]: the steps when a Tx enters the node are unmodified, except
        * CheckTx (optionally) returns lane information (if TX valid)
        * The lane information is used to decide to which transaction list the Tx will be added
    * Broadcast flow:
      * [Rough description]: we need to extend the loop in the broadcast goroutine. Several channels now, not just one
        * To decide
          * if we're going for (simpler) solution where with max 2 lanes,
            the broadcast goroutine itself could manage the priority of lanes within the existing `select` at the end of the loop
          * if we're going for (future-proof) solution where max lanes can be configurable,
            we need to refactor method `TxWaitChan` (or the code that sends the data to that channel) to manage the lane priority there.
            In this case, the broadcast goroutine will be mainly unmodified.
    * Reap flow
      * We go from the reap loop (currently just one), to probably a nested one
        * the outer loop goes through all TX lists (remember, one per lane), in decreasing priority order
        * the current `break` statements in the inner loop (limit of bytes or gas reached), should also break from the outer loop
    * Exit flow (Re-Check, Update). Update: unconditionally; Re-Check, only if App says TX is now invalid
      * Depends on how the multiple TX lists are implemented (see discussion above)
      * Update contains Re-check at the end
      * Remember Update is done while holding the mempool lock (not further CheckTx... they have to wait)
      * Changes:
        * We could update the different TX lists in parallel
        * `Update` method (implementing `Update` method in `Mempool` interface) is implemented by `CListMempool` so
          * If we go for several `CListMempool`s, the we'll have to call `Update` on each of them with all the TXs every time
            * or move `Update` method out of `CListMempool` (where?)
          * Also, the way we will do `ReCheck` would complicated if we had N `Update` calls
            * We will do `ReCheck` lane-by-lane, in FIFO order within a lane
          * So, this seems to make us lean toward: 1 `CListMempool` containing N `txs` lists
* If each lane has a priority, then:
  * Consensus reaps transactions from higher-priority lanes first, in fifo order
  * Higher-priority lanes get to be gossiped first, and gossip within a lane is still in FIFO order
  * Transactions in higher-priority lanes are processed (CheckTx'd) first.
  * In my view, there will be a _native_ or _default_ lane, similar to the _native VLAN_ in an 802.1Q network.
  * I would add that the duality lane - priority introduces a powerful indirection
    * The app can just define the lane of a transaction in `CheckTx`, but the priority of the lane
      itself can configured (and fine-tuned) elsewhere (also the app?, operators?)
  * After discussion with Osmosis and Injective, there's a dilemma
    * For Osmosis, a two-lane solution ("native" and "priority") would probably be enough.
      * This means that we could defer the lane and priority configuration, and not tackle it as part of the MVP
    * For Injective, their use case seems to be more demanding: many lanes (4?, 9?),
      with one lane even having less prio than the native lane
      * If we want to address that use case fully, our MVP _does need_ to tackle the design of lane & prio configuration
      * so, a more ambitious MVP, implying more energy, more time. We must be careful with what we promise for when
* Each lanes informs consensus when there are txs available.
  *  `TxsAvailable` is in our laundry list
* The list of lanes and their priorities: are consensus parameters? are defined by the app?
  *  $10**6 question :-)
    * let's discuss it next week with our users. Update: Didn't happen so far :-)
    * TODO
      * Part 1: static config. How-to
      * Part 2: config changes; what happens with ongoing TXs

Some considerations:

* The `Tx` message needs an extra `lane` field. If empty, the transaction doesn't have a category (and the message is compatible with the current format).
  * yes, _native lane_ (see above)
  * Update: we decided to
    * use different p2p channels for different lanes (we will need to _reserve_ a channel range for this in p2p)
    * not trust info from peers, since we anyway need to run `CheckTx` to check the TX's validity
      * so not trusting lane info from peers is aligned with our current model on checking validity of TXs received from peers
  * So, at least in a first version, the lane info won't be shipped with TXs when broadcasting them
* Byzantine nodes may gossip transactions on a high-priority lane to get preferential treatment.
  * Idea: we may consider banning peers that send transactions with incorrect lane information.
    * Decision was not to ship lane info, but we're still using different channels for different lanes
      * What shall we do when we see a peer misusing a p2p channel?
        * Likely not clear what to do, will heavily depend on use case. So
          * let's wait for someone to hit this so we have a use case to solve
          * and less things to do for MVP
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
