module ConsensusReactor {

    type Proc = int
    const Procs: Set[Proc]                                      //All processes in the system

    type Payload = str


//====== [CONS-GOSSIP-VOCABULARY] GOSSIP-I
    type CONSMessage =                                         //TODO: Define payloads?
        | { tag: "proposal", height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "prevote", height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "precommit", height: int, round: int, payload: Payload, signer: Proc}
    

//====== [GOSSIP-P2P-VOCABULARY] P2P-I
    var maxConn: Proc -> int

    type GOSSIPMessage = 
        | {tag: "cons", consMsg: CONSMessage}
        | {tag: "placeholder", payload: Payload}

    var uMsgs: Proc -> Set[{dst:Proc, msg: GOSSIPMessage}]
    var rMsgs: Proc -> Set[{src:Proc, msg: GOSSIPMessage}]
    var nes : Proc -> Set[Proc]                                 //Neighbor sets
    var igs : Proc -> Set[Proc]                                 //Nodes to be ignored.



//====== CONS Internals
    var cbMsgs: Proc -> List[CONSMessage]

//===== GOSSIP Internals
    var bMsgs : Proc -> Set[CONSMessage]                        //Witness variable of messages broadcast by CONS
    var dMsgs : Proc -> Set[CONSMessage]                        //Witness variable of messages delivered to CONS and not discarded


//======== Shared definitions
    //[DEF-BROADCAST]
    //The broadcast API
    action broadcast(p,m): (Proc, CONSMessage) => bool = all {
        bMsgs' = bMsgs.set(p, bMsgs.get(p).union(Set(m)))
    }

    //[DEF-DELIVERY]
    //Messages delivered but not yet superseded
    def delivered(p) = dMsgs.get(p)

    //[DEF-SUPERSESSION]
    //The supersession operator  
    def SSS(lhs,rhs): (CONSMessage, CONSMessage) => bool = {
        //TODO: Most messages are deprecated by the another message of the same type 
        //but from a new round on the same height. Votes must be kept until a new height
        //is started.
        true
    }

    //m is SuperSedeD in the set s
    def isSSDInSet(s,m): (Set[CONSMessage], CONSMessage) => bool = {
        s.exists(m1 => m1.SSS(m))
    }

    def setRange(i,j) = range(i,j).indices()

    //m is SuperSedeD in the set s
    def isSSDInList(l,m): (List[CONSMessage], CONSMessage) => bool = {
        setRange(0,l.length()).exists(i => l.nth(i).SSS(m))
    }

    //[DEF-UNICAST]
    //The P2P unicast API.
    action unicast(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(nes.get(p)),
        uMsgs' = uMsgs.set(p, uMsgs.get(p).union(Set({dst:q, msg:m})))
    }


//================================= REQ-CONS-GOSSIP

    //[REQ-CONS-GOSSIP-BROADCAST.1]
    //Satisfied by [DEF-BROADCAST]

    //[REQ-CONS-GOSSIP-DELIVERY.1]
    //Satisfied by [DEF-DELIVERY]

    //[REQ-CONS-GOSSIP-BROADCAST.2]
    temporal RCGBroadcast2a = {
        Procs.forall(p =>
            bMsgs.get(p).forall(m1 =>
                not( 
                    Procs.exists(q =>
                        bMsgs.get(q).exists(m2 => 
                            SSS(m2,m1)
                        )
                    )
                ) implies eventually (
                    Procs.forall(q => 
                        m1.in(dMsgs.get(q))
                    )
                )
            )
        )
    }

    temporal RCGBroadcast2b = {
        Procs.forall(p =>
            dMsgs.get(p).exists(m1 =>
                not( 
                    Procs.exists(q =>
                        bMsgs.get(q).exists(m2 => 
                            SSS(m2,m1)
                        )
                    )
                ) implies eventually (
                    Procs.forall(q => 
                        m1.in(dMsgs.get(q))
                    )
                )
            )
        )
    }

/*
    temporal RCGBroadcast2 = all {
        RCGBroadcast2a,
        RCGBroadcast2b,
    }
*/

    //[REQ-CONS-GOSSIP-DELIVERY.2]
    //TODO: Am I missing an implication here?

    temporal RCGDelivery2 = {
        Procs.forall(p => 
            dMsgs.get(p).forall(m => 
                always (
                    or (
                        m.in(dMsgs.get(p)),
                        isSSDInSet(dMsgs.get(p),m)
                    )
                )
            )
        )
    }

//================================= PROV-CONS-GOSSIP

    //[PROV-CONS-GOSSIP-SUPERSESSION.1]
    //Provided by [DEF-SUPERSESSION]

    //[PROV-CONS-GOSSIP-SUPERSESSION.2]
    //TODO: []PCGSupersession2
    def PCGSupersession2(c) = (
        Procs.forall(p => 
            setRange(0, length(cbMsgs.get(p)) - 1).filter(i => 
                not (
                    isSSDInList(cbMsgs.get(p).slice(i,length(cbMsgs.get(p))), cbMsgs.get(p).nth(i))
                )
            ).size() < c
        )
    )

    //[PROV-CONS-GOSSIP-SUPERSESSION.3]
    //TODO: []PCGSupersession3
    def PCGSupersession3 = (
        Procs.forall(p => 
            setRange(1,length(cbMsgs.get(p))).forall(i =>
                not(isSSDInList(cbMsgs.get(p).slice(0,i), cbMsgs.get(p).nth(i)))
            )
        )
    )


//================================= REQ-GOSSIP-CONS
    //[REQ-GOSSIP-CONS-SUPERSESSION.1]
    //Satisfied by [DEF-SUPERSESSION]

    //[REQ-GOSSIP-CONS-SUPERSESSION.2]
    pure def nonSSDMsgs(msgs): (Set[CONSMessage]) => Set[CONSMessage] = 
        msgs.filter(m => not(isSSDInSet(msgs, m)))

    temporal RGCSupersession2 = {
        Int.exists(c => 
            Procs.forall(p =>
                always (
                    size(nonSSDMsgs(bMsgs.get(p))) < c
                )
            )
        )
    }

//================================= PROV-GOSSIP-CONS

    //[PROV-GOSSIP-CONS-BROADCAST.1]
    //Satisfied by [DEF-BROADCAST]


    //[PROV-GOSSIP-CONS-BROADCAST.2]
    //Satisfied by [LOCAL_CAST.1] + [GLOBAL_CAST.1]
    //TODO: show that [LOCAL_CAST.1] + [GLOBAL_CAST.1] + [REQ-GOSSIP-CONS-SUPERSESSION.2] => Best effort communication + Bounded memory usage.


    //[LOCAL_CAST.1]
    temporal localCast = {
        Procs.forall(p => 
            nes.get(p).forall(q => 
                bMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                m.in(dMsgs.get(q)),
                                isSSDInSet(dMsgs.get(p) ,m),
                                q.notin(nes.get(p))
                            }
                        )
                    )
                )
            )
        )
    }


    //[GLOBAL_CAST.1]
    temporal globalCast = {
        Procs.forall(p => 
            dMsgs.get(p).forall(m =>
                always (
                    eventually (
                        or {
                            m.in(bMsgs.get(p)),
                            isSSDInSet(bMsgs.get(p), m)
                        }
                    )
                )                
            )
        )
    }



//===== REQ-GOSSIP-P2P

    //[REQ-GOSSIP-P2P-UNICAST.1]
    // Provided by the [DEF-UNICAST]

    //[REQ-GOSSIP-P2P-UNICAST.2]
    temporal RGPUnicast2 = {
        Procs.forall(p => 
            nes.get(p).forall(q => 
                uMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                in({src:p, msg:m.msg}, rMsgs.get(q)),
                                q.notin(nes.get(p))
                            }
                        )
                    )
                )
            )
        )
    }

    //[REQ-GOSSIP-P2P-CONCURRENT_CONN]
    //Provided by construction of nes, addPeer and removePeer

    action addPeer(p,q): (Proc,Proc) => bool = all {
        q.notin(nes.get(p)),
        size(nes.get(p)) < maxConn.get(p),
        nes' = nes.set(p, (nes.get(p).union(Set(q)))),
//        size(nes'.get(p)) < maxConn.get(p)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        q.notin(nes.get(p)),
        nes' = nes.set(p, nes.get(p).exclude(Set(q)))
    }

    //[REQ-GOSSIP-P2P-IGNORING]
    //TODO: []RGPIgnoring
    temporal RGPIgnoring = {
        Procs.forall(p => 
            igs.get(p).forall(q => 
                q.notin(nes.get(p))
            )
        )
    }
}


