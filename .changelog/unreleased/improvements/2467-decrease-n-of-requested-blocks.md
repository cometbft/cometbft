- `[blocksync]` make the max number of downloaded blocks dynamic.
  Previously it was a const 600. Now it's `peersCount * maxRequestsPerPeer (10)` 
  [\#2467](https://github.com/cometbft/cometbft/pull/2467)
