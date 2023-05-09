------------------------------- MODULE Network -------------------------------
EXTENDS Base

CONSTANTS
    \* The network topology.
    \* @type: NODE_ID -> Set(NODE_ID);
    Peers

ASSUME \A n \in NodeIds: n \notin Peers[n]

\* Network state
VARIABLES
    \* For each node, a set of incoming messages not yet processed.
    \* @typeAlias: MSG = [sender: NODE_ID, tx: TX];
    \* @type: NODE_ID -> Set(MSG);
    msgs

TxMsgs == [sender: NodeIds, tx: Txs]

TypeOK ==
    msgs \in [NodeIds -> SUBSET TxMsgs]

Init ==
    msgs = [x \in NodeIds |-> {}]

SendTo(msg, peer) == 
    msgs' = [msgs EXCEPT ![peer] = @ \union {msg}]

Receive(nodeId, msg) ==
    msgs' = [msgs EXCEPT ![nodeId] = @ \ {msg}]

IncomingMsgs(nodeId) ==
    msgs[nodeId]

ReceivedMsg(nodeId, msg) ==
    \E m \in msgs[nodeId]: m = msg

================================================================================
Created by Hernán Vanzetto on 9 May 2023
