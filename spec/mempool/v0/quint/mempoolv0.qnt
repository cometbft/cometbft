// -*- mode: Bluespec; -*-
/** Mempool V0
  * 
  */
// Assumption: The network topology is fixed: nodes do not leave or join the
// network, peers do not change.

// One of the goals of this spec is to make their actions and data structures
// easily mapped to the code.

module mempoolv0 {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"

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
    val Steps: Set[Step] = Set("init", "CheckTx", "ReceiveCheckTxResponse", 
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
        height: Height, // height at which tx was validated
        senders: Set[NodeId] // peers that send tx, or none if tx was emmitted by a client
    }
    
    var mempool: NodeId -> Set[MempoolTx]
    
    // The transactions received by a node.
    var cache: NodeId -> Set[Tx]
    
    // The last block Update()'d to.
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

    action addToMempool(nodeId, tx, h, senderId) = all {
        val senders = if (senderId == noNode) Set() else Set(senderId)
        val memTx = { tx: tx, height: h, senders: senders }
        mempool' = mempool.put(nodeId, mempool.get(nodeId).setAdd(memTx)),
    }

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
    action init = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        cache' = NodeIds.mapBy(_ => Set()),
        step' = { node: noNode, name: "init" },
        error' = (noNode, noError),
        height' = NodeIds.mapBy(_ => 0),
        Network_init,
        Chain_init,
        ABCI_init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    action CheckTx(nodeId: NodeId, tx: Tx, senderId: NodeId): bool = all {
        if (mempoolIsFull(nodeId)) all {
            setError(nodeId, MempoolIsFullError),
            mempool' = mempool,
            cache' = cache,
            ABCI_unchanged,
        } else all {
            if (inCache(nodeId, tx)) all {
        		// Record new sender for the tx we've already seen.
                // Note it's possible a tx is still in the cache but no longer in the mempool
                // (eg. after committing a block, txs are removed from mempool but not cache),
                // so we only record the sender for txs still in the mempool.
                addSender(nodeId, tx, senderId),
                setError(nodeId, TxInCacheError),
                cache' = cache,
                ABCI_unchanged,
            } else all {
                addToCache(nodeId, tx),
                ABCI_SendRequestNewCheckTx(nodeId, tx, senderId, height.get(nodeId) + 1),
                setError(nodeId, noError),
                mempool' = mempool,
            }
        },
        height' = height,
    }

    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    action CheckTx_RPC(nodeId, tx) = all {
        setStep(nodeId, "CheckTx"),
        CheckTx(nodeId, tx, noNode),
        Network_unchanged,
        Chain_unchanged,
    }

    /* Receive some transaction from a client via RPC. */
    action CheckTxRPC(nodeId) =
        nondet tx = oneOf(Txs)
        CheckTx_RPC(nodeId, tx)

    // Callback function to process a CheckTx response.
    // Note that tx and sender are arguments to CListMempool.resCbFirstTime.
    // [CListMempool.resCbFirstTime]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L369
    action ReceiveCheckTxResponse(nodeId) = all {
        setStep(nodeId, "ReceiveCheckTxResponse"),
        ABCI_checkTxRequests(nodeId).nonEmpty(),
        nondet request = oneOf(ABCI_checkTxRequests(nodeId)) all {
            val response = ABCI_responseFor(nodeId, request)
            all {
                ABCI_RemoveRequest(nodeId, request),
                if (response.err == noError) {
                    if (mempoolIsFull(nodeId)) all {
                        forceRemoveFromCache(nodeId, request.tx), // There may be enough space for tx later.
                        setError(nodeId, MempoolIsFullError),
                        mempool' = mempool,
                    } else all { 
                        require(not(inMempool(nodeId, request.tx))),
                        val senderId = ABCI_senderFor(nodeId, request)
                        val h = ABCI_heightFor(nodeId, request) // height at which tx was validated
                        addToMempool(nodeId, request.tx, h, senderId),
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
        Network_unchanged,
        Chain_unchanged,
    }

    /** Consensus' BlockExecutor calls Update to update the mempool after executing txs.
      * txResults are the results of ResponseFinalizeBlock for every tx in txs.
      * BlockExecutor holds the mempool lock while calling this function.
      */
    // [CListMempool.Update]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
    action Update(nodeId: NodeId, h: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        setStep(nodeId, "Update"),
        require(h > height.get(nodeId)),
        require(txs.nonEmpty()),

        // Update mempool's height.
        height' = height.put(nodeId, h),

        // Remove all txs from the mempool.
        removeFromMempool(nodeId, txs),

        // Update cache for all transactions.
        // - Add valid committed txs to the cache (in case they are missing).
        //      It may happen that tx has never been seen by this node before,
        //      so we add it to the cache just for this case.
        // - Remove invalid txs, if keepInvalidTxsInCache is false.
        //      If we know that tx is invalid, we need to remove it from the cache.
        val validTxs = txs.filter(tx => txValidResults.get(tx))
        val invalidTxs = txs.filter(tx => not(txValidResults.get(tx)) and not(Configs.get(nodeId).keepInvalidTxsInCache)) 
        cache' = cache.put(nodeId, cache.get(nodeId).union(validTxs).exclude(invalidTxs)),

        // Either recheck non-committed txs to see if they became invalid
        // or just notify there're some txs left.
        val txsStillInMempool = mempoolTxs(nodeId).exclude(txs)
        if (txsStillInMempool.nonEmpty()) {
            // CHECK: in the code, ReceiveRecheckTxResponse may be called concurrently.
            ABCI_SendRequestRecheckTxs(nodeId, txsStillInMempool)
        } else ABCI_unchanged,

        setError(nodeId, noError),
        Network_unchanged,
        Chain_unchanged,
    }

    // Consensus reactor updates the mempool and rechecks outstanding txs.
    action Consensus_Update(nodeId: NodeId): bool = all {
        require(not(Chain_isEmpty)),
        require(height.get(nodeId) <= Chain_latestHeight),
        val h = height.get(nodeId) + 1
        any {
            all {
                // this node is not the proposer: take the block from the chain
                val txs: Set[Tx] = Chain_getBlock(height.get(nodeId))
                val txValidResults: Tx -> bool = txs.mapBy(tx => isValidAt(tx, h)) // all txs should be valid because they are committed
                Update(nodeId, h, txs, txValidResults),
            },
            all {
                // this block is the proposer: take the block from the mempool
                nondet txs: Set[Tx] = oneOf(mempoolTxs(nodeId).powerset().setRemove(Set()))
                val txValidResults: Tx -> bool = txs.mapBy(tx => isValidAt(tx, h))
                Update(nodeId, h, txs, txValidResults),
            },
        }
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
        Network_unchanged,
        Chain_unchanged,
    }

    //--------------------------------------------------------------------------
    // P2P Reactor actions
    //--------------------------------------------------------------------------
    /* Receive a Txs message from a peer, and calls CheckTx on each Tx. */
    // [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
    action P2P_ReceiveTxs(nodeId) = all {
        setStep(nodeId, "P2P_ReceiveTxs"),
        require(Network_incomingMsgs(nodeId).nonEmpty()),
        nondet msg = oneOf(Network_incomingMsgs(nodeId))
        all {
            CheckTx(nodeId, msg.tx, msg.senderId),
            receiveMsg(nodeId, msg),
            Chain_unchanged,
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
            setError(nodeId, noError),
            mempool' = mempool,
            cache' = cache,
            height' = height,
            ABCI_unchanged,
            Chain_unchanged,
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
        setError(nodeId, noError),
        mempool' = mempool, 
        height' = height, 
        cache' = cache, 
        Network_unchanged,
        Chain_unchanged,
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
            require(txs.forall(tx => isValidAt(tx, Chain_latestHeight + 1))),
            Chain_newBlockFrom(txs),
            mempool' = mempool, 
            height' = height, 
            cache' = cache, 
            setError(noNode, noError),
            ABCI_unchanged,
            Network_unchanged,
        }
}

module tests {
    import base.* from "./base"
    import basicSpells.* from "./basicSpells"
    // import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import mempoolv0.* //from "./mempoolv0"

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
        ABCI_unchanged,
        Network_unchanged,
        Chain_unchanged,
    }

    // action assert__(__action: bool, __cond: bool): bool =
    //     __action.then(assert_(__cond))

    run test1 = init
        .then(assert_(isValidAt("tx1", 1)))
        .then(CheckTx_RPC("n1", "tx1"))
        .then(ABCI_ProcessRequest("n1"))
        .then(ReceiveCheckTxResponse("n1"))
        .then(assert_(inMempool("n1", "tx1") and inCache("n1", "tx1")))
    
    run test2 = test1
        .then(Update("n1", 1, Set("tx1"), Map("tx1" -> true)))
        .then(assert_(not(inMempool("n1", "tx1")) and inCache("n1", "tx1")))

    run test3 = test1
        .then(P2P_SendTx_("n1", "n2"))
        // .then(Update("n1", 1, Set("tx1"), Map("tx1" -> true)))
        // .then(assert_(not(inMempool("n1", "tx1")) and inCache("n1", "tx1")))
        .then(P2P_ReceiveTxs("n2"))
        .then(assert_(not(inMempool("n2", "tx1")) and inCache("n2", "tx1")))

    run test4 = test3
        .then(ABCI_ProcessRequest("n2"))
        .then(ReceiveCheckTxResponse("n2"))
        .then(assert_(inMempool("n2", "tx1") and inCache("n2", "tx1")))

    run test5 = test4
        .then(Update("n2", 1, Set("tx1"), Map("tx1" -> "tx1".isValidAt(1))))
        .then(assert_(not(inMempool("n2", "tx1")) and inCache("n2", "tx1")))

    run test8 = test1
        .then(Update("n1", 2, Set("tx1"), Map("tx1" -> "tx1".isValidAt(2))))

    run test9 = test1
        .then(Update("n1", 1, Set("tx2"), Map("tx2" -> true)))

    run testInvalid1 = init
        .then(CheckTx_RPC("n1", "tx3")) // tx3 is invalid at height 0 according to isValidAt
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
    val s1 = ns("n1")
    val s2 = ns("n2")
    val s3 = ns("n3")

    def state = (
        ("step", step),
        ("error", error),
        ("chain", chain),
    )
}

module __repl__ {
    import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import mempoolv0.* //from "./mempoolv0"
    // import properties.* 
    import tests.* 
} //-- __repl__
