//This module implements P2P without using the network
// - everything sent is immediately delivered
// - there are no duplications
// - there are no losses
// as long as the nodes are connected.
module P2PInstCheat {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"
    import P2PI.*    from "./P2PI"

    import P2PAbstract as PA from "./P2PAbstract"
    export P2PAbstract as PA


//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p: ProcId, q:ProcId, m: P2PMessage): bool = all {
        PA::areConnected(p,q),
        val msg = {src:p, dst:q, msg:m}
        all {
            not(msg.in(PA::networkState.msgsRcvd)),
            PA::networkState' = PA::networkState.with("msgsSent", PA::networkState.msgsSent.union(Set(msg)))
                                                .with("msgsRcvd", PA::networkState.msgsRcvd.union(Set(msg)))
        }
    }

    def received(p): ProcId => Set[{src:ProcId, dst:ProcId, msg: P2PMessage}] =
        PA::networkState.msgsRcvd.filter(m => m.dst == p)


//===== Steps
    action receiveMessageStep = all {
        PA::unchangedAll,
        false   //This forces this action to fail and not stutter.
    }
}
