# Upgrading CometBFT

This guide provides instructions for upgrading to specific versions of CometBFT.
## Unreleased

### ABCI Changes

Non-replay-protected vote extensions are added to CometBFT to allow applications to sign a blob of bytes by CometBFT using validator's private key infrastructure. Compared to existing vote extensions which are replay protected by containing meta information as chain-id, height and round (canonical vote extension) the data of the non-replay-protected vote extensions is signed as provided by the application. The non-replay-protected vote extension field is an optional part of the vote extension mechanism and it is up to the application to protect this field against replay attacks. If the application does not populate the non-replay-protected field and does not read it either in `VerifyVoteExtension` or `PrepareProposal` there are no security implications to consider by the application, other than limiting the fields's max length (`MaxVoteExtensionSize`) in `VerifyVoteExtension`, just as it is done with the replay protected part

## v1.x

CometBFT `v1.0` includes some substantial breaking API changes that will hopefully
allow future changes to be rolled out quicker.

### Versioning

As of v1.0, the CometBFT team provides the following guarantees relating to
versioning:

- **Major version** bumps, such as v1.0.0 to v2.0.0, would generally involve
  changes that _force_ users to perform a coordinated upgrade in order to use
  the new version, such as protocol-breaking changes (e.g. changes to how block
  hashes are computed and thus what the network considers to be "valid blocks",
  or how the consensus protocol works, or changes that affect network-level
  compatibility between nodes, etc.).
- **Minor version** bumps, such as v1.1.0 to v1.2.0, are reserved for rolling
  out new features or substantial changes that do not force a coordinated
  upgrade (i.e. not protocol-breaking), but could potentially break Go APIs.
- **Patch version** bumps, such as v1.0.0 to v1.0.1, are reserved for
  bug/security fixes that are not protocol- or Go API-breaking.

### Building CometBFT

The minimum Go version has been bumped to [v1.23][go123].

### Upgrading Guide (`v0.38` -> `v1.0`)

Starting with the `v1.0` release, instead of providing detailed information
about new features, changes, and other relevant details for upgrading to CometBFT `v1.0` in this document,
we have created a comprehensive upgrading guide from the previous `v0.38.x` release line to this new `v1.0` release.
This guide can be utilized as a valuable resource when upgrading to the CometBFT `v1.0` release.

The upgrading guide includes detailed information about major new features in CometBFT `v1.0`, such as PBTS,
Data Companion API, several enhancements, configuration and genesis updates for a smoother
transition to the new `v1.0` version.

Please see more information on the [Upgrading from CometBFT v0.38 to v1.0](/docs/guides/upgrades/v0.38-to-v1.0.md) guide.

## v0.38.0

This release introduces state machine-breaking changes, as well as substantial changes
on the ABCI interface and indexing. It therefore requires a
coordinated upgrade.

### Config Changes

- The field `Version` in the mempool section has been removed. The priority
  mempool (what was called version `v1`) has been removed (see below), thus
  there is only one implementation of the mempool available (what was called
  `v0`).
- Config fields `TTLDuration` and `TTLNumBlocks`, which were only used by the
  priority mempool, have been removed.

### Mempool Changes

- The priority mempool (what was referred in the code as version `v1`) has been
  removed. There is now only one mempool (what was called version `v0`), that
  is, the default implementation as a queue of transactions.
- In the protobuf message `ResponseCheckTx`, fields `sender`, `priority`, and
  `mempool_error`, which were only used by the priority mempool, were removed
  but still kept in the message as "reserved".

### ABCI Changes

- The `ABCIVersion` is now `2.0.0`.
- Added new ABCI methods `ExtendVote`, and `VerifyVoteExtension`.
  Applications upgrading to v0.38.0 must implement these methods as described
  [here](./spec/abci/abci%2B%2B_comet_expected_behavior.md#adapting-existing-applications-that-use-abci)
- Removed methods `BeginBlock`, `DeliverTx`, `EndBlock`, and replaced them by
  method `FinalizeBlock`. Applications upgrading to `v0.38.0` must refactor
  the logic handling the methods removed to handle `FinalizeBlock`.
- The Application's hash (or any data representing the Application's current state)
  is known by the time `FinalizeBlock` finishes its execution.
  Accordingly, the `app_hash` parameter has been moved from `ResponseCommit`
  to `ResponseFinalizeBlock`.
- Field `signed_last_block` in structure `VoteInfo` has been replaced by the
  more expressive `block_id_flag`. Applications willing to keep the semantics
  of `signed_last_block` can now use the following predicate
    - `voteInfo.block_id_flag != BlockIDFlagAbsent`
- For further details, please see the updated [specification](spec/abci/README.md)

### `block_results` RPC endpoint - query result display change (breaking)

- When returning a block, all block events are displayed within the `finalize_block_events` field.
 For blocks generated with older versions of CometBFT,  that means that block results that appeared
 as `begin_block_events` and `end_block_events` are merged into `finalize_block_events`.
 For users who rely on the events to be grouped by the function they were generated by, this change
 is breaking.

### kvindexer changes to indexing block events

The changes described here are internal to the implementation of the kvindexer, and they are transparent to the
user. However, if you own a fork with a modified version of the indexer, you should be aware of these changes.

- Indexer key for block events will not contain information about the function that returned the event.
The events were indexed by their attributes, event type, the function that returned them, the height and
event sequence. The functions returning events in old (pre `v0.38.0`) versions of CometBFT were `BeginBlock` or `EndBlock`.
As events are returned now only via `FinalizeBlock`, the value of this field has no use, and has been removed.
The main motivation is the reduction of the storage footprint.

Events indexed with previous CometBFT or Tendermint Core versions, will still be transparently processed.
There is no need to re-index the events. This function field is not exposed to queries, and was not
visible to users. However, if you forked CometBFT and changed the indexer code directly to accommodate for this,
this will impact your code.

## v0.37.0

This release introduces state machine-breaking changes, and therefore requires a
coordinated upgrade.

### Go API

When upgrading from the v0.34 release series, please note that the Go module has
now changed to `github.com/cometbft/cometbft`.

### ABCI Changes

- The `ABCIVersion` is now `1.0.0`.
- Added new ABCI methods `PrepareProposal` and `ProcessProposal`. For details,
  please see the [spec](spec/abci/README.md). Applications upgrading to
  v0.37.0 must implement these methods, at the very minimum, as described
  [here](./spec/abci/abci++_app_requirements.md)
- Deduplicated `ConsensusParams` and `BlockParams`.
  In the v0.34 branch they are defined both in `abci/types.proto` and `types/params.proto`.
  The definitions in `abci/types.proto` have been removed.
  In-process applications should make sure they are not using the deleted
  version of those structures.
- In v0.34, messages on the wire used to be length-delimited with `int64` varint
  values, which was inconsistent with the `uint64` varint length delimiters used
  in the P2P layer. Both now consistently use `uint64` varint length delimiters.
- Added `AbciVersion` to `RequestInfo`.
  Applications should check that CometBFT's ABCI version matches the one they expect
  in order to ensure compatibility.
- The `SetOption` method has been removed from the ABCI `Client` interface.
  The corresponding Protobuf types have been deprecated.
- The `key` and `value` fields in the `EventAttribute` type have been changed
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
example, `TMHOME` becomes `CMTHOME`.

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
[go123]: https://go.dev/blog/go1.23
[pbts-spec]: ./spec/consensus/proposer-based-timestamp/README.md
