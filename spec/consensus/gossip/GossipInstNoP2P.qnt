// This module implements GOSSIP-I without using P2P. Nodes compare their local views and converge 2-by-2.
module GossipInstNoP2P {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"

    import GossipI.* from "./GossipI"
    import GossipAbstract as GA from "./GossipAbstract"
    export GossipAbstract as GA

    val stateAll = {nada: "nada"}

    action unchangedAll = true
    action init = true

    // This is a "cheating" version of gossip
    // It describes the abstract wanted behavior, which is to simply merge local views.
    action gossip(p,q) = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        val qA = GA::GossipState.get(q).adds
        val qD = GA::GossipState.get(q).dels
        all {
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", union(pA,qA))
                                                                            .with("dels", union(pD,qD)))
                                              .set(q, GA::GossipState.get(q).with("adds", union(pA,qA))
                                                                            .with("dels", union(pD,qD)))
        }
    }

    action gossipStep(p,q) = all {
        gossip(p,q),
        GA::P2PInst::PA::unchangedAll,
    }



//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (ProcId, GossipMessage) => bool = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        all {
            not(pD.contains(t)),
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", pA.unionE(t)))
        }
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, ts): (ProcId, Set[GossipMessage]) => bool = all {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        all {
            ts.forall(t => pA.contains(t)),
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", exclude(pA,ts))
                                                                            .with("dels", union(pD,ts)))
        }
    }
}