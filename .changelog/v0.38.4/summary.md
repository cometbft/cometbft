*January 22, 2024*

This release is aimed at those projects that have a dependency on CometBFT,
release line `v0.38.x`, and make use of function `SaveBlockStoreState` in package
`github.com/cometbft/cometbft/store`. This function changed its signature in `v0.38.3`.
This new release reverts the signature change so that upgrading to the latest release
of CometBFT on `v0.38.x` does not require any change in the code depending on CometBFT.

