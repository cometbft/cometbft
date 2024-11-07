// -*- mode: Bluespec; -*-

/*******************************************************************************
Flood is a basic _push_ gossip protocol: every time a node receives a
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

    // Nodes communicates only one type of message carrying a full transaction.
    type Message =
        | TxMsg(TX)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Upon receiving a transaction message from a peer, the node attempts to
    // add the transaction to the mempool.
    action handleMessage(node, _msgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_msgs, Some(sender), tx)
        }

    // A node attempts to add a transaction to the mempool. It will proceed
    // differently if it has seen before the transaction or not. 
    //
    // The sender is optional, meaning that if there's no sender, the
    // transaction comes from a user. If tx comes from a peer, the node needs to
    // update the list of incoming messages to remove the message from sender.
    action tryAddTx(node, _msgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_msgs, optionalSender, tx)
        else
            node.processDuplicateTx(_msgs, optionalSender, tx)

    // Attempt to add a first-time transaction tx to the mempool: add tx to
    // cache, add tx to pool and update tx's senders if tx is valid, and update
    // messages.
    action tryAddFirstTimeTx(node, _msgs, optionalSender, tx) = all {
        state' = state.update(node, st => {
            cache: st.cache.setAdd(hash(tx)),
            pool: if (valid(tx)) st.pool.append(tx) else st.pool,
            senders: 
                if (valid(tx)) 
                    st.senders.addSender(tx, optionalSender) 
                else st.senders,
            ...st }),
        msgs' = _msgs,
        peers' = peers,
    }

    // Process a duplicate transaction tx: update tx's senders if tx is in pool,
    // and update messages.
    action processDuplicateTx(node, _msgs, optionalSender, tx) =  all {
        state' = state.update(node, st => { 
            senders: 
                if (st.pool.includes(tx)) 
                    st.senders.addSender(tx, optionalSender) 
                else st.senders, 
            ...st }),
        msgs' = _msgs,
        peers' = peers,
    }

    //--------------------------------------------------------------------------
    // mkTargetNodes is used by the transaction dissemination process to define
    // the set of peers to which `node` will send `tx`.
    //
    // For Flood, we want to send tx to all peers except tx's senders.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), List())
        node.Peers().exclude(txSenders.listToSet())

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

    // All messages have been delivered.
    val allMsgsDelivered = 
        NodeIDs.forall(node => node.Msgs().listIsEmpty())

    // The pools of all nodes contain a given tx.
    def txInAllPools(tx) =
        NodeIDs.forall(n => n.Pool().includes(tx))

    // All available transactions are in the pool of all nodes.
    val allTxsGossipped =
        NodeIDs.forall(node => size(node.Pool().listToSet()) == size(Txs))

    // Gossip: The first time a node receives some transaction from a user, the
    // transaction eventually reaches all nodes (nodes may receive the
    // transaction more than once).
}
