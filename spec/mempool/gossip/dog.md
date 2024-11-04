## DOG gossip protocol

The Dynamic Optimal Graph (DOG) protocol is built on top of the [Flood protocol][flood]. All types,
messages, and data structures in Flood are also present in this specification.

## Description

...

## Definitions

- A transaction is received *for the first time* by a node, when the node does not have the
  transaction in its cache. Otherwise the transaction is a *duplicate*.

## Messages

In addition to `Tx(tx)` from [Flood][flood], two new kind of messages are added:
- `HaveTx(txID)`
  - "I have this tx already; don't send me any more txs from the same source as txID."
  - for cutting cycles, where txID is a hash of the full transaction, that is, a (small) bytes array
  - `HaveTx` messages carry a transaction hash, so their size is typically small compared to Tx(tx)
    messages.
- `Reset` 
  - "My situation has changed; reset my routing data on your side."
  - for dynamic re-routing when a node disconnects (more on this later).

Alternative names: 
- `HaveTx(txID)` → `SendLess(txID)`/`SendLessFrom(txID)`
- `Reset` → `SendMore`

## Parameters

- `config.TargetRedundancy: float`: The redundancy level that the gossip protocol should aim to
  maintain.
- `config.TargetRedundancyDeltaPercent: float`: Value in the range `[0, 1)` that defines the bounds
of acceptable redundancy levels; redundancy +- redundancy*delta TxsPerAdjustment: int
- `config.TxsPerAdjustment: int`: How many (first-time) transactions should the node receive before
  attempting to adjust redundancy.

On startup, define the constants:
- `delta := config.TargetRedundancy * config.TargetRedundancyDeltaPercent`
- `redundancyLowerBound := config.TargetRedundancy - delta`
- `redundancyUpperBound := config.TargetRedundancy + delta`

When `config.TargetRedundancy = 0`, the Redundancy Control mechanism is partially disabled (see below).

## State

For routing:
- `disabledRoutes: Set[(NodeID, NodeID)]`

A route is a tuple `(A,B)` where `A`, `B` have `NodeID` type. We also write routes as `A -> B`.

For the Redundancy Control mechanism:
- `firstTimeTxs: int`: counts received transactions that are not in the cache
- `duplicateTxs: int`
- `isHaveTxBlocked: bool`

## Initial state

- `disabledRoutes := Set()`
- `firstTimeTxs := 0`
- `duplicateTxs := 0`
- `isHaveTxBlocked := false`

## State transitions

### Handling incoming messages

Upon receiving a `Tx(tx)` message from `sender`, do `handleTxMessage(tx, sender)` where:
```
def handleTxMessage(tx, sender):
  if not(cache.contains(tx.ID)):
R1:
    cache.add(tx.ID)
    if valid(tx):
      pool.append(tx)
      if sender != nil:
        senders[tx.ID].add(sender.ID)
    firstTimeTxs++
    if firstTimeTxs == config.TxsPerAdjustment:
      adjustRedundancy()
  else:
R2:
  if sender != nil and pool.contains(tx):
    senders[tx.ID].add(sender.ID)
  duplicateTxs++
  if not(isHaveTxBlocked):
    peer.send(HaveTx(tx.ID))
    isHaveTxBlocked := true
```

Rule `R1` as the same as in [Flood][flood] except that at the end it increases the `firstTimeTxs`
counter and calls `adjustRedundancy()` every `config.TxsPerAdjustment` transactions received for the
first time.

```
def adjustRedundancy():
  redundancy := duplicateTxs / firstTimeTxs
  if redundancy < r.redundancyLowerBound:
    peer.send(Reset)
  if redundancy >= redundancyUpperBound:
    isHaveTxBlocked := false
  firstTimeTxs, duplicateTxs := 0, 0
```

When `config.TargetRedundancy = 0`, the lower and upper bounds are also equal to 0. Then every
`config.TxsPerAdjustment` received transactions `adjustRedundancy` will unblock `HaveTx` but it will
not send `Reset` messages.

Rule `R2` increases the `duplicateTxs` counter and sends a `HaveTx` message if the RC mechanism is
not blocking it.

```
def handleHaveTxMessage(txID, sender):
R3:
  disabledRoutes.add(senders(txID) → A)
```

```
def handleResetMessage(sender):
R4:
  disabledRoutes.remove(sender.ID) // remove all disabled routes with sender.ID as source or target
```
where:
- `disabledRoutes.add(A -> B)` adds the route `A -> B` to the list of disabled routes.
- `disabledRoutes.remove(A)` removes all routes where node ID `A` is a source or a target.

With rule `R3`, we don't want to cut all routes from B to A, only those that come from the
transaction's original sender, that's why we need to take the source into account in rule `R3`.

Rule `R4` is for allowing traffic to flow again to A and nodes will dynamically adapt to the new
traffic, closing routes when needed.

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

The dissemination process iterates over `pool` reading the next transaction to send to `peer`. It
sends the transaction only if the node did not receive it from `peer` (tag `D1`).

### Peer disconnection

A node calls `removePeer(peer)` when it detects that `peer` is disconnected.
```
def removePeer(peer):
D:
  disabledRoutes.remove(peer.ID)
  peers.exclude(peer).send(Reset) // Broadcast a `Reset` message to all other peers
```

Rule D signals other peers that the node's situation has changed and its routing tables should be
reset so that data can be re-routed if needed.

## Properties

### Latency

The protocol implicitly favors routes with low latency, by cutting routes to peers that send the
duplicate transaction at a later time.

There are no extra round trip messages as in push-pull gosssip protocols.

### Dynamic network topology

What happens when a node joins the network? The new node and its peers will automatically adapt and
close routes if they start receiving duplicate transactions. 

What happens when a node is disconnected from the network, for whatever reason? Its peers need to
update their routes that have the node as source or target. Here's where we use rule D and the `Reset`
message:

If node A detects that its peer B is disconnected, remove any route that has B as source or target,
and broadcast to all other peers a `Reset` message. This is to signal A's peers that A's situation has
changed and its routing data should be reset so it can be rerouted if needed.

On receiving a `Reset` message from A, remove any route that has A as source or target. This will
allow traffic to flow again to A and nodes will dynamically adapt to the new traffic, closing routes
when needed.

### Tolerance to malicious behaviour

A node sending `HaveTx` messages can only stop peers from sending transactions to itself; it cannot
affect how other nodes receive transactions. Similarly for `Reset` messages, which affect only the
routes of the node sending it.

Sending the same transaction to multiple nodes simultaneously. This is not necessarily an attack. A
user may decide to send its transaction to multiple nodes to increase the chances of being included
in a block.

### Fairness

Some nodes may end up handling most of the traffic.

[flood]: flood.md
