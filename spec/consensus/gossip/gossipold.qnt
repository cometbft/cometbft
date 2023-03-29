module reactor {
    //TODO Reorganize once order does not matter.


//===== Global parameters
    type Proc = int

    //const Procs: Set[Proc]                                    //All processes in the system //TODO: Use once fully supported.
    pure val Procs = Set(1,2,3,4,5)

    type Payload = str


//===== Pure Helper functions
    pure def minusE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def minusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).exclude(Set(e)))
    pure def minusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).exclude(es))

    pure def plusE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def plusE2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).union(Set(e)))
    pure def plusEs2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).union(es))




//====== [VOC-CONS-GOSSIP-TYPES] GOSSIP-I
    type CONSTuple =                                          //TODO: Define payloads?
        | { tag: "proposal",  height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "prevote",   height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "precommit", height: int, round: int, payload: Payload, signer: Proc}

 //====== [VOC-GOSSIP-P2P-TYPES] P2P-I
    type GOSSIPMessage =
        | {tag: "query" }
        | {tag: "query_response" }
        | {tag: "update", updates: Set[CONSTuple] }




//===== GOSSIP Internals
    var localViewAdds : Proc -> Set[CONSTuple]                        //Witness variable of messages broadcast by CONS
    var localViewDels : Proc -> Set[CONSTuple]                        //Witness variable of messages broadcast by CONS

    var sMsgs: Proc -> Set[{dst:Proc, msg: GOSSIPMessage}]            //Unicast sent messages
    var rMsgs: Proc -> Set[{src:Proc, msg: GOSSIPMessage}]            //Unicast received messages

    //Helper
    val gossipInternals = {adds: localViewAdds, dels: localViewDels, sent: sMsgs, received: rMsgs}

    action unchangedGossip() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
        rMsgs' = rMsgs,
        sMsgs' = sMsgs
    }

    action unchangedGossipMsgs() = all {
        rMsgs' = rMsgs,
        sMsgs' = sMsgs
    }

    action unchangedGossipView() = all {
        localViewAdds' = localViewAdds,
        localViewDels' = localViewDels,
    }

// This is a "cheating" version of gossip
// It describes the wanted behavior, which is to simply merged local views.
    action gossipCheat(p,q) = {
        val lvPA = localViewAdds.get(p)
        val lvPD = localViewDels.get(p)
        val lvQA = localViewAdds.get(q)
        val lvQD = localViewDels.get(q)

        all {
            localViewAdds' = localViewAdds.set(p, union(lvPA,lvQA))
                                        .set(q, union(lvPA,lvQA)),
            localViewDels' = localViewDels.set(p, union(lvPD,lvQD))
                                        .set(q, union(lvPD,lvQD)),

            //unchanged
            unchangedGossipMsgs
        }
    }

// This is a version of gossip that uses messages from the P2P layer.
    action gossip(p,q) = all {
        //send query message
        //send query response
        //receive query response
        //identify differences
        //send differences
        //receive differences
        //merge

        //TODO: temporary unchanged
        unchangedGossip,
    }

//===== P2P Internals
    var maxCon : Proc -> int                                    //Maximum number of connections
    var igProcs : Proc -> Set[Proc]                                 //Nodes to be ignored.
    var cnProcs  : Proc -> Set[Proc]                                 //Neighbor sets

    //Helper
    val p2pInternals = {maxCon: maxCon, ignore: igProcs, connections: cnProcs}

    action addPeer(p,q): (Proc,Proc) => bool = all {
        not(cnProcs.get(p).contains(q)),
        not(cnProcs.get(q).contains(p)),
        size(cnProcs.get(p)) < maxCon.get(p),
        size(cnProcs.get(q)) < maxCon.get(q),
        cnProcs' = plusE2(plusE2(cnProcs, p, q), q, p)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        cnProcs' = minusE2(cnProcs, p, q)
    }

    action unchangedP2P = all {
        maxCon' = maxCon,
        igProcs' = igProcs,
        cnProcs' = cnProcs
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



//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
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

    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        sMsgs' = plusE2(sMsgs, p, {dst:q, msg:m})
    }

    def received(p) = rMsgs.get(p)

//======== other definitions
    def globalView: Set[CONSTuple] = {
        val allAdds = localViewAdds.keys().fold(Set(), (acc, k) => acc.union(localViewAdds.get(k)))
        val allDels = localViewDels.keys().fold(Set(), (acc, k) => acc.union(localViewDels.get(k)))

        allAdds.exclude(allDels)
    }


    def supersededByNewHeight(p): Proc => Set[CONSTuple] = {
        //chose the highest height
        val highestHeight = localView(p).fold(0, (a, e) => if (e.height > a) e.height else a)
        //chose the entries with height smaller and return
        localView(p).filter(e => e.height < highestHeight)
    }


    // Removes superseded entries from the tuple space
    action forgetSuperseded(p): Proc => bool = {
        val ssdByHeight = supersededByNewHeight(p)
        val ssdByNewRound = Set() //TODO: superSededByNewRound(p)
        val ssdByNewProof = Set() //TODO: superSededByNewProof(p)

        all {
            removeTuples(p, ssdByHeight.union(ssdByNewRound).union(ssdByNewProof)),

            unchangedGossipMsgs,
        }
    }



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
        always (
            globalView.forall(t =>
                eventually (
                    any {
                        Procs.forall(p => localView(p).contains(t)),
                        not (globalView.contains(t))
                    }
                )
            )
        )
    }

//================================= REQ-GOSSIP-P2P
    //[REQ-GOSSIP-P2P-UNICAST]
    temporal RGPUnicastProp = {
        Procs.forall(p =>
            cnProcs.get(p).forall(q =>
                sMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                rMsgs.get(q).contains({src:p, msg:m.msg}),
                                not (cnProcs.get(p).contains(q))
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
                cnProcs.get(p).size() < maxCon.get(p)
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

    temporal RGPIgnoringProp: bool =
        always (RGPIgnoring)


//===== Init (anything ending with Init sets the whole state and can serve as the spec `init` clause)
    action initCons: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }

    action initGossip: bool = all {
        sMsgs' = Procs.mapBy(_ => Set()),
        rMsgs' = Procs.mapBy(_ => Set())
    }

    action initP2P: bool = all {
        maxCon' = Procs.mapBy(_ => 3),
        igProcs' = Procs.mapBy(_ => Set()),
        cnProcs' = Procs.mapBy(_ => Set())
    }


    action randomEntryInit: bool = all {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        localViewAdds' = Procs.mapBy(p => if (p == 1) Set(entry) else Set()),
        localViewDels' = Procs.mapBy(_ => Set()),

        initGossip,
        initP2P,
    }

    action init: bool = all {
        initCons,
        initGossip,
        initP2P,
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    action garbageCollect = all {
        nondet p = oneOf(Procs)
        forgetSuperseded(p),
        unchangedP2P,
    }


    //Gossip without using P2P
    action gossipStepCheat(p,q) = all {
        gossipCheat(p,q),
        unchangedP2P,
    }

    //Gossip using P2P
    action gossipStep(p,q) = all {
        gossip(p,q),
        unchangedP2P
    }

    //Gossip using P2P
    action forgetSupersededStep = all {
        nondet p = Procs.oneOf()

        forgetSuperseded(p),
        unchangedP2P
    }

    //Update the connections
    action randomDropConnectionStep() = all {
        Procs.filter(ap => cnProcs.get(ap).size() > 0).size() > 0,
        nondet p = Procs.filter(ap => cnProcs.get(ap).size() > 0).oneOf()
        nondet q = cnProcs.get(p).oneOf()

        removePeer(p,q),
        maxCon' = maxCon,
        igProcs' = igProcs,
        unchangedGossip,
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
                unchangedGossip,
            }
        }
    }

    action addEntryStep(p, e) = all {
            addTuple(p, e),

            //unchanged
            unchangedGossipMsgs,
            localViewDels' = localViewDels,
            unchangedP2P
    }


    action randomAddEntryStep() = all {
        nondet p = oneOf(Procs)
        nondet tag = oneOf(Set("proposal","prevote","precommit"))
        nondet height = 1.to(10).oneOf()
        nondet round = 1.to(10).oneOf()
        val entry = {tag:tag, height:height, round:round, payload:"Payload"}

        addEntryStep(p,entry),
    }

    action step: bool = any {
        all { //gossip
            //chose partners
            nondet p = oneOf(Procs)
            nondet q = oneOf(minusE(Procs,p))

            gossipStepCheat(p,q)
            //gossipStep(p,q)
        },
        randomConnectionStep,
        //randomDropConnectionStep,
        //randomAddEntryStep,
        //forgetSupersededStep,
    }












//================ Test and Debug

//==== Actions

    action fullConnectStep() = {
        all {
            cnProcs' = Procs.mapBy(p => Procs),
            maxCon' = maxCon,
            igProcs' = igProcs,
            unchangedGossip,
        }
    }

    action ringConnectStep() = {
        all {
            cnProcs' = Procs.mapBy(p => if (p < size(Procs)) Set(p+1) else Set(1)),
            maxCon' = maxCon,
            igProcs' = igProcs,
            unchangedGossip,
        }
    }

    action starConnectStep() = {
        all {
            cnProcs' = Procs.mapBy(p => if (p ==1 ) Procs.minusE(1) else Set(1)),
            maxCon' = maxCon,
            igProcs' = igProcs,
            unchangedGossip,
        }
    }

    action randomConnectStep() = {
        //TODO: Describe a random connected graph, without iteration, efficiently
        //nondet f = Procs.setOfMaps(powerset(Procs)).oneOf()
        all {
            cnProcs' = Procs.mapBy(p => if (p < size(Procs)-1) Set(p+1, p+2) else if (p == size(Procs)-1) Set(p+1,1) else Set(1,2)),
            maxCon' = maxCon,
            igProcs' = igProcs,
            unchangedGossip,
        }
    }



    action stutter = all {
        unchangedGossip,
        unchangedP2P
    }



//===== Runs
    run sendReceiveInARingTest: bool = {
        init
            //connect some nodes
            .then(ringConnectStep)
            //send and receive messages
            //TODO
    }

    run sendReceiveInFullyConnectedTest: bool = {
        init
            //connect some nodes
            .then(fullConnectStep)
            //send and receive messages
            //TODO
    }

    run gossipInARingTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        init
            //connect some nodes
            .then(ringConnectStep)
            //add one entry
            .then(addEntryStep(1, entry))
            //gossip
            .then(gossipStepCheat(1,2))
            .then(gossipStepCheat(2,3))
            .then(gossipStepCheat(3,4))
            .then(gossipStepCheat(4,5))
            .then(gossipStepCheat(1,2))
            .then(all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                stutter
            })
    }

   run gossipInFullTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}
        init
            //connect some nodes
            .then(fullConnectStep)
            //add one entry
            .then(addEntryStep(1, entry))
            //gossip
            .then(gossipStepCheat(1,2))
            .then(gossipStepCheat(1,3))
            .then(gossipStepCheat(3,4))
            .then(gossipStepCheat(3,5))
            .then(all {
                //All see the entry
                Procs.forall(p => localView(p).contains(entry)),
                stutter
            })
    }

    run gossipStarTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        init
            //connect some nodes
            .then(starConnectStep)
            //add one entries
            .then(addEntryStep(1, entry))
            //gossip
            .then(step)
            .repeated(5)
            .then(all {
                //Some see the entry
                Procs.minusE(1).exists(p => localView(p).contains(entry)),
                stutter
            })
    }

    run gossipRandomTest: bool = {
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL"}

        init
            //connect some nodes
            .then(randomConnectStep)
            //add one entries
            .then(addEntryStep(1, entry))
            //gossip
            .then(step)
            .repeated(5)
            .then(all {
                //Some see the entry
                Procs.minusE(1).exists(p => localView(p).contains(entry)),
                stutter
            })
    }

}

/*
To test this spec

- quint test reactor.qnt //Runs all tests
- quint run --max-samples 10000 --max-steps 10 --invariant RGPIgnoring reactor.qnt
*/
