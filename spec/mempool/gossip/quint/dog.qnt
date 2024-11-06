// -*- mode: Bluespec; -*-

// 👉 This is WORK IN PROGRESS; not everything here is correct 👈 

/*******************************************************************************
Dynamic Optimal Graph (DOG) gossip protocol.

The DOG protocol is built on top of the Flood protocol. All types, messages, and
data structures in Flood are also present in this specification.

Briefly, if node A receives from B a transaction that it already has, this means
that there must exist a cycle in the network topology, so A tells B to stop
sending transactions. So the idea is simply to cut cycles.

The protocol adds two new messages to Flood: 
- HaveTxMsg(TxID) for cutting cycles, and 
- ResetMsg for dynamic re-routing when a node disconnects (more on this later).

HaveTxMsg rules:
- H1. If node A receives a transaction tx from node B that it already received
  (that is, it's in the cache), then A replies to B with a message
  HaveTxMsg(tx).
    - HaveTxMsg messages contains a transaction hash, so its size is
      insignificant. 
- H2. If node B receives a HaveTxMsg(tx) message from node A, it disables the
  route sender(tx) → A. 
    - A route in a node is a tuple `sourceNodeID → targetNodeID`, with source
      and target being any of the node's peers. Initially, all nodes have all
      their routes enabled.
    - We don't want to cut all routes from B to A, only those that come from the
      transaction's original sender, that's why we need to take the source into
      account.
    - The protocol implicitly favours routes with low latency, by cutting routes
      to peers that send the duplicate tx at a later time.
- H3. Before node A sends a transaction tx to B, check if the route sender(tx) →
  B is enabled. 
    - Txs received from users are always gossiped to all peers.
- H4. After node A sends a HaveTxMsg message to B, start a timer on A and don't
  let A send any other HaveTxMsg message while the timer is on. 
    - This is to avoid a specific interleaving of messages that could render
      node A isolated from gossip traffic.
    - While the timer is on, we allow incoming traffic that flows from B to
      gradually stop. During this time, the traffic is diverted to other routes
      from other peers, while A still receives a small number of duplicate
      transactions (it's visible in the picture above).
    - The suggested minimum timeout should be the time it takes
      already-dispatched messages from B to arrive to A, so it's enough for the
      traffic from B to fully stop. We don't care to receive a few more
      duplicate transactions, therefore the timer could be safely extended to be
      sure that A does not become isolated.

ResetMsg rules:
- R1. If node A detects that its peer B is disconnected, remove any route that
  has B as source or target, and broadcast to all other peers a message
  ResetMsg.
    - This is to signal A's peers that A's situation has changed and its routing
      data should be reset so it can be rerouted if needed.
- R2. On receiving a ResetMsg message from A, remove any route that has A as
  source or target. 
    - This will allow traffic to flow again to A and nodes will dynamically
      adapt to the new traffic, closing routes when needed.

*******************************************************************************/
module dog {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    import mempool as M from "./mempool"

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    type Message =
        | TxMsg(TX)
        // HaveTxMsg messages carry a transaction hash, which is typically much
        // smaller than full transactions.
        // "I already have this tx; don't send me more txs from the same
        // source."
        | HaveTxMsg(TX)
        // "My situation has changed; reset my routing data on your side."
        | ResetMsg

    //--------------------------------------------------------------------------
    // Routing
    //--------------------------------------------------------------------------

    // A route is a tuple `(source, target)`. We also write it as `A -> B`.
    type Route = (NodeID, NodeID)

    // The set of disabled routes on each node.
    var disabledRoutes: NodeID -> Set[Route]

    def DisabledRoutes(node) = disabledRoutes.get(node)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state. All data structures are initially empty.
    action D_init = all {
        M::init,
        disabledRoutes' = NodeIDs.mapBy(node => Set()),
    }

    // Handle valid messages for this protocol.
    action handleMessage(node, updatedMsgs, sender, msg) =
        match msg {
        | TxMsg(tx) => node.tryAddTx(updatedMsgs, Some(sender), tx)
        | HaveTxMsg(tx) => node.handleHaveTxMessage(updatedMsgs, sender, tx)
        | ResetMsg => node.handleResetMessage(updatedMsgs, sender)
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
            disabledRoutes' = disabledRoutes,
        } else all {
            // Reply HaveTxMsg(tx)
            //
            // TODO: send HaveTxMsg message only if there's no incoming message
            //       from sender(tx) and sender(tx) has the route to this node
            //       disabled. This is to simulate the timeout that disables
            //       sending HaveTxMsg.
            msgs' = match senderOpt {
                | Some(sender) => node.send(updatedMsgs, Set(sender), HaveTxMsg(tx))
                | None => updatedMsgs
                },
            state' = state.mapPut(node, st => { senders: updatedTxSenders, ...st }),
            peers' = peers,
            disabledRoutes' = disabledRoutes,
        }

    // If senderOptional has a value, add it to tx's list of senders.
    pure def addTxSenders(node, _state, tx, senderOptional) = 
        val senders = _state.get(node).senders
        match senderOptional {
        | Some(sender) => senders.mapPut(hash(tx), ls => ls.setAdd(sender))
        | None => senders
        }

    // Disable the route (sender(tx) -> sender).
    pure def disable(_disabledRoutes, node, txSenders, sender) =
        if (not(txSenders.listIsEmpty()))
            _disabledRoutes.mapPut(node, rs => rs.setAdd((txSenders[0], sender))) // TODO: check
        else _disabledRoutes 

    // Handle a received HaveTx message.
    // Disable the route (sender(tx) -> sender).
    action handleHaveTxMessage(node, updatedMsgs, sender, tx) = all {
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set()) // in most cases, there exists one and only one sender for tx (??)
        disabledRoutes' = disabledRoutes.disable(node, txSenders.setToList(), sender), // TODO: check setToList
        msgs' = updatedMsgs,
        peers' = peers,
        state' = state,
    }

    // Handle a received Reset message.
    // Remove any route that has sender as source or target. 
    action handleResetMessage(node, updatedMsgs, sender) = all {
        disabledRoutes' = disabledRoutes.remove(node, sender),
        msgs' = updatedMsgs,
        peers' = peers,
        state' = state,
    }

    // Remove from routes any route that has sender as source or target. 
    pure def remove(routes, node, sender) = 
        routes.mapPut(node, rs => rs.filter(route => not(sender.isSourceOrTargetIn(route))))

    // Is the node a source or target in the given route.
    pure def isSourceOrTargetIn(node, route) =
        node == route._1 or node == route._2

    // TODO: a malicious node sends the same tx to one or more of its peers,
    // maybe in coordination with another malicious node.
    // action gossipDuplicate(node) = ...

    //--------------------------------------------------------------------------
    // Send tx to all peers except if the route `sender(tx) -> peer` is
    // disabled.
    //
    // Broadcast tx to all the peers of node, except if peer is sender(tx) or if
    // the route sender(tx) -> peer is disabled, where sender(tx) is any of the
    // peers that have sent tx to node.
    def mkTargetNodes(node, tx) =
        val txSenders = node.Senders().mapGetDefault(hash(tx), Set())
        val disabledTargets = node.DisabledRoutes().filter(r => txSenders.contains(r._1)).map(r => r._2)
        // TODO: remove exclude(txSenders)?
        node.Peers().exclude(txSenders).exclude(disabledTargets)

    //--------------------------------------------------------------------------
    // All possible state transitions.
    action nextState = 
        val nodes = peers.currentNodes()
        any {
            nondet node = oneOf(nodes)
            all {
                node.disseminateNextTx(mkTargetNodes, TxMsg),
                disabledRoutes' = disabledRoutes,
            },

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
                disabledRoutes' = disabledRoutes,
            },
            
            // Pick a node to disconnect only if it is not the only node in the
            // network.
            nondet node = oneOf(nodes)
            all {
                require(nodes != Set(node)),
                node.disconnect(),
                state' = state,
                disabledRoutes' = disabledRoutes,
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
