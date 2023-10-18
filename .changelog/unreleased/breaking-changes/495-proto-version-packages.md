- `[proto]` Renamed the packages from `tendermint.*` to `cometbft.*`
  and introduced versioned packages to distinguish between proto definitions
  released in 0.34.x, 0.37.x, 0.38.x, and 0.39.x versions.
  Prior to the eventual 1.0 release, the versioned packages are suffixed with
  `.v1beta1`, `.v1beta2`, and so on.
  Relocated generated Go code into a new `api` folder and changed the import
  paths accordingly.
  ([\#495](https://github.com/cometbft/cometbft/pull/495)
  [\#1504](https://github.com/cometbft/cometbft/issues/1504))
