// -*- mode: Bluespec; -*-
module network {
    import basicSpells.* from "./basicSpells"
    import params as Params from "./base"
    import base.* from "./base"

    // const Peers: NodeId -> Set[NodeId]
    val Peers: NodeId -> Set[NodeId] = Params::__Peers

    //--------------------------------------------------------------------------
    // Network state
    type TxsMsg = { tag: str, senderId: NodeId, tx: Tx }
    var msgs: NodeId -> Set[TxsMsg] // map from receiver to messages to process

    //--------------------------------------------------------------------------
    action Network_init =
        msgs' = NodeIds.mapBy(_ => Set())

    action sendTo(msg: TxsMsg, peer: NodeId): bool =
        msgs' = msgs.mapPut(peer, (ms) => ms.setAdd(msg))

    action receiveMsg(nodeId: NodeId, msg: TxsMsg): bool =
        msgs' = msgs.mapPut(nodeId, (ms) => ms.setRemove(msg))    

    action Network_unchanged =
        msgs' = msgs

    //--------------------------------------------------------------------------
    def Network_incomingMsgs(nodeId) =
        msgs.get(nodeId)
}