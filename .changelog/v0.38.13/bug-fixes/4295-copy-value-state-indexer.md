- `[state/indexer]` Fix the tx_search results not returning all results by changing the logic in the indexer to copy the key and values instead of reusing an iterator. This issue only arises when upgrading to cometbft-db v0.13 or later.
  ([\#4295](https://github.com/cometbft/cometbft/issues/4295)). Special thanks to @faddat for reporting the issue.

