// -*- mode: Bluespec; -*-
module state {
    import spells.* from "./spells"
    import network.* from "./network"
    export network.*

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

    type NodeState = {         
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
    val emptyNodeState = {
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

}
