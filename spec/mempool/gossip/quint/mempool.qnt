// -*- mode: Bluespec; -*-

// File generated from markdown using lmt. DO NOT EDIT.

module mempool {
    import spells.* from "./spells"
    import p2p.* from "./p2p"
    export p2p.*

    //--------------------------------------------------------------------------
    // Types
    //--------------------------------------------------------------------------
    type TX = str
    pure def valid(tx) = true
    type TxID = str
    pure def hash(tx: TX): TxID = tx

    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    const Txs: Set[TX]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    var state: NodeID -> MempoolState
    
    type MempoolState = {
        cache: Set[TxID],
        pool: List[TX],
        poolIndex: int,
        senders: TxID -> List[NodeID]
    }
    
    // Auxiliary definitions
    pure def addSender(_txSenders, tx, optionalSender) = 
        match optionalSender {
        | Some(sender) => _txSenders.update(hash(tx), ss => 
            if (ss.includes(sender)) ss else ss.append(sender))
        | None => _txSenders
        }
    def sendersOf(node, tx) = 
        node.Senders().mapGetDefault(hash(tx), List()).listToSet()
    def Cache(node) = state.get(node).cache
    def Pool(node) = state.get(node).pool
    def PoolIndex(node) = state.get(node).poolIndex
    def Senders(node) = state.get(node).senders

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action init = all {
        P2P_init,
        state' = NodeIDs.mapBy(n => initialMempoolState),
    }
    val initialMempoolState = {
        pool: List(),
        cache: Set(),
        poolIndex: 0,
        senders: Map(),
    }
    action receiveTxFromUser(node, tx, _tryAddTx) =
        node._tryAddTx(incomingMsgs, None, tx)
    action disseminateNextTx(node, _mkTargetNodes, _mkTxMsg) = all {
        // Check that the current index is within bounds. 
        require(node.PoolIndex() < node.Pool().length()),
        // Get from the pool the next transaction to disseminate.
        val tx = node.Pool()[node.PoolIndex()]
        all {
            // Wrap tx in a message and send it to the target nodes.
            incomingMsgs' = 
                node.multiSend(incomingMsgs, _mkTargetNodes(node, tx), _mkTxMsg(tx)),
            // Increase pool index.
            state' = state.update(node, st => { poolIndex: st.poolIndex + 1, ...st }),
            peers' = peers,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    val uniqueTxsInPool = 
        NodeIDs.forall(node => size(node.Pool().listToSet()) == length(node.Pool()))

}
