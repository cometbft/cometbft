// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module p2p {
    import spells.* from "./spells"

    //--------------------------------------------------------------------------
    // Types
    //--------------------------------------------------------------------------
    type NodeID = str
    
    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    const NodeIDs: Set[NodeID]
    const InitialPeers: NodeID -> Set[NodeID]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    var incomingMsgs: NodeID -> List[(NodeID, msg)]
    var peers: NodeID -> Set[NodeID]
    pure def multiSend(node, _incomingMsgs, targetNodes, msg) =
        _incomingMsgs.updateMultiple(targetNodes, ms => ms.append((node, msg)))
    pure def send(node, _incomingMsgs, targetNode, msg) =
        node.multiSend(_incomingMsgs, Set(targetNode), msg)
    
    // Auxiliary definitions
    def IncomingMsgs(node) = incomingMsgs.get(node)
    def Peers(node) = peers.get(node)
    val nodesInNetwork = NodeIDs.filter(node => node.Peers().nonEmpty())
    val nodesNotInNetwork = NodeIDs.exclude(nodesInNetwork)
    pure def disconnect(_peers, node) =
        // TODO: check that the network does not become disconnected; we don't want to model that.
        _peers.put(node, Set())
    def disconnectedPeers(node) = 
        node.Peers().filter(p => not(node.in(p.Peers())))

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    action P2P_init = all {
        incomingMsgs' = NodeIDs.mapBy(_ => List()),
        peers' = NodeIDs.mapBy(n => InitialPeers.get(n)),
    }
    action receiveFromPeer(node, handleMessage) = all {
        require(length(node.IncomingMsgs()) > 0),
        // We model receiving of a message as taking the head of the list of
        // incoming messages and leaving the tail.
        val someMsg = node.IncomingMsgs().head()
        val sender = someMsg._1
        val msg = someMsg._2
        val _incomingMsgs = incomingMsgs.update(node, tail)
        handleMessage(node, _incomingMsgs, sender, msg)
    }
    action joinNetwork(node, peerSet) = all {
        // The node must not be connected to the network.
        require(node.Peers().isEmpty()),
        peers' = peers
            // Assign to node the set of new peers.
            .put(node, peerSet)
            // Add node as a new peer to the set of connecting peers.
            .updateMultiple(peerSet, ps => ps.join(node)),
        incomingMsgs' = incomingMsgs,
    }
    action pickNodeAndJoin = all {
        // Pick a node that is not connected to the network.
        require(NodeIDs.exclude(nodesInNetwork).nonEmpty()),
        nondet node = oneOf(NodeIDs.exclude(nodesInNetwork))
        // Pick a non-empty set of nodes in the network to be the node's peers.
        nondet peerSet = oneOf(nodesInNetwork.powerset().exclude(Set()))
        node.joinNetwork(peerSet),
    }
    
    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    val bidirectionalNetwork =
        NodeIDs.forall(nodeA => 
            nodeA.Peers().forall(nodeB => nodeA.in(nodeB.Peers())))
    temporal allMsgsDelivered = 
        eventually(NodeIDs.forall(node => length(node.IncomingMsgs()) == 0))
    // TODO: Invariant: all nodes in the network are always connected.

}
