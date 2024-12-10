# Mempool

This specification of a mempool defines essential types and data structures needed to keep a list of
pending transactions ("the mempool"), as well as generic actions to disseminate transactions. Those
generic actions are then instantiated with specific functions that define the behaviour of the
gossip protocols.

The mempool is built on top of a [P2P layer](p2p.md), which declares many definitions found here.

## Types

### Transactions

A transaction is uniquely identified by a string, which represents its content (typically
implemented as an array of bytes).
```bluespec "types" +=
type TX = str
```

Transactions are validated by an external entity. The validation function must be deterministic. In
the actual implementation, the mempool makes a CheckTx ABCI call to the application, which validates
the transaction. 
```bluespec "types" +=
pure def valid(tx) = true
```

In this simplified specification we model all transactions as valid. To model invalid transactions,
`valid` should be declared as a model parameter (a `const`) and instantiated with a deterministic
function of type `(TX) => bool`.

### Transaction IDs

A transaction identifier, computed as the hash of the transaction (typically a short array of
bytes).
```bluespec "types" +=
type TxID = str
pure def hash(tx: TX): TxID = tx
```

## Parameters
        
The set of all possible transactions.
```bluespec "params" +=
const AllTxs: Set[TX]
```

## State

Each node has a mempool state.
```bluespec "state" +=
var mempool: NodeID -> MempoolState
```

We define `MempoolState` as a data structure with the following fields.

#### Cache of already received transaction IDs

We assume the cache never overflows, i.e., it can grow indefinitely.
```bluespec "mempoolstate" +=
cache: Set[TxID],
```

#### List of uncommitted or pending transactions ("the mempool")

This list is used for storaging transactions and for picking transactions to disseminate to peers.
```bluespec "mempoolstate" +=
txs: List[TX],
```

We make the following assumptions about the mempool:
- It does not have a maximum capacity.
- New entries are only appended. We do not model when entries are removed.

A transaction that is in the `txs` list, must also be in `cache` (assuming an infinite cache), but
not necessarily the inverse. The reason a transaction is in `cache` but not in `txs` is either
because: 
- the transaction was initially invalid and never got into `txs`, 
- the transaction became invalid after it got in `txs` and thus got evicted when it was revalidated,
  or
- the transaction was committed to a block and got removed from `txs`.

All these scenarios are not modeled here. Then `cache` and `txs` will always have the same content
and one of the two is actually redundant in this spec.

#### Index to the next transaction to disseminate

A mempool iterator traverses the entries in `txs` one at a time.
```bluespec "mempoolstate" +=
txsIndex: int,
```
We model transaction dissemination using one dissemination process (`disseminateNextTx`) that
iterates on the list of transactions reading one entry per step, and atomically multicasts one
transaction message to all connected peers.

In the implementation there is one dissemination process per peer, each with its own iterator (and
thus a separate index per iterator) with a `next()` method to retrieve the next entry in the `txs`
list. If it reaches the end of the list, it blocks until a new entry is added. All iterators read
concurrently from `txs`.

<details>
  <summary>Auxiliary definitions</summary>

```bluespec "auxstate" +=
def Cache(node) = mempool.get(node).cache
def Txs(node) = mempool.get(node).txs
def TxsIndex(node) = mempool.get(node).txsIndex
```
</details>

## Initial state

The initial state of a mempool:
```bluespec "actions" +=
action MP_init = all {
    P2P_init,
    mempool' = NodeIDs.mapBy(n => initialMempoolState),
}
```
where:
```bluespec "actions" +=
val initialMempoolState = {
    cache: Set(),
    txs: List(),
    txsIndex: 0,
}
```

## State transitions (actions)

### Handling incoming transactions

Users create transactions and send them to one of the nodes in the network. Nodes receive
transactions either directly from users or in messages from peers. Transaction from users have no
sender.

Action `receiveTxFromUser` models a `node` receiving transaction `tx` from a user.
```bluespec "actions" +=
action receiveTxFromUser(node, tx, _tryAddTx) =
    node._tryAddTx(incomingMsgs, None, tx)
```
The function parameter `_tryAddTx(incomingMsgs, optionalSender, tx)` defines how transactions are
added to the mempool.

Typically, users send (full) transactions to the node via an RPC endpoint. Users are allowed to
submit the same transaction more than once and to multiple nodes.

This action is enabled only if the transaction is not in the mempool. In the actual mempool
implementation we have the cache that prevents this scenario.

### Transaction dissemination

Action `disseminateNextTx` models a `node` traversing the `txs` list while sending transactions to
its peers. It takes the transaction pointed by `txsIndex` and atomically sends it to a set of target
peers.

The following function parameters define to who `node` will send transactions:
- `_mkTargetNodes(node, tx)` returns the set of peers to which `node`
  will send `tx`.
- `_mkTxMsg(tx)` is a wrapper function that returns the specific message
  type used by the gossip protocol.
```bluespec "actions" +=
action disseminateNextTx(node, _mkTargetNodes, _mkTxMsg) = all {
    // Check that the current index is within bounds. 
    require(node.TxsIndex() < node.Txs().length()),
    // Get from the mempool the next transaction to disseminate.
    val tx = node.Txs()[node.TxsIndex()]
    all {
        // Wrap transaction in a message and send it to the target nodes.
        incomingMsgs' = 
            node.multiSend(incomingMsgs, _mkTargetNodes(node, tx), _mkTxMsg(tx)),
        // Increase index.
        mempool' = mempool.update(node, st => { txsIndex: st.txsIndex + 1, ...st }),
        peers' = peers,
    }
}
```

The index must not exceed the `txs`'s length. This pre-condition models when the iterator is at the
end of the list and it's blocked waiting for a new entry to be appended to the list.

In the actual implementation, there is a separate goroutine for each peer, so not all transactions
are sent at the same time.

## Properties

_**Invariant**_ Transaction lists do not have repeated entries.
```bluespec "properties" +=
val uniqueTxsInMempool = 
    NodeIDs.forall(node => size(node.Txs().listToSet()) == length(node.Txs()))
```

<!--
```bluespec quint/mempool.qnt +=
// -*- mode: Bluespec; -*-

// File generated from markdown using https://github.com/driusan/lmt. DO NOT EDIT.

module mempool {
    import spells.* from "./spells"
    import p2p.* from "./p2p"
    export p2p.*

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
    
    type MempoolState = {
        <<<mempoolstate>>>
    }
    
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
