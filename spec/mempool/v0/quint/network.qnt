// -*- mode: Bluespec; -*-
module network {
    import basicSpells.* from "./basicSpells"
    import base.* from "./base"

    // const Peers: NodeId -> Set[NodeId]
    val Peers: NodeId -> Set[NodeId] = Map(
        "n1" -> Set("n2", "n3"),
        "n2" -> Set("n1", "n3"),
        "n3" -> Set("n1", "n2")
    )

    //--------------------------------------------------------------------------
    // Network state
    type TxsMsg = { tag: str, senderId: NodeId, tx: Tx }
    var msgs: NodeId -> Set[TxsMsg] // map from receiver to messages to process

    //--------------------------------------------------------------------------
    action Network_init =
        msgs' = NodeIds.mapBy(_ => Set())

    action sendTo(msg: TxsMsg, peer: NodeId): bool =
        msgs' = msgs.put(peer, msgs.get(peer).setAdd(msg))

    action receiveMsg(nodeId: NodeId, msg: TxsMsg): bool =
        msgs' = msgs.put(nodeId, msgs.get(nodeId).setRemove(msg))    

    action Network_unchanged =
        msgs' = msgs

    //--------------------------------------------------------------------------
    def Network_incomingMsgs(nodeId) =
        msgs.get(nodeId)
}