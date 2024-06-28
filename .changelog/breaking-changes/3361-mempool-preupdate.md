- `[mempool]` Revert adding the method `PreUpdate()` to the `Mempool` interface, recently introduced
  in the previous patch release. Its logic is now moved into the `Lock` method. With this change,
  the `Mempool` interface is the same as before the previous patch.
  ([\#3361](https://github.com/cometbft/cometbft/pull/3361))
