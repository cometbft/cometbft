// -*- mode: Bluespec; -*-
/** Mempool V0
  *
  * One of the goals of this spec is to easily map actions and data structures
  * to the implementation. 
  */

// Assumptions/Simplifications: 
// - The network topology is fixed: nodes do not leave or join the network, peers do not change.

module mempoolv0 {
    import basicSpells.* from "./basicSpells"
    import params as Params from "./base"
    import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import history.* from "./history"

    //--------------------------------------------------------------------------
    // Spec parameters
    //--------------------------------------------------------------------------
    val MempoolMaxSize: int = Params::__MempoolMaxSize
    // assume(MempoolMaxSize > 0)

    type Config = { keepInvalidTxsInCache: bool }
    val Configs: NodeId -> Config = Params::__Configs

    //--------------------------------------------------------------------------
    // Auxiliary variables
    //--------------------------------------------------------------------------
    type Step = str
    val Steps: Set[Step] = Set("init", "CheckTx", "ReceiveCheckTxResponse", 
        "Update", "ReceiveRecheckTxResponse", "P2P_ReceiveTxs", "P2P_SendTx", 
        "ABCI::ProcessCheckTxRequest")

    var _step: { node: NodeId, name: Step, args: str -> str }

    action setStep(node, s, args) =
        _step' = { node: node, name: s, args: args }

    var _error: { node: NodeId, error: Error }

    action setError(node, err) =
        _error' = { node: node, error: err }

    //--------------------------------------------------------------------------
    // Mempool state
    //--------------------------------------------------------------------------
    // [mempoolTx]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L671
    type MempoolTx = { 
        tx: Tx, 
        height: Height, // height at which tx was validated
        senders: Set[NodeId] // peers that send tx, or none if tx was emmitted by a client
    }

    val noMempoolTx = { tx: "", height: 0, senders: Set() }

    var mempool: NodeId -> Set[MempoolTx]

    // The transactions received by a node.
    var cache: NodeId -> Set[Tx]

    // The last block Update()'d to.
    var mempoolHeight: NodeId -> Height

    def Mempool(nodeId) = mempool.get(nodeId)
    def Cache(nodeId) = cache.get(nodeId)
    def MempoolHeight(nodeId) = mempoolHeight.get(nodeId)

    //--------------------------------------------------------------------------
    // Cache auxiliary definitions
    //--------------------------------------------------------------------------
    action addToCache(node, tx) =
        cache' = cache.mapPut(node, (txs) => txs.setAdd(tx))

    action forceRemoveFromCache(node: NodeId, tx: Tx): bool =
        cache' = cache.mapPut(node, (txs) => txs.setRemove(tx))

    action removeFromCache(node, invalidTx_) =
        if (not(Configs.get(node).keepInvalidTxsInCache))
            node.forceRemoveFromCache(invalidTx_)
        else cache' = cache

    //--------------------------------------------------------------------------
    // Mempool auxiliary definitions
    //--------------------------------------------------------------------------
    def mempoolIsEmpty(node) =
        node.Mempool().isEmpty()

    def mempoolIsFull(node) =
        size(node.Mempool()) > MempoolMaxSize

    def mempoolTxs(node: NodeId): Set[Tx] = 
        node.Mempool().map(e => e.tx)

    // Returned value is Optional: either none or a singleton.
    def memTxFor(node: NodeId, tx: Tx): Set[MempoolTx] =
        node.Mempool().filter(e => e.tx == tx)

    // Get the element from the singleton or none in any other case
    def chooseSomeMemTxFor(node: NodeId, tx: Tx): MempoolTx =
        // node.memTxFor(tx).chooseSome()
        val memTxs = node.memTxFor(tx).setToList()
        if (memTxs.length() == 1) memTxs[0] else noMempoolTx

    // Returned value is Optional: either none or a singleton.
    def sendersFor(node: NodeId, tx: Tx): Set[NodeId] = 
        node.memTxFor(tx).map(e => e.senders).flatten()

    def inMempool(node, tx) =
        tx.in(node.mempoolTxs())

    def senderFor(node, tx) = 
        node.Mempool().map(e => e.tx)

    action addToMempool(node, tx, h, senderId) =
        val senders = if (senderId == noNode) Set() else Set(senderId)
        val memTx = { tx: tx, height: h, senders: senders }
        mempool' = mempool.mapPut(node, (memTxs) => memTxs.setAdd(memTx))

    action removeFromMempool(node: NodeId, txs: Set[Tx]): bool =
        mempool' = mempool.mapPut(node, (memTxs) => memTxs.filter(e => e.tx.notIn(txs)))

    action addSender(node, tx, senderId) =
        if (node.inMempool(tx)) {
            val oldMemTx: MempoolTx = node.chooseSomeMemTxFor(tx)
            val newMemTx = oldMemTx.with("senders", oldMemTx.senders.union(senderId))
            mempool' = mempool.mapPut(node, (memTxs) => memTxs.setRemove(oldMemTx).setAdd(newMemTx))
        } else mempool' = mempool

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action init = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        cache' = NodeIds.mapBy(_ => Set()),
        _step' = { node: noNode, name: "init", args: Map() },
        _error' = { node: noNode, error: noError },
        mempoolHeight' = NodeIds.mapBy(_ => 0),
        History_init,
        Network_init,
        Chain_init,
        ABCI_init,
    }
    
    /** Validate a transaction and add it to the mempool. */
    // [CListMempool.CheckTx]: https://github.com/CometBFT/cometbft/blob/5a8bd742619c08e997e70bc2bbb74650d25a141a/mempool/clist_mempool.go#L202
    //-- __subaction__: this action is not part of the next-state relation
    action CheckTx(node: NodeId, tx: Tx, senderId: NodeId): bool = all {
        if (node.mempoolIsFull()) all {
            node.setError(MempoolIsFullError),
            mempool' = mempool,
            cache' = cache,
            ABCI_unchanged,
        } else all {
            if (tx.in(node.Cache())) all {
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
                node.ABCI_SendRequestNewCheckTx(tx, senderId, node.MempoolHeight() + 1),
                node.setError(noError),
                mempool' = mempool,
            }
        },
        mempoolHeight' = mempoolHeight,
    }

    /* Receive a specific transaction from a client via RPC. */
    // [Environment.BroadcastTxAsync]: https://github.com/CometBFT/cometbft/blob/111d252d75a4839341ff461d4e0cf152ca2cc13d/rpc/core/mempool.go#L22
    //-- __subaction__: this action is not part of the next-state relation
    action CheckTx_RPC(node, tx) = all {
        node.setStep("CheckTx", Map("node" -> tx)),
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
        node.setStep("ReceiveCheckTxResponse", Map()),
        node.ABCI_checkTxRequests().nonEmpty(),
        nondet request = oneOf(node.ABCI_checkTxRequests()) 
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
                    node.History_recordMempoolTx(request.tx),
                    cache' = cache,
                }
            } else all { // ignore invalid transaction
                node.removeFromCache(request.tx),
                node.setError(InvalidTxWarning),
                mempool' = mempool,
                History_unchanged,
            }
        },
        mempoolHeight' = mempoolHeight,
        Network_unchanged,
        Chain_unchanged,
    }

    /** Consensus' BlockExecutor calls Update to update the mempool after executing txs.
      * txResults are the results of ResponseFinalizeBlock for every tx in txs.
      * BlockExecutor holds the mempool lock while calling this function.
      */
    // [CListMempool.Update]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L577
    //-- __subaction__: this action is not part of the next-state relation
    action Update(node: NodeId, h: Height, txs: Set[Tx], txValidResults: Tx -> bool): bool = all {
        node.setStep("Update", Map("node" -> node)),
        require(h > node.MempoolHeight()),
        require(txs.nonEmpty()),

        // Update mempool's height.
        mempoolHeight' = mempoolHeight.put(node, h),

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
            node.ABCI_SendRequestRecheckTxs(txsStillInMempool, h)
        else ABCI_unchanged,

        node.setError(noError),
        History_unchanged,
        Network_unchanged,
    }

    // Consensus reactor updates the mempool with the latest block just created
    // and rechecks outstanding txs.
    action Consensus_Update(node: NodeId): bool = all {
        require(not(Chain_isEmpty)),
        require(node.MempoolHeight() <= Chain_latestHeight),
        val h = node.MempoolHeight() + 1 // block's height
        any {
            // This node is not the proposer: take the block from the chain.
            // All txs are valid because they are committed in the chain.
            val txs: Set[Tx] = Chain_getBlock(node.MempoolHeight())
            val txValidResults: Tx -> bool = txs.mapBy(_ => true)
            all {
                node.Update(h, txs, txValidResults),
                Chain_unchanged,
            },
            all {
                // This node is the proposer: take the block from the mempool.
                require(node.mempoolTxs().nonEmpty()),
                nondet txs: Set[Tx] = oneOf(node.mempoolTxs().nonEmptyPowerset())
                val txValidResults: Tx -> bool = txs.mapBy(tx => isValidAt(tx, h))
                all {
                    node.Update(h, txs, txValidResults),
                    Chain_newBlockFrom(txs),
                }
            },
        }
    }

    // Callback function to process a CheckTx response, when rechecking.
    // [CListMempool.resCbRecheck]: https://github.com/CometBFT/cometbft/blob/6498d67efdf0a539e3ca0dc3e4a5d7cb79878bb2/mempool/clist_mempool.go#L432
    action ReceiveRecheckTxResponse(node: NodeId): bool = all {
        node.setStep("ReceiveRecheckTxResponse", Map("node" -> node)),
        node.ABCI_recheckTxRequests().nonEmpty(),
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
        mempoolHeight' = mempoolHeight,
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
        node.setStep("P2P_ReceiveTxs", Map("node" -> node)),
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
    //-- __subaction__: this action is not part of the next-state relation
    action P2P_SendTx_(node: NodeId, peer: NodeId): bool = all {
        node.setStep("P2P_SendTx", Map("node" -> node, "peer" -> peer)),
        require(not(node.mempoolIsEmpty())),
        nondet tx = oneOf(node.mempoolTxs())
        val msg = { tag: "Txs", senderId: node, tx: tx }
        all {
            // If the msg was not already sent to this peer.
            require(peer.Network_incomingMsgs().forall(m => m != msg)),
            // If the peer is not a tx's sender.
            require(peer.notIn(node.sendersFor(tx))),
            sendTo(msg, peer),
            node.setError(noError),
            node.History_recordGossippedTx(tx),
            mempool' = mempool,
            cache' = cache,
            mempoolHeight' = mempoolHeight,
            ABCI_unchanged,
            Chain_unchanged,
        }
    }

    // [Reactor.broadcastTxRoutine]: https://github.com/CometBFT/cometbft/blob/5049f2cc6cf519554d6cd90bcca0abe39ce4c9df/mempool/reactor.go#L132
    action P2P_SendTx(node: NodeId): bool =
        nondet peer = oneOf(Peers.get(node))
        node.P2P_SendTx_(peer)

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
        }

    /* There may be action in other nodes too. */
    action ChainNext = 
        val notCommittedTxs = Txs.exclude(Chain_allTxsInChain)
        all {
            require(notCommittedTxs.nonEmpty()),
            nondet txs = oneOf(notCommittedTxs.nonEmptyPowerset())
            all {
                require(txs.forall(tx => isValidAt(tx, Chain_latestHeight + 1))),
                Chain_newBlockFrom(txs),
                setStep(noNode, "ChainNext", txs.mapBy(x => x)),
                setError(noNode, noError),
                mempool' = mempool, 
                mempoolHeight' = mempoolHeight, 
                cache' = cache, 
                History_unchanged,
                ABCI_unchanged,
                Network_unchanged,
            }
        }
    
    action step = any { NodeNext, ChainNext }
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
    val tx1 = "tx1"
    val tx2 = "tx2"
    val tx3 = "tx3"

    action assert_(__cond: bool): bool = all {
        __cond,
        mempool' = mempool, 
        mempoolHeight' = mempoolHeight, 
        cache' = cache,
        _step' = _step,
        _error' = _error,
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
        .then(assert_(tx1.isValidAt(1)))
        .then(node1.CheckTx_RPC(tx1))
        .then(node1.ReceiveCheckTxResponse())
        .then(assert_(node1.inMempool(tx1) and tx1.in(node1.Cache())))
    
    run test2 = test1
        .then(all{ node1.Update(1, Set(tx1), Map(tx1 -> true)), Chain_unchanged })
        .then(assert_(not(node1.inMempool(tx1)) and tx1.in(node1.Cache())))

    run test3 = test1
        .then(node1.P2P_SendTx_(node2))
        .then(node2.P2P_ReceiveTxs())
        .then(assert_(not(node2.inMempool(tx1)) and tx1.in(node2.Cache())))

    run test4 = test3
        .then(node2.ReceiveCheckTxResponse())
        .then(assert_(node2.inMempool(tx1) and tx1.in(node2.Cache())))

    run test5 = test4
        .then(all { node2.Update(1, Set(tx1), Map(tx1 -> tx1.isValidAt(1))), Chain_unchanged })
        .then(assert_(not(node2.inMempool(tx1)) and tx1.in(node2.Cache())))

    run testInvalid1 = init
        .then(CheckTx_RPC(node1, tx3)) // tx3 is invalid at height 0 according to isValidAt
        .then(ReceiveCheckTxResponse(node1))

    //--------------------------------------------------------------------------
    def nodeState(node) = (
        ("mempool", node.Mempool()), 
        ("cache", node.Cache()),
        ("height", node.MempoolHeight()),
        ("abci requests", requestResponses.get(node)),
        ("p2p incoming msgs", msgs.get(node)),
    )
    def ns(nodeId) = nodeState(nodeId)
    val s1 = ns("n1")
    val s2 = ns("n2")
    val s3 = ns("n3")

    def state = (
        ("step", _step),
        ("error", _error),
        ("chain", chain),
    )
}

module __repl__ {
    import base.* from "./base"
    import chain.* from "./chain"
    import abciServers.* from "./abciServers"
    import network.* from "./network"
    import mempoolv0.*
    // import properties.* from "./properties"
    import tests.* 
} //-- __repl__
