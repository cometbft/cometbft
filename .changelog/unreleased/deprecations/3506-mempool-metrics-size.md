- `[mempool/metrics]` Mark metrics `mempool_size` and `mempool_size_bytes` as
  deprecated, as now they can be obtain, respectively, as the sum of
  `mempool_lane_size` and `mempool_lane_bytes`
  ([\#3506](https://github.com/cometbft/cometbft/issue/3506)).
