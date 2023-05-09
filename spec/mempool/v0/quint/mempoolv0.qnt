/** Mempool V0
  * 
  */
// Assumption: The network topology is fixed: nodes do not leave or join the
// network, peers do not change.

// One of the goals of this spec is to make their actions and data structures
// easily mapped to the code.

module Base {
    // Nodes
    type NodeId = str
    val NoNode: NodeId = "no-node"
    // const NodeIds: Set[NodeId]
    val NodeIds: Set[NodeId] = Set(NoNode, "n1", "n2", "n3")

    // Transactions
    type Tx = str
    val InvalidTx: Tx = "invalid"
    // const Txs: Set[Tx]
    val Txs: Set[Tx] = Set("tx1", "tx2", "invalid")
    def isValid(tx) = tx == InvalidTx
    type Error = str

    // Error
    val NoError: Error = "none"

    // Heights
    type Height = int
    val Heights = Set(0, 1, 2, 3, 4)
}

module ABCIMessages {
    import Base.*

    val InvalidTxError: Error = "invalid-tx"

    // https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L94
    val CheckTxTypes = Set("New", "Recheck")

    // https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L99
    type Request = {
        tag: str, 
        tx: Tx, 
        checkTxType: str
    }

    // https://github.com/CometBFT/cometbft/blob/4790ea3e46475064d5475c787427ae926c5a9e94/proto/tendermint/abci/types.proto#L254
    type Response = {
        tag: str, 
        err: Error // called `code` in protobuf
    }
}

// An ABCI server receives asynchronous ABCI requests and replies with ABCI responses.
module ABCIServers {
    import Base.*
    import ABCIMessages.*

    val NoResponse: Response = {tag: "NoResponse", err: NoError}

    type ABCIRequests = {tag: str, tx: Tx}
    val ABCIRequestTags = Set("CheckTx", "RecheckTx")
    
    // The ABCIServer receives requests in a queue (a set actually) and responds
    // to a queue (set) of responses.
    var requestResponses: NodeId -> (Request -> Response)
    var requestSender: NodeId -> Request -> NodeId

    //--------------------------------------------------------------------------
    // Initial state and actions
    //--------------------------------------------------------------------------

    action ABCI_Init = all {
        requestResponses' = NodeIds.mapBy(_ => Set().mapBy(_ => NoResponse)),
        requestSender' = NodeIds.mapBy(_ => Set().mapBy(_ => NoNode)),
    }

    // The app receives a CheckTx request and creates a response.
    action ABCI_ProcessCheckTxRequest(nodeId) = {
        nondet req: Request = oneOf(requestResponses.get(nodeId).keys()) 
        all {
            // req.tag.in(Set("CheckTx", "RecheckTx")),
            requestResponses.get(nodeId).get(req) == NoResponse,
            val err = if (isValid(req.tx)) NoError else InvalidTxError
            val resp = {tag: req.tag, err: err}
            requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(req, resp)),
            requestSender' = requestSender,
        }
    }

    action ABCI_SendRequestNewCheckTx(nodeId, tx, senderId) = 
        val req = {tag: "CheckTx", tx: tx, checkTxType: "New"}
        all {
            requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(req, NoResponse)),
            requestSender' = requestSender.put(nodeId, requestSender.get(nodeId).put(req, senderId)),
        }

    action ABCI_SendRequestRecheckTxs(nodeId, txs) = all {
        val reqs = txs.map(tx => {tag: "CheckTx", tx: tx, checkTxType: "Recheck"})
        requestResponses' = requestResponses.put(nodeId, {
            val m = requestResponses.get(nodeId)
            m.keys().mapBy(req => if (req.in(reqs)) NoResponse else m.get(req))
        }),
        requestSender' = requestSender,
    }

    action ABCI_RemoveRequest(nodeId, request) = all {
        requestResponses.get(nodeId).get(request) != NoResponse,
        requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(request, NoResponse)), // TODO: remove entry instead of setting to NoResponse
        requestSender' = requestSender.put(nodeId, requestSender.get(nodeId).put(request, NoNode)), // TODO: remove entry instead of setting to NoNode
    }

    action ABCI_Unchanged = all {
        requestResponses' = requestResponses,
        requestSender' = requestSender,
    }

    //--------------------------------------------------------------------------
    // Auxiliary definitions
    //--------------------------------------------------------------------------

    def ABCI_ResponseFor(nodeId, request) = requestResponses.get(nodeId).get(request)
    def ABCI_SenderFor(nodeId, request) = requestSender.get(nodeId).get(request)

    def Requests(nodeId: NodeId, checkTxType: str): Set[Request] =
        requestResponses.get(nodeId).keys()
            .filter(r => and {
                r.checkTxType == checkTxType,
                requestResponses.get(nodeId).get(r) != NoResponse
            })
    def ABCI_CheckRequests(nodeId) = Requests(nodeId, "New")
    def ABCI_RecheckRequests(nodeId) = Requests(nodeId, "Recheck")
}

module Network {
    import Base.* 

    // const Peers: NodeId -> Set[NodeId]
    val Peers: NodeId -> Set[NodeId] = Map(
        "n1" -> Set("n2", "n3"),
        "n2" -> Set("n1", "n3"),
        "n3" -> Set("n1", "n2")
    )

    //--------------------------------------------------------------------------
    // Network state
    type TxsMsg = {tag: str, senderId: NodeId, tx: Tx}
    var msgs: NodeId -> Set[TxsMsg] // map from receiver to messages to process

    //--------------------------------------------------------------------------
    /* Network auxiliary definitions */
    action SendTo(msg: TxsMsg, peer: NodeId): bool =
        msgs' = msgs.put(peer, msgs.get(peer).union(Set(msg)))

    action ReceiveMsg(nodeId: NodeId, msg: TxsMsg): bool =
        msgs' = msgs.put(nodeId, msgs.get(nodeId).exclude(Set(msg)))    
}

module MempoolV0 {
    import Base.* 
    import ABCIServers.* // as ABCI
    import Network.*

    //--------------------------------------------------------------------------
    // Spec parameters
    //--------------------------------------------------------------------------
    // const MempoolMaxSize: int
    val MempoolMaxSize: int = 2
    // assume(MempoolMaxSize > 0)

    type Config = {keepInvalidTxsInCache: bool}
    // const Configs: NodeId -> Config
    // Configs = Map("n1" -> {keepInvalidTxsInCache -> false}, "n2" -> {keepInvalidTxsInCache -> false}, "n2" -> {keepInvalidTxsInCache -> false}), 
    val Configs: NodeId -> Config = NodeIds.mapBy(_ => {keepInvalidTxsInCache: false})

    //--------------------------------------------------------------------------
    // Node metadata
    //--------------------------------------------------------------------------
    type Step = str
    val Steps: Set[Step] = Set("Init", "CheckTx", "ReceiveCheckTxResponse", 
        "Update", "ReceiveRecheckTxResponse", "P2P_ReceiveTxs", "P2P_SendTx", 
        "ABCI::ProcessCheckTxRequest")

    var step: NodeId -> Step
    var error: NodeId -> Error

    /* Auxiliary definitions */
    action setStep(nodeId, s) =
        step' = step.put(nodeId, s)

    action setError(nodeId, err) =
        error' = error.put(nodeId, err)

    //--------------------------------------------------------------------------
    // Mempool state
    //--------------------------------------------------------------------------
    // [mempoolTx]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L671
    type MempoolTx = {tx: Tx, height: Height}
    
    var mempool: NodeId -> Set[MempoolTx]
    var sender: NodeId -> Tx -> Set[NodeId]
    var cache: NodeId -> Set[Tx]
    var height: NodeId -> Height

    //--------------------------------------------------------------------------
    /* Sender auxiliary definitions */
    action addSender(nodeId, tx, senderId) = {
        val senderTxMap = sender.get(nodeId)
        val senderIds = senderTxMap.get(tx)
        sender' = sender.put(nodeId, senderTxMap.put(tx, senderIds.union(Set(senderId))))
    }

    action removeSender(nodeId, txs) = {
        val txSenderMap: Tx -> Set[NodeId] = sender.get(nodeId)
        val updatedKeys = txSenderMap.keys().exclude(txs)
        sender' = sender.put(nodeId, updatedKeys.mapBy(tx => txSenderMap.get(tx)))
    }

    //--------------------------------------------------------------------------
    /* Cache auxiliary definitions */
    def inCache(nodeId, tx) =
        cache.get(nodeId).contains(tx)

    action addToCache(nodeId, tx) =
        cache' = cache.put(nodeId, cache.get(nodeId).union(Set(tx)))

    action forceRemoveFromCache(nodeId: NodeId, tx: Tx): bool =
        cache' = cache.put(nodeId, cache.get(nodeId).exclude(Set(tx)))

    action removeFromCache(nodeId, tx) =
        if (Configs.get(nodeId).keepInvalidTxsInCache) {
            forceRemoveFromCache(nodeId, tx)
        } else cache' = cache

    //--------------------------------------------------------------------------
    /* Mempool auxiliary definitions */    
    def mempoolIsEmpty(nodeId) =
        mempool.get(nodeId) == Set()

    def mempoolIsFull(nodeId) =
        size(mempool.get(nodeId)) > MempoolMaxSize

    def mempoolTxs(nodeId) = 
        mempool.get(nodeId).map(e => e.tx)

    def inMempool(nodeId, tx) =
        tx.in(mempoolTxs(nodeId))

    action addToMempool(nodeId, tx, h, senderId) = all {
        mempool' = mempool.put(nodeId, mempool.get(nodeId).union(Set({tx: tx, height: h}))),
        addSender(nodeId, tx, senderId),
    }

    action removeFromMempool(nodeId: NodeId, txs: Set[Tx]): bool = all {
        mempool' = mempool.put(nodeId, mempool.get(nodeId).filter(e => e.tx.in(txs))),
        removeSender(nodeId, txs),
    }

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action Init = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        sender' = NodeIds.mapBy(_ => Map()),
        cache' = NodeIds.mapBy(_ => Set()),
        msgs' = NodeIds.mapBy(_ => Set()),
        step' = NodeIds.mapBy(_ => "Init"),
        error' = NodeIds.mapBy(_ => NoError),
        height' = NodeIds.mapBy(_ => 0),
        ABCI_Init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    action CheckTx(nodeId: NodeId, tx: Tx, senderId: NodeId): bool = all {
        mempool' = mempool,
        if (mempoolIsFull(nodeId)) all {
            sender' = sender,
            cache' = cache,
            setError(nodeId, "ErrMempoolIsFull"),
            ABCI_Unchanged,
        } else all {
            if (inCache(nodeId, tx)) all {
        		// Record new sender for the tx we've already seen.
                // Note it's possible a tx is still in the cache but no longer in the mempool
                // (eg. after committing a block, txs are removed from mempool but not cache),
                // so we only record the sender for txs still in the mempool.
                if (inMempool(nodeId, tx)) {
                    addSender(nodeId, tx, senderId)
                } else {
                    sender' = sender
                },
                cache' = cache,
                setError(nodeId, "ErrTxInCache"),
                ABCI_Unchanged,
            } else all {
                sender' = sender,
                addToCache(nodeId, tx),
                setError(nodeId, NoError),
                ABCI_SendRequestNewCheckTx(nodeId, tx, senderId),
            }
        },
        height' = height,
    }

    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    action CheckTxRPC_(nodeId, tx) = all {
        setStep(nodeId, "CheckTx"),
        CheckTx(nodeId, tx, NoNode),
        msgs' = msgs,
    }

    /* Receive some transaction from a client via RPC. */
    action CheckTxRPC(nodeId) = {
        nondet tx = oneOf(Txs)
        CheckTxRPC_(nodeId, tx)
    }

    // Callback function to process a CheckTx response.
    // Note that tx and sender are arguments to CListMempool.resCbFirstTime.
    // [CListMempool.resCbFirstTime]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L369
    action ReceiveCheckTxResponse(nodeId: NodeId): bool = all {
        setStep(nodeId, "ReceiveCheckTxResponse"),
        nondet request = oneOf(ABCI_CheckRequests(nodeId)) all {
            val response = ABCI_ResponseFor(nodeId, request)
            val senderId = ABCI_SenderFor(nodeId, request)
            all {
                ABCI_RemoveRequest(nodeId, request),
                if (response.err == NoError) {
                    if (mempoolIsFull(nodeId)) all {
                        forceRemoveFromCache(nodeId, request.tx), // There may be enough space for tx later.
                        setError(nodeId, "ErrMempoolIsFull"),
                        mempool' = mempool,
                        sender' = sender,
                    } else all { 
                        addToMempool(nodeId, request.tx, height.get(nodeId), senderId),
                        cache' = cache,
                        setError(nodeId, NoError),
                    }
                } else all { // ignore invalid transaction
                    mempool' = mempool,
                    sender' = sender,
                    removeFromCache(nodeId, request.tx),
                    setError(nodeId, NoError),
                }
            }
        },
        msgs' = msgs,
        height' = height,
    }

    /** Consensus' BlockExecutor calls Update to update the mempool after executing txs.
      * txResults are the results of ResponseFinalizeBlock for every tx in txs.
      * BlockExecutor holds the mempool lock while calling this function.
      */
    // [CListMempool.Update] https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
    action Update(nodeId: NodeId, h: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        setStep(nodeId, "Update"),
        txs != Set(),

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
        } else {
            ABCI_Unchanged
        },

        msgs' = msgs,
        error' = error,
    }

    // Callback function to process a CheckTx response, when rechecking.
    // [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L432
    action ReceiveRecheckTxResponse(nodeId: NodeId): bool = all {
        setStep(nodeId, "ReceiveRecheckTxResponse"),
        nondet request = oneOf(ABCI_RecheckRequests(nodeId))
        all {
            val response = ABCI_ResponseFor(nodeId, request)
            all {
                inMempool(nodeId, request.tx),
                ABCI_RemoveRequest(nodeId, request),
                if (response.err == NoError) all {
                    // Tx became invalidated due to newly committed block.
                    removeFromMempool(nodeId, Set(request.tx)),
                    removeFromCache(nodeId, request.tx),
                } else all {
                    mempool' = mempool,
                    sender' = sender,
                    cache' = cache,
                }
            }
        },
        msgs' = msgs,
        error' = error,
        height' = height,
    }

    //--------------------------------------------------------------------------
    // P2P Reactor actions
    //--------------------------------------------------------------------------

    /* Receive a Txs message from a peer, and calls CheckTx on each Tx. */
    // [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
    action P2P_ReceiveTxs(nodeId) = all {
        setStep(nodeId, "P2P_ReceiveTxs"),
        msgs.get(nodeId) != Set(),
        nondet msg = oneOf(msgs.get(nodeId))
        all {
            CheckTx(nodeId, msg.tx, msg.senderId),
            ReceiveMsg(nodeId, msg),
        }
    }

    /* The reactor loops through its mempool and sends the tx one by one to all
     * each of its peers. */
    // [Reactor.broadcastTxRoutine] https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
    action P2P_SendTx(nodeId: NodeId): bool = all {
        setStep(nodeId, "P2P_SendTx"),
        not(mempoolIsEmpty(nodeId)),
        nondet peer = oneOf(Peers.get(nodeId))
        nondet tx = oneOf(mempoolTxs(nodeId))
        val msg = {tag: "Txs", senderId: nodeId, tx: tx}
        all {
            // If the msg was not already sent to this peer.
            msgs.get(peer).forall(m => m != msg),
            // If the peer is not a tx's sender.
            tx.in(sender.get(nodeId).keys()) implies not(peer.in(sender.get(nodeId).get(tx))),
            SendTo(msg, peer),
            mempool' = mempool,
            cache' = cache,
            height' = height,
            sender' = sender,
            error' = error,
            ABCI_Unchanged,
        }
    }

    //--------------------------------------------------------------------------
    // The ABCI application process a request and generates a response
    action ABCI_ProcessRequest(nodeId) = all {
        setStep(nodeId, "ABCI::ProcessCheckTxRequest"),
        ABCI_ProcessCheckTxRequest(nodeId),
        mempool' = mempool, 
        sender' = sender, 
        height' = height, 
        cache' = cache, 
        error' = error, 
        msgs' = msgs,
    }

    /* All possible actions */
    action Next =
        nondet nodeId = oneOf(NodeIds)
        any {
            // Receive a transaction from a client via RPC
            CheckTxRPC(nodeId),

            // Receive a (New) CheckTx response from the application
            ReceiveCheckTxResponse(nodeId),

            // Consensus reactor updates the mempool and rechecks txs
            nondet txs = oneOf(Txs.powerset().exclude(Set(Set())))
            nondet txValidResults = oneOf(txs.setOfMaps(Bool))
            Update(nodeId, height.get(nodeId) + 1, txs, txValidResults),

            // Receive a (Recheck) CheckTx response from the application
            ReceiveRecheckTxResponse(nodeId),

            // Receive a Txs message from a peer
            P2P_ReceiveTxs(nodeId),

            // Send a transaction in the mempool to a peer
            P2P_SendTx(nodeId),

            // The ABCI application process a request and generates a response
            ABCI_ProcessRequest(nodeId),
        }

    //--------------------------------------------------------------------------
    run Test1 = Init
        .then(CheckTxRPC_("n1", "tx1"))
        .then(ABCI_ProcessRequest("n1"))
        .then(ReceiveCheckTxResponse("n1"))
        .then(Update("n1", 1, Set("tx1"), Map("tx1" -> true)))

    //--------------------------------------------------------------------------
    def NodeState(nodeId) = (
        ("step", step.get(nodeId)),
        ("error", error.get(nodeId)),
        ("mempool", mempool.get(nodeId)), 
        ("sender", sender.get(nodeId)),
        ("cache", cache.get(nodeId)),
        ("height", height.get(nodeId)),
    )

}
