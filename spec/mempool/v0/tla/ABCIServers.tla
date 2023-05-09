------------------------------ MODULE ABCIServers ------------------------------
(******************************************************************************)
(* An ABCI server receives asynchronous ABCI requests and replies with ABCI
responses. *)
(******************************************************************************)
EXTENDS ABCIMessages, Base, Maps, TLC

\* @type: RESPONSE;
NoResponse == [tag |-> "NoResponse", error |-> NoError]
\* @type: Set(RESPONSE);
Responses == ResponseCheckTx \union {NoResponse}

\* The ABCIServer receives requests in a queue (a set actually) and responds
\* to a queue (set) of responses.
VARIABLES 
    \* @type: NODE_ID -> REQUEST -> RESPONSE;
    requestResponses,
    \* @type: NODE_ID -> REQUEST -> NODE_ID;
    requestSenders

TypeOK ==
    /\ IsFuncMap(requestResponses, NodeIds, RequestCheckTx, Responses)
    /\ IsFuncMap(requestSenders, NodeIds, RequestCheckTx, NodeIds \cup {NoNode})

--------------------------------------------------------------------------------
(******************************************************************************)
(* Auxiliary definitions *)
(******************************************************************************)

\* @type: (NODE_ID, TX, NODE_ID) => Bool;
SendRequestNewCheckTx(nodeId, tx, sender) == 
    LET req == [tag |-> "CheckTx", tx |-> tx, checkTxType |-> "New"] IN
    /\ requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPut(@, req, NoResponse)]
    /\ requestSenders' = [requestSenders EXCEPT ![nodeId] = MapPut(@, req, sender)]

\* @type: (NODE_ID, Set(TX)) => Bool;
SendRequestRecheckTxs(nodeId, txs) == 
    LET reqs == {[tag |-> "CheckTx", tx |-> tx, checkTxType |-> "Recheck"]: tx \in txs} IN
    /\ requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPutMany(@, reqs, NoResponse)]
    /\ requestSenders' = requestSenders

ResponseFor(nodeId, request) == requestResponses[nodeId][request]
SenderFor(nodeId, request) == requestSenders[nodeId][request]

Requests(nodeId, checkTxType) == 
    { r \in DOMAIN requestResponses[nodeId]: 
        /\ r.checkTxType = checkTxType 
        /\ requestResponses[nodeId][r] # NoResponse }

CheckRequests(nodeId) == Requests(nodeId, "New")
RecheckRequests(nodeId) == Requests(nodeId, "Recheck")

\* @type: (NODE_ID, REQUEST) => Bool;
RemoveRequest(nodeId, request) ==
    /\ requestResponses[nodeId][request] # NoResponse
    /\ requestResponses' = [requestResponses EXCEPT ![nodeId] = MapRemove(@, request)]
    /\ requestSenders' = [requestSenders EXCEPT ![nodeId] = MapRemove(@, request)]

vars == <<requestResponses, requestSenders>>
Unchanged == UNCHANGED vars

--------------------------------------------------------------------------------
(******************************************************************************)
(* Actions *)
(******************************************************************************)

\* EmptyMap is not accepted by Apalache's typechecker.
\* @type: REQUEST -> RESPONSE;
EmptyMapResponses == [x \in {} |-> NoResponse]
\* @type: REQUEST -> NODE_ID;
EmptyMapNodeIds == [x \in {} |-> NoNode]

Init ==
    /\ requestResponses = [n \in NodeIds |-> EmptyMapResponses]
    /\ requestSenders = [n \in NodeIds |-> EmptyMapNodeIds]

\* The app receives a request and creates a response.
ProcessCheckTxRequest(nodeId) == 
    \* /\ PrintT(<<"ProcessCheckTxRequest", nodeId>>)
    /\ \E request \in DOMAIN requestResponses[nodeId]:
        /\ requestResponses[nodeId][request] = NoResponse
        /\ LET err == IF isValid(request.tx) THEN NoError ELSE InvalidTxError IN
           LET response == [tag |-> request.tag, error |-> err] IN
           requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPut(@, request, response)]
    /\ requestSenders' = requestSenders

================================================================================
Created by Hernán Vanzetto on 1 May 2023
