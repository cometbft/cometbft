// -*- mode: Bluespec; -*-

/*******************************************************************************
Flood is a simple *push* gossip protocol: every time a node receives a
transaction, it forwards (or pushes) the transaction to all of its peers, except
to the peer(s) from which it received the transaction.
*******************************************************************************/
module flood {
    import spells.* from "./spells"

    //--------------------------------------------------------------------------
    // Base types
    //--------------------------------------------------------------------------

    // A transaction is uniquely identified by a string, which represents its
    // content (typically implemented as an array of bytes).
    type TX = str

    // Transactions are validated by an external entity. The validation function
    // must be deterministic. In the implementation, it is the application via a
    // CheckTx ABCI call that validates transctions. In this simplified spec, we
    // model all transactions as valid.
    pure def valid(tx) = true

    // A transaction identifier, computed as the hash of the transaction
    // (typically a short array of bytes).
    type TxID = str
    pure def hash(tx: TX): TxID = tx
    
    // A node internally identifies other nodes with an ID (a string).
    type NodeID = str    

    //--------------------------------------------------------------------------
    // Model parameters
    //--------------------------------------------------------------------------
        
    // The set of all possible transactions.
    const Txs: Set[TX]
    
    // The set of all nodes that participate in the network, even those that are
    // not initially connected.
    const NodeIDs: Set[NodeID]

    // Initial network topology. The topology is defined by the set of peers
    // each node has (the values of the mapping).
    const InitialPeers: NodeID -> Set[NodeID]

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------

    // In this protocol, nodes communicates only one type of message, `TxMsg`,
    // that includes a full transaction.
    type Message =
        | TxMsg(TX)

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    // The protocol is modelled as a state transition system. A state is defined
    // by the values of the following variables.

    // Each node has a list of incoming messages. Messages arrive in order and
    // are appended to the list. Each message has a sender.
    var msgs: NodeID -> List[(NodeID, Message)]

    type NodeState = { 
        // The dynamic network topology. Each node has a set of peers.
        peers: Set[NodeID],
        
        // Cache of already received transaction IDs.
        cache: Set[TxID],

        // List of uncommitted/pending transactions ("the mempool"). For
        // checking if the node already received a transaction and for picking
        // transactions to send to peers.
        pool: List[TX],

        // Index of the last broadcasted transaction in the pool.
        //
        // In this specification, we model transaction dissemination using one
        // iterator. One process iterates on the pool and multicasts a mempool
        // entry to all peers, one per step and atomically.
        //
        // In the real implementation there is one dissemination process per
        // peer, each with its own iterator (and thus a separate pool index per
        // iterator) with a `next()` method to retrieve the next entry in the
        // list. If it reaches the end of the list, it blocks until a new entry
        // is added. All iterators read concurrently from the pool.
        poolIndex: int,

        // Each transaction in the pool has a set of peers from which the node
        // received the transaction. Note that a transaction may be in the cache
        // but not in the pool. In that case, the transaction won't have
        // senders.
        //
        // Senders are only needed for disseminating (valid) transactions that
        // are in the mempool. That is why we remove the senders when the
        // transaction is removed from the mempool (not modelled here).
        senders: TxID -> Set[NodeID]
    }

    // The state of all nodes.
    var state: NodeID -> NodeState

    // A node's empty state with a given set of peers.
    pure def emptyNodeState(peers) = {
        peers: peers,
        pool: List(),
        cache: Set(),
        poolIndex: 0,
        senders: Map(),
    }

    // Auxiliary definitions
    def Msgs(node) = msgs.get(node)
    def Peers(node) = state.get(node).peers
    def Cache(node) = state.get(node).cache
    def Pool(node) = state.get(node).pool
    def PoolIndex(node) = state.get(node).poolIndex
    def Senders(node) = state.get(node).senders

    //--------------------------------------------------------------------------
    // Network boilerplate and auxiliary definitions
    //--------------------------------------------------------------------------
    // We assume that nodes have reliable P2P channels to all other nodes and
    // that messages are delivered in the order in which they were sent.
    //
    // In the implementation, messages are transamitted via the `Mempool`
    // channel on the P2P layer.

    // A node sends a message to the given set of peers.
    pure def send(node, _msgs, peers, msg) =
        _msgs.mapPutMany(peers, p => _msgs.get(p).append((node, msg)))

    // A node receives one of the incoming messages from a peer and handles it
    // according to its type.
    action receiveFromPeer(node) = all {
        require(node.Msgs().listNonEmpty()),
        val someMsg = node.Msgs().head()
        val sender = someMsg._1
        val msg = someMsg._2
        // We model receiving of a message as taking the head of the list of
        // incoming messages.
        val updatedMsgs = msgs.mapPut(node, ms => tail(ms))
        handleMessage(node, updatedMsgs, sender, msg)
    }

    // Nodes connected to the network in the current state.
    pure def currentNodes(_state) = 
        _state.keys().map(n => _state.get(n).peers).flatten()

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state. All data structures are initially empty.
    action init = all {
        msgs' = NodeIDs.mapBy(_ => List()),
        state' = NodeIDs.mapBy(n => emptyNodeState(InitialPeers.get(n))),
    }

    //--------------------------------------------------------------------------
    // Transactions are created by users who send them to one of the nodes in
    // the network. Nodes receive transactions either from users or from peers.
    // Transaction messages sent from users have no sender.

    // A node receives a given transaction from a user.
    //
    // Typically a user sends a (full) transaction to a node via an RPC
    // endpoint. Users are allowed to submit the same tx more than once and to
    // multiple nodes. This action is only enabled if the transaction is not in
    // the pool. In the current mempool implementation we have the cache that
    // prevents this scenario.
    action receiveTxFromUser(node, tx) = 
        node.tryAddTx(msgs, None, tx)

    // Handle valid messages for this protocol.
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
        val updatedTxSenders = 
            match senderOpt {
            | Some(sender) => node.Senders().mapPut(hash(tx), ls => ls.setAdd(sender))
            | None => node.Senders()
            }
        if (not(node.Cache().contains(hash(tx)))) all {
            // Add tx to cache, add tx to pool if valid, and update senders.
            state' = state.mapPut(node, st => {
                cache: st.cache.setAdd(hash(tx)),
                pool: if (valid(tx)) st.pool.append(tx) else st.pool,
                senders: updatedTxSenders,
                ...st }),
            msgs' = updatedMsgs,
        } else all {
            // Just update the senders.
            state' = state.mapPut(node, st => { senders: updatedTxSenders, ...st }),
            msgs' = updatedMsgs,
        }

    //--------------------------------------------------------------------------
    // A node iterates on the pool sending transactions to its peers. It takes
    // the next transaction from the iterator and disseminates it to all its
    // peers at once. 
    //
    // The pool index must not exceed the pool's length. This models when the
    // iterator is at the end of the list and it's blocked waiting for a new
    // entry to be appended to the list.
    //
    // In the current implementation, there is a separate goroutine for each
    // peer, so not all txs are sent at the same time.
    action disseminateNextTx(node) = all {
        // Check that the current index is within bounds. 
        require(node.PoolIndex() < node.Pool().length()),
        // Get from the pool the next transaction to disseminate.
        val tx = node.Pool()[node.PoolIndex()]
        all {
            // Send tx to all peers except the senders of tx.
            val txSenders = node.Senders().mapGetDefault(hash(tx), Set())
            msgs' = node.send(msgs, node.Peers().exclude(txSenders), TxMsg(tx)),
            // Update pool index.
            state' = state.mapPut(node, st => { poolIndex: st.poolIndex + 1, ...st }),
        }
    }

    //--------------------------------------------------------------------------
    // A node joins the network by connecting to a given set of peers. All its
    // peers add the new node to their list of peers.
    action joinNetwork(node: NodeID, newPeers: Set[NodeID]): bool = all {
        // The node must not be connected to the network.
        require(node.Peers().nonEmpty()),
        state' = state
            // Set an empty state with the new peers to the new node.
            .mapPut(node, st => emptyNodeState(newPeers))
            // Add node as a new peer to the nodes currently in the network.
            .mapPutMany(state.currentNodes(), n => { 
                peers: n.Peers().setAdd(node), 
                ...state.get(n) }),
        msgs' = msgs,
    }

    // A node gets disconnected from the network. All its peers are immediately
    // aware that the node is no longer one of their peers, so their state is
    // updated accordingly.
    //
    // TODO: the network must not become disconnected; we don't want to model
    // that.
    action disconnectNetwork(node) = all {
        state' = state
            // Clean node's state and remove all its peers.
            .mapPut(node, _ => emptyNodeState(Set()))
            // Remove node from other peers' state.
            .mapPutMany(state.currentNodes(), n => {
                peers: n.Peers().except(node),
                ...state.get(n) }),
        msgs' = msgs,
    }

    //--------------------------------------------------------------------------
    // All possible state transitions.
    action nextState = 
        val nodes = state.currentNodes()
        any {
            nondet node = oneOf(nodes)
            node.disseminateNextTx(),

            nondet node = oneOf(nodes)
            nondet tx = oneOf(Txs)
            node.receiveTxFromUser(tx),

            nondet node = oneOf(nodes)
            node.receiveFromPeer(),

            val otherNodes = NodeIDs.exclude(nodes) // nodes not connected to the network
            all {
                require(otherNodes.nonEmpty()),
                nondet node = oneOf(otherNodes)
                // Pick a non-empty set of nodes in the network to be the node's
                // peers.
                nondet peers = oneOf(nodes.powerset().exclude(Set()))
                node.joinNetwork(peers),
            },
            
            // Pick a node to disconnect only if it is not the only node in the
            // network.
            nondet node = oneOf(nodes)
            all {
                require(nodes != Set(node)),
                node.disconnectNetwork()
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
