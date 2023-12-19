- Change `store`, `state`, `evidence` and `light` databases key format. The
  keys are now grouped by height and lexicographically sorted using the
  [orderedcode](https://github.com/google/orderedcode) library.

  To upgrade, run `cometbft migrate-db`. **Before doing so, it's highly
  recommended to backup your database**. If you're using `goleveldb`, you can
  run `cometbft experimental-compact-goleveldb` to compact the database after
  the migration is done.
  ([\#1814](https://github.com/cometbft/cometbft/pull/1814))
