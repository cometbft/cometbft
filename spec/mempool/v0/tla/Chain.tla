--------------------------------- MODULE Chain ---------------------------------
EXTENDS Base, Maps, Integers

VARIABLES
    \* @type: HEIGHT -> Set(TX);
    chain

--------------------------------------------------------------------------------
Max(S) == CHOOSE x \in S: \A y \in S: x >= y

LatestHeight == Max(DOMAIN chain \union {1})

IsEmpty ==
    DOMAIN chain = {}

GetBlock(h) ==
    chain[h]

AllTxsInChain ==
    UNION { chain[h] : h \in DOMAIN chain }

--------------------------------------------------------------------------------
TypeOK ==
    chain \in Maps(Heights, SUBSET Txs)

Init ==
    chain = [x \in {} |-> {}]

SetNext(txs) == 
    chain' = MapPut(chain, LatestHeight + 1, txs)

Unchanged == 
    UNCHANGED chain

================================================================================
Created by Hernán Vanzetto on 9 May 2023
