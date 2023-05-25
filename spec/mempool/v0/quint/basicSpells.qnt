// -*- mode: Bluespec; -*-
// From: https://github.com/informalsystems/quint/blob/main/examples/spells/basicSpells.qnt
module basicSpells {
    /// An annotation for writing preconditions.
    pure def require(__cond: bool): bool = __cond

    /// Remove a set element.
    pure def setRemove(__set: Set[a], __elem: a): Set[a] = {
        __set.exclude(Set(__elem))
    }

    /// Remove a map entry.
    pure def mapRemove(__map: a -> b, __key: a): a -> b = {
        __map.keys().setRemove(__key).mapBy(__k => __map.get(__k))
    }
}