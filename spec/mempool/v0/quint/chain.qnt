// -*- mode: Bluespec; -*-
module chain {
    import base.* from "./base"

    type Chain = List[Set[Tx]]
    var chain: Chain

    val latestHeight = 
        chain.length()

    val chainIsEmpty = 
        chain.length() == 0

    def getBlock(h) =
        chain[h]

    val allTxsInChain =
        chain.foldl(Set(), (s, x) => s.union(x))

    //--------------------------------------------------------------------------
    action init =
        chain' = List()

    action newBlockFrom(txs) =
        chain' = chain.append(txs)

    action unchanged = 
        chain' = chain

    //--------------------------------------------------------------------------
    // All transactions in the chain up to height `h`.
    def txsUpTo(ch: Chain, h: Height): Set[Tx] = 
        if (ch.length() == 0) 
            Set()
        else
            ch.slice(0, h).listToSet().flatten()

    run txsUpToTest1 = init
        .then(all {
            assert(chain.txsUpTo(0) == Set("")),
            unchanged,
        })

    run txsUpToTest = txsUpToTest1
        .then(newBlockFrom(Set("tx1")))
        .then(newBlockFrom(Set("tx2", "tx3")))
        .then(all {
            assert(chain.txsUpTo(1) == Set("tx1")),
            assert(chain.txsUpTo(2) == Set("tx1", "tx2", "tx3")),
            unchanged,
        })

    // The transaction `tx` is committed to the chain, at most at height `h`.
    def committedUpTo(tx: Tx, h: Height): bool = 
        tx.in(chain.txsUpTo(h))

}
