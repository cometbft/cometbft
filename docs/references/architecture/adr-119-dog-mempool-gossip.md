# ADR 119: Dynamic Optimal Graph (DOG) gossip protocol

## Changelog

- 2024-11-04: Initial notes (@hvanz and @jmalicevic)
- 2024-11-25: Applied comments from @hvanz and added information on decided metrics (@jmalicevic)
- 2024-11-27: CAT mempool description (@hvanz)
- 2024-12-04: libp2p gossipSub description (@jmalicevic)

## Status
Accepted: Tracking issue [\#3297].


## Context

The current transaction dissemination protocol of the mempool sends each received 
transaction to all the peers a node is connected to, except for the sender.
In case transactions were received via RPC, they are broadcasted to all
connected peers. 
While this approach provides resilience to Byzantine attacks and optimal transaction latency, it generates significant network traffic, with nodes receiving an exponential number of duplicate transactions.

Benchmarks have also confirmed a large portion of the sent and received bytes 
by the network layer of a node is due to transaction gossiping. 

DOG is a protocol that aims to reduce the number of duplicate transactions received
while maintaining the resilience to attacks and low transaction latency. 

## Alternative Approaches

The existing alternative approaches for transaction gossiping are:
- Flood - the current transaction gossiping protocol in CometBFT
- CAT Mempool - the transaction gossiping proposed by Celestia
- Limiting number of peers to send a transaction to - Experimental protocol in CometBFT
- libp2p gossipSub protocol


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

CometBFT allows operators to configure `p2p.experimental_max_gossip_connections_to_non_persistent_peers` and `p2p.experimental_max_gossip_connections_to_persistent_peers`  as a maximum number of peers to send  transactions to ([#1558](https://github.com/cometbft/cometbft/pull/1558) and [#1584](https://github.com/cometbft/cometbft/pull/1584)). 
This reduces bandwidth compared to when they are disabled, but there is no rule to determine which peers transactions
are not forwarded to. DOG can be considered an enhanced, more informed version of this protocol.

A node can use either this or DOG, but not both at the same time.


### libp2p's gossipSub protocol

 [gossipsub](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md), the gossip protocol of `libp2p`, aims to reduce the bandwidth used for gossiping by limiting the number of peers a node exchange data with. The peering is controlled by a router keeping track of peering state and information about the topics peers are subscribed to. 

 Peers form meshes, where nodes within a mesh are subscribed to the same topic and the size of the mesh is defined with a configuration parameter. 

 `gossipSub` allows peers to indicate they already have a particular transaction so that other peers do not forward it to them, by sending a `IHAVE` message. That, however this is done for each individual message. DOG on the other hand, instructs peers to stop forwarding anything from a given peer with one message. 

 There is [ongoing work](https://github.com/libp2p/specs/pull/413/files) within `gossipSub` that might achieve something similar. A peer can send a message instructing another peer not to forward anything to it. DOG on the other hand blocks only one route, allowing traffic to continue through the peer.

## Decision

The CometBFT team has decided to implement the protocol on top of the existing mempool
(`flood` ) dissemination protocol, but make it optional.
The protocol can be activated/deactivated
via a new config flag `mempool.dog_protocol_enabled`. It will be enabled by default.
 More details on this in the sections below. 

## Detailed Design

We start the design section with a general description of how the algorithm works. The protocol is explained in detail in the accompanying [specification](https://github.com/cometbft/cometbft/blob/main/spec/mempool/gossip/dog.md). 

CometBFT nodes cache received transactions and store the IDs of all the peers that have sent it in a `senders` list. 

Upon receiving a duplicate transaction on node `N` , DOG uses this list to retrieve a `sender`. The `sender` is informed that `N` has received this transaction already. Upon receiving this information, the `sender` looks up who sent this transaction to him, and disables any transaction forwarding between this node and `N`. 

Namely, let's assume that transaction `tx_1` was sent by node `B` to node `C` which forwarded it to `A`. `A` already has the transaction and notifies node `C` about this. Node `C` will look up the senders of transaction `tx_1` and in the future disable transaction forwarding from node `B` to node `A`.

In reality, if node `C` has received `tx_1` from multiple senders (`X`,`Y`, `B`), it will pick the first one that has sent it and disable that route. 

Node `C` keeps a map of `disabledRoutes` per peer and uses it to determine whether a transaction should be forwarded. 

Entries for a particular peer are reset if a peer sends a `ResetRoute` message or if a peer disconnects.
More details on the messages exchanged to achieve this can be found in the section on newly introduced code changes. 


### Redundancy control

In an ideal setting, receiving a transaction more than once is not needed. But in a Byzantine setting, allowing for only one route for a transaction can introduce an attack vector. 

Operators can thus configure a desired redundancy and the gossip protocol will adjust route disabling based on this setting. Redundancy is defined as the ratio between duplicate and unique transactions. 

The redundancy is set to `1` and impacts the gossiping of transactions as follows:


```go
  redundancy := duplicateTxs / firstTimeTxs
  if redundancy < r.redundancyLowerBound:
    peer.send(Reset)
  if redundancy > redundancyUpperBound
    enableRouteDisabling() 
```

The number of unique and duplicate transactions is updated as transactions come in and,
periodically (every `1s`), DOG recomputes the system's redundancy. 

A redundancy of `1` implies that we allow one duplicate transaction for every unique transaction.

#### Redundancy adjustment triggering

As explained, redundancy is adjusted periodically. It is also explicitly triggered when a peer is disconnected. 
While not required to adjusted immediately on a disconnect, we trigger the re-evaluation of existing redundancy 
as an optimization to speed up the time between changes in the network and the state of the system. 

### Impacted areas of code

The changes are constrained to the mempool reactor, and two new p2p control messages.

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
 
#### Breaking changes and new additions

- The `Mempool` interface is expanded with a method : 
```go
// GetSenders returns the list of node IDs from which we have received a transaction.
GetSenders(txKey types.TxKey) ([]p2p.ID, error)
``` 
- The `Entry` interface in the mempool package is expanded with a method:
```go
// Senders returns the list of registered peers that sent us the transaction.
	Senders() []p2p.ID
```
- We introduce a new communication channel, called `MempoolControlChannel`. The channel
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
	disabledRoutes map[p2p.ID]p2pIDSet
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


#### Monitoring

The impact of the protocol on operations can also be observed by looking at the following metrics that already exist:

- `mempool.already_received_txs` - the number of redundant transactions. When DOG is enabled these values should drop. 

- `p2p.message_send_bytes_total` - This metric shows the number of bytes sent per message type. Without DOG, the transactions tend to dominate the number of bytes, while, when enabled, the block parts should dominate it.

- `p2p.message_receive_bytes_total` - This metric shows the number of bytes received per message type. As with the metric above, with DOG the dominating messages should be block parts. 

The received and sent bytes metrics are reported by message type which enables operators to observe the number of `HaveTx` and `ResetRoute` messages sent.


This ADR introduces a set of metrics into the `mempool` module. They can be used to observe the parameters of the protocol: 


- `mempool.disabled_routes` - Number of disabled routes.

- `mempool.redundancy` -  The current level of redundancy computed as the ratio between duplicate and unique transactions. 


### Configuration parameters


- `mempool.dog_protocol_enabled: bool`: Enabling or disabling the DOG protocol. `true` by default.
The only reason for this is the incompatibility of DOG with the existing experimental feature
that disables sending transactions to all peers. 


The [specification](https://github.com/cometbft/cometbft/spec/mempool/gossip/dog.md) of the protocol introduces 3 additional variables. 
Out of the tree, we have made the following configuration parameters:
 
- `mempool.dog_adjust_interval: time.Duration`: Set to `1s` by default. Indicates how often the redundancy controller readjusts 
the redundancy and has a chance to trigger sending of `HaveTx` or `ResetRoute` messages. As with the delta, we 
have not observed a reduction in redundancy due to a lower interval. Most likely due to the fact that, regardless
of the interval, it takes a certain amount of time for the information to propagate through the network.<br/>See [\#4598] for more information about the experiments. 

- `mempool.dog_target_redundancy: float`: Set to `1`. The redundancy level that the gossip protocol should aim to
achieve. An acceptable value, based on our tests was also `0,5`. While still having, for each unique transaction, one duplicate, the number of transaction messages in the system was significantly lower and we did not see a great benefit in using `0.5` over `1`. Therefore we set the default to a safer value in terms of resilience to attacks.
Overall increasing this value above `2` would lead to too many duplicates without an improvement in the tolerance of the system to byzantine attacks. 
<br/>See [\#4597] for more information about the experiments. 

The third parameter defines the bounds of acceptable redundancy levels and cannot be configured: 
- `TargetRedundancyDeltaPercent: float`: Set to `0.1` (10%). It defines the bounds
of acceptable redundancy levels. The actual redundancy will be: `redundancy +- redundancy*delta`. A lower delta 
did not lead to a visible reduction in redundancy, while slightly increasing the number of control messages sent. 
We have therefore opted to not reduce this value further. It does not impact the protocol as much as the target redundancy or the redundancy adjustment interval.<br/>See [\#4569] for more information about the experiments. 

Part of the work on the protocol was extensive testing of its performance and impact on the network, as well 
as the impact of the network configuration and load on the protocol itself. 
Issue [\#4543] covers the experiments performed on DOG.

We have also tested scenarios where we send each transaction to 100 of the 200 nodes. Typically we only send each transaction to one node. DOG was able 
to reduce the number of duplicates in the system, even under a high load, even when the target redundancy is `1`. 

After extensive testing, we did not find compelling evidence and differences in a variety of runs to warrant the
tuning of the default values chosen. We have left the target redundancy and redundancy adjustment interval 
as something operators can adjust so that users can experiment in their testnets. But none of our experiments
found a degradation in performance and correctness for these value. 

More details on the experiments used to make this decision
 can be found in issue [\#4543].



## User recommendation

- The Entire network should use DOG. Otherwise the impact will be minimal. 


- The protocol implicitly favors faster routes, by cutting routes through peers that send the duplicate transaction at a later time.


- If the frequency of `HaveTx` messages is too high, nodes will have too many routes cut.

-  We have performed some tests with a target redundancy of `0.1`. While we did not observe any single node missing out on transactions, we have not thoroughly tested this, in particular in the presence of perturbations (nodes going down) or heavy load. We thus do not advise to lower this parameter without thorough testing. 

- In networks that have slow connections between nodes (latencies bigger than 500ms), it is recommended to increase `config.dog_adjust_interval` to a higher value, at least as high as the maximum round-trip time (RTT) in the network.

## Consequences

Overall DOG reduces the bandwidth used by CometBFT, freeing up the network for consensus related messages. 
For DOG to work as expected, the entire network should be using this, although it is not required for the network
to behave correctly. It should however not be used in combination with the parameters `p2p.experimental_max_gossip_connections_to_non_persistent_peers` and `p2p.experimental_max_gossip_connections_to_persistent_peers` set to `true`.

### Positive

- Except for enabling DOG on all nodes, users do not have to change any other aspects of their application. 
- Gossiping using DOG leads to a significant reduction in network traffic caused by transactions. 
  Moreover, in our experiments most metrics have improved, with no observed degradation in any area. The 
  reduction of redundant messages has a system-wide positive impact, contributing to faster consensus, lower 
  transaction latency, and more efficient resource utilization (#4606).
- Routes with lower latencies are implicitly favoured by cutting routes through peers that sent a duplicate transaction later in time.

### Negative

- It takes around 5 to 15 minutes for nodes to build a stable routing table, depending on network conditions and node configuration. This shouldn't pose a problem as it is expected for nodes to run for a long time.

### Neutral

- The new protocol is not compatible with the existing experimental feature that limits disseminating transactions up to a specified number of peers. If both are enabled simultaneously, the mempool will work but not as expected.

- Mixed networks are supported, nodes not having DOG enabled will not receive the new messages related to the protocol. But the network will not benefit from the protocol as expected. 

## References


[\#4320]: https://github.com/cometbft/cometbft/issues/4320
[\#3297]: https://github.com/cometbft/cometbft/issues/3297
[\#1472]: https://github.com/cometbft/cometbft/pull/1472
[\#2027]: https://github.com/cometbft/cometbft/issues/2027
[\#4569]: https://github.com/cometbft/cometbft/issues/4596
[\#4598]: https://github.com/cometbft/cometbft/issues/4598
[\#4597]: https://github.com/cometbft/cometbft/issues/4597

* [FLOOD](https://github.com/cometbft/cometbft/blob/main/spec/mempool/gossip/flood.md)
* [DOG Specification](https://github.com/cometbft/cometbft/blob/main/spec/mempool/gossip/dog.md). 

