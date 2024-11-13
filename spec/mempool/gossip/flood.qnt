// -*- mode: Bluespec; -*-

/*******************************************************************************
Flood is a basic _push_ gossip protocol: every time a node receives a
transaction, it forwards (or "pushes") the transaction to all its peers, except
to the peer(s) from which it received the transaction.
*******************************************************************************/
module flood {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    export mempool.*

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    // Nodes communicates only one type of message carrying a full transaction.
    type Message =
        | TxMsg(TX)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Upon receiving a message with transaction `tx` from a peer (i.e., the
    // `sender`), the `node` attempts to add `tx` to its mempool. 
    //
    // The argument `_incomingMsgs` is passed just to update the queues of
    // incoming messages, when applicable (Flood does not reply with any message
    // but DOG does).
    action handleMessage(node, _incomingMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
        }

    // A node attempting to add a transaction to its mempool processes the
    // transaction according to whether it has seen it before, that is, if the
    // transaction exists in the mempool cache. We call a *first-time*
    // transaction to one that the node does not have in its cache. A
    // *duplicate* transaction is one that the node has received multiple times,
    // and thus it's cached.
    //
    // The sender is optional. When there's a sender, it means that the
    // transaction comes from a peer; otherwise it comes directly from a user.
    action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
        else
            node.processDuplicateTx(_incomingMsgs, optionalSender, tx)

    // `tryAddFirstTimeTx` attempts to add a first-time transaction `tx` to a
    // `node`'s mempool: (1) it caches `tx`, (2) if `tx` is valid, it adds `tx`
    // to the pool and (3) updates its senders.
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

    // `processDuplicateTx` processes a duplicate transaction `tx` by updating
    // the list of senders, only if `tx` is already in pool.
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

    //--------------------------------------------------------------------------
    // `mkTargetNodes` defines the set of peers to whom `node` will send `tx`.
    // It is passed as an argument to the generic transaction dissemination
    // action.
    //
    // In Flood, a node sends a transaction to all its peers except those who
    // previously sent it.
    def mkTargetNodes(node, tx) =
        node.Peers().exclude(node.sendersOf(tx))

    //--------------------------------------------------------------------------
    // All possible state transitions in the protocol.
    action step = any {
        // Transaction dissemination: node sends transaction to subset of peers.
        nondet node = oneOf(nodesInNetwork)
        node.disseminateNextTx(mkTargetNodes, TxMsg),

        // User-submitted transactions: node receives a transaction from a user.
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
        },

        // Nodes leaving the network.
        all {
            pickNodeAndDisconnect,
            state' = state,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------

    // A given transaction is in the pool of all nodes.
    def txInAllPools(tx) =
        NodeIDs.forall(n => n.Pool().includes(tx))

    // If a transaction is in the pool of any node, then eventually the
    // transaction will reach the pool of all nodes (maybe more than once, and
    // assuming transactions are not removed from mempools).
    temporal txInPoolGetsDisseminated = 
        Txs.forall(tx => 
            NodeIDs.exists(node =>
                node.Pool().includes(tx) implies eventually(txInAllPools(tx))))

    // Invariant: if node A sent a transaction tx to node B (A is in the list of
    // tx's senders), then B does not send tx to A (the message won't be in A's
    // incoming messages).
    val noSendToSender =
        NodeIDs.forall(nodeA => 
            NodeIDs.forall(nodeB => 
                Txs.forall(tx =>
                    nodeB.sendersOf(tx).contains(nodeA) 
                    implies
                    not(nodeA.IncomingMsgs().includes((nodeB, TxMsg(tx))))
        )))

}
