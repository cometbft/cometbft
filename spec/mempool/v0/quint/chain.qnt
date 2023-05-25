// -*- mode: Bluespec; -*-
module chain {
    import base.* from "./base"

    var chain: List[Set[Tx]]

    val Chain_latestHeight = 
        chain.length() - 1

    val Chain_isEmpty = 
        chain.length() == 0

    def Chain_getBlock(h) =
        chain[h]

    val Chain_allTxsInChain =
        chain.foldl(Set(), (s, x) => s.union(x))

    //--------------------------------------------------------------------------
    action Chain_Init =
        chain' = List()

    action Chain_NewBlockFrom(txs) =
        chain' = chain.append(txs)

    action Chain_Unchanged = 
        chain' = chain
}
