// -*- mode: Bluespec; -*-

// 👉 This is WORK IN PROGRESS; not everything here is correct 👈 

/*******************************************************************************
Dynamic Optimal Graph (DOG) gossip protocol.

The DOG protocol is built on top of the Flood protocol. All types, messages, and
data structures in Flood are also present in this specification.

Briefly, if node A receives from B a transaction that it already has, this means
that there must exist a cycle in the network topology, so A tells B to stop
sending transactions. The DOG protocol adds two new messages to Flood: 
- HaveTxMsg(TxID) for cutting cycles, and 
- ResetMsg for dynamic re-routing when a node disconnects.
*******************************************************************************/
module dog {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    import flood as F from "./flood"

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    type Message =
        | TxMsg(TX)
        // HaveTxMsg messages carry a transaction hash, which is typically much
        // smaller than full transactions.
        // "I already have this tx; don't send me more txs from the same
        // source."
        | HaveTxMsg(TX)
        // "My situation has changed; reset my routing data on your side."
        | ResetMsg

    //--------------------------------------------------------------------------
    // Routing
    //--------------------------------------------------------------------------

    // A route is a tuple `(source, target)`. We also write it as `A -> B`.
    type Route = (NodeID, NodeID)

    // The set of disabled routes on each node.
    var disabledRoutes: NodeID -> Set[Route]

    // Auxiliary definitions
    def DisabledRoutes(node) = disabledRoutes.get(node)
    
    //--------------------------------------------------------------------------
    // Disable a routes (sender(tx) -> sender), where sender(tx) is the first
    // element of txSenders.
    //
    // TODO: fix definition
    pure def addRoute(_disabledRoutes, node, txSenders, sender) =
        if (not(txSenders.listIsEmpty()))
            _disabledRoutes.update(node, rs => rs.setAdd((txSenders[0], sender))) // TODO: check
        else _disabledRoutes 

    // Remove any route that has peer as source or target. 
    pure def removeRoute(routes: Set[Route], peer: NodeID): Set[Route] = 
        routes.filter(route => not(peer.isSourceOrTargetIn(route)))

    // Is the node a source or target in the given route.
    pure def isSourceOrTargetIn(node, route) =
        node == route._1 or node == route._2

    //--------------------------------------------------------------------------
    // Redundancy Control (RC) mechanism
    //--------------------------------------------------------------------------

    // Definition: A transaction is received *for the first time* by a node,
    // when the node does not have the transaction in its cache. Otherwise the
    // transaction is a *duplicate*.

    type RedundancyControlState = {
        // Counter for transactions received for the first time by a node.
        firstTimeTxs: int,
        // Counter for duplicate transactions received by a node. 
        duplicateTxs: int,
        // Whether the node can reply with a HaveTx message upon receiving a
        // duplicate transaction.
        isHaveTxBlocked: bool,
    }

    // Each node keeps a state of the RC mechanism.
    var redundancyControl: NodeID -> RedundancyControlState

    // Auxiliary definitions
    def RC(node) = redundancyControl.get(node)
    pure def increaseFirstTimeTxs(rc) = { firstTimeTxs: rc.firstTimeTxs+1, ...rc }
    pure def increaseDuplicateTxs(rc) = { duplicateTxs: rc.duplicateTxs+1, ...rc }
    pure def blockHaveTx(rc) = { isHaveTxBlocked: false, ...rc }
    
    //--------------------------------------------------------------------------
    // Parameters/Configuration
    //--------------------------------------------------------------------------

    // The redundancy level that the gossip protocol should aim to maintain.
    //
    // When TargetRedundancyCfg is 0, the Redundancy Control mechanism is partially
    // disabled (see `adjustRedundancy` below).
    //
    // It should be a real type, but reals are not currently supported by Quint.
    const TargetRedundancyCfg: int

    // Value in the range `[0, 100)` that defines the bounds of acceptable
    // redundancy levels; redundancy +- redundancy * delta
    //
    // It should be a real type, but reals are not currently supported by Quint.
    const TargetRedundancyDeltaPercentCfg: int

    //  How many (first-time) transactions should the node receive before
    //  attempting to adjust redundancy.
    const TxsPerAdjustmentCfg: int

    // Constants
    val _delta = TargetRedundancyCfg * TargetRedundancyDeltaPercentCfg / 100
    val redundancyLowerBound = TargetRedundancyCfg - _delta
    val redundancyUpperBound = TargetRedundancyCfg + _delta

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state.
    action D_init = all {
        F::init,
        disabledRoutes' = NodeIDs.mapBy(_ => Set()),
        redundancyControl' = NodeIDs.mapBy(_ => { 
            firstTimeTxs: 0, duplicateTxs: 0, isHaveTxBlocked: false 
        })
    }

    action handleMessage(node, updatedMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(updatedMsgs, Some(sender), tx)
        | HaveTxMsg(tx) => node.handleHaveTxMessage(updatedMsgs, sender, tx)
        | ResetMsg => node.handleResetMessage(updatedMsgs, sender)
        }

    // The same as in Flood. The difference is in the tryAddTx* functions.
    action tryAddTx(node, updatedMsgs, optionalSender, tx) = 
        if (not(node.Cache().contains(hash(tx))))
            node.tryAddFirstTimeTx(updatedMsgs, optionalSender, tx)
        else
            node.tryAddDuplicateTx(updatedMsgs, optionalSender, tx)

    // Upon receiving tx for the first time, add tx to cache, add tx to pool and
    // update tx's senders if tx is valid, and update messages (the same as in
    // Flood). Additionally, increase `rc.firstTimeTxs`, and every
    // `TxsPerAdjustmentCfg` transactions received for the first time, call
    // `adjustRedundancy()`.
    action tryAddFirstTimeTx(node, updatedMsgs, optionalSender, tx) = 
        val rc1 = node.RC().increaseFirstTimeTxs()
        val adjustResult = 
            if (rc1.firstTimeTxs == TxsPerAdjustmentCfg) 
                rc1.adjustRedundancy() 
            else (rc1, false)
        val rc2 = adjustResult._1
        val sendReset = adjustResult._2
        all {
            val updatedMsgs2 = 
                val targets = optionalSender.optionToSet() // may be empty
                if (sendReset) 
                    node.send(updatedMsgs, targets, ResetMsg)
                else updatedMsgs
            node.F::tryAddFirstTimeTx(updatedMsgs2, optionalSender, tx),
            redundancyControl' = redundancyControl.update(node, _ => rc2),
            disabledRoutes' = disabledRoutes,
        }

    // Return an updated RC state and whether to reply with a Reset message.
    //
    // When `TargetRedundancyCfg = 0`, the lower and upper bounds are also equal to
    // 0. Then every `TxsPerAdjustmentCfg` received transactions `adjustRedundancy`
    // will unblock `HaveTx` but it will not send `Reset` messages.
    pure def adjustRedundancy(rc: RedundancyControlState): (RedundancyControlState, bool) = 
        val updatedRC = { firstTimeTxs: 0, duplicateTxs: 0, ...rc }
        val redundancy = rc.duplicateTxs / rc.firstTimeTxs
        if (redundancy < redundancyLowerBound)
            (updatedRC, true)
        else if (redundancy >= redundancyUpperBound)
            (updatedRC.blockHaveTx(), false)
        else 
            (updatedRC, false)

    // Upon receiving an already seen transaction, update tx's senders if tx is
    // in pool (and thus it's valid), and update messages (the same as in
    // Flood). Additionally, increase `duplicateTxs` and reply a `HaveTx`
    // message if the RC mechanism is not blocking it (and there's a sender).
    action tryAddDuplicateTx(node, updatedMsgs, optionalSender, tx) =
        // Reply HaveTx if tx comes from a peer.
        //
        // TODO: send HaveTxMsg message only if there's no incoming message
        //       from sender(tx) and sender(tx) has the route to this node
        //       disabled. This is to simulate the timeout that disables
        //       sending HaveTxMsg.
        val updatedMsgs2 = 
            val targets = optionalSender.optionToSet() // may be empty
            if (not(node.RC().isHaveTxBlocked))
                node.send(updatedMsgs, targets, HaveTxMsg(tx))
            else updatedMsgs
        all {
            node.F::tryAddDuplicateTx(updatedMsgs2, optionalSender, tx),
            redundancyControl' = redundancyControl.update(node, increaseDuplicateTxs),
            disabledRoutes' = disabledRoutes,
        }

    // Upon receiving HaveTxMsg(tx), disable the route (sender(tx) -> sender).
    // This will decrease the traffic to the sender.
    //
    // With this rule, we don't want to cut all routes from sender to node, only
    // those that come from the tx's original sender. That is why we need to
    // take the source into account in rule `R3`.
    action handleHaveTxMessage(node, updatedMsgs, sender, tx) = all {
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set()) // in most cases, there exists one and only one sender for tx (??)
        disabledRoutes' = disabledRoutes.addRoute(node, txSenders.setToList(), sender), // TODO: check setToList
        msgs' = updatedMsgs,
        peers' = peers,
        state' = state,
        redundancyControl' = redundancyControl,
    }

    // Upon receiving ResetMsg, remove any route that has sender as source or
    // target. This will allow traffic to flow again to the sender, and other
    // nodes will dynamically adapt to the new traffic, closing routes when
    // needed.
    action handleResetMessage(node, updatedMsgs, sender) = all {
        disabledRoutes' = disabledRoutes.update(node, drs => drs.removeRoute(sender)),
        msgs' = updatedMsgs,
        peers' = peers,
        state' = state,
        redundancyControl' = redundancyControl,
    }

    //--------------------------------------------------------------------------
    // For dissemination. Send tx to all peers except if the route `sender(tx)
    // -> peer` is disabled.
    //
    // Broadcast tx to all the peers of node, except if peer is sender(tx) or if
    // the route sender(tx) -> peer is disabled, where sender(tx) is any of the
    // peers that have sent tx to node.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set())
        val disabledTargets = node.DisabledRoutes().filter(r => txSenders.contains(r._1)).map(r => r._2)
        // TODO: remove exclude(txSenders)?
        node.Peers().exclude(txSenders).exclude(disabledTargets)

    //--------------------------------------------------------------------------
    // When a node disconnects from the network, its peers signal their peers
    // that their situation has changed and their routing tables should be reset
    // so that data can be re-routed if needed.
    action onDisconnect(node) = all {
        // All node's peers detect that node has disconnect and send a Reset
        // message to all their peers.
        val updatedMsgs = node.Peers().fold(msgs, 
            (_msgs, peer) => peer.send(_msgs, peer.Peers(), ResetMsg)
        )
        node.disconnect(updatedMsgs),
        // The node's peers remove node from their routing tables.
        disabledRoutes' = disabledRoutes.updateMultiple(node.Peers(), drs => drs.removeRoute(node)),
        state' = state,
        redundancyControl' = redundancyControl,
    }

    //--------------------------------------------------------------------------
    // TODO: a malicious node sends the same tx to one or more of its peers,
    // maybe in coordination with another malicious node.
    // action gossipDuplicate(node) = ...

    //--------------------------------------------------------------------------
    // All possible state transitions.
    action nextState = any {
        // Transaction dissemination.
        nondet node = oneOf(nodesInNetwork)
        all {
            node.disseminateNextTx(mkTargetNodes, TxMsg),
            disabledRoutes' = disabledRoutes,
            redundancyControl' = redundancyControl,
        },

        // Handling received transactions from users.
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(Txs)
        node.receiveTxFromUser(tx, tryAddTx),

        // Handling received messages from peers.
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),

        // Nodes joining the network.
        all {
            joinSomeNode,
            disabledRoutes' = disabledRoutes,
            redundancyControl' = redundancyControl,
        },

        // Nodes leaving the network.
        all {
            // Pick a node to disconnect only if it is not the only node in
            // the network.
            require(size(nodesInNetwork) > 1),
            nondet node = oneOf(nodesInNetwork) 
            onDisconnect(node)
        }
    }

}
