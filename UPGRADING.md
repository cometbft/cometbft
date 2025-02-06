# Upgrading CometBFT

This guide provides instructions for upgrading to specific versions of CometBFT.

## v0.34.35

It is recommended that CometBFT be built with Go v1.22+ since v1.21 is no longer
supported.

## v0.34.33

It is recommended that CometBFT be built with Go v1.21+ since v1.20 is no longer
supported.

## v0.34.29

It is recommended that CometBFT be built with Go v1.20+ since v1.19 is no longer
supported.

## v0.34.28

For users explicitly making use of the Go APIs provided in the `crypto/merkle`
package, please note that, in order to fix a potential security issue, we had to
make a breaking change here. This change should only affect a small minority of
users. For more details, please see
[\#557](https://github.com/cometbft/cometbft/issues/557).

## v0.34.27

This is the first official release of CometBFT, forked originally from
[Tendermint Core v0.34.24][v03424] and subsequently updated in Informal Systems'
public fork of Tendermint Core for [v0.34.25][v03425] and [v0.34.26][v03426].

### Upgrading from Tendermint Core

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

### Home directory

CometBFT, by default, will consider its home directory in `~/.cometbft` from now
on instead of `~/.tendermint`.

### Environment variables

The environment variable prefixes have now changed from `TM` to `CMT`. For
example, `TMHOME` or `TM_HOME` become `CMTHOME` or `CMT_HOME`.

We have implemented a fallback check in case `TMHOME` is still set and `CMTHOME`
is not, but you will start to see a warning message in the logs if the old
`TMHOME` variable is set. This fallback check will be removed entirely in a
subsequent major release of CometBFT.

### Building CometBFT

CometBFT must be compiled using Go 1.19 or higher. The use of Go 1.18 is not
supported, since this version has reached end-of-life with the release of [Go 1.20][go120].

### Troubleshooting

If you run into any trouble with this upgrade, please [contact us][discussions].

---

For historical upgrading instructions for Tendermint Core v0.34.24 and earlier,
please see the [Tendermint Core upgrading instructions][tmupgrade].

[v03424]: https://github.com/tendermint/tendermint/releases/tag/v0.34.24
[v03425]: https://github.com/informalsystems/tendermint/releases/tag/v0.34.25
[v03426]: https://github.com/informalsystems/tendermint/releases/tag/v0.34.26
[discussions]: https://github.com/cometbft/cometbft/discussions
[tmupgrade]: https://github.com/tendermint/tendermint/blob/35581cf54ec436b8c37fabb43fdaa3f48339a170/UPGRADING.md
[go120]: https://go.dev/blog/go1.20
