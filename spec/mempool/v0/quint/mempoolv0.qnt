// -*- mode: Bluespec; -*-
/** Mempool V0
  *
  * One of the goals of this spec is to easily map actions and data structures
  * to the implementation. 
  */

// Assumptions: 
// - The network topology is fixed: nodes do not leave or join the network, peers do not change.

module mempoolv0 {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import history.* from "./history"

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
    // Auxiliary variables
    //--------------------------------------------------------------------------
    type Step = str
    val Steps: Set[Step] = Set("init", "CheckTx", "ReceiveCheckTxResponse", 
        "Update", "ReceiveRecheckTxResponse", "P2P_ReceiveTxs", "P2P_SendTx", 
        "ABCI::ProcessCheckTxRequest")

    var step: { node: NodeId, name: Step }

    action setStep(nodeId, s) =
        step' = { node: nodeId, name: s }

    var error: { node: NodeId, error: Error }

    action setError(nodeId, err) =
        error' = { node: nodeId, error: err }

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

    def Mempool(nodeId) = mempool.get(nodeId)
    def Cache(nodeId) = cache.get(nodeId)
    def Height(nodeId) = height.get(nodeId)

    //--------------------------------------------------------------------------
    /* Cache auxiliary definitions */
    def inCache(node, tx) =
        node.Cache().contains(tx)

    action addToCache(node, tx) =
        cache' = cache.mapPut(node, (txs) => txs.setAdd(tx))

    action forceRemoveFromCache(node: NodeId, tx: Tx): bool =
        cache' = cache.mapPut(node, (txs) => txs.setRemove(tx))

    action removeFromCache(node, invalidTx_) =
        if (not(Configs.get(node).keepInvalidTxsInCache))
            node.forceRemoveFromCache(invalidTx_)
        else cache' = cache

    //--------------------------------------------------------------------------
    /* Mempool auxiliary definitions */    
    def mempoolIsEmpty(node) =
        node.Mempool().isEmpty()

    def mempoolIsFull(node) =
        size(node.Mempool()) > MempoolMaxSize

    def mempoolTxs(node: NodeId): Set[Tx] = 
        node.Mempool().map(e => e.tx)

    def memTxFor(node: NodeId, tx: Tx): Set[MempoolTx] =
        node.Mempool().filter(e => e.tx == tx)
    
    def sendersFor(node: NodeId, tx: Tx): Set[NodeId] = 
        node.memTxFor(tx).map(e => e.senders).flatten()

    def inMempool(node, tx) =
        tx.in(node.mempoolTxs())

    def senderFor(node, tx) = 
        node.Mempool().map(e => e.tx)

    action addToMempool(node, tx, h, senderId) = all {
        val senders = if (senderId == noNode) Set() else Set(senderId)
        val memTx = { tx: tx, height: h, senders: senders }
        mempool' = mempool.mapPut(node, (memTxs) => memTxs.setAdd(memTx)),
        node.History_recordMempoolTx(tx),
    }

    action removeFromMempool(node: NodeId, txs: Set[Tx]): bool =
        mempool' = mempool.mapPut(node, (memTxs) => memTxs.filter(e => not(e.tx.in(txs))))

    action addSender(node, tx, senderId) =
        if (node.inMempool(tx)) {
            val oldMemTx: MempoolTx = node.memTxFor(tx).chooseSome()
            val newMemTx = oldMemTx.with("senders", oldMemTx.senders.union(senderId))
            mempool' = mempool.mapPut(node, (memTxs) => memTxs.setRemove(oldMemTx).setAdd(newMemTx))
        } else mempool' = mempool

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action init = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        cache' = NodeIds.mapBy(_ => Set()),
        step' = { node: noNode, name: "init" },
        error' = { node: noNode, error: noError },
        height' = NodeIds.mapBy(_ => 0),
        History_init,
        Network_init,
        Chain_init,
        ABCI_init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    action CheckTx(node: NodeId, tx: Tx, senderId: NodeId): bool = all {
        if (node.mempoolIsFull()) all {
            node.setError(MempoolIsFullError),
            mempool' = mempool,
            cache' = cache,
            ABCI_unchanged,
        } else all {
            if (node.inCache(tx)) all {
        		// Record new sender for the tx we've already seen.
                // Note it's possible a tx is still in the cache but no longer in the mempool
                // (eg. after committing a block, txs are removed from mempool but not cache),
                // so we only record the sender for txs still in the mempool.
                node.addSender(tx, senderId),
                node.setError(TxInCacheError),
                cache' = cache,
                ABCI_unchanged,
            } else all {
                node.addToCache(tx),
                node.ABCI_SendRequestNewCheckTx(tx, senderId, height.get(node) + 1),
                node.setError(noError),
                mempool' = mempool,
            }
        },
        height' = height,
    }

    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    action CheckTx_RPC(node, tx) = all {
        node.setStep("CheckTx"),
        node.CheckTx(tx, noNode),
        node.History_recordSubmittedTx(tx),
        Network_unchanged,
        Chain_unchanged,
    }

    /* Receive some transaction from a client via RPC. */
    action CheckTxRPC(node) =
        nondet tx = oneOf(Txs)
        node.CheckTx_RPC(tx)

    // Callback function to process a CheckTx response.
    // Note that tx and sender are arguments to CListMempool.resCbFirstTime.
    // [CListMempool.resCbFirstTime]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L369
    action ReceiveCheckTxResponse(node) = all {
        node.setStep("ReceiveCheckTxResponse"),
        node.ABCI_checkTxRequests().nonEmpty(),
        nondet request = oneOf(node.ABCI_checkTxRequests()) all {
            val response = node.ABCI_responseFor(request)
            all {
                node.ABCI_RemoveRequest(request),
                if (response.err == noError) {
                    if (node.mempoolIsFull()) all {
                        node.forceRemoveFromCache(request.tx), // There may be enough space for tx later.
                        node.setError(MempoolIsFullError),
                        mempool' = mempool,
                        History_unchanged,
                    } else all { 
                        require(not(node.inMempool(request.tx))),
                        val senderId = node.ABCI_senderFor(request)
                        val h = node.ABCI_heightFor(request) // height at which tx was validated
                        node.addToMempool(request.tx, h, senderId),
                        node.setError(noError),
                        cache' = cache,
                    }
                } else all { // ignore invalid transaction
                    node.removeFromCache(request.tx),
                    node.setError(InvalidTxWarning),
                    mempool' = mempool,
                    History_unchanged,
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
    action Update(node: NodeId, h: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        node.setStep("Update"),
        require(h > height.get(node)),
        require(txs.nonEmpty()),

        // Update mempool's height.
        height' = height.put(node, h),

        // Remove all txs from the mempool.
        node.removeFromMempool(txs),

        // Update cache for all transactions.
        // - Add valid committed txs to the cache (in case they are missing).
        //      It may happen that tx has never been seen by this node before,
        //      so we add it to the cache just for this case.
        // - Remove invalid txs, if keepInvalidTxsInCache is false.
        //      If we know that tx is invalid, we need to remove it from the cache.
        val validTxs = txs.filter(tx => txValidResults.get(tx))
        val invalidTxs = txs.filter(tx => not(txValidResults.get(tx)) and not(Configs.get(node).keepInvalidTxsInCache)) 
        cache' = cache.mapPut(node, (ts) => ts.union(validTxs).exclude(invalidTxs)),

        // Either recheck non-committed txs to see if they became invalid
        // or just notify there're some txs left.
        val txsStillInMempool = node.mempoolTxs().exclude(txs)
        if (txsStillInMempool.nonEmpty())
            // CHECK: in the code, ReceiveRecheckTxResponse may be called concurrently.
            node.ABCI_SendRequestRecheckTxs(txsStillInMempool)
        else ABCI_unchanged,

        node.setError(noError),
        History_unchanged,
        Network_unchanged,
        Chain_unchanged,
    }

    // Consensus reactor updates the mempool and rechecks outstanding txs.
    action Consensus_Update(node: NodeId): bool = all {
        require(not(Chain_isEmpty)),
        require(height.get(node) <= Chain_latestHeight),
        val h = height.get(node) + 1
        any {
            all {
                // this node is not the proposer: take the block from the chain
                val txs: Set[Tx] = Chain_getBlock(height.get(node))
                val txValidResults: Tx -> bool = txs.mapBy(tx => isValidAt(tx, h)) // all txs should be valid because they are committed
                node.Update(h, txs, txValidResults),
            },
            all {
                // this block is the proposer: take the block from the mempool
                nondet txs: Set[Tx] = oneOf(node.mempoolTxs().powerset().setRemove(Set()))
                val txValidResults: Tx -> bool = txs.mapBy(tx => isValidAt(tx, h))
                node.Update(h, txs, txValidResults),
            },
        }
    }

    // Callback function to process a CheckTx response, when rechecking.
    // [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L432
    action ReceiveRecheckTxResponse(node: NodeId): bool = all {
        node.setStep("ReceiveRecheckTxResponse"),
        nondet request = oneOf(node.ABCI_recheckTxRequests())
        all {
            val response = node.ABCI_responseFor(request)
            all {
                require(node.inMempool(request.tx)),
                node.ABCI_RemoveRequest(request),
                if (response.err != noError) all {
                    // Tx became invalidated due to newly committed block.
                    node.removeFromMempool(Set(request.tx)),
                    node.removeFromCache(request.tx),
                } else all {
                    mempool' = mempool,
                    cache' = cache,
                }
            }
        },
        node.setError(noError),
        height' = height,
        History_unchanged,
        Network_unchanged,
        Chain_unchanged,
    }

    //--------------------------------------------------------------------------
    // P2P Reactor actions
    //--------------------------------------------------------------------------
    /* Receive a Txs message from a peer, and calls CheckTx on each Tx. */
    // [Reactor.Receive]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/mempool/reactor.go#L93
    action P2P_ReceiveTxs(node) = all {
        node.setStep("P2P_ReceiveTxs"),
        require(node.Network_incomingMsgs().nonEmpty()),
        nondet msg = oneOf(node.Network_incomingMsgs())
        all {
            node.CheckTx(msg.tx, msg.senderId),
            node.receiveMsg(msg),
            History_unchanged,
            Chain_unchanged,
        }
    }

    /* The reactor loops through its mempool and sends the tx one by one to all
     * each of its peers. */
    action P2P_SendTx_(node: NodeId, peer: NodeId): bool = all {
        node.setStep("P2P_SendTx"),
        require(not(node.mempoolIsEmpty())),
        nondet tx = oneOf(node.mempoolTxs())
        val msg = { tag: "Txs", senderId: node, tx: tx }
        all {
            // If the msg was not already sent to this peer.
            Network_incomingMsgs(peer).forall(m => m != msg),
            // If the peer is not a tx's sender.
            not(peer.in(node.sendersFor(tx))),
            sendTo(msg, peer),
            node.setError(noError),
            mempool' = mempool,
            cache' = cache,
            height' = height,
            History_unchanged,
            ABCI_unchanged,
            Chain_unchanged,
        }
    }

    // [Reactor.broadcastTxRoutine]: https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
    action P2P_SendTx(node: NodeId): bool =
        nondet peer = oneOf(Peers.get(node))
        node.P2P_SendTx_(peer)

    //--------------------------------------------------------------------------
    // The ABCI application process a request and generates a response
    action ABCI_ProcessRequest(node) = all {
        node.setStep("ABCI::ProcessCheckTxRequest"),
        node.ABCI_ProcessCheckTxRequest(),
        node.setError(noError),
        mempool' = mempool, 
        height' = height, 
        cache' = cache, 
        History_unchanged,
        Network_unchanged,
        Chain_unchanged,
    }

    /* All possible actions on a node. */
    action NodeNext =
        nondet node = oneOf(NodeIds)
        any {
            // Receive a transaction from a client via RPC
            node.CheckTxRPC(),

            // Receive a (New) CheckTx response from the application
            node.ReceiveCheckTxResponse(),

            // Consensus reactor updates the mempool and rechecks txs
            node.Consensus_Update(),

            // Receive a (Recheck) CheckTx response from the application
            node.ReceiveRecheckTxResponse(),

            // Receive a Txs message from a peer
            node.P2P_ReceiveTxs(),

            // Send a transaction in the mempool to a peer
            node.P2P_SendTx(),

            // The ABCI application process a request and generates a response
            node.ABCI_ProcessRequest(),
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
            History_unchanged,
            ABCI_unchanged,
            Network_unchanged,
        }
}

module tests {
    import base.* from "./base"
    import basicSpells.* from "./basicSpells"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import history.* from "./history"
    import mempoolv0.* //from "./mempoolv0"

    val node1 = "n1"
    val node2 = "n2"
    val node3 = "n3"

    action assert_(__cond: bool): bool = all {
        __cond,
        mempool' = mempool, 
        height' = height, 
        cache' = cache,
        step' = step, 
        error' = error,
        History_unchanged,
        ABCI_unchanged,
        Network_unchanged,
        Chain_unchanged,
    }

    // action assert__(__action: bool, __cond: bool): bool =
    //     __action.then(assert_(__cond))

    //--------------------------------------------------------------------------
    // Tests
    //--------------------------------------------------------------------------
    run test1 = init
        .then(assert_("tx1".isValidAt(1)))
        .then(node1.CheckTx_RPC("tx1"))
        .then(node1.ABCI_ProcessRequest())
        .then(node1.ReceiveCheckTxResponse())
        .then(assert_(node1.inMempool("tx1") and node1.inCache("tx1")))
    
    run test2 = test1
        .then(node1.Update(1, Set("tx1"), Map("tx1" -> true)))
        .then(assert_(not(node1.inMempool("tx1")) and node1.inCache("tx1")))

    run test3 = test1
        .then(node1.P2P_SendTx_(node2))
        .then(node2.P2P_ReceiveTxs())
        .then(assert_(not(node2.inMempool("tx1")) and node2.inCache("tx1")))

    run test4 = test3
        .then(node2.ABCI_ProcessRequest())
        .then(node2.ReceiveCheckTxResponse())
        .then(assert_(node2.inMempool("tx1") and node2.inCache("tx1")))

    run test5 = test4
        .then(node2.Update(1, Set("tx1"), Map("tx1" -> "tx1".isValidAt(1))))
        .then(assert_(not(node2.inMempool("tx1")) and node2.inCache("tx1")))

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
