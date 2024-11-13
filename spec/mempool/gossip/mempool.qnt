// -*- mode: Bluespec; -*-
/*******************************************************************************
The mempool module defines essential types and data structures needed to keep a
list of pending transactions ("the mempool"), as well as generic actions to
disseminate transactions. Those generic actions are then instantiated with
specific functions that define the behaviour of the gossip protocols.
*******************************************************************************/
module mempool {
    import spells.* from "./spells"
    import p2p.* from "./p2p"
    export p2p.*

    //--------------------------------------------------------------------------
    // Types
    //--------------------------------------------------------------------------

    // A transaction is uniquely identified by a string, which represents its
    // content (typically implemented as an array of bytes).
    type TX = str

    // Transactions are validated by an external entity. The validation function
    // must be deterministic. In the actual implementation, the mempool makes a
    // CheckTx ABCI call to the application, which validates the transaction. 
    //
    // In this simplified specification we model all transactions as valid. To
    // model invalid transactions, `valid` should be declared as a model
    // parameter (a `const`) and instantiated with a deterministic function of
    // type `(TX) => bool`.
    pure def valid(tx) = true

    // A transaction identifier, computed as the hash of the transaction
    // (typically a short array of bytes).
    type TxID = str
    pure def hash(tx: TX): TxID = tx
    
    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
        
    // The set of all possible transactions.
    const Txs: Set[TX]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------

    // Mempool's data structures.
    type MempoolState = {         
        // Cache of already received transaction IDs. In this spec we assume the
        // cache never overflows, i.e., it can grow indefinitely.
        cache: Set[TxID],

        // List of uncommitted or pending transactions ("the mempool"). Used for
        // picking transactions to disseminate to peers.
        //
        // We make the following assumptions about the mempool:
        // - It does not have a maximum capacity.
        // - New entries are only appended. We do not model when entries are
        //   removed.
        //
        // A transaction that is in `pool`, must also be in `cache` (assuming an
        // infinite cache), but not necessarily the inverse. The reason a
        // transaction is in `cache` but not in `pool` is either because: 
        // - the transaction was initially invalid and never got into `pool`, 
        // - the transaction became invalid after it got in `pool` and thus got
        //   evicted when it was revalidated, or
        // - the transaction was committed to a block and got removed from
        //   `pool`.
        //
        // All these scenarios are not modeled here. Then `cache` and `pool`
        // will always have the same content and one of the two is actually
        // redundant in this spec.
        pool: List[TX],

        // Index to the next transaction in pool to disseminate. A pool iterator
        // traverses the entries in pool one at a time.
        //
        // We model transaction dissemination using one dissemination process
        // (`disseminateNextTx`) that iterates on the list of transactions
        // reading one mempool entry per step, and atomically multicasts one
        // transaction message to all connected peers.
        //
        // In the implementation there is one dissemination process per peer,
        // each with its own iterator (and thus a separate pool index per
        // iterator) with a `next()` method to retrieve the next entry in the
        // list. If it reaches the end of the list, it blocks until a new entry
        // is added. All iterators read concurrently from the pool.
        poolIndex: int,

        // Each entry in the mempool has a set of peers from which the node
        // received the transaction. 
        //
        // We define it as a list instead of a set because the DOG protocol
        // needs to know who is the first sender of a transaction.
        //
        // Note that when a transaction is in the cache but not in the pool, it
        // won't have any sender. Senders are only needed for disseminating
        // (valid) transactions that are in the mempool.
        senders: TxID -> List[NodeID]
    }

    // Each node has a mempool state.
    var state: NodeID -> MempoolState

    // A node's empty state with a given set of peers.
    val emptyMempoolState = {
        pool: List(),
        cache: Set(),
        poolIndex: 0,
        senders: Map(),
    }

    // `addSender` adds a sender to `tx`'s list of senders (`_txSenders`), if
    // `optionalSender` has a value.
    pure def addSender(_txSenders, tx, optionalSender) = 
        match optionalSender {
        | Some(sender) => _txSenders.update(hash(tx), ss => ss.append(sender))
        | None => _txSenders
        }

    // Auxiliary definitions
    def Cache(node) = state.get(node).cache
    def Pool(node) = state.get(node).pool
    def PoolIndex(node) = state.get(node).poolIndex
    def Senders(node) = state.get(node).senders

    // The set of senders of transaction `tx`.
    def sendersOf(node, tx) = 
        node.Senders().mapGetDefault(hash(tx), List()).listToSet()

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // Initial state.
    action init = all {
        msgs' = NodeIDs.mapBy(_ => List()),
        peers' = NodeIDs.mapBy(n => InitialPeers.get(n)),
        state' = NodeIDs.mapBy(n => emptyMempoolState),
    }

    // Users create transactions and send them to one of the nodes in the
    // network. Nodes receive transactions either directly from users or in
    // messages from peers. Transaction from users have no sender.

    // `receiveTxFromUser` is a generic action that models a node receiving
    // transaction `tx` from a user. The function `_tryAddTx` can be
    // instantiated by the gossip protocol to define how to add transactions to
    // the mempool:
    // - `_tryAddTx(msgs, optionalSender, tx)` attempts to add `tx` to the
    //   mempool, given the node ID of the sender in case `tx` comes from
    //   another node.
    //
    // Typically, users send (full) transactions to the node via an RPC
    // endpoint. Users are allowed to submit the same transaction more than once
    // and to multiple nodes.
    //
    // This action is enabled only if the transaction is not in the pool. In the
    // actual mempool implementation we have the cache that prevents this
    // scenario.
    action receiveTxFromUser(node, tx, _tryAddTx) =
        node._tryAddTx(msgs, None, tx)

    // `disseminateNextTx` is a generic action that models a node traversing the
    // pool while sending transactions to its peers. It takes the next
    // transaction from the iterator and atomically sends it to a set of target
    // peers.
    //
    // The following arguments are functions that define the to who it will send
    // transactions:
    // - `_mkTargetNodes(node, tx)` returns the set of peers to which `node`
    //   will send `tx`.
    // - `_mkTxMsg(tx)` is a wrapper function that returns the specific message
    //   type used by the gossip protocol.
    //
    // The pool index must not exceed the pool's length. This pre-condition
    // models when the iterator is at the end of the list and it's blocked
    // waiting for a new entry to be appended to the list.
    //
    // In the actual implementation, there is a separate goroutine for each
    // peer, so not all txs are sent at the same time.
    action disseminateNextTx(node, _mkTargetNodes, _mkTxMsg) = all {
        // Check that the current index is within bounds. 
        require(node.PoolIndex() < node.Pool().length()),
        // Get from the pool the next transaction to disseminate.
        val tx = node.Pool()[node.PoolIndex()]
        all {
            // Wrap tx in a message and send it to the target nodes.
            msgs' = node.send(msgs, _mkTargetNodes(node, tx), _mkTxMsg(tx)),
            // Increase pool index.
            state' = state.update(node, st => { poolIndex: st.poolIndex + 1, ...st }),
            peers' = peers,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------

    // Invariant: pools do not have repeated entries.
    val uniqueTxsInPool = 
        NodeIDs.forall(node => size(node.Pool().listToSet()) == length(node.Pool()))
    
}
