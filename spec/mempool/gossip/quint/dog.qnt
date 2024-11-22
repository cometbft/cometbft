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
    // Parameters
    //--------------------------------------------------------------------------
    const TargetRedundancy: int
    const TargetRedundancyDeltaPercent: int
    val _delta = TargetRedundancy * TargetRedundancyDeltaPercent / 100
    val lowerBound = TargetRedundancy - _delta
    val upperBound = TargetRedundancy + _delta
    const adjustInterval: int

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
    def RC(node) = rc.get(node)
    val initialRCState = { firstTimeTxs: 0, duplicateTxs: 0, isHaveTxBlocked: false }
    pure def increaseFirstTimeTxs(_rc) = { firstTimeTxs: _rc.firstTimeTxs + 1, ..._rc }
    pure def increaseDuplicateTxs(_rc) = { duplicateTxs: _rc.duplicateTxs + 1, ..._rc }
    pure def resetCounters(_rc) = { firstTimeTxs: 0, duplicateTxs: 0, ..._rc }
    pure def blockHaveTx(_rc) = { isHaveTxBlocked: true, ..._rc }
    pure def redundancy(_rc) = _rc.duplicateTxs / _rc.firstTimeTxs
    pure def adjustRedundancy(node, _rc) =
        if (_rc.firstTimeTxs + _rc.duplicateTxs == 0)
            (_rc, false)
        else if (_rc.redundancy() < lowerBound)
            (_rc, true)
        else if (_rc.redundancy() >= upperBound)
            (_rc.blockHaveTx(), false)
        else 
            (_rc, false)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action DOG_init = all {
        Flood::init,
        dr' = NodeIDs.mapBy(_ => Set()),
        rc' = NodeIDs.mapBy(_ => initialRCState)
    }

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
    action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
        else
            node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
    action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = 
        all {
            node.Flood::tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx),
            rc' = rc.update(node, increaseFirstTimeTxs),
            dr' = dr,
        }
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
    pure def replyHaveTx(node, _incomingMsgs, _rc, optionalSender, tx) =
        if (optionalSender.isSome() and not(_rc.isHaveTxBlocked))
            val targets = optionalSender.optionToSet()
            (node.multiSend(_incomingMsgs, targets, HaveTxMsg(hash(tx))), _rc.blockHaveTx())
        else (_incomingMsgs, _rc)
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
        val txSenders = node.sendersOf(hash(tx))
        val disabledTargets = node.DisabledRoutes()
            // Keep routes whose source is tx's first sender, if any.
            .filter(r => if (length(txSenders) > 0) r._1 == txSenders[0] else false)
            // Map routes to targets.
            .map(r => r._2)
        node.Peers().exclude(txSenders.listToSet()).exclude(disabledTargets)
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
        // User-initiated transactions
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(AllTxs)
        node.receiveTxFromUser(tx, tryAddTx),
        // Peer message handling
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),
        // Transaction dissemination
        nondet node = oneOf(nodesInNetwork)
        all {
            node.disseminateNextTx(mkTargetNodes, TxMsg),
            senders' = senders,
            dr' = dr,
            rc' = rc,
        },
        // Node joins network
        all {
            pickNodeAndJoin,
            mempool' = mempool,
            senders' = senders,
            dr' = dr,
            rc' = rc,
        },
        // Node disconnects from network
        all {
            // Pick a node that is not the only node in the network.
            require(size(nodesInNetwork) > 1),
            nondet nodeToDisconnect = oneOf(nodesInNetwork) 
            disconnectAndUpdateRoutes(nodeToDisconnect),
        },
        // Redundancy Controller process loop
        nondet node = oneOf(nodesInNetwork)
        node.controlLoopIteration(),
    }

}
