module Helpers {
    //===== Pure Helper functions
    pure def minusE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def minusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).exclude(Set(e)))
    pure def minusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).exclude(es))

    pure def plusE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def plusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).union(Set(e)))
    pure def plusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).union(es))
}

module Globals {
    type Proc = int

    //XXX: Ideally this should be a const, but because it would have to be
    //instantiated on every import, there would be duplication.
    //All processes in the system
    val Procs = Set(1,2,3,4,5)
    //Maximum number of outbound connections per node.
    val MaxConNum = 4
}


//Defines the P2P Interface
module P2PI {
 //====== [VOC-GOSSIP-P2P-TYPES] P2P-I
    type GOSSIPMessage =
        | {tag: "query" }
        | {tag: "query_response" }
        | {tag: "update", tuple: Set[(str, int, int, str, int)]}

//==== Exported methods P2P-I
    //action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool
    //def received(p): Set[GOSSIPMessage]

}

//Provides staandard behavior for P2PI implementations
module P2PAbstract {
    import Helpers.*
    import Globals.*
    import P2PI.*


//==== Common to implementationtions, but probably not actually part of the P2P-I.
    //===== P2P Internals                                       //TODO: Move some of these to P2PI
    var igProcs : Proc -> Set[Proc]                                 //Nodes to be ignored.
    var cnProcs  : Proc -> Set[Proc]                                 //Neighbor sets

    var sMsgs: Set[{src:Proc, dst:Proc, msg: GOSSIPMessage}]            //Unicast sent messages.
    var rMsgs: Set[{src:Proc, dst:Proc, msg: GOSSIPMessage}]          //Unicast received messages

    //Helper
    val p2pInternals = {ignore: igProcs, connections: cnProcs, sent:sMsgs, received:rMsgs}

    action PAunchangedAll = all {
        igProcs' = igProcs,
        cnProcs' = cnProcs,
        rMsgs' = rMsgs,
        sMsgs' = sMsgs

    }



    //Adds bidirectional connection if both nodes are below limit.
    action addPeer(p,q): (Proc,Proc) => bool = all {
        not(cnProcs.get(p).contains(q)),
        size(cnProcs.get(p)) < MaxConNum,
        not(cnProcs.get(q).contains(p)),
        size(cnProcs.get(q)) < MaxConNum,
        cnProcs' = plusE2(plusE2(cnProcs, p, q), q, p)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        cnProcs' = minusE2(minusE2(cnProcs, p, q), q, p)
    }


    action ignoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = plusE2(igProcs, p, q)
    }

    action doNotIgnoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = minusE2(igProcs, p, q)
    }

//==== Init
    action PAinit: bool = all {
        igProcs' = Procs.mapBy(_ => Set()),
        cnProcs' = Procs.mapBy(_ => Set()),
        sMsgs' = Set(),
        rMsgs' = Set()
    }

//==== Steps
    //Defined for interface compatibility, but must never succeed since it would be a stutter only.
    action PAstep = all {
        PAunchangedAll,
        false
    }


//===== Invariants
//================================= REQ-GOSSIP-P2P
    //[REQ-GOSSIP-P2P-UNICAST]
    val RGPUnicast = {
        sMsgs.forall(m => or {
            rMsgs.contains(m),
            not (cnProcs.get(m.src).contains(m.dst))
        })
    }

    temporal RGPUnicastProp = always (eventually (RGPUnicast))

    //[REQ-GOSSIP-P2P-CONCURRENT_CONN]
    val RGPConcurrent = {
        Procs.forall(p =>
                cnProcs.get(p).size() < MaxConNum
        )
    }

    temporal RGPConcurrentProp: bool =
        always (RGPConcurrent)

    //[REQ-GOSSIP-P2P-IGNORING]
    val RGPIgnoring: bool = {
        Procs.forall(p =>
            val cnP = cnProcs.get(p)
            val igP = igProcs.get(p)
            val igCn = cnP.intersect(igP)
                size(igCn) == 0
        )
    }

    temporal RGPIgnoringProp =
        always (RGPIgnoring)
}

//This module implements P2P without using the network
// - everything sent is immediately delivered
// - there are no duplications
// - there are no losses
// as long as the nodes are connected.
module P2PInstCheat {
    import Helpers.*
    import Globals.*
    import P2PI.*

    import P2PAbstract.* //as PA

    action PInstUnchangedAll = PAunchangedAll

//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        val msg = {src:p, dst:q, msg:m}
        all {
            not(msg.in(rMsgs)),
            sMsgs' = plusE(sMsgs, msg),
            rMsgs' = plusE(rMsgs, msg)
        }
    }

    def received(p): Proc => Set[{src:Proc, dst:Proc, msg: GOSSIPMessage}] =
        rMsgs.filter(m => m.dst == p)


//==== Init
    action PInstInit: bool = PAinit


//===== Steps
    action deliverMessageStep = all {PInstUnchangedAll, true}

    action PInstStep = all {PInstUnchangedAll, false}
}



//This module implements P2P "using" the network
// - everything sent is put in a pending set
// - elements from the pending set are delivered randomly
// - there are no duplications
// - there are no losses
module P2PInstNetwork {
    import Helpers.*
    import Globals.*
    import P2PI.*

    import P2PAbstract.* // as PA


    type OutGoingMsg = {dst:Proc, msg: GOSSIPMessage}
    type IncomingMsg = {src:Proc, msg: GOSSIPMessage}

    action PInstUnchangedAll = PAunchangedAll

    def nonDelMsgs(p): Proc => Set[{src: Proc, dst: Proc, msg:GOSSIPMessage}] = {
        sMsgs.filter(m => m.src == p and not(m.in(rMsgs)))
    }

//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        sMsgs' = plusE(sMsgs, {src:p, dst:q, msg:m}),
        rMsgs' = rMsgs
    }

    def received(p) = rMsgs.filter(m => m.dst == p)


//======Init
    action PInstInit: bool = PAinit

//===== Step
    action deliverMessageStep = {
        val nonDelMsgsAll = sMsgs.filter(m => not(m.in(rMsgs)) and m.dst.in(cnProcs.get(m.src)))
        all {
            size(nonDelMsgsAll) > 0,
            val m = nonDelMsgsAll.oneOf()
            all {
                rMsgs' = plusE(rMsgs, m),
                sMsgs' = sMsgs,

                igProcs' = igProcs,
                cnProcs' = cnProcs
            }
        }
    }

    action PInstStep = any {
        deliverMessageStep,
    }
}


module P2PTests {
    import Helpers.*
    import Globals.*
    import P2PAbstract.*

    //import P2PInstCheat.* //as PInst
    import P2PInstNetwork.* //as PInst

//======== Tests

    //Connects 2 nodes node towards forming a ring.
    action connectRingStep() = {
        //Find a node not in the ring.
        val nonConnected = Procs.filter(p => if (p < size(Procs)) not(cnProcs.get(p).contains(p+1)) else not(cnProcs.get(p).contains(1)))
        all {
            cnProcs' = Procs.mapBy(p => if (p < size(Procs)) Set(p+1) else Set(1)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectFullStep() = {
        all {
            cnProcs' = Procs.mapBy(p => Procs),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectStarStep() = {
        all {
            cnProcs' = Procs.mapBy(p => if (p ==1 ) Procs.minusE(1) else Set(1)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectRandomStep() = {
        //TODO: Describe a random connected graph, without iteration, efficiently
        //nondet f = Procs.setOfMaps(powerset(Procs)).oneOf()
        //The graph below is actually a double link ring.
        all {
            cnProcs' = Procs.mapBy(p => if (p < size(Procs)-1) Set(p+1, p+2) else if (p == size(Procs)-1) Set(p+1,1) else Set(1,2)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

//init and step, used for random walks a fully configured "stack"
    action PTInit = PInstInit

    action PTStep = PInstStep

//===== Runs/Debug
    run sendReceiveInARingTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectRingStep)
            //send and receive messages
            .then(
                all {
                    send(1,2,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(2,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(deliverMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                all {
                    send(1,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }.fail())
    }


    run sendReceiveInFullyConnectedTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectFullStep)
            //send and receive messages
            .then(
                all {
                    send(1,2,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(1,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(2,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(deliverMessageStep.repeated(3))
            .then(all {RGPUnicast, PInstUnchangedAll})
    }

    run sendReceiveInStarConnectedTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectStarStep)
            //send and receive messages
            .then(
                all {
                    send(1,2,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(1,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(deliverMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                all {
                    send(2,3,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }.fail()
            )
    }

    run sendReceiveInRandomConnectedTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectRandomStep)
            //send and receive messages
            .then(
                val p = cnProcs.keys().filter(pa => cnProcs.get(pa).size() > 0 ).oneOf()
                val q = cnProcs.get(p).oneOf()
                all {
                    send(p,q,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                val p = cnProcs.keys().filter(pa => cnProcs.get(pa).size() > 0 ).oneOf()
                val q = cnProcs.get(p).oneOf()
                all {
                    send(p,q,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(deliverMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                val p = cnProcs.keys().filter(pa => cnProcs.get(pa) != Procs ).oneOf()
                val q = Procs.exclude(cnProcs.get(p)).oneOf()
                all {
                    send(p,q,{gat: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }.fail()
            )

    }

//===== Test Steps
    //Update the connections
    action randomDropConnectionStep() = all {
        Procs.filter(ap => cnProcs.get(ap).size() > 0).size() > 0,
        nondet p = Procs.filter(ap => cnProcs.get(ap).size() > 0).oneOf()
        nondet q = cnProcs.get(p).oneOf()

        removePeer(p,q),
        igProcs' = igProcs,
        sMsgs' = sMsgs,
        rMsgs' = rMsgs
    }

    action randomConnectionStep() = all {
        //Are there two processes that can still connect to each other.
        val nonFull = Procs.filter(p => cnProcs.get(p).size() < MaxConNum)
        val nonCon = nonFull.filter(p => nonFull.exists(q => and(q != p , not (cnProcs.get(p).contains(q)))))
        all {
            nonCon.size() > 0,
            all {
                nondet p = nonCon.oneOf()
                nondet q = nonFull.minusE(p).filter(aq => not (cnProcs.get(p).contains(aq))).oneOf()

                addPeer(p,q),
                sMsgs' = sMsgs,
                rMsgs' = rMsgs,
                igProcs' = igProcs,
            }
        }
    }

    action randomMessageStep = {
        val p = Procs.oneOf()
        val q = Procs.oneOf()
        all {
            send(p,q,{gat: "query"}),

            cnProcs' = cnProcs,
            igProcs' = igProcs,
        }
    }

    action randomStep = any {
        randomConnectionStep,
        randomDropConnectionStep,
        randomMessageStep
    }

    //quint test gossip.qnt --main P2PTests--verbosity=3
    //quint run gossip.qnt --main P2PTests --step randomStep --verbosity=3
    //quint run gossip.qnt --main P2PTests --init PTInit --step PTStep --verbosity=3
}



//Define the Gossip Interface
module GossipI {
    import Globals.*

//====== [VOC-CONS-GOSSIP-TYPES] GOSSIP-I
    type CONSTuple =
        | { tag: "proposal",  height: int, round: int, payload: str, signer: Proc}
        | { tag: "prevote",   height: int, round: int, payload: str, signer: Proc}
        | { tag: "precommit", height: int, round: int, payload: str, signer: Proc}

//==== Exported methods GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    //action addTuple(p,t): (Proc, CONSTuple) => bool

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    //action removeTuples(p, es): (Proc, Set[CONSTuple]) => bool

    //[DEF-READ-TUPLE]
    //The local view
    //def localView(p): Proc => Set[CONSTuple]
}

module GossipAbstract {
    import Helpers.*
    import Globals.*

    import GossipI.*
}


// This module implements GOSSIP-I without really gossiping. All tuples are immediately available to all.
module GossipInstCheat {
    import Helpers.*
    import Globals.*

    import GossipI.*        //GI
    import GossipAbstract.* //GA

//===== GOSSIP Internals
    var globalViewAdds : Set[CONSTuple]
    var globalViewDels : Set[CONSTuple]

    //Helper
    def globalView: Set[CONSTuple] = {
        globalViewAdds.exclude(globalViewDels)
    }

    //Debug
    val fullState = {adds: globalViewAdds, dels: globalViewDels}

    //Unchanged
    action GInstUnchangedAll() = all {
        globalViewAdds' = globalViewAdds,
        globalViewDels' = globalViewDels,
    }



//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p: Proc, t: CONSTuple): bool = all {
        not(globalViewDels.contains(t)),
        globalViewAdds' = plusE(globalViewAdds, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p: Proc, es: Set[CONSTuple]): bool = all {
        globalViewDels' = globalViewDels.union(es),
        globalViewAdds' = globalViewAdds.exclude(es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = globalView


//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action GInstInit: bool = all {
        globalViewAdds' = Set(),
        globalViewDels' = Set()
    }


//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    action addTupleStep(p, e) = all {
            addTuple(p, e),
            //unchanged
            globalViewDels' = globalViewDels,
    }

    //Step does not need to do anything, since everything is imediatelly gossiped.
        //Gossip using P2P
    action gossipStep(p,q) = all {
        GInstUnchangedAll
    }

    action GInstStep: bool = true
}

// This module implements GOSSIP-I without using P2P. Nodes compare their local views and converge 2-by-2.
module GossipInstNoP2P {
    import Helpers.*
    import Globals.*
    import GossipI.*
    import GossipAbstract.*


//===== GOSSIP Internals
    var localViewAdds : Proc -> Set[CONSTuple]
    var localViewDels : Proc -> Set[CONSTuple]

    //Helper
    def globalView: Set[CONSTuple] = {
        val allAdds = localViewAdds.keys().fold(Set(), (acc, k) => acc.union(localViewAdds.get(k)))
        val allDels = localViewDels.keys().fold(Set(), (acc, k) => acc.union(localViewDels.get(k)))

        allAdds.exclude(allDels)
    }

    //Debug
    val fullState = {adds: localViewAdds, dels: localViewDels}

    //Unchanged
    action GInstUnchangedAll() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
    }


    // This is a "cheating" version of gossip
    // It describes the abstract wanted behavior, which is to simply merge local views.
    action gossip(p,q) = {
        val lvPA = localViewAdds.get(p)
        val lvPD = localViewDels.get(p)
        val lvQA = localViewAdds.get(q)
        val lvQD = localViewDels.get(q)

        all {
            localViewAdds' = localViewAdds.set(p, union(lvPA,lvQA))
                                        .set(q, union(lvPA,lvQA)),
            localViewDels' = localViewDels.set(p, union(lvPD,lvQD))
                                        .set(q, union(lvPD,lvQD)),
        }
    }


//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (Proc, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = plusE2(localViewAdds, p, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (Proc, Set[CONSTuple]) => bool = all {
        localViewDels' = plusEs2(localViewDels, p, es),
        localViewAdds' = minusEs2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = localViewAdds.get(p).exclude(localViewDels.get(p))



//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action GInstInit: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }


//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Gossip using P2P
    action gossipStep(p,q) = all {
        gossip(p,q),
    }

    action addTupleStep(p, e) = all {
            addTuple(p, e),
            //unchanged
            localViewDels' = localViewDels,
    }

    action step: bool = any {
        //chose partners
        nondet p = oneOf(Procs)
        nondet q = oneOf(minusE(Procs,p))

        gossipStep(p,q)
    }


}


//This module implements GOSSIP-I using P2P-I. Nodes converge by exchanging messages.
module GossipInstP2P {
    import Helpers.*
    import Globals.*
    import P2PI.*
    import GossipI.*
    import GossipAbstract.* //as GA

    var localViewAdds : Proc -> Set[CONSTuple]
    var localViewDels : Proc -> Set[CONSTuple]

    //Helper
    def globalView: Set[CONSTuple] = {
        val allAdds = localViewAdds.keys().fold(Set(), (acc, k) => acc.union(localViewAdds.get(k)))
        val allDels = localViewDels.keys().fold(Set(), (acc, k) => acc.union(localViewDels.get(k)))

        allAdds.exclude(allDels)
    }

    val fullState = {adds: localViewAdds, dels: localViewDels}

    //Unchanged
    action GInstUnchangedAll() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
    }

    // This is a version of gossip that uses messages from the P2P layer.
    action gossip(p,q) = all {
        //TODO: any
        //send query message
        //send query response
        //receive query response
        //identify differences
        //send differences
        //receive differences
        //merge

        //TODO: temporary unchanged
        GInstUnchangedAll,
    }

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (Proc, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = plusE2(localViewAdds, p, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (Proc, Set[CONSTuple]) => bool = all {
        localViewDels' = plusEs2(localViewDels, p, es),
        localViewAdds' = minusEs2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = localViewAdds.get(p).exclude(localViewDels.get(p))

//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action GInstInit: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Gossip using P2P
    action gossipStep(p,q) = all {
        gossip(p,q),
    }

    action addTupleStep(p, e) = all {
            addTuple(p, e),
            //unchanged
            localViewDels' = localViewDels,
    }

    action step: bool = any {
        //chose partners
        nondet p = oneOf(Procs)
        nondet q = oneOf(minusE(Procs,p))

        gossipStep(p,q)
    }

/*
    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        localViewAdds' = Procs.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = Procs.mapBy(_ => Set()),
    }
*/

}

//This module checks for GOSSIP-I invariants on some implementation.
module GossipInv {
    import Globals.*

    import GossipInstCheat.* //as GI
//    import GossipNoP2P as G
//    import GossipP2P as G

//===== Invariants
//================================= TS-VALIDITY
    val NoDoubleMessage: bool = {
        Procs.forall(p =>
            Procs.forall(q =>
                val lvP = localView(p)
                lvP.forall(mp =>
                    val lvQ = localView(q)
                    lvQ.forall(mq =>
                        and (
                            mp.tag == mq.tag,
                            mp.height == mq.height,
                            mp.round == mq.round,
                            mp.signer == mq.signer
                        ).implies(mp.payload == mq.payload)
                    )
                )
            )
        )
    }

    val NoSpuriousProposals: bool = {
        Procs.forall(p =>
            Procs.forall(q =>
                val lvP = localView(p).filter(m => m.tag == "proposal")
                lvP.forall(mp =>
                    val lvQ = localView(q).filter(m => m.tag == "proposal")
                    lvQ.forall(mq =>
                        and (
                            mp.height == mq.height,
                            mp.round == mq.round
                        ).implies(mp.signer == mq.signer)
                    )
                )
            )
        )
    }

    temporal TSValidityProp: bool =
        always (and (NoDoubleMessage, NoSpuriousProposals))

//================================= REQ-CONS-GOSSIP
    //[REQ-CONS-GOSSIP-CONVERGENCE]
    temporal RCGConvergence = {
            globalView.forall(t =>
                eventually (
                    any {
                        Procs.forall(p => localView(p).contains(t)),
                        not (globalView.contains(t))
                    }
                )
            )
    }

    temporal RCGConvergenceProp = always (RCGConvergence)

}

/*

//This method implements the supersession of state, wherever it actually sits.
module SupersessionOracle {
//   import Globals.*
//    import GossipI.*

//    import GossipI.*
    import GossipCheat as G
    //TODO: remove once Quint supports it.
    type Proc = G::Proc
    val Procs = G::Procs
    type CONSTuple = G::CONSTuple



//===== These are related to how supersession and related could be implemented

    def supersededByNewHeight(p): Proc => Set[CONSTuple] = {
        //chose the highest height
        val highestHeight = G::localView(p).fold(0, (a, e) => if (e.height > a) e.height else a)
        //chose the entries with height smaller and return
        G::localView(p).filter(e => e.height < highestHeight)
    }

    // Removes superseded entries from the tuple space
    action forgetSuperseded(p): Proc => bool = all {
        val ssdByHeight = supersededByNewHeight(p)
        val ssdByNewRound = Set() //TODO: superSededByNewRound(p)
        val ssdByNewProof = Set() //TODO: superSededByNewProof(p)

        G::removeTuples(p, ssdByHeight.union(ssdByNewRound).union(ssdByNewProof)),
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    action step: bool = any {
        nondet p = oneOf(G::Procs)
        //forgetSuperseded(p),

        G::step
    }
}

*/


//==== Gossip Tests

module GossipTests {
    import Globals.*
    import P2PI.*

    import P2PAbstract.*
    //import P2PInstCheat.* //as PInst
    import P2PInstNetwork.* //as PInst
    import P2PTests.* //as PT

//    import GossipInstCheat.* //as GInst
    import GossipInstNoP2P.*
//    import GossipInstP2P.*


    action GTInit = all {
        PInstInit,
        GInstInit
    }

    action GTUnchangedAll = all {
        PInstUnchangedAll,
        GInstUnchangedAll
    }

//===== Runs/Debug
    run gossipInRingTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GTInit
            //connect some nodes
            .then(
                all {
                    connectRingStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
                PInstUnchangedAll
            })
            //gossip
            .then((all {
                val pqs = tuples(Procs,Procs).filter(t => t._1 != t._2 and localView(t._1) != localView(t._2))
                if (pqs.size() > 0) all {
                    val pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                    PInstUnchangedAll
                } else
                GTUnchangedAll
            }).repeated(size(Procs)))
            .then( all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            })
        }

    run gossipCliqueTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GTInit
            //connect some nodes
            .then(
                all {
                    connectFullStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
                PInstUnchangedAll
            })
            //gossip
            .then((all {
                val pqs = tuples(Procs,Procs).filter(t => t._1 != t._2 and localView(t._1) != localView(t._2))
                if (pqs.size() > 0) all {
                    val pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                    PInstUnchangedAll
                } else
                GTUnchangedAll
            }).repeated(size(Procs)))
            .then( all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            })
        }

    run gossipInStarTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GTInit
            //connect some nodes
            .then(
                all {
                    connectStarStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
                PInstUnchangedAll
            })
            //gossip
            .then((all {
                val pqs = tuples(Procs,Procs).filter(t => t._1 != t._2 and localView(t._1) != localView(t._2))
                if (pqs.size() > 0) all {
                    val pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                    PInstUnchangedAll
                } else
                GTUnchangedAll
            }).repeated(size(Procs)))
            .then( all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            })
        }

    run gossipInRandomTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GTInit
            //connect some nodes
            .then(
                all {
                    connectRandomStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
                PInstUnchangedAll
            })
            //gossip
            .then((all {
                val pqs = tuples(Procs,Procs).filter(t => t._1 != t._2 and localView(t._1) != localView(t._2))
                if (pqs.size() > 0) all {
                    val pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                    PInstUnchangedAll
                } else
                GTUnchangedAll
            }).repeated(size(Procs)))
            .then( all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            })
        }


}

/*


module ConsensusMock {
    import Globals.*
    import GossipAbstract.*

    action randomAddEntryStep() = all {
        nondet p = oneOf(Procs)
        nondet tag = oneOf(Set("proposal","prevote","precommit"))
        nondet height = 1.to(10).oneOf()
        nondet round = 1.to(10).oneOf()
        val entry = {tag:tag, height:height, round:round, payload:"Payload"}

        addEntryStep(p,entry),
    }

}

        action randomEntryInit: bool = all {
            val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
            globalViewAdds' = Set(entry),
            globalViewDels' = Set(),
        }

    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        localViewAdds' = Procs.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = Procs.mapBy(_ => Set()),
    }

*/


/*
To test this spec

- quint test reactor.qnt //Runs all tests
- quint test gossip.qnt --main GossipTests --verbosity=3
- quint run --max-samples 10000 --max-steps 10 --invariant RGPIgnoring reactor.qnt
*/


