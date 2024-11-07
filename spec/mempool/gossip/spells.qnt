// -*- mode: Bluespec; -*-
module spells {

    //--------------------------------------------------------------------------
    // Basic
    //--------------------------------------------------------------------------

    /// An annotation for writing preconditions.
    pure def require(__cond: bool): bool = __cond

    //--------------------------------------------------------------------------
    // Options
    //--------------------------------------------------------------------------

    type Option[a] =
        | Some(a)
        | None

    pure def isSome(__opt) =
        match __opt {
        | Some(_) => true
        | None => false
        }

    pure def isNone(__opt) = not(isSome(__opt))

    pure def optionMap(__opt: Option[a], __f: a => b): Option[b] =
        match __opt {
        | Some(a) => Some(__f(a))
        | None => None
        }

    pure def optionFlatten(__opt: Option[Option[a]]): Option[a] =
        match __opt {
        | Some(o) => o
        | None => None
        }

    pure def optionGetDefault(__opt: Option[a], __default: a): a =
        match __opt {
        | Some(o) => o
        | None => __default
        }

    pure def optionToSet(__opt: Option[a]): Set[a] =
        match __opt {
        | Some(o) => Set(o)
        | None => Set()
        }

    //--------------------------------------------------------------------------
    // Sets
    //--------------------------------------------------------------------------

    pure def join(__set: Set[a], __elem: a): Set[a] =
        __set.union(Set(__elem))

    pure def isEmpty(__set: Set[a]): bool =
        __set == Set()

    pure def nonEmpty(__set: Set[a]): bool =
        __set != Set()

    pure def except(__set: Set[a], __elem: a): Set[a] =
        __set.exclude(Set(__elem))

    //--------------------------------------------------------------------------
    // Maps
    //--------------------------------------------------------------------------

    /// Update a map entry using the previous value.
    ///
    /// @param __map the map to update
    /// @param __key the key to search for
    /// @param __f a function that returns the new value for __key 
    ///   when applied to __key's old value
    /// @returns a new map equal to __map except that __key maps
    ///   to __f applied to __key's old value
    pure def update(__map: a -> b, __key: a, __f: b => b): (a -> b) = 
        __map.put(__key, __f(__map.get(__key)))

    /// Update multiple entries in a map.
    ///
    /// @param __map the map to update
    /// @param __keys the set of keys to be updated in __map
    /// @param __values a function from __map values to new values
    /// @returns a new map equal to __map except that for each key k in __keys 
    ///   will map to __values(k)
    pure def updateMultiple(__map: a -> b, __keys: Set[a], __values: b => b): (a -> b) = 
        __map.keys().union(__keys).mapBy(k => 
            if (k.in(__keys)) __values(__map.get(k)) else __map.get(k)
        )

    pure def mapRemoveMultiple(__map: a -> b, __keys: Set[a]): (a -> b) = 
        __map.keys().filter(k => not(k.in(__keys))).mapBy(k => __map.get(k))

    pure def mapRemove(__map: a -> b, __key: a): (a -> b) = 
        mapRemoveMultiple(__map, Set(__key))

    pure def mapGet(__map: a -> b, x: a): Option[b] = 
        if (__map.keys().contains(x)) Some(__map.get(x)) else None

    pure def mapGetDefault(__map: a -> b, __x: a, __default: b): b = 
        if (__map.keys().contains(__x)) __map.get(__x) else __default

    //--------------------------------------------------------------------------
    // Lists
    //--------------------------------------------------------------------------

    pure def listIsEmpty(__list: List[a]): bool =
        __list == List()

    pure def listNonEmpty(__list: List[a]): bool =
        __list != List()

    pure def listToSet(__list: List[a]): Set[a] =
        __list.foldl(Set(), (__s, __x) => __s.join(__x))

    pure def setToList(__set: Set[a]): List[a] =
        __set.fold(List(), (__l, __x) => __l.append(__x))

    pure def includes(__list: List[a], x: a): bool =
        __list.listToSet().contains(x)

    pure def headOption(__list: List[a]): Option[a] =
        if (__list.length() > 0) Some(__list.head()) else None

}