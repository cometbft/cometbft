module GossipAbstract {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"

    import GossipI.* from "./GossipI"

    //XOR import
        import P2PInstCheat as P2PInst from "./P2PInstCheat"
        export P2PInstCheat as P2PInst
        //import P2PInstNetwork as P2PInst from "./P2PInstNetwork"
        //export P2PInstNetwork as P2PInst

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