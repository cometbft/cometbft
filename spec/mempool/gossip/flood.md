# Flood protocol

## Description

Flood is a simple *push* gossip protocol: every time it receives a transaction, it forwards or
pushes it to all of its peers, except to the peer(s3) from which it received the transaction, if any.

## Types and notation

Types:
- `Tx` is the content of a transaction (typically an array of bytes).
- `TxID` is a transaction identifier, computed as the hash of the transaction (a short array of
  bytes).
- `NodeID` is a string used internally by the node to identify its peers.

In the following,
- `tx.ID = Hash(tx)` is the hash of transaction `tx`, of `TxID` type, and
- `peer.ID` is the identifier of `peer`, of `NodeID` type.

## Network

The network is comprised by a set of nodes. Each node has a set of peers. This defines the topology
of the network.

We assume that a node has reliable P2P channels to each peer.
- `peer.send(msg)` sends `msg` to `peer`.

## Messages

The Flood gossip protocol only communicates `Tx(tx)` messages, where `tx` is a full transaction of
arbitrary size.

Messages are transamitted via the `Mempool` channel on the P2P layer.

## State

The values of the following data structures partially define the state of the protocol.
- `pool: List[Tx]`: a concurrent list of pending transactions ("the mempool").
- `cache: Set[TxID]`: a set of transaction IDs (hashes).
- `senders: map TxID → Set[NodeID]`: for each transaction in `pool`, a set of node IDs from which
  the node received the transaction. To keep track of all peers that send each transaction.
- There is one iterator `iter` per peer to traverse `pool`. It has only one method `next()` to
  retrieve the next entry in the list. If it reaches the end of the list, it blocks until a new
  entry is added. All iterators read concurrently from `pool`.

## Initial state

All data structures are initially empty when the mempool starts.
- `pool = List()`
- `cache = Set()`
- `senders = Map()`

## State transitions

### Handling incoming messages

Upon receiving a `Tx(tx)` message from `sender`, do `handleTxMessage(tx, sender)` where:
```
def handleTxMessage(tx, sender):
  if not(cache.contains(tx.ID)):
R1: 
    cache.add(tx.ID)
    pool.append(tx)
    if sender != nil:
      senders[tx.ID].add(sender.ID)
  else:
R2: 
    if sender != nil:
      senders[tx.ID].add(sender.ID)
```
- When the message is sent from a user, and not from a peer, the message has no sender, then
`sender` is `nil`.
- The node received the transaction either for the first time (tag `R1`) or it was received before
  (`R2`).
- All actions in the tags should be executed atomically.

### Dissemination

A node runs one `disseminateTo(peer)` process per connected `peer`:
```
def disseminateTo(peer):
  iter := pool.newIterator()
  while true:
    tx := iter.next()
D1:
    if not(senders[tx.ID].contains(peer.ID)):
      peer.send(Tx(tx))
```

The process iterates over `pool` reading the next transaction to send to `peer`. It sends the
transaction only if the node did not receive it from `peer` (tag `D1`).

## Properties

Pros:
+ Optimal latency (given the constrains of the network topology)
+ Tolerate malicious behaviour (BFT)

Cons:
- Bandwidth: exponential redundancy
