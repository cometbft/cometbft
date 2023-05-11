------------------------------ MODULE MempoolV0MC ------------------------------
(******************************************************************************)
(* Model Checker parameters                                                   *)
(******************************************************************************)
EXTENDS MempoolV0

instance_NodeIds == {"n1", "n2"}
instance_Txs == {"invalid-tx", "tx1", "tx2"}

instance_MempoolMaxSize == 3
instance_Configs == [x \in NodeIds |->
    CASE x = "n1" -> [keepInvalidTxsInCache |-> FALSE]
      [] x = "n2" -> [keepInvalidTxsInCache |-> FALSE]
    \*   [] x = "n3" -> [keepInvalidTxsInCache |-> FALSE]
    \*   [] OTHER -> [keepInvalidTxsInCache |-> FALSE]
]

instance_Peers == [x \in NodeIds |->
    CASE x = "n1" -> {"n2"}
      [] x = "n2" -> {"n1"}
    \*   [] x = "n3" -> {"n1", "n2"}
    \*   [] OTHER -> {}
]

--------------------------------------------------------------------------------
(******************************************************************************)
(* Model instantiation for Apalache. *)
(* https://apalache.informal.systems/docs/apalache/parameters.html?highlight=constinit#constinit-predicate *)
(******************************************************************************)

ConstInit ==
    /\ NodeIds = instance_NodeIds
    /\ Txs = instance_Txs
    /\ MempoolMaxSize = instance_MempoolMaxSize
    /\ Configs = instance_Configs
    /\ Peers = instance_Peers

--------------------------------------------------------------------------------

\* @type: <<NODE_ID, Str, Str -> Str>>;
View == 
    <<
        step.node,
        step.name, 
        error
    >>
================================================================================
