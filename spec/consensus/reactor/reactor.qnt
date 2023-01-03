module ConsensusReactor {

    type Proc = int
    const Procs : Set[Proc]                                      //All processes in the system
    
    // [CONS-GOSSIP-VOCABULARY] GOSSIP-I
    type StateMessage =                                         //TODO: Define payloads?
        | { tag: "proposal", height: int, round: int, payload: str, signer: Proc}
        | { tag: "prevote", height: int, round: int, payload: str, signer: Proc}
        | { tag: "precommit", height: int, round: int, payload: str, signer: Proc}
    


    var Ne : Procs -> Set[Procs]                                 //Neighbor sets
    var BMsgs : Procs -> Set[StateMessage]                       //Messages broadcast by State
    var DMsgs : Procs -> Set[StateMessage]                       //Messages delivered to State

    var UMsgs : Procs -> Set[StateMessage]                       //Messages unicast by State
    var Msgs : Procs -> Set[StateMessage]                        //Messages to be sent/forwarded (subset BMsgs[] U UMsgs[] U DMsgs[])


    // [GOSSIP-P2P-VOCABULARY] P2P-I







    //[DEF-SUPERSESSION]
    // The supersession operator  
    //[PROV-CONS-GOSSIP-SUPERSESSION]
    //[REQ-GOSSIP-CONS-SUPERSESSION.1]
    //Vocabulary x Requirement
    def SSS(lhs,rhs): (StateMessage, StateMessage) => bool = {
        //TODO
        true
    }

    //[DEF-BROADCAST]
    //[REQ-CONS-GOSSIP-BROADCAST]
    //Vocabulary x Requirement
    action broadcast(p,m): (Process, StateMessage) = {
        Msgs' = Msgs.set(p, Msgs[p].union(Set(m)))
    }

//================================= REQ-CONS-GOSSIP

    //[REQ-CONS-GOSSIP-KEEP_NON_SUPERSEDED]
    //TODO: Am I missing an implication here?
    temporal placeholder1 = {
        Procs.forall(p => 
            DMsgs.get(p).forall(m =>
                always (
                    or {
                        m.in(DMsgs.get(p)),
                        DMsgs.get(p).exists(m1 => m1.SSS(m))
                    }
                )
            )
        )
    }


//================================= REQ-GOSSIP-CONS

    //[REQ-GOSSIP-STATE-SUPERSESSION.2]
    pure def isSupersededIn(msg, msgs): (Message, Set[StateMessage]) => bool = {
        msgs.filter(m => m.SSS(msg)) != Set()
    }

    pure def nonSupersededMsgs(p): (Proc) => Set[StateMessage] = 
        Msgs.get(p).filter(m => not(isSupersededIn(m, Msgs[p])))

    temporal placeholder2 = {
        Int.exists(c => 
            Proc.forall(p => 
                always (
                    size(nonSupersededMsgs(p)) < c
                )
            )
        )
    }

    //[PROV-GOSSIP-STATE-NEIGHBOR_CAST.1]
    temporal placeholder3 = {
        Procs.forall(p => 
            Ne.get(p).forall(q => 
                BMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                m.in(DMsgs.get(q)),
                                BMsgs.get(p).exists(m1 => SSS(m1,m)),
                                q.notin(Ne.get(p))
                            }
                        )
                    )
                )
            )
        )
    }


    //[PROV-GOSSIP-STATE-NEIGHBOR_CAST.2]
    temporal placeholder4 = {
        Procs.forall(p => 
            DMsgs.get(p).forall(m =>
                always (
                    implies (
                        eventually (
                            or {
                                m.in(Msgs.get(p)),
                                Msgs.get(p).exists(m1 => SSS(m1,m)),
                            }
                        )
                    )
                )                
            )
        )
    }

    //[REQ-GOSSIP-P2P-UNICAST.1]
    action unicast(p,q,m): (Proc,Proc,Message) =
    all {
        Ne.get(p).contains(q),
        UMsgs' = UMsgs.set(p, UMsgs.get(p).union(Set((q,m)))
    }

    //[REQ-GOSSIP-P2P-UNICAST.2]
    temporal placeholder5 = {
        Procs.forall(p => 
            Ne.get(p).forall(q => 
                UMsgs.get(p).forall(m =>
                    always (
                        eventually (
                            or {
                                m.in(DMsgs.get(q)),
                                UMsgs.get(p).exists(m1 => SSS(m1,m)),
                                q.notin(Ne.get(p))
                            }
                        )
                    )
                )
            )
        )
    }



}



    //GOSSIP internals
    type GossipMessage = 
        | {tag: "state-msg", msg: StateMessage, sndr: Proc}
        | {tag: "other-msg", height: int, round: int, payload: str, sndr: Proc}

