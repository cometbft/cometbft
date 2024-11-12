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

    // Upon receiving a transaction message from a peer, the node attempts to
    // add the transaction to its mempool.
    action handleMessage(node, _msgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(_msgs, Some(sender), tx)
        }

    // A node attempting to add a transaction to its mempool processes the
    // transaction according to whether it has been seen it before, that is, if
    // it exists in the mempool cache. We call a *first-time* transaction to one
    // that the node does not have in its cache. A *duplicate* transaction is
    // one that the node has received multiple times, and thus it's in its
    // cache.
    //
    // The sender is optional. When there's a sender, it means that the
    // transaction comes from a peer; otherwise it comes directly from a user.
    action tryAddTx(node, _msgs, optionalSender, tx) = 
        if (not(hash(tx).in(node.Cache())))
            node.tryAddFirstTimeTx(_msgs, optionalSender, tx)
        else
            node.processDuplicateTx(_msgs, optionalSender, tx)

    // `tryAddFirstTimeTx` attempts to add a first-time transaction to the
    // mempool (not in cache): (1) it adds the transaction to the cache, (2) if
    // the transaction is valid, it adds the transaction to the pool and (3)
    // updates its senders.
    action tryAddFirstTimeTx(node, _msgs, optionalSender, tx) = all {
        state' = state.update(node, st => {
            cache: st.cache.join(hash(tx)),
            pool: if (valid(tx)) st.pool.append(tx) else st.pool,
            senders: 
                if (valid(tx)) 
                    st.senders.addSender(tx, optionalSender) 
                else st.senders,
            ...st }),
        msgs' = _msgs,
        peers' = peers,
    }

    // `processDuplicateTx` processes a duplicate transaction (in cache) by
    // updating the list of senders, only if the transaction is in pool.
    action processDuplicateTx(node, _msgs, optionalSender, tx) = all {
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
    // `mkTargetNodes` defines the set of peers to whom `node` will send `tx`.
    // It is passed as an argument to the generic transaction dissemination
    // action.
    //
    // In Flood, a node sends a transaction to all its peers except those who
    // previously sent it.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), List())
        node.Peers().exclude(txSenders.listToSet())

    //--------------------------------------------------------------------------
    // All possible state transitions in the protocol.
    action nextState = any {
        // Transaction dissemination: node sends transaction to subset of peers.
        nondet node = oneOf(nodesInNetwork)
        node.disseminateNextTx(mkTargetNodes, TxMsg),

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
        NodeIDs.forall(node => length(node.Msgs()) == 0)

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
