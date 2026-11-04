- `[state/indexer]` Fix unbounded memory use when pruning the kv tx and block
  indexers: `TxIndex.Prune` now closes the iterator it opens for each indexed
  event attribute and flushes its batch every 1000 pruned txs as intended, and
  `BlockerIndexer.Prune` flushes only after a delete instead of on every key.
  ([\#6134](https://github.com/cometbft/cometbft/issues/6134))
