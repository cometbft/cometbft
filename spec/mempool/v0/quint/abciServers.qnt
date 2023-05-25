// -*- mode: Bluespec; -*-

// An ABCI server receives asynchronous ABCI requests and replies with ABCI responses.
module abciServers {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"
    import abciMessages.* from "./abciMessages"

    type ABCIRequests = { tag: str, tx: Tx }
    val ABCIRequestTags = Set("CheckTx", "RecheckTx")
    
    // The ABCIServer receives requests in a queue (a set actually) and responds
    // to a queue (set) of responses. Each request maps to the node id of the
    // sender and to the response, if available.
    var requestSenderResponses: NodeId -> (Request -> (NodeId, Height, Response))

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
    
    def ABCI_heightFor(nodeId, request) = 
        requestSenderResponses.get(nodeId).get(request)._2

    def ABCI_responseFor(nodeId, request) = 
        requestSenderResponses.get(nodeId).get(request)._3

    def hasResponse(nodeId, request) =
        ABCI_responseFor(nodeId, request) != noResponse

    def requestsWithResponse(nodeId: NodeId, checkTxType: str): Set[Request] =
        requestSenderResponses.get(nodeId).keys()
            .filter(r => r.checkTxType == checkTxType and nodeId.hasResponse(r))
    def ABCI_checkTxRequests(nodeId) = requestsWithResponse(nodeId, "New")
    def ABCI_recheckTxRequests(nodeId) = requestsWithResponse(nodeId, "Recheck")

    val emptyResponse = (noNode, -1, noResponse)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action ABCI_init =
        requestSenderResponses' = NodeIds.mapBy(_ => Map())

    // For a "New" request we store the sender.
    action ABCI_SendRequestNewCheckTx(nodeId, tx, senderId, h) = 
        val req = { tag: "CheckTx", tx: tx, checkTxType: "New" }
        requestSenderResponses' = requestSenderResponses.put(nodeId, 
            requestSenderResponses.get(nodeId).put(req, (senderId, h, noResponse)))

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
            val h = ABCI_heightFor(nodeId, req)
            val err = if (req.tx.isValidAt(h)) noError else InvalidTxError
            val resp = { tag: req.tag, err: err }
            requestSenderResponses' = requestSenderResponses.put(nodeId, 
                requestSenderResponses.get(nodeId).put(req, (senderId, h, resp))),
        }

    action ABCI_RemoveRequest(nodeId, request) = all {
        require(hasResponse(nodeId, request)), // We keep this condition to allow the action to return false; otherwise it will always evaluate to true.
        requestSenderResponses' = requestSenderResponses.put(nodeId, 
            requestSenderResponses.get(nodeId).mapRemove(request)), 
    }

    action ABCI_unchanged =
        requestSenderResponses' = requestSenderResponses

}
