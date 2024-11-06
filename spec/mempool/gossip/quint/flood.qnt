// -*- mode: Bluespec; -*-

/*******************************************************************************
Flood is a simple *push* gossip protocol: every time a node receives a
transaction, it forwards (or pushes) the transaction to all of its peers, except
to the peer(s) from which it received the transaction.
*******************************************************************************/
module flood {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    export mempool.*

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    // In this protocol, nodes communicates only one type of message, `TxMsg`,
    // that includes a full transaction.
    type Message =
        | TxMsg(TX)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Upon receiving a TxMsg(tx) message, try adding tx to the mempool.
    action handleMessage(node, updatedMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(updatedMsgs, Some(sender), tx)
        }

    // Upon receiving a transaction, try to add it to the mempool. The sender is
    // optional, meaning that if there's no sender, the transaction comes from a
    // user. If tx comes from a peer, the node needs to update the list of
    // incoming messages.
    //
    // A transaction that is in `pool`, must also be in `cache` (assuming an
    // infinite cache), but not necessarily the inverse. The reason a
    // transaction is in `cache` but not in `pool` is either because: 
    // - the transaction was initially invalid and never got into `pool`, 
    // - the transaction became invalid after it got in `pool` and thus got
    //   evicted while revalidating it, or
    // - the transaction was committed to a block and got removed from `pool`.
    action tryAddTx(node, updatedMsgs, optionalSender, tx) = 
        if (not(node.Cache().contains(hash(tx))))
            node.tryAddFirstTimeTx(updatedMsgs, optionalSender, tx)
        else
            node.tryAddDuplicateTx(updatedMsgs, optionalSender, tx)

    // Upon receiving tx for the first time, add tx to cache, add tx to pool and
    // update tx's senders if tx is valid, and update messages.
    action tryAddFirstTimeTx(node, updatedMsgs, optionalSender, tx) = all {
        state' = state.update(node, st => {
            cache: st.cache.setAdd(hash(tx)),
            pool: if (valid(tx)) st.pool.append(tx) else st.pool,
            senders: if (valid(tx)) st.senders.addSender(tx, optionalSender) else st.senders,
            ...st }),
        msgs' = updatedMsgs,
        peers' = peers,
    }

    // Upon receiving an already seen transaction, update tx's senders if tx is
    // in pool (and thus it's valid), and update messages.
    action tryAddDuplicateTx(node, updatedMsgs, optionalSender, tx) =  all {
        state' = state.update(node, st => { 
            senders: if (st.pool.includes(tx)) st.senders.addSender(tx, optionalSender) else st.senders, 
            ...st }),
        msgs' = updatedMsgs,
        peers' = peers,
    }

    // Return the target nodes for disseminating tx from node. In this case,
    // Flood will send tx to all peers except the tx's senders.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set())
        node.Peers().exclude(txSenders)

    //--------------------------------------------------------------------------
    // All possible state transitions.
    action nextState = any {
        // Transaction dissemination.
        nondet node = oneOf(nodesInNetwork)
        node.disseminateNextTx(mkTargetNodes, TxMsg),

        // Handling received transactions from users.
        nondet node = oneOf(nodesInNetwork)
        nondet tx = oneOf(Txs)
        node.receiveTxFromUser(tx, tryAddTx),

        // Handling received messages from peers.
        nondet node = oneOf(nodesInNetwork)
        node.receiveFromPeer(handleMessage),

        // Nodes joining or leaving the network.
        joinSomeNode,
        disconnectSomeNode(nodesInNetwork),
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------

    // All messages have been delivered.
    val allMsgsDelivered = 
        NodeIDs.forall(node => node.Msgs().listIsEmpty())

    // The pools of all nodes contain a given tx.
    def txInAllPools(tx) =
        NodeIDs.forall(n => state.get(n).pool.includes(tx))

    // All available transactions are in the pool of all nodes.
    val allTxsGossipped =
        NodeIDs.forall(node => node.Pool().listToSet().size() == Txs.size())

    // Gossip: The first time a node receives some transaction from a user, the
    // transaction eventually reaches all nodes (nodes may receive the
    // transaction more than once).
}
