---
order: 3
---

# Upgrading from Tendermint Core

CometBFT was originally forked from [Tendermint Core v0.34.24][v03424] and
subsequently updated in Informal Systems' public fork of Tendermint Core for
[v0.34.25][v03425] and [v0.34.26][v03426].

If you already make use of Tendermint Core (either the original Tendermint Core
v0.34.24, or Informal Systems' public fork), you can upgrade to CometBFT
v0.34.27 by replacing your dependency in your `go.mod` file:

```bash
go mod edit -replace github.com/tendermint/tendermint=github.com/cometbft/cometbft@v0.34.27
```

We make use of the original module URL in order to minimize the impact of
switching to CometBFT. This is only possible in our v0.34 release series, and we
will be switching our module URL to `github.com/cometbft/cometbft` in the next
major release.

## Home directory

CometBFT, by default, will consider its home directory in `~/.cometbft` from now
on instead of `~/.tendermint`.

## Environment variables

The environment variable prefixes have now changed from `TM` to `CMT`. For
example, `TMHOME` or `TM_HOME` become `CMTHOME` or `CMT_HOME`.

We have implemented a fallback check in case `TMHOME` is still set and `CMTHOME`
is not, but you will start to see a warning message in the logs if the old
`TMHOME` variable is set. This fallback check will be removed entirely in a
subsequent major release of CometBFT.

## Building CometBFT

If you are building CometBFT from scratch, please note that it must be compiled
using Go 1.22 or higher.

[v03424]: https://github.com/tendermint/tendermint/releases/tag/v0.34.24
[v03425]: https://github.com/informalsystems/tendermint/releases/tag/v0.34.25
[v03426]: https://github.com/informalsystems/tendermint/releases/tag/v0.34.26
