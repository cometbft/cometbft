// -*- mode: Bluespec; -*-
module chain {
    import base.* from "./base"

    type Chain = List[Set[Tx]]
    var chain: Chain

    val Chain_latestHeight = 
        chain.length() - 1

    val Chain_isEmpty = 
        chain.length() == 0

    def Chain_getBlock(h) =
        chain[h]

    val Chain_allTxsInChain =
        chain.foldl(Set(), (s, x) => s.union(x))

    //--------------------------------------------------------------------------
    action Chain_init =
        chain' = List()

    action Chain_newBlockFrom(txs) =
        chain' = chain.append(txs)

    action Chain_unchanged = 
        chain' = chain

    //--------------------------------------------------------------------------
    // All transactions in the chain up to height `h`.
    def txsUpTo(ch: Chain, h: Height): Set[Tx] = 
        if (ch.length() == 0) 
            Set()
        else
            ch.slice(0, h).listToSet().flatten()

    run txsUpToTest1 = Chain_init
        .then(all {
            assert(chain.txsUpTo(0) == Set("")),
            Chain_unchanged,
        })

    run txsUpToTest = txsUpToTest1
        .then(Chain_newBlockFrom(Set("tx1")))
        .then(Chain_newBlockFrom(Set("tx2", "tx3")))
        .then(all {
            assert(chain.txsUpTo(1) == Set("tx1")),
            assert(chain.txsUpTo(2) == Set("tx1", "tx2", "tx3")),
            Chain_unchanged,
        })

    // The transaction `tx` is committed to the chain, at most at height `h`.
    def Chain_committedUpTo(tx: Tx, h: Height): bool = 
        tx.in(chain.txsUpTo(h))

}
