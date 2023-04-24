// Pure Helper functions
module Helpers {
    pure def excludeE(s, e): (Set[a],a) => Set[a] = s.exclude(Set(e))
    pure def unionE(s, e): (Set[a],a) => Set[a] = s.union(Set(e))
    pure def print(msg, toPrint) = true
}

// Global parameters.
// Must be imported as `Globals.*`
module Globals {
    type ProcId = int

    //XXX Quint: Ideally this should be a const, but because it would have to be
    //instantiated on every import, there would be duplication.
    //All processes in the system
    pure val ProcIds = Set(1,2,3,4,5)

    pure val DEFAULT_MAX_INBOUND_CONNECTION_NUMBER = 4
    pure val DEFAULT_MAX_OUTBOUND_CONNECTION_NUMBER = 4

}


// The P2P Interface
// Must be imported as `P2PI.*`
// Commented actions are implemented by P2PAbstract or P2PInst*
module P2PI {
 //====== [VOC-GOSSIP-P2P-TYPES] P2P-I
    type P2PMessage =
        | {tag: "query_request" }
        | {tag: "query_response", tuples: Set[{tag:str, height:int, round:int,              signer:int}] }
        | {tag: "update_request", tuples: Set[{tag:str, height:int, round:int, payload:str, signer:int}]}

//==== Exported methods P2P-I
    //action send(p,q,m): (ProcId, ProcId, P2PMessage) => bool
    //def received(p): Set[P2PMessage]
    //action ignoreNode(p:ProcId, q:ProcId, flag:bool): bool

}

//Provides standard behavior for P2PI implementations
module P2PAbstract {
    import Helpers.*
    import Globals.*
    import P2PI.*

//==== Common to implementationtions, but not actually part of the P2P-I.
    //===== P2P Internals                                       //TODO: Move some of these to P2PI

    type PeerNodeState = {
        ignoreNodes: Set[ProcId],                               //Nodes to be ignored.
        connectedNodes: Set[ProcId],                            //Neighbor sets
        maxInbCon: int,                                         //Maximum number of inbound connections
        maxOutCon: int,                                         //Maximum number of outbound connections
    }

    type NetworkState = {
        msgsSent: Set[{src:ProcId, dst:ProcId, msg: P2PMessage}],         //Unicast sent messages.
        msgsRcvd: Set[{src:ProcId, dst:ProcId, msg: P2PMessage}],         //Unicast received messages
    }

    var pid2peerState: ProcId -> PeerNodeState
    var networkState: NetworkState

    val stateAll = {pid2peerState: pid2peerState, Network:networkState}

    action unchangedAll = all {
        pid2peerState' = pid2peerState,
        networkState' = networkState
    }

    //Tare p and q connected at the P2P level?
    def areConnected(p: ProcId, q: ProcId): bool =
        pid2peerState.get(p).connectedNodes.contains(q)

    //Adds bidirectional connection if both nodes are below limits.
    action connectPeers(p: ProcId, q: ProcId): bool = {
        val pState = pid2peerState.get(p)
        val qState = pid2peerState.get(q)
        all {
            not(areConnected(p,q)),
            size(pState.connectedNodes) < pState.maxOutCon,
            size(qState.connectedNodes) < pState.maxInbCon,

            pid2peerState' = pid2peerState.set(p, pState.with("connectedNodes", pState.connectedNodes.union(Set(q))))
                                .set(q, qState.with("connectedNodes", qState.connectedNodes.union(Set(p))))
        }
    }

    action disconnectPeers(p:ProcId, q:ProcId): bool = {
        val pState = pid2peerState.get(p)
        val qState = pid2peerState.get(q)
        pid2peerState' = pid2peerState.set(p, pState.with("connectedNodes", pState.connectedNodes.exclude(Set(q))))
                            .set(q, qState.with("connectedNodes", qState.connectedNodes.exclude(Set(p))))
    }


    action ignoreNode(p:ProcId, q:ProcId, flag:bool): bool = {
        val pState = pid2peerState.get(p)
        if (flag) {
            pid2peerState' = pid2peerState.set(p, pState.with("ignoreNodes", pState.ignoreNodes.union(Set(q))))
        } else {
            pid2peerState' = pid2peerState.set(p, pState.with("ignoreNodes", pState.ignoreNodes.exclude(Set(q))))
        }
    }

//init and step must exist to be called from extensions/instantiators.
//==== Init
    action init: bool = all {
        pid2peerState' = ProcIds.mapBy(p => {
                    ignoreNodes: Set(),
                    connectedNodes: Set(),
                    maxInbCon: DEFAULT_MAX_INBOUND_CONNECTION_NUMBER,
                    maxOutCon: DEFAULT_MAX_OUTBOUND_CONNECTION_NUMBER
        }),
        networkState' = {msgsSent: Set(), msgsRcvd: Set()}
    }


//===== Invariants
//================================= REQ-GOSSIP-P2P
    //[REQ-GOSSIP-P2P-UNICAST]
    //XXX Quint: not supported.
    temporal RGPUnicastProp = {
        always (networkState.msgsRcvd.forall(m =>
            eventually (or {
                networkState.msgsRcvd.contains(m),
                not (areConnected(m.src, m.dst))
            })
        ))
    }

    //[REQ-GOSSIP-P2P-CONCURRENT_CONN]
    def RGPConcurrent = {
        ProcIds.forall(p =>
            val pState = pid2peerState.get(p)
            size(pState.connectedNodes) < pState.maxInbCon + pState.maxOutCon
        )
    }

    temporal RGPConcurrentProp = always (RGPConcurrent)

    //[REQ-GOSSIP-P2P-IGNORING]
    def RGPIgnoring: bool = {
        ProcIds.forall(p =>
            val pState = pid2peerState.get(p)
            val cnP = pState.connectedNodes
            val igP = pState.ignoreNodes
            val igCn = cnP.intersect(igP)
                size(igCn) == 0
        )
    }

    temporal RGPIgnoringProp: bool = always (RGPIgnoring)


    //======== Tests
    //Connects nodes as a ring.
    action connectRingStep() = all {
        //Find a node not in the ring.
        val connectionMap = ProcIds.mapBy(p => if (p == 1) Set(ProcIds.size(), 2)
                                                else if (p == ProcIds.size()) Set(1, p-1)
                                                else Set(p-1, p+1))
        pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        networkState' = networkState
    }

    //Connects nodes completely.
    action connectFullStep() = all {
        val connectionMap = ProcIds.mapBy(p => ProcIds.excludeE(p))
        pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        networkState' = networkState
    }

    //Connects nodes as a star around node 1.
    action connectStarStep() = all {
        val connectionMap = ProcIds.mapBy(p => if (p == 1 ) ProcIds.excludeE(1) else Set(1))
        pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        networkState' = networkState
    }

    action connectRandomStep() = all {
        //TODO: Describe a random connected graph, without iteration, efficiently
        nondet connectionMapWithLoops = setOfMaps(ProcIds, powerset(ProcIds)).oneOf()
        val connectionMap = connectionMapWithLoops.keys().mapBy(p => connectionMapWithLoops.get(p).excludeE(p))
        pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        networkState' = networkState
    }

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

    import P2PAbstract as PA
    export P2PAbstract as PA


//====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p: ProcId, q:ProcId, m: P2PMessage): bool = all {
        PA::areConnected(p,q),
        val msg = {src:p, dst:q, msg:m}
        all {
            not(msg.in(PA::networkState.msgsRcvd)),
            PA::networkState' = PA::networkState.with("msgsSent", PA::networkState.msgsSent.union(Set(msg)))
                                                .with("msgsRcvd", PA::networkState.msgsRcvd.union(Set(msg)))
        }
    }

    def received(p): ProcId => Set[{src:ProcId, dst:ProcId, msg: P2PMessage}] =
        PA::networkState.msgsRcvd.filter(m => m.dst == p)


//===== Steps
    action receiveMessageStep = all {
        PA::unchangedAll //XXX Quint: this should be just unchangedAll
    }
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

    import P2PAbstract as PA
    export P2PAbstract as PA

    type OutGoingMsg = {dst:ProcId, msg: P2PMessage}
    type IncomingMsg = {src:ProcId, msg: P2PMessage}

    def nonDelMsgs(p): ProcId => Set[{src: ProcId, dst: ProcId, msg:P2PMessage}] = {
        PA::networkState.msgsSent.filter(m => m.src == p and not(m.in(PA::networkState.msgsRcvd)))
    }

    //====== [VOC-GOSSIP-P2P-ACTIONS] P2P-I
    //[DEF-UNICAST]
    //The P2P unicast API.
    action send(p,q,m): (ProcId, ProcId, P2PMessage) => bool = all {
        val pState = PA::pid2peerState.get(p)
        pState.connectedNodes.contains(q),

        val msg = {src:p, dst:q, msg:m}
        all {
            print("sending", msg),
            not(PA::networkState.msgsSent.contains(msg)),

            PA::networkState' = PA::networkState.with("msgsSent", PA::networkState.msgsSent.union(Set(msg)))
        },
    }

    def received(p) = PA::networkState.msgsRcvd.filter(m => m.dst == p)


//===== Step
    action receiveMessageStep = {
        val nonDelMsgsAll = PA::networkState.msgsSent.filter(m => not(m.in(PA::networkState.msgsRcvd))
                                                        and m.dst.in(PA::pid2peerState.get(m.src).connectedNodes))
        all {
            size(nonDelMsgsAll) > 0,
            nondet msg = nonDelMsgsAll.oneOf()
            all {
                print("receiving", msg),
                PA::networkState' = PA::networkState.with("msgsRcvd", PA::networkState.msgsRcvd.union(Set(msg))),
                PA::pid2peerState' = PA::pid2peerState
            }
        }
    }

    action step = any {
        receiveMessageStep,
    }

}


//quint test gossip_ref.qnt --main P2PTests --verbosity=5
module P2PTests {
    import Helpers.*
    import Globals.*

    //XOR import
        import P2PInstCheat as P2PInst
        //import P2PInstNetwork as P2PInst

    action unchangedAll = all {
        P2PInst::PA::unchangedAll,
    }

    action init = all {
        P2PInst::PA::init,
    }

    //===== Runs/Debug
    run sendReceiveInRingTopTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectRingStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(2))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.fail())
    }

    run sendReceiveInFullyConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectFullStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(3))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
    }

    run sendReceiveInStarConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectStarStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,5,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(3))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.fail()
            )
    }

    run sendReceiveInRandomConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectRandomStep)
            //send and receive messages
            .then(
                all {
                    nondet p = P2PInst::PA::pid2peerState.keys().filter(pa => P2PInst::PA::pid2peerState.get(pa).connectedNodes.size() > 0).oneOf()
                    nondet q = P2PInst::PA::pid2peerState.get(p).connectedNodes.oneOf()
                    P2PInst::send(p,q,{tag: "query_request"}),
                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.repeated(1)
            )
            .then(P2PInst::receiveMessageStep.repeated(1))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
    }
}

//Define the Gossip Interface
module GossipI {
    import Globals.*

//====== [VOC-CONS-GOSSIP-TYPES] GOSSIP-I
    type GossipMessage =
        | { tag: "proposal",  height: int, round: int, payload: str, signer: ProcId}
        | { tag: "prevote",   height: int, round: int, payload: str, signer: ProcId}
        | { tag: "precommit", height: int, round: int, payload: str, signer: ProcId}

//==== Exported methods GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    //action addTuple(p,t): (ProcId, GossipMessage) => bool

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    //action removeTuples(p, es): (ProcId, Set[GossipMessage]) => bool

    //[DEF-READ-TUPLE]
    //The local view
    //def localView(p): ProcId => Set[GossipMessage]
}

module GossipAbstract {
    import Helpers.*
    import Globals.*

    import GossipI.*

    //XOR import
        //import P2PInstCheat as P2PInst
        //export P2PInstCheat as P2PInst
        import P2PInstNetwork as P2PInst
        export P2PInstNetwork as P2PInst

    //===== GOSSIP Internals
    type GOSSIPNodeState = {
        adds: Set[GossipMessage],
        dels: Set[GossipMessage]
    }

    var GossipState: ProcId -> GOSSIPNodeState

    val stateAll = {GossipState: GossipState}

    action unchangedAll() = all {
        GossipState' = GossipState,
    }


    //Helper
    def globalView: Set[GossipMessage] = {
        val allAdds = GossipState.keys().fold(Set(), (acc, k) => acc.union(GossipState.get(k).adds))
        val allDels = GossipState.keys().fold(Set(), (acc, k) => acc.union(GossipState.get(k).dels))

        allAdds.exclude(allDels)
    }

    //====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    //action addTuple(p,t): (ProcId, GossipMessage) => bool

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    //action removeTuples(p, ts): (ProcId, Set[GossipMessage]) => bool

    //[DEF-READ-TUPLE]
    //The local view
    def localView(p) =
        val pA = GossipState.get(p).adds
        val pD = GossipState.get(p).dels
        pA.exclude(pD)


    action init: bool = all {
        GossipState' = ProcIds.mapBy(_ => {adds: Set(), dels: Set()}),
    }

    def areConnected(p: ProcId, q: ProcId): bool =
        P2PInst::PA::areConnected(p,q)

}


// This module implements GOSSIP-I without really gossiping. All tuples are immediately available to all.
module GossipInstCheat {
    import Helpers.*
    import Globals.*

    import GossipI.*

    import GossipAbstract as GA
    export GossipAbstract as GA

    val stateAll = {nada: "nada"}

    action unchangedAll = true
    action init = true

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p: ProcId, t: GossipMessage): bool = {
        val pA = GA::GossipState.get(p).adds
        val pAN = pA.unionE(t)
        val pD = GA::GossipState.get(p).dels
        all {
            not(pD.contains(t)),
            GA::GossipState' = ProcIds.mapBy(_ => {adds: pAN, dels: pD})
        }
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p: ProcId, ts: Set[GossipMessage]): bool = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        val pDN = pD.union(ts)
        all {
            ts.forall(t => pA.contains(t)),
            GA::GossipState' = ProcIds.mapBy(_ => {adds: pA, dels: pDN})

        }
    }

//===== Steps (anything ending in Step sets the whole state and may be used in `run`)
    //Step does not need to do anything, since everything is imediatelly gossiped.
    //Gossip using P2P
    action gossip(p,q) = GA::unchangedAll
}


// This module implements GOSSIP-I without using P2P. Nodes compare their local views and converge 2-by-2.
module GossipInstNoP2P {
    import Helpers.*
    import Globals.*

    import GossipI.*
    import GossipAbstract as GA
    export GossipAbstract as GA

    val stateAll = {nada: "nada"}

    action unchangedAll = true
    action init = true

    // This is a "cheating" version of gossip
    // It describes the abstract wanted behavior, which is to simply merge local views.
    action gossip(p,q) = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        val qA = GA::GossipState.get(q).adds
        val qD = GA::GossipState.get(q).dels
        all {
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", union(pA,qA))
                                                                            .with("dels", union(pD,qD)))
                                              .set(q, GA::GossipState.get(q).with("adds", union(pA,qA))
                                                                            .with("dels", union(pD,qD)))
        }
    }

//====== [VOC-CONS-GOSSIP-ACTIONS] GOSSIP-I
    //[DEF-ADD-TUPLE]
    //Adds a tuple to the local view.
    action addTuple(p,t): (ProcId, GossipMessage) => bool = {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        all {
            not(pD.contains(t)),
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", pA.unionE(t)))
        }
    }

    //[DEF-REMOVE-TUPLE]
    //REMOVES a set of tuples from the local view.
    action removeTuples(p, ts): (ProcId, Set[GossipMessage]) => bool = all {
        val pA = GA::GossipState.get(p).adds
        val pD = GA::GossipState.get(p).dels
        all {
            ts.forall(t => pA.contains(t)),
            GA::GossipState' = GA::GossipState.set(p, GA::GossipState.get(p).with("adds", exclude(pA,ts))
                                                                            .with("dels", union(pD,ts)))
        }
    }
}




//This module implements GOSSIP-I using P2P-I. Nodes converge by exchanging messages.
module GossipInstP2P {
    import Helpers.*
    import Globals.*

    import P2PI.*

    import GossipI.*
    import GossipAbstract as GA
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





//==== Gossip Tests

module GossipTests {
    import Helpers.*
    import Globals.*

    //XOR import
        import GossipInstCheat as GInst
        //import GossipInstNoP2P as GInst
        //import GossipInstP2P as GInst

    var stepCounter: int
    val stateAll = {
        stepCounter:stepCounter,
        pa:GInst::GA::P2PInst::PA::stateAll,
        pa:GInst::GA::stateAll
    }

    action init = all {
        stepCounter' = 1
    }

    action unchangedAll = {stepCounter' = stepCounter}
    action increment = {stepCounter' = stepCounter + 1}

    def gossipablePairs =
        tuples(ProcIds,ProcIds).filter(t => and (
            t._1 != t._2,
            GInst::GA::localView(t._1) != GInst::GA::localView(t._2),
            GInst::GA::areConnected(t._1, t._2)
    ))

//===== Runs/Debug
    run gossipInRingTopoTest: bool = {
        val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectRingStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    unchangedAll
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            unchangedAll
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossip(pq._1, pq._2),
                        GInst::GA::P2PInst::PA::unchangedAll,
                        GInst::unchangedAll,

                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                unchangedAll,
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            unchangedAll,
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,

        })
    }

    run gossipCliqueTopoTest: bool = {
        val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
        all {
            GInst::GA::P2PInst::PA::connectFullStep,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            unchangedAll
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            unchangedAll
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossip(pq._1, pq._2),
                        GInst::GA::P2PInst::PA::unchangedAll,
                        GInst::unchangedAll,

                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                unchangedAll,
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            unchangedAll,
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,

        })
    }

    run gossipInStarTopoTest: bool = {
            val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectStarStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    unchangedAll
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            unchangedAll
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossip(pq._1, pq._2),
                        GInst::GA::P2PInst::PA::unchangedAll,
                        GInst::unchangedAll,

                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                unchangedAll,
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            unchangedAll,
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,

        })
    }

        run gossipInRandomTopoTest: bool = {
            val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectRandomStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    unchangedAll
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            unchangedAll
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossip(pq._1, pq._2),
                        GInst::GA::P2PInst::PA::unchangedAll,
                        GInst::unchangedAll,

                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                unchangedAll,
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            unchangedAll,
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,

        })
    }
}