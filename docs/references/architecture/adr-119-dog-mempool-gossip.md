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
that had originally sent the transaction. The exact point when the notification is sent is defined below as a redundancy threshold. (TODO LiNK)

Namely, let's assume that transaction `tx_1` was sent by node `B` to node `C` which forwarded it to `A`. `A` already has the transaction and notifies node `C` about this. Node `C` will look up the senders of transaction `tx_1` and in the future disable transaction forwarding from node `B` to node `A`.

In reality, if node `C` ha received `tx_1` from multiple senders (`X`,`Y`, `B`), it will pick the first one that has sent it and disable that route. 

When a connection to a peer is lost a node sends a `Reset` message to its peers. It
signals to the peers that they should delete all routing information related to this node. This assures that, in case the network topology changes, or a peer becomes unavailable, the node is able to discover transactions via a previously disabled route. 

### Redundancy control

In an ideal setting, receiving a transaction more than once is not needed. But in a Byzantine setting, allowing for only one route for a transaction can introduce an attack vector. 

Operators can thus configure a desired redundancy and the gossip protocol will adjust route disabling based on this setting. Redundancy is defined as the ratio between duplicate and unique transactions. Operators can also define how much they tolerate the redundancy to deviate from the desired level.

By default, the redundancy is set to 1 and impacts the gossiping of transactions as follows:


```
  redundancy := duplicateTxs / firstTimeTxs
  if redundancy < r.redundancyLowerBound:
    peer.send(Reset)
  if redundancy > redundancyUpperBound
    enableRouteDisabling() 
```

Redundancy is adjusted whenever a new transaction is added into the mempool.
(TODO - revisit whether this should also be done on duplicates.)


### Impacted areas of code

#### Mempool reactor


The bulk of the changes is constrained to the mempool reactor. 
We introduce a new communication channel, called `Mempool Control Channel`. The channel
is used to transmit the messages needed for the implementation of the protocol. 

Additionally, the mempool reactor is extended with a 
`gossipRouter` and a `redundancyControl` struct to keep track of the redundancy
of transactions in the system. 
Each node is keeps track of the routes it disabled between its peers. 

```
type gossipRouter struct {
	mtx cmtsync.RWMutex
	// A set of `source -> target` routes that are disabled for disseminating transactions. Source
	// and target are node IDs.
	disabledRoutes map[nodekey.ID]p2pIDSet
}
``` 


```
type redundancyControl struct {
	txsPerAdjustment int64

	// Pre-computed upper and lower bounds of accepted redundancy.
	lowerBound float64
	upperBound float64

	mtx          cmtsync.RWMutex
	firstTimeTxs int64 // number of transactions received for the first time
	duplicates   int64 // number of duplicate transactions

	// If true, do not send HaveTx messages.
	blockHaveTx atomic.Bool
}

```



#### New p2p messages

The protocol introduces two new p2p messages whose protobuf definition is given below:

```
type HaveTx struct {
	TxKey []byte 
}
```

```
type Reset struct {

}
```

#### Monitoring

This ADR introduces a set of metrics which can be used to observe the parameters of the protocol. 

- `HaveTxMsgsReceived`
- `ResetMsgsSent`
- `DisabledRoutes`
- `Redundancy`

The impact of the protocol on operations can also be observed by looking at the following metrics that already exist:

-`AlreadyReceivedTx` - the number of redundant transactions. When DOG is enabled these values should drop. 

-`BytesReceived` - This metric shows the number of bytes received per message type. Without DOG, the transactions dominate the number of bytes, while, when enabled, the block parts dominate.

Newly introduced metrics are: 


- `HaveTxMsgsReceived` -  Number of HaveTx messages received (cumulative). 

- `ResetMsgsSent` - Number of Reset messages sent (cumulative).

- `DisabledRoutes` - Number of disabled routes.

- `Redundancy` -  The current level of redundancy computed as the ratio between duplicate and unique transactions. 



### Configuration parameters


- `config.EnableDOGProtocol: bool`: Enabling or disabling the DOG protocol
- `config.TargetRedundancy: float`: The redundancy level that the gossip protocol should aim to
  maintain.
- `config.TargetRedundancyDeltaPercent: float`: Value in the range `[0, 1)` that defines the bounds
of acceptable redundancy levels; redundancy +- redundancy*delta TxsPerAdjustment: int
- `config.TxsPerAdjustment: int`: How many (first-time) transactions should the node receive before
  attempting to adjust redundancy.
  <!--  TODO : Should this be evaluated also on duplicate transactions. -->

On startup, define the constants:
- `delta := config.TargetRedundancy * config.TargetRedundancyDeltaPercent`
- `redundancyLowerBound := config.TargetRedundancy - delta`
- `redundancyUpperBound := config.TargetRedundancy + delta`

## User recommendation

- Entire network should use DOG


> The protocol implicitly favors routes with low latency, by cutting routes to peers that send the duplicate transaction at a later time.

- Is this true? We don't cut ties with the sender of the transaction, rather we are telling it to cut ties with some other node that has sent us the tx before that. 
So sender could still forward us transactions but not the ones received from this particular node. 

> If the frequency of `HaveTx` messages is too high, nodes will have too many routes cut. 



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

- The new protocol is not compatible with the existing experimental feature that limits disseminating transactions up to a specified number of peers. If both are enabled simultaneously, the mempool will work but not as expected. This needs to be clear in the documentation.

### Neutral

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- {reference link}
