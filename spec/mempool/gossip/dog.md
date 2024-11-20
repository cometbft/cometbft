# Dynamic Optimal Graph (DOG) gossip protocol

The DOG protocol optimizes network efficiency by dynamically managing how transactions are routed to
peers: if node A receives a transaction from node B that it already has in its cache, it implies
that there is a cycle in the network topology. Node A will message B to stop sending transactions,
and B will close one of the "routes" that sends transactions to A, thus cutting the cycle.

Additionally, for keeping nodes resilient to Byzantine attacks, the protocol has a Redundancy
Control mechanism that maintains a minimum, pre-defined level of transaction redundancy. If a node
is not receiving enough duplicate transactions, it will message its peers to request additional
ones.

The DOG protocol is built on top of the [Flood protocol](flood.md). DOG's spec uses many of the same
types, messages, and data structures as Flood. In principle, it is possible to enable or disable DOG
in nodes running Flood. Nodes running DOG and Flood can co-exist in the network, though performance
will be optimal only if all nodes enable DOG.

**Table of contents**
  - [Messages](#messages)
  - [Routing](#routing)
  - [Redundancy Control](#redundancy-control)
    - [Redundancy level](#redundancy-level)
    - [How to adjust redundancy](#how-to-adjust-redundancy)
    - [When to adjust redundancy](#when-to-adjust-redundancy)
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

In addition to the `TxMsg` data message present in Flood, DOG adds two control messages. The size of
the control messages is negligible compared to `TxMsg`, which carries a full transaction.
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

The protocol has a routing mechanism to filter transaction messages that nodes sent to their peers.

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

We define the following functions on routes:

* `disableRoute` disables the route `source -> target` by adding it to a set of disabled routes.
    ```bluespec "routing" +=
    pure def disableRoute(_dr, node, source, target) =
        _dr.update(node, routes => routes.join((source, target)))
    ```

* `enableRoute` enables all routes to `peer` or from `peer` by removing any disabled route that has
`peer` as source or target. 
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

The Redundancy Controller (RC) is a closed-loop mechanism that dynamically adjusts the level of
redundant transactions that a node receives. 

Each node has a controller that periodically measures the redundancy
level and tries to keep it within pre-configured accepted bounds by sending control messages
(`HaveTx` and `Reset`) to its peers.
```bluespec "rc" +=
var rc: NodeID -> RedundancyController
```

The data structure `RedundancyController` has the following fields:
* A counter of transactions received for the first time by the node.
    ```bluespec "rcstate" +=
    firstTimeTxs: int,
    ```
* A counter of duplicate transactions received by the node.
    ```bluespec "rcstate" +=
    duplicateTxs: int,
    ```
* A flag indicating whether the node is allowed to reply with a `HaveTx` message upon receiving a
  duplicate transaction.
    ```bluespec "rcstate" +=
    isHaveTxBlocked: bool,
    ```

As part of its initial configuration, each node needs to set two parameters: 
1) the desired redundancy that the controller should keep as target and 
2) how often it should make adjustments.

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

### Redundancy level

We define the _redundancy level_ as the proportion of duplicate transactions over first-time
transactions (as defined [here](flood.md#adding-transactions-to-the-mempool)). 
```bluespec "rc" +=
pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs
```
For example, a redundancy of 1 means that for each transaction that the node has received for the
first time, the node has received one transaction that is a duplicate (not necessarily a duplicate
of the first-time transactions received).

### How to adjust redundancy

Function `adjustRedundancy` computes the current `redundancy` level and returns an updated RC state
and updated list of incoming messages.
- If in the iteration there were no received transactions (first-time + duplicates transactions is
  0), the controller should not do anything. In particular it should not send `Reset` messages
  because we want to keep the set of disabled routes stable.
- If `redundancy` is less than the lower acceptable limit, the redundancy level should be increased,
  so it sends a `Reset` message to a random peer,
- If `redundancy` is higher thant the upper acceptable limit, the redundancy level should be
  decreased, then it will unblock in RC to allow sending `HaveTx` messages the next time the node
  receives a duplicate transaction.
- If the redundancy level is within acceptable limits, nothing should change.
```bluespec "rc" +=
pure def adjustRedundancy(node, _incomingMsgs, _rc, randomPeerToSendReset) =
    val red = _rc.redundancy()
    if (_rc.firstTimeTxs + _rc.duplicateTxs == 0)
        (_rc, _incomingMsgs)
    else if (red < lowerBound)
        (_rc.resetCounters(), node.multiSend(_incomingMsgs, Set(randomPeerToSendReset), ResetMsg))
    else if (red >= upperBound)
        (_rc.resetCounters().blockHaveTx(), _incomingMsgs)
    else 
        (_rc.resetCounters(), _incomingMsgs)
```
Every call to `adjustRedundancy` will reset the counters, so that the redundancy level in each
iteration is computed independently of past measurements.

Note that when the target redundancy is 0, the lower and upper bounds are also equal to 0. In this
case, `adjustRedundancy` will unblock `HaveTx` but it will not send `Reset` messages.

### When to adjust redundancy

The Redundancy Controller runs `redundancyControlLoop` in a separate process calling
`adjustRedundancy` on the pre-defined `adjustInterval` duration (see below).

Sending too many `HaveTx` messages immediately after receiving a duplicate transaction can cut all
incoming routes to the node, rendering it isolated from transaction traffic. The solution is to wait
a certain time after sending a `HaveTx` message, to allow peers to process the message, and allow
traffic to take other routes. After that moment, it is safe to send `HaveTx` messages again when
receiving new duplicate transactions.

<details>
  <summary>Control loop</summary>

In Quint we cannot specify that an action is enabled on time intervals. 
```bluespec "actions" +=
action redundancyControlLoop(node) = all {
    require(node.RC().adjustNow()),
    nondet randomPeerToSendReset = oneOf(node.Peers())
    val res = node.adjustRedundancy(incomingMsgs, node.RC(), randomPeerToSendReset)
    all {
        incomingMsgs' = res._2,
        peers' = peers,
        mempool' = mempool,
        senders' = senders,
        rc' = rc.put(node, res._1),
        dr' = dr,
    }
}
```
</details>


## Parameters

The protocol has two parameters that each node must configure: the target redundancy level, and the
duration of adjustment intervals.

### Target redundancy

`TargetRedundancy` is the desired redundancy level of a node. A certain level that the Redundancy
Controller aims to maintain (within the boundaries defined by below).
```bluespec "params" +=
const TargetRedundancy: int
```
When `TargetRedundancy` is 0, the Redundancy Control mechanism is partially disabled:
`adjustRedundancy` will be able only to block `HaveTx` messages but not to send `Reset` messages.
This makes the protocol not resistant to malicious behaviour. Therefore, in practice, it is not
recommended to set it to 0 in Byzantine networks.

This value should be a real type, but reals are not currently supported by Quint.

#### Target bounds

Based on the target redundancy, the following constants define the accepted bounds of redundancy
level.
```bluespec "params" +=
val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
val lowerBound = TargetRedundancy - _delta
val upperBound = TargetRedundancy + _delta
```
where:
```bluespec "params" +=
val TargetRedundancyDeltaPercent: int = 5
```
defines the tolerance of acceptable redundancy below and above `TargetRedundancy`, as a percentage
in the range `[0, 100)`.

### Time interval of adjustments

`adjustInterval` is the duration in milliseconds that the controller must wait before the next call to `adjustRedundancy`.
```bluespec "params" +=
const adjustInterval: int
```
The suggested default value is 1000 milliseconds or higher.


Quint does not allow to specify time constraints.

`adjustNow` determines if it is time to adjust redundancy.  
```bluespec "params" +=
pure def adjustNow(rc) = true
```

## Initial state

DOG's initial state is based on Flood's initial state.
```bluespec "init_action" +=
action DOG_init = all {
    Flood::init,
    dr' = NodeIDs.mapBy(_ => Set()),
    rc' = NodeIDs.mapBy(_ => initialRCState)
}
```

## State transitions (actions)

There are 6 possible state transitions in the protocol. In the rest of the section we describe the
missing details of each step.

1. User-initiated transactions: node receives a transaction from a user
    ```bluespec "steps" +=
    // User-initiated transactions
    nondet node = oneOf(nodesInNetwork)
    nondet tx = oneOf(AllTxs)
    node.receiveTxFromUser(tx, tryAddTx),
    ```

2. Peer message handling: node processes messages received from peers
    ```bluespec "steps" +=
    // Peer message handling
    nondet node = oneOf(nodesInNetwork)
    node.receiveFromPeer(handleMessage),
    ```

3. Transaction dissemination: node sends transaction to subset of peers
    ```bluespec "steps" +=
    // Transaction dissemination
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
    // Node joins network
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
    // Node disconnects from network
    all {
        // Pick a node that is not the only node in the network.
        require(size(nodesInNetwork) > 1),
        nondet nodeToDisconnect = oneOf(nodesInNetwork) 
        disconnectAndUpdateRoutes(nodeToDisconnect),
    },
    ```

6. Redundancy Controller process
    ```bluespec "steps" +=
    // Redundancy Controller process loop
    nondet node = oneOf(nodesInNetwork)
    node.redundancyControlLoop(),
    ```

### Adding transactions to the mempool

`tryAddTx` defines how a node add a transaction to its mempool. The following code is the same as in
Flood, the difference is in the two functions that process the transaction.
```bluespec "actions" +=
action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
    if (not(hash(tx).in(node.Cache())))
        node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
    else
        node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
```

#### Adding first-time transactions

`tryAddFirstTimeTx` attempts to add a received first-time transaction `tx` to the mempool by
performing the same updates as in Flood (it adds `tx` to `cache`; if `tx` is valid, it appends `tx`
to the mempool and updates `tx`'s senders). Additionally, it increases the `node`'s
`rc.firstTimeTxs` counter.
```bluespec "actions" +=
action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = 
    all {
        node.Flood::tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx),
        rc' = rc.update(node, increaseFirstTimeTxs),
        dr' = dr,
    }
```

#### Handling duplicate transactions

`processDuplicateTx` processes a received duplicate transaction `tx` by updating the list of senders
if `tx` is in the mempool, and the list of incoming messages, the same as in Flood. Additionally,
1. it increases `duplicateTxs` and 
2. replies a `HaveTx` message if the RC mechanism is not blocking it (and there's a sender). 

```bluespec "actions" +=
action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) =
    val _rc = node.RC().increaseDuplicateTxs()
    val updatedVars = node.replyHaveTx(_incomingMsgs, _rc, optionalSender, tx)
    val _incomingMsgs1 = updatedVars._1
    val _rc1 = updatedVars._2
    all {
        node.Flood::processDuplicateTx(_incomingMsgs1, optionalSender, tx),
        rc' = rc.put(node, _rc1),
        dr' = dr,
    }
```
where `replyHaveTx` will send a `HaveTx` message if `tx` comes from a peer and `HaveTx` messages are
not blocked:
```bluespec "actions" +=
pure def replyHaveTx(node, _incomingMsgs, _rc, optionalSender, tx) =
    if (optionalSender.isSome() and not(_rc.isHaveTxBlocked))
        val targets = optionalSender.optionToSet()
        (node.multiSend(_incomingMsgs, targets, HaveTxMsg(hash(tx))), _rc.blockHaveTx())
    else (_incomingMsgs, _rc)
```

### Handling incoming messages

In this subsection we define how to handle each type of message received from a peer (the `sender`).
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
        val txSenders = node.sendersOf(txID)
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
    val txSenders = node.sendersOf(hash(tx)).listToSet()
    val disabledTargets = node.DisabledRoutes()
        // Keep only routes whose source is one of tx's senders.
        .filter(r => r._1.in(txSenders))
        // Keep routes' targets.
        .map(r => r._2)
    node.Peers().exclude(txSenders).exclude(disabledTargets)
```

### Nodes disconnecting from the network

When a node disconnects from the network, its peers signal their own peers that their situation has
changed, so that their routing tables are reset. In this way, if needed, data via those nodes can be
re-routed through other nodes.
```bluespec "actions" +=
action disconnectAndUpdateRoutes(nodeToDisconnect) = all {
    // All node's peers send a Reset message to all their peers.
    val updatedIncomingMsgs = nodeToDisconnect.Peers().fold(incomingMsgs, 
        (inMsgs, peer) => peer.multiSend(inMsgs, peer.Peers(), ResetMsg))
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
    import flood.sendersOf from "./flood"

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------
    <<<messages>>>

    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    <<<params>>>

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
    // Actions
    //--------------------------------------------------------------------------
    <<<init_action>>>

    <<<actions>>>

    action step = any {
        <<<steps>>>
    }

}
```
-->

[quint]: https://quint-lang.org/
