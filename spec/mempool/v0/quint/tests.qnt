module tests {
    import base.* from "./base"
    import basicSpells.* from "./basicSpells"
    import errors.* from "./errors"
    import mempoolv0.* from "./mempoolv0"
    import chain as Chain from "./chain"
    import abciServers as ABCI from "./abciServers"
    import network as Network from "./network"
    import history as History from "./history"

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
        Chain::unchanged,
        ABCI::unchanged,
        Network::unchanged,
        History::unchanged,
    }

    // action assert__(__action: bool, __cond: bool): bool =
    //     __action.then(assert_(__cond))

    //--------------------------------------------------------------------------
    // Tests
    //--------------------------------------------------------------------------
    run test1 = init
        .then(node1.CheckTx_RPC(tx1))
        .then(node1.ReceiveCheckTxResponse())
        .then(assert_(node1.inMempool(tx1) and tx1.in(node1.Cache())))
    
    run test2 = test1
        .then(all{ node1.Update(1, Set(tx1), Map(tx1 -> true)), Chain::unchanged })
        .then(assert_(not(node1.inMempool(tx1)) and tx1.in(node1.Cache())))

    run test3 = test1
        .then(node1.P2P_SendTx_(node2))
        .then(node2.P2P_ReceiveTxs())
        .then(assert_(not(node2.inMempool(tx1)) and tx1.in(node2.Cache())))

    run test4 = test3
        .then(node2.ReceiveCheckTxResponse())
        .then(assert_(node2.inMempool(tx1) and tx1.in(node2.Cache())))

    run test5 = test4
        .then(all { node2.Update(1, Set(tx1), Map(tx1 -> tx1.isValidAt(1))), Chain::unchanged })
        .then(assert_(not(node2.inMempool(tx1)) and tx1.in(node2.Cache())))

    run testInvalid1 = init
        .then(assert_(not(tx3.isValidAt(0))))
        .then(CheckTx_RPC(node1, tx3))
        .then(ReceiveCheckTxResponse(node1))

    //--------------------------------------------------------------------------
    // an initial state with a non-empty chain
    action init1 = all {
        mempool' = NodeIds.mapBy(_ => Set()),
        cache' = NodeIds.mapBy(_ => Set()),
        _step' = { node: noNode, name: "init", args: Map() },
        _error' = { node: noNode, error: noError },
        mempoolHeight' = NodeIds.mapBy(_ => 0),
        History::init,
        Network::init,
        Chain::chain' = List(Set("tx1")),
        ABCI::init,
    }

    run testUpdate1 = init1
        .then(node1.BlockExecutor_MempoolUpdate_nonProposer())
        .then(assert_(not(node1.inMempool(tx1)) and tx1.in(node1.Cache())))

    //--------------------------------------------------------------------------
    // P2P
    //--------------------------------------------------------------------------
    run testNet1 = test1
        .then(P2P_SendTx_(node1, node2))
        .then(assert_(node2.Network::incomingMsgs().nonEmpty()))
        .then(P2P_ReceiveTxs(node2))
        .then(assert_(node2.Network::incomingMsgs().isEmpty()))
        .then(assert_(tx1.in(node2.Cache())))

    //--------------------------------------------------------------------------
    // Helper functions for debugging
    //--------------------------------------------------------------------------
    def nodeState(node) = (
        ("mempool", node.Mempool()), 
        ("cache", node.Cache()),
        ("height", node.MempoolHeight()),
        ("abci requests", ABCI::requestResponses.get(node)),
        ("p2p incoming msgs", Network::msgs.get(node)),
    )
    def ns(nodeId) = nodeState(nodeId)
    val s1 = ns("n1")
    val s2 = ns("n2")
    val s3 = ns("n3")

    def state = (
        ("step", _step),
        ("error", _error),
        ("chain", Chain::chain),
    )

    def history = (
        ("historicMempoolTxs", History::historicMempoolTxs),
        ("txsReceivedViaRPC", History::txsReceivedViaRPC),
        ("gossippedTxs", History::gossippedTxs),
    )
}

module __repl__ {
    import base.* from "./base"
    import chain as Chain from "./chain"
    // import abciServers as ABCI from "./abciServers"
    import network as Network from "./network"
    import mempoolv0.* from "./mempoolv0"
    import tests.* 
    // import properties.* from "./properties"
} //-- __repl__
