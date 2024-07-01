- `[mempool]` Revert adding the method `PreUpdate()` to the `Mempool` interface, recently introduced
  in the previous patch release (`v0.37.7`). Its logic is now moved into the `Lock` method. With this change,
  the `Mempool` interface is the same as in `v0.37.6`.
  ([\#3363](https://github.com/cometbft/cometbft/pull/3363))
