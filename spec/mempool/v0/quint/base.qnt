// -*- mode: Bluespec; -*-
module base {
    // Nodes
    type NodeId = str
    val noNode: NodeId = "no-node"
    // const NodeIds: Set[NodeId]
    val NodeIds: Set[NodeId] = Set("n1", "n2", "n3")

    // Heights
    type Height = int
    val Heights = 0.to(3)

    // Transactions
    type Tx = str
    // val invalidTx: Tx = "invalid-tx"
    // const Txs: Set[Tx]
    val Txs: Set[Tx] = Set("tx1", "tx2", "tx3")
    def isValidAt(tx: Tx, h: Height): bool = any {
        h == 1 and tx.in(Set("tx1")),
        h == 2 and tx.in(Set("tx2", "tx3")),
        h == 3 and tx.in(Set("tx2", "tx3")),
    }

    // Error
    type Error = str
    val noError: Error = "none"

    //--------------------------------------------------------------------------
    // More spells
    //--------------------------------------------------------------------------
    pure def setAdd(__set: Set[a], __elem: a): Set[a] =
        __set.union(Set(__elem))

    pure def isEmpty(__set: Set[a]): bool =
        __set == Set()

    pure def nonEmpty(__set: Set[a]): bool =
        __set != Set()

    pure def notIn(__x: a, __set: Set[a]): bool =
        not(__x.in(__set))

    pure def listAsSet(__list: List[a]): Set[a] =
        __list.foldl(Set(), (__s, __x) => __s.setAdd(__x))

    /// Update a map entry using the previous value.
    ///
    /// @param __map the map to update
    /// @param __key the key to search for
    /// @param __f a function that returns the new value for __key 
    ///   when applied to __key's old value
    /// @returns a new map equal to __map except that __key maps
    ///   to __f applied to __key's old value
    pure def mapPut(__map: a -> b, __key: a, __f: b => b): (a -> b) = 
        __map.put(__key, __f(__map.get(__key)))

}
