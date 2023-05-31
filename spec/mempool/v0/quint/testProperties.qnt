// Test properties that define scenarios for extracting traces
module testProperties {
    import base.* from "./base"
    import chain as Chain from "./chain"
    import mempoolv0.* from "./mempoolv0"

    //--------------------------------------------------------------------------
    // a node whose mempool is not empty
    val mempoolNonEmpty = 
        nondet node = oneOf(NodeIds)
        node.mempoolTxs().nonEmpty()
    val notMempoolNonEmpty = not(mempoolNonEmpty)

    //--------------------------------------------------------------------------
    // a node whose cache has tx that is not in the mempool
    val InCacheNotInMempool = 
        nondet node = oneOf(NodeIds)
        all {
            node.Cache().nonEmpty(),
            nondet tx = oneOf(node.Cache())
            tx.notIn(node.mempoolTxs())
        }
    val notInCacheNotInMempool = not(InCacheNotInMempool)

    //--------------------------------------------------------------------------
    // a node whose cache has tx that is also in the mempool
    val InCacheInMempool = 
        nondet node = oneOf(NodeIds)
        all {
            node.Cache().nonEmpty(),
            nondet tx = oneOf(node.Cache())
            tx.in(node.mempoolTxs())
        }
    val notInCacheInMempool = not(InCacheInMempool)

    //--------------------------------------------------------------------------
    // two different nodes have the same tx in mempool
    val SameTx = 
        nondet tx = oneOf(Txs)
        nondet n1 = oneOf(NodeIds)
        nondet n2 = oneOf(NodeIds)
        all {
            n1 != n2,
            tx.in(n1.mempoolTxs()),
            tx.in(n2.mempoolTxs()),
        }
    val notSameTx = not(SameTx)

    //--------------------------------------------------------------------------
    val EmptyChain = Chain::chainIsEmpty

    val FullChain = Chain::allTxsInChain == Set("tx1", "tx2")
    val notFullChain = not(FullChain)    
}
