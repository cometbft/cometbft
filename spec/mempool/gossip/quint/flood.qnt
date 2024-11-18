// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module flood {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    export mempool.*

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------
    type Message =
        | TxMsg(TX)

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    var senders: NodeID -> TxID -> List[NodeID]
    
    // Auxiliary definitions
    def Senders(node) = senders.get(node)
    def sendersOf(node, tx) = 
        node.Senders().mapGetDefault(hash(tx), List()).listToSet()
    pure def addSender(_txSenders, tx, optionalSender) = 
        match optionalSender {
        | Some(sender) => _txSenders.update(hash(tx), ss => 
            if (ss.includes(sender)) ss else ss.append(sender))
        | None => _txSenders
        }

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action init = all {
        MP_init,
        senders' = NodeIDs.mapBy(n => Map()),
    }
    action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
        else
            node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
    action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = all {
        mempool' = mempool.update(node, st => {
            cache: st.cache.join(hash(tx)),
            txs: if (valid(tx)) st.txs.append(tx) else st.txs,
            ...st }),
        senders' = senders.update(node, ss =>
            if (valid(tx)) ss.addSender(tx, optionalSender) else ss),
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
    }
    action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) = all {
        senders' = senders.update(node, ss =>
            if (node.Txs().includes(tx)) ss.addSender(tx, optionalSender) else ss),
        mempool' = mempool,
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
    }
    action handleMessage(node, _incomingMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
        }
    def mkTargetNodes(node, tx) =
        node.Peers().exclude(node.sendersOf(tx))

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
        },
        all {
            pickNodeAndJoin,
            mempool' = mempool,
            senders' = senders,
        },
        all {
            pickNodeAndDisconnect,
            mempool' = mempool,
            senders' = senders,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    def txInAllMempools(tx) =
        NodeIDs.forall(n => n.Txs().includes(tx))
    temporal txInPoolGetsDisseminated = 
        AllTxs.forall(tx => 
            NodeIDs.exists(node =>
                node.Txs().includes(tx) implies eventually(txInAllMempools(tx))))
    val dontSendBackToSender =
        NodeIDs.forall(nodeA => 
            NodeIDs.forall(nodeB => 
                AllTxs.forall(tx =>
                    nodeB.sendersOf(tx).contains(nodeA) 
                    implies
                    not(nodeA.IncomingMsgs().includes((nodeB, TxMsg(tx))))
        )))

}
