# Upgrading CometBFT

This guide provides instructions for upgrading to specific versions of CometBFT.

## Unreleased

### Config Changes

* A new config field, `BootstrapPeers` has been introduced as a means of
  adding a list of addresses to the addressbook upon initializing a node. This is an
  alternative to `PersistentPeers`. `PersistentPeers` shold be only used for
  nodes that you want to keep a constant connection with i.e. sentry nodes
* The field `Version` in the mempool section has been removed. The priority
  mempool (what was called version `v1`) has been removed (see below), thus
  there is only one implementation of the mempool available (what was called
  `v0`).
* Config fields `TTLDuration` and `TTLNumBlocks`, which were only used by the priority
  mempool, have been removed.

### Mempool Changes

* The priority mempool (what was referred in the code as version `v1`) has been
  removed. There is now only one mempool (what was called version `v0`), that
  is, the default implementation as a queue of transactions. 
* In the protobuf message `ResponseCheckTx`, fields `sender`, `priority`, and
  `mempool_error`, which were only used by the priority mempool, were removed
  but still kept in the message as "reserved".

## v0.37.0

This release introduces state machine-breaking changes, and therefore requires a
coordinated upgrade.

### Go API

When upgrading from the v0.34 release series, please note that the Go module has
now changed to `github.com/cometbft/cometbft`.

### ABCI Changes

* The `ABCIVersion` is now `2.0.0`.
* Added new ABCI methods `ExtendVote`, and `VerifyVoteExtension`.
  Applications upgrading to v0.38.0 must implement these methods as described
  [here](./spec/abci/abci%2B%2B_comet_expected_behavior.md#adapting-existing-applications-that-use-abci)
* Removed methods `BeginBlock`, `DeliverTx`, `EndBlock`, and replaced them by
  method `FinalizeBlock`. Applications upgrading to v0.38.0 must refactor
  the logic handling the methods removed to handle `FinalizeBlock`.
* The Application's hash (or any data representing the Application's current state)
  is known by the time `FinalizeBlock` finishes its execution.
  Accordingly, the `app_hash` parameter has been moved from `ResponseCommit`
  to `ResponseFinalizeBlock`.
* For details, please see the updated [specification](spec/abci/README.md)


### RPC

If you rely on the `/tx_search` or `/block_search` endpoints for event querying,
please note that the default behaviour of these endpoints has changed in a way
that might break your queries. The original behaviour was poorly specified,
which did not respect event boundaries.

Please see
[tendermint/tendermint\#9712](https://github.com/tendermint/tendermint/issues/9712)
for context on the bug that was addressed that resulted in this behaviour
change.

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
