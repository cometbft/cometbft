# ADR 119: Dynamic Optimal Graph (DOG) gossip protocol

## Changelog

- 2024-11-04: Initial notes (@hvanz and @jmalicevic)
- 2024-11-25: Applied comments from @hvanz and added information on decided metrics (@jmalicevic)

## Status
Accepted: [\#3297][tracking-issue].


## Context

The current transaction dissemination protocol of the mempool sends each received 
transaction to all the peers a node is connected to, except for the sender.
In case transactions were received via RPC, they are broadcasted to all
connected peers. 
While resilient to Byzantine attacks, this type of transaction gossiping is 
causing a lot of network traffic and nodes receiving duplicate transactions
very frequently. 

Benchmarks have also confirmed a large portion of the sent and received bytes 
by the network layer of a node is due to transaction gossiping. 

DOG is a protocol that aims to reduce the number of duplicate transactions received
while maintaining the resilience to attacks. 

## Alternative Approaches

The existing alternative approaches for transaction gossiping are:
- FLOOD - the current transaction gossiping protocol in CometBFT
- CAT Mempool - the transaction gossiping proposed by Celestia
- Limiting number of peers to send a transaction to - Experimental protocol in CometBFT
- Etherium's transaction gossiping protocol.


### FLOOD - the current transaction gossiping protocol in CometBFT

[Flood](https://github.com/cometbft/cometbft/blob/main/spec/mempool/gossip/flood.md) is the protocol currently used by CometBFT to gossip transactions. If a transaction
was received via RPC, it is gossiped to all connected peers. Transactions received from other peers
are forwarded to all the peers which have not already sent this particular transaction.

This is enabled by CometBFT keeping the ID of each peer that has sent a transaction.
This does not prevent a node to send the transaction to a node that has already
received this transactions from its peers, thus creating a lot of duplicates. 

This is the exact problem DOG was designed to tackle.

### CAT protocol

[Content-Addressable Transaction (CAT)](https://github.com/celestiaorg/celestia-core/blob/feature/cat/docs/celestia-architecture/adr-009-cat-pool.md) is a "push-pull" gossip protocol originally implemented by Celestia on top of the priority-mempool (aka v1), which existed in Tendermint Core until v0.36 and CometBFT until v0.37. 

With CAT, nodes forward ("push") transactions received from users via RPC endpoints to their peers. When a node `A` receives a transaction from another node, it will notify its peers that it has the transaction with a `SeenTx(hash(tx))` message. A node `B`, upon receiving the `SeenTx` message from `A`, will check whether the transaction is in its mempool. If not, it will "pull" the transaction from `A` by sending a `WantTx(hash(tx))` message, to which `A` responds with a `Tx` message carrying the full transaction. While `SeenTx` and `HaveTx` messages are much smaller than `Tx` messages, these additional communication steps introduce latency when disseminating transactions. The CAT protocol was especially designed for networks with low throughput and large transaction sizes. 

Efforts to port the CAT mempool to CometBFT were documented in [\#2027]. Experimental results on a small testnet from [\#1472] showed that CAT effectively reduces bandwidth usage. However, its impact on latency was not evaluated in those tests. Porting CAT was finally deprioritised in favor of DOG.




### Limiting the number of peers a transaction is forwarded to

CometBFT allows operators to configure `p2p.experimental_max_gossip_connections_to_non_persistent_peers` and `p2p.experimental_max_gossip_connections_to_persistent_peers`  as a maximum number of peers to send  transactions to. 
This reduces bandwitdth compared to when they are disabled but there is no rule to determine which peers transactions
are not forwarded to. DOG can be looked at as an enahanced, more informed version of this protocol.

A node should use either this or DOG. 

## Decision

The CometBFT team has decided to implement the protocol on top of the existing mempool
(`flood` ) dissemination protocol, but make it optional. The protocol can be activated/deactivated
via a new config flag `mempool.enable_dog_protocol`. More details on this in the sections below. 

## Detailed Design

We start the design section with a general description of how the algorithm works. The protocol is explained in detail in the accompanying [specification](https://github.com/cometbft/cometbft/spec/mempool/gossip/dog.md). 

CometBFT caches received transactions and stores the IDs of all the peers that have sent it in a `senders` list. When a transaction enters the mempool,
if it was already received it can be retrieved from the cache. When this happens a certain number of times, the receiving node notifies the sender to stop forwarding any transaction from the peer 
that had originally sent the transaction. The exact point when the notification is sent is defined below as a redundancy threshold. (TODO LiNK)

Namely, let's assume that transaction `tx_1` was sent by node `B` to node `C` which forwarded it to `A`. `A` already has the transaction and notifies node `C` about this. Node `C` will look up the senders of transaction `tx_1` and in the future disable transaction forwarding from node `B` to node `A`.

In reality, if node `C` ha received `tx_1` from multiple senders (`X`,`Y`, `B`), it will pick the first one that has sent it and disable that route. 

Node `C` keeps a map of `disabledRoutes` per peer and uses it to determine whether a transaction should be forwarded. 

Entries for a particular peer are reset if a peer sends a `ResetRoute` message or if a peer disconnects. 

### Redundancy control

In an ideal setting, receiving a transaction more than once is not needed. But in a Byzantine setting, allowing for only one route for a transaction can introduce an attack vector. 

Operators can thus configure a desired redundancy and the gossip protocol will adjust route disabling based on this setting. Redundancy is defined as the ratio between duplicate and unique transactions. 

<!-- Operators can also define how much they tolerate the redundancy to deviate from the desired level. -->

The redundancy is set to `1` and impacts the gossiping of transactions as follows:


```go
  redundancy := duplicateTxs / firstTimeTxs
  if redundancy < r.redundancyLowerBound:
    peer.send(Reset)
  if redundancy > redundancyUpperBound
    enableRouteDisabling() 
```

The number of unique and duplicate transactions is updated as transactions come in and,
periodically (every `1s`), DOG recomputes the redundnacy of the system. 

A redundancy of `1` implies that we allow for each unoque we tolerate a duplicate of the same transaction.

#### Redundancy adjustment triggering

As explained, redundancy is adjusted periodicially. It is also explicitly triggered when a peer is disconnected. 
While not required to adjusted immediately on a disconnect, we trigger the re-evaluation of existing redundancy 
as an optimization to speed up the time between changes in the network and the state of the syetem. 

### Impacted areas of code

The changes are constrained to the mempool reactor. 


#### Breaking changes and new additions

- The `Mempool` interface is expanded with a method : 
```go
// GetSenders returns the list of node IDs from which we have received a transaction.
GetSenders(txKey types.TxKey) ([]nodekey.ID, error)
``` 
- The `Entry` interface in the mempool package is expanded with a method:
```go
// Senders returns the list of registered peers that sent us the transaction.
	Senders() []nodekey.ID
```
- We introduce a new communication channel, called `Mempool Control Channel`. The channel
is used to transmit the messages needed for the implementation of the protocol. The channel ID
is `31` with priority `10` (for context, the mempool channel sending transaction has a priority of `5`).

Additionally, the mempool reactor is extended with a 
`gossipRouter` and a `redundancyControl` struct to keep track of the redundancy
of transactions in the system. 
Each node keeps track of the routes it disabled between its peers. 

```go
type gossipRouter struct {
	mtx cmtsync.RWMutex
	// A set of `source -> target` routes that are disabled for disseminating transactions. Source
	// and target are node IDs.
	disabledRoutes map[nodekey.ID]p2pIDSet
}
``` 


```go
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


#### Added logic to existing functions

The overall behaviour of existing functions remains the same when DOG is disabled. 

Enabling DOG causes the following changes in behaviour:

`Receive`

The function now handles messages from two channels. The `MempoolControlChannel`
transmits `HaveTx` and `ResetRoute` messages that keep track of routes between peers. 

`TryAddTx` 

When DOG is enabled we count the number of unique and duplicate transactions, and send
`HaveTx` messages if the redundancy is too high. 

Once a `HaveTx` message is sent,
sending further `HaveTx` messages is blocked. It is unblocked by the redundancy control mechanism 
once it is triggered again. This will be discussed below, but overall the frequency
of `HaveTx` messages should be aligned with the time it takes to send a message and 
adjust routing between peers. 

`broadcastTxRoutine`

Before sending a transaction, we filter the peers to send to based on the `disabledRoutes`. Again,
only if DOG is enabled. 

`RemovePeer`

When a peer is removed, we remove any existing entries related to this peer from the `disabledRoutes` 
map and explicitly trigger redundancy re-adjustment. The later is not strictly needed, but 
can lead to quicker propagation of this information through the network.

Note that we had considered a scenario where, upon removing a peer, a node sends `ResetRoute` messages to all its peers. 
This lead to routes being disabled too many times, and the process of stabilizing redundancy would be re-triggered
too frequently. The adapted logic relies in the redundancy controller to decide whether a `ResetRoute` 
message should be sent. 


#### New p2p messages

The protocol introduces two new p2p messages whose protobuf definition is given below:

```proto
message HaveTx {
  bytes tx_key = 1;
}
```

```proto
message ResetRoute {
}
```

We expand the abstract definition of a mempool message with these message types as well:


```proto
// Message is an abstract mempool message.
message Message {
  // Sum of all possible messages.
  oneof sum {
    Txs txs = 1;
    HaveTx have_tx = 2;
    ResetRoute reset_route = 3;
  }
}
```

#### Monitoring

The impact of the protocol on operations can also be observed by looking at the following metrics that already exist:

- `mempool.already_received_txs` - the number of redundant transactions. When DOG is enabled these values should drop. 

- `p2p.message_send_bytes_total` - This metric shows the number of bytes sent per message type. Without DOG, the transactions tend to dominate the number of bytes, while, when enabled, the block parts should dominate it.

- `p2p.message_receive_bytes_total` - This metric shows the number of bytes received per message type. As with the metric abobe, with DOG the dominating messages should be block parts.


This ADR introduces a set of metrics into the `mempool` module. They can be used to observe the parameters of the protocol: 


<!-- - `HaveTxMsgsReceived` -  Number of HaveTx messages received (cumulative). 

- `ResetMsgsSent` - Number of Reset messages sent (cumulative). -->

- `mempool.disabled_routes` - Number of disabled routes.

- `mempool.redundancy` -  The current level of redundancy computed as the ratio between duplicate and unique transactions. 


### Configuration parameters


- `mempool.enable_dog_protocol: bool`: Enabling or disabling the DOG protocol. `false` by default.
The only reason for this is the incompatibility of DOG with the existing experimental feature
that disables sending transactions to all peers. 


The [specification]() of the protocol introduces 3 additional variables. 
In the first implementation of the protocol, we were inclined to expose them as configuration parameters.

Part of the work on the protocol was extensive testing of its performance and impact on the network, as well 
as the impact of the network configuration and load on the protocol itself. 
Issue [\#4320] covers the experiments performed on DOG. All the findings can be found in [TODO](link)

We have also tested scenarios where we send the same load to 100 of the 200 nodes. DOG was able 
to reduce the number of duplicates in the system, even under a high load. 

After extensive testing, we did not find compelling evidence and differences in a variety of runs to varrant the
tuning the default values chosen. Therefore, in the first version of the protocol,
we are fixing a set of values for them. 

More details on the experiments used to make this decision
 can be found in Issue [\#4320].

- `mempool.target_redundancy: float`: Set to `1`. The redundancy level that the gossip protocol should aim to
achieve. Increasing this value above `2` would lead to too many duplicates. Lowering it would reduce the redundancy 
but not by a very significant amount during high load or network updates.<!-- TODO verify this claim i nthe results for perturbations and load --> As lower values introduce the potential
for attacks on nodes, we decided to leave the value of `1` as the target redundancy. 


- `TargetRedundancyDeltaPercent: float`: Set to `0.2` (20%). It defines the bounds
of acceptable redundancy levels. The actual redundancy will be: `redundancy +- redundancy*delta`. A lower delta 
did not lead to a visible reduction in redundancy, while slightly increasing the number of control messages sent. 
We have therefore opted to not reduce this value further. 

- `config.AdjustmentInterval: time.Duration`: Set to `1s`. Indicates how often the redundancy controller readjusts 
the redundancy and has a chance to trigger sending of `HaveTx` or `ResetRoute` messages. As with the delta, we 
have not observed a reduction in redundancy due to a lower interval. Most likely due to the fact that, regardless
of the interval, it takes a certain amount of time for the information to propagate through the network. 

## User recommendation

- The Entire network should use DOG. Otherwise the impact will be minimal. 


> The protocol implicitly favors routes with low latency, by cutting routes to peers that send the duplicate transaction at a later time.

- Is this true? We don't cut ties with the sender of the transaction, rather we are telling it to cut ties with some other node that has sent us the tx before that. 
So sender could still forward us transactions but not the ones received from this particular node. 

- If the frequency of `HaveTx` messages is too high, nodes will have too many routes cut.


## Consequences

Overall DOG reduces the bandwidth used by CometBFT, freeing up the network for consensus related messages. 
For DOG to work as expected, the entire network should be using this, although it is not required for the network
to behave correctly. It should however not be used in combination with the parameters `p2p.experimental_max_gossip_connections_to_non_persistent_peers` and `p2p.experimental_max_gossip_connections_to_persistent_peers` set to `true`.

### Positive

- Except for enabling DOG on all nodes, users do not have to change any other aspects of their application. 
- Gossiping using DOG leads to a significant reduction in network traffic caused by transactions. 

### Negative

- The new protocol is not compatible with the existing experimental feature that limits disseminating transactions up to a specified number of peers. If both are enabled simultaneously, the mempool will work but not as expected.

### Neutral

- At this point, DOG cannot be fine tuned with too many parameters. The reason for this is that the protocol was very resilient based on all our tests. But this is something that might be revisited when users start using it in production. 
- Mixed networks are supported, nodes not having DOG enabled will not receive the new messages related to the protocol. But the network will not benefit from the protocol as expected. 

## References


[\#4320]: https://github.com/cometbft/cometbft/issues/4320
[\#3297]: https://github.com/cometbft/cometbft/issues/3297
[\#1472]: https://github.com/cometbft/cometbft/pull/1472
[\#2027]: https://github.com/cometbft/cometbft/issues/2027

