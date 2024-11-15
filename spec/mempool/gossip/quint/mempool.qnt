// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

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
    const AllTxs: Set[TX]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    var mempool: NodeID -> MempoolState
    
    type MempoolState = {
        cache: Set[TxID],
        txs: List[TX],
        txsIndex: int,
    }
    
    // Auxiliary definitions
    def Cache(node) = mempool.get(node).cache
    def Txs(node) = mempool.get(node).txs
    def TxsIndex(node) = mempool.get(node).txsIndex

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action MP_init = all {
        P2P_init,
        mempool' = NodeIDs.mapBy(n => initialMempoolState),
    }
    val initialMempoolState = {
        cache: Set(),
        txs: List(),
        txsIndex: 0,
    }
    action receiveTxFromUser(node, tx, _tryAddTx) =
        node._tryAddTx(incomingMsgs, None, tx)
    action disseminateNextTx(node, _mkTargetNodes, _mkTxMsg) = all {
        // Check that the current index is within bounds. 
        require(node.TxsIndex() < node.Txs().length()),
        // Get from the mempool the next transaction to disseminate.
        val tx = node.Txs()[node.TxsIndex()]
        all {
            // Wrap transaction in a message and send it to the target nodes.
            incomingMsgs' = 
                node.multiSend(incomingMsgs, _mkTargetNodes(node, tx), _mkTxMsg(tx)),
            // Increase index.
            mempool' = mempool.update(node, st => { txsIndex: st.txsIndex + 1, ...st }),
            peers' = peers,
        }
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    val uniqueTxsInMempool = 
        NodeIDs.forall(node => size(node.Txs().listToSet()) == length(node.Txs()))

}
