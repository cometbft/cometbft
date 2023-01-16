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
and by extension of Tendermint consensus protocol.
Nodes that play the role of validators in the consensus protocol,
in their turn, retrieve from their local copy of the mempool
pending transactions to be included in proposed blocks.
Validator nodes are therefore the recipients of the transactions stored and
transported by the mempool protocol.

The goal of the mempool protocol is then to convey transactions
from the nodes that act as entry points of Tendermint
to the nodes whose role is order transactions.
