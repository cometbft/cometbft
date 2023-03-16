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
        | {tag: "cons", consMsg: CONSTuple}                     //Wrapper for CONSTuple
        | {tag: "placeholder", payload: Payload}

//===== GOSSIP Internals
    var localViewAdds : Proc -> Set[CONSTuple]                        //Witness variable of messages broadcast by CONS
    var localViewDels : Proc -> Set[CONSTuple]                        //Witness variable of messages broadcast by CONS

    var sMsgs: Proc -> Set[{dst:Proc, msg: GOSSIPMessage}]            //Unicast sent messages
    var rMsgs: Proc -> Set[{src:Proc, msg: GOSSIPMessage}]            //Unicast received messages

//===== P2P Internals
    var maxConn : Proc -> int                                    //Maximum number of connections
    var igProcs : Proc -> Set[Proc]                                 //Nodes to be ignored.
    var cnProcs  : Proc -> Set[Proc]                                 //Neighbor sets

    action addPeer(p,q): (Proc,Proc) => bool = all {
        not(cnProcs.get(p).contains(q)),
        size(cnProcs.get(p)) < maxConn.get(p),
        cnProcs' = includeElem2(cnProcs, p, q)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        cnProcs' = excludeElem2(cnProcs, p, q)
    }

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I

    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (Proc, CONSTuple) => bool = all {
        not(localViewDels.get(p).contains(t)),
        localViewAdds' = includeElem2(localViewAdds, p, t)
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, es): (Proc, Set[CONSTuple]) => bool = all {
        localViewDels' = includeElems2(localViewDels, p, es),
        localViewAdds' = excludeElems2(localViewAdds, p, es)
    }

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) = localViewAdds.get(p)

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

/*
//================================= REQ-CONS-GOSSIP
    //[REQ-CONS-GOSSIP-CONVERGENCE]
    temporal RCGConvergence = {
        globalView().forall(t =>
            eventually any {
                Procs.forall(p => localView(p).contains(t)),
                not globalView.contains(t)
            }
        )
    }


//================================= REQ-GOSSIP-P2P
    //[REQ-GOSSIP-P2P-UNICAST]
    temporal RGPUnicast = {
        Procs.forall(p =>
            cnProcs.get(p).forall(q =>
                sMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                rMsgs.get(q).contains({src:p, msg:m.msg}),
                                not cnProcs.get(p).contains(q)
                            }
                        )
                    )
                )
            )
        )
    }

    //[REQ-GOSSIP-P2P-CONCURRENT_CONN]
    temporal RGPConcurrent = {
        Procs.forall(p =>
            always {
                cnProcs.get(p).size() < maxConn.get(p)
            }
        )
    }


    //[REQ-GOSSIP-P2P-IGNORING]
    temporal RGPIgnoring = {
        Procs.forall(p =>
            always {
                size(cnProcs.get(p).intersect(igs.get(p)) == 0)
            }
        )
    }

*/







//===== Init
    action consInit: bool = all {
        localViewAdds' = Procs.mapBy(_ => Set()),
        localViewDels' = Procs.mapBy(_ => Set())
    }


//===== P2P Internals


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


//===== Invariants


//===== Runs
    run simpleRun: bool = {
        init

        /*.then(
                nondet p = oneOf(Procs)
                addTuple(p, {tag: "proposal", height:1, round:0, payload: "TX1", signer: 1})
        )*/


        //TODO: then deliver
    }
}
