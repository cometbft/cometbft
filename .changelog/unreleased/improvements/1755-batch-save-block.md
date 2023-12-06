- `[store]` Save block using a single DB batch ([\#1755](https://github.com/cometbft/cometbft/pull/1755))
  Only if block is less than 640kB, otherwise each block part is saved individually.
