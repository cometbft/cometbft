module ConsensusReactor {

    type Proc = int

    //const Procs: Set[Proc]                                    //All processes in the system //TODO: Use once fully supported.
    pure val Procs = Set(1,2,3,4,5)

    type Payload = str

//===== Helper functions
    pure def excludeElem(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def excludeElem2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).exclude(Set(e)))

    pure def excludeElems(s, es): (Set[a], Set[a]) => Set[a] = s.exclude(es)
    pure def excludeElems2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).exclude(es))

    pure def includeElem(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def includeElem2(m, p, e): (i -> Set[a], i, a) => i -> Set[a] = m.set(p, m.get(p).union(Set(e)))

    pure def includeElems(s, es): (Set[a],Set[a]) => Set[a] = s.union(es)
    pure def includeElems2(m, p, es): (i -> Set[a], i, Set[a]) => i -> Set[a] = m.set(p, m.get(p).union(es))

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

            //unchanged
            unchangedGossipMsgs
        }
    }

// This is a version of gossip that uses messages from the P2P layer.
    action gossipReal(p,q) = all {
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

// Removes superseded entries from the tuple space
    action forgetUseless(p) = all {
        unchangedGossipMsgs,
        //TODO: temporary unchanged
        unchangedGossipView
        //TODO: clean the state of this process
    }

//===== P2P Internals
    var maxConn : Proc -> int                                    //Maximum number of connections
    var igProcs : Proc -> Set[Proc]                                 //Nodes to be ignored.
    var cnProcs  : Proc -> Set[Proc]                                 //Neighbor sets

    //Helper
    val consInternals = {maxCon: maxConn, ignore: igProcs, connections: cnProcs}

    action addPeer(p,q): (Proc,Proc) => bool = all {
        not(cnProcs.get(p).contains(q)),
        not(cnProcs.get(q).contains(p)),
        size(cnProcs.get(p)) < maxConn.get(p),
        size(cnProcs.get(q)) < maxConn.get(q),
        cnProcs' = includeElem2(includeElem2(cnProcs, p, q), q, p)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        cnProcs' = excludeElem2(cnProcs, p, q)
    }

    action unchangedP2P = all {
        maxConn' = maxConn,
        igProcs' = igProcs,
        cnProcs' = cnProcs
    }


//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I

    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (Proc, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = includeElem2(localViewAdds, p, t),
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (Proc, Set[CONSTuple]) => bool = all {
        localViewDels' = includeElems2(localViewDels, p, es),
        localViewAdds' = excludeElems2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = localViewAdds.get(p).exclude(localViewDels.get(p))

//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    action setMaxCon(p, m): (int,int) => bool = {
        maxConn' = maxConn.set(p, m)
    }

    def maxCon(p) = {
        maxConn.get(p)
    }

    action ignoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = includeElem2(igProcs, p, q)
    }

    action doNotIgnoreNode(p, q): (Proc, Proc) => bool = {
        igProcs' = excludeElem2(igProcs, p, q)
    }

    def neighbors(p) = cnProcs.get(p)

    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(cnProcs.get(p)),
        sMsgs' = includeElem2(sMsgs, p, {dst:q, msg:m})
    }

    def received(p) = rMsgs.get(p)

//======== Shared definitions
    def globalView: Set[CONSTuple] = {
        val allAdds = localViewAdds.keys().fold(Set(), (acc, k) => acc.union(localViewAdds.get(k)))
        val allDels = localViewDels.keys().fold(Set(), (acc, k) => acc.union(localViewDels.get(k)))

        allAdds.exclude(allDels)
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
                cnProcs.get(p).size() < maxConn.get(p)
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






//===== Init
    action consInit: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }

    action gossipInit: bool = all {
        sMsgs' = Procs.mapBy(_ => Set()),
        rMsgs' = Procs.mapBy(_ => Set())
    }

    action p2pInit: bool = all {
        maxConn' = Procs.mapBy(_ => 3),
        igProcs' = Procs.mapBy(_ => Set()),
        cnProcs' = Procs.mapBy(_ => Set())
    }

    action init: bool = all {
        consInit,
        gossipInit,
        p2pInit,
    }

//===== Steps
    action step: bool = any {
        all {
            //chose partners
            nondet p = oneOf(Procs)
            nondet q = oneOf(excludeElem(Procs,p))
            gossip(p,q),
            unchangedP2P,
        },
        all {
            nondet p = oneOf(Procs)
            forgetUseless(p),
            unchangedP2P,
        }
    }

    action stepReal: bool = any {
        all {
            //chose partners
            nondet p = oneOf(Procs)
            nondet q = oneOf(excludeElem(Procs,p))
            gossipReal(p,q),
            unchangedP2P
        },
        all {
            nondet p = oneOf(Procs)
            forgetUseless(p),
            unchangedP2P
        }
    }



//===== Runs

    action addRandomEntry() = all {
            nondet p = oneOf(Procs)
            nondet tag = oneOf(Set("proposal","prevote","precommit"))
            nondet height = 1.to(10).oneOf()
            nondet round = 1.to(10).oneOf()
            val payload = "TX" //TODO+ height + " " + round

            all {
                addTuple(p, {tag: tag, height:height, round:round, payload: payload, signer: p}),

                //unchanged
                unchangedGossipMsgs,
                localViewDels' = localViewDels,
                unchangedP2P
            }
    }

    action addRandomConnection() = all {
        nondet p = oneOf(Procs)
        nondet q = oneOf(excludeElem(Procs,p))

        all {
            addPeer(p,q),
            maxConn' = maxConn,
            igProcs' = igProcs,
            unchangedGossip,
        }
    }

    run sendReceiveRun: bool = {
        init
            //connect some nodes
            .then(addRandomConnection)
            .then(addRandomConnection)
            .then(addRandomConnection)
            //send and receive messages
    }

    run gossipRun: bool = {
        init
            .then(addRandomConnection)
            .then(addRandomConnection)
            .then(addRandomConnection)
            //add some entries
            .then(addRandomEntry)
            .then(addRandomEntry)
            .then(addRandomEntry)
            //gossip
            .then(step)
            .then(step)
            .then(step)
    }

    run gossipRealRun: bool = {
        init
            //connect some nodes
            .then(addRandomConnection)
            .then(addRandomConnection)
            .then(addRandomConnection)
            //add some entries
            .then(addRandomEntry)
            .then(addRandomEntry)
            .then(addRandomEntry)
            //gossip
            .then(stepReal)
            .then(stepReal)
            .then(stepReal)
    }
}
