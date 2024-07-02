# Mempool cache

The *cache* stores the hashes of recently seen transactions. Its purpose is to avoid processing and
adding duplicate transactions to the mempool.

## Adding transactions to the cache

A node that receives a transaction, before trying to add it to the mempool, puts it in the cache, if
it is not already there. Only if it is not in the cache, the node will attempt to add it to the
mempool. If the transaction exists in the cache, it is discarded with an error message.

Another way a transaction may enter the cache is when the node's execution module updates the
mempool. A valid transaction that was included in a block by another node may have never been seen
by the current node, so we added to the cache.

:memo: _Definition_: We say that a node receives a transaction `tx` _for the first time_ when the
node receives `tx` and `tx` is not in the cache.

Because of how transactions are evicted from the cache (see below), it is possible by this
definition that a transaction is received for the first time on multiple occasions.

## Capacity

The cache has a *capacity*, a maximum number of transactions that it can hold. The cache's capacity
MUST be bigger than the mempool's capacity (in number of transactions). For example, in the current
implementation, the cache doubles the capacity of the mempool. 

The cache enforces a FIFO policy for keeping transactions: a transaction will be evicted from the
cache if it is the cache's oldest entry and a new transaction is being added. Therefore, a
transaction stays in the cache longer than it stays in the mempool.

## Removing transactions from the cache

A transaction is removed from the cache also when the application deems the transaction as invalid.
Because transactions are evicted when the cache is full, a transaction could be removed from the
cache and then added back at a later time.
