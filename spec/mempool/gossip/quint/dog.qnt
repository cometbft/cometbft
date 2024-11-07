// -*- mode: Bluespec; -*-

// 👉 This is WORK IN PROGRESS; not everything here is correct 👈 

/*******************************************************************************
Dynamic Optimal Graph (DOG) gossip protocol.

The DOG protocol is built on top of the Flood protocol. All types, messages, and
data structures in Flood are present in this specification.

Briefly, what the protocol does is the following. If node A receives from B a
transaction that it already has, this means that there must exist a cycle in the
network topology, so A tells B to stop sending transactions. Conversely, if A
does not receive enough transactions, it will ask its peers to send more. For
that, the protocol adds two new messages to Flood: 
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
        // A message carrying a full transaction as payload (same as in Flood).
        | TxMsg(TX)
        // HaveTxMsg messages carry a transaction hash, typically much smaller
        // than full transactions. A node sends this message to signal that it
        // has the transaction, so that the receiver must cut some route that is
        // forming a cycle in the network topology. The goal for the sender is
        // to receive less transactions.
        | HaveTxMsg(TxID)
        // A node sends a Reset message to signal that it is not receiving
        // enough transactions, so that the receiver must re-enable some route
        // to the sender if possible. This message is usually sent when a peer
        // is disconnected or sending less transactions than before, so that the
        // sender receives more transactions.
        | ResetMsg

    //--------------------------------------------------------------------------
    // Routing
    //--------------------------------------------------------------------------
    // The protocol uses a routing mechanism on the nodes to filter TxMsg
    // messages sent to their peers.

    // A route is a tuple `(source, target)`. We also write it as `source ->
    // target`. Routes are defined within a node, and source and target are
    // peers connected to that node.
    type Route = (NodeID, NodeID)

    // The set of disabled routes per node. By default, a node has all of its
    // routes enabled, so its set of disabled routes is empty. A node A will
    // send `TxMsg(tx)` to peer B if the route `sender(tx) -> B` is enabled,
    // that is, the route is not in this set.
    var disabledRoutes: NodeID -> Set[Route]

    // Disable the route `source -> target` by adding it to a set of disabled
    // routes.
    pure def addRoute(_disabledRoutes, node, source, target) =
        _disabledRoutes.update(node, routes => routes.setAdd((source, target)))

    // Remove any route that has peer as source or target. 
    pure def removeRoute(_disabledRoutes, peer) = 
        _disabledRoutes.filter(route => not(peer.isSourceOrTargetIn(route)))

    // Auxiliary definitions
    def DisabledRoutes(node) = disabledRoutes.get(node)
    pure def isSourceOrTargetIn(node, route) = node == route._1 or node == route._2

    //--------------------------------------------------------------------------
    // Redundancy Controller (RC)
    //--------------------------------------------------------------------------
    // The Redundancy Controller is a closed-loop mechanism to auto-adjust the
    // level of redundant transactions tha each node receives. A node configures
    // a target redundancy level that it aims to maintain (within certain
    // boundaries). The closed-loop controller periodically measures the
    // redundancy level and either send messages to its peers or abstain from
    // sending messages to try to decrease or increase the number of
    // transactions received.
    //
    // Definition: A transaction is received *for the first time* by a node,
    // when the node does not have the transaction in its cache. When a node
    // receives a transaction more than once, it is called a *duplicate*
    // transaction.

    // A Redundancy Controller needs the following data structures.
    type RedundancyControlState = {
        // Counter of transactions received for the first time by a node.
        firstTimeTxs: int,
        // Counter of duplicate transactions received by a node. 
        duplicateTxs: int,
        // Whether the node is allowed to reply with a HaveTx message upon
        // receiving a duplicate transaction.
        isHaveTxBlocked: bool,
    }

    // Each node keeps an RC state to control its redundancy level.
    var rc: NodeID -> RedundancyControlState

    // The redundancy level is defined as the proportion of duplicate
    // transactions over the number of transactions received for the first time.
    // For example, a redundancy of 1 means that for each transaction that the
    // node has received for the first time, the node received one transaction
    // that is duplicate (not necessarily of the same transaction).
    pure def redundancyLevel(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs

    // adjustRedundancy returns an updated RC state and whether to reply with a
    // Reset message.
    //
    // Periodically, the node measures the redundancy level. If it is within
    // acceptable bounds (as defined below), it does nothing. Otherwise it will
    // either send Reset messages to peers in order to receive more transactions
    // and thus increase its redundancy level, or it will allow sending HaveTx
    // messages to peers so that they disable routes, thus decreasing the
    // traffic to the node.
    //
    // On every adjustment, the counters are reset, so that redundancy values
    // are measured within the adjustment period.
    //
    // When the target redundancy is 0, the lower and upper bounds are also
    // equal to 0. Then every `TxsPerAdjustment` received transactions
    // `adjustRedundancy` will unblock `HaveTx` but it will not send `Reset`
    // messages.
    pure def adjustRedundancy(_rc) = 
        if (_rc.redundancyLevel() < redundancyLowerBound)
            (_rc.resetCounters(), true)
        else if (_rc.redundancyLevel() >= redundancyUpperBound)
            (_rc.resetCounters().blockHaveTx(), false)
        else 
            (_rc.resetCounters(), false)

    // Auxiliary definitions
    def RC(node) = rc.get(node)
    val initialRCState = { firstTimeTxs: 0, duplicateTxs: 0, isHaveTxBlocked: false }
    pure def increaseFirstTimeTxs(_rc) = { firstTimeTxs: _rc.firstTimeTxs + 1, ..._rc }
    pure def increaseDuplicateTxs(_rc) = { duplicateTxs: _rc.duplicateTxs + 1, ..._rc }
    pure def resetCounters(_rc) = { firstTimeTxs: 0, duplicateTxs: 0, ..._rc }
    pure def blockHaveTx(_rc) = { isHaveTxBlocked: false, ..._rc }
    
    //--------------------------------------------------------------------------
    // Parameters/Configuration
    //--------------------------------------------------------------------------

    // The redundancy level that the protocol aims to maintain, within the
    // boundaries defined by TargetRedundancyDeltaPercent.
    //
    // When TargetRedundancy is 0, the Redundancy Control mechanism is partially
    // disabled (see `adjustRedundancy` below).
    //
    // It should be a real type, but reals are not currently supported by Quint.
    const TargetRedundancy: int

    // Value in the range `[0, 100)` that defines the bounds of acceptable
    // redundancy levels; redundancy +- redundancy * delta
    //
    // It should be a real type, but reals are not currently supported by Quint.
    //
    // This value was initially modelled as a model parameter (a `const`), but
    // we decided to keep it fixed.
    val TargetRedundancyDeltaPercent: int = 5

    // The number of first-time transactions received by a node before it
    // adjusts its redundancy level.
    const TxsPerAdjustment: int

    // Constants
    val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
    val redundancyLowerBound = TargetRedundancy - _delta
    val redundancyUpperBound = TargetRedundancy + _delta

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state.
    action D_init = all {
        F::init,
        disabledRoutes' = NodeIDs.mapBy(_ => Set()),
        rc' = NodeIDs.mapBy(_ => initialRCState)
    }

    action handleMessage(node, updatedMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(updatedMsgs, Some(sender), tx)
        | HaveTxMsg(txID) => node.handleHaveTxMessage(updatedMsgs, sender, txID)
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
    // `TxsPerAdjustment` transactions received for the first time, call
    // `adjustRedundancy()`.
    action tryAddFirstTimeTx(node, updatedMsgs, optionalSender, tx) = 
        val rc1 = node.RC().increaseFirstTimeTxs()
        val adjustResult = 
            if (rc1.firstTimeTxs == TxsPerAdjustment) 
                rc1.adjustRedundancy() 
            else (rc1, false)
        val rc2 = adjustResult._1
        val sendReset = adjustResult._2
        all {
            val updatedMsgs1 = 
                val targets = optionalSender.optionToSet() // may be empty
                if (sendReset) 
                    node.send(updatedMsgs, targets, ResetMsg)
                else updatedMsgs
            node.F::tryAddFirstTimeTx(updatedMsgs1, optionalSender, tx),
            rc' = rc.put(node, rc2),
            disabledRoutes' = disabledRoutes,
        }

    // Upon receiving an already seen transaction, the node updates tx's senders
    // if tx is in pool (and thus it's valid), and updates the list of incoming
    // messages (the same as in Flood). Additionally, it increases
    // `duplicateTxs` and reply a `HaveTx` message if the RC mechanism is not
    // blocking it and there's a sender.
    action tryAddDuplicateTx(node, updatedMsgs, optionalSender, tx) =
        // Reply HaveTxMsg if tx comes from a peer.
        val updatedMsgs1 = 
            val targets = optionalSender.optionToSet() // may be empty
            if (not(node.RC().isHaveTxBlocked))
                node.send(updatedMsgs, targets, HaveTxMsg(hash(tx)))
            else updatedMsgs
        all {
            node.F::tryAddDuplicateTx(updatedMsgs1, optionalSender, tx),
            rc' = rc.update(node, increaseDuplicateTxs),
            disabledRoutes' = disabledRoutes,
        }

    // Upon receiving HaveTxMsg(txID), disable the route (sender(txID) ->
    // sender). This will decrease the traffic to the sender.
    //
    // We don't want to cut all routes from sender to node, only the route that
    // has as source tx's original sender.
    action handleHaveTxMessage(node, updatedMsgs, sender, txID) = all {
        disabledRoutes' = 
            val txSenders = node.Senders().mapGetDefault(txID, List())
            if (length(txSenders) > 0)
                // Since we need to disable only one route, we need to choose
                // one of tx's senders. We choose the first sender in the list,
                // which is the first peer that sent the transaction to `node`.
                // The other senders in the list also sent tx but at a later
                // time, meaning that routes coming from those senders are
                // probably disabled, and most of the traffic comes from the
                // first one.
                disabledRoutes.addRoute(node, txSenders[0], sender)
            else disabledRoutes,
        msgs' = updatedMsgs,
        peers' = peers,
        state' = state,
        rc' = rc,
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
        rc' = rc,
    }

    //--------------------------------------------------------------------------
    // Return the set of peers to which the node will send tx when disseminating
    // transactions. As in Flood, DOG will filter out tx's senders.
    // Additionally, DOG will not send tx to peer if the route `sender(tx) ->
    // peer` is disabled.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), List())
        val disabledTargets = node.DisabledRoutes()
            // Keep only routes whose source is one of tx's senders.
            .filter(r => r._1.in(txSenders.listToSet()))
            // Keep routes' targets.
            .map(r => r._2)
        node.Peers().exclude(txSenders.listToSet()).exclude(disabledTargets)

    //--------------------------------------------------------------------------
    // When a node disconnects from the network, its peers signal their own
    // peers that their situation has changed, that their routing tables should
    // be reset. In this way, data via those nodes can be re-routed if needed.
    action D_disconnect(node) = all {
        // All node's peers detect that node has disconnect and send a Reset
        // message to all their peers.
        val updatedMsgs = node.Peers().fold(msgs, 
            (_msgs, peer) => peer.send(_msgs, peer.Peers(), ResetMsg)
        )
        node.disconnect(updatedMsgs),
        // The node's peers remove node from their routing tables.
        disabledRoutes' = disabledRoutes.updateMultiple(node.Peers(), drs => drs.removeRoute(node)),
        state' = state,
        rc' = rc,
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
            rc' = rc,
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
            pickNodeAndJoin,
            state' = state,
            disabledRoutes' = disabledRoutes,
            rc' = rc,
        },

        // Nodes leaving the network.
        all {
            // Pick a node to disconnect only if it is not the only node in
            // the network.
            require(size(nodesInNetwork) > 1),
            nondet node = oneOf(nodesInNetwork) 
            D_disconnect(node),
        }
    }

}
