// Pure Helper functions
module Helpers {
    pure def excludeE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def unionE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def print(msg, toPrint) = true
}