- `[mempool]` Revert adding the method `PreUpdate()` to the `Mempool` interface, recently introduced
  in the previous patch release (v0.38.8). Its logic is now moved into the `Lock` method. With this change,
  the `Mempool` interface is the same as in v0.38.7.
  ([\#3361](https://github.com/cometbft/cometbft/pull/3361))
