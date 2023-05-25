// -*- mode: Bluespec; -*-
module history {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"

    // For each node, all transactions ever added to the mempool.
    var historyMempoolTxs: NodeId -> Set[Tx]

    // For each node, all transactions ever submitted by a client via RPC.
    var historySubmittedTxs: NodeId -> Set[Tx]

    //--------------------------------------------------------------------------
    action History_init = all {
        historyMempoolTxs' = NodeIds.mapBy(_ => Set()),
        historySubmittedTxs' = NodeIds.mapBy(_ => Set()),
    }

    action History_recordMempoolTx(nodeId: NodeId, tx: Tx): bool = all {
        historyMempoolTxs' = historyMempoolTxs.put(nodeId, historyMempoolTxs.get(nodeId).setAdd(tx)),
        historySubmittedTxs' = historySubmittedTxs,
    }

    action History_recordSubmittedTx(nodeId, tx) = all {
        historyMempoolTxs' = historyMempoolTxs,
        historySubmittedTxs' = historySubmittedTxs.put(nodeId, historySubmittedTxs.get(nodeId).setAdd(tx)),
    }

    action History_unchanged = all {
        historyMempoolTxs' = historyMempoolTxs,
        historySubmittedTxs' = historySubmittedTxs,
    }

    //--------------------------------------------------------------------------

    // for debugging in repl
    def history = (
        ("MempoolTxs", historyMempoolTxs),
        ("SubmittedTxs", historySubmittedTxs),
    )

}