- `[pubsub|kvindexer]` Pubsub queries are now able to parse big integers (larger than int64). Very big floats 
   are also properly parsed into very big integers instead of being truncated to int64. This PR introduces the following Go-API
   breaking changes:
    - `LowerBoundValue()/UpperBoundValue()` does not recognize `int64` anymore, supports rather `big.Int`
    -  `Conditions()` recognizes only `big.Int` and is not supporting `int64` anymore
  ([\#771](https://github.com/cometbft/cometbft/pull/771))