- [`consensus`] Check for late votes in the reactor, preventing them from entering the 
single-threaded consensus logic. This change is a performance optimization that reduces the number 
of redundant votes that are processed by the consensus logic.
  ([\#3157](https://github.com/cometbft/cometbft/issues/3157)
