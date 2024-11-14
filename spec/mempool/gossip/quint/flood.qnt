// -*- mode: Bluespec; -*-

// File generated from markdown using lmt. DO NOT EDIT.

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
    // Actions
    //--------------------------------------------------------------------------
    action handleMessage(node, _incomingMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
        }
    action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
        else
            node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
    action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = all {
        state' = state.update(node, st => {
            cache: st.cache.join(hash(tx)),
            pool: if (valid(tx)) st.pool.append(tx) else st.pool,
            senders: 
                if (valid(tx)) 
                    st.senders.addSender(tx, optionalSender) 
                else st.senders,
            ...st }),
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
    }
    action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) = all {
        state' = state.update(node, st => { 
            senders: 
                if (st.pool.includes(tx)) 
                    st.senders.addSender(tx, optionalSender) 
                else st.senders, 
            ...st }),
        incomingMsgs' = _incomingMsgs,
        peers' = peers,
    }
    def mkTargetNodes(node, tx) =
        node.Peers().exclude(node.sendersOf(tx))

    action step = any {
        nondet node = oneOf(nodesInNetwork)
        node.disseminateNextTx(mkTargetNodes, TxMsg),
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(Txs)
        node.receiveTxFromUser(tx, tryAddTx),
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),
        all {
            pickNodeAndJoin,
            state' = state,
        },
        all {
            pickNodeAndDisconnect,
            state' = state,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    def txInAllPools(tx) =
        NodeIDs.forall(n => n.Pool().includes(tx))
    temporal txInPoolGetsDisseminated = 
        Txs.forall(tx => 
            NodeIDs.exists(node =>
                node.Pool().includes(tx) implies eventually(txInAllPools(tx))))
    val noSendToSender =
        NodeIDs.forall(nodeA => 
            NodeIDs.forall(nodeB => 
                Txs.forall(tx =>
                    nodeB.sendersOf(tx).contains(nodeA) 
                    implies
                    not(nodeA.IncomingMsgs().includes((nodeB, TxMsg(tx))))
        )))

}
