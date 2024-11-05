// -*- mode: Bluespec; -*-
module mempool {
    import spells.* from "./spells"
    import p2p.* from "./p2p"
    export p2p.*

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
    
    //--------------------------------------------------------------------------
    // Model parameters
    //--------------------------------------------------------------------------
        
    // The set of all possible transactions.
    const Txs: Set[TX]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    // The protocol is modelled as a state transition system. A state is defined
    // by the values of the following variables.

    type MempoolState = {         
        // Cache of already received transaction IDs.
        cache: Set[TxID],

        // List of uncommitted/pending transactions ("the mempool"). For picking
        // transactions to send to peers. 
        //
        // We do not model consensus removing mempool entries.
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
    var state: NodeID -> MempoolState

    // A node's empty state with a given set of peers.
    val emptyMempoolState = {
        pool: List(),
        cache: Set(),
        poolIndex: 0,
        senders: Map(),
    }

    // Auxiliary definitions
    def Cache(node) = state.get(node).cache
    def Pool(node) = state.get(node).pool
    def PoolIndex(node) = state.get(node).poolIndex
    def Senders(node) = state.get(node).senders

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state. All data structures are initially empty.
    action init = all {
        msgs' = NodeIDs.mapBy(_ => List()),
        peers' = NodeIDs.mapBy(n => InitialPeers.get(n)),
        state' = NodeIDs.mapBy(n => emptyMempoolState),
    }

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
    action receiveTxFromUser(node, tx, _tryAddTx) =
        node._tryAddTx(msgs, None, tx)

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
    action disseminateNextTx(node, mkTargetNodes, mkTxMsg) = all {
        // Check that the current index is within bounds. 
        require(node.PoolIndex() < node.Pool().length()),
        // Get from the pool the next transaction to disseminate.
        val tx = node.Pool()[node.PoolIndex()]
        all {
            // Send tx to all peers except the senders of tx.
            msgs' = node.send(msgs, mkTargetNodes(node, tx), mkTxMsg(tx)),
            peers' = peers,
            // Update pool index.
            state' = state.mapPut(node, st => { poolIndex: st.poolIndex + 1, ...st }),
        }
    }
}
