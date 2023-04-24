module Helpers {
    //===== Pure Helper functions
    pure def minusE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def minusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).exclude(Set(e)))
    pure def minusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).exclude(es))

    pure def plusE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def plusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).union(Set(e)))
    pure def plusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).union(es))

    pure def print(msg, toPrint) = true
}

module Globals {
    type ProcId = int

    //XXX: Ideally this should be a const, but because it would have to be
    //instantiated on every import, there would be duplication.
    //All processes in the system
    val ProcIds = Set(1,2,3,4,5)
    //Maximum number of outbound connections per node.
    val MaxConNum = 4
}


//Defines the P2P Interface
module P2PI {
 //====== [VOC-GOSSIP-P2P-TYPES] P2P-I
    type GOSSIPMessage =
        | {tag: "query" }
        | {tag: "query_response", tuples: Set[{tag:str, height: int, round:int, signer:int}] }
        | {tag: "update", tuples: Set[{tag:str, height:int, round:int, payload:str, signer:int}]}

//==== Exported methods P2P-I
    //action send(p,q,m): (ProcId, ProcId, GOSSIPMessage) => bool
    //def received(p): Set[GOSSIPMessage]

}

//Provides staandard behavior for P2PI implementations
module P2PAbstract {
    import Helpers.*
    import Globals.*
    import P2PI.*


//==== Common to implementationtions, but probably not actually part of the P2P-I.
    //===== P2P Internals                                       //TODO: Move some of these to P2PI
    var igProcs : ProcId -> Set[ProcId]                                 //Nodes to be ignored.
    var cnProcs  : ProcId -> Set[ProcId]                                 //Neighbor sets

    var sMsgs: Set[{src:ProcId, dst:ProcId, msg: GOSSIPMessage}]            //Unicast sent messages.
    var rMsgs: Set[{src:ProcId, dst:ProcId, msg: GOSSIPMessage}]          //Unicast received messages

    //Helper
    val p2pInternals = {ignore: igProcs, connections: cnProcs, sent:sMsgs, received:rMsgs}

    action PAunchangedAll = all {
        igProcs' = igProcs,
        cnProcs' = cnProcs,
        rMsgs' = rMsgs,
        sMsgs' = sMsgs
    }



    //Adds bidirectional connection if both nodes are below limit.
    action addPeer(p,q): (ProcId,ProcId) => bool = all {
        not(cnProcs.get(p).contains(q)),
        size(cnProcs.get(p)) < MaxConNum,
        not(cnProcs.get(q).contains(p)),
        size(cnProcs.get(q)) < MaxConNum,
        cnProcs' = plusE2(plusE2(cnProcs, p, q), q, p)
    }

    action removePeer(p,q): (ProcId,ProcId) => bool = all {
        cnProcs' = minusE2(minusE2(cnProcs, p, q), q, p)
    }


    action ignoreNode(p, q): (ProcId, ProcId) => bool = {
        igProcs' = plusE2(igProcs, p, q)
    }

    action doNotIgnoreNode(p, q): (ProcId, ProcId) => bool = {
        igProcs' = minusE2(igProcs, p, q)
    }

//==== Init
    action PAinit: bool = all {
        igProcs' = ProcIds.mapBy(_ => Set()),
        cnProcs' = ProcIds.mapBy(_ => Set()),
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
        ProcIds.forall(p =>
                cnProcs.get(p).size() < MaxConNum
        )
    }

    temporal RGPConcurrentProp: bool =
        always (RGPConcurrent)

    //[REQ-GOSSIP-P2P-IGNORING]
    val RGPIgnoring: bool = {
        ProcIds.forall(p =>
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
    action send(p,q,m): (ProcId, ProcId, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        val msg = {src:p, dst:q, msg:m}
        all {
            not(msg.in(rMsgs)),
            sMsgs' = plusE(sMsgs, msg),
            rMsgs' = plusE(rMsgs, msg)
        }
    }

    def received(p): ProcId => Set[{src:ProcId, dst:ProcId, msg: GOSSIPMessage}] =
        rMsgs.filter(m => m.dst == p)


//==== Init
    action PInstInit: bool = PAinit


//===== Steps
    action receiveMessageStep = all {PInstUnchangedAll, true}

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


    type OutGoingMsg = {dst:ProcId, msg: GOSSIPMessage}
    type IncomingMsg = {src:ProcId, msg: GOSSIPMessage}

    action PInstUnchangedAll = PAunchangedAll

    def nonDelMsgs(p): ProcId => Set[{src: ProcId, dst: ProcId, msg:GOSSIPMessage}] = {
        sMsgs.filter(m => m.src == p and not(m.in(rMsgs)))
    }

//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (ProcId, ProcId, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        val msg = {src:p, dst:q, msg:m}
        all {
            print("sending", msg),
            not(msg.in(sMsgs)),
            sMsgs' = plusE(sMsgs, msg),
        },
        rMsgs' = rMsgs
    }

    def received(p) = rMsgs.filter(m => m.dst == p)


//======Init
    action PInstInit: bool = PAinit

//===== Step
    action receiveMessageStep = {
        val nonDelMsgsAll = sMsgs.filter(m => not(m.in(rMsgs)) and m.dst.in(cnProcs.get(m.src)))
        all {
            size(nonDelMsgsAll) > 0,
            nondet m = nonDelMsgsAll.oneOf()
            all {
                print("receiving", m),
                rMsgs' = plusE(rMsgs, m),
                sMsgs' = sMsgs,

                igProcs' = igProcs,
                cnProcs' = cnProcs
            }
        }
    }

    action PInstStep = any {
        receiveMessageStep,
    }
}


module P2PTests {
    import Helpers.*
    import Globals.*
    import P2PAbstract.*

    //import P2PInstCheat.* //as PInst
    import P2PInstNetwork.* //as PInst

//======== Tests

    //Connects nodes node towards forming a ring.
    action connectRingStep() = {
        all {
            cnProcs' = ProcIds.mapBy(p => if (p == 1) Set(ProcIds.size(), 2) else if (p == ProcIds.size()) Set(1, p-1) else Set(p-1, p+1)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectFullStep() = {
        all {
            cnProcs' = ProcIds.mapBy(p => ProcIds),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectStarStep() = {
        all {
            cnProcs' = ProcIds.mapBy(p => if (p ==1 ) ProcIds.minusE(1) else Set(1)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

    action connectRandomStep() = {
        //TODO: Describe a random connected graph, without iteration, efficiently
        //nondet f = ProcIds.setOfMaps(powerset(ProcIds)).oneOf()
        //The graph below is actually a double link ring.
        all {
            cnProcs' = ProcIds.mapBy(p => if (p < size(ProcIds)-1) Set(p+1, p+2) else if (p == size(ProcIds)-1) Set(p+1,1) else Set(1,2)),
            igProcs' = igProcs,

            sMsgs' = sMsgs,
            rMsgs' = rMsgs
        }
    }

//init and step, used for random walks a fully configured "stack"
    action PTInit = PInstInit

    action PTStep = PInstStep

//===== Runs/Debug
    run sendReceiveInRingTopTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectRingStep)
            //send and receive messages
            .then(
                all {
                    send(1,2,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(2,3,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(receiveMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                all {
                    send(1,3,{tag: "query"}),

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
                    send(1,2,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(1,3,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(2,3,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(receiveMessageStep.repeated(3))
            .then(all {RGPUnicast, PInstUnchangedAll})
    }

    run sendReceiveInStarConnectedTest: bool = {
        PInstInit
            //connect some nodes
            .then(connectStarStep)
            //send and receive messages
            .then(
                all {
                    send(1,2,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                all {
                    send(1,3,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(receiveMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                all {
                    send(2,3,{tag: "query"}),

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
                nondet p = cnProcs.keys().filter(pa => cnProcs.get(pa).size() > 0 ).oneOf()
                nondet q = cnProcs.get(p).oneOf()
                all {
                    send(p,q,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(
                nondet p = cnProcs.keys().filter(pa => cnProcs.get(pa).size() > 0 ).oneOf()
                nondet q = cnProcs.get(p).oneOf()
                all {
                    send(p,q,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }
            )
            .then(receiveMessageStep.repeated(2))
            .then(all {RGPUnicast, PInstUnchangedAll})
            .then(
                nondet p = cnProcs.keys().filter(pa => cnProcs.get(pa) != ProcIds ).oneOf()
                nondet q = ProcIds.exclude(cnProcs.get(p)).oneOf()
                all {
                    send(p,q,{tag: "query"}),

                    cnProcs' = cnProcs,
                    igProcs' = igProcs,
                }.fail()
            )

    }

//===== Test Steps
    //Update the connections
    action randomDropConnectionStep() = all {
        ProcIds.filter(ap => cnProcs.get(ap).size() > 0).size() > 0,
        nondet p = ProcIds.filter(ap => cnProcs.get(ap).size() > 0).oneOf()
        nondet q = cnProcs.get(p).oneOf()

        removePeer(p,q),
        igProcs' = igProcs,
        sMsgs' = sMsgs,
        rMsgs' = rMsgs
    }

    action randomConnectionStep() = all {
        //Are there two processes that can still connect to each other.
        val nonFull = ProcIds.filter(p => cnProcs.get(p).size() < MaxConNum)
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
        nondet p = ProcIds.oneOf()
        nondet q = ProcIds.oneOf()
        all {
            send(p,q,{tag: "query"}),

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
        | { tag: "proposal",  height: int, round: int, payload: str, signer: ProcId}
        | { tag: "prevote",   height: int, round: int, payload: str, signer: ProcId}
        | { tag: "precommit", height: int, round: int, payload: str, signer: ProcId}

//==== Exported methods GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    //action addTuple(p,t): (ProcId, CONSTuple) => bool

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    //action removeTuples(p, es): (ProcId, Set[CONSTuple]) => bool

    //[DEF-READ-TUPLE]
    //The local view
    //def localView(p): ProcId => Set[CONSTuple]
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
    action addTuple(p: ProcId, t: CONSTuple): bool = all {
        not(globalViewDels.contains(t)),
        globalViewAdds' = plusE(globalViewAdds, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p: ProcId, es: Set[CONSTuple]): bool = all {
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
    var localViewAdds : ProcId -> Set[CONSTuple]
    var localViewDels : ProcId -> Set[CONSTuple]

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
    action addTuple(p,t): (ProcId, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = plusE2(localViewAdds, p, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (ProcId, Set[CONSTuple]) => bool = all {
        localViewDels' = plusEs2(localViewDels, p, es),
        localViewAdds' = minusEs2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = localViewAdds.get(p).exclude(localViewDels.get(p))



//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action GInstInit: bool = all {
        localViewAdds' = ProcIds.mapBy(_ => Set()),
        localViewDels' = ProcIds.mapBy(_ => Set())
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
        nondet p = oneOf(ProcIds)
        nondet q = oneOf(minusE(ProcIds,p))

        gossipStep(p,q)
    }


}


//This module implements GOSSIP-I using P2P-I. Nodes converge by exchanging messages.
module GossipInstP2P {
    import Helpers.*
    import Globals.*

    import P2PI.*
    import P2PAbstract.*

    //import P2PInstCheat.* //as PInst
    import P2PInstNetwork.* //as PInst

    import GossipI.*
    import GossipAbstract.* //as GA

    var localViewAdds : ProcId -> Set[CONSTuple]
    var localViewDels : ProcId -> Set[CONSTuple]
    var processed: Set[{src:ProcId, dst:ProcId, msg: GOSSIPMessage}]

    //Helper
    def globalView: Set[CONSTuple] = {
        val allAdds = localViewAdds.keys().fold(Set(), (acc, k) => acc.union(localViewAdds.get(k)))
        val allDels = localViewDels.keys().fold(Set(), (acc, k) => acc.union(localViewDels.get(k)))

        allAdds.exclude(allDels)
    }

    val fullState = {adds: localViewAdds, dels: localViewDels}

    //The local view
    def localView(p) = localViewAdds.get(p).exclude(localViewDels.get(p))

    //Unchanged
    action GInstUnchangedAll() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
        processed' = processed
    }



    //send query message
    action sendQuery(p: ProcId, q: ProcId): bool = all{
        send(p,q, {tag: "query"}),
        cnProcs' = cnProcs,
        igProcs' = igProcs,
        GInstUnchangedAll,
    }

    def entryToSummary(e: {tag:str, height:int, round:int, payload:str, signer:int}): {tag:str, height:int, round:int, signer:int} = {
        {tag:e.tag, height:e.height, round:e.round, signer:e.signer}
    }

    def getMySummary(p: ProcId): Set[{tag:str, height:int, round:int, signer:int}] = {
        localView(p).map(e => entryToSummary(e))
    }

    def getMyTuples(p: ProcId): Set[{tag:str, height:int, round:int, payload:str, signer:int}] = {
        localView(p)
    }

    //send query response
    action respondQuery(p: ProcId, q: ProcId): bool = all{
        nondet m = received(q)
                .exclude(processed)
                .filter(e => e.src == p)
                .filter(e => e.msg.tag == "query")
                .oneOf()
        val myTuples = getMySummary(q)
        all {
            send(q,p, {tag: "query_response", tuples: myTuples}),

            cnProcs' = cnProcs,
            igProcs' = igProcs,
            localViewAdds' = localViewAdds,
            localViewDels' = localViewDels,
            processed' = plusE(processed, m),
        }
    }

    //receive query response
    //identify differences
    //send differences
    action processResponse(p: ProcId, q: ProcId): bool = all{
        nondet m = received(p).exclude(processed).filter(e => e.src == q).filter(e => e.msg.tag == "query_response").oneOf()

        val myTuples = getMyTuples(p)
        val myUpdates = myTuples.filter(e => not ( m.msg.tuples.contains(entryToSummary(e))))
        all {
            send(p,q, {tag: "update", tuples: myUpdates}),
            cnProcs' = cnProcs,
            igProcs' = igProcs,
            localViewAdds' = localViewAdds,
            localViewDels' = localViewDels,
            processed' = plusE(processed, m),
        }
    }

    //receive differences
    //merge
    action processDiff(p: ProcId, q: ProcId): bool = all{
        val ms = received(q)
                .exclude(processed)
                .filter(e => e.src == p)
                .filter(e => e.msg.tag == "update")

        all {
            print("processDiff", ms),
            ms != Set(),
            nondet m = ms.oneOf()
            all {
                localViewAdds' = plusEs2(localViewAdds, q, m.msg.tuples),
                processed' = plusE(processed, m),
            },
            localViewDels' = localViewDels,
            PInstUnchangedAll,
        }
    }

    // This is a version of gossip that uses messages from the P2P layer.
    action gossip(p,q) = any {
        sendQuery(p,q),
        respondQuery(p,q),
        processResponse(p,q),
        processDiff(p,q),
    }

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (ProcId, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = plusE2(localViewAdds, p, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (ProcId, Set[CONSTuple]) => bool = all {
        localViewDels' = plusEs2(localViewDels, p, es),
        localViewAdds' = minusEs2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //See localView(p), above

//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action GInstInit: bool = all {
        localViewAdds' = ProcIds.mapBy(_ => Set()),
        localViewDels' = ProcIds.mapBy(_ => Set()),
        processed' = Set(),

        PInstInit,
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Gossip using P2P
    action p2pStep = all {
            PInstStep,
            GInstUnchangedAll,
        }

    action gossipStep(p,q) = any {
        gossip(p,q),
    }


    action addTupleStep(p, e) = all {
            addTuple(p, e),
            //unchanged
            localViewDels' = localViewDels,
            processed' = processed,
            PAunchangedAll,
    }

    action step: bool = any {
        p2pStep,

        //chose partners
        nondet p = oneOf(ProcIds)
        nondet q = oneOf(minusE(ProcIds,p))
        gossipStep(p,q),
    }
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
        ProcIds.forall(p =>
            ProcIds.forall(q =>
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
        ProcIds.forall(p =>
            ProcIds.forall(q =>
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
                        ProcIds.forall(p => localView(p).contains(t)),
                        not (globalView.contains(t))
                    }
                )
            )
    }

    temporal RCGConvergenceProp = always (RCGConvergence)

}

/*

//This module implements the supersession of state, wherever it actually sits.
module SupersessionOracle {
//   import Globals.*
//    import GossipI.*

//    import GossipI.*
    import GossipCheat as G
    //TODO: remove once Quint supports it.
    type ProcId = G::ProcId
    val ProcIds = G::ProcIds
    type CONSTuple = G::CONSTuple



//===== These are related to how supersession and related could be implemented

    def supersededByNewHeight(p): ProcId => Set[CONSTuple] = {
        //chose the highest height
        val highestHeight = G::localView(p).fold(0, (a, e) => if (e.height > a) e.height else a)
        //chose the entries with height smaller and return
        G::localView(p).filter(e => e.height < highestHeight)
    }

    // Removes superseded entries from the tuple space
    action forgetSuperseded(p): ProcId => bool = all {
        val ssdByHeight = supersededByNewHeight(p)
        val ssdByNewRound = Set() //TODO: superSededByNewRound(p)
        val ssdByNewProof = Set() //TODO: superSededByNewProof(p)

        G::removeTuples(p, ssdByHeight.union(ssdByNewRound).union(ssdByNewProof)),
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    action step: bool = any {
        nondet p = oneOf(G::ProcIds)
        //forgetSuperseded(p),

        G::step
    }
}

*/


//==== Gossip Tests

module GossipTests {
    import Helpers.*
    import Globals.*
    import P2PI.*

    import P2PAbstract.*
    //import P2PInstCheat.* //as PInst
    //import P2PInstNetwork.* //as PInst
    import P2PTests.* //as PT

//    import GossipInstCheat.* //as GInst
//    import GossipInstNoP2P.*
    import GossipInstP2P.*

    var stepCounter: int

    action testInit = all {
        stepCounter' = 1
    }

    action testUnchangedAll = {stepCounter' = stepCounter}
    action testIncrement = {stepCounter' = stepCounter + 1}

    def gossipablePairs =
        tuples(ProcIds,ProcIds).filter(t => and (
        t._1 != t._2,
        localView(t._1) != localView(t._2),
        cnProcs.get(t._1).contains(t._2)
    ))


//===== Runs/Debug
    run gossipInRingTopoTest: bool = {
        val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all { GInstInit, testInit})
            //connect some nodes
            .then(
                all {
                    connectRingStep,
                    GInstUnchangedAll,
                    testUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(signer, entry),
                testUnchangedAll
            })
            //gossip
            .then(( all{
                testIncrement,
                print("outer step", stepCounter),
                val pqs = gossipablePairs
                if (pqs.size() > 0) all {
                    print ("step yeah", pqs),
                    print("local views", localViewAdds),
                    any{
                        nondet pq = pqs.oneOf()
                        all {
                            print("pq", pq),
                            gossipStep(pq._1, pq._2), //Send or process received messages.
                        },
                        p2pStep                   //Receive messages
                    },
                } else all {
                    print ("step nope", pqs),
                    print("local views", localViewAdds),
                    //Either it progresses or all have the entry.
                    //ProcIds.forall(p => localView(p).contains(entry)),
                    PAunchangedAll,
                    GInstUnchangedAll
                }
            //*2 gossip flows per link. *7 steps needed per gossip.
            }).repeated(ProcIds.size()*2*7))
            .then( all {
                //All see the entry
                ProcIds.forall(p => localView(p).contains(entry)),
                PAunchangedAll,
                GInstUnchangedAll,
                testUnchangedAll
            })
       }

    run gossipCliqueTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GInstInit
            //connect some nodes
                .then(
                all {
                    connectFullStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
            })
            //gossip
            .then((all {
                val pqs = gossipablePairs
                if (pqs.size() > 0) all {
                    nondet pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                } else {
                    all {
                        PAunchangedAll,
                        GInstUnchangedAll
                    }
                }
            }).repeated(size(ProcIds)))
            .then( all {
                //All see the entry
                ProcIds.forall(p => localView(p).contains(entry)),
                PAunchangedAll,
                GInstUnchangedAll
            })
        }

    run gossipInStarTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GInstInit
/*             //connect some nodes
            .then(
                all {
                    connectStarStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
            })
            //gossip
            .then((all {
                val pqs = gossipablePairs
                if (pqs.size() > 0) all {
                    nondet pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                } else
                GTUnchangedAll
            }).repeated(size(ProcIds)))
            .then( all {
                //All see the entry
                ProcIds.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            }) */
        }

    run gossipInRandomTopoTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        GInstInit
/*             //connect some nodes
            .then(
                all {
                    connectRandomStep,
                    GInstUnchangedAll
                })
            //add one entry
            .then(all {
                addTupleStep(1, entry),
            })
            //gossip
            .then((all {
                val pqs = gossipablePairs
                if (pqs.size() > 0) all {
                    nondet pq = pqs.oneOf()
                    gossipStep(pq._1, pq._2),
                } else
                GTUnchangedAll
            }).repeated(size(ProcIds)))
            .then( all {
                //All see the entry
                ProcIds.forall(p => localView(p).contains(entry)),
                GTUnchangedAll
            }) */
        }


}

/*


module ConsensusMock {
    import Globals.*
    import GossipAbstract.*

    action randomAddEntryStep() = all {
        nondet p = oneOf(ProcIds)
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
        localViewAdds' = ProcIds.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = ProcIds.mapBy(_ => Set()),
    }

*/


/*
To test this spec

- quint test reactor.qnt //Runs all tests
- quint test gossip.qnt --main GossipTests --verbosity=3
- quint run --max-samples 10000 --max-steps 10 --invariant RGPIgnoring reactor.qnt
*/


