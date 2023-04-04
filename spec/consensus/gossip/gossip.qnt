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

    //All processes in the system
    //const Procs: Set[Proc]
    val Procs = Set(1,2,3,4,5)
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
    import Globals as G
    //TODO: remove once Quint supports it.
    type Proc = G::Proc
    val Procs = G::Procs

    import P2PI.*


//==== Common to implementationtions, but probably not actually part of the P2P-I.
    //===== P2P Internals                                       //TODO: Move some of these to P2PI
    var maxCon : Proc -> int                                    //Maximum number of outbound connections
    var igProcs : Proc -> Set[Proc]                                 //Nodes to be ignored.
    var cnProcs  : Proc -> Set[Proc]                                 //Neighbor sets

    var sMsgs: Proc -> Set[{dst:Proc, msg: GOSSIPMessage}]            //Unicast sent messages.
    var rMsgs: Proc -> Set[{src:Proc, msg: GOSSIPMessage}]            //Unicast received messages

    //Helper
    val p2pInternals = {maxCon: maxCon, ignore: igProcs, connections: cnProcs, sent:sMsgs, received:rMsgs}

    action unchangedAll = all {
        maxCon' = maxCon,
        igProcs' = igProcs,
        cnProcs' = cnProcs,
        rMsgs' = rMsgs,
        sMsgs' = sMsgs

    }

    action addPeer(p,q): (Proc,Proc) => bool = all {
        not(cnProcs.get(p).contains(q)),
        size(cnProcs.get(p)) < maxCon.get(p),
        cnProcs' = plusE2(cnProcs, p, q)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        cnProcs' = minusE2(cnProcs, p, q)
    }

    action setMaxCon(p, m): (int,int) => bool = {
        maxCon' = maxCon.set(p, m)
    }

    def maxConNum(p) = {
        maxCon.get(p)
    }

    action ignoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = plusE2(igProcs, p, q)
    }

    action doNotIgnoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = minusE2(igProcs, p, q)
    }

    def neighbors(p) = cnProcs.get(p)

//==== Init
    action init: bool = all {
        maxCon' = Procs.mapBy(_ => 3),
        igProcs' = Procs.mapBy(_ => Set()),
        cnProcs' = Procs.mapBy(_ => Set()),
        sMsgs' = Procs.mapBy(_ => Set()),
        rMsgs' = Procs.mapBy(_ => Set())
    }

//==== Steps
    //Update the connections
    action randomDropConnectionStep() = all {
        Procs.filter(ap => cnProcs.get(ap).size() > 0).size() > 0,
        nondet p = Procs.filter(ap => cnProcs.get(ap).size() > 0).oneOf()
        nondet q = cnProcs.get(p).oneOf()

        removePeer(p,q),
        maxCon' = maxCon,
        igProcs' = igProcs,
    }

    action randomConnectionStep() = all {
        //Are there two processes that can still connect to each other.
        val nonFull = Procs.filter(p => cnProcs.get(p).size() < maxCon.get(p))
        val nonCon = nonFull.filter(p => nonFull.exists(q => and(q != p , not (cnProcs.get(p).contains(q)))))
        all {
            nonCon.size() > 0,
            all {
                nondet p = nonCon.oneOf()
                nondet q = nonFull.minusE(p).filter(aq => not (cnProcs.get(p).contains(aq))).oneOf()

                addPeer(p,q),
                maxCon' = maxCon,
                igProcs' = igProcs,
            }
        }
    }

    action randomstep = any {
        randomConnectionStep,
        randomDropConnectionStep,
    }

    action step = unchangedAll

}



//This module implements P2P without using the network
// - everything sent is immediately delivered
// - there are no duplications
// - there are no losses
module P2PCheat {
    import Helpers.*
    //import Globals as G
    import P2PI.*

    import P2PAbstract as PA
    //TODO: remove once Quint supports it.
    type Proc = PA::Proc
    val Procs = PA::Procs
    val maxCon = PA::maxCon                                     //Maximum number of outbound connections
    val igProcs = PA::igProcs                                   //Nodes to be ignored.
    val cnProcs = PA::cnProcs                                   //Neighbor sets
    val sMsgs = PA::sMsgs                                       //Unicast sent messages.
    val rMsgs = PA::rMsgs                                       //Unicast received messages


//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(PA::cnProcs.get(p)),
        not({src:p, msg:m}.in(PA::rMsgs.get(q))),
        PA::sMsgs' = plusE2(PA::sMsgs, p, {dst:q, msg:m}),
        PA::rMsgs' = plusE2(PA::rMsgs, q, {src:q, msg:m})
    }

    def received(p): Proc => Set[{src:Proc, msg: GOSSIPMessage}] =
        PA::rMsgs.get(p)

    action unchangedAll = PA::unchangedAll

//==== Init
    action init: bool = PA::init


//===== Steps
    action deliverMessageStep = PA::unchangedAll

    action step: bool = PA::step
}


//This module implements P2P "using" the network
// - everything sent is put in a pending set
// - elements from the pending set are delivered randomly
// - there are no duplications
// - there are no losses
module P2PNetwork {
    import Helpers.*
    //import Globals.*
    import P2PI.*

    import P2PAbstract as PA
    //TODO: remove once Quint supports it.
    type Proc = PA::Proc
    val Procs = PA::Procs
    val maxCon = PA::maxCon                                     //Maximum number of outbound connections
    val igProcs = PA::igProcs                                   //Nodes to be ignored.
    val cnProcs = PA::cnProcs                                   //Neighbor sets
    val sMsgs = PA::sMsgs                                       //Unicast sent messages.
    val rMsgs = PA::rMsgs                                       //Unicast received messages


    type OutGoingMsg = {dst:Proc, msg: GOSSIPMessage}
    type IncomingMsg = {src:Proc, msg: GOSSIPMessage}

    //Helpers
    def outToInc(s, m): (Proc, OutGoingMsg) => IncomingMsg =
        {src:s, msg:m.msg}

    def incToOut(d, m): (Proc,IncomingMsg) => OutGoingMsg =
        {dst:d, msg:m.msg}

    def nonDelMsgs(p): Proc => Set[{dst: Proc, msg:GOSSIPMessage}] = {
        PA::sMsgs.get(p).filter(m => not(outToInc(p,m).in(PA::rMsgs.get(m.dst))))
    }

//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(PA::cnProcs.get(p)),
        PA::sMsgs' = plusE2(PA::sMsgs, p, {dst:q, msg:m}),
        PA::rMsgs' = PA::rMsgs
    }

    def received(p) = PA::rMsgs.get(p)

    action unchangedAll = PA::unchangedAll

//======Init
    action init: bool = PA::init

//===== Step
    action deliverMessageStep = {
        val p = PA::sMsgs.keys().filter(ap => nonDelMsgs(ap).size() > 0).oneOf()
        val m = nonDelMsgs(p).oneOf()
        all {
            PA::rMsgs' = plusE2(PA::rMsgs, m.dst, outToInc(p,m)),
            PA::sMsgs' = minusE2(PA::sMsgs, p, m),

            PA::maxCon' = PA::maxCon,
            PA::igProcs' = PA::igProcs,
            PA::cnProcs' = PA::cnProcs
        }
    }

    action step = any {
        PA::step,
        deliverMessageStep,
    }
}




module P2PInvariants {
    //import Globals.*
    //import P2PAbstract as PA

    import P2PCheat as P
    //import P2PNetwork as P
    //TODO: remove once Quint supports it.
    type Proc = P::Proc
    val Procs = P::Procs

//================================= REQ-GOSSIP-P2P
    //[REQ-GOSSIP-P2P-UNICAST]
    temporal RGPUnicastProp = {
        Procs.forall(p =>
            P::cnProcs.get(p).forall(q =>
                P::sMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                P::rMsgs.get(q).contains({src:p, msg:m.msg}),
                                not (P::cnProcs.get(p).contains(q))
                            }
                        )
                    )
                )
            )
        )
    }


    //[REQ-GOSSIP-P2P-CONCURRENT_CONN]
    val RGPConcurrent = {
        Procs.forall(p =>
                P::cnProcs.get(p).size() < P::maxCon.get(p)
        )
    }

    temporal RGPConcurrentProp: bool =
        always (RGPConcurrent)

    //[REQ-GOSSIP-P2P-IGNORING]
    val RGPIgnoring: bool = {
        Procs.forall(p =>
            val cnP = P::cnProcs.get(p)
            val igP = P::igProcs.get(p)
            val igCn = cnP.intersect(igP)
                size(igCn) == 0
        )
    }

    temporal RGPIgnoringProp =
        always (RGPIgnoring)


    action step = P::step

    action init = P::init
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
    import Globals as G
    //TODO: remove once Quint supports it.
    type Proc = G::Proc
    val Procs = G::Procs

    import GossipI as GI
    //TODO: remove once Quint supports it.
    type CONSTuple = GI::CONSTuple


}

// This module implements GOSSIP-I without really gossiping. All tuples are immediately available to all.
module GossipCheat {
    import Helpers.*
    //import Globals as G

//    import GossipI.*
    import GossipAbstract as GA
    //TODO: remove once Quint supports it.
    type Proc = GA::Proc
    val Procs = GA::Procs
    type CONSTuple = GA::CONSTuple

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
    action unchangedGossip() = all {
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
    action init: bool = all {
        globalViewAdds' = Set(),
        globalViewDels' = Set()
    }

    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        globalViewAdds' = Set(entry),
        globalViewDels' = Set(),
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    action addEntryStep(p, e) = all {
            addTuple(p, e),
            //unchanged
            globalViewDels' = globalViewDels,
    }

    //Step does not need to do anything, since everything is imediatelly gossiped.
    action step: bool = true
}

// This module implements GOSSIP-I without using P2P. Nodes compare their local views and converge 2-by-2.
module GossipNoP2P {
    import Helpers.*
    //import Globals as G

//    import GossipI.*
    import GossipAbstract as GA
    //TODO: remove once Quint supports it.
    type Proc = GA::Proc
    val Procs = GA::Procs
    type CONSTuple = GA::CONSTuple

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
    action unchangedGossip() = all {
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
    action init: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }

    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        localViewAdds' = Procs.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = Procs.mapBy(_ => Set()),
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Gossip using P2P
    action gossipStep(p,q) = all {
        gossip(p,q),
    }

    action addEntryStep(p, e) = all {
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

    action stutter = all {
        unchangedGossip,
    }
}


//This module implements GOSSIP-I using P2P-I. Nodes converge by exchanging messages.
module GossipP2P {

    import Helpers.*
    //import Globals as G

//    import GossipI.*
    import GossipAbstract as GA
    //TODO: remove once Quint supports it.
    type Proc = GA::Proc
    val Procs = GA::Procs
    type CONSTuple = GA::CONSTuple

    import P2PI.*

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
    action unchanged() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
    }


    // This is a version of gossip that uses messages from the P2P layer.
    action gossip(p,q) = all {
        //TODO
        //send query message
        //send query response
        //receive query response
        //identify differences
        //send differences
        //receive differences
        //merge

        //TODO: temporary unchanged
        unchanged,
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
    action init: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }

    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        localViewAdds' = Procs.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = Procs.mapBy(_ => Set()),
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Gossip using P2P
    action gossipStep(p,q) = all {
        gossip(p,q),
    }
    action addEntryStep(p, e) = all {
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

    action stutter = all {
        unchanged,
    }
}

//This module checks for GOSSIP-I invariants on some implementation.
module GossipInv {
    //import Globals.*

    import GossipCheat as G
//    import GossipNoP2P as G
//    import GossipP2P as G

//===== Invariants
//================================= TS-VALIDITY
    val NoDoubleMessage: bool = {
        G::Procs.forall(p =>
            G::Procs.forall(q =>
                val lvP = G::localView(p)
                lvP.forall(mp =>
                    val lvQ = G::localView(q)
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
        G::Procs.forall(p =>
            G::Procs.forall(q =>
                val lvP = G::localView(p).filter(m => m.tag == "proposal")
                lvP.forall(mp =>
                    val lvQ = G::localView(q).filter(m => m.tag == "proposal")
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
        always (
            G::globalView.forall(t =>
                eventually (
                    any {
                        G::Procs.forall(p => G::localView(p).contains(t)),
                        not (G::globalView.contains(t))
                    }
                )
            )
        )
    }

}


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



//========= P2P Tests
module P2PTestHelper {
    import Helpers.*
    //import Globals.*
    //import P2PI.*
    //import P2PAbstract as PA

    import P2PCheat as P
    //import P2PNetwork as P

    action connectRingStep() = {
        all {
            P::cnProcs' = P::Procs.mapBy(p => if (p < size(P::Procs)) Set(p+1) else Set(0)),
            P::maxCon' = P::maxCon,
            P::igProcs' = P::igProcs,

            P::sMsgs' = P::sMsgs,
            P::rMsgs' = P::rMsgs
        }
    }

    action connectFullStep() = {
        all {
            P::cnProcs' = P::Procs.mapBy(p => P::Procs),
            P::maxCon' = P::maxCon,
            P::igProcs' = P::igProcs,

            P::sMsgs' = P::sMsgs,
            P::rMsgs' = P::rMsgs
        }
    }

    action connectStarStep() = {
        all {
            P::cnProcs' = P::Procs.mapBy(p => if (p ==1 ) P::Procs.minusE(1) else Set(1)),
            P::maxCon' = P::maxCon,
            P::igProcs' = P::igProcs,

            P::sMsgs' = P::sMsgs,
            P::rMsgs' = P::rMsgs
        }
    }

    action connectRandomStep() = {
        //TODO: Describe a random connected graph, without iteration, efficiently
        //nondet f = Procs.setOfMaps(powerset(Procs)).oneOf()
        //The graph below is actually a double link ring.
        all {
            P::cnProcs' = P::Procs.mapBy(p => if (p < size(P::Procs)-1) Set(p+1, p+2) else if (p == size(P::Procs)-1) Set(p+1,1) else Set(1,2)),
            P::maxCon' = P::maxCon,
            P::igProcs' = P::igProcs,

            P::sMsgs' = P::sMsgs,
            P::rMsgs' = P::rMsgs
        }
    }
}


module P2PTests {
    import Helpers.*
    //import Globals(Procs = Set(1,2,3,4,5)) as G
    //import P2PI.*

    //import P2PAbstract as PA

    import P2PTestHelper.*

    import P2PCheat as P
    //import P2PNetwork as P

//init and step, used for random walks a fully configured "stack"
    action init = P::init

    action step = P::step

//===== Runs/Debug
    run NOOPTest: bool = {
        P::init
    }

    run sendReceiveInARingTest: bool = {
        P::init
            //connect some nodes
            .then(connectRingStep)
            //send and receive messages
            .then(
                all {
                    P::send(1,2,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(
                all {
                    P::send(2,3,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(P::deliverMessageStep).repeated(2)
            .then(all {
                size(P::Procs.filter(p => size(P::received(p)) > 0)) == 2,

                P::unchangedAll
            })
    }

    run sendReceiveInFullyConnectedTest: bool = {
        P::init
            //connect some nodes
            .then(connectFullStep)
            //send and receive messages
            .then(
                all {
                    P::send(1,2,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(
                all {
                    P::send(2,3,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(P::deliverMessageStep).repeated(2)
            .then(all {
                size(P::Procs.filter(p => size(P::received(p)) > 0)) == 2,

                P::unchangedAll
            })
    }

    run sendReceiveInStarConnectedTest: bool = {
        P::init
            //connect some nodes
            .then(connectStarStep)
            //send and receive messages
            .then(
                all {
                    P::send(1,2,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(
                all {
                    P::send(2,3,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(P::deliverMessageStep).repeated(2)
            .then(all {
                size(P::Procs.filter(p => size(P::received(p)) > 0)) == 2,

                P::unchangedAll
            })
    }

    run sendReceiveInRandomConnectedTest: bool = {
        P::init
            //connect some nodes
            .then(connectRandomStep)
            //send and receive messages
            .then(
                all {
                    P::send(1,2,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(
                all {
                    P::send(2,3,{gat: "query"}),

                    P::cnProcs' = P::cnProcs,
                    P::maxCon' = P::maxCon,
                    P::igProcs' = P::igProcs,
                }
            )
            .then(P::deliverMessageStep).repeated(2)
            .then(all {
                size(P::Procs.filter(p => size(P::received(p)) > 0)) == 2,

                P::unchangedAll
            })
    }
}



/*
*/
/*

//==== Gossip Tests

module GossipRingTests {
    import Globals.*
    import P2PI.*

    import P2PAbstract as P
    //import P2PNetwork as P

    import GossipGlobal(Param = true) as G
//    import GossipAbstract(Param(Param = true) as G
//    import GossipP2P(Param = true) as G

    action connectStep() = {
        all {
            P::cnProcs' = Procs.mapBy(p => if (p < size(Procs)) Set(p+1) else Set(0)),
            P::maxCon' = P::maxCon,
            P::igProcs' = P::igProcs,
            G::unchangedGossip,
        }
    }

//===== Runs/Debug
    action stutter = all {
        P::stutter,
        G::stutter,
    }

    run sendReceiveInARingTest: bool = {
        G::init
            //connect some nodes
            .then(connectStep)
            //send and receive messages
            //TODO
    }

    run gossipInARingTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        G::init
            //connect some nodes
            .then(connectStep)
            //add one entry
            .then(G::addEntryStep(1, entry))
            //gossip
            .then(G::gossipStepCheat(1,2))
            .then(G::gossipStepCheat(2,3))
            .then(G::gossipStepCheat(3,4))
            .then(G::gossipStepCheat(4,5))
            .then(G::gossipStepCheat(1,2))
            .then(all {
                //All see the entry
                G::Procs.forall(p => G::localView(p).contains(entry)),
                stutter
            })
    }
}


module GossipFullTests {
    import GossipAbstract(Procs = Set(1,2,3,4,5)) as G

//===== Runs/Debug
    action stutter = all {
        G::stutter,
    }

    action connectStep() = {
        all {
            G::cnProcs' = G::Procs.mapBy(p => G::Procs),
            G::maxCon' = G::maxCon,
            G::igProcs' = G::igProcs,
            G::unchangedGossip,
        }
    }

    run sendReceiveInFullyConnectedTest: bool = {
        G::init
            //connect some nodes
            .then(connectStep)
            //send and receive messages
            //TODO
    }

   run gossipInFullTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        G::init
            //connect some nodes
            .then(connectStep)
            //add one entry
            .then(G::addEntryStep(1, entry))
            //gossip
            .then(G::gossipStepCheat(1,2))
            .then(G::gossipStepCheat(1,3))
            .then(G::gossipStepCheat(3,4))
            .then(G::gossipStepCheat(3,5))
            .then(all {
                //All see the entry
                G::Procs.forall(p => G::localView(p).contains(entry)),
                stutter
            })
    }
}




module GossipStarTests {

    import GossipAbstract(Procs = Set(1,2,3,4,5)) as G

    action connectStep() = {
        all {
            G::cnProcs' = G::Procs.mapBy(p => if (p ==1 ) G::Procs.G::minusE(1) else Set(1)),
            G::maxCon' = G::maxCon,
            G::igProcs' = G::igProcs,
            G::unchangedGossip,
        }
    }

//===== Runs/Debug
    action stutter = all {
        G::stutter,
    }

    run gossipStarTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        G::init
            //connect some nodes
            .then(connectStep)
            //add one entries
            .then(G::addEntryStep(1, entry))
            //gossip
            .then(G::step)
            .repeated(5)
            .then(all {
                //Some see the entry
                G::Procs.G::minusE(1).exists(p => G::localView(p).contains(entry)),
                stutter
            })
    }

    run gossipRandomTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        G::init
            //connect some nodes
            .then(connectStep)
            //add one entries
            .then(G::addEntryStep(1, entry))
            //gossip
            .then(G::step)
            .repeated(5)
            .then(all {
                //Some see the entry
                G::Procs.G::minusE(1).exists(p => G::localView(p).contains(entry)),
                stutter
            })
    }

}


module GossipRandomTests {

    import GossipAbstract(Procs = Set(1,2,3,4,5)) as G

//===== Runs/Debug
    action stutter = all {
        G::stutter,
    }

    action randomConnectStep() = {
        //TODO: Describe a random connected graph, without iteration, efficiently
        //nondet f = Procs.setOfMaps(powerset(Procs)).oneOf()
        all {
            G::cnProcs' = G::Procs.mapBy(p => if (p < size(G::Procs)-1) Set(p+1, p+2) else if (p == size(G::Procs)-1) Set(p+1,1) else Set(1,2)),
            G::maxCon' = G::maxCon,
            G::igProcs' = G::igProcs,
            G::unchangedGossip,
        }
    }


    run gossipRandomTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        G::init
            //connect some nodes
            .then(randomConnectStep)
            //add one entries
            .then(G::addEntryStep(1, entry))
            //gossip
            .then(G::step)
            .repeated(5)
            .then(all {
                //Some see the entry
                G::Procs.G::minusE(1).exists(p => G::localView(p).contains(entry)),
                stutter
            })
    }

}


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

/*
To test this spec

- quint test reactor.qnt //Runs all tests
- quint run --max-samples 10000 --max-steps 10 --invariant RGPIgnoring reactor.qnt
*/
