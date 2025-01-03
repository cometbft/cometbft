# P2P

This module specifies a P2P layer as needed for the gossip protocols. It includes the definitions of
nodes, peers, network topology, sending messages, nodes joining and leaving the network.

## Types

Nodes are identified by a string.
```bluespec "types"
type NodeID = str
```

## Parameters

The set of all possible node IDs, even those that are not initially connected to the network.
```bluespec "params"
const NodeIDs: Set[NodeID]
```

Initial network topology. A topology is defined by the set of peers each node has.
```bluespec "params" +=
const InitialPeers: NodeID -> Set[NodeID]
```

## State

To model network communication, each node has a queue (a list) of incoming messages. Node A sends a
message to a node B by appending the message to B's queue. We use queues to model that messages
arrive in order, as we assume this is guaranteed by the transport layer. Messages have a sender (a
node ID).

The type variable `msg` can be instantiated on the message types of different protocols.

```bluespec "state"
var incomingMsgs: NodeID -> List[(NodeID, msg)]
```

In the actual implementation, transaction messages are transmitted on the `Mempool` data channel of
the P2P layer. Control messages are usually transmitted on other channels with different priorities.
Here we model a single, reliable channel.

The dynamic network topology. Each node has a set of peers that is updated when nodes join or leave
the network.

```bluespec "state" +=
var peers: NodeID -> Set[NodeID]
```

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "auxstate" +=
def IncomingMsgs(node) = incomingMsgs.get(node)
def Peers(node) = peers.get(node)
```

Function `multiSend` sends message `msg` to a set of `targetNodes`. It updates a list of incoming
messages `_incomingMsgs`. `targetNodes` can be empty, in which case `_incomingMsgs` will stay the
same.
```bluespec "state" +=
pure def multiSend(node, _incomingMsgs, targetNodes, msg) =
    _incomingMsgs.updateMultiple(targetNodes, ms => ms.append((node, msg)))
pure def send(node, _incomingMsgs, targetNode, msg) =
    node.multiSend(_incomingMsgs, Set(targetNode), msg)
```

A node is in the network if it has peers:
```bluespec "auxstate" +=
val nodesInNetwork = NodeIDs.filter(node => node.Peers().nonEmpty())
val nodesNotInNetwork = NodeIDs.exclude(nodesInNetwork)
```

A node disconnects from the network when it does not have peers.
```bluespec "auxstate" +=
pure def disconnect(_peers, node) =
    // TODO: check that the network does not become disconnected; we don't want to model that.
    _peers.put(node, Set())
```

The set of `node`'s peers that are not themselves connected to `node`.
```bluespec "auxstate" +=
def DisconnectedPeers(node) = 
    node.Peers().filter(p => not(node.in(p.Peers())))
```
</details>

## Initial state

The initial state of the P2P layer:
```bluespec "actions" +=
action P2P_init = all {
    incomingMsgs' = NodeIDs.mapBy(_ => List()),
    peers' = NodeIDs.mapBy(n => InitialPeers.get(n)),
}
```

## State transitions (actions)

A node receives one of the incoming messages from a peer and handles it according to its type.
```bluespec "actions" +=
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
```

A node joins the network by connecting to a given set of peers. All those peers add the new node to
their list of peers.
```bluespec "actions" +=
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
```

Non-deterministically pick a node and its peers to join the network.
```bluespec "actions" +=
action pickNodeAndJoin = all {
    // Pick a node that is not connected to the network.
    require(NodeIDs.exclude(nodesInNetwork).nonEmpty()),
    nondet node = oneOf(NodeIDs.exclude(nodesInNetwork))
    // Pick a non-empty set of nodes in the network to be the node's peers.
    nondet peerSet = oneOf(nodesInNetwork.powerset().exclude(Set()))
    node.joinNetwork(peerSet),
}
```

## Properties

_**Invariant**_ Peer relationships are bidirectional or symmetrical: if node A has B as peer, then B
has A as peer.
```bluespec "properties" +=
val bidirectionalNetwork =
    NodeIDs.forall(nodeA => 
        nodeA.Peers().forall(nodeB => nodeA.in(nodeB.Peers())))
```

_**Property**_ Eventually all messages are delivered (there are no incoming messages).
```bluespec "properties" +=
temporal allMsgsDelivered = 
    eventually(NodeIDs.forall(node => length(node.IncomingMsgs()) == 0))
```

```bluespec "properties" +=
// TODO: Invariant: all nodes in the network are always connected.
```

<!--
```bluespec quint/p2p.qnt +=
// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module p2p {
    import spells.* from "./spells"

    //--------------------------------------------------------------------------
    // Types
    //--------------------------------------------------------------------------
    <<<types>>>
    
    //--------------------------------------------------------------------------
    // Parameters
    //--------------------------------------------------------------------------
    <<<params>>>

    //--------------------------------------------------------------------------
    // State
    //--------------------------------------------------------------------------
    <<<state>>>
    
    // Auxiliary definitions
    <<<auxstate>>>

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    <<<actions>>>
    
    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    <<<properties>>>

}
```
-->
