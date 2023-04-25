// The P2P Interface
// Must be imported as `P2PI.*`
module P2PI {
 //====== [VOC-GOSSIP-P2P-TYPES] P2P-I
    type P2PMessage =
        | {tag: "query_request" }
        | {tag: "query_response", tuples: Set[{tag:str, height:int, round:int,              signer:int}] }
        | {tag: "update_request", tuples: Set[{tag:str, height:int, round:int, payload:str, signer:int}]}

//==== Exported methods P2P-I
    // Commented actions are implemented by P2PAbstract or P2PInst*
    //action send(p,q,m): (ProcId, ProcId, P2PMessage) => bool
    //def received(p): Set[P2PMessage]
    //action ignoreNode(p:ProcId, q:ProcId, flag:bool): bool
}