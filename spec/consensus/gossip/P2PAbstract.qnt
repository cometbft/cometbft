//Provides standard behavior for P2PI implementations
module P2PAbstract {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"
    import P2PI.*    from "./P2PI"

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
        //TODO: Describe a random connected graph, efficiently
        //nondet connectionMapWithLoops = setOfMaps(ProcIds, powerset(ProcIds)).oneOf()
        //val connectionMap = connectionMapWithLoops.keys().mapBy(p => connectionMapWithLoops.get(p).excludeE(p))
        //pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        val connectionMap = Map(1 -> Set(3,5),
                                2 -> Set(5),
                                3 -> Set(1,5,4),
                                4 -> Set(3),
                                5 -> Set(1,3,2))
        pid2peerState' = ProcIds.mapBy(p => pid2peerState.get(p).with("connectedNodes", connectionMap.get(p))),
        networkState' = networkState
    }
}