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

### ABCI Changes

* The `ABCIVersion` is now `1.0.0`.

* Added new ABCI methods `PrepareProposal` and `ProcessProposal`. For details,
  please see the [spec](spec/abci/README.md). Applications upgrading to
  v0.37.0 must implement these methods, at the very minimum, as described
  [here](./spec/abci/abci++_app_requirements.md)
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

### Mempool Changes

* The priority mempool (what was referred in the code as version `v1`) has been
  removed. There is now only one mempool (what was called version `v0`), that
  is, the default implementation as a queue of transactions. 
* In the protobuf message `ResponseCheckTx`, fields `sender`, `priority`, and
  `mempool_error`, which were only used by the priority mempool, were removed
  but still kept in the message as "reserved".
