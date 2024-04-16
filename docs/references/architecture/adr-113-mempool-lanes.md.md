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
  * @sergio-mena: for an MVP, probably some (most?) of those can be common to all lanes.
    In further versions, we can see which of those above need to be per-lane
  * @sergio-mena: it seems clear to me that each lane will need its own Clist, right from the beginning
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
    * Exit flow (Re-Check, Update). Update: unconditionally; Re-Check, only of App say TX is now invalid
      * Depends on how the multiple TX lists are implemented (see discussion above)
      * WE'RE HERE

* Each lane has a priority then:
  * Consensus reaps transactions from higher-priority lanes first.
  * Higher-priority lanes get to be gossiped first (routing).
  * Transactions in higher-priority lanes are processed (CheckTx'd) first.
  * @sergio-mena: In my view, there will be a _native_ or _default_ lane, similar to the _native VLAN_ in an 802.1Q network.
  * @sergio-mena: I would add that the duality lane - priority introduces a powerful indirection
    * The app can just define the lane of a transaction in `CheckTx`, but the priority of the lane
      itself can configured (and fine-tuned) elsewhere (also the app?, operators?)
* Each lanes informs consensus when there are txs available.
  * sergio-mena: `TxsAvailable` is in our laundry list
* The list of lanes and their priorities: are consensus parameters? are defined by the app?
  * sergio-mena: $10**6 question :-)
    * let's discuss it next week with our users

Some considerations:

* The `Tx` message needs an extra `lane` field. If empty, the transaction doesn't have a category (and the message is compatible with the current format).
  * sergio-mena: yes, _native lane_ (see above)
* Byzantine nodes may gossip transactions on a high-priority lane to get preferential treatment. We may consider banning peers that send transactions with incorrect lane information.
  * sergio-mena: good point, we need to discuss further. A good first-hand solution is
    * every node attaches the lanes to the TXs received locally, just as every node calls `CheckTx` today
    * nodes transmit TXs without lane information, as they're not supposed to trust each other
* Before implementing lanes, is it better to first modularise the current mempool?
  * sergio-mena: my guess is _no_, at least for an MVP

## Consequences

TODO

### Positive

### Negative

### Neutral

## References

TODO

- Reference list
