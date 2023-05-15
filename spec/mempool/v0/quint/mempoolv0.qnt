// -*- mode: Bluespec; -*-
/** Mempool V0
  * 
  */
// Assumption: The network topology is fixed: nodes do not leave or join the
// network, peers do not change.

// One of the goals of this spec is to make their actions and data structures
// easily mapped to the code.

// From: https://github.com/informalsystems/quint/blob/main/examples/spells/basicSpells.qnt
module basicSpells {
    /// An annotation for writing preconditions.
    pure def require(__cond: bool): bool = __cond

    /// Remove a set element.
    pure def setRemove(__set: Set[a], __elem: a): Set[a] = {
        __set.exclude(Set(__elem))
    }

    /// Remove a map entry.
    pure def mapRemove(__map: a -> b, __key: a): a -> b = {
        __map.keys().setRemove(__key).mapBy(__k => __map.get(__k))
    }
}

module Base {
    // Nodes
    type NodeId = str
    val noNode: NodeId = "no-node"
    // const NodeIds: Set[NodeId]
    val NodeIds: Set[NodeId] = Set(noNode, "n1", "n2", "n3")

    // Transactions
    type Tx = str
    val invalidTx: Tx = "invalid-tx"
    // const Txs: Set[Tx]
    val Txs: Set[Tx] = Set("tx1", "tx2", invalidTx)
    def isValid(tx) = tx != invalidTx
    type Error = str

    // Error
    val noError: Error = "none"

    // Heights
    type Height = int
    val Heights = 0.to(4)

    //--------------------------------------------------------------------------
    // More spells
    //--------------------------------------------------------------------------
    pure def setAdd(__set: Set[a], __elem: a): Set[a] =
        __set.union(Set(__elem))

    pure def isEmpty(__set: Set[a]): bool =
        __set == Set()

    pure def nonEmpty(__set: Set[a]): bool =
        __set != Set()

}

/* Messages as defined in the protobuf files */
module ABCIMessages {
    import Base.*

    // [tendermint.abci.CheckTxType]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L94
    val CheckTxTypes = Set("New", "Recheck")

    // [tendermint.abci.RequestCheckTx]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L99
    type Request = {
        tag: str, // The only value that we use is "CheckTx".
        tx: Tx, 
        checkTxType: str,
    }

    // [tendermint.abci.ResponseCheckTx]: https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L254
    type Response = {
        tag: str, 
        err: Error, // called `code` in protobuf
    }
}

// An ABCI server receives asynchronous ABCI requests and replies with ABCI responses.
module ABCIServers {
    import basicSpells.* //from "./basicSpells"
    import Base.*
    import ABCIMessages.*

    type ABCIRequests = { tag: str, tx: Tx }
    val ABCIRequestTags = Set("CheckTx", "RecheckTx")
    
    // The ABCIServer receives requests in a queue (a set actually) and responds
    // to a queue (set) of responses. Each request maps to the node id of the
    // sender and to the response, if available.
    var requestSenderResponses: NodeId -> (Request -> (NodeId, Response))

    val noResponse: Response = { tag: "", err: noError }

    //--------------------------------------------------------------------------
    // Errors
    //--------------------------------------------------------------------------
    val InvalidTxError: Error = "err:invalid-tx"
    val MempoolIsFullError: Error = "err:mempool-full"
    val TxInCacheError: Error = "err:tx-in-cache"
    val InvalidTxWarning: Error = "warn:invalid-tx"

    //--------------------------------------------------------------------------
    // Auxiliary definitions
    //--------------------------------------------------------------------------
    def ABCI_senderFor(nodeId, request) = 
        (requestSenderResponses.get(nodeId).get(request))._1
    
    def ABCI_responseFor(nodeId, request) = 
        requestSenderResponses.get(nodeId).get(request)._2

    def hasResponse(nodeId, request) =
        ABCI_responseFor(nodeId, request) != noResponse

    def requestsWithResponse(nodeId: NodeId, checkTxType: str): Set[Request] =
        requestSenderResponses.get(nodeId).keys()
            .filter(r => r.checkTxType == checkTxType and nodeId.hasResponse(r))
    def ABCI_checkTxRequests(nodeId) = requestsWithResponse(nodeId, "New")
    def ABCI_recheckTxRequests(nodeId) = requestsWithResponse(nodeId, "Recheck")

    val emptyResponse = (noNode, noResponse)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action ABCI_Init =
        requestSenderResponses' = NodeIds.mapBy(_ => Map())

    // For a "New" request we store the sender.
    action ABCI_SendRequestNewCheckTx(nodeId, tx, senderId) = 
        val req = { tag: "CheckTx", tx: tx, checkTxType: "New" }
        requestSenderResponses' = requestSenderResponses.put(nodeId, 
            requestSenderResponses.get(nodeId).put(req, (senderId, noResponse)))

    // For "Recheck" we send multiple requests at once.
    action ABCI_SendRequestRecheckTxs(nodeId, txs) =
        val newRequests = txs.map(tx => { tag: "CheckTx", tx: tx, checkTxType: "Recheck" })
        // Add all the new requests, with no sender and no response.
        requestSenderResponses' = requestSenderResponses.put(nodeId, {
            val m = requestSenderResponses.get(nodeId)
            m.keys().mapBy(req => if (req.in(newRequests)) emptyResponse else m.get(req))
        })

    // The app receives a CheckTx request and creates a response.
    action ABCI_ProcessCheckTxRequest(nodeId) =
        nondet req: Request = oneOf(requestSenderResponses.get(nodeId).keys()) 
        all {
            require(not(nodeId.hasResponse(req))),
            val senderId = ABCI_senderFor(nodeId, req)
            val err = if (req.tx.isValid()) noError else InvalidTxError
            val resp = { tag: req.tag, err: err }
            requestSenderResponses' = requestSenderResponses.put(nodeId, 
                requestSenderResponses.get(nodeId).put(req, (senderId, resp))),
        }

    action ABCI_RemoveRequest(nodeId, request) = all {
        require(hasResponse(nodeId, request)), // We keep this condition to allow the action to return false; otherwise it will always evaluate to true.
        requestSenderResponses' = requestSenderResponses.put(nodeId, 
            requestSenderResponses.get(nodeId).mapRemove(request)), 
    }

    action ABCI_Unchanged =
        requestSenderResponses' = requestSenderResponses

}

module Network {
    import basicSpells.* //from "./basicSpells"
    import Base.* 

    // const Peers: NodeId -> Set[NodeId]
    val Peers: NodeId -> Set[NodeId] = Map(
        "n1" -> Set("n2", "n3"),
        "n2" -> Set("n1", "n3"),
        "n3" -> Set("n1", "n2")
    )

    //--------------------------------------------------------------------------
    // Network state
    type TxsMsg = { tag: str, senderId: NodeId, tx: Tx }
    var msgs: NodeId -> Set[TxsMsg] // map from receiver to messages to process

    //--------------------------------------------------------------------------
    action Network_Init =
        msgs' = NodeIds.mapBy(_ => Set())

    action sendTo(msg: TxsMsg, peer: NodeId): bool =
        msgs' = msgs.put(peer, msgs.get(peer).setAdd(msg))

    action receiveMsg(nodeId: NodeId, msg: TxsMsg): bool =
        msgs' = msgs.put(nodeId, msgs.get(nodeId).setRemove(msg))    

    action Network_Unchanged =
        msgs' = msgs

    //--------------------------------------------------------------------------
    def Network_incomingMsgs(nodeId) =
        msgs.get(nodeId)
}

module Chain {
    import Base.* 

    var chain: List[Set[Tx]]

    val Chain_latestHeight = 
        chain.length() - 1

    val Chain_isEmpty = 
        chain.length() == 0

    def Chain_getBlock(h) =
        chain[h]

    val Chain_allTxsInChain =
        chain.foldl(Set(), (s, x) => s.union(x))

    //--------------------------------------------------------------------------
    action Chain_Init =
        chain' = List()

    action Chain_NewBlockFrom(txs) =
        chain' = chain.append(txs)

    action Chain_Unchanged = 
        chain' = chain

}

module MempoolV0 {
    import basicSpells.* //from "./basicSpells"
    import Base.* 
    import Chain.*
    import ABCIServers.* // as ABCI
    import Network.*

    //--------------------------------------------------------------------------
    // Spec parameters
    //--------------------------------------------------------------------------
    // const MempoolMaxSize: int
    val MempoolMaxSize: int = 2
    // assume(MempoolMaxSize > 0)

    type Config = { keepInvalidTxsInCache: bool }
    // const Configs: NodeId -> Config
    // Configs = Map("n1" -> { keepInvalidTxsInCache -> false }, "n2" -> { keepInvalidTxsInCache -> false }, "n2" -> { keepInvalidTxsInCache -> false }), 
    val Configs: NodeId -> Config = NodeIds.mapBy(_ => { keepInvalidTxsInCache: false })

    //--------------------------------------------------------------------------
    // Node metadata
    //--------------------------------------------------------------------------
    type Step = str
    val Steps: Set[Step] = Set("Init", "CheckTx", "ReceiveCheckTxResponse", 
        "Update", "ReceiveRecheckTxResponse", "P2P_ReceiveTxs", "P2P_SendTx", 
        "ABCI::ProcessCheckTxRequest")

    var step: { node: NodeId, name: Step }
    var error: (NodeId, Error)

    /* Auxiliary definitions */
    action setStep(nodeId, s) =
        step' = { node: nodeId, name: s }

    action setError(nodeId, err) =
        error' = (nodeId, err)

    //--------------------------------------------------------------------------
    // Mempool state
    //--------------------------------------------------------------------------
    // [mempoolTx]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L671
    type MempoolTx = { 
        tx: Tx, 
        height: Height, 
        senders: Set[NodeId]
    }
    
    var mempool: NodeId -> Set[MempoolTx]
    var cache: NodeId -> Set[Tx]
    var height: NodeId -> Height

    //--------------------------------------------------------------------------
    /* Cache auxiliary definitions */
    def inCache(nodeId, tx) =
        cache.get(nodeId).contains(tx)

    action addToCache(nodeId, tx) =
        cache' = cache.put(nodeId, cache.get(nodeId).setAdd(tx))

    action forceRemoveFromCache(nodeId: NodeId, tx: Tx): bool =
        cache' = cache.put(nodeId, cache.get(nodeId).setRemove(tx))

    action removeFromCache(nodeId, invalidTx_) =
        if (not(Configs.get(nodeId).keepInvalidTxsInCache)) {
            forceRemoveFromCache(nodeId, invalidTx_)
        } else cache' = cache

    //--------------------------------------------------------------------------
    /* Mempool auxiliary definitions */    
    def mempoolIsEmpty(nodeId) =
        mempool.get(nodeId).isEmpty()

    def mempoolIsFull(nodeId) =
        size(mempool.get(nodeId)) > MempoolMaxSize

    def mempoolTxs(nodeId: NodeId): Set[Tx] = 
        mempool.get(nodeId).map(e => e.tx)

    def memTxFor(nodeId: NodeId, tx: Tx): Set[MempoolTx] =
        mempool.get(nodeId).filter(e => e.tx == tx)
    
    def sendersFor(nodeId: NodeId, tx: Tx): Set[NodeId] = 
        memTxFor(nodeId, tx).map(e => e.senders).flatten()

    def inMempool(nodeId, tx) =
        tx.in(mempoolTxs(nodeId))

    def senderFor(nodeId, tx) = 
        mempool.get(nodeId).map(e => e.tx)

    action addToMempool(nodeId, tx, h, senderId) =
        val senders = if (senderId == noNode) Set() else Set(senderId)
        val memTx = { tx: tx, height: h, senders: senders }
        mempool' = mempool.put(nodeId, mempool.get(nodeId).setAdd(memTx))

    action removeFromMempool(nodeId: NodeId, txs: Set[Tx]): bool =
        mempool' = mempool.put(nodeId, mempool.get(nodeId).filter(e => e.tx.in(txs)))

    action addSender(nodeId, tx, senderId) =
        if (inMempool(nodeId, tx)) {
            val oldMemTx: MempoolTx = memTxFor(nodeId, tx).chooseSome()
            val newMemTx = oldMemTx.with("senders", oldMemTx.senders.union(senderId))
            mempool' = mempool.put(nodeId, mempool.get(nodeId).setRemove(oldMemTx).setAdd(newMemTx))
        } else mempool' = mempool

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action Init = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        cache' = NodeIds.mapBy(_ => Set()),
        step' = { node: noNode, name: "Init" },
        error' = (noNode, noError),
        height' = NodeIds.mapBy(_ => 0),
        Network_Init,
        Chain_Init,
        ABCI_Init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    action CheckTx(nodeId: NodeId, tx: Tx, senderId: NodeId): bool = all {
        if (mempoolIsFull(nodeId)) all {
            mempool' = mempool,
            cache' = cache,
            setError(nodeId, MempoolIsFullError),
            ABCI_Unchanged,
        } else all {
            if (inCache(nodeId, tx)) all {
        		// Record new sender for the tx we've already seen.
                // Note it's possible a tx is still in the cache but no longer in the mempool
                // (eg. after committing a block, txs are removed from mempool but not cache),
                // so we only record the sender for txs still in the mempool.
                if (inMempool(nodeId, tx)) {
                    addSender(nodeId, tx, senderId)
                } else mempool' = mempool,
                cache' = cache,
                setError(nodeId, TxInCacheError),
                ABCI_Unchanged,
            } else all {
                mempool' = mempool,
                addToCache(nodeId, tx),
                setError(nodeId, noError),
                ABCI_SendRequestNewCheckTx(nodeId, tx, senderId),
            }
        },
        height' = height,
    }

    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    action CheckTxRPC_(nodeId, tx) = all {
        setStep(nodeId, "CheckTx"),
        CheckTx(nodeId, tx, noNode),
        Network_Unchanged,
        Chain_Unchanged,
    }

    /* Receive some transaction from a client via RPC. */
    action CheckTxRPC(nodeId) =
        nondet tx = oneOf(Txs)
        CheckTxRPC_(nodeId, tx)

    // Callback function to process a CheckTx response.
    // Note that tx and sender are arguments to CListMempool.resCbFirstTime.
    // [CListMempool.resCbFirstTime]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L369
    action ReceiveCheckTxResponse(nodeId) = all {
        setStep(nodeId, "ReceiveCheckTxResponse"),
        ABCI_checkTxRequests(nodeId) != Set(),
        nondet request = oneOf(ABCI_checkTxRequests(nodeId)) all {
            val response = ABCI_responseFor(nodeId, request)
            val senderId = ABCI_senderFor(nodeId, request)
            all {
                ABCI_RemoveRequest(nodeId, request),
                if (response.err == noError) {
                    if (mempoolIsFull(nodeId)) all {
                        forceRemoveFromCache(nodeId, request.tx), // There may be enough space for tx later.
                        setError(nodeId, MempoolIsFullError),
                        mempool' = mempool,
                    } else all { 
                        addToMempool(nodeId, request.tx, height.get(nodeId), senderId), // TODO: check if tx is in mempool
                        cache' = cache,
                        setError(nodeId, noError),
                    }
                } else all { // ignore invalid transaction
                    mempool' = mempool,
                    removeFromCache(nodeId, request.tx),
                    setError(nodeId, InvalidTxWarning),
                }
            }
        },
        height' = height,
        Network_Unchanged,
        Chain_Unchanged,
    }

    /** Consensus' BlockExecutor calls Update to update the mempool after executing txs.
      * txResults are the results of ResponseFinalizeBlock for every tx in txs.
      * BlockExecutor holds the mempool lock while calling this function.
      */
    // [CListMempool.Update]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
    action Update(nodeId: NodeId, h: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        setStep(nodeId, "Update"),
        require(txs != Set()),

        // Update mempool's height.
        height' = height.put(nodeId, h),

        // Remove all txs from the mempool.
        removeFromMempool(nodeId, txs),

        // Update cache for all transactions.
        // - Add valid committed txs to the cache (in case they are missing).
        // - Remove invalid txs, if keepInvalidTxsInCache is false.
        val validTxs = txs.filter(tx => txValidResults.get(tx))
        val invalidTxs = txs.filter(tx => not(txValidResults.get(tx)) and not(Configs.get(nodeId).keepInvalidTxsInCache)) 
        cache' = cache.put(nodeId, cache.get(nodeId).union(validTxs).exclude(invalidTxs)),

        // Either recheck non-committed txs to see if they became invalid
        // or just notify there're some txs left.
        if (not(mempoolIsEmpty(nodeId))) {
            // CHECK: in the code, ReceiveRecheckTxResponse may be called concurrently.
            ABCI_SendRequestRecheckTxs(nodeId, txs)
        } else ABCI_Unchanged,

        setError(nodeId, noError),
        Network_Unchanged,
        Chain_Unchanged,
    }

    action Consensus_Update(nodeId: NodeId): bool = all {
        require(not(Chain_isEmpty)),
        require(height.get(nodeId) <= Chain_latestHeight),
        val txs: Set[Tx] = Chain_getBlock(height.get(nodeId))
        nondet txValidResults: Tx -> Bool = oneOf(txs.setOfMaps(Bool))
        Update(nodeId, height.get(nodeId), txs, txValidResults),
    }

    // Callback function to process a CheckTx response, when rechecking.
    // [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L432
    action ReceiveRecheckTxResponse(nodeId: NodeId): bool = all {
        setStep(nodeId, "ReceiveRecheckTxResponse"),
        nondet request = oneOf(ABCI_recheckTxRequests(nodeId))
        all {
            val response = ABCI_responseFor(nodeId, request)
            all {
                require(inMempool(nodeId, request.tx)),
                ABCI_RemoveRequest(nodeId, request),
                if (response.err != noError) all {
                    // Tx became invalidated due to newly committed block.
                    removeFromMempool(nodeId, Set(request.tx)),
                    removeFromCache(nodeId, request.tx),
                } else all {
                    mempool' = mempool,
                    cache' = cache,
                }
            }
        },
        setError(nodeId, noError),
        height' = height,
        Network_Unchanged,
        Chain_Unchanged,
    }

    //--------------------------------------------------------------------------
    // P2P Reactor actions
    //--------------------------------------------------------------------------
    /* Receive a Txs message from a peer, and calls CheckTx on each Tx. */
    // [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
    action P2P_ReceiveTxs(nodeId) = all {
        setStep(nodeId, "P2P_ReceiveTxs"),
        require(Network_incomingMsgs(nodeId) != Set()),
        nondet msg = oneOf(Network_incomingMsgs(nodeId))
        all {
            CheckTx(nodeId, msg.tx, msg.senderId),
            receiveMsg(nodeId, msg),
            Chain_Unchanged,
        }
    }

    /* The reactor loops through its mempool and sends the tx one by one to all
     * each of its peers. */
    action P2P_SendTx_(nodeId: NodeId, peer: NodeId): bool = all {
        setStep(nodeId, "P2P_SendTx"),
        require(not(mempoolIsEmpty(nodeId))),
        nondet tx = oneOf(mempoolTxs(nodeId))
        val msg = { tag: "Txs", senderId: nodeId, tx: tx }
        all {
            // If the msg was not already sent to this peer.
            Network_incomingMsgs(peer).forall(m => m != msg),
            // If the peer is not a tx's sender.
            not(peer.in(sendersFor(nodeId, tx))),
            sendTo(msg, peer),
            mempool' = mempool,
            cache' = cache,
            height' = height,
            setError(nodeId, noError),
            ABCI_Unchanged,
            Chain_Unchanged,
        }   
    }

    // [Reactor.broadcastTxRoutine]: https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
    action P2P_SendTx(nodeId: NodeId): bool =
        nondet peer = oneOf(Peers.get(nodeId))
        P2P_SendTx_(nodeId, peer)

    //--------------------------------------------------------------------------
    // The ABCI application process a request and generates a response
    action ABCI_ProcessRequest(nodeId) = all {
        setStep(nodeId, "ABCI::ProcessCheckTxRequest"),
        ABCI_ProcessCheckTxRequest(nodeId),
        mempool' = mempool, 
        height' = height, 
        cache' = cache, 
        setError(nodeId, noError),
        Network_Unchanged,
        Chain_Unchanged,
    }

    /* All possible actions on a node. */
    action NodeNext =
        nondet nodeId = oneOf(NodeIds)
        any {
            // Receive a transaction from a client via RPC
            CheckTxRPC(nodeId),

            // Receive a (New) CheckTx response from the application
            ReceiveCheckTxResponse(nodeId),

            // Consensus reactor updates the mempool and rechecks txs
            Consensus_Update(nodeId),

            // Receive a (Recheck) CheckTx response from the application
            ReceiveRecheckTxResponse(nodeId),

            // Receive a Txs message from a peer
            P2P_ReceiveTxs(nodeId),

            // Send a transaction in the mempool to a peer
            P2P_SendTx(nodeId),

            // The ABCI application process a request and generates a response
            ABCI_ProcessRequest(nodeId),
        }

    /* There may be action in other nodes too. */
    action ChainNext =
        nondet txs = oneOf(Txs.exclude(Chain_allTxsInChain).powerset().setRemove(Set()))
        all {
            require(txs.forall(tx => isValid(tx))),
            Chain_NewBlockFrom(txs),
            mempool' = mempool, 
            height' = height, 
            cache' = cache, 
            setError(noNode, noError),
            ABCI_Unchanged,
            Network_Unchanged,
        }

    //--------------------------------------------------------------------------
    // Tests
    //--------------------------------------------------------------------------
    action assert_(__cond: bool): bool = all {
        __cond,
        mempool' = mempool, 
        height' = height, 
        cache' = cache,
        step' = step, 
        error' = error,
        ABCI_Unchanged,
        Network_Unchanged,
        Chain_Unchanged,
    }

    // action assert__(__action: bool, __cond: bool): bool =
    //     __action.then(assert_(__cond))

    run test1 = Init
        .then(CheckTxRPC_("n1", "tx1"))
        .then(ABCI_ProcessRequest("n1"))
        .then(ReceiveCheckTxResponse("n1"))
        .then(assert_(inMempool("n1", "tx1")))
        .then(assert_(inCache("n1", "tx1")))
    
    run test2 = test1
        .then(Update("n1", 1, Set("tx1"), Map("tx1" -> true)))
        .then(P2P_SendTx_("n1", "n2"))
        .then(P2P_ReceiveTxs("n2"))
        .then(ABCI_ProcessRequest("n2"))
        .then(ReceiveCheckTxResponse("n2"))
        .then(assert_(inMempool("n2", "tx1")))
        .then(assert_(inCache("n2", "tx1")))

    run test3 = test1
        .then(Update("n1", 1, Set("tx1"), Map("tx1" -> false)))

    run test4 = test1
        .then(Update("n1", 1, Set("tx2"), Map("tx2" -> true)))

    run testInvalid1 = Init
        .then(CheckTxRPC_("n1", invalidTx))
        .then(ABCI_ProcessRequest("n1"))
        .then(ReceiveCheckTxResponse("n1"))

    //--------------------------------------------------------------------------
    def nodeState(nodeId) = (
        ("mempool", mempool.get(nodeId)), 
        ("cache", cache.get(nodeId)),
        ("height", height.get(nodeId)),
        ("abci requests", requestSenderResponses.get(nodeId)),
        ("p2p incoming msgs", msgs.get(nodeId)),
    )
    def ns(nodeId) = nodeState(nodeId)

    def state = (
        ("step", step),
        ("error", error),
        ("chain", chain),
    )
}

module __repl__ {
    import Base.*
    import ABCIServers.*
    import Chain.*
    import Network.*
    import MempoolV0.*
} //-- __repl__
