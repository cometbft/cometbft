// -*- mode: Bluespec; -*-

// 👉 This is WORK IN PROGRESS; not everything here is correct 👈 

/*******************************************************************************
Dynamic Optimal Graph (DOG) gossip protocol.

The DOG protocol is built on top of the Flood protocol. Types, messages, and
data structures in Flood are used in this specification. In priciple, it is
possible to enable or disable DOG in nodes running Flood, and nodes running the
two protocols can co-exist in the network, though not at the best performance.

Briefly, what the protocol does is the following. If node A receives from B a
transaction that it already has, this means that there must exist a cycle in the
network topology, so A tells B to stop sending transactions. Conversely, if A
does not receive enough transactions, it will ask its peers to send more. For
that, the protocol adds two new messages to Flood: 
- HaveTxMsg(TxID) for cutting cycles in the network, and 
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
        // `TxMsg` messages carry a full transaction (same as in Flood).
        | TxMsg(TX)
        // `HaveTxMsg` messages carry a transaction hash, typically much smaller
        // than full transactions. A node sends this message to signal that it
        // has the transaction, so that the receiver must cut some route that is
        // forming a cycle in the network topology. The goal for the sender is
        // to receive less transactions.
        | HaveTxMsg(TxID)
        // `ResetMsg` is sent by a node to signal that is not receiving enough
        // transactions, so that the receiver must re-enable some route to the
        // sender if possible. This message is usually sent when a peer is
        // disconnected or sending less transactions than before, so that the
        // sender receives more transactions.
        | ResetMsg

    // In the real implementation, `TxMsg` is sent via a P2P mempool data
    // channel, and `HaveTxMsg` and `ResetMsg` via the P2P mempool control
    // channel.

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
    var dr: NodeID -> Set[Route]

    // Disable the route `source -> target` by adding it to a set of disabled
    // routes.
    pure def disableRoute(_dr, node, source, target) =
        _dr.update(node, routes => routes.join((source, target)))

    // Enable all routes to peer or from peer by removing any disabled route
    // that has peer as source or target. 
    pure def enableRoute(_dr, peer) = 
        _dr.filter(route => not(peer.isSourceOrTargetIn(route)))

    // Auxiliary definitions
    def DisabledRoutes(node) = dr.get(node)
    pure def isSourceOrTargetIn(node, route) = node == route._1 or node == route._2

    //--------------------------------------------------------------------------
    // Redundancy Controller (RC)
    //--------------------------------------------------------------------------
    // The Redundancy Controller is a closed-loop mechanism to auto-adjust the
    // level of redundant transactions that a node receives. A node configures a
    // target redundancy level that it aims to maintain and how often it should
    // make adjustments. The controller periodically computes the redundancy
    // level and tries to adjust it in order to keep it within certain
    // pre-defined accepted bounds (as defined below).

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
    // transactions over first-time transactions. For example, a redundancy of 1
    // means that for each transaction that the node has received for the first
    // time, the node has received one transaction that is a duplicate (not
    // necessarily a duplicate of the first-time transaction).
    pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs

    // adjustRedundancy returns an updated RC state and whether to reply with a
    // Reset message.
    //
    // Periodically, the RC of the node computes the redundancy level. If the
    // level is too low, it will try to increase it by sending to one of its
    // peers a Reset message, thus increasing the number of transactions
    // received. If the level is too high, it will try to decrease it by
    // allowing sending HaveTx messages, so that the number of transactions
    // received decline. Otherwise it does nothing.
    //
    // On every adjustment, the counters are reset. This means that redundancy
    // is computed for the period elapsed since the last adjustment.
    //
    // When the target redundancy is 0, the lower and upper bounds are also
    // equal to 0. Then every `TxsPerAdjustment` received transactions
    // `adjustRedundancy` will unblock `HaveTx` but it will not send `Reset`
    // messages.
    pure def adjustRedundancy(_rc) = 
        if (_rc.redundancy() < redundancyLowerBound)
            (_rc.resetCounters(), true)
        else if (_rc.redundancy() >= redundancyUpperBound)
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

    // `TargetRedundancy` is a threshold level that the protocol aims to
    // maintain, within the boundaries defined by
    // `TargetRedundancyDeltaPercent`.
    //
    // When `TargetRedundancy` is 0, the Redundancy Control mechanism is
    // partially disabled: `adjustRedundancy` will be able only to block HaveTx
    // messages but not to send Reset messages. This makes the protocol not
    // resistant to malicious behaviour. Therefore, in practice, it is not
    // recommended to set it to 0 in Byzantine networks.
    //
    // It should be a real type, but reals are not currently supported by Quint.
    const TargetRedundancy: int

    // Value in the range `[0, 100)` as a percentage of the target redundancy
    // that defines the lower and upper bounds of acceptable redundancy levels.
    //
    // It should be a real type, but reals are not currently supported by Quint.
    //
    // This value was initially modelled as a model parameter (a `const`), but
    // we decided to keep it fixed.
    val TargetRedundancyDeltaPercent: int = 5

    // The number of first-time transactions that the Redundancy Controller must
    // wait for the node to receive before it calls `adjustRedundancy`.
    const TxsPerAdjustment: int

    // Whether is it time to make an adjustment to the redundancy level.
    pure def adjustNow(rc) = 
        rc.firstTimeTxs == TxsPerAdjustment / rc.redundancy()

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
        dr' = NodeIDs.mapBy(_ => Set()),
        rc' = NodeIDs.mapBy(_ => initialRCState)
    }

    // How to handle messages from peers.
    action handleMessage(node, _msgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_msgs, Some(sender), tx)
        | HaveTxMsg(txID) => node.handleHaveTxMessage(_msgs, sender, txID)
        | ResetMsg => node.handleResetMessage(_msgs, sender)
        }

    // The same as in Flood. The difference is in the two functions that process
    // the transaction.
    action tryAddTx(node, _msgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_msgs, optionalSender, tx)
        else
            node.processDuplicateTx(_msgs, optionalSender, tx)

    // Attempt to add a received first-time transaction tx to the mempool: add
    // tx to cache, add tx to pool and update tx's senders if tx is valid, and
    // update messages (the same as in Flood). Additionally, increase
    // `rc.firstTimeTxs`, and every `TxsPerAdjustment` transactions received for
    // the first time, call `adjustRedundancy()`.
    action tryAddFirstTimeTx(node, _msgs, optionalSender, tx) = 
        val _rc1 = node.RC().increaseFirstTimeTxs()
        val _result = if (_rc1.adjustNow()) _rc1.adjustRedundancy() else (_rc1, false)
        val _rc2 = _result._1
        val sendReset = _result._2
        all {
            val updatedMsgs = 
                val targets = optionalSender.optionToSet() // may be empty
                if (sendReset) node.send(_msgs, targets, ResetMsg) else _msgs
            node.F::tryAddFirstTimeTx(updatedMsgs, optionalSender, tx),
            rc' = rc.put(node, _rc2),
            dr' = dr,
        }

    // Process a received duplicate transaction `tx`: increase `duplicateTxs`
    // and reply a `HaveTx` message if the RC mechanism is not blocking it (and
    // there's a sender). As in Flood, update the list of senders if `tx` is in
    // `pool` (and thus it's valid), and update the list of incoming messages. 
    action processDuplicateTx(node, _msgs, optionalSender, tx) =
        // Reply HaveTxMsg if tx comes from a peer.
        val updatedMsgs = 
            val targets = optionalSender.optionToSet() // may be empty
            if (not(node.RC().isHaveTxBlocked))
                node.send(_msgs, targets, HaveTxMsg(hash(tx)))
            else _msgs
        all {
            node.F::processDuplicateTx(updatedMsgs, optionalSender, tx),
            rc' = rc.update(node, increaseDuplicateTxs),
            dr' = dr,
        }

    // Upon receiving `HaveTxMsg(txID)`, disable the route `sender(txID) ->
    // sender`. This will decrease the traffic to the sender.
    //
    // We don't want to cut all routes from sender to node, only the route that
    // has as source tx's original sender.
    action handleHaveTxMessage(node, _msgs, sender, txID) = all {
        dr' = 
            val txSenders = node.Senders().mapGetDefault(txID, List())
            if (length(txSenders) > 0)
                // Since we want to disable only one route, we need to choose
                // one of tx's senders. We choose the first sender in the list,
                // which is the first peer that sent the transaction to `node`.
                // The other senders in the list also sent the transaction but
                // at a later time, meaning that routes coming from those
                // senders are probably disabled, and most of the traffic comes
                // from the first one.
                dr.disableRoute(node, txSenders[0], sender)
            else dr,
        msgs' = _msgs,
        peers' = peers,
        state' = state,
        rc' = rc,
    }

    // Upon receiving ResetMsg, remove any route that has sender as source or
    // target. This will allow traffic to flow again to the sender, and other
    // nodes will dynamically adapt to the new traffic, closing routes when
    // needed.
    action handleResetMessage(node, _msgs, sender) = all {
        dr' = dr.update(node, drs => drs.enableRoute(sender)),
        msgs' = _msgs,
        peers' = peers,
        state' = state,
        rc' = rc,
    }

    // When a node disconnects from the network, its peers signal their own
    // peers that their situation has changed, that their routing tables should
    // be reset. In this way, data via those nodes can be re-routed if needed.
    action disconnectAndUpdateRoutes(nodeToDisconnect) = all {
        // All node's peers detect that node has disconnect and send a Reset
        // message to all their peers.
        val updatedMsgs = nodeToDisconnect.Peers().fold(msgs, 
            (_msgs, peer) => peer.send(_msgs, peer.Peers(), ResetMsg)
        )
        nodeToDisconnect.disconnectNetwork(updatedMsgs),
        // The node's peers enable all routes to node in their routing tables.
        dr' = dr.updateMultiple(nodeToDisconnect.Peers(), 
            drs => drs.enableRoute(nodeToDisconnect)),
        state' = state,
        rc' = rc,
    }

    //--------------------------------------------------------------------------
    // Transaction dissemination: As in Flood, DOG will filter out the
    // transaction's senders. Additionally, DOG will not send `tx` to a peer if
    // the route `sender(tx) -> peer` is disabled.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), List())
        val disabledTargets = node.DisabledRoutes()
            // Keep only routes whose source is one of tx's senders.
            .filter(r => r._1.in(txSenders.listToSet()))
            // Keep routes' targets.
            .map(r => r._2)
        node.Peers().exclude(txSenders.listToSet()).exclude(disabledTargets)

    //--------------------------------------------------------------------------
    // All possible state transitions in the protocol.
    action nextState = any {
        // Transaction dissemination: node sends transaction to subset of peers.
        nondet node = oneOf(nodesInNetwork)
        all {
            node.disseminateNextTx(mkTargetNodes, TxMsg),
            dr' = dr,
            rc' = rc,
        },

        // User-initiated transactions: node receives a transaction from a user.
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(Txs)
        node.receiveTxFromUser(tx, tryAddTx),

        // Peer message handling: node processes messages received from peers.
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),

        // Nodes joining the network.
        all {
            pickNodeAndJoin,
            state' = state,
            dr' = dr,
            rc' = rc,
        },

        // Nodes leaving the network.
        all {
            // Pick a node that is not the only node in the network.
            require(size(nodesInNetwork) > 1),
            nondet nodeToDisconnect = oneOf(nodesInNetwork) 
            disconnectAndUpdateRoutes(nodeToDisconnect),
        }
    }

}
