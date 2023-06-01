// -*- mode: Bluespec; -*-
module chain {
    import base.* from "./base"

    type Chain = List[Set[Tx]]
    var chain: Chain

    val latestHeight = 
        chain.length()

    val chainIsEmpty = 
        chain.length() == 0

    def getBlockAt(h) =
        chain[h - 1]

    val allTxsInChain =
        chain.foldl(Set(), (s, x) => s.union(x))

    //--------------------------------------------------------------------------
    action init =
        chain' = List()

    action addNewBlock(txs) =
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

    // The transaction `tx` is committed to the chain, at most at height `h`.
    def committedUpTo(tx: Tx, h: Height): bool = 
        tx.in(chain.txsUpTo(h))

    //--------------------------------------------------------------------------
    run chainTest = (chain' = List())
        .then(all {
            assert(latestHeight == 0),
            assert(allTxsInChain == Set()),
            assert(chain.txsUpTo(0) == Set()),
            unchanged,
        })
        .then(addNewBlock(Set("tx1")))
        .then(all {
            assert(latestHeight == 1),
            assert(allTxsInChain == Set("tx1")),
            assert(chain.txsUpTo(1) == Set("tx1")),
            assert(getBlockAt(1) == Set("tx1")),
            unchanged,
        })
        .then(addNewBlock(Set("tx2", "tx3")))
        .then(all {
            assert(latestHeight == 2),
            assert(allTxsInChain == Set("tx1", "tx2", "tx3")),
            assert(chain.txsUpTo(0) == Set()),
            assert(chain.txsUpTo(1) == Set("tx1")),
            assert(chain.txsUpTo(2) == Set("tx1", "tx2", "tx3")),
            assert(getBlockAt(2) == Set("tx2", "tx3")),
            unchanged,
        })
}
