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

    type Message =
        | TxMsg(TX)
        | HaveTxMsg(TxID)
        | ResetMsg

    //--------------------------------------------------------------------------
    // Routing
    //--------------------------------------------------------------------------
    type Route = (NodeID, NodeID)
    var dr: NodeID -> Set[Route]
    pure def disableRoute(_dr, node, source, target) =
        _dr.update(node, routes => routes.join((source, target)))
    pure def enableRoute(_dr, peer) = 
        _dr.filter(route => not(peer.isSourceOrTargetIn(route)))
    def DisabledRoutes(node) = dr.get(node)
    pure def isSourceOrTargetIn(node, route) = node == route._1 or node == route._2

    //--------------------------------------------------------------------------
    // Redundancy Controller
    //--------------------------------------------------------------------------
    type RedundancyController = {
        firstTimeTxs: int,
        duplicateTxs: int,
        isHaveTxBlocked: bool,
    }
    var rc: NodeID -> RedundancyController
    pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs
    pure def adjustRedundancy(_rc) = 
        if (_rc.redundancy() < redundancyLowerBound)
            (_rc.resetCounters(), true)
        else if (_rc.redundancy() >= redundancyUpperBound)
            (_rc.resetCounters().blockHaveTx(), false)
        else 
            (_rc.resetCounters(), false)
    def RC(node) = rc.get(node)
    val initialRCState = { firstTimeTxs: 0, duplicateTxs: 0, isHaveTxBlocked: false }
    pure def increaseFirstTimeTxs(_rc) = { firstTimeTxs: _rc.firstTimeTxs + 1, ..._rc }
    pure def increaseDuplicateTxs(_rc) = { duplicateTxs: _rc.duplicateTxs + 1, ..._rc }
    pure def resetCounters(_rc) = { firstTimeTxs: 0, duplicateTxs: 0, ..._rc }
    pure def blockHaveTx(_rc) = { isHaveTxBlocked: true, ..._rc }

    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    const TargetRedundancy: int
    val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
    val redundancyLowerBound = TargetRedundancy - _delta
    val redundancyUpperBound = TargetRedundancy + _delta
    val TargetRedundancyDeltaPercent: int = 5
    const TxsPerAdjustment: int
    pure def adjustNow(rc) =
        val threshold = TxsPerAdjustment / rc.redundancy()
        rc.firstTimeTxs >= max(MinTxsPerAdjustment, min(MaxTxsPerAdjustment, threshold))
    val MinTxsPerAdjustment = 10
    val MaxTxsPerAdjustment = 1000

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action DOG_init = all {
        Flood::init,
        dr' = NodeIDs.mapBy(_ => Set()),
        rc' = NodeIDs.mapBy(_ => initialRCState)
    }
    action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
        else
            node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
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
    action handleMessage(node, _incomingMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
        | HaveTxMsg(txID) => node.handleHaveTxMessage(_incomingMsgs, sender, txID)
        | ResetMsg => node.handleResetMessage(_incomingMsgs, sender)
        }
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
    action handleResetMessage(node, _incomingMsgs, sender) = all {
        dr' = dr.update(node, drs => drs.enableRoute(sender)),
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
        mempool' = mempool,
        senders' = senders,
        rc' = rc,
    }
    def mkTargetNodes(node, tx) =
        val txSenders = node.sendersOf(hash(tx)).listToSet()
        val disabledTargets = node.DisabledRoutes()
            // Keep only routes whose source is one of tx's senders.
            .filter(r => r._1.in(txSenders))
            // Keep routes' targets.
            .map(r => r._2)
        node.Peers().exclude(txSenders).exclude(disabledTargets)
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

    action step = any {
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(AllTxs)
        node.receiveTxFromUser(tx, tryAddTx),
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),
        nondet node = oneOf(nodesInNetwork)
        all {
            node.disseminateNextTx(mkTargetNodes, TxMsg),
            senders' = senders,
            dr' = dr,
            rc' = rc,
        },
        all {
            pickNodeAndJoin,
            mempool' = mempool,
            senders' = senders,
            dr' = dr,
            rc' = rc,
        },
        all {
            // Pick a node that is not the only node in the network.
            require(size(nodesInNetwork) > 1),
            nondet nodeToDisconnect = oneOf(nodesInNetwork) 
            disconnectAndUpdateRoutes(nodeToDisconnect),
        }
    }

}
