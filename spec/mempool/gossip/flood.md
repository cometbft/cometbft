# Flood gossip protocol

Flood is a basic _push_ gossip protocol: every time a node receives a transaction, it forwards (or
"pushes") the transaction to all its peers, except to the peer(s) from which it received the
transaction.

This protocol is built on top of the [mempool](mempool.md) and [p2p](p2p.md) modules.

**Table of contents**
  - [Messages](#messages)
  - [State](#state)
  - [Initial state](#initial-state)
  - [State transitions (actions)](#state-transitions-actions)
    - [Adding transactions to the mempool](#adding-transactions-to-the-mempool)
      - [Adding first-time transactions](#adding-first-time-transactions)
      - [Handling duplicate transactions](#handling-duplicate-transactions)
    - [Handling incoming messages](#handling-incoming-messages)
    - [Transaction dissemination](#transaction-dissemination)
  - [Properties](#properties)

> This document was written using the literature programming paradigm. Code snippets are written in
> [Quint][quint] and can get "tangled" into a Quint file.

## Messages

Nodes communicates only one type of message carrying a full transaction.
```bluespec "messages" +=
type Message =
    | TxMsg(TX)
```

## State

Flood's state consists of the underlying [mempool](mempool.md) state (variable `mempool`) and
[P2P](p2p.md) state (variables `incomingMsgs` and `peers`).

Additionally, for each transaction in each node's mempool, we keep track of the peer IDs from whom
the node received the transaction. 
```bluespec "state" +=
var senders: NodeID -> TxID -> List[NodeID]
```
We define the senders as a list instead of a set because the DOG protocol needs to know who is the
first sender of a transaction.

Note that a transaction won't have a sender when it is in the cache but not in the mempool. Senders
are only needed for disseminating (valid) transactions that are in the mempool.

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "auxstate" +=
def Senders(node) = senders.get(node)
```

The set of senders of transaction `tx`:
```bluespec "auxstate" +=
def sendersOf(node, tx) = 
    node.Senders().mapGetDefault(hash(tx), List()).listToSet()
```

Function `addSender` adds a sender to `tx`'s list of senders (`_txSenders`), if `optionalSender` has
a value that's not already in the list.
```bluespec "auxstate" +=
pure def addSender(_txSenders, tx, optionalSender) = 
    match optionalSender {
    | Some(sender) => _txSenders.update(hash(tx), ss => 
        if (ss.includes(sender)) ss else ss.append(sender))
    | None => _txSenders
    }
```
</details>

## Initial state

Flood's initial state is the underlying mempool's initial state (`MP_init`) and an empty mapping of
transactions to senders.
```bluespec "actions" +=
action init = all {
    MP_init,
    senders' = NodeIDs.mapBy(n => Map()),
}
```

## State transitions (actions)

These are the state transitions of the system. Note that generic actions are imported from the
[mempool](mempool.md) and [p2p](p2p.md) specs. The missing implementation details (`tryAddTx`,
`handleMessage`, `mkTargetNodes`) are described in the rest of the section.

1. User-submitted transactions: when a node receives a transaction from a user, it tries to add it
   to the mempool.
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    nondet tx = oneOf(AllTxs)
    node.receiveTxFromUser(tx, tryAddTx),
    ```

2. Peer message handling: a node processes messages received from a peer.
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    node.receiveFromPeer(handleMessage),
    ```

3. Transaction dissemination: a node sends a transaction in its mempool to a subset of target nodes.
    ```bluespec "steps" +=
    nondet node = oneOf(nodesInNetwork)
    all {
        node.disseminateNextTx(mkTargetNodes, TxMsg),
        senders' = senders,
    },
    ```

4. A node joins the network.
    ```bluespec "steps" +=
    all {
        pickNodeAndJoin,
        mempool' = mempool,
        senders' = senders,
    },
    ```

5. A node disconnects from the network.
    ```bluespec "steps" +=
    all {
        pickNodeAndDisconnect,
        mempool' = mempool,
        senders' = senders,
    }
    ```

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
2. if `tx` is valid, it appends `tx` to `txs`, and
3. updates its senders.
```bluespec "actions" +=
action tryAddFirstTimeTx(node, _incomingMsgs, optionalSender, tx) = all {
    mempool' = mempool.update(node, st => {
        cache: st.cache.join(hash(tx)),
        txs: if (valid(tx)) st.txs.append(tx) else st.txs,
        ...st }),
    senders' = senders.update(node, ss =>
        if (valid(tx)) ss.addSender(tx, optionalSender) else ss),
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
}
```

#### Handling duplicate transactions

Action `processDuplicateTx` processes a duplicate transaction `tx` by updating the list of senders,
only if `tx` is already in the mempool (`txs`).
```bluespec "actions" +=
action processDuplicateTx(node, _incomingMsgs, optionalSender, tx) = all {
    senders' = senders.update(node, ss =>
        if (node.Txs().includes(tx)) ss.addSender(tx, optionalSender) else ss),
    mempool' = mempool,
    incomingMsgs' = _incomingMsgs,
    peers' = peers,
}
```

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

### Transaction dissemination 

In Flood, a node sends a transaction to all its peers except those who previously sent it.

`mkTargetNodes` defines the set of peers to whom `node` will send `tx`. It is passed as an argument
to the generic transaction dissemination action.
```bluespec "actions" +=
def mkTargetNodes(node, tx) =
    node.Peers().exclude(node.sendersOf(tx))
```

## Properties

Function `txInAllMempools` returns `true` if the given transaction `tx` is in the mempool of all
nodes.
```bluespec "properties" +=
def txInAllMempools(tx) =
    NodeIDs.forall(n => n.Txs().includes(tx))
```

_**Property**_ If a transaction is in the mempool of any node, then eventually the transaction will
reach the mempool of all nodes (maybe more than once, and assuming transactions are not removed from
mempools).
```bluespec "properties" +=
temporal txInPoolGetsDisseminated = 
    AllTxs.forall(tx => 
        NodeIDs.exists(node =>
            node.Txs().includes(tx) implies eventually(txInAllMempools(tx))))
```

_**Invariant**_ If node A sent a transaction `tx` to node B (A is in the list of `tx`'s senders),
then B does not send `tx` to A (the message won't be in A's incoming messages).
```bluespec "properties" +=
val dontSendBackToSender =
    NodeIDs.forall(nodeA => 
        NodeIDs.forall(nodeB => 
            AllTxs.forall(tx =>
                nodeB.sendersOf(tx).contains(nodeA) 
                implies
                not(nodeA.IncomingMsgs().includes((nodeB, TxMsg(tx))))
    )))
```

<!--
```bluespec quint/flood.qnt +=
// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module flood {
    import spells.* from "./spells"
    import mempool.* from "./mempool"
    export mempool.*

    //--------------------------------------------------------------------------
    // Messages
    //--------------------------------------------------------------------------
    <<<messages>>>

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
