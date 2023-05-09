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

    // Errors
    type Error = str
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
        error: Error // called `code` in protobuf
    }
}

// An ABCI server receives asynchronous ABCI requests and replies with ABCI responses.
module ABCIServers {
    import Base.*
    import ABCIMessages.*

    val NoResponse: Response = {tag: "NoResponse", error: NoError}

    type ABCIRequests = {tag: str, tx: Tx}
    val ABCIRequestTags = Set("CheckTx", "RecheckTx")
    
    // The ABCIServer receives requests in a queue (a set actually) and responds
    // to a queue (set) of responses.
    var requestResponses: NodeId -> (Request -> Response)
    var requestSenders: NodeId -> Request -> NodeId

    //--------------------------------------------------------------------------
    // Initial state and actions
    //--------------------------------------------------------------------------

    action ABCI_Init = all {
        requestResponses' = NodeIds.mapBy(_ => Set().mapBy(_ => NoResponse)),
        requestSenders' = NodeIds.mapBy(_ => Set().mapBy(_ => NoNode)),
    }

    // The app receives a CheckTx request and creates a response.
    action ABCI_ProcessCheckTxRequest(nodeId) = {
        nondet req: Request = oneOf(requestResponses.get(nodeId).keys()) 
        all {
            // req.tag.in(Set("CheckTx", "RecheckTx")),
            requestResponses.get(nodeId).get(req) == NoResponse,
            val error = if (isValid(req.tx)) NoError else InvalidTxError
            val resp = {tag: req.tag, error: error}
            requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(req, resp)),
            requestSenders' = requestSenders,
        }
    }

    action ABCI_SendRequestNewCheckTx(nodeId, tx, sender) = 
        val req = {tag: "CheckTx", tx: tx, checkTxType: "New"}
        all {
            requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(req, NoResponse)),
            requestSenders' = requestSenders.put(nodeId, requestSenders.get(nodeId).put(req, sender)),
        }

    action ABCI_SendRequestRecheckTxs(nodeId, txs) = all {
        val reqs = txs.map(tx => {tag: "CheckTx", tx: tx, checkTxType: "Recheck"})
        requestResponses' = requestResponses.put(nodeId, {
            val m = requestResponses.get(nodeId)
            m.keys().mapBy(req => if (req.in(reqs)) NoResponse else m.get(req))
        }),
        requestSenders' = requestSenders,
    }

    action ABCI_RemoveRequest(nodeId, request) = all {
        requestResponses.get(nodeId).get(request) != NoResponse,
        requestResponses' = requestResponses.put(nodeId, requestResponses.get(nodeId).put(request, NoResponse)), // TODO: remove entry instead of setting to NoResponse
        requestSenders' = requestSenders.put(nodeId, requestSenders.get(nodeId).put(request, NoNode)), // TODO: remove entry instead of setting to NoNode
    }

    action ABCI_Unchanged = all {
        requestResponses' = requestResponses,
        requestSenders' = requestSenders,
    }

    //--------------------------------------------------------------------------
    // Auxiliary definitions
    //--------------------------------------------------------------------------

    def ABCI_ResponseFor(nodeId, request) = requestResponses.get(nodeId).get(request)
    def ABCI_SenderFor(nodeId, request) = requestSenders.get(nodeId).get(request)

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
    type TxsMsg = {sender: NodeId, tx: Tx}
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

    // const MempoolMaxSize: int
    val MempoolMaxSize: int = 2
    // assume(MempoolMaxSize > 0)

    type Config = {keepInvalidTxsInCache: bool}
    // const Configs: NodeId -> Config
    // Configs = Map("n1" -> {keepInvalidTxsInCache -> false}, "n2" -> {keepInvalidTxsInCache -> false}, "n2" -> {keepInvalidTxsInCache -> false}), 
    val Configs: NodeId -> Config = NodeIds.mapBy(_ => {keepInvalidTxsInCache: false})
    
    type Step = str
    val Steps: Set[Step] = Set("Init", "CheckTx", "ReceiveCheckTxResponse", 
        "Update", "ReceiveRecheckTxResponse", "P2P_ReceiveTxs", "P2P_SendTx", 
        "ABCI::ProcessCheckTxRequest")
    
    //--------------------------------------------------------------------------
    // Nodes state
    var mempools: NodeId -> Set[Tx]
    var senders: NodeId -> Tx -> Set[NodeId]
    var caches: NodeId -> Set[Tx]
    var step: NodeId -> Step
    var errors: NodeId -> Error
    var chainHeight: NodeId -> Height

    //--------------------------------------------------------------------------
    /* Sender auxiliary definitions */
    action addSender(nodeId, tx, sender) = {
        val sendersTxMap = senders.get(nodeId)
        val senderIds = sendersTxMap.get(tx)
        senders' = senders.put(nodeId, sendersTxMap.put(tx, senderIds.union(Set(sender))))
    }

    action removeSenders(nodeId, txs) = {
        val txSendersMap: Tx -> Set[NodeId] = senders.get(nodeId)
        val updatedKeys = txSendersMap.keys().exclude(txs)
        senders' = senders.put(nodeId, updatedKeys.mapBy(tx => txSendersMap.get(tx)))
    }

    //--------------------------------------------------------------------------
    /* Cache auxiliary definitions */
    def inCache(nodeId, tx) =
        caches.get(nodeId).contains(tx)

    action addToCache(nodeId, tx) =
        caches' = caches.put(nodeId, caches.get(nodeId).union(Set(tx)))

    action forceRemoveFromCache(nodeId: NodeId, tx: Tx): bool =
        caches' = caches.put(nodeId, caches.get(nodeId).exclude(Set(tx)))

    action removeFromCache(nodeId, tx) =
        if (Configs.get(nodeId).keepInvalidTxsInCache) {
            forceRemoveFromCache(nodeId, tx)
        } else caches' = caches

    //--------------------------------------------------------------------------
    /* Mempool auxiliary definitions */    
    def mempoolIsEmpty(nodeId) =
        mempools.get(nodeId) == Set()

    def mempoolIsFull(nodeId) =
        size(mempools.get(nodeId)) > MempoolMaxSize

    def inMempool(nodeId, tx) =
        tx.in(mempools.get(nodeId))

    //TODO? add also height
    action addToMempool(nodeId, tx, sender) = all {
        mempools' = mempools.put(nodeId, mempools.get(nodeId).union(Set(tx))),
        addSender(nodeId, tx, sender),
    }

    action removeFromMempool(nodeId: NodeId, txs: Set[Tx]): bool = all {
        mempools' = mempools.put(nodeId, mempools.get(nodeId).exclude(txs)),
        removeSenders(nodeId, txs),
    }

    //--------------------------------------------------------------------------
    action setError(nodeId, err) =
        errors' = errors.put(nodeId, err)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action Init = all {
        mempools' = NodeIds.mapBy(_ => Set()),
        senders' = NodeIds.mapBy(_ => Map()),
        caches' = NodeIds.mapBy(_ => Set()),
        step' = NodeIds.mapBy(_ => "Init"),
        errors' = NodeIds.mapBy(_ => NoError),
        msgs' = NodeIds.mapBy(_ => Set()),
        chainHeight' = NodeIds.mapBy(_ => 0),
        ABCI_Init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    action CheckTx(nodeId: NodeId, tx: Tx, senderId: NodeId): bool = all {
        mempools' = mempools,
        if (mempoolIsFull(nodeId)) all {
            senders' = senders,
            caches' = caches,
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
                    senders' = senders
                },
                caches' = caches,
                setError(nodeId, "ErrTxInCache"),
                ABCI_Unchanged,
            } else all {
                senders' = senders,
                addToCache(nodeId, tx),
                setError(nodeId, NoError),
                ABCI_SendRequestNewCheckTx(nodeId, tx, senderId),
            }
        },
        chainHeight' = chainHeight,
    }


    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    action CheckTxRPC_(nodeId, tx) = all {
        step' = step.put(nodeId, "CheckTx"),
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
        step' = step.put(nodeId, "ReceiveCheckTxResponse"),
        nondet request = oneOf(ABCI_CheckRequests(nodeId)) all {
            val response = ABCI_ResponseFor(nodeId, request)
            val senderId = ABCI_SenderFor(nodeId, request)
            all {
                ABCI_RemoveRequest(nodeId, request),
                if (response.error == NoError) {
                    if (mempoolIsFull(nodeId)) all {
                        mempools' = mempools,
                        senders' = senders,
                        forceRemoveFromCache(nodeId, request.tx),
                        setError(nodeId, "ErrMempoolIsFull"),
                    } else all {
                        addToMempool(nodeId, request.tx, senderId), //TODO? add also height
                        caches' = caches,
                        setError(nodeId, NoError),
                    }
                } else all { // ignore invalid transaction
                    mempools' = mempools,
                    senders' = senders,
                    removeFromCache(nodeId, request.tx),
                    setError(nodeId, NoError),
                }
            }
        },
        msgs' = msgs,
        chainHeight' = chainHeight,
    }

    /** Consensus' BlockExecutor calls Update to update the mempool after executing txs.
      * txResults are the results of ResponseFinalizeBlock for every tx in txs.
      * BlockExecutor holds the mempool lock while calling this function.
      */
    // [CListMempool.Update] https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
    action Update(nodeId: NodeId, height: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        step' = step.put(nodeId, "Update"),
        txs != Set(),

        chainHeight' = chainHeight.put(nodeId, height),
        
        // Remove all txs from the mempool.
        removeFromMempool(nodeId, txs),

        // Update cache for all transactions.
        // - Add valid committed txs to the cache (in case they are missing).
        // - Remove invalid txs, if keepInvalidTxsInCache is false.
        val validTxs = txs.filter(tx => txValidResults.get(tx))
        val invalidTxs = txs.filter(tx => not(txValidResults.get(tx)) and not(Configs.get(nodeId).keepInvalidTxsInCache)) 
        caches' = caches.put(nodeId, caches.get(nodeId).union(validTxs).exclude(invalidTxs)),

        // Either recheck non-committed txs to see if they became invalid
        // or just notify there're some txs left.
        if (not(mempoolIsEmpty(nodeId))) {
            // NOTE: globalCb may be called concurrently.
            ABCI_SendRequestRecheckTxs(nodeId, txs)
        } else {
            ABCI_Unchanged
        },

        msgs' = msgs,
        errors' = errors,
    }

    // Callback function to process a CheckTx response, when rechecking.
    // [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L432
    action ReceiveRecheckTxResponse(nodeId: NodeId): bool = all {
        step' = step.put(nodeId, "ReceiveRecheckTxResponse"),
        nondet request = oneOf(ABCI_RecheckRequests(nodeId))
        all {
            val response = ABCI_ResponseFor(nodeId, request)
            all {
                inMempool(nodeId, request.tx),
                ABCI_RemoveRequest(nodeId, request),
                if (response.error == NoError) all {
                    // Tx became invalidated due to newly committed block.
                    removeFromMempool(nodeId, Set(request.tx)),
                    removeFromCache(nodeId, request.tx),
                } else all {
                    mempools' = mempools,
                    senders' = senders,
                    caches' = caches,
                }
            }
        },
        errors' = errors,
        msgs' = msgs,
        chainHeight' = chainHeight,
    }

    //--------------------------------------------------------------------------
    // P2P Reactor actions
    //--------------------------------------------------------------------------

    /* Receive a Txs message from a peer, and calls CheckTx on each Tx (only one 
     * for now). */
    // [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
    action P2P_ReceiveTxs(nodeId) = all {
        step' = step.put(nodeId, "P2P_ReceiveTxs"),
        msgs.get(nodeId) != Set(),
        nondet msg = oneOf(msgs.get(nodeId))
        all {
            CheckTx(nodeId, msg.tx, msg.sender),
            ReceiveMsg(nodeId, msg),
        }
    }

    /* The reactor loops through its mempool and sends the tx one by one to all
     * each of its peers. */
    // [Reactor.broadcastTxRoutine] https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
    action P2P_SendTx(nodeId: NodeId): bool = all {
        step' = step.put(nodeId, "P2P_SendTx"),
        mempools.get(nodeId) != Set(),
        nondet peer = oneOf(Peers.get(nodeId))
        nondet tx = oneOf(mempools.get(nodeId))
        val msg = {sender: nodeId, tx: tx}
        all {
            // If the msg was not already sent to this peer.
            msgs.get(peer).forall(m => m != msg),
            // If the peer is not a tx's sender.
            tx.in(senders.get(nodeId).keys()) implies not(peer.in(senders.get(nodeId).get(tx))),
            SendTo(msg, peer),
            mempools' = mempools,
            caches' = caches,
            senders' = senders,
            errors' = errors,
            chainHeight' = chainHeight,
            ABCI_Unchanged,
        }
    }

    //--------------------------------------------------------------------------
    // The ABCI application process a request and generates a response
    action ABCIProcessRequest(nodeId) = all {
        step' = step.put(nodeId, "ABCI::ProcessCheckTxRequest"),
        ABCI_ProcessCheckTxRequest(nodeId),
        mempools' = mempools, 
        senders' = senders, 
        caches' = caches, 
        errors' = errors, 
        chainHeight' = chainHeight, 
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
                        
            // Consensus reactor updates the mempool
            nondet txs = oneOf(Txs.powerset().exclude(Set(Set())))
            nondet txValidResults = oneOf(txs.setOfMaps(Bool))
            Update(nodeId, chainHeight.get(nodeId) + 1, txs, txValidResults),

            // Receive a (Recheck) CheckTx response from the application
            ReceiveRecheckTxResponse(nodeId),

            // Receive a Txs message from a peer
            P2P_ReceiveTxs(nodeId),
            
            // Send a transaction in the mempool to a peer
            P2P_SendTx(nodeId),

            // The ABCI application process a request and generates a response
            ABCIProcessRequest(nodeId),
        }

    //--------------------------------------------------------------------------
    run Test1 = Init
        .then(CheckTxRPC_("n1", "tx1"))
        .then(ABCIProcessRequest("n1"))
        .then(ReceiveCheckTxResponse("n1"))
        .then(Update("n1", 1, Set("tx1"), Map("tx1" -> true)))

    //--------------------------------------------------------------------------
    def NodeState(nodeId) = (
        ("mempool", mempools.get(nodeId)), 
        ("sender", senders.get(nodeId)),
        ("cache", caches.get(nodeId)),
        ("step", step.get(nodeId)),
        ("error", errors.get(nodeId)),
        ("chainHeight", chainHeight.get(nodeId)),
    )

}
