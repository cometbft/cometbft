- `[abci]` Change the proto-derived enum type and constant aliases to the
  buf-recommended naming conventions adopted in the `abci/v1beta4` proto package
  ([\#736](https://github.com/cometbft/cometbft/issues/736)).
- `[proto]` The `Type` enum field is now required to be set to a value other
  than the default `CHECK_TX_TYPE_UNKNOWN` for a valid `CheckTxRequest`
  ([\#736](https://github.com/cometbft/cometbft/issues/736)).
