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

    // def Chain_inChain(tx, h) = ...

    //--------------------------------------------------------------------------
    action Chain_init =
        chain' = List()

    action Chain_newBlockFrom(txs) =
        chain' = chain.append(txs)

    action Chain_unchanged = 
        chain' = chain
}
