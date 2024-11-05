// -*- mode: Bluespec; -*-
module p2p {
    import spells.* from "./spells"

    //--------------------------------------------------------------------------
    // Base types
    //--------------------------------------------------------------------------
    
    // A node internally identifies other nodes with an ID (a string).
    type NodeID = str    

    //--------------------------------------------------------------------------
    // Model parameters
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
    // The protocol is modelled as a state transition system. A state is defined
    // by the values of the following variables.

    // Each node has a list of incoming messages. Messages arrive in order and
    // are appended to the list. Each message has a sender.
    //
    // In the implementation, transaction messages are transamitted via the
    // `Mempool` channel on the P2P layer. Control messages may be transmitted
    // through other channel. This variable models only one channel.
    var msgs: NodeID -> List[(NodeID, m)]

    // The dynamic network topology. Each node has a set of peers.
    var peers: NodeID -> Set[NodeID]

    // Auxiliary definitions
    def Msgs(node) = msgs.get(node)
    def Peers(node) = peers.get(node)

    //--------------------------------------------------------------------------
    // Boilerplate and auxiliary definitions
    //--------------------------------------------------------------------------
    // We assume that nodes have reliable P2P channels to all other nodes and
    // that messages are delivered in the order in which they were sent.

    // A node sends a message to the given set of target nodes.
    pure def send(node, _msgs, targetNodes, msg) =
        _msgs.mapPutMany(targetNodes, p => _msgs.get(p).append((node, msg)))

    // A node receives one of the incoming messages from a peer and handles it
    // according to its type.
    action receiveFromPeer(node, handleMessage) = all {
        require(node.Msgs().listNonEmpty()),
        val someMsg = node.Msgs().head()
        val sender = someMsg._1
        val msg = someMsg._2
        // We model receiving of a message as taking the head of the list of
        // incoming messages.
        val updatedMsgs = msgs.mapPut(node, ms => tail(ms))
        handleMessage(node, updatedMsgs, sender, msg)
    }

    // Nodes connected to the network in the current state.
    pure def currentNodes(_peers: NodeID -> Set[NodeID]): Set[NodeID] = 
        NodeIDs.filter(n => _peers.get(n).nonEmpty())

    //--------------------------------------------------------------------------
    // A node joins the network by connecting to a given set of peers. All its
    // peers add the new node to their list of peers.
    action joinNetwork(node: NodeID, newPeers: Set[NodeID]): bool = all {
        // The node must not be connected to the network.
        require(node.Peers().isEmpty()),
        peers' = peers
            // Assign to node the set of new peers.
            .put(node, newPeers)
            // Add node as a new peer to the nodes currently in the network.
            .mapPutMany(peers.currentNodes(), n => peers.get(n).setAdd(node)),
        msgs' = msgs,
    }

    // A node gets disconnected from the network. All its peers are immediately
    // aware that the node is no longer one of their peers, so their state is
    // updated accordingly.
    //
    // TODO: the network must not become disconnected; we don't want to model
    // that.
    action disconnectNetwork(node) = all {
        val p = peers.currentNodes()
        peers' = peers
            // Clean node's state and remove all its peers.
            .put(node, Set())
            // Remove node from other peers' state.
            .mapPutMany(p, n => peers.get(n).except(node)),
        msgs' = msgs,
    }

}
