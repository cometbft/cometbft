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
    \* For each node, we keep a mapping from requests to 
    \* - the sender of the request, if available, and
    \* - to the response from the application, when available.
    \* @type: NODE_ID -> REQUEST -> <<NODE_ID, RESPONSE>>;
    requestResponses

TypeOK ==
    IsFuncMap(requestResponses, NodeIds, RequestCheckTx, (NodeIds \cup {NoNode}) \X Responses)

--------------------------------------------------------------------------------
(******************************************************************************)
(* Auxiliary definitions *)
(******************************************************************************)

SenderFor(nodeId, request) == requestResponses[nodeId][request][1]
ResponseFor(nodeId, request) == requestResponses[nodeId][request][2]

HasResponse(nodeId, request) ==
    ResponseFor(nodeId, request) # NoResponse

Requests(nodeId, checkTxType) == 
    { r \in DOMAIN requestResponses[nodeId]: 
        /\ r.checkTxType = checkTxType 
        /\ HasResponse(nodeId, r) }

CheckRequests(nodeId) == Requests(nodeId, "New")
RecheckRequests(nodeId) == Requests(nodeId, "Recheck")

--------------------------------------------------------------------------------
(******************************************************************************)
(* Actions *)
(******************************************************************************)

\* EmptyMap is not accepted by Apalache's typechecker.
\* @type: REQUEST -> <<NODE_ID, RESPONSE>>;
EmptyMapResponses == [x \in {} |-> <<NoNode, NoResponse>>]

Init ==
    requestResponses = [n \in NodeIds |-> EmptyMapResponses]

\* @type: (NODE_ID, Set(TX)) => Bool;
SendRequestRecheckTxs(nodeId, txs) == 
    LET reqs == {[tag |-> "CheckTx", tx |-> tx, checkTxType |-> "Recheck"]: tx \in txs} IN
    requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPutMany(@, reqs, <<NoNode, NoResponse>>)]

\* @type: (NODE_ID, TX, NODE_ID) => Bool;
SendRequestNewCheckTx(nodeId, tx, sender) == 
    LET req == [tag |-> "CheckTx", tx |-> tx, checkTxType |-> "New"] IN
    requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPut(@, req, <<sender, NoResponse>>)]

\* The app receives a request and creates a response.
ProcessCheckTxRequest(nodeId) == 
    \E request \in DOMAIN requestResponses[nodeId]:
        /\ ~ HasResponse(nodeId, request)
        /\ LET sender == SenderFor(nodeId, request) IN
           LET err == IF isValid(request.tx) THEN NoError ELSE InvalidTxError IN
           LET response == [tag |-> request.tag, error |-> err] IN
           requestResponses' = [requestResponses EXCEPT ![nodeId] = MapPut(@, request, <<sender, response>>)]

\* @type: (NODE_ID, REQUEST) => Bool;
RemoveRequest(nodeId, request) ==
    /\ HasResponse(nodeId, request)
    /\ requestResponses' = [requestResponses EXCEPT ![nodeId] = MapRemove(@, request)]

Unchanged == UNCHANGED requestResponses

================================================================================
Created by Hernán Vanzetto on 1 May 2023
