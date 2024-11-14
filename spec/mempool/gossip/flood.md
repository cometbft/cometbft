# Flood gossip protocol

Flood is a basic _push_ gossip protocol: every time a node receives a transaction, it forwards (or
"pushes") the transaction to all its peers, except to the peer(s) from which it received the
transaction.

This protocol is built on top of a [mempool module](mempool.md) and a [p2p layer](p2p.md).

> This document is written using the literature programming paradigm. Code snippets are written in
> [Quint][quint] and can get "tangled" into a Quint file.

## Messages

Nodes communicates only one type of message carrying a full transaction.
```bluespec "messages" +=
type Message =
    | TxMsg(TX)
```

## State

Flood's state is just the underlying [mempool](mempool.md) state (variable `state`). It does not
need extra data structures.

## Initial state

Flood's initial state is the underlying mempool's initial state (`init`).

## State transitions (actions)

### Handling incoming messages

Upon receiving a message with transaction `tx` from a peer (i.e., the `sender`), the `node` attempts
to add `tx` to its mempool. 
```bluespec "actions" +=
action handleMessage(node, _incomingMsgs, sender, msg) =
    match msg {
    | TxMsg(tx) => node.tryAddTx(_incomingMsgs, Some(sender), tx)
    }
```
> The argument `_incomingMsgs` is passed just to update the queues of incoming messages, when
applicable (Flood does not reply with any message but DOG does).

### Adding transactions to the mempool

A node attempting to add a transaction to its mempool processes the transaction according to whether
it has seen it before, that is, if the transaction exists in the mempool cache.
- A *first-time* transaction is one that the node does not have in its cache. 
- A *duplicate* transaction is one that the node has received multiple times, and thus it's cached.

```bluespec "actions" +=
action tryAddTx(node, _incomingMsgs, optionalSender, tx) = 
    if (not(hash(tx).in(node.Cache())))
        node.tryAddFirstTimeTx(_incomingMsgs, optionalSender, tx)
    else
        node.processDuplicateTx(_incomingMsgs, optionalSender, tx)
```
In this action the sender is optional. When there's a sender, it means that the transaction comes
from a peer; otherwise it comes directly from a user.

#### Adding first-time transactions

`tryAddFirstTimeTx` attempts to add a first-time transaction `tx` to a
`node`'s mempool:
1. it caches `tx`, 
2. if `tx` is valid, it adds `tx` to the pool, and
3. updates its senders.
```bluespec "actions" +=
action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = all {
    state' = state.update(node, st => {
        cache: st.cache.join(hash(tx)),
        pool: if (valid(tx)) st.pool.append(tx) else st.pool,
        senders: 
            if (valid(tx)) 
                st.senders.addSender(tx, optionalSender) 
            else st.senders,
        ...st }),
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
}
```

#### Handling duplicate transactions

`processDuplicateTx` processes a duplicate transaction `tx` by updating the list of senders, only if
`tx` is already in `pool`.
```bluespec "actions" +=
action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) = all {
    state' = state.update(node, st => { 
        senders: 
            if (st.pool.includes(tx)) 
                st.senders.addSender(tx, optionalSender) 
            else st.senders, 
        ...st }),
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
}
```

### Transaction dissemination 

In Flood, a node sends a transaction to all its peers except those who previously sent it.

`mkTargetNodes` defines the set of peers to whom `node` will send `tx`. It is passed as an argument
to the generic transaction dissemination action.
```bluespec "actions" +=
def mkTargetNodes(node, tx) =
    node.Peers().exclude(node.sendersOf(tx))
```

### All state transitions

Summing up, these are all the possible state transitions of the protocol.

#### Transaction dissemination: node sends transaction to subset of peers
```bluespec "steps" +=
nondet node = oneOf(nodesInNetwork)
node.disseminateNextTx(mkTargetNodes, TxMsg),
```

#### User-submitted transactions: node receives a transaction from a user
```bluespec "steps" +=
nondet node = oneOf(nodesInNetwork)
nondet tx = oneOf(Txs)
node.receiveTxFromUser(tx, tryAddTx),
```

#### Peer message handling: node processes messages received from peers
```bluespec "steps" +=
nondet node = oneOf(nodesInNetwork)
node.receiveFromPeer(handleMessage),
```

#### Nodes joining the network
```bluespec "steps" +=
all {
    pickNodeAndJoin,
    state' = state,
},
```

#### Nodes leaving the network
```bluespec "steps" +=
all {
    pickNodeAndDisconnect,
    state' = state,
}
```

## Properties

`txInAllPools` defines if a given transaction is in the pool of all nodes.
```bluespec "properties" +=
def txInAllPools(tx) =
    NodeIDs.forall(n => n.Pool().includes(tx))
```

_**Property**_ If a transaction is in the pool of any node, then eventually the transaction will
reach the pool of all nodes (maybe more than once, and assuming transactions are not removed from
mempools).
```bluespec "properties" +=
temporal txInPoolGetsDisseminated = 
    Txs.forall(tx => 
        NodeIDs.exists(node =>
            node.Pool().includes(tx) implies eventually(txInAllPools(tx))))
```

_**Invariant**_ If node A sent a transaction tx to node B (A is in the list of tx's senders), then B
does not send tx to A (the message won't be in A's incoming messages).
```bluespec "properties" +=
val noSendToSender =
    NodeIDs.forall(nodeA => 
        NodeIDs.forall(nodeB => 
            Txs.forall(tx =>
                nodeB.sendersOf(tx).contains(nodeA) 
                implies
                not(nodeA.IncomingMsgs().includes((nodeB, TxMsg(tx))))
    )))
```

<!--
```bluespec quint/flood.qnt +=
// -*- mode: Bluespec; -*-

// File generated from markdown using lmt. DO NOT EDIT.

module flood {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    export mempool.*

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------
    <<<messages>>>

    //--------------------------------------------------------------------------
    // Actions
    //--------------------------------------------------------------------------
    <<<actions>>>

    action step = any {
        <<<steps>>>
    }

    //--------------------------------------------------------------------------
    // Properties
    //--------------------------------------------------------------------------
    <<<properties>>>

}
```
-->

[quint]: https://quint-lang.org/
