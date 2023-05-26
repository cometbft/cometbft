// -*- mode: Bluespec; -*-
module properties {
    import base.* from "./base"
    import chain.* from "./chain"
    import history.* from "./history"
    import mempoolv0.* from "./mempoolv0"

    // All gossipped transactions are or were in the mempool.
    val invariant1 = NodeIds.forall(node => Txs.forall(tx => 
        node.History_gossipped(tx) implies node.History_hadInMempool(tx)
    ))

    // Transactions committed to the chain cannot be in or have belonged to the mempool.
    val invariant2 = NodeIds.forall(node => Txs.forall(tx => 
        tx.Chain_committedUpTo(node.Height_()) implies not(node.History_hadInMempool(tx))
    ))

    // Only valid transactions are present in the mempool.
    val invariant3 = Txs.forall(tx => NodeIds.forall(node => 
        node.inMempool(tx) implies tx.isValidAt(node.Height_())
    ))

    temporal safety = always(all {
        invariant1,
        invariant2,
        invariant3,
    })

    // If a transaction tx is in the mempool of a node, then eventually tx will remain
    // forever committed to the chain or be forever invalid.
    temporal invariant4 = Txs.forall(tx => NodeIds.forall(node => always(
        node.inMempool(tx) implies eventually(always(tx.Chain_committedUpTo(node.Height_()) or not(tx.isValidAt(node.Height_()))))
    )))

    temporal liveness = {
        invariant4
    }

}