# ADR 119: Dynamic Optimal Graph (DOG) gossip protocol

## Changelog

- 2024-11-04: Initial notes (@hvanz and @jmalicevic)
- 2024-11-25: Applied comments from @hvanz and added information on decided metrics (@jmalicevic)

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

The existing alternative approaches for transaction gossiping are:
- FLOOD - the current transaction gossiping protocol in CometBFT
- CAT Mempool - the transaction gossiping proposed by Celestia
- Limiting number of peers to send a transaction to - Experimental protocol in CometBFT
- Etherium's transaction gossiping protocol.


### FLOOD - the current transaction gossiping protocol in CometBFT

Currently, CometBFT gossips each transaction to all the connected peers, except the sender
of the transaction. Thus, for every transaction, CometBFT keeps a list of senders to make sure
it is not sent again to a node that has sent it. 
This does not prevent a node to send the transaction to a node that has sent it to the 
sender, thus creating a lot of duplicates. 

This is the exact problem DOG was designed to tackle.



## Decision

The CometBFT team has decided to implement the protocol on top of the existing mempool
(`flood` ) dissemination protocol, but make it optional. The protocol can be activated/deactivated
via a new config flag `enable_dog_protocol`. More details on this in the sections below. 

## Detailed Design

We start the design section with a general description of how the algorithm works. 

CometBFT caches received transactions and stores the IDs of all the peers that have sent it in a `senders` list. When a transaction enters the mempool,
if it was already received it can be retrieved from the cache. When this happens a certain number of times, the receiving node notifies the sender to stop forwarding any transaction from the peer 
that had originally sent the transaction. The exact point when the notification is sent is defined below as a redundancy threshold. (TODO LiNK)

Namely, let's assume that transaction `tx_1` was sent by node `B` to node `C` which forwarded it to `A`. `A` already has the transaction and notifies node `C` about this. Node `C` will look up the senders of transaction `tx_1` and in the future disable transaction forwarding from node `B` to node `A`.

In reality, if node `C` ha received `tx_1` from multiple senders (`X`,`Y`, `B`), it will pick the first one that has sent it and disable that route. 

Node `C` keeps a map of `disabledRoutes` per peer and uses it to determine whether a transaction should be forwarded. 

Entries for a particular peer are reset if a peer sends a `Reset` message or if a peer disconnects. 

### Redundancy control

In an ideal setting, receiving a transaction more than once is not needed. But in a Byzantine setting, allowing for only one route for a transaction can introduce an attack vector. 

Operators can thus configure a desired redundancy and the gossip protocol will adjust route disabling based on this setting. Redundancy is defined as the ratio between duplicate and unique transactions. 

<!-- Operators can also define how much they tolerate the redundancy to deviate from the desired level. -->

By default, the redundancy is set to `1` and impacts the gossiping of transactions as follows:


```go
  redundancy := duplicateTxs / firstTimeTxs
  if redundancy < r.redundancyLowerBound:
    peer.send(Reset)
  if redundancy > redundancyUpperBound
    enableRouteDisabling() 
```

The number of unique and duplicate transactions is upadated as transactions come in and,
periodically (every `1s`), DOG recomputes the redundnacy of the system. 


#### Redundancy adjustment triggering

As explained, redundancy is adjusted periodicially. It is also explicitly triggered when a peer is disconnected. <!-- TODO justify why-->

### Impacted areas of code

The bulk of the changes is constrained to the mempool reactor. 


#### Breaking changes and new additions

- The `Mempool` interface is expanded with a method : 
```go
// GetSenders returns the list of node IDs from which we receive the given transaction.
GetSenders(txKey types.TxKey) ([]nodekey.ID, error)
``` 
- The `Entry` interface in the mempool package is expanded with a method:
```go
// Senders returns the list of registered peers that sent us the transaction.
	Senders() []nodekey.ID
```
- We introduce a new communication channel, called `Mempool Control Channel`. The channel
is used to transmit the messages needed for the implementation of the protocol. The channel ID
is `31` and it has a priority of `10` .

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
transmits `HaveTx` and `Reset` messages that keep track of routes between peers. 

`TryAddTx` 

When DOG is enabled we count the number of unique and duplicate transactions, and send
`HaveTx` messages if the redundancy is too high. 

Once a `HaveTx` message is sent,
sending further messages is blocked. It is unblocked by the redundancy control mechanism 
once it is triggered again. This will be discussed below, but overall the frequency
of `HaveTx` messages should be aligned with the time it takes to send a message and 
adjust routing between peers. 

`broadcastTxRoutine`

Before sending a transaction, we filter the peers to send to based on the `disabeldRoutes`. Again,
only if DOG is enabled. 

`RemovePeer`

When a peer is removed, we remove any existing entries related to this peer from the `DisabledRoutes` 
map and explicitly trigger redundancy re-adjustment. The later is not strictly needed, but 
can lead to quicker propagation of this information through the network.

Note that we had concidered a scenario where, upon removing a peer, a node sends `Reset` messages to all its peers. 
This lead to routes beeing disabled too many times, and the process of stabilizing redundancy would be re-triggered
too frequently. The adapted logic relies in the redundancy controller to decide whether a `Reset` 
message should be sent. 


#### New p2p messages

The protocol introduces two new p2p messages whose protobuf definition is given below:

```proto
type HaveTx struct {
	TxKey []byte 
}
```

```
type Reset struct {

}
```

#### Monitoring

The impact of the protocol on operations can also be observed by looking at the following metrics that already exist:

- `AlreadyReceivedTx` - the number of redundant transactions. When DOG is enabled these values should drop. 

- `BytesReceived` - This metric shows the number of bytes received per message type. Without DOG, the transactions dominate the number of bytes, while, when enabled, the block parts dominate.

This ADR introduces a set of metrics which can be used to observe the parameters of the protocol: 


<!-- - `HaveTxMsgsReceived` -  Number of HaveTx messages received (cumulative). 

- `ResetMsgsSent` - Number of Reset messages sent (cumulative). -->

- `DisabledRoutes` - Number of disabled routes.

- `Redundancy` -  The current level of redundancy computed as the ratio between duplicate and unique transactions. 



### Configuration parameters


- `config.EnableDOGProtocol: bool`: Enabling or disabling the DOG protocol. `true` by default.
<!-- TODO verify this. But not a single benchmark
showed a reason to not enable it. --->
- `config.TargetRedundancy: float`: The redundancy level that the gossip protocol should aim to
  maintain. It is `1` by default. It cannot be `0` as this would open the node
  to byzantine attacks.
   <!-- TODO should we remove this too? If yes, remove the above sentence 
  regarding operator configuration -->

In the first version of the protocol, we were inclined to expose more fine tuning
parameters to operators. But after extensive experimentation we have settled on values for 
each of them. 


- `TargetRedundancyDeltaPercent: float`: Value in the range `[0, 1)` that defines the bounds
of acceptable redundancy levels; redundancy +- redundancy*delta TxsPerAdjustment: int. It's value 
is `0.2`  or `20%` of the set redundancy. <!-- We should remove target reduncany, oeoprators can set weird values there -->
- `config.AdjustmentInterval: time.Duration`: Indicates how often the redundancy controler readjusts 
the redundancy and has a chance to trigger sending of `HaveTx` or `Reset` messages. It's value is `1s`. 

#### Deriving the values of the redundancy controller

Part of the work on the protocol was extensive testing of its performance and impact of the network.
Issue [\#4320] covers the experiments performed on DOG. All the findings can be found in [TODO](link)

 We have also tested scenarios where we send the same load to 100 of the 200 nodes and DOG was able 
 to reduce the number of duplicates in the system, even under a high load. 

## User recommendation

- Entire network should use DOG. Otherwise the impact will be minimal. 


> The protocol implicitly favors routes with low latency, by cutting routes to peers that send the duplicate transaction at a later time.

- Is this true? We don't cut ties with the sender of the transaction, rather we are telling it to cut ties with some other node that has sent us the tx before that. 
So sender could still forward us transactions but not the ones received from this particular node. 

- If the frequency of `HaveTx` messages is too high, nodes will have too many routes cut.
- Except for enabling DOG on all nodes, users do not have to change any other aspects of their application. 

## Consequences

Gossiping using DOG leads to a significant reductio in network traffic caused by transactions. 

### Positive

### Negative

- The new protocol is not compatible with the existing experimental feature that limits disseminating transactions up to a specified number of peers. If both are enabled simultaneously, the mempool will work but not as expected. This needs to be clear in the documentation.

### Neutral

## References


[\#4320]: https://github.com/cometbft/cometbft/issues/4320
- {reference link}
