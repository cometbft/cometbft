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

    // Upon receiving a tx message, try add tx to the pool.
    action handleMessage(node, updatedMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(updatedMsgs, Some(sender), tx)
        }

    // Try to add a (valid) transaction to the pool. The sender is optional,
    // meaning that if there's no sender, the transaction comes from a user. If
    // tx comes from a peer, the node needs to update the list of incoming
    // messages.
    //
    // A transaction that is in `pool`, must also be in `cache` (assuming an
    // infinite cache), but not necessarily the inverse. The reason a
    // transaction is in `cache` but not in `pool` is either because: 
    // - the transaction was initially invalid and never got into `pool`, 
    // - the transaction became invalid after it got in `pool` and thus got
    //   evicted while revalidating it, or
    // - the transaction was committed to a block and got removed from `pool`.
    action tryAddTx(node, updatedMsgs, senderOpt, tx) = 
        // If there's a sender, add it to tx's set of senders.
        val updatedTxSenders = node.addTxSenders(state, tx, senderOpt)
        if (not(node.Cache().contains(hash(tx)))) all {
            // Add tx to cache, add tx to pool if valid, and update senders.
            state' = state.mapPut(node, st => {
                cache: st.cache.setAdd(hash(tx)),
                pool: if (valid(tx)) st.pool.append(tx) else st.pool,
                senders: updatedTxSenders,
                ...st }),
            msgs' = updatedMsgs,
            peers' = peers,
        } else all {
            // Just update the senders.
            state' = state.mapPut(node, st => { senders: updatedTxSenders, ...st }),
            msgs' = updatedMsgs,
            peers' = peers,
        }

    // If senderOptional has a value, add it to tx's list of senders.
    pure def addTxSenders(node, _state, tx, senderOptional) = 
        val senders = _state.get(node).senders
        match senderOptional {
        | Some(sender) => senders.mapPut(hash(tx), ls => ls.setAdd(sender))
        | None => senders
        }

    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set())
        node.Peers().exclude(txSenders)

    //--------------------------------------------------------------------------
    // All possible state transitions.
    action nextState = 
        val nodes = peers.currentNodes()
        any {
            nondet node = oneOf(nodes)
            node.disseminateNextTx(mkTargetNodes, TxMsg),

            nondet node = oneOf(nodes)
            nondet tx = oneOf(Txs)
            node.receiveTxFromUser(tx, tryAddTx),

            nondet node = oneOf(nodes)
            node.receiveFromPeer(handleMessage),

            val otherNodes = NodeIDs.exclude(nodes) // nodes not connected to the network
            all {
                require(otherNodes.nonEmpty()),
                nondet node = oneOf(otherNodes)
                // Pick a non-empty set of nodes in the network to be the node's
                // peers.
                nondet peerSet = oneOf(nodes.powerset().exclude(Set()))
                node.join(peerSet),
                state' = state,
            },
            
            // Pick a node to disconnect only if it is not the only node in the
            // network.
            nondet node = oneOf(nodes)
            all {
                require(nodes != Set(node)),
                node.disconnect(),
                state' = state,
            },
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
