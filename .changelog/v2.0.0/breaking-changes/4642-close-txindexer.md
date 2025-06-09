- `[state]`:
  - `TxIndexer` no longer leaks its database's goroutines.
  - `TxIndexer` now exposes a `Close` method to close the indexer's database when done with it.
  ([\#4642](https://github.com/cometbft/cometbft/pull/4642))