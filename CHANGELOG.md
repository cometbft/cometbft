# CHANGELOG

## v0.37.14

*December 20, 2024*

This release adjusts `reconnectBackOffBaseSeconds` to increase reconnect retries to up
1 day (~24 hours).

The `reconnectBackOffBaseSeconds` is increased by a bit over 10% (from
3.0 to 3.4 seconds) so this would not affect reconnection retries too
much.

### BUG FIXES

- `[mocks]` Mockery `v2.49.0` broke the mocks. We had to add a `.mockery.yaml` to
properly handle this change.
  ([\#4521](https://github.com/cometbft/cometbft/pull/4521))

### IMPROVEMENTS

- `[p2p]` fix exponential backoff logic to increase reconnect retries close to 24 hours
 ([\#3519](https://github.com/cometbft/cometbft/issues/3519))

## v0.37.13

*October 31, 2024*

This release rollbacks cometbft-db version to v0.9.5 due to the breaking change
introduced in v0.13.0. Users wishing to use the latest version of cometbft-db
are advised to upgrade to CometBFT v0.38.

### DEPENDENCIES

- `[deps]` Rollback cometbft-db version to v0.9.5 due to the breaking change
  that was introduced in v0.13.0
  ([\#4369](https://github.com/cometbft/cometbft/pull/4369)).

## v0.37.12

*October 24, 2024*

This patch release addresses the issue where tx_search was not returning all results, which only arises when upgrading
to CometBFT-DB version 0.13 or later. It includes a fix in the state indexer to resolve this problem. We recommend
upgrading to this patch release if you are affected by this issue.

### BUG FIXES

- `[state/indexer]` Fix the tx_search results not returning all results by changing the logic in the indexer to copy the key and values instead of reusing an iterator. This issue only arises when upgrading to cometbft-db v0.13 or later.
  ([\#4295](https://github.com/cometbft/cometbft/issues/4295)). Special thanks to @faddat for reporting the issue.

### DEPENDENCIES

- `[go/runtime]` Bump Go version to 1.22
  ([\#4072](https://github.com/cometbft/cometbft/pull/4072))
- Bump cometbft-db version to v0.14.1
  ([\#4326](https://github.com/cometbft/cometbft/pull/4326))

### FEATURES

- `[crypto]` use decred secp256k1 directly ([#4329](https://github.com/cometbft/cometbft/pull/4329))

## v0.37.11

*September 3, 2024*

This release includes a security fix for the light client and is recommended
for all users.

### BUG FIXES

- `[light]` Cross-check proposer priorities in retrieved validator sets
  ([\#ASA-2024-009](https://github.com/cometbft/cometbft/security/advisories/GHSA-g5xx-c4hv-9ccc))
- `[privval]` Retry accepting a connection ([\#2047](https://github.com/cometbft/cometbft/pull/2047))
- `[privval]` Ignore duplicate privval listen when already connected ([\#3828](https://github.com/cometbft/cometbft/issues/3828)

### DEPENDENCIES

- `[crypto/secp256k1]` Adjust to breaking interface changes in
  `btcec/v2` latest release, while avoiding breaking changes to
  local CometBFT functions
  ([\#3728](https://github.com/cometbft/cometbft/pull/3728))
- pinned mockery's version to v2.49.2 to prevent potential
  changes in mocks after each new release of mockery
  ([\#4605](https://github.com/cometbft/cometbft/pull/4605))

### IMPROVEMENTS

- `[types]` Check that proposer is one of the validators in `ValidateBasic`
  ([\#ASA-2024-009](https://github.com/cometbft/cometbft/security/advisories/GHSA-g5xx-c4hv-9ccc))

## v0.37.10

*August 12, 2024*

This release contains a few minor bug fixes and performance improvements.

### IMPROVEMENTS

- `[indexer]` Fixed ineffective select break statements; they now
  point to their enclosing for loop label to exit
  ([\#3544](https://github.com/cometbft/cometbft/issues/3544))

## v0.37.9

*July 16, 2024*

This release contains a few minor bug fixes and performance improvements.

### BUG FIXES

- `[p2p]` Node respects configured `max_num_outbound_peers` limit when dialing
  peers provided by a seed node
  ([\#486](https://github.com/cometbft/cometbft/issues/486))
- `[blocksync]` Do not stay in blocksync if the node's validator voting power
  is high enough to block the chain while it is not online
  ([\#3406](https://github.com/cometbft/cometbft/pull/3406))

### IMPROVEMENTS

- `[p2p/conn]` Remove the usage of a synchronous pool of buffers in secret connection, storing instead the buffer in the connection struct. This reduces the synchronization primitive usage, speeding up the code.
  ([\#3403](https://github.com/cometbft/cometbft/issues/3403))

## v0.37.8

*July 1, 2024*

This release reverts the API-breaking change to the `Mempool` interface introduced in the last patch
release (v0.37.7) while still keeping the performance improvement added to the mempool. It also
includes a minor fix to the RPC endpoints `/tx` and `/tx_search`.

### BREAKING CHANGES

- `[mempool]` Revert adding the method `PreUpdate()` to the `Mempool` interface, recently introduced
  in the previous patch release (`v0.37.7`). Its logic is now moved into the `Lock` method. With this change,
  the `Mempool` interface is the same as in `v0.37.6`.
  ([\#3363](https://github.com/cometbft/cometbft/pull/3363))

### BUG FIXES

- `[rpc]` Fix nil pointer error in `/tx` and `/tx_search` when block is
  absent ([\#3352](https://github.com/cometbft/cometbft/issues/3352))

## v0.37.7

*June 27, 2024*

This release is centered around two topics: mempool and performance improvements. The mempool will stop accepting new
transactions if the node can't keep up with rechecking the number of transactions already in the mempool.
It also contains a few bug fixes.

### BREAKING CHANGES

- `[mempool]` Add to the `Mempool` interface a new method `PreUpdate()`. This method should be
  called before acquiring the mempool lock, to signal that a new update is coming. Also add to
  `ErrMempoolIsFull` a new field `RecheckFull`.
  ([\#3314](https://github.com/cometbft/cometbft/pull/3314))

### BUG FIXES

- `[blockstore]` Fix invalid blocks received in blocksync mode, added banning peer option
  ([\#ASA-2024-008](https://github.com/cometbft/cometbft/security/advisories/GHSA-hg58-rf2h-6rr7))
- `[consensus]` Fix a race condition in the consensus timeout ticker. Race is caused by two timeouts being scheduled at the same time.
  ([\#3092](https://github.com/cometbft/cometbft/pull/2136))

### IMPROVEMENTS

- `[event-bus]` Remove the debug logs in PublishEventTx, which were noticed production slowdowns.
  ([\#2911](https://github.com/cometbft/cometbft/pull/2911))
- `[state/execution]` Cache the block hash computation inside of the Block Type, so we only compute it once.
  ([\#2924](https://github.com/cometbft/cometbft/pull/2924))
- `[consensus/state]` Remove a redundant `VerifyBlock` call in `FinalizeCommit`
  ([\#2928](https://github.com/cometbft/cometbft/pull/2928))
- `[p2p/channel]` Speedup `ProtoIO` writer creation time, and thereby speedup channel writing by 5%.
  ([\#2949](https://github.com/cometbft/cometbft/pull/2949))
- `[p2p/conn]` Minor speedup (3%) to connection.WritePacketMsgTo, by removing MinInt calls.
  ([\#2952](https://github.com/cometbft/cometbft/pull/2952))
- `[internal/bits]` 10x speedup creating initialized bitArrays, which speedsup extendedCommit.BitArray(). This is used in consensus vote gossip.
  ([\#2959](https://github.com/cometbft/cometbft/pull/2841)).
- `[blockstore]` Remove a redundant `Header.ValidateBasic` call in `LoadBlockMeta`, 75% reducing this time.
  ([\#2964](https://github.com/cometbft/cometbft/pull/2964))
- `[p2p/conn]` Speedup connection.WritePacketMsgTo, by reusing internal buffers rather than re-allocating.
  ([\#2986](https://github.com/cometbft/cometbft/pull/2986))
- [`blockstore`] Use LRU caches in blockstore, significiantly improving consensus gossip routine performance
  ([\#3003](https://github.com/cometbft/cometbft/issues/3003)
- [`consensus`] Improve performance of consensus metrics by lowering string operations
  ([\#3017](https://github.com/cometbft/cometbft/issues/3017)
- [`protoio`] Remove one allocation and new object call from `ReadMsg`,
  leading to a 4% p2p message reading performance gain.
  ([\#3018](https://github.com/cometbft/cometbft/issues/3018)
- `[mempool]` Before updating the mempool, consider it as full if rechecking is still in progress.
  This will stop accepting transactions in the mempool if the node can't keep up with re-CheckTx.
  This improvement is implemented only in the v0 mempool.
  ([\#3314](https://github.com/cometbft/cometbft/pull/3314))

## v0.37.6

*April 26, 2024*

This release contains a few bug fixes and performance improvements. It also
bumps Go version to 1.21.

### BUG FIXES

- `[state]` Fix rollback to a specific height
  ([\#2136](https://github.com/cometbft/cometbft/pull/2136))
- [`bits`] prevent `BitArray.UnmarshalJSON` from crashing on 0 bits
  ([\#2774](https://github.com/cometbft/cometbft/pull/2774))

### DEPENDENCIES

- Bump Go version used to v1.21 since v1.20 has reached EOL
  ([\#2817](https://github.com/cometbft/cometbft/pull/2817))

### IMPROVEMENTS

- `[state/indexer]` Lower the heap allocation of transaction searches
  ([\#2839](https://github.com/cometbft/cometbft/pull/2839))
- `[internal/bits]` 10x speedup and remove heap overhead of bitArray.PickRandom (used extensively in consensus gossip)
  ([\#2841](https://github.com/cometbft/cometbft/pull/2841)).
- `[libs/json]` Lower the memory overhead of JSON encoding by using JSON encoders internally
  ([\#2846](https://github.com/cometbft/cometbft/pull/2846)).

## v0.37.5

*March 12, 2024*

This release fixes a security bug in the light client. It also introduces many
improvements to the block sync in collaboration with the
[Osmosis](https://osmosis.zone/) team.

### BUG FIXES

- `[mempool]` The calculation method of tx size returned by calling proxyapp should be consistent with that of mempool
  ([\#1687](https://github.com/cometbft/cometbft/pull/1687))
- `[evidence]` When `VerifyCommitLight` & `VerifyCommitLightTrusting` are called as part
  of evidence verification, all signatures present in the evidence must be verified
  ([\#1749](https://github.com/cometbft/cometbft/pull/1749))

### IMPROVEMENTS

- `[types]` Validate `Validator#Address` in `ValidateBasic` ([\#1715](https://github.com/cometbft/cometbft/pull/1715))
- `[abci]` Increase ABCI socket message size limit to 2GB ([\#1730](https://github.com/cometbft/cometbft/pull/1730): @troykessler)
- `[blocksync]` make the max number of downloaded blocks dynamic.
  Previously it was a const 600. Now it's `peersCount * maxPendingRequestsPerPeer (20)`
  [\#2467](https://github.com/cometbft/cometbft/pull/2467)
- `[blocksync]` Request a block from peer B if we are approaching pool's height
  (less than 50 blocks) and the current peer A is slow in sending us the
  block [\#2475](https://github.com/cometbft/cometbft/pull/2475)
- `[blocksync]` Request the block N from peer B immediately after getting
  `NoBlockResponse` from peer A
  [\#2475](https://github.com/cometbft/cometbft/pull/2475)
- `[blocksync]` Sort peers by download rate (the fastest peer is picked first)
  [\#2475](https://github.com/cometbft/cometbft/pull/2475)

## v0.37.4

*November 27, 2023*

This release provides the **nop** mempool for applications that want to build
their own mempool. Using this mempool effectively disables all mempool
functionality in CometBFT, including transaction dissemination and the
`broadcast_tx_*` endpoints.

Also fixes a small bug in the mempool for an experimental feature, and reverts
the change from v0.37.3 that bumped the minimum Go version to v1.21.

### BUG FIXES

- `[mempool]` Avoid infinite wait in transaction sending routine when
  using experimental parameters to limiting transaction gossiping to peers
  ([\#1654](https://github.com/cometbft/cometbft/pull/1654))

### FEATURES

- `[mempool]` Add `nop` mempool ([\#1643](https://github.com/cometbft/cometbft/pull/1643))

  If you want to use it, change mempool's `type` to `nop`:

  ```toml
  [mempool]

  # The type of mempool for this node to use.
  #
  # Possible types:
  # - "flood" : concurrent linked list mempool with flooding gossip protocol
  # (default)
  # - "nop"   : nop-mempool (short for no operation; the ABCI app is responsible
  # for storing, disseminating and proposing txs). "create_empty_blocks=false"
  # is not supported.
  type = "nop"
  ```

## v0.37.3

*November 17, 2023*

This release contains, among other things, an opt-in, experimental feature to
help reduce the bandwidth consumption associated with the mempool's transaction
gossip.

### BREAKING CHANGES

- `[p2p]` Remove unused UPnP functionality
  ([\#1113](https://github.com/cometbft/cometbft/issues/1113))

### BUG FIXES

- `[state/indexer]` Respect both height params while querying for events
   ([\#1529](https://github.com/cometbft/cometbft/pull/1529))

### FEATURES

- `[node/state]` Add Go API to bootstrap block store and state store to a height
  ([\#1057](https://github.com/tendermint/tendermint/pull/#1057)) (@yihuang)
- `[metrics]` Add metric for mempool size in bytes `SizeBytes`.
  ([\#1512](https://github.com/cometbft/cometbft/pull/1512))

### IMPROVEMENTS

- `[crypto/sr25519]` Upgrade to go-schnorrkel@v1.0.0 ([\#475](https://github.com/cometbft/cometbft/issues/475))
- `[node]` Make handshake cancelable ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
- `[node]` Close evidence.db OnStop ([cometbft/cometbft\#1210](https://github.com/cometbft/cometbft/pull/1210): @chillyvee)
- `[mempool]` Add experimental feature to limit the number of persistent peers and non-persistent
  peers to which the node gossip transactions (only for "v0" mempool).
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[config]` Add mempool parameters `experimental_max_gossip_connections_to_persistent_peers` and
  `experimental_max_gossip_connections_to_non_persistent_peers` for limiting the number of peers to
  which the node gossip transactions. 
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))

## v0.37.2

*June 14, 2023*

Provides several minor bug fixes, as well as fixes for several low-severity
security issues.

### BUG FIXES

- `[state/kvindex]` Querying event attributes that are bigger than int64 is now
  enabled. We are not supporting reading floats from the db into the indexer
  nor parsing them into BigFloats to not introduce breaking changes in minor
  releases. ([\#771](https://github.com/cometbft/cometbft/pull/771))
- `[pubsub]` Pubsub queries are now able to parse big integers (larger than
  int64). Very big floats are also properly parsed into very big integers
  instead of being truncated to int64.
  ([\#771](https://github.com/cometbft/cometbft/pull/771))

### IMPROVEMENTS

- `[rpc]` Remove response data from response failure logs in order
  to prevent large quantities of log data from being produced
  ([\#654](https://github.com/cometbft/cometbft/issues/654))

### SECURITY FIXES

- `[rpc/jsonrpc/client]` **Low severity** - Prevent RPC
  client credentials from being inadvertently dumped to logs
  ([\#787](https://github.com/cometbft/cometbft/pull/787))
- `[cmd/cometbft/commands/debug/kill]` **Low severity** - Fix unsafe int cast in
  `debug kill` command ([\#793](https://github.com/cometbft/cometbft/pull/793))
- `[consensus]` **Low severity** - Avoid recursive call after rename to
  `(*PeerState).MarshalJSON`
  ([\#863](https://github.com/cometbft/cometbft/pull/863))
- `[mempool/clist_mempool]` **Low severity** - Prevent a transaction from
  appearing twice in the mempool
  ([\#890](https://github.com/cometbft/cometbft/pull/890): @otrack)

## v0.37.1

*April 26, 2023*

This release fixes several bugs, and has had to introduce one small Go
API-breaking change in the `crypto/merkle` package in order to address what
could be a security issue for some users who directly and explicitly make use of
that code.

### BREAKING CHANGES

- `[crypto/merkle]` Do not allow verification of Merkle Proofs against empty trees (`nil` root). `Proof.ComputeRootHash` now panics when it encounters an error, but `Proof.Verify` does not panic
  ([\#558](https://github.com/cometbft/cometbft/issues/558))

### BUG FIXES

- `[consensus]` Unexpected error conditions in `ApplyBlock` are non-recoverable, so ignoring the error and carrying on is a bug. We replaced a `return` that disregarded the error by a `panic`.
  ([\#496](https://github.com/cometbft/cometbft/pull/496))
- `[consensus]` Rename `(*PeerState).ToJSON` to `MarshalJSON` to fix a logging data race
  ([\#524](https://github.com/cometbft/cometbft/pull/524))
- `[light]` Fixed an edge case where a light client would panic when attempting
  to query a node that (1) has started from a non-zero height and (2) does
  not yet have any data. The light client will now, correctly, not panic
  _and_ keep the node in its list of providers in the same way it would if
  it queried a node starting from height zero that does not yet have data
  ([\#575](https://github.com/cometbft/cometbft/issues/575))

### IMPROVEMENTS

- `[jsonrpc/client]` Improve the error message for client errors stemming from
  bad HTTP responses.
  ([cometbft/cometbft\#638](https://github.com/cometbft/cometbft/pull/638))

## v0.37.0

*March 6, 2023*

This is the first CometBFT release with ABCI 1.0, which introduces the
`PrepareProposal` and `ProcessProposal` methods, with the aim of expanding the
range of use cases that application developers can address. This is the first
change to ABCI towards ABCI++, and the full range of ABCI++ functionality will
only become available in the next major release with ABCI 2.0. See the
[specification](./spec/abci/) for more details.

In the v0.34.27 release, the CometBFT Go module is still
`github.com/tendermint/tendermint` to facilitate ease of upgrading for users,
but in this release we have changed this to `github.com/cometbft/cometbft`.

Please also see our [upgrading guidelines](./UPGRADING.md) for more details on
upgrading from the v0.34 release series.

Also see our [QA results](https://docs.cometbft.com/v0.37/qa/v037/cometbft) for
the v0.37 release.

We'd love your feedback on this release! Please reach out to us via one of our
communication channels, such as [GitHub
Discussions](https://github.com/cometbft/cometbft/discussions), with any of your
questions, comments and/or concerns.

See below for more details.

### BREAKING CHANGES

- The `TMHOME` environment variable was renamed to `CMTHOME`, and all environment variables starting with `TM_` are instead prefixed with `CMT_`
  ([\#211](https://github.com/cometbft/cometbft/issues/211))
- `[p2p]` Reactor `Send`, `TrySend` and `Receive` renamed to `SendEnvelope`,
  `TrySendEnvelope` and `ReceiveEnvelope` to allow metrics to be appended to
  messages and measure bytes sent/received.
  ([\#230](https://github.com/cometbft/cometbft/pull/230))
- Bump minimum Go version to 1.20
  ([\#385](https://github.com/cometbft/cometbft/issues/385))
- [config] The boolean key `fastsync` is deprecated and replaced by
    `block_sync`. ([\#9259](https://github.com/tendermint/tendermint/pull/9259))
    At the same time, `block_sync` is also deprecated. In the next release,
    BlocSync will always be enabled and `block_sync` will be removed.
    ([\#409](https://github.com/cometbft/cometbft/issues/409))
- `[abci]` Make length delimiter encoding consistent
  (`uint64`) between ABCI and P2P wire-level protocols
  ([\#5783](https://github.com/tendermint/tendermint/pull/5783))
- `[abci]` Change the `key` and `value` fields from
  `[]byte` to `string` in the `EventAttribute` type.
  ([\#6403](https://github.com/tendermint/tendermint/pull/6403))
- `[abci/counter]` Delete counter example app
  ([\#6684](https://github.com/tendermint/tendermint/pull/6684))
- `[abci]` Renamed `EvidenceType` to `MisbehaviorType` and `Evidence`
  to `Misbehavior` as a more accurate label of their contents.
  ([\#8216](https://github.com/tendermint/tendermint/pull/8216))
- `[abci]` Added cli commands for `PrepareProposal` and `ProcessProposal`.
  ([\#8656](https://github.com/tendermint/tendermint/pull/8656))
- `[abci]` Added cli commands for `PrepareProposal` and `ProcessProposal`.
  ([\#8901](https://github.com/tendermint/tendermint/pull/8901))
- `[abci]` Renamed `LastCommitInfo` to `CommitInfo` in preparation for vote
  extensions. ([\#9122](https://github.com/tendermint/tendermint/pull/9122))
- Change spelling from British English to American. Rename
  `Subscription.Cancelled()` to `Subscription.Canceled()` in `libs/pubsub`
  ([\#9144](https://github.com/tendermint/tendermint/pull/9144))
- `[abci]` Removes unused Response/Request `SetOption` from ABCI
  ([\#9145](https://github.com/tendermint/tendermint/pull/9145))
- `[config]` Rename the fastsync section and the
  fast\_sync key blocksync and block\_sync respectively
  ([\#9259](https://github.com/tendermint/tendermint/pull/9259))
- `[types]` Reduce the use of protobuf types in core logic. `ConsensusParams`,
  `BlockParams`, `ValidatorParams`, `EvidenceParams`, `VersionParams` have
  become native types.  They still utilize protobuf when being sent over
  the wire or written to disk.  Moved `ValidateConsensusParams` inside
  (now native type) `ConsensusParams`, and renamed it to `ValidateBasic`.
  ([\#9287](https://github.com/tendermint/tendermint/pull/9287))
- `[abci/params]` Deduplicate `ConsensusParams` and `BlockParams` so
  only `types` proto definitions are use. Remove `TimeIotaMs` and use
  a hard-coded 1 millisecond value to ensure monotonically increasing
  block times. Rename `AppVersion` to `App` so as to not stutter.
  ([\#9287](https://github.com/tendermint/tendermint/pull/9287))
- `[abci]` New ABCI methods `PrepareProposal` and `ProcessProposal` which give
  the app control over transactions proposed and allows for verification of
  proposed blocks. ([\#9301](https://github.com/tendermint/tendermint/pull/9301))

### BUG FIXES

- `[consensus]` Fixed a busy loop that happened when sending of a block part failed by sleeping in case of error.
  ([\#4](https://github.com/informalsystems/tendermint/pull/4))
- `[state/kvindexer]` Fixed the default behaviour of the kvindexer to index and
  query attributes by events in which they occur. In 0.34.25 this was mitigated
  by a separated RPC flag. @jmalicevic
  ([\#77](https://github.com/cometbft/cometbft/pull/77))
- `[state/kvindexer]` Resolved crashes when event values contained slashes,
  introduced after adding event sequences in
  [\#77](https://github.com/cometbft/cometbft/pull/77). @jmalicevic
  ([\#382](https://github.com/cometbft/cometbft/pull/382))
- `[consensus]` ([\#386](https://github.com/cometbft/cometbft/pull/386)) Short-term fix for the case when `needProofBlock` cannot find previous block meta by defaulting to the creation of a new proof block. (@adizere)
  - Special thanks to the [Vega.xyz](https://vega.xyz/) team, and in particular to Zohar (@ze97286), for reporting the problem and working with us to get to a fix.
- `[docker]` enable cross platform build using docker buildx
  ([\#9073](https://github.com/tendermint/tendermint/pull/9073))
- `[consensus]` fix round number of `enterPropose`
  when handling `RoundStepNewRound` timeout.
  ([\#9229](https://github.com/tendermint/tendermint/pull/9229))
- `[docker]` ensure Docker image uses consistent version of Go
  ([\#9462](https://github.com/tendermint/tendermint/pull/9462))
- `[p2p]` prevent peers who have errored from being added to `peer_set`
  ([\#9500](https://github.com/tendermint/tendermint/pull/9500))
- `[blocksync]` handle the case when the sending
  queue is full: retry block request after a timeout
  ([\#9518](https://github.com/tendermint/tendermint/pull/9518))

### FEATURES

- `[abci]` New ABCI methods `PrepareProposal` and `ProcessProposal` which give
  the app control over transactions proposed and allows for verification of
  proposed blocks. ([\#9301](https://github.com/tendermint/tendermint/pull/9301))

### IMPROVEMENTS

- `[e2e]` Add functionality for uncoordinated (minor) upgrades
  ([\#56](https://github.com/tendermint/tendermint/pull/56))
- `[tools/tm-signer-harness]` Remove the folder as it is unused
  ([\#136](https://github.com/cometbft/cometbft/issues/136))
- `[p2p]` Reactor `Send`, `TrySend` and `Receive` renamed to `SendEnvelope`,
  `TrySendEnvelope` and `ReceiveEnvelope` to allow metrics to be appended to
  messages and measure bytes sent/received.
  ([\#230](https://github.com/cometbft/cometbft/pull/230))
- `[abci]` Added `AbciVersion` to `RequestInfo` allowing
  applications to check ABCI version when connecting to CometBFT.
  ([\#5706](https://github.com/tendermint/tendermint/pull/5706))
- `[cli]` add `--hard` flag to rollback command (and a boolean to the `RollbackState` method). This will rollback
   state and remove the last block. This command can be triggered multiple times. The application must also rollback
   state to the same height.
  ([\#9171](https://github.com/tendermint/tendermint/pull/9171))
- `[crypto]` Update to use btcec v2 and the latest btcutil.
  ([\#9250](https://github.com/tendermint/tendermint/pull/9250))
- `[rpc]` Added `header` and `header_by_hash` queries to the RPC client
  ([\#9276](https://github.com/tendermint/tendermint/pull/9276))
- `[proto]` Migrate from `gogo/protobuf` to `cosmos/gogoproto`
  ([\#9356](https://github.com/tendermint/tendermint/pull/9356))
- `[rpc]` Enable caching of RPC responses
  ([\#9650](https://github.com/tendermint/tendermint/pull/9650))
- `[consensus]` Save peer LastCommit correctly to achieve 50% reduction in gossiped precommits.
  ([\#9760](https://github.com/tendermint/tendermint/pull/9760))

---

CometBFT is a fork of [Tendermint Core](https://github.com/tendermint/tendermint) as of late December 2022.

## Bug bounty

Friendly reminder, we have a [bug bounty program](https://hackerone.com/cosmos).

## Previous changes

For changes released before the creation of CometBFT, please refer to the Tendermint Core [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/a9feb1c023e172b542c972605311af83b777855b/CHANGELOG.md).

