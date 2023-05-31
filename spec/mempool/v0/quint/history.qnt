// -*- mode: Bluespec; -*-
module history {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"

    // For each node, all transactions ever added to the mempool.
    var historicMempoolTxs: NodeId -> Set[Tx]

    // For each node, all transactions ever received from a client via RPC.
    var txsReceivedViaRPC: NodeId -> Set[Tx]

    // For each node, all transactions ever gossiped to its peers.
    var gossippedTxs: NodeId -> Set[Tx]

    //--------------------------------------------------------------------------

    def _MempoolTxs(node) = historicMempoolTxs.get(node)
    def _ReceivedTxs(node) = txsReceivedViaRPC.get(node)
    def _GossippedTxs(node) = gossippedTxs.get(node)

    // A node has or has had tx in its mempool.
    def hadInMempool(node, tx) = tx.in(node._MempoolTxs())

    // A node has received tx submitted by a client via RPC.
    def receivedViaRPC(node, tx) = tx.in(node._ReceivedTxs())

    // A node has gossipped tx to a peer.
    def gossipped(node, tx) = tx.in(node._GossippedTxs())

    //--------------------------------------------------------------------------
    action init = all {
        historicMempoolTxs' = NodeIds.mapBy(_ => Set()),
        txsReceivedViaRPC' = NodeIds.mapBy(_ => Set()),
        gossippedTxs' = NodeIds.mapBy(_ => Set()),
    }

    action recordMempoolTx(nodeId, tx) = all {
        historicMempoolTxs' = historicMempoolTxs.mapPut(nodeId, (txs) => txs.setAdd(tx)),
        txsReceivedViaRPC' = txsReceivedViaRPC,
        gossippedTxs' = gossippedTxs,
    }

    action recordSubmittedTx(nodeId, tx) = all {
        historicMempoolTxs' = historicMempoolTxs,
        txsReceivedViaRPC' = txsReceivedViaRPC.mapPut(nodeId, (txs) => txs.setAdd(tx)),
        gossippedTxs' = gossippedTxs,
    }

    action recordGossippedTx(nodeId, tx) = all {
        historicMempoolTxs' = historicMempoolTxs,
        txsReceivedViaRPC' = txsReceivedViaRPC,
        gossippedTxs' = gossippedTxs.mapPut(nodeId, (txs) => txs.setAdd(tx)),
    }

    action unchanged = all {
        historicMempoolTxs' = historicMempoolTxs,
        txsReceivedViaRPC' = txsReceivedViaRPC,
        gossippedTxs' = gossippedTxs,
    }

    //--------------------------------------------------------------------------

    // for debugging in repl
    def history = (
        ("historicMempoolTxs", historicMempoolTxs),
        ("txsReceivedViaRPC", txsReceivedViaRPC),
        ("gossippedTxs", gossippedTxs),
    )

}
