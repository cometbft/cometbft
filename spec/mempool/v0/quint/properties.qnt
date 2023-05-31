// -*- mode: Bluespec; -*-
module properties {
    import base.* from "./base"
    import chain as Chain from "./chain"
    import history as History from "./history"
    import mempoolv0.* from "./mempoolv0"

    // All gossipped transactions are or were in the mempool.
    val invariant1 = NodeIds.forall(node => Txs.forall(tx => 
        node.History::gossipped(tx) implies node.History::hadInMempool(tx)
    ))

    // Transactions committed to the chain cannot be in the mempool.
    val invariant2 = NodeIds.forall(node => Txs.forall(tx => 
        tx.Chain::committedUpTo(node.MempoolHeight()) implies not(node.inMempool(tx))
    ))

    // Only valid transactions are present in the mempool.
    val invariant3 = Txs.forall(tx => NodeIds.forall(node => 
        node.inMempool(tx) implies tx.isValidAt(node.chooseSomeMemTxFor(tx).height)
    ))

    temporal safety = always(all {
        invariant1,
        invariant2,
        invariant3,
    })

    // If a transaction tx is in the mempool of a node, then eventually tx will remain
    // forever committed to the chain or be forever invalid.
    temporal invariant4 = Txs.forall(tx => NodeIds.forall(node => always(
        node.inMempool(tx) implies eventually(always(tx.Chain::committedUpTo(node.MempoolHeight()) or not(tx.isValidAt(node.MempoolHeight()))))
    )))

    temporal liveness = {
        invariant4
    }
}
