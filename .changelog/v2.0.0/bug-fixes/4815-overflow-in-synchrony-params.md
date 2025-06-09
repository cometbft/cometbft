- `[consensus]` Fix overflow in synchrony parameters in `linux/amd64` architecture.
  Cap `SynchronyParams.MessageDelay` to 24hrs.
  Cap `SynchronyParams.Precision` to 30 sec.
  ([\#4815](https://github.com/cometbft/cometbft/issues/4815))
