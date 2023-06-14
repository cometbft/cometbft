- `[state/kvindex]` Querying event attributes that are bigger than int64 is now
  enabled. We are not supporting reading floats from the db into the indexer
  nor parsing them into BigFloats to not introduce breaking changes in minor
  releases. ([\#771](https://github.com/cometbft/cometbft/pull/771))
