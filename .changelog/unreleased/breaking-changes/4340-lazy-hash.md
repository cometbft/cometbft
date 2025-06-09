- `[log]` LazyBlockHash -> LazyHash
  * LazyBlockHash replaced with more generic LazyHash which lazily evaluates
    a tx or block hash when the stringer interface is invoked. Good for use
    with debug statements so the item is only hashed when print is invoked
  * tx `Hash` ret type changed to HexBytes to fit this interface
  [\#4340](https://github.com/cometbft/cometbft/pull/4340)
