- `[statesync]` If the node can't discover snapshots for 2 min
  (`statesync.max_discovery_time`), switch to blocksync. Remove
  `statesync.discovery_time` from the configuration. If
  `statesync.max_discovery_time` is zero, the node will be retrying
  indefinitely.
  [\#3878](https://github.com/cometbft/cometbft/issues/3878)
