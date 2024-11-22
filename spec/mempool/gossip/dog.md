# Dynamic Optimal Graph (DOG) gossip protocol

The DOG protocol introduces two novel features to optimize network bandwidth utilization while
ensuring robustness against Byzantine attacks and preserving low latency performance.

* **Dynamic Routing.** DOG implements a routing mechanism that filters data disseminated from a node
  to its peers. A node `A` receiving from `B` a transaction that is already present in its cache
  implies that there is a cycle in the network topology. Node `A` will message `B` indicating to
  stop sending transactions and `B` will close one of the "routes" that sends transactions to `A`,
  thus cutting the cycle. Eventually, transactions will have only one path to reach all nodes in the
  network, with the routes forming a superposition of spanning trees--the optimal connection
  structure for disseminating data across the network.

* **Redundancy Control mechanism.** For keeping nodes resilient to Byzantine attacks, the protocol
  maintains a minimum level of redundant transactions. If a node is not receiving enough duplicate
  transactions, it will request its peers to re-activate a previously disabled route.

The DOG protocol is built on top of the [Flood protocol](flood.md). This spec re-uses many of the
same types, messages, and data structures defined in Flood's spec. 

**Table of contents**
  - [Messages](#messages)
  - [Dynamic Routing](#dynamic-routing)
  - [Redundancy Control](#redundancy-control)
    - [Computing redundancy](#computing-redundancy)
    - [How to adjust](#how-to-adjust)
    - [When to adjust](#when-to-adjust)
  - [Parameters](#parameters)
  - [Initial state](#initial-state)
  - [State transitions (actions)](#state-transitions-actions)
    - [Adding transactions to the mempool](#adding-transactions-to-the-mempool)
    - [Handling incoming messages](#handling-incoming-messages)
    - [Transaction dissemination](#transaction-dissemination)
    - [Nodes disconnecting from the network](#nodes-disconnecting-from-the-network)

> This document was written using the literature programming paradigm. Code snippets are written in
> [Quint][quint] and can get "tangled" into a Quint file.

## Messages

In addition to the `TxMsg` data message present in Flood, DOG adds two control messages. 
```bluespec "messages" +=
type Message =
```

* Transaction message (same as in Flood).
    ```bluespec "messages" +=
        | TxMsg(TX)
    ```

* A node sends `HaveTxMsg` messages to signal that it already received the transaction. The receiver
  will cut a route related to `tx` that is forming a cycle in the network topology.
    ```bluespec "messages" +=
        | HaveTxMsg(TxID)
    ```

* A node sends `ResetMsg` messages to signal that it is not receiving enough transactions. The
  receiver should re-enable some route to the node if possible.
    ```bluespec "messages" +=
        | ResetMsg
    ```

Note that the size of `HaveTxMsg` and `ResetMsg` is negligible compared to `TxMsg`, which carries a
full transaction.

## Dynamic Routing

The protocol has a routing mechanism to filter transaction messages that nodes sent to their peers.

A _route_ is a tuple `(source, target)` representing the flow of transactions between two nodes. A
route is defined within a node, and `source` and `target` are peers connected to that node. 
```bluespec "routing" +=
type Route = (NodeID, NodeID)
```
We also write a route as `source -> target`.

Each node maintains a set of disabled routes (`dr`) to manage active connections.
```bluespec "routing" +=
var dr: NodeID -> Set[Route]
```
By default, all routes are enabled, that is, the set of disabled routes is empty. A node `B` will
send a transaction `tx` to peer `C` only if the route `A -> C` is not in `B`'s set of disabled
routes, where `A` is the first node in `tx`'s list of senders (see the section on [Transaction
dissemination](#transaction-dissemination)).

We define the following functions on routes:

* `disableRoute` adds the route `source -> target` to a set of disabled routes.
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

The Redundancy Controller (RC) implements a closed-loop mechanism that dynamically adjusts the level
of redundant transactions received by a node.

Each node has a _controller_ that periodically measures the redundancy level (defined below) and
tries to keep it within pre-configured bounds by sending `HaveTx` and `Reset` messages to peers.
```bluespec "rc" +=
var rc: NodeID -> RedundancyController
```

The data structure `RedundancyController` contains the following fields:
* A counter of transactions received for the first time by the node.
    ```bluespec "rcstate" +=
    firstTimeTxs: int,
    ```
* A counter of duplicate transactions received by the node.
    ```bluespec "rcstate" +=
    duplicateTxs: int,
    ```
* A flag indicating whether the node is temporarily blocked from replying with a `HaveTx` message
  upon receiving a duplicate transaction.
    ```bluespec "rcstate" +=
    isHaveTxBlocked: bool,
    ```

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

### Computing redundancy

The _redundancy level_ is defined as the proportion of duplicate transactions relative to first-time
transactions.
```bluespec "rc" +=
pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs
```
For example, a redundancy of 0.5 means that, for every two first-time transactions, the node
receives one duplicate transaction (not necessarily a duplicate of any of those two first-time
transactions).

### How to adjust

Function `adjustRedundancy` computes the current `redundancy` level and determines the controller
actions by returning an updated RC state and whether the controller should send a `Reset` message:
- If no transactions were received during the last iteration, the controller should not react in
  order to preserve the current state of the routes.
- If `redundancy` is too low, the controller should request more transactions by sending a `Reset`
  message to a random peer.
- If `redundancy` is too high, the controller should signal peers to reduce traffic by temporarily
  allowing to reply with a `HaveTx` message the next time the node receives a duplicate transaction.
- If `redundancy` is within acceptable limits, the controller takes no action.
```bluespec "rc" +=
pure def adjustRedundancy(node, _rc) =
    if (_rc.firstTimeTxs + _rc.duplicateTxs == 0)
        (_rc, false)
    else if (_rc.redundancy() < lowerBound)
        (_rc, true)
    else if (_rc.redundancy() >= upperBound)
        (_rc.blockHaveTx(), false)
    else 
        (_rc, false)
```
Note that if the target redundancy is 0 (see the [parameters](#parameters)), the lower and upper
bounds are also equal to 0. Then `adjustRedundancy` will be able to unblock `HaveTx` but it will
never send `Reset` messages

### When to adjust

The Redundancy Controller runs in a separate thread a control loop that periodically calls
`adjustRedundancy` to maintain the target redundancy level. 

Action `controlLoopIteration` models the behavior of one iteration, which computes the current
redundancy level and then determines necessary actions such as sending a `Reset` message to a
randomly chosen peer. Additionally, after each adjustment, it resets the transaction counters so
that redundancy is computed independently of past measurements.
```bluespec "actions" +=
action controlLoopIteration(node) = all {
    val res = node.adjustRedundancy(node.RC())
    val updatedNodeRC = res._1
    val sendReset = res._2
    nondet randomPeer = oneOf(node.Peers())
    all {
        incomingMsgs' = 
            if (sendReset) 
                node.multiSend(incomingMsgs, Set(randomPeer), ResetMsg)
            else incomingMsgs,
        rc' = rc.put(node, updatedNodeRC.resetCounters()),
        peers' = peers,
        mempool' = mempool,
        senders' = senders,
        dr' = dr,
    }
}
```

Adjustments should be paced by a timer to account for message propagation delays. If redundancy
adjustments are made too frequently, a node risks isolation as all peers may cut routes prematurely.
The timer should align with the network’s maximum round-trip time (RTT) to allow `HaveTx` and
`Reset` messages to propagate and take effect before initiating further adjustments.

> An alternative to triggering `adjustRedundancy` at fixed time intervals is to base it on the
> number of received transactions. While this approach eliminates dependency on time constraints, it
> introduces significant security risks. An attacker could exploit this mechanism by sending bursts
> of numerous small transactions, causing the node to trigger `adjustRedundancy` too frequently.
> This might result in the near-continuous activation of `HaveTx` messages, leading peers to
> systematically cut routes to the node, ultimately isolating the node from transaction traffic.

> The `controlLoopIteration` action should have a pre-condition that enables to fire the state
> transition only on specified time intervals, such as the parameter `adjustInterval` defined below.
> However, Quint does not allow to (easily) specify time constraints. Since we are modeling the
> protocol as a state transition system, we just model `controlLoopIteration` as always being
> enabled. A real implementation will naturally use a timer.

## Parameters

The protocol has the following parameters that each node must configure during initialization: 
1) the desired redundancy level that the controller should keep as target,
2) a percentage of the target value that defines the accepted lower and upper bounds,
3) the time interval between redundancy adjustments.

* `TargetRedundancy` specifies the level of redundancy the controller aims to maintain (within
  specified bounds).
    ```bluespec "params" +=
    const TargetRedundancy: int
    ```
    A target equal to 0 partially disables the Redundancy Control mechanism: the controller can
    block `HaveTx` messages but cannot send `Reset` messages. Zero redundancy minimizes bandwidth
    usage, achieving the lowest possible message overhead. In non-Byzantines networks, this is the
    best possible scenario. However, in Byzantine networks it could potentially render nodes
    isolated from transaction data. Therefore, the target should be set to a value greater than 0.
    Experimental results suggest a value between 0.5 and 1, which is a safe number that does not
    result in excessive duplicate transactions.

    > Note: `TargetRedundancy` should ideally be a real number, but reals are not currently
    > supported by Quint.

* `TargetRedundancyDeltaPercent` is a percentage (a number in the range `[0, 100)`) that defines the
  acceptable bounds for redundancy levels as a deviation from `TargetRedundancy`.
    ```bluespec "params" +=
    const TargetRedundancyDeltaPercent: int
    ```
    From this value the protocol derives the constants:
    ```bluespec "params" +=
    val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
    val lowerBound = TargetRedundancy - _delta
    val upperBound = TargetRedundancy + _delta
    ```
    This range provides flexibility, allowing the redundancy level to fluctuate while staying within
	tolerable limits. For example, a `TargetRedundancy` of 0.5 with a `TargetRedundancyDeltaPercent`
	of 20% would allow a redundancy level between 0.4 and 0.6.
    
    Based on experimentation, a `TargetRedundancyDeltaPercent` of around 20% strikes a good balance
	between adaptability and stability.

	> Note: Similar to `TargetRedundancy`, this parameter could also be defined as a real type if it
	were allowed in Quint.

* `adjustInterval` defines the time (in milliseconds) that the controller waits between successive
  calls of `adjustRedundancy`.
    ```bluespec "params" +=
    const adjustInterval: int
    ```
	This interval should allow sufficient time for control messages (`HaveTx` and `Reset`) to
	propagate through the network and take effect.
	
    A minimum value for `adjustInterval` depends on the network’s round-trip time (RTT). Assuming that
    global latency typically stays below 500ms (maybe an excessive number, need references here), the
    interval should be set to at least 1000ms to ensure stability and avoid over-adjustment.

    Optimal values depend on empirical measurements of network latency. However, in practice, values
    above 1000ms are recommended to allow message processing and delivery times in diverse network
    environments. For instance, a node with 50 peers will have a maximum of `50 * (50 - 1) = 2450`
    routes. Hypothetically, removing one route on every adjustment, at one adjustment per second, it
    would take 40.8 minutes to remove all routes from the node.

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

The following are all possible state transitions allowed in the protocol. The rest of the section
describes the missing details of each step.

1. A node receives a transaction from a user and tries to add it to its mempool.
    ```bluespec "steps" +=
    // User-initiated transactions
    nondet node = oneOf(nodesInNetwork)
    nondet tx = oneOf(AllTxs)
    node.receiveTxFromUser(tx, tryAddTx),
    ```

2. A node handles a message received from a peer.
    ```bluespec "steps" +=
    // Peer message handling
    nondet node = oneOf(nodesInNetwork)
    node.receiveFromPeer(handleMessage),
    ```

3. A node disseminates a transaction in its mempool to a subset of peers.
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

3. A node joins the network (same as in Flood).
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

4. A nodes disconnects from the network: its peers must update their routes.
    ```bluespec "steps" +=
    // Node disconnects from network
    all {
        // Pick a node that is not the only node in the network.
        require(size(nodesInNetwork) > 1),
        nondet nodeToDisconnect = oneOf(nodesInNetwork) 
        disconnectAndUpdateRoutes(nodeToDisconnect),
    },
    ```

5. The Redundancy Controller periodically tries to adjust the redundancy level.
    ```bluespec "steps" +=
    // Redundancy Controller process loop
    nondet node = oneOf(nodesInNetwork)
    node.controlLoopIteration(),
    ```

### Adding transactions to the mempool

`tryAddTx` defines how a node adds a transaction to its mempool. The following code is the same as
in Flood; the difference is in the two functions that process the transaction.
```bluespec "actions" +=
action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
    if (not(hash(tx).in(node.Cache())))
        node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
    else
        node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
```

* **Adding first-time transactions**

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

* **Handling duplicate transactions**

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
    Immediately after sending the `HaveTx` message the controller will block sending a new one.

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

* **Handling `HaveTx` messages**

    Upon receiving `HaveTxMsg(txID)` from `sender`, `node` disables the route `txSender -> sender`,
    where `txSender` is the first node in `txID`'s list of senders, if `txID` actually comes from a
    peer. This action will decrease the traffic to `sender`.
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
    The list of `tx`’s senders contains the node IDs from which `node` received the transaction,
    ordered by the arrival time of the corresponding messages. To avoid disabling the routes from
    all those senders at once, the protocol picks the first sender in the list, which is the first
    peer from which `node` received `tx`. Subsequent entries in the list are nodes whose transaction
    messages arrived later as duplicates. As such, most routes from those peers to `node` will
    eventually be disabled, with most traffic coming primarily from the first peer.

* **Handling Reset messages**

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

As in Flood, DOG will filter out the transaction's senders. Additionally, `node` will not send `tx`
to peer `B` if the route `A -> B` is disabled, where `A` is the first node in the list of `tx`'s
senders.
```bluespec "actions" +=
def mkTargetNodes(node, tx) =
    val txSenders = node.sendersOf(hash(tx))
    val disabledTargets = node.DisabledRoutes()
        // Keep routes whose source is tx's first sender, if any.
        .filter(r => if (length(txSenders) > 0) r._1 == txSenders[0] else false)
        // Map routes to targets.
        .map(r => r._2)
    node.Peers().exclude(txSenders.listToSet()).exclude(disabledTargets)
```
The protocol selects the first sender in the list based on the same reasoning applied when handling
received `HaveTx` messages. This sender is likely responsible for the majority of traffic related to
the transaction, as it was the first to forward it to the node. Subsequent senders in the list only
sent the transaction later as duplicates, and their routes are more likely already disabled.

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
