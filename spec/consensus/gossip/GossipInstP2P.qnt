//This module implements GOSSIP-I using P2P-I. Nodes converge by exchanging messages.
module GossipInstP2P {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"
    import P2PI.*    from "./P2PI"

    import GossipI.* from "./GossipI"
    import GossipAbstract as GA from "./GossipAbstract"
    export GossipAbstract as GA

    type CONSTupleNoLoad =
        | { tag: "proposal",  height: int, round: int, signer: ProcId}
        | { tag: "prevote",   height: int, round: int, signer: ProcId}
        | { tag: "precommit", height: int, round: int, signer: ProcId}

    var processed: Set[{src:ProcId, dst:ProcId, msg: P2PMessage}]

    action unchangedAll() = all {
        processed' = processed
    }

    action init: bool = all {
        processed' = Set(),
    }

    val stateAll = {processed: processed}


    //send query message
    action sendQuery(p: ProcId, q: ProcId): bool = all {
        //networkState
        GA::P2PInst::send(p,q, {tag: "query_request"}),
        //pid2peerState
        GA::P2PInst::PA::pid2peerState' = GA::P2PInst::PA::pid2peerState,

        GA::unchangedAll,
        unchangedAll,
    }


    def dropPayload(e: GossipMessage): CONSTupleNoLoad = {
        {tag:e.tag, height:e.height, round:e.round, signer:e.signer}
    }

    def getTuplesNoLoad(p: ProcId): Set[CONSTupleNoLoad] = {
        GA::localView(p).map(e => dropPayload(e))
    }

    def getTuples(p: ProcId): Set[GossipMessage] = {
        GA::localView(p)
    }

    //send query response
    action respondQuery(p: ProcId, q: ProcId): bool = all {
        nondet m = GA::P2PInst::received(q)
                                .exclude(processed)
                                .filter(e => e.src == p)
                                .filter(e => e.msg.tag == "query_request")
                                .oneOf()
        val myTuples = getTuplesNoLoad(q)
        all {
            //networkState
            GA::P2PInst::send(q,p, {tag: "query_response", tuples: myTuples}),
            //pid2peerState
            GA::P2PInst::PA::pid2peerState' = GA::P2PInst::PA::pid2peerState,
            GA::unchangedAll,
            processed' = processed.unionE(m),
        }
    }

    //receive query response
    //identify differences
    //send differences
    action processResponse(p: ProcId, q: ProcId): bool = all{
        nondet m = GA::P2PInst::received(p)
                            .exclude(processed)
                            .filter(e => e.src == q)
                            .filter(e => e.msg.tag == "query_response")
                            .oneOf()

        val myTuples = getTuples(p)
        val myUpdates = myTuples.filter(e => not ( m.msg.tuples.contains(dropPayload(e))))
        all {
            //networkState
            GA::P2PInst::send(p,q, {tag: "update_request", tuples: myUpdates}),
            //pid2peerState
            GA::P2PInst::PA::pid2peerState' = GA::P2PInst::PA::pid2peerState,
            GA::unchangedAll,
            processed' = processed.unionE(m),
        }
    }


    //Adds a tuple to the local view.
    action addTuples(p: ProcId, ts: Set[GossipMessage]): bool = {
        val pA = GA::GossipState.get(p).adds
        val pAN = pA.union(ts)
        val pD = GA::GossipState.get(p).dels
        all {
            ts.forall(t => not(pD.contains(t))),
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", pAN))
        }
    }

    //receive differences
    //merge
    action processDiff(p: ProcId, q: ProcId): bool = all{
        val ms = GA::P2PInst::received(q)
                            .exclude(processed)
                            .filter(e => e.src == p)
                            .filter(e => e.msg.tag == "update_request")

        all {
            print("processDiff", ms),
            ms != Set(),
            nondet m = ms.oneOf()
            all {
                addTuples(q, m.msg.tuples),
                processed' = processed.unionE(m),
            },
            //networkState
            //pid2peerState
            GA::P2PInst::PA::unchangedAll,
        }
    }
   // This is a version of gossip that uses messages from the P2P layer.
    action gossip(p,q) = any {
        sendQuery(p,q),
        respondQuery(p,q),
        processResponse(p,q),
        processDiff(p,q),
    }

    action gossipStep(p,q) = gossip(p,q)

    //====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p: ProcId, t:GossipMessage): bool = all {
        addTuples(p, Set(t))
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p: ProcId, ts: Set[GossipMessage]): bool = {
        val pA = GA::GossipState.get(p).adds
        val pAN = pA.exclude(ts)
        val pD = GA::GossipState.get(p).dels
        val pDN = pD.union(ts)
        all {
            ts.forall(t => not(pD.contains(t))),
            GA::GossipState' = GA::GossipState.set(p, {adds: pA, dels: pDN})
        }
    }
}