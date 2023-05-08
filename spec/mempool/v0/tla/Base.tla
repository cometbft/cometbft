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

\* @typeAlias: ERROR = Str;
\* @type: ERROR;
NoError == "none"

isValid(tx) ==
    TRUE

\* @typeAlias: HEIGHT = Int;
\* @type: HEIGHT;
FirstHeight == 0
\* @type: Set(HEIGHT);
Heights == {FirstHeight, 1, 2, 3, 4}


--------------------------------------------------------------------------------
\* Bounded sequences
\* @  type: (Set(a)) => Set(Seq(a));
\* @type: (Set(a)) => Set(Int -> a);
BSeq(S) == UNION { [1 .. k -> S] : k \in 0 .. 2 }

================================================================================
Created by Hernán Vanzetto on 1 May 2023
