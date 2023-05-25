// -*- mode: Bluespec; -*-
module base {
    // Nodes
    type NodeId = str
    val noNode: NodeId = "no-node"
    // const NodeIds: Set[NodeId]
    val NodeIds: Set[NodeId] = Set(noNode, "n1", "n2", "n3")

    // Transactions
    type Tx = str
    val invalidTx: Tx = "invalid-tx"
    // const Txs: Set[Tx]
    val Txs: Set[Tx] = Set("tx1", "tx2", invalidTx)
    def isValid(tx) = tx != invalidTx
    type Error = str

    // Error
    val noError: Error = "none"

    // Heights
    type Height = int
    val Heights = 0.to(4)

    //--------------------------------------------------------------------------
    // More spells
    //--------------------------------------------------------------------------
    pure def setAdd(__set: Set[a], __elem: a): Set[a] =
        __set.union(Set(__elem))

    pure def isEmpty(__set: Set[a]): bool =
        __set == Set()

    pure def nonEmpty(__set: Set[a]): bool =
        __set != Set()
}
