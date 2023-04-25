# Upgrading CometBFT

This guide provides instructions for upgrading to specific versions of CometBFT.

## v0.37.1

For users explicitly making use of the Go APIs provided in the `crypto/merkle`
package, please note that, in order to fix a potential security issue, we had to
make a breaking change here. This change should only affect a small minority of
users. For more details, please see
[\#557](https://github.com/cometbft/cometbft/issues/557).

## v0.37.0

This release introduces state machine-breaking changes, and therefore requires a
coordinated upgrade.

### Go API

When upgrading from the v0.34 release series, please note that the Go module has
now changed to `github.com/cometbft/cometbft`.

### ABCI Changes

* The `ABCIVersion` is now `1.0.0`.
* Added new ABCI methods `PrepareProposal` and `ProcessProposal`. For details,
  please see the [spec](spec/abci/README.md). Applications upgrading to
  v0.37.0 must implement these methods, at the very minimum, as described
  [here](./spec/abci/abci%2B%2B_comet_expected_behavior.md#adapting-existing-applications-that-use-abci)
* Deduplicated `ConsensusParams` and `BlockParams`.
  In the v0.34 branch they are defined both in `abci/types.proto` and `types/params.proto`.
  The definitions in `abci/types.proto` have been removed.
  In-process applications should make sure they are not using the deleted
  version of those structures.
* In v0.34, messages on the wire used to be length-delimited with `int64` varint
  values, which was inconsistent with the `uint64` varint length delimiters used
  in the P2P layer. Both now consistently use `uint64` varint length delimiters.
* Added `AbciVersion` to `RequestInfo`.
  Applications should check that CometBFT's ABCI version matches the one they expect
  in order to ensure compatibility.
* The `SetOption` method has been removed from the ABCI `Client` interface.
  The corresponding Protobuf types have been deprecated.
* The `key` and `value` fields in the `EventAttribute` type have been changed
  from type `bytes` to `string`. As per the [Protocol Buffers updating
  guidelines](https://developers.google.com/protocol-buffers/docs/proto3#updating),
  this should have no effect on the wire-level encoding for UTF8-encoded
  strings.

### RPC

If you rely on the `/tx_search` or `/block_search` endpoints for event querying,
please note that the default behaviour of these endpoints has changed in a way
that might break your queries. The original behaviour was poorly specified,
which did not respect event boundaries.

Please see
[tendermint/tendermint\#9712](https://github.com/tendermint/tendermint/issues/9712)
for context on the bug that was addressed that resulted in this behaviour
change.

---

For historical upgrading instructions for Tendermint Core v0.34.24 and earlier,
please see the [Tendermint Core upgrading instructions][tmupgrade].

[tmupgrade]: https://github.com/tendermint/tendermint/blob/35581cf54ec436b8c37fabb43fdaa3f48339a170/UPGRADING.md
