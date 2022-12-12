// Pure Helper functions
module Spells {
    pure def excludeE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def unionE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def print(msg, toPrint) = true

    pure def has(__map: a -> b, __key: a): bool = {
        __map.keys().contains(__key)
    }

    pure def getOrElse(__map: a -> b, __key: a, __default: b): b = {
        if (__map.has(__key)) {
            __map.get(__key)
        } else {
            __default
        }
    }
}