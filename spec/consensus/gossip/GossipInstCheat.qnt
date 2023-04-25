// This module implements GOSSIP-I without really gossiping. All tuples are immediately available to all.
module GossipInstCheat {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"

    import GossipI.* from "./GossipI"

    import GossipAbstract as GA from "./GossipAbstract"
    export GossipAbstract as GA

    val stateAll = {nada: "nada"}

    action unchangedAll = true
    action init = true

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p: ProcId, t: GossipMessage): bool = {
        val pA = GA::GossipState.get(p).adds
        val pAN = pA.unionE(t)
        val pD = GA::GossipState.get(p).dels
        all {
            not(pD.contains(t)),
            GA::GossipState' = ProcIds.mapBy(_ => {adds: pAN, dels: pD})
        }
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p: ProcId, ts: Set[GossipMessage]): bool = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        val pDN = pD.union(ts)
        all {
            ts.forall(t => pA.contains(t)),
            GA::GossipState' = ProcIds.mapBy(_ => {adds: pA, dels: pDN})
        }
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Step does not need to do anything, since everything is imediatelly gossiped.
    //Gossip using P2P
    action gossip(p,q) = all {
        GA::unchangedAll,
        GA::P2PInst::PA::unchangedAll,
    }
}