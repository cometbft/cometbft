- [`consensus`] Check for late votes in the reactor, preventing them from being delivered to the consensus logic, as a performance optimization.
single-threaded consensus logic. This change is a performance optimization that reduces the number 
of redundant votes that are processed by the consensus logic.
  ([\#3154](https://github.com/cometbft/cometbft/issues/3154)
