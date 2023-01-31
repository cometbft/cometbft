- `[p2p]` Reactor `Send`, `TrySend` and `Receive` renamed to `EnvelopeSend` to
  allow a metric to be appended to the message and measure bytes sent/received
  by message type.
  ([\#230](https://github.com/cometbft/cometbft/pull/230))