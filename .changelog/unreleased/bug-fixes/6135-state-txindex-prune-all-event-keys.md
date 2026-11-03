- `[state/txindex]` Fix `TxIndex.Prune` leaving part of the pruned txs' event
  keys on disk: it now deletes the keys by prefix instead of a numeric
  event-sequence range, which string ordering cut short (e.g. `$es$93` sorts
  after `$es$9223372036854775807`).
  ([\#6135](https://github.com/cometbft/cometbft/issues/6135))
