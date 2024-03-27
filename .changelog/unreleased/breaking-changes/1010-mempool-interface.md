`[mempool]` Change the signature of `CheckTx` in the `Mempool` interface to
`CheckTx(tx types.Tx) (*abcicli.ReqRes, error)`. Also, add new method
`SetTxRemovedCallback`.
([\#1010](https://github.com/cometbft/cometbft/issues/1010))