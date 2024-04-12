# ADR 113: Mempool Lanes

## Changelog

- 2024-04-12: Initial notes from @hvanz

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
* Additionally, there should be a Lane Coordinator, with a cache for uncategorised txs. These are transactions received from RPC endpoints.
  * After doing CheckTx, the transaction will have a category, and only then it can be put in a lane. (So a tx in a lane has been validated at least once by some node.)
* Each lane has a priority then:
  * Consensus reaps transactions from higher-priority lanes first.
  * Higher-priority lanes get to be gossiped first (routing).
  * Transactions in higher-priority lanes are processed (CheckTx'd) first.
* Each lanes informs consensus when there are txs available.
* The list of lanes and their priorities: are consensus parameters? are defined by the app?

Some considerations:

* The `Tx` message needs an extra `lane` field. If empty, the transaction doesn't have a category (and the message is compatible with the current format).
* Byzantine nodes may gossip transactions on a high-priority lane to get preferential treatment. We may consider banning peers that send transactions with incorrect lane information.
* Before implementing lanes, is it better to first modularise the current mempool?

## Consequences

TODO

### Positive

### Negative

### Neutral

## References

TODO

- Reference list
