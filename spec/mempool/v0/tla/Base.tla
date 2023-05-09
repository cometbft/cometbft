--------------------------------- MODULE Base ----------------------------------
EXTENDS Integers

CONSTANTS
    \* @typeAlias: NODE_ID = Str;
    \* @type: Set(NODE_ID);
    NodeIds,
    \* @typeAlias: TX = Str;
    \* @type: Set(TX);
    Txs

\* @type: NODE_ID;
NoNode == "no-node"
ASSUME NoNode \notin NodeIds

\* @type: TX;
InvalidTx == "invalid-tx"
isValid(tx) == tx \notin {InvalidTx}

\* @typeAlias: ERROR = Str;
\* @type: ERROR;
NoError == "none"

\* @typeAlias: HEIGHT = Int;
\* @type: Set(HEIGHT);
Heights == {1, 2, 3, 4}

--------------------------------------------------------------------------------
\* Bounded sequences
\* @  type: (Set(a)) => Set(Seq(a));
\* @type: (Set(a)) => Set(Int -> a);
BSeq(S) == UNION { [1 .. k -> S] : k \in 0 .. 2 }

================================================================================
Created by Hernán Vanzetto on 1 May 2023
