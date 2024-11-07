// -*- mode: Bluespec; -*-
module p2p {
    import spells.* from "./spells"

    // A node internally identifies other nodes with an ID (a string).
    type NodeID = str    

    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    
    // The set of all nodes that participate in the network, even those that are
    // not initially connected.
    const NodeIDs: Set[NodeID]

    // Initial network topology. The topology is defined by the set of peers
    // each node has (the values of the mapping).
    const InitialPeers: NodeID -> Set[NodeID]

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    // The protocols are modelled as a state transition systems. States are
    // defined by the values of the variables.

    // Each node has a queue (a list) of incoming messages. Messages arrive in
    // order and are appended to the queue. Each message has a sender (a node
    // ID).
    //
    // In the implementation, transaction messages are transamitted via the
    // `Mempool` channel on the P2P layer. Control messages may be transmitted
    // through other channel. This variable models only one channel.
    var msgs: NodeID -> List[(NodeID, m)]

    // The dynamic network topology. Each node has a set of peers that is
    // updated when nodes join or leave the network.
    var peers: NodeID -> Set[NodeID]

    // Auxiliary definitions
    def Msgs(node) = msgs.get(node)
    def Peers(node) = peers.get(node)

    //--------------------------------------------------------------------------
    // Boilerplate
    //--------------------------------------------------------------------------
    // We assume that nodes have reliable P2P channels to all other nodes and
    // that messages are delivered in the order in which they were sent.

    // A node sends a message to the given set of target nodes.
    pure def send(node, _msgs, targetNodes, msg) =
        if (targetNodes.nonEmpty())
            _msgs.updateMultiple(targetNodes, ms => ms.append((node, msg)))
        else _msgs

    // A node receives one of the incoming messages from a peer and handles it
    // according to its type.
    action receiveFromPeer(node, handleMessage) = all {
        require(length(node.Msgs()) > 0),
        val someMsg = node.Msgs().head()
        val sender = someMsg._1
        val msg = someMsg._2
        // We model receiving of a message as taking the head of the list of
        // incoming messages.
        val updatedMsgs = msgs.update(node, tail)
        handleMessage(node, updatedMsgs, sender, msg)
    }

    // A node is in the network if it has peers.
    val nodesInNetwork = NodeIDs.filter(node => node.Peers().nonEmpty())
    val nodesNotInNetwork = NodeIDs.exclude(nodesInNetwork)

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------

    // A node joins the network by connecting to a given set of peers. All those
    // peers add the new node to their list of peers.
    action join(node, peerSet) = all {
        // The node must not be connected to the network.
        require(node.Peers().isEmpty()),
        peers' = peers
            // Assign to node the set of new peers.
            .put(node, peerSet)
            // Add node as a new peer to the set of connecting peers.
            .updateMultiple(peerSet, ps => ps.setAdd(node)),
        msgs' = msgs,
    }

    action pickNodeAndJoin = all {
        // Pick a node that is not connected to the network.
        require(NodeIDs.exclude(nodesInNetwork).nonEmpty()),
        nondet node = oneOf(NodeIDs.exclude(nodesInNetwork))
        // Pick a non-empty set of nodes in the network to be the node's
        // peers.
        nondet peerSet = oneOf(nodesInNetwork.powerset().exclude(Set()))
        node.join(peerSet),
    }

    // A node gets disconnected from the network. All its peers are immediately
    // aware that the node is no longer one of their peers, so their state is
    // updated accordingly.
    //
    // TODO: the network must not become disconnected; we don't want to model
    // that.
    action disconnect(disconnectedNode, updatedMsgs) = all {
        peers' = peers
            // Clean node's state and remove all its peers.
            .put(disconnectedNode, Set())
            // Remove node from other peers' state.
            .updateMultiple(nodesInNetwork, ps => ps.except(disconnectedNode)),
        msgs' = updatedMsgs,
    }

    action pickNodeAndDisconnect = all {
        // Pick a node to disconnect only if it is not the only node in the
        // network.
        require(size(nodesInNetwork) > 1),
        nondet disconnectedNode = oneOf(nodesInNetwork) 
        disconnect(disconnectedNode, msgs),
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------

    // Peer relationships are bidirectional or symmetrical: if node A has B as
    // peer, then B has A as peer.
    val peerBidirectional =
        NodeIDs.forall(node => node.Peers().forall(peer => node.in(peer.Peers())))


    // TODO: Network is not disconnected.

}
