- `[p2p]` Extracted TCP transport into its own package - `transport/tcp`
  * Updated `Transport` interface;
  * Moved `NetAddress`, `NodeInfo` and `NodeKey` into separate packages -
  `netaddress`, `nodeinfo`, `nodekey` accordingly;
  * Internalized `fuzz` package.
  [\#4301](https://github.com/cometbft/cometbft/issues/4301)
