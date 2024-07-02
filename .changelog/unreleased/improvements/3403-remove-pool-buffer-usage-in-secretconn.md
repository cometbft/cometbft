- `[p2p/secretconn]` Remove the usage of pooled buffers in secret connection, instead storing the buffer in the connection struct. This reduces the synchronization primitive usage, speeding up the code.
  ([\#3403](https://github.com/cometbft/cometbft/issues/3403))
