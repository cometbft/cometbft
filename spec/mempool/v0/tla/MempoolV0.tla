------------------------------- MODULE MempoolV0 -------------------------------
(******************************************************************************)
(* Mempool V0                                                                 *)
(******************************************************************************)

(* Assumption: The network topology is fixed: nodes do not leave or join the
network, peers do not change. *)

(* One of the goals of this spec is to make their actions and data structures
easily mapped to the code to be able to apply MBT. *)

EXTENDS Base, Integers, Sequences, Maps, TLC, FiniteSets

CONSTANTS 
    \* @type: Int;
    MempoolMaxSize,
    
    \* The configuration of each node
    \* @typeAlias: CONFIG = [keepInvalidTxsInCache: Bool];
    \* @type: NODE_ID -> CONFIG;
    Configs,

ASSUME MempoolMaxSize > 0

--------------------------------------------------------------------------------
(******************************************************************************)
(* The ABCI application *)
(******************************************************************************)
VARIABLES 
    \* @type: NODE_ID -> REQUEST -> RESPONSE;
    requestResponses,
    \* @type: NODE_ID -> REQUEST -> NODE_ID;
    requestSenders

ABCI == INSTANCE ABCIServers

--------------------------------------------------------------------------------
\* Node states
VARIABLES
    \* @type: NODE_ID -> Set(TX);
    mempool,
    \* @type: NODE_ID -> TX -> Set(NODE_ID);
    sender,
    \* @type: NODE_ID -> Set(TX);
    cache,
    \* @type: NODE_ID -> HEIGHT;
    chainHeight

VARIABLES
    \* @type: NODE_ID -> Str;
    step,
    \* @type: NODE_ID -> ERROR;
    error,

\* vars == <<mempool, sender, cache, chainHeight, step, error>>

--------------------------------------------------------------------------------
\* Network state
VARIABLES
    \* @typeAlias: MSG = [sender: NODE_ID, tx: TX];
    \* @type: NODE_ID -> Set(MSG);
    msgs

Network == INSTANCE Network
--------------------------------------------------------------------------------
Steps == {"Init", "CheckTx", "ReceiveCheckTxResponse", 
    "Update", "ReceiveRecheckTxResponse", 
    "P2P_ReceiveTxs", "P2P_SendTx", "ABCI!ProcessCheckTxRequest"}

\* @type: Set(ERROR);
Errors == {NoError, "ErrMempoolIsFull", "ErrTxInCache"}

TypeOK == 
    /\ mempool \in [NodeIds -> SUBSET Txs]
    /\ IsFuncMap(sender, NodeIds, Txs, SUBSET NodeIds)
    /\ cache \in [NodeIds -> SUBSET Txs]
    /\ chainHeight \in [NodeIds -> Heights]
    /\ step \in [NodeIds -> Steps]
    /\ error \in [NodeIds -> Errors]
    /\ Network!TypeOK
    /\ ABCI!TypeOK

\* EmptyMap is not accepted by Apalache's typechecker.
EmptyMapNodeIds == [x \in {} |-> {}]

Init == 
    /\ mempool = [x \in NodeIds |-> {}]
    /\ sender = [x \in NodeIds |-> EmptyMapNodeIds]
    /\ cache = [x \in NodeIds |-> {}]
    /\ chainHeight = [x \in NodeIds |-> FirstHeight]
    /\ step = [x \in NodeIds |-> "Init"]
    /\ error = [x \in NodeIds |-> NoError]
    /\ Network!Init
    /\ ABCI!Init

--------------------------------------------------------------------------------
(******************************************************************************)
(* Auxiliary definitions *)
(******************************************************************************)

setStep(nodeId, s) ==
    step' = [step EXCEPT ![nodeId] = s]

setError(nodeId, err) ==
    error' = [error EXCEPT ![nodeId] = err]

(******************************************************************************)
(* Sender *)
(******************************************************************************)

\* @type: (NODE_ID, TX, NODE_ID) => Bool;
addSender(nodeId, tx, senderId) ==
    sender' = [sender EXCEPT ![nodeId] = 
        MapPut(@, tx, (IF tx \in DOMAIN @ THEN @[tx] ELSE {}) \union {senderId})]

removeSenders(nodeId, txs) ==
    sender' = [sender EXCEPT ![nodeId] = MapRemoveMany(@, txs)] 

(******************************************************************************)
(* Mempool *)
(******************************************************************************)

inMempool(nodeId, tx) ==
    tx \in mempool[nodeId]

\* @type: (NODE_ID, TX, NODE_ID) => Bool;
addToMempool(nodeId, tx, senderId) ==
    /\ mempool' = [mempool EXCEPT ![nodeId] = @ \cup {tx}]
    /\ addSender(nodeId, tx, senderId)
    \* /\ setTxHeight(nodeId, tx)

\* @type: (NODE_ID, Set(TX)) => Bool;
removeFromMempool(nodeId, txs) ==
    /\ mempool' = [mempool EXCEPT ![nodeId] = @ \ txs]
    /\ removeSenders(nodeId, txs)

mempoolIsEmpty(nodeId) ==
    mempool[nodeId] = {}

mempoolIsFull(nodeId) ==
    Cardinality(mempool[nodeId]) > MempoolMaxSize

(******************************************************************************)
(* Cache *)
(******************************************************************************)

\* @type: (NODE_ID, TX) => Bool;
inCache(nodeId, tx) ==
    tx \in cache[nodeId]

\* @type: (NODE_ID, TX) => Bool;
addToCache(nodeId, tx) ==
    cache' = [cache EXCEPT ![nodeId] = @ \union {tx}]

\* @type: (NODE_ID, TX) => Bool;
forceRemoveFromCache(nodeId, tx) ==
    cache' = [cache EXCEPT ![nodeId] = @ \ {tx}]

\* @type: (NODE_ID, TX) => Bool;
removeFromCache(nodeId, tx) ==
    IF Configs[nodeId].keepInvalidTxsInCache
    THEN forceRemoveFromCache(nodeId, tx)
    ELSE cache' = cache

--------------------------------------------------------------------------------

(* Validate a transaction received either from a client through an RPC endpoint
or from a peer via P2P. If valid, add it to the mempool. *)
\* [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
\* @type: (NODE_ID, TX, NODE_ID) => Bool;
CheckTx(nodeId, tx, senderId) ==
    /\ setStep(nodeId, "CheckTx")
    /\ mempool' = mempool
    /\ IF mempoolIsFull(nodeId) THEN
            /\ sender' = sender
            /\ cache' = cache
            /\ setError(nodeId, "ErrMempoolIsFull")
            /\ ABCI!Unchanged
        ELSE IF inCache(nodeId, tx) THEN
            \* Record new sender for the tx we've already seen.
            \* Note it's possible a tx is still in the cache but no longer in the mempool
            \* (eg. after committing a block, txs are removed from mempool but not cache),
            \* so we only record the sender for txs still in the mempool.
            /\ IF inMempool(nodeId, tx)
                THEN addSender(nodeId, tx, senderId)
                ELSE sender' = sender
            /\ cache' = cache
            /\ setError(nodeId, "ErrTxInCache")
            /\ ABCI!Unchanged
        ELSE
            /\ sender' = sender
            /\ addToCache(nodeId, tx)
            /\ setError(nodeId, NoError)
            /\ ABCI!SendRequestNewCheckTx(nodeId, tx, senderId)
    /\ UNCHANGED chainHeight

\* Receive a specific transaction from a client via RPC. */
\* [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
CheckTxRPC(nodeId, tx) ==
    /\ setStep(nodeId, "CheckTx")
    /\ CheckTx(nodeId, tx, NoNode)
    /\ UNCHANGED msgs

\* Callback that handles the response to a CheckTx request to a transaction sent
\* for the first time.
\* Note: tx and sender are arguments to the function resCbFirstTime.
\* [CListMempool.resCbFirstTime]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L369
\* @type: (NODE_ID) => Bool;
ReceiveCheckTxResponse(nodeId) ==
    /\ setStep(nodeId, "ReceiveCheckTxResponse")
    /\ \E request \in ABCI!CheckRequests(nodeId):
        LET response == ABCI!ResponseFor(nodeId, request) IN
        LET senderId == ABCI!SenderFor(nodeId, request) IN
        /\ IF response.error = NoError THEN
                IF mempoolIsFull(nodeId) THEN
                    /\ mempool' = mempool
                    /\ sender' = sender
                    /\ forceRemoveFromCache(nodeId, request.tx)
                    /\ setError(nodeId, "ErrMempoolIsFull")
                ELSE
                    /\ addToMempool(nodeId, request.tx, senderId)
                    /\ cache' = cache
                    /\ setError(nodeId, NoError)
           ELSE \* ignore invalid transaction
                /\ mempool' = mempool
                /\ sender' = sender
                /\ removeFromCache(nodeId, request.tx)
                /\ setError(nodeId, NoError)
        /\ ABCI!RemoveRequest(nodeId, request)
    /\ UNCHANGED <<chainHeight, msgs>>

\* Callback that handles the response to a CheckTx request to a transaction sent
\* after the first time (on Update).
\* [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L432
\* @type: (NODE_ID) => Bool;
ReceiveRecheckTxResponse(nodeId) == 
    /\ setStep(nodeId, "ReceiveRecheckTxResponse")
    /\ \E request \in ABCI!RecheckRequests(nodeId):
        LET response == ABCI!ResponseFor(nodeId, request) IN
        /\ inMempool(nodeId, request.tx)
        /\ IF response.error = NoError THEN
                \* Tx became invalidated due to newly committed block.
                /\ removeFromMempool(nodeId, {request.tx})
                /\ removeFromCache(nodeId, request.tx)
           ELSE /\ mempool' = mempool
                /\ sender' = sender
                /\ cache' = cache
        /\ ABCI!RemoveRequest(nodeId, request)
    /\ UNCHANGED <<error, chainHeight, msgs>>

(* The consensus reactors first reaps a list of transactions from the mempool,
executes the transactions in the app, adds them to a newly block, and finally
updates the mempool. The list of transactions is taken in FIFO order but we
don't care about the order in this spec. Then we model the mempool txs as a set
instead of a sequence of transactions. *)
(* BlockExecutor calls Update to update the mempool after executing txs.
txResults are the results of ResponseFinalizeBlock for every tx in txs.
BlockExecutor holds the mempool lock while calling this function. *)
\* [CListMempool.Update] https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
\* @type: (NODE_ID, HEIGHT, Set(TX), (TX -> Bool)) => Bool;
Update(nodeId, height, txs, txValidResults) ==
    /\ setStep(nodeId, "Update")
    /\ txs # {}
    
    /\ chainHeight' = [chainHeight EXCEPT ![nodeId] = height]
        \* TODO: need to model consensus: all nodes should create the same block
    
    \* Remove all txs from the mempool.
    /\ removeFromMempool(nodeId, txs)
    
    \* update cache for all transactions
    \* Add valid committed txs to the cache (in case they are missing).
    \* And remove invalid txs, if keepInvalidTxsInCache is false.
    /\ LET 
            validTxs == {tx \in txs: txValidResults[tx]}
            invalidTxs == {tx \in txs: ~ txValidResults[tx] /\ ~ Configs[nodeId].keepInvalidTxsInCache} 
       IN
       cache' = [cache EXCEPT ![nodeId] = (@ \union validTxs) \ invalidTxs]

    \* Either recheck non-committed txs to see if they became invalid
    \* or just notify there're some txs left.
    /\ IF mempoolIsEmpty(nodeId) THEN
            ABCI!Unchanged
       ELSE 
            \* NOTE: globalCb may be called concurrently.
            ABCI!SendRequestRecheckTxs(nodeId, txs)

    /\ UNCHANGED <<error, msgs>>

(* Receive a transaction from a peer and validate it with CheckTx. *)
\* [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
P2P_ReceiveTxs(nodeId) == 
    /\ setStep(nodeId, "P2P_ReceiveTxs")
    /\ \E msg \in Network!IncomingMsgs(nodeId):
        /\ CheckTx(nodeId, msg.tx, msg.sender)
        /\ Network!Receive(nodeId, msg)

(* The mempool reactor loops through its mempool and sends transactions one by
one to each of its peers. *)
\* [Reactor.broadcastTxRoutine] https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
P2P_SendTx(nodeId) ==
    /\ setStep(nodeId, "P2P_SendTx")
    /\ \E peer \in Network!Peers[nodeId], tx \in mempool[nodeId]:
        LET msg == [sender |-> nodeId, tx |-> tx] IN
        \* If the msg was not already sent to this peer.
        /\ ~ Network!ReceivedMsg(peer, msg)
        \* If the peer is not a tx's sender.
        /\ tx \in DOMAIN sender[nodeId] => peer \notin sender[nodeId][tx]
        /\ Network!SendTo(msg, peer)
        /\ UNCHANGED <<mempool, cache, error, chainHeight, sender>>
        /\ ABCI!Unchanged

--------------------------------------------------------------------------------
Next == 
    \E nodeId \in NodeIds:
        \* Receive some transaction from a client via RPC
        \/ \E tx \in Txs: CheckTxRPC(nodeId, tx)    

        \* Receive a (New) CheckTx response from the application
        \/ ReceiveCheckTxResponse(nodeId)

        \* Consensus reactor updates the mempool
        \/ \E txs \in (SUBSET Txs \ {{}}): 
            \E txValidResults \in [txs -> BOOLEAN]:
                Update(nodeId, chainHeight[nodeId] + 1, txs, txValidResults)

        \* Receive a (Recheck) CheckTx response from the application
        \/ ReceiveRecheckTxResponse(nodeId)

        \* Receive a transaction from a peer
        \/ P2P_ReceiveTxs(nodeId)

        \* Send a transaction in the mempool to a peer
        \/ P2P_SendTx(nodeId)

        \* The ABCI application process a request and generates a response
        \/  /\ ABCI!ProcessCheckTxRequest(nodeId)
            /\ setStep(nodeId, "ABCI!ProcessCheckTxRequest")
            /\ UNCHANGED <<mempool, sender, cache, error, chainHeight, msgs>>

--------------------------------------------------------------------------------
(******************************************************************************)
(* Test scenarios *)
(******************************************************************************)

EmptyCache == 
    \E nodeId \in NodeIds:
        /\ mempool[nodeId] # {}
        /\ cache[nodeId] = {}
NotEmptyCache == ~ EmptyCache

NonEmptyCache == 
    \E nodeId \in NodeIds:
        /\ mempool[nodeId] # {}
        /\ cache[nodeId] # {}
NotNonEmptyCache == ~ NonEmptyCache

ReceiveRecheckTxResponseProp ==
    \E nodeId \in NodeIds:
        step[nodeId] = "ReceiveRecheckTxResponse"
NotReceiveRecheckTxResponse == ~ ReceiveRecheckTxResponseProp

SendTxTest ==
    \E nodeId \in NodeIds:
        step[nodeId] = "SendTx"
NotSendTxTest == ~ SendTxTest

\* @ typeAlias: STATE = [step: NODE_ID -> Str];
\* @typeAlias: STATE = [requestResponses: NODE_ID -> REQUEST -> RESPONSE, requestSenders: NODE_ID -> REQUEST -> NODE_ID, mempool: NODE_ID -> Set(TX), sender: NODE_ID -> TX -> Set(NODE_ID), cache: NODE_ID -> Set(TX), step: NODE_ID -> Str, error: NODE_ID -> ERROR, chainHeight: NODE_ID -> HEIGHT];
\* @type: Seq(STATE) => Bool;
SendThenCheck(trace) ==
    \E i, j \in DOMAIN trace: i < j /\
        \E n \in NodeIds:
            LET state1 == trace[i] IN 
            LET state2 == trace[j] IN
            /\ state1.step[n] = "SendTx"
            /\ state2.step[n] = "CheckTx" 
            \* /\ Len(trace) = 10

\* @type: Seq(STATE) => Bool;
NotSendThenCheck(trace) == ~ SendThenCheck(trace)

================================================================================
Created by Hernán Vanzetto on 1 May 2023
