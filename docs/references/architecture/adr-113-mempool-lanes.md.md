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
* Additionally, there should be a Lane Coordinator, with a cache for uncategorised txs. These are transactions received from RPC endpoints.
  * After doing CheckTx, the transaction will have a category, and only then it can be put in a lane. (So a tx in a lane has been validated at least once by some node.)
  * @sergio-mena: In my view, there will be a _native_ or _default_ lane, similar to the _native VLAN_ in an 802.1Q network.
  * @sergio-mena: Open question: does `CheckTx` attach a "category"? Or directly a lane?
* Each lane has a priority then:
  * Consensus reaps transactions from higher-priority lanes first.
  * Higher-priority lanes get to be gossiped first (routing).
  * Transactions in higher-priority lanes are processed (CheckTx'd) first.
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
