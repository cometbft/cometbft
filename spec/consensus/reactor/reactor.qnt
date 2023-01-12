module ConsensusReactor {

    type Proc = int
    const Procs: Set[Proc]                                      //All processes in the system

    type Payload = str


//====== [CONS-GOSSIP-VOCABULARY] GOSSIP-I
    type CONSMessage =                                         //TODO: Define payloads?
        | { tag: "proposal", height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "prevote", height: int, round: int, payload: Payload, signer: Proc}
        | { tag: "precommit", height: int, round: int, payload: Payload, signer: Proc}
    
    var bMsgs : Proc -> Set[CONSMessage]                       //Messages broadcast by CONS
    var dMsgs : Proc -> Set[CONSMessage]                       //Messages delivered to CONS


//====== [GOSSIP-P2P-VOCABULARY] P2P-I
    var maxConn: Proc -> int

    type GOSSIPMessage = 
        | {tag: "cons", consMsg: CONSMessage}
        | {tag: "placeholder", payload: Payload}

    var uMsgs: Proc -> Set[{dst:Proc, msg: GOSSIPMessage}]
    var rMsgs: Proc -> Set[{src:Proc, msg: GOSSIPMessage}]
    var nes : Proc -> Set[Proc]                                 //Neighbor sets
    var igs : Proc -> Set[Proc]                                 //Nodes to be ignored.





//======== Shared definitions
    //[DEF-BROADCAST]
    //The broadcast API
    action broadcast(p,m): (Proc, CONSMessage) => bool = all {
        bMsgs' = bMsgs.set(p, bMsgs.get(p).union(Set(m)))
    }

    //[DEF-SUPERSESSION]
    //The supersession operator  
    def SSS(lhs,rhs): (CONSMessage, CONSMessage) => bool = {
        //TODO
        true
    }

    //m is SuperSedeD in the set s
    def isSSDInSet(s,m): (Set[CONSMessage], CONSMessage) => bool = {
        s.exists(m1 => m1.SSS(m))
    }

    //[DEF-UNICAST]
    //The P2P unicast API.
    action unicast(p,q,m): (Proc, Proc, GOSSIPMessage) => bool = all {
        q.in(nes.get(p)),
        uMsgs' = uMsgs.set(p, uMsgs.get(p).union(Set({dst:q, msg:m})))
    }



//====== CONS Internals
//===== GOSSIP Internals


//================================= REQ-CONS-GOSSIP

    //[REQ-CONS-GOSSIP-BROADCAST.1]
    //Satisfied by [DEF-BROADCAST] and use of bMsgs

    //[REQ-CONS-GOSSIP-DELIVERY.1]
    //Satisfied by usage of dMsgs

    //[REQ-CONS-GOSSIP-BROADCAST.2]
    //TODO


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
        size(nes'.get(p)) < maxConn.get(p)
    }

    action removePeer(p,q): (Proc,Proc) => bool = all {
        q.notin(nes.get(p)),
        nes' = nes.set(p, nes.get(p).exclude(Set(q)))
    }

    //[REQ-GOSSIP-P2P-IGNORING]
    temporal RGPIgnoring = always {
        Procs.forall(p => 
            igs.get(p).forall(q => 
                q.notin(nes.get(p))
            )
        )
    }

//GOSSIP internals
/*
    type GossipMessage = 
        | {tag: "state-msg", msg: CONSMessage, sndr: Proc}
        | {tag: "other-msg", height: int, round: int, payload: str, sndr: Proc}

*/

}







/*
    var height: Proc -> int
    var round: Proc -> int
    var step: Proc -> Steps
    var decision: List[Payload]

    type Steps =  
        | {tag: "proposal"}
        | {tag: "prevote"}
        | {tag: "precommit"}

    action InitGossipI = all {
        height' = Procs.mapBy(_ => 0),
        round' = Procs.mapBy(_ => 0),
        decision' = Procs.mapBy(_ => List())
    }
*/

//===== Anything below this line actually belongs to the implementation.qnt and will be moved
//===== there once I learn how modularize quint specs.


