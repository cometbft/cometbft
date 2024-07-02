- [`state`] Run SaveABCIResponses and app.Commit in parallel within ApplyBlock,
  lowering latency for blocks with many large ABCI responses.
  ([\#3291](https://github.com/cometbft/cometbft/issues/3291)
