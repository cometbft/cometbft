# Dynamic Optimal Graph (DOG) gossip protocol

The DOG protocol optimizes network efficiency by dynamically managing how transactions are routed to
peers: if node A receives a transaction from node B that it already has in its cache, it implies
that there is a cycle in the network topology. Node A will message B to stop sending transactions,
and B will close one of the "routes" that sends transactions to A, thus cutting the cycle.

Additionally, for keeping nodes resilient to Byzantine attacks, the protocol has a Redundancy
Controller that maintains a minimum, pre-defined level of transaction redundancy. If a node is not
receiving enough duplicate transactions, it will message its peers to request additional ones.

The DOG protocol is built on top of the [Flood protocol](flood.md). DOG's spec uses many of the same
types, messages, and data structures as Flood. In principle, it is possible to enable or disable DOG
in nodes running Flood. Nodes running DOG and Flood can co-exist in the network, though performance
will be optimal only if all nodes enable DOG.

**Table of contents**
  - [Messages](#messages)
  - [Routing](#routing)
    - [Functions on routes](#functions-on-routes)
  - [Redundancy Control](#redundancy-control)
    - [Redundancy level](#redundancy-level)
  - [Parameters](#parameters)
    - [Target redundancy](#target-redundancy)
      - [Target bounds](#target-bounds)
    - [Number of transactions per adjustment](#number-of-transactions-per-adjustment)
  - [Initial state](#initial-state)
  - [State transitions (actions)](#state-transitions-actions)
    - [Adding transactions to the mempool](#adding-transactions-to-the-mempool)
      - [Adding first-time transactions](#adding-first-time-transactions)
      - [Handling duplicate transactions](#handling-duplicate-transactions)
    - [Handling incoming messages](#handling-incoming-messages)
      - [Handling HaveTx messages](#handling-havetx-messages)
      - [Handling Reset messages](#handling-reset-messages)
    - [Transaction dissemination](#transaction-dissemination)
    - [Nodes disconnecting from the network](#nodes-disconnecting-from-the-network)

> This document was written using the literature programming paradigm. Code snippets are written in
> [Quint][quint] and can get "tangled" into a Quint file.

## Messages

In addition to the `TxMsg` data message, DOG has two control messages. The size of the control
messages is negligible compared to `TxMsg`, which carries a full transaction.
```bluespec "messages" +=
type Message =
```

* Transaction message (same as in Flood).
```bluespec "messages" +=
    | TxMsg(TX)
```

* A node sends `HaveTxMsg` messages to signal that it already received the transaction. The receiver
  will cut a route related to tx that is forming a cycle in the network topology.
```bluespec "messages" +=
    | HaveTxMsg(TxID)
```

* A node sends `ResetMsg` messages to signal that it is not receiving enough transactions. The
  receiver should re-enable some route to the node if possible.
```bluespec "messages" +=
    | ResetMsg
```

## Routing

The protocol has a routing mechanism to filter `TxMsg` messages that nodes sent to their peers.

A route is a tuple `(source, target)`. We also write it as `source -> target`. Routes are defined
within a node, and `source` and `target` are peers connected to that node.
```bluespec "routing" +=
type Route = (NodeID, NodeID)
```

Routing is defined by a set of disabled routes on each node.
```bluespec "routing" +=
var dr: NodeID -> Set[Route]
```
By default, a node has all of its routes enabled, so the set of disabled routes is empty. When
disseminating some transaction `tx`, a node `A` will send `TxMsg(tx)` to peer `B` if the route
`sender(tx) -> B` is enabled, that is, the route is not in this set.

### Functions on routes

Disable the route `source -> target` by adding it to a set of disabled routes.
```bluespec "routing" +=
pure def disableRoute(_dr, node, source, target) =
    _dr.update(node, routes => routes.join((source, target)))
```

Enable all routes to peer or from peer by removing any disabled route that has peer as source or
target. 
```bluespec "routing" +=
pure def enableRoute(_dr, peer) = 
    _dr.filter(route => not(peer.isSourceOrTargetIn(route)))
```

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "routing" +=
def DisabledRoutes(node) = dr.get(node)
pure def isSourceOrTargetIn(node, route) = node == route._1 or node == route._2
```
</details>

## Redundancy Control

The Redundancy Controller (RC) is a closed-loop mechanism that auto-adjust the level of redundant
transactions that a node receives. As part of its initial configuration, a node needs to set to
parameters (1) the target redundancy that it aims to maintain and (2) how often it should make
adjustments. The controller periodically computes the redundancy level and tries to keep it within
certain pre-defined accepted bounds (as defined below).

Each node has an RC to adjust its redundancy level.
```bluespec "rc" +=
var rc: NodeID -> RedundancyController
```
where `RedundancyController` is a data structure with the following fields:

* A counter of transactions received for the first time by a node.
```bluespec "rcstate" +=
firstTimeTxs: int,
```

* A counter of duplicate transactions received by a node.
```bluespec "rcstate" +=
duplicateTxs: int,
```

* A flag stating whether the node is allowed to reply with a `HaveTx` message upon receiving a
  duplicate transaction.
```bluespec "rcstate" +=
isHaveTxBlocked: bool,
```

### Redundancy level

We define the _redundancy level_ as the proportion of duplicate transactions over first-time
transactions. For example, a redundancy of 1 means that for each transaction that the node has
received for the first time, the node has received one transaction that is a duplicate (not
necessarily a duplicate of the first-time transactions received).
```bluespec "rc" +=
pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs
```

`adjustRedundancy` returns an updated RC state and whether to reply with a `Reset` message.
```bluespec "rc" +=
pure def adjustRedundancy(_rc) = 
    if (_rc.redundancy() < redundancyLowerBound)
        (_rc.resetCounters(), true)
    else if (_rc.redundancy() >= redundancyUpperBound)
        (_rc.resetCounters().blockHaveTx(), false)
    else 
        (_rc.resetCounters(), false)
```
Periodically, the controller computes the redundancy level. If the level is too low, it will try to
increase it by sending to one of its peers a `Reset` message, thus increasing the number of
transactions received. If the level is too high, it will try to decrease it by allowing sending
`HaveTx` messages, so that the number of transactions received decline. Otherwise it does nothing.

On every adjustment, the counters are reset. This means that redundancy is computed for the period
elapsed since the last adjustment.

Note that when the target redundancy is 0, the lower and upper bounds are also equal to 0. Then
every `TxsPerAdjustment` received transactions `adjustRedundancy` will unblock `HaveTx` but it will
not send `Reset` messages.

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "rc" +=
def RC(node) = rc.get(node)
val initialRCState = { firstTimeTxs: 0, duplicateTxs: 0, isHaveTxBlocked: false }
pure def increaseFirstTimeTxs(_rc) = { firstTimeTxs: _rc.firstTimeTxs + 1, ..._rc }
pure def increaseDuplicateTxs(_rc) = { duplicateTxs: _rc.duplicateTxs + 1, ..._rc }
pure def resetCounters(_rc) = { firstTimeTxs: 0, duplicateTxs: 0, ..._rc }
pure def blockHaveTx(_rc) = { isHaveTxBlocked: true, ..._rc }
```
</details>

## Parameters

The protocol has two parameters that must be configured.

### Target redundancy

`TargetRedundancy` is the desired redundancy level of a node. A certain level that the Redundancy
Controller aims to maintain (within the boundaries defined by below).
```bluespec "params" +=
const TargetRedundancy: int
```
When `TargetRedundancy` is 0, the Redundancy Control mechanism is partially disabled:
`adjustRedundancy` will be able only to block HaveTx messages but not to send `Reset` messages. This
makes the protocol not resistant to malicious behaviour. Therefore, in practice, it is not
recommended to set it to 0 in Byzantine networks.

This value should be a real type, but reals are not currently supported by Quint.

#### Target bounds

Based on the target redundancy, the following constants define the accepted bounds of redundancy
level.
```bluespec "params" +=
val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
val redundancyLowerBound = TargetRedundancy - _delta
val redundancyUpperBound = TargetRedundancy + _delta
```
where:
```bluespec "params" +=
val TargetRedundancyDeltaPercent: int = 5
```
defines the tolerance of acceptable redundancy below and above `TargetRedundancy`, as a percentage
in the range `[0, 100)`.

### Number of transactions per adjustment

`TxsPerAdjustment` is the number of first-time transactions that the node must wait to receive
before it calls `adjustRedundancy`.
```bluespec "params" +=
const TxsPerAdjustment: int
```

`adjustNow` determines if it is time to adjust redundancy. It returns true when the controller
counts more than a certain number of first-time transactions. 
```bluespec "params" +=
pure def adjustNow(rc) =
    val threshold = TxsPerAdjustment / rc.redundancy()
    rc.firstTimeTxs >= max(MinTxsPerAdjustment, min(MaxTxsPerAdjustment, threshold))
```
where:
```bluespec "params" +=
val MinTxsPerAdjustment = 10
val MaxTxsPerAdjustment = 1000
```
are the minimum and maximum allowed values of threshold, which we use to keep it within a safe
range.

The `threshold` triggers adjustments is proportionally inverse to the redundancy level. When
redundancy is high, the node needs to decrease it fast, so adjustments happen more frequently. When
redundancy is low, adjustments will happen less often.


We don't want to trigger redundancy adjustments on intervals of a fixed number of transactions. If
the network is receiving a low transaction load, nodes may need to wait too much to adjust. While if
the load is high, adjustments will happen very frequently. When a node joins a network, redundancy
will be high and, ideally, we want the controller to make adjustments more often to reach the target
redundancy faster. This requirement is not strictly needed because we are interested in decreasing
the transaction bandwidth in the long run, and not necessarily as soon as when a node joins the
network.

## Initial state

DOG's initial state is based on Flood's initial state.
```bluespec "actions" +=
action DOG_init = all {
    Flood::init,
    dr' = NodeIDs.mapBy(_ => Set()),
    rc' = NodeIDs.mapBy(_ => initialRCState)
}
```

## State transitions (actions)

There are 5 possible state transitions in the protocol. In the rest of the section we describe the
missing details of each step.

1. User-initiated transactions: node receives a transaction from a user
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    nondet tx = oneOf(AllTxs)
    node.receiveTxFromUser(tx, tryAddTx),
    ```

2. Peer message handling: node processes messages received from peers
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    node.receiveFromPeer(handleMessage),
    ```

3. Transaction dissemination: node sends transaction to subset of peers
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    all {
        node.disseminateNextTx(mkTargetNodes, TxMsg),
        senders' = senders,
        dr' = dr,
        rc' = rc,
    },
    ```

4. Nodes joining the network (same as in Flood)
    ```bluespec "steps" +=
    all {
        pickNodeAndJoin,
        mempool' = mempool,
        senders' = senders,
        dr' = dr,
        rc' = rc,
    },
    ```

5. Nodes leaving the network
    ```bluespec "steps" +=
    all {
        // Pick a node that is not the only node in the network.
        require(size(nodesInNetwork) > 1),
        nondet nodeToDisconnect = oneOf(nodesInNetwork) 
        disconnectAndUpdateRoutes(nodeToDisconnect),
    }
    ```

### Adding transactions to the mempool

`tryAddTx` defines how a node add a transaction to its mempool. The logic is the same as in Flood,
the difference is in the two functions that process the transaction.
```bluespec "actions" +=
action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
    if (not(hash(tx).in(node.Cache())))
        node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
    else
        node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
```

#### Adding first-time transactions

`tryAddFirstTimeTx` attempts to add a received first-time transaction tx to the mempool: 
1. it adds `tx` to `cache`, 
2. if `tx` is valid, it adds `tx` to `pool`, and
3. if `tx` is valid, it updates `tx`'s senders. 

All these actions are taken from Flood. Additionally,

4. it increases `rc.firstTimeTxs`, and 
5. on every `TxsPerAdjustment` transactions received for the first time, call `adjustRedundancy()`.
```bluespec "actions" +=
action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = 
    val _rc1 = node.RC().increaseFirstTimeTxs()
    val _result = if (_rc1.adjustNow()) _rc1.adjustRedundancy() else (_rc1, false)
    val _rc2 = _result._1
    val sendReset = _result._2
    all {
        val updatedIncomingMsgs = 
            val targets = optionalSender.optionToSet() // may be empty
            if (sendReset) node.multiSend(_incomingMsgs, targets, ResetMsg) else _incomingMsgs
        node.Flood::tryAddFirstTimeTx(updatedIncomingMsgs, optionalSender, tx),
        rc' = rc.put(node, _rc2),
        dr' = dr,
    }
```

#### Handling duplicate transactions

`processDuplicateTx` processes a received duplicate transaction `tx`: 
1. it increases `duplicateTxs` and 
2. replies a `HaveTx` message if the RC mechanism is not blocking it (and there's a sender). 

As in Flood, it updates the list of senders if `tx` is in `pool` (and thus it's valid), and update
the list of incoming messages. 
```bluespec "actions" +=
action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) =
    // Reply `HaveTxMsg` if `tx` comes from a peer.
    val updatedIncomingMsgs = 
        val targets = optionalSender.optionToSet() // may be empty
        if (not(node.RC().isHaveTxBlocked))
            node.multiSend(_incomingMsgs, targets, HaveTxMsg(hash(tx)))
        else _incomingMsgs
    all {
        node.Flood::processDuplicateTx(updatedIncomingMsgs, optionalSender, tx),
        rc' = rc.update(node, increaseDuplicateTxs),
        dr' = dr,
    }
```

### Handling incoming messages

`handleMessage` defines how to handle each type of message received from a peer (the `sender`).
```bluespec "actions" +=
action handleMessage(node, _incomingMsgs, sender, msg) =
    match msg {
    | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
    | HaveTxMsg(txID) => node.handleHaveTxMessage(_incomingMsgs, sender, txID)
    | ResetMsg => node.handleResetMessage(_incomingMsgs, sender)
    }
```

#### Handling HaveTx messages

Upon receiving `HaveTxMsg(txID)`, disable the route `sender(txID) -> sender`, if `node` has at least
a sender for `txID`. This will decrease the traffic to `sender`.
```bluespec "actions" +=
action handleHaveTxMessage(node, _incomingMsgs, sender, txID) = all {
    dr' = 
        val txSenders = node.Senders().mapGetDefault(txID, List())
        if (length(txSenders) > 0)
            dr.disableRoute(node, txSenders[0], sender)
        else dr,
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
    mempool' = mempool,
    senders' = senders,
    rc' = rc,
}
```

We don't want to cut all routes from `sender` to `node`, only the route that has as source one the
peer IDs in `txSenders`, which may be many. We need to choose one, so we pick the first in the list,
which is the first peer that sent `tx` to `node`. Other peers in the list also sent `tx` but at a
later time, meaning that routes coming from those nodes are probably disabled, and most of the
traffic comes from the first one.

#### Handling Reset messages

Upon receiving `ResetMsg`, remove any route that has `sender` as source or target. 
```bluespec "actions" +=
action handleResetMessage(node, _incomingMsgs, sender) = all {
    dr' = dr.update(node, drs => drs.enableRoute(sender)),
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
    mempool' = mempool,
    senders' = senders,
    rc' = rc,
}
```
This will allow traffic to flow again to the sender, and other nodes will dynamically adapt to the
new traffic, closing routes when needed.

### Transaction dissemination

As in Flood, DOG will filter out the transaction's senders. Additionally, DOG will not send `tx` to
a peer if the route `sender(tx) -> peer` is disabled.
```bluespec "actions" +=
def mkTargetNodes(node, tx) =
    val txSenders = node.Senders().mapGetDefault(hash(tx), List())
    val disabledTargets = node.DisabledRoutes()
        // Keep only routes whose source is one of tx's senders.
        .filter(r => r._1.in(txSenders.listToSet()))
        // Keep routes' targets.
        .map(r => r._2)
    node.Peers().exclude(txSenders.listToSet()).exclude(disabledTargets)
```

### Nodes disconnecting from the network

When a node disconnects from the network, its peers signal their own peers that their situation has
changed, so that their routing tables are reset. In this way, data via those nodes can be re-routed
through other nodes if needed.
```bluespec "actions" +=
action disconnectAndUpdateRoutes(nodeToDisconnect) = all {
    // All node's peers detect that node has disconnect and send a Reset
    // message to all their peers.
    val updatedIncomingMsgs = nodeToDisconnect.Peers().fold(incomingMsgs, 
        (_incomingMsgs, peer) => peer.multiSend(_incomingMsgs, peer.Peers(), ResetMsg)
    )
    nodeToDisconnect.disconnectNetwork(updatedIncomingMsgs),
    // The node's peers enable all routes to node in their routing tables.
    dr' = dr.updateMultiple(nodeToDisconnect.Peers(), 
        drs => drs.enableRoute(nodeToDisconnect)),
    mempool' = mempool,
    senders' = senders,
    rc' = rc,
}
```

<!--
```bluespec quint/dog.qnt +=
// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module dog {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    import flood as Flood from "./flood"
    import flood.senders from "./flood"
    import flood.Senders from "./flood"

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    <<<messages>>>

    //--------------------------------------------------------------------------
    // Routing
    //--------------------------------------------------------------------------
    <<<routing>>>

    //--------------------------------------------------------------------------
    // Redundancy Controller
    //--------------------------------------------------------------------------
    type RedundancyController = {
        <<<rcstate>>>
    }
    <<<rc>>>

    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    <<<params>>>

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    <<<actions>>>

    action step = any {
        <<<steps>>>
    }

}
```
-->

[quint]: https://quint-lang.org/
