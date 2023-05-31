// -*- mode: Bluespec; -*-

// An ABCI server implemented by an application receives asynchronous ABCI
// requests and replies with ABCI responses.
module abciServers {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"
    import abciMessages.* from "./abciMessages"

    type ABCIRequests = { tag: str, tx: Tx }
    val ABCIRequestTags = Set("CheckTx", "RecheckTx")
    
    // The server receives requests in a queue (a set actually) and responds
    // to a queue (a set) of responses. Each request maps to: the node id of the
    // sender, the height at which it was requested, and the response, if any.
    var requestResponses: NodeId -> (Request -> (NodeId, Height, Response))

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
    def RequestResponses(node) = requestResponses.get(node)
    def ResponsesTo(node, request) = requestResponses.get(node).get(request)
    def Requests(node) = node.RequestResponses().keys()
    
    def ABCI_senderFor(node, request) = node.ResponsesTo(request)._1
    def ABCI_heightFor(node, request) = node.ResponsesTo(request)._2
    def ABCI_responseFor(node, request) = node.ResponsesTo(request)._3

    def ABCI_checkTxRequests(node) = 
        node.Requests().filter(r => r.checkTxType == "New")
    def ABCI_recheckTxRequests(node) = 
        node.Requests().filter(r => r.checkTxType =="Recheck")

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action ABCI_init =
        requestResponses' = NodeIds.mapBy(_ => Map())

    // For a "New" request we store the sender.
    action ABCI_SendRequestNewCheckTx(nodeId, tx, senderId, h) = 
        val req = { tag: "CheckTx", tx: tx, checkTxType: "New" }
        val err = if (req.tx.isValidAt(h)) noError else InvalidTxError
        val resp = { tag: req.tag, err: err }
        requestResponses' = requestResponses.mapPut(nodeId, 
            (rs) => rs.put(req, (senderId, h, resp)))

    // For "Recheck" we send multiple requests at once.
    action ABCI_SendRequestRecheckTxs(nodeId, txs, h) =
        val newRequests = txs.map(tx => { tag: "CheckTx", tx: tx, checkTxType: "Recheck" })
        requestResponses' = requestResponses.mapPut(nodeId, (rs) =>
            rs.mapPutMany(newRequests, (req) => {
                val err = if (req.tx.isValidAt(h)) noError else InvalidTxError
                (noNode, h, { tag: req.tag, err: err }) 
            }))

    action ABCI_RemoveRequest(nodeId, request) =
        requestResponses' = requestResponses.mapPut(nodeId, (rs) => rs.mapRemove(request))

    action ABCI_unchanged =
        requestResponses' = requestResponses

}
