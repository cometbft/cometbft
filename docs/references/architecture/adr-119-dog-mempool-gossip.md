# ADR 119: Dynamic Optimal Graph (DOG) gossip protocol

## Changelog

- 2024-04-12: Initial notes (@hvanz and @jmalicevic)

## Status
Accepted: [#3297][tracking-issue].


## Context

The current transaction dissemination protocol of the mempool sends each received 
transaction to all the peers a node is connected to, except for the sender. 
While resilient to Byzantine attacks, this type of transaction gossiping is 
causing a lot of network traffic and nodes receiving duplicate transactions
very frequently. 

Benchmarks have also confirmed a large portion of the sent and received bytes 
is due to transaction gossiping. 

DOG is a protocol that aims to reduce the number of duplicate transactions sent
while maintaining the resilience to attacks. 

## Alternative Approaches

> This section contains information around alternative options that are considered
> before making a decision. It should contain a explanation on why the alternative
> approach(es) were not chosen.

## Decision

The CometBFT team has decided to implement the protocol on top of the existing mempool
(`flood` ) dissemination protocol but make it optional. The protocol can be activated/deactivate
via a new config flag `enable_dog_protocol`. More details on this in the sections below. 

## Detailed Design

We start the design section with a general description of how the algorithm works. 

CometBFT, at the moment, caches received transactions. For all transactions, CometBFT stores the IDs of all the peers that have sent it in a `senders` list. When a transaction enters the mempool,
if it was already received it can be retrieved from the cache. When this happens a certain number of times, the receiving node notifies the sender to stop forwarding any transaction from the peer 
that had originally sent the transaction.

Namely, let's assume that transaction `tx_1` was sent by node `B` to node `C` which forwarded it to `A`. `A` already has the transaction and notifies node `C` about this. Node `C` will look up the senders of transaction `tx_1` and in the future disable transaction forwarding from node `B` to node `A`.

In reality, if node `C` ha received `tx_1` from multiple senders (`X`,`Y`, `B`), it will pick the first one that has sent it and disable that route. 
> The protocol implicitly favors routes with low latency, by cutting routes to peers that send the duplicate transaction at a later time.

- Is this true? We don't cut ties with the sender of the transaction, rather we are telling it to cut ties with some other node that has sent us the tx before that. 
So sender could still forward us transactions but not the ones received from this particular node. 

- What if the sender is the only one sending the duplicate transactions (senders ): len(senders) == 1 and senders[0] is itself?  (probably ok)
Should we be doing this? 

Q: What if node C disables a route Y->A because Y is the sender[0]. 

Q: 


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

### Negative

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- {reference link}
