** This (unfinished) document was used to collect a common understanding of the mempool. We will need to see how/whether this should be incorporated in the specification at some moment **

# Mempool

The mempool is a distributed pool of pending transactions.
A pending transaction is a valid transaction that has been submitted by a
client of the blockchain but has not yet been committed to the blockchain.
The mempool is thus fed with client transactions,
that a priori can be submitted to any node in the network.
And it is consumed by the consensus protocol, more specifically by validator nodes,
which retrieve from the mempool transactions to be included in proposed blocks.

More concretely, every node participating in the mempool protocol maintains a
local copy of the mempool, namely a list of pending transactions.
Nodes that expose an interface to receive transactions from clients
append the submitted transactions to their local copy of the mempool.
These nodes are the entry point of the mempool protocol,
and by extension of the consensus protocol.
Nodes that play the role of validators in the consensus protocol,
in their turn, retrieve from their local copy of the mempool
pending transactions to be included in proposed blocks.
Validator nodes are therefore the recipients of the transactions stored and
transported by the mempool protocol.

The goal of the mempool protocol is then to convey transactions
from the nodes that act as entry points of the network
to the nodes whose role is to order transactions.

## Interactions

### RPC server

Clients submit transactions through the RPC endpoints offered by certain
(public) nodes, which add the submitted transactions to the mempool.

### ABCI application

The mempool should only store and convey valid transactions.
It is up to the ABCI application to define whether a transaction is valid.

Transactions received by a node are sent to the application to be validated,
through the CheckTx method from the mempool ABCI connection.
This applies for both transactions received from a client and transactions
received from a peer in the mempool protocol.
Transactions that are validated by the application are appended to the local
copy of the mempool.
Transactions considered invalid by the application are droped, therefore are
not added to the local copy of the mempool.

The validity of a transaction may depend on the state of the application.
In particular, some transactions that were valid considering a given state of
the application can become invalid when the state of the application is updated.
The state of the application is updated when a commited block of transactions
is delivered to the application for being executed.
Thus, whenever a new block is committed, the list of pending transactions
stored in the mempool is updated to exclude the executed transactions and
 sent to the application to be validated against the
new state of the application.
Transactions that have become invalid with the new state of application are
then removed from the mempool.

### Consensus: validators

The consensus protocol consumes pending transactions stored in the mempool to
build blocks to be proposed.
More precisely, the consensus protocol requests to the mempool a list of
pending transactions that respects certain limits, in terms of the number of
transactions returned, their total size in bytes, and their required gas.
The mempool then returns the longest prefix of its local list of pending
transactions that respects the limits established by the consensus protocol.
This means that the order with which the transactions are stored in the mempool
is preserved when transactions are provided to the consensus protocol.

> Notice that the transactions provided to the consensus protocol are not
> removed from the mempool, as they are still pending transactions albeit being
> included in a proposed block.

As proposing blocks is a prerogative of nodes acting as validators,
only validator nodes interact with the mempool in this way.

### Consensus: all nodes

The consensus protocol is responsible for committing blocks of transactions to
the blockchain.
Once a block is committed to the blockchain, all transactions included in the
block should be removed from the mempool, as they are no any longer pending.
The consensus protocol thus, as part of the procedure to commit a block,
informs the mempool the list of transactions included in the committed block.
The mempool then removes from its local list of pending transactions the
transactions that were included in the committed block, and therefore are no
longer pending.
This procedure precedes the re-validation of transactions against the new state
of the application, which is part of this same procedure to commit a block.

> **Note**    
> Notice that a node can commit blocks to the blockchain through different
> procedures, such as via the block sync protocol.
> The above operation should be part of these other procedures, and should be
> performed whenever a node commits a new block to the blockchain.

## Formalization Attempt

Let `committed(tx, h)` return true iff the transaction `tx` is committed by
height `h` of the blockchain.
This means that `tx` was included in a block `B_k` of transactions with height `k <= h`.

> Notice that `committed(tx, h) => committed(tx, h')` for all `h' > h`.

Lets `valid(tx, h)` to return true iff the transaction `tx` is valid at height
`h` of the blockchain.

> The height of the blockchain reflects the state of the application.
> Some transactions can be valid at a given state (height), but no longer valid
> in another state (height).

We need to assume for the `valid` predicate a property similar to the one we
have for the `committed` predicate.
That is, that from a given height the state of the transaction is not changed
anymore.
While this is trivial for the `committed` predicate, we need to assume that
behavior for the `valid` predicate.

So, we need to assume that there is a height `h*` from which a transaction `tx`
becomes invalid, and this state does not change for higher heights.
That is, we assume that there is a height `h*` so that for all `h >= h* =>
valid(tx, h) == FALSE`.

This property is realistic as applications are expected to implement some form
of replay protection.
This means that once a transaction `tx` is committed, the application should
reject processing the same transaction `tx` again.
This is achieved by rejecting `tx` as invalid from the moment that `tx` is
committed and executed by the application.
So, a replay protection procedure can be define as:
`committed(tx, h) => h' > h \and !valid(tx, h')`

Given these definitions, the main property of the mempool is the following:

 - If a transaction `tx` is added to the local mempool of a correct node, then
   eventually we have a height `h*` for which either:
   - `committed(tx, h*)` is true
   - or `!valid(tx, h*)` and `h > h* => !valid(tx, h)`
