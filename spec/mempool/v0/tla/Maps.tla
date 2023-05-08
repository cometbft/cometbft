--------------------------------- MODULE Maps ----------------------------------
(******************************************************************************)
(* Maps represent partial functions.                                          *)
(*                                                                            *)
(* In TLA+ functions are total, that is, they are defined over all elements   *)
(* of their domain S. A map represents a partial function over S, that is, a  *)
(* function whose domain is a subset of S, or possibly S itself if the        *)
(* function is actually total.                                                *)
(******************************************************************************)

\* The set of all maps with (partial) domain S and codomain T.
\* @type: (Set(a), Set(b)) => Set(a -> b);
Maps(S, T) == UNION { [A -> T]: A \in SUBSET S }

\* @type: (a -> b, Set(a), Set(b)) => Bool;
IsMap(f, S, T) ==
    /\ f = [x \in DOMAIN f |-> f[x]]
    /\ DOMAIN f \subseteq S
    /\ \A x \in DOMAIN f: f[x] \in T

THEOREM MapsDef ==
    ASSUME NEW f, NEW S, NEW T
    PROVE f \in Maps(S, T) <=> IsMap(f, S, T)

\* The empty tuple is the only function in TLA+ with an empty domain.
EmptyMap == <<>>

\* @type: (a -> b, Set(a), b) => a -> b;
MapPutMany(map, keys, value) ==
    [k \in (DOMAIN map) \cup keys |-> IF k \in keys THEN value ELSE map[k]]

\* @type: (a -> b, a, b) => a -> b;
MapPut(map, key, value) ==
    MapPutMany(map, {key}, value)

\* @type: (a -> b, Set(a)) => a -> b;
MapRemoveMany(map, keys) ==
    [k \in DOMAIN map \ keys |-> map[k]]

\* @type: (a -> b, a) => a -> b;
MapRemove(map, key) ==
    MapRemoveMany(map, {key})

\* IsFuncMap(f, S, T, U) <=> f \in [S -> Maps(T, U)]
\* @type: (a -> b -> c, Set(a), Set(b), Set(c)) => Bool;
IsFuncMap(f, S, T, U) ==
    /\ f = [x \in DOMAIN f |-> f[x]]
    /\ DOMAIN f = S
    /\ \A x \in DOMAIN f: IsMap(f[x], T, U)

================================================================================
Created by Hernán Vanzetto on 10 August 2022
Updated by Hernán Vanzetto on 1 May 2023
