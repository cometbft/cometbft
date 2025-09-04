# Dynamic Optimal Graph (DOG) gossip protocol

The DOG protocol introduces two novel features to optimize network bandwidth utilization while
preserving low latency performance and ensuring robustness against Byzantine attacks.

* **Dynamic Routing.** DOG implements a routing mechanism that filters data disseminated from a node
  to its peers. When a node `A` receives from node `B` a transaction that is already present in its
  cache, this means that there is a cycle in the network topology. In this case, `A` will message
  `B` indicating to not send any more transactions, and `B` will close one of the "routes" that
  are used to forward transactions to `A`, thus cutting the cycle. Eventually, transactions will have only one
  path to reach all nodes in the network, with the resulting routes forming a superposition of
  spanning trees--the optimal P2P connection structure for disseminating data across the network.

* **Redundancy Control mechanism.** For keeping nodes resilient to Byzantine attacks, the protocol
  maintains a minimum level of transaction redundancy. Nodes periodically measure the redundancy
  level of received transactions and decide if they should request peers for more or less
  transactions. If a node is not receiving enough duplicate transactions, it will request its peers
  to re-activate a previously disabled route. This ensures a steady yet controlled flow of data.

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

* A node sends a `HaveTxMsg` message to signal that it already received a transaction. The receiver
  will cut a route related to `tx` that is forming a cycle in the network topology.
    ```bluespec "messages" +=
        | HaveTxMsg(TxID)
    ```

* A node sends a `ResetRouteMsg` message to signal that it is not receiving enough transactions. The
  receiver should, if possible, re-enable some route to the node.
    ```bluespec "messages" +=
        | ResetRouteMsg
    ```

Note that the size of `HaveTxMsg` and `ResetRouteMsg` is negligible compared to `TxMsg`, which
carries a full transaction.

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
send a transaction `tx`, received from peer `A`, to peer `C` only if the route `A -> C` is not in `B`'s set of disabled
routes. Since a transaction can be received from multiple peers, we define its sender `A` as the first node in `tx`'s list of senders (see the section on [Transaction
dissemination](#transaction-dissemination)).

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "routing" +=
def DisabledRoutes(node) = dr.get(node)
pure def disableRoute(routes, route) = routes.join(route)
pure def enableRoute(routes, route) = routes.exclude(Set(route))
pure def isSourceOrTargetIn(node, route) = node == route._1 or node == route._2
pure def routesWithSource(routes, source) = routes.filter(r => r._1 == source)
pure def routesWithTarget(routes, target) = routes.filter(r => r._2 == target)
pure def mapTargets(routes) = routes.map(r => r._2)
```

`resetRoutes` re-enables all routes to `peer` or from `peer` by removing any disabled route that has
`peer` as source or target. 
```bluespec "routing" +=
pure def resetRoutes(routes, peer) = 
    routes.filter(route => not(peer.isSourceOrTargetIn(route)))
```
</details>

## Redundancy Control

Each node implements a Redundancy Controller (RC) with a closed-loop feedback mechanism, commonly
used in control systems for dynamic self-regulation. The controller periodically monitors the level
of redundant transactions received and adjusts accordingly by sending `HaveTx` and `ResetRoute`
messages to peers. This ensures the redundancy level remains within predefined bounds, adapting to
changes in network conditions in real time.
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
pure def unblockHaveTx(_rc) = { isHaveTxBlocked: false, ..._rc }
pure def blockHaveTx(_rc) = { isHaveTxBlocked: true, ..._rc }
```
</details>

### Computing redundancy

The _redundancy level_ is calculated as the ratio of duplicate transactions to first-time
transactions.
```bluespec "rc" +=
pure def redundancy(_rc) =
    if (_rc.firstTimeTxs == 0) 
        upperBound 
    else 
        _rc.duplicateTxs / _rc.firstTimeTxs
```
If the number of first-time transactions is 0, the redundancy level is set to a predefined maximum
value (the constant `upperBound` defined below) to prompt the controller to reduce redundancy.
Conversely, the redundancy level is set to 0 if there are no duplicate transactions, signaling the
controller to increase redundancy. 

For example, a redundancy of 0.5 means that, for every two first-time transactions received, the
node receives one duplicate transaction (not necessarily a duplicate of any of those two first-time
transactions).

### How to adjust

Function `controllerActions` computes the current `redundancy` level and determines which actions
the controller will take by returning an updated controller state and whether the controller should
send a `ResetRoute` message:
- If no transactions were received during the last iteration, the controller should not react in
  order to preserve the current state of the routes.
- If `redundancy` is too low, the controller should request more transactions by sending a
  `ResetRoute` message to a random peer.
- If `redundancy` is too high, the controller should signal peers to reduce traffic by temporarily
  allowing to reply with a `HaveTx` message the next time the node receives a duplicate transaction.
- If `redundancy` is within acceptable limits, the controller takes no action.
```bluespec "rc" +=
pure def controllerActions(_rc) =
    if (_rc.firstTimeTxs + _rc.duplicateTxs == 0)
        (_rc, false)
    else if (_rc.redundancy() < lowerBound)
        (_rc, true)
    else if (_rc.redundancy() >= upperBound)
        (_rc.unblockHaveTx(), false)
    else 
        (_rc, false)
```
Note that if the target redundancy is 0 (see [parameters](#parameters)), the lower and upper bounds
are also equal to 0. Then `controllerActions` will be able to unblock `HaveTx` but it will never
send `ResetRoute` messages.

An important aspect of the controller actions is that, on each iteration, the controller allows the
node to send at most one `HaveTx` or one `ResetRoute` message, as explained next.

### When to adjust

The Redundancy Controller runs in a separate thread a control loop that periodically calls
`adjustRedundancy` in order to reach and maintain the target redundancy level.

The `adjustRedundancy` action:
1. First it calls `controllerActions` to compute the current redundancy level and determines the
next steps such as unblocking `HaveTx` messages or sending a `ResetRoute` message to a randomly
chosen peer.
2. After making the adjustment, it resets the transaction counters so that redundancy is computed
independently of past measurements.
```bluespec "actions" +=
action adjustRedundancy(node) =
    nondet randomPeer = oneOf(node.Peers())
    val res = node.RC().controllerActions()
    val updatedNodeRC = res._1
    val sendResetRoute = res._2
    all {
        incomingMsgs' = 
            if (sendResetRoute) 
                node.send(incomingMsgs, randomPeer, ResetRouteMsg)
            else incomingMsgs,
        rc' = rc.put(node, updatedNodeRC.resetCounters()),
    }
```

Adjustments should be paced by a timer to account for message propagation delays. If redundancy
adjustments are made too frequently, a node risks isolation as all peers may cut routes prematurely.
The timer should align with the network’s maximum round-trip time (RTT) to allow `HaveTx` and
`ResetRoute` messages to propagate and take effect before initiating further adjustments.
Consequently, the controller is designed to send at most one control message per iteration in order
to prevent over-correction.

For example, suppose node `A` receives a duplicate transaction from `B` and replies with a `HaveTx`
message. Until `B` receives and process the `HaveTx` message, thus cutting a route to `A`, it will
pass at least a round-trip time (RTT) until `A` stops seeing traffic from `B`. In the meantime, `A`
may still continue to receive duplicates from `B` and other peers, causing `A`'s redundancy level to
be high. During that time, `A` should not send `HaveTx` messages to other peers because that may end
up cutting all routes to it. See the `adjustInterval` [parameter](#parameters) for more details.

> An alternative to triggering `adjustRedundancy` at fixed time intervals is to base it on the
> number of received transactions. While this approach eliminates dependency on time constraints, it
> introduces vulnerabilities that could destabilize nodes. For example, an attacker could exploit
> this mechanism by sending bursts of numerous small transactions to a node, causing the node to
> trigger `adjustRedundancy` too frequently. This results in the near-continuous activation of
> `HaveTx` messages, leading nodes to repeatedly alternating between sending `HaveTx` and
> `ResetRoute` messages. On testnets, we have observed that nodes continue to operate normally,
> except that bandwidth does not decrease as expected.

## Parameters

The following parameters must be configured by each node at initialization.

* `TargetRedundancy`: the desired redundancy level that the controller aims to keep as target
  (within specified bounds).
    ```bluespec "params" +=
    const TargetRedundancy: int
    ```
    A target equal to 0 partially disables the Redundancy Control mechanism: the controller can
    block `HaveTx` messages but cannot send `ResetRoute` messages. Zero redundancy minimizes
    bandwidth usage, achieving the lowest possible message overhead. In non-Byzantines networks,
    this is the best possible scenario. However, in Byzantine networks it could potentially render
    nodes isolated from transaction data. Therefore, the target should be set to a value greater
    than 0. Experimental results suggest a value between 0.5 and 1, which is a safe number that does
    not result in excessive duplicate transactions.

    > Note: `TargetRedundancy` should ideally be specified as a real number, but reals are not
    > currently supported by Quint.

* `TargetRedundancyDeltaPercent`: a percentage (a number in the open interval `(0, 100)`) of
  `TargetRedundancy` that defines acceptable lower and upper bounds for redundancy levels as a
  deviation from the target value.
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
    > were allowed by Quint.

* `adjustInterval`: the time (in milliseconds) that the controller waits between successive
  calls of `adjustRedundancy`.
    ```bluespec "params" +=
    const adjustInterval: int
    ```
	This interval should allow sufficient time for control messages (`HaveTx` and `ResetRoute`) to
	propagate through the network and take effect.
	
    A minimum value for `adjustInterval` depends on the network’s round-trip time (RTT). Assuming that
    global latency typically stays below 500ms (maybe an excessive number, need references here), the
    interval should be set to at least 1000ms to ensure stability and avoid over-adjustment.

    Optimal values depend on empirical measurements of network latency. However, in practice, a value of
    1000ms or above is recommended to allow message processing and delivery times in diverse network
    environments and load scenarios.
    
    This value is also related to the number of peers a node has, which determines the maximum number of
    routes. For instance, a node with 50 peers will have a maximum of `50 * (50 - 1) = 2450` routes.
    Hypothetically, removing one route on every adjustment, at one adjustment per second, it would take
    40.8 minutes to remove all routes from the node.

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

The following are all the state transitions allowed in the protocol. The rest of the section
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

3. A node disseminates a transaction currently in its mempool to a subset of peers.
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

  > In the actual implementation, each peer has its own independent dissemination goroutine,
  > resulting in transactions being sent to different peers at different times. However, for
  > simplicity, in this spec we model all of these actions in one atomic step.

4. A node joins the network (same as in Flood).
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

5. A node disconnects from the network.
    ```bluespec "steps" +=
    // Node disconnects from network
    all {
        require(size(nodesInNetwork) > 1),
        nondet node = oneOf(nodesInNetwork) 
        peers' = peers.disconnect(node),
        incomingMsgs' = incomingMsgs,
        mempool' = mempool,
        senders' = senders,
        dr' = dr,
        rc' = rc,
    },
    ```

6. A node detects that a peer is disconnected from the network.
    ```bluespec "steps" +=
    // Node detects a peer is disconnected
    nondet node = oneOf(nodesInNetwork)
    all {
        require(node.disconnectedPeers().nonEmpty()),
        nondet peer = oneOf(node.disconnectedPeers()) 
        node.updateDisconnectedPeer(peer),
        mempool' = mempool,
        senders' = senders,
    },
    ```

7. The Redundancy Controller periodically tries to adjust the redundancy level.
    ```bluespec "steps" +=
    // Redundancy Controller process loop
    all {
        nondet node = oneOf(nodesInNetwork)
        node.adjustRedundancy(),
        peers' = peers,
        mempool' = mempool,
        senders' = senders,
        dr' = dr,
    },
    ```

    > This action should ideally have a pre-condition that enables the state transition only on
    > specified time intervals, as defined by the `adjustInterval` parameter. However, Quint does
    > not natively support specifying time constraints without relying on workarounds such as using
    > counters as clocks. As a result, this action is modelled as always being enabled. In a
    > real-world implementation, redundancy adjustments would be naturally triggered by a timer set
    > to the duration specified by `adjustInterval`.
    
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
    | ResetRouteMsg => node.handleResetRouteMessage(_incomingMsgs, sender)
    }
```

* **Handling `HaveTx` messages**

    Upon receiving `HaveTxMsg(txID)` from `sender`, `node` disables the route `firstSender -> sender`, 
    where `firstSender` is the first node in `txID`'s list of senders, if `txID` actually
    comes from a peer. This action will decrease the traffic to `sender`.
    ```bluespec "actions" +=
    action handleHaveTxMessage(node, _incomingMsgs, sender, txID) = all {
        val txSenders = node.sendersOf(txID)
        dr' = dr.update(node, drs => 
            if (length(txSenders) > 0) drs.disableRoute((txSenders[0], sender)) else drs),
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
    peer from which `node` received `tx` for the first time. Subsequent entries in the list are
    nodes whose transaction messages arrived later as duplicates. As such, most routes from those
    peers to `node` will eventually be disabled, with most traffic coming primarily from the first
    peer.

* **Handling ResetRoute messages**

    Upon receiving `ResetRouteMsg`, `node` re-enables a random disabled route that has `sender` as
    target.
    ```bluespec "actions" +=
    action handleResetRouteMessage(node, _incomingMsgs, sender) = all {
        nondet randomRoute = oneOf(node.DisabledRoutes().routesWithTarget(sender))
        dr' = dr.update(node, drs => drs.enableRoute(randomRoute)),
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
        mempool' = mempool,
        senders' = senders,
        rc' = rc,
    }
    ```
    This will allow some traffic to flow again to `sender`. Other nodes will dynamically adapt to
    the new traffic, closing routes when needed.

    The protocol re-enables only one route per `ResetRoute` message to allow traffic to `sender` to
    increase gradually. If that peer still needs more transactions, it will send another
    `ResetRoute` message at a later time.

### Transaction dissemination

As in Flood, DOG will filter out the transaction's senders. Additionally, `node` will not send `tx`
to peer `B` if the route `A -> B` is disabled, where `A` is the first node in the list of `tx`'s
senders.
```bluespec "actions" +=
def mkTargetNodes(node, tx) =
    val txSenders = node.sendersOf(hash(tx))
    val disabledTargets = 
        if (length(txSenders) > 0)
            node.DisabledRoutes().routesWithSource(txSenders[0]).mapTargets()
        else Set()
    node.Peers()
        .exclude(txSenders.listToSet())
        .exclude(disabledTargets)
```
The protocol selects the first sender in the list based on the same reasoning applied when handling
received `HaveTx` messages. The first sender is likely responsible for the majority of traffic
related to the transaction, as it was the first to forward it to the node. Subsequent senders in the
list only sent the transaction later as duplicates, and their routes are more likely already
disabled.

### Nodes disconnecting from the network

When a `node` detects that a `peer` has disconnected from the network, 
1. it updates its set of active peers,
2. it updates its routing table by resetting all routes that have `peer` as either a source or target, and
3. it triggers a redundancy adjustment.
```bluespec "actions" +=
action updateDisconnectedPeer(node, peer) = all {
    peers' = peers.update(node, ps => ps.exclude(Set(peer))),
    dr' = dr.update(node, drs => drs.resetRoutes(peer)),
    node.adjustRedundancy(),
}
```
Calling `adjustRedundancy` is not strictly needed here because the protocol will make an adjustment
on the next iteration of the Redundancy Controller. This is just an improvement to trigger it
sooner.

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
