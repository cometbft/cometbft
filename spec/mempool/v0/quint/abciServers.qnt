// -*- mode: Bluespec; -*-

// An ABCI server (an application) receives asynchronous ABCI requests and
// replies with ABCI responses.
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

    def hasResponseFor(node, request) =
        node.ABCI_responseFor(request) != noResponse

    def requestsWithResponse(node: NodeId, checkTxType: str): Set[Request] =
        node.Requests().filter(r => r.checkTxType == checkTxType and node.hasResponseFor(r))

    def ABCI_checkTxRequests(node) = node.requestsWithResponse("New")
    def ABCI_recheckTxRequests(node) = node.requestsWithResponse("Recheck")

    val emptyResponse = (noNode, -1, noResponse)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action ABCI_init =
        requestResponses' = NodeIds.mapBy(_ => Map())

    // For a "New" request we store the sender.
    action ABCI_SendRequestNewCheckTx(nodeId, tx, senderId, h) = 
        val req = { tag: "CheckTx", tx: tx, checkTxType: "New" }
        requestResponses' = requestResponses.mapPut(nodeId, 
            (rs) => rs.put(req, (senderId, h, noResponse)))

    // For "Recheck" we send multiple requests at once.
    action ABCI_SendRequestRecheckTxs(nodeId, txs) =
        val newRequests = txs.map(tx => { tag: "CheckTx", tx: tx, checkTxType: "Recheck" })
        // Add all the new requests, with no sender and no response.
        requestResponses' = requestResponses.mapPut(nodeId, (rs) =>
            rs.keys().mapBy(req => if (req.in(newRequests)) emptyResponse else rs.get(req)))

    // The app receives a CheckTx request and creates a response.
    action ABCI_ProcessCheckTxRequest(node) =
        nondet req: Request = oneOf(node.Requests())
        all {
            require(not(node.hasResponseFor(req))),
            val senderId = node.ABCI_senderFor(req)
            val h = node.ABCI_heightFor(req)
            val err = if (req.tx.isValidAt(h)) noError else InvalidTxError
            val resp = { tag: req.tag, err: err }
            requestResponses' = requestResponses.mapPut(node, (rs) => rs.put(req, (senderId, h, resp))),
        }

    action ABCI_RemoveRequest(nodeId, request) = all {
        require(hasResponseFor(nodeId, request)), // We keep this condition to allow the action to return false; otherwise it will always evaluate to true.
        requestResponses' = requestResponses.mapPut(nodeId, (rs) => rs.mapRemove(request)), 
    }

    action ABCI_unchanged =
        requestResponses' = requestResponses

}
