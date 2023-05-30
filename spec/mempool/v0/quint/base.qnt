// -*- mode: Bluespec; -*-
module params {

    val __NodeIds = Set("n1", "n2")

    val __Heights = 0.to(2)

    val __Txs = Set("tx1", "tx2")

    def __isValidAt(tx, h) = or {
        h == 1 and tx.in(Set("tx1")),
        h == 2 and tx.in(Set("tx2")),
    }

}

module base {
    import params.*

    // Nodes
    type NodeId = str
    val noNode: NodeId = "no-node"
    // const NodeIds: Set[NodeId]
    val NodeIds: Set[NodeId] = __NodeIds

    // Heights
    type Height = int
    val Heights = __Heights

    // Transactions
    type Tx = str
    // val invalidTx: Tx = "invalid-tx"
    // const Txs: Set[Tx]
    val Txs: Set[Tx] = __Txs
    def isValidAt(tx: Tx, h: Height): bool = __isValidAt(tx, h)
    
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

    pure def nonEmptyPowerset(__set: Set[a]): Set[Set[a]] =
        __set.powerset().exclude(Set(Set()))

    //--------------------------------------------------------------------------
    pure def listToSet(__list: List[a]): Set[a] =
        __list.foldl(Set(), (__s, __x) => __s.setAdd(__x))

    pure def setToList(__set: Set[a]): List[a] =
        __set.fold(List(), (__l, __x) => __l.append(__x))

    //--------------------------------------------------------------------------
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
