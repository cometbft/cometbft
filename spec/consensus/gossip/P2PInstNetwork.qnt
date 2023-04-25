//This module implements P2P "using" the network
// - everything sent is put in a pending set
// - elements from the pending set are delivered randomly
// - there are no duplications
// - there are no losses
module P2PInstNetwork {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"
    import P2PI.*    from "./P2PI"

    import P2PAbstract as PA from "./P2PAbstract"
    export P2PAbstract as PA

    type OutGoingMsg = {dst:ProcId, msg: P2PMessage}
    type IncomingMsg = {src:ProcId, msg: P2PMessage}

    def nonDelMsgs(p): ProcId => Set[{src: ProcId, dst: ProcId, msg:P2PMessage}] = {
        PA::networkState.msgsSent.filter(m => m.src == p and not(m.in(PA::networkState.msgsRcvd)))
    }

    //====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (ProcId, ProcId, P2PMessage) => bool = all {
        val pState = PA::pid2peerState.get(p)
        pState.connectedNodes.contains(q),

        val msg = {src:p, dst:q, msg:m}
        all {
            print("sending", msg),
            not(PA::networkState.msgsSent.contains(msg)),

            PA::networkState' = PA::networkState.with("msgsSent", PA::networkState.msgsSent.union(Set(msg)))
        },
    }

    def received(p) = PA::networkState.msgsRcvd.filter(m => m.dst == p)


//===== Step
    action receiveMessageStep = {
        val nonDelMsgsAll = PA::networkState.msgsSent.filter(m => not(m.in(PA::networkState.msgsRcvd))
                                                        and m.dst.in(PA::pid2peerState.get(m.src).connectedNodes))
        all {
            size(nonDelMsgsAll) > 0,
            nondet msg = nonDelMsgsAll.oneOf()
            all {
                print("receiving", msg),
                PA::networkState' = PA::networkState.with("msgsRcvd", PA::networkState.msgsRcvd.union(Set(msg))),
                PA::pid2peerState' = PA::pid2peerState
            }
        }
    }

    action step = any {
        receiveMessageStep,
    }
}