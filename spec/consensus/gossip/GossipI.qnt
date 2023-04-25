//Define the Gossip Interface
module GossipI {
    import Globals.* from "./Globals"

//====== [VOC-CONS-GOSSIP-TYPES] GOSSIP-I
    type GossipMessage =
        | { tag: "proposal",  height: int, round: int, payload: str, signer: ProcId}
        | { tag: "prevote",   height: int, round: int, payload: str, signer: ProcId}
        | { tag: "precommit", height: int, round: int, payload: str, signer: ProcId}

//==== Exported methods GOSSIP-I
    //Commented actions are implemented by GossipAbstract or GossipInst*

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