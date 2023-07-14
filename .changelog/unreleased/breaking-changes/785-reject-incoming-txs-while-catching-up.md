- `[rpc]` The endpoints `broadcast_tx_*` now returns an error when the node is
  catching up, that is, performing block sync or state sync.
  ([\#785](https://github.com/cometbft/cometbft/issues/785))
- `[mempool]` When the node is catching up, that is, performing block sync or
  state sync, the mempool reactor now discards incoming transactions from peers,
  and it also does not propagate transactions to peers.
  ([\#785](https://github.com/cometbft/cometbft/issues/785))
