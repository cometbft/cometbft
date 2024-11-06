# CHANGELOG

## v0.38.15

*November 6, 2024*

This release supersedes [`v0.38.14`](#v03814), which mistakenly updated the Go version to
`1.23`, introducing an unintended breaking change. It sets the Go version back
to `1.22.7` by reverting [\#4297](https://github.com/cometbft/cometbft/pull/4297).

The release includes the bug fixes, performance improvements, and importantly,
the fix for the security vulnerability in the vote extensions (VE) validation
logic that were part of `v0.38.14`. For more details, please refer to [ASA-2024-011](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj).

## v0.38.14

*November 6, 2024*

This release fixes a security vulnerability in the vote extensions (VE)
validation logic. For more details, please refer to
[ASA-2024-011](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj).

We recommend upgrading ASAP if you’re using vote extensions (VE).

### BUG FIXES

- `[consensus]` Do not panic if the validator index of a `Vote` message is out
  of bounds, when vote extensions are enabled
  ([\#ABC-0021](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj))

### DEPENDENCIES

- Bump cometbft-db version to v0.15.0
  ([\#4297](https://github.com/cometbft/cometbft/pull/4297))
- `[go/runtime]` Bump Go version to 1.23
  ([\#4297](https://github.com/cometbft/cometbft/pull/4297))

### IMPROVEMENTS

- `[p2p]` fix exponential backoff logic to increase reconnect retries close to 24 hours
 ([\#3519](https://github.com/cometbft/cometbft/issues/3519))

## v0.38.13

*October 24, 2024*

This patch release addresses the issue where tx_search was not returning all results, which only arises when upgrading
to CometBFT-DB version 0.13 or later. It includes a fix in the state indexer to resolve this problem. We recommend
upgrading to this patch release if you are affected by this issue.

### BUG FIXES

- `[metrics]` Call unused `rejected_txs` metric in mempool
  ([\#4019](https://github.com/cometbft/cometbft/pull/4019))
- `[state/indexer]` Fix the tx_search results not returning all results by changing the logic in the indexer to copy the key and values instead of reusing an iterator. This issue only arises when upgrading to cometbft-db v0.13 or later.
  ([\#4295](https://github.com/cometbft/cometbft/issues/4295)). Special thanks to @faddat for reporting the issue.

### DEPENDENCIES

- `[go/runtime]` Bump Go version to 1.22
  ([\#4073](https://github.com/cometbft/cometbft/pull/4073))
- Bump cometbft-db version to v0.14.1
  ([\#4321](https://github.com/cometbft/cometbft/pull/4321))

### FEATURES

- `[crypto]` use decred secp256k1 directly ([#4294](https://github.com/cometbft/cometbft/pull/4294))

### IMPROVEMENTS

- `[metrics]` Add `evicted_txs` metric to mempool
  ([\#4019](https://github.com/cometbft/cometbft/pull/4019))
- `[log]` Change "mempool is full" log to debug level
  ([\#4123](https://github.com/cometbft/cometbft/pull/4123)) Special thanks to @yihuang.

## v0.38.12

*September 3, 2024*

This release includes a security fix for the light client and is recommended
for all users.

### BUG FIXES

- `[light]` Cross-check proposer priorities in retrieved validator sets
  ([\#ASA-2024-009](https://github.com/cometbft/cometbft/security/advisories/GHSA-g5xx-c4hv-9ccc))
- `[privval]` Ignore duplicate privval listen when already connected ([\#3828](https://github.com/cometbft/cometbft/issues/3828)

### DEPENDENCIES

- `[crypto/secp256k1]` Adjust to breaking interface changes in
  `btcec/v2` latest release, while avoiding breaking changes to
  local CometBFT functions
  ([\#3728](https://github.com/cometbft/cometbft/pull/3728))

### IMPROVEMENTS

- `[types]` Check that proposer is one of the validators in `ValidateBasic`
  ([\#ASA-2024-009](https://github.com/cometbft/cometbft/security/advisories/GHSA-g5xx-c4hv-9ccc))
- `[e2e]` Add `log_level` option to manifest file
  ([#3819](https://github.com/cometbft/cometbft/pull/3819)).
- `[e2e]` Add `log_format` option to manifest file
  ([#3836](https://github.com/cometbft/cometbft/issues/3836)).

## v0.38.11

*August 12, 2024*

This release fixes a panic in consensus where CometBFT would previously panic
if there's no extension signature in non-nil Precommit EVEN IF vote extensions
themselves are disabled.

It also includes a few other bug fixes and performance improvements.

### BUG FIXES

- `[types]` Only check IFF vote is a non-nil Precommit if extensionsEnabled
  types ([\#3565](https://github.com/cometbft/cometbft/issues/3565))

### IMPROVEMENTS

- `[indexer]` Fixed ineffective select break statements; they now
  point to their enclosing for loop label to exit
  ([\#3544](https://github.com/cometbft/cometbft/issues/3544))

## v0.38.10

*July 16, 2024*

This release fixes a bug in `v0.38.x` that prevented ABCI responses from being
correctly read when upgrading from `v0.37.x` or below. It also includes a few other
bug fixes and performance improvements.

### BUG FIXES

- `[p2p]` Node respects configured `max_num_outbound_peers` limit when dialing
  peers provided by a seed node
  ([\#486](https://github.com/cometbft/cometbft/issues/486))
- `[rpc]` Fix an issue where a legacy ABCI response, created on `v0.37` or before, is not returned properly in `v0.38` and up
  on the `/block_results` RPC endpoint.
  ([\#3002](https://github.com/cometbft/cometbft/issues/3002))
- `[blocksync]` Do not stay in blocksync if the node's validator voting power
  is high enough to block the chain while it is not online
  ([\#3406](https://github.com/cometbft/cometbft/pull/3406))

### IMPROVEMENTS

- `[p2p/conn]` Update send monitor, used for sending rate limiting, once per batch of packets sent
  ([\#3382](https://github.com/cometbft/cometbft/pull/3382))
- `[libs/pubsub]` Allow dash (`-`) in event tags
  ([\#3401](https://github.com/cometbft/cometbft/issues/3401))
- `[p2p/conn]` Remove the usage of a synchronous pool of buffers in secret connection, storing instead the buffer in the connection struct. This reduces the synchronization primitive usage, speeding up the code.
  ([\#3403](https://github.com/cometbft/cometbft/issues/3403))

## v0.38.9

*July 1, 2024*

This release reverts the API-breaking change to the Mempool interface introduced in the last patch
release (v0.38.8) while still keeping the performance improvement added to the mempool. It also
includes a minor fix to the RPC endpoints /tx and /tx_search.

### BREAKING CHANGES

- `[mempool]` Revert adding the method `PreUpdate()` to the `Mempool` interface, recently introduced
  in the previous patch release (v0.38.8). Its logic is now moved into the `Lock` method. With this change,
  the `Mempool` interface is the same as in v0.38.7.
  ([\#3361](https://github.com/cometbft/cometbft/pull/3361))

### BUG FIXES

- `[rpc]` Fix nil pointer error in `/tx` and `/tx_search` when block is
  absent ([\#3352](https://github.com/cometbft/cometbft/issues/3352))

## v0.38.8

*June 27, 2024*

This release contains a few bug fixes and performance improvements.

### BREAKING CHANGES

- `[mempool]` Add to the `Mempool` interface a new method `PreUpdate()`. This method should be
  called before acquiring the mempool lock, to signal that a new update is coming. Also add to
  `ErrMempoolIsFull` a new field `RecheckFull`.
  ([\#3314](https://github.com/cometbft/cometbft/pull/3314))

### BUG FIXES

- `[blockstore]` Added peer banning in blockstore
  ([\#ABC-0013](https://github.com/cometbft/cometbft/security/advisories/GHSA-hg58-rf2h-6rr7))
- `[blockstore]` Send correct error message when vote extensions do not align with received packet
  ([\#ABC-0014](https://github.com/cometbft/cometbft/security/advisories/GHSA-hg58-rf2h-6rr7))
- [`mempool`] Fix data race when rechecking with async ABCI client
  ([\#1827](https://github.com/cometbft/cometbft/issues/1827))
- `[consensus]` Fix a race condition in the consensus timeout ticker. Race is caused by two timeouts being scheduled at the same time.
  ([\#3092](https://github.com/cometbft/cometbft/pull/2136))
- `[types]` Do not batch verify a commit if the validator set keys have different
  types. ([\#3195](https://github.com/cometbft/cometbft/issues/3195)

### IMPROVEMENTS

- `[config]` Added `recheck_timeout` mempool parameter to set how much time to wait for recheck
  responses from the app (only applies to non-local ABCI clients).
  ([\#1827](https://github.com/cometbft/cometbft/issues/1827/))
- `[rpc]` Add a configurable maximum batch size for RPC requests.
  ([\#2867](https://github.com/cometbft/cometbft/pull/2867)).
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
  ([\#3314](https://github.com/cometbft/cometbft/pull/3314))

## v0.38.7

*April 26, 2024*

This release contains a few bug fixes and performance improvements.

### BUG FIXES

- [`mempool`] Panic when a CheckTx request to the app returns an error
  ([\#2225](https://github.com/cometbft/cometbft/pull/2225))
- [`bits`] prevent `BitArray.UnmarshalJSON` from crashing on 0 bits
  ([\#2774](https://github.com/cometbft/cometbft/pull/2774))

### FEATURES

- [`node`] Add `BootstrapStateWithGenProvider` to boostrap state using a custom
  genesis doc provider ([\#2793](https://github.com/cometbft/cometbft/pull/2793))

### IMPROVEMENTS

- `[state/indexer]` Lower the heap allocation of transaction searches
  ([\#2839](https://github.com/cometbft/cometbft/pull/2839))
- `[internal/bits]` 10x speedup and remove heap overhead of bitArray.PickRandom (used extensively in consensus gossip)
  ([\#2841](https://github.com/cometbft/cometbft/pull/2841)).
- `[libs/json]` Lower the memory overhead of JSON encoding by using JSON encoders internally
  ([\#2846](https://github.com/cometbft/cometbft/pull/2846)).

## v0.38.6

*March 12, 2024*

This release fixes a security bug in the light client. It also introduces many
improvements to the block sync in collaboration with the
[Osmosis](https://osmosis.zone/) team.

### BUG FIXES

- `[privval]` Retry accepting a connection ([\#2047](https://github.com/cometbft/cometbft/pull/2047))
- `[state]` Fix rollback to a specific height
  ([\#2136](https://github.com/cometbft/cometbft/pull/2136))

### FEATURES

- `[e2e]` Add `block_max_bytes` option to the manifest file.
  ([\#2362](https://github.com/cometbft/cometbft/pull/2362))

### IMPROVEMENTS

- `[blocksync]` Avoid double-calling `types.BlockFromProto` for performance
  reasons ([\#2016](https://github.com/cometbft/cometbft/pull/2016))
- `[e2e]` Add manifest option `load_max_txs` to limit the number of transactions generated by the
  `load` command. ([\#2094](https://github.com/cometbft/cometbft/pull/2094))
- `[jsonrpc]` enable HTTP basic auth in websocket client ([#2434](https://github.com/cometbft/cometbft/pull/2434))
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

## v0.38.5

*January 24, 2024*

This release fixes a problem introduced in `v0.38.3`: if an application
updates the value of ConsensusParam `VoteExtensionsEnableHeight` to the same value
(actually a "noop" update) this is accepted in `v0.38.2` but rejected under some
conditions in `v0.38.3` and `v0.38.4`. Even if rejecting a useless update would make sense
in general, in a point release we should not reject a set of inputs to
a function that was previuosly accepted (unless there is a good reason
for it). The goal of this release is to accept again all "noop" updates, like `v0.38.2` did.

### IMPROVEMENTS

- `[consensus]` Add `chain_size_bytes` metric for measuring the size of the blockchain in bytes
  ([\#2093](https://github.com/cometbft/cometbft/pull/2093))

## v0.38.4

*January 22, 2024*

This release is aimed at those projects that have a dependency on CometBFT,
release line `v0.38.x`, and make use of function `SaveBlockStoreState` in package
`github.com/cometbft/cometbft/store`. This function changed its signature in `v0.38.3`.
This new release reverts the signature change so that upgrading to the latest release
of CometBFT on `v0.38.x` does not require any change in the code depending on CometBFT.

### IMPROVEMENTS

- `[e2e]` Add manifest option `VoteExtensionsUpdateHeight` to test
  vote extension activation via `InitChain` and `FinalizeBlock`.
  Also, extend the manifest generator to produce different values
  of this new option
  ([\#2065](https://github.com/cometbft/cometbft/pull/2065))

## v0.38.3

*January 17, 2024*

This release addresses a high impact security issue reported in advisory
([ASA-2024-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-qr8r-m495-7hc4)).
There are other non-security bugs fixes that have been addressed since
`v0.38.2` was released, as well as some improvements.
Please check the list below for further details.

### BUG FIXES

- `[consensus]` Fix for "Validation of `VoteExtensionsEnableHeight` can cause chain halt"
  ([ASA-2024-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-qr8r-m495-7hc4))
- `[mempool]` Fix data races in `CListMempool` by making atomic the types of `height`, `txsBytes`, and
  `notifiedTxsAvailable`. ([\#642](https://github.com/cometbft/cometbft/pull/642))
- `[mempool]` The calculation method of tx size returned by calling proxyapp should be consistent with that of mempool
  ([\#1687](https://github.com/cometbft/cometbft/pull/1687))
- `[evidence]` When `VerifyCommitLight` & `VerifyCommitLightTrusting` are called as part
  of evidence verification, all signatures present in the evidence must be verified
  ([\#1749](https://github.com/cometbft/cometbft/pull/1749))
- `[crypto]` `SupportsBatchVerifier` returns false
  if public key is nil instead of dereferencing nil.
  ([\#1825](https://github.com/cometbft/cometbft/pull/1825))
- `[blocksync]` wait for `poolRoutine` to stop in `(*Reactor).OnStop`
  ([\#1879](https://github.com/cometbft/cometbft/pull/1879))

### IMPROVEMENTS

- `[types]` Validate `Validator#Address` in `ValidateBasic` ([\#1715](https://github.com/cometbft/cometbft/pull/1715))
- `[abci]` Increase ABCI socket message size limit to 2GB ([\#1730](https://github.com/cometbft/cometbft/pull/1730): @troykessler)
- `[state]` Save the state using a single DB batch ([\#1735](https://github.com/cometbft/cometbft/pull/1735))
- `[store]` Save block using a single DB batch if block is less than 640kB, otherwise each block part is saved individually
  ([\#1755](https://github.com/cometbft/cometbft/pull/1755))
- `[rpc]` Support setting proxy from env to `DefaultHttpClient`.
  ([\#1900](https://github.com/cometbft/cometbft/pull/1900))
- `[rpc]` Use default port for HTTP(S) URLs when there is no explicit port ([\#1903](https://github.com/cometbft/cometbft/pull/1903))
- `[crypto/merkle]` faster calculation of hashes ([#1921](https://github.com/cometbft/cometbft/pull/1921))

## v0.38.2

*November 27, 2023*

This release provides the **nop** mempool for applications that want to build their own mempool.
Using this mempool effectively disables all mempool functionality in CometBFT, including transaction dissemination and the `broadcast_tx_*` endpoints.

Also fixes a small bug in the mempool for an experimental feature.

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

## v0.38.1

*November 17, 2023*

This release contains, among other things, an opt-in, experimental feature to
help reduce the bandwidth consumption associated with the mempool's transaction
gossip.

### BUG FIXES

- `[state/indexer]` Respect both height params while querying for events
   ([\#1529](https://github.com/cometbft/cometbft/pull/1529))

### FEATURES

- `[metrics]` Add metric for mempool size in bytes `SizeBytes`.
  ([\#1512](https://github.com/cometbft/cometbft/pull/1512))

### IMPROVEMENTS

- `[mempool]` Add experimental feature to limit the number of persistent peers and non-persistent
  peers to which the node gossip transactions.
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[config]` Add mempool parameters `experimental_max_gossip_connections_to_persistent_peers` and
  `experimental_max_gossip_connections_to_non_persistent_peers` for limiting the number of peers to
  which the node gossip transactions.
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))

## v0.38.0

*September 12, 2023*

This release includes the second part of ABCI++, called ABCI 2.0.
ABCI 2.0 introduces ABCI methods `ExtendVote` and `VerifyVoteExtension`.
These new methods allow the application to add data (opaque to CometBFT),
called _vote extensions_ to precommit votes sent by validators.
These vote extensions are made available to the proposer(s) of the next height.
Additionally, ABCI 2.0 coalesces `BeginBlock`, `DeliverTx`, and `EndBlock`
into one method, `FinalizeBlock`, whose `Request*` and `Response*`
data structures contain the sum of all data previously contained
in the respective `Request*` and `Response*` data structures in
`BeginBlock`, `DeliverTx`, and `EndBlock`.
See the [specification](./spec/abci/) for more details on ABCI 2.0.

### BREAKING CHANGES

- `[mempool]` Remove priority mempool.
  ([\#260](https://github.com/cometbft/cometbft/issues/260))
- `[config]` Remove `Version` field from `MempoolConfig`.
  ([\#260](https://github.com/cometbft/cometbft/issues/260))
- `[protobuf]` Remove fields `sender`, `priority`, and `mempool_error` from
  `ResponseCheckTx`. ([\#260](https://github.com/cometbft/cometbft/issues/260))
- `[crypto/merkle]` Do not allow verification of Merkle Proofs against empty trees (`nil` root). `Proof.ComputeRootHash` now panics when it encounters an error, but `Proof.Verify` does not panic
  ([\#558](https://github.com/cometbft/cometbft/issues/558))
- `[state/kvindexer]` Remove the function type from the event key stored in the database. This should be breaking only
for people who forked CometBFT and interact directly with the indexers kvstore.
  ([\#774](https://github.com/cometbft/cometbft/pull/774))
- `[rpc]` Removed `begin_block_events` and `end_block_events` from `BlockResultsResponse`.
  The events are merged into one field called `finalize_block_events`.
  ([\#9427](https://github.com/tendermint/tendermint/issues/9427))
- `[pubsub]` Added support for big integers and big floats in the pubsub event query system.
  Breaking changes: function `Number` in package `libs/pubsub/query/syntax` changed its return value.
  ([\#797](https://github.com/cometbft/cometbft/pull/797))
- `[kvindexer]` Added support for big integers and big floats in the kvindexer.
  Breaking changes: function `Number` in package `libs/pubsub/query/syntax` changed its return value.
  ([\#797](https://github.com/cometbft/cometbft/pull/797))
- `[mempool]` Application can now set `ConsensusParams.Block.MaxBytes` to -1
  to have visibility on all transactions in the
  mempool at `PrepareProposal` time.
  This means that the total size of transactions sent via `RequestPrepareProposal`
  might exceed `RequestPrepareProposal.max_tx_bytes`.
  If that is the case, the application MUST make sure that the total size of transactions
  returned in `ResponsePrepareProposal.txs` does not exceed `RequestPrepareProposal.max_tx_bytes`,
  otherwise CometBFT will panic.
  ([\#980](https://github.com/cometbft/cometbft/issues/980))
- `[node/state]` Add Go API to bootstrap block store and state store to a height. Make sure block sync starts syncing from bootstrapped height.
  ([\#1057](https://github.com/tendermint/tendermint/pull/#1057)) (@yihuang)
- `[state/store]` Added Go functions to save height at which offline state sync is performed.
  ([\#1057](https://github.com/tendermint/tendermint/pull/#1057)) (@jmalicevic)
- `[p2p]` Remove UPnP functionality
  ([\#1113](https://github.com/cometbft/cometbft/issues/1113))
- `[node]` Removed `ConsensusState()` accessor from `Node`
  struct - all access to consensus state should go via the reactor
  ([\#1120](https://github.com/cometbft/cometbft/pull/1120))
- `[state]` Signature of `ExtendVote` changed in `BlockExecutor`.
  It now includes the block whose precommit will be extended, an the state object.
  ([\#1270](https://github.com/cometbft/cometbft/pull/1270))
- `[state]` Move pruneBlocks from node/state to state/execution.
  ([\#6541](https://github.com/tendermint/tendermint/pull/6541))
- `[abci]` Move `app_hash` parameter from `Commit` to `FinalizeBlock`
  ([\#8664](https://github.com/tendermint/tendermint/pull/8664))
- `[abci]` Introduce `FinalizeBlock` which condenses `BeginBlock`, `DeliverTx`
  and `EndBlock` into a single method call
  ([\#9468](https://github.com/tendermint/tendermint/pull/9468))
- `[p2p]` Remove unused p2p/trust package
  ([\#9625](https://github.com/tendermint/tendermint/pull/9625))
- `[rpc]` Remove global environment and replace with constructor
  ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[node]` Move DBContext and DBProvider from the node package to the config
  package. ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[inspect]` Add a new `inspect` command for introspecting
  the state and block store of a crashed tendermint node.
  ([\#9655](https://github.com/tendermint/tendermint/pull/9655))
- `[metrics]` Move state-syncing and block-syncing metrics to
  their respective packages. Move labels from block_syncing
  -> blocksync_syncing and state_syncing -> statesync_syncing
  ([\#9682](https://github.com/tendermint/tendermint/pull/9682))

### BUG FIXES

- `[kvindexer]` Forward porting the fixes done to the kvindexer in 0.37 in PR \#77
  ([\#423](https://github.com/cometbft/cometbft/pull/423))
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
- `[abci]` Restore the snake_case naming in JSON serialization of
  `ExecTxResult` ([\#855](https://github.com/cometbft/cometbft/issues/855)).
- `[consensus]` Avoid recursive call after rename to (*PeerState).MarshalJSON
  ([\#863](https://github.com/cometbft/cometbft/pull/863))
- `[mempool/clist_mempool]` Prevent a transaction to appear twice in the mempool
  ([\#890](https://github.com/cometbft/cometbft/pull/890): @otrack)
- `[docker]` Ensure Docker image uses consistent version of Go.
  ([\#9462](https://github.com/tendermint/tendermint/pull/9462))
- `[abci-cli]` Fix broken abci-cli help command.
  ([\#9717](https://github.com/tendermint/tendermint/pull/9717))

### DEPRECATIONS

- `[rpc/grpc]` Mark the gRPC broadcast API as deprecated.
  It will be superseded by a broader API as part of
  [\#81](https://github.com/cometbft/cometbft/issues/81)
  ([\#650](https://github.com/cometbft/cometbft/issues/650))

### FEATURES

- `[node/state]` Add Go API to bootstrap block store and state store to a height
  ([\#1057](https://github.com/tendermint/tendermint/pull/#1057)) (@yihuang)
- `[proxy]` Introduce `NewConnSyncLocalClientCreator`, which allows local ABCI
  clients to have the same concurrency model as remote clients (i.e. one mutex
  per client "connection", for each of the four ABCI "connections").
  ([tendermint/tendermint\#9830](https://github.com/tendermint/tendermint/pull/9830)
  and [\#1145](https://github.com/cometbft/cometbft/pull/1145))
- `[proxy]` Introduce `NewUnsyncLocalClientCreator`, which allows local ABCI
  clients to have the same concurrency model as remote clients (i.e. one
  mutex per client "connection", for each of the four ABCI "connections").
  ([\#9830](https://github.com/tendermint/tendermint/pull/9830))
- `[abci]` New ABCI methods `VerifyVoteExtension` and `ExtendVote` allow validators to validate the vote extension data attached to a pre-commit message and allow applications to let their validators do more than just validate within consensus ([\#9836](https://github.com/tendermint/tendermint/pull/9836))

### IMPROVEMENTS

- `[blocksync]` Generate new metrics during BlockSync
  ([\#543](https://github.com/cometbft/cometbft/pull/543))
- `[jsonrpc/client]` Improve the error message for client errors stemming from
  bad HTTP responses.
  ([cometbft/cometbft\#638](https://github.com/cometbft/cometbft/pull/638))
- `[rpc]` Remove response data from response failure logs in order
  to prevent large quantities of log data from being produced
  ([\#654](https://github.com/cometbft/cometbft/issues/654))
- `[pubsub/kvindexer]` Numeric query conditions and event values are represented as big floats with default precision of 125.
  Integers are read as "big ints" and represented with as many bits as they need when converting to floats.
  ([\#797](https://github.com/cometbft/cometbft/pull/797))
- `[node]` Make handshake cancelable ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
- `[consensus]` New metrics (counters) to track duplicate votes and block parts.
  ([\#896](https://github.com/cometbft/cometbft/pull/896))
- `[mempool]` Application can now set `ConsensusParams.Block.MaxBytes` to -1
  to gain more control on the max size of transactions in a block.
  It also allows the application to have visibility on all transactions in the
  mempool at `PrepareProposal` time.
  ([\#980](https://github.com/cometbft/cometbft/pull/980))
- `[node]` Close evidence.db OnStop ([cometbft/cometbft\#1210](https://github.com/cometbft/cometbft/pull/1210): @chillyvee)
- `[state]` Make logging `block_app_hash` and `app_hash` consistent by logging them both as hex.
  ([\#1264](https://github.com/cometbft/cometbft/pull/1264))
- `[crypto/merkle]` Improve HashAlternatives performance
  ([\#6443](https://github.com/tendermint/tendermint/pull/6443))
- `[p2p/pex]` Improve addrBook.hash performance
  ([\#6509](https://github.com/tendermint/tendermint/pull/6509))
- `[crypto/merkle]` Improve HashAlternatives performance
  ([\#6513](https://github.com/tendermint/tendermint/pull/6513))
- `[pubsub]` Performance improvements for the event query API
  ([\#7319](https://github.com/tendermint/tendermint/pull/7319))

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

## v0.34.27

*Feb 27, 2023*

This is the first official release of CometBFT - a fork of [Tendermint
Core](https://github.com/tendermint/tendermint). This particular release is
intended to be compatible with the Tendermint Core v0.34 release series.

For details as to how to upgrade to CometBFT from Tendermint Core, please see
our [upgrading guidelines](./UPGRADING.md).

If you have any questions, comments, concerns or feedback on this release, we
would love to hear from you! Please contact us via [GitHub
Discussions](https://github.com/cometbft/cometbft/discussions),
[Discord](https://discord.gg/cosmosnetwork) (in the `#cometbft` channel) or
[Telegram](https://t.me/CometBFT).

Special thanks to @wcsiu, @ze97286, @faddat and @JayT106 for their contributions
to this release!

### BREAKING CHANGES

- Rename binary to `cometbft` and Docker image to `cometbft/cometbft`
  ([\#152](https://github.com/cometbft/cometbft/pull/152))
- The `TMHOME` environment variable was renamed to `CMTHOME`, and all
  environment variables starting with `TM_` are instead prefixed with `CMT_`
  ([\#211](https://github.com/cometbft/cometbft/issues/211))
- Use Go 1.19 to build CometBFT, since Go 1.18 has reached end-of-life.
  ([\#360](https://github.com/cometbft/cometbft/issues/360))

### BUG FIXES

- `[consensus]` Fixed a busy loop that happened when sending of a block part
  failed by sleeping in case of error.
  ([\#4](https://github.com/informalsystems/tendermint/pull/4))
- `[state/kvindexer]` Resolved crashes when event values contained slashes,
  introduced after adding event sequences.
  (\#[383](https://github.com/cometbft/cometbft/pull/383): @jmalicevic)
- `[consensus]` Short-term fix for the case when `needProofBlock` cannot find
  previous block meta by defaulting to the creation of a new proof block.
  ([\#386](https://github.com/cometbft/cometbft/pull/386): @adizere)
  - Special thanks to the [Vega.xyz](https://vega.xyz/) team, and in particular
    to Zohar (@ze97286), for reporting the problem and working with us to get to
    a fix.
- `[p2p]` Correctly use non-blocking `TrySendEnvelope` method when attempting to
  send messages, as opposed to the blocking `SendEnvelope` method. It is unclear
  whether this has a meaningful impact on P2P performance, but this patch does
  correct the underlying behaviour to what it should be
  ([tendermint/tendermint\#9936](https://github.com/tendermint/tendermint/pull/9936))

### DEPENDENCIES

- Replace [tm-db](https://github.com/tendermint/tm-db) with
  [cometbft-db](https://github.com/cometbft/cometbft-db)
  ([\#160](https://github.com/cometbft/cometbft/pull/160))
- Bump tm-load-test to v1.3.0 to remove implicit dependency on Tendermint Core
  ([\#165](https://github.com/cometbft/cometbft/pull/165))
- `[crypto]` Update to use btcec v2 and the latest btcutil
  ([tendermint/tendermint\#9787](https://github.com/tendermint/tendermint/pull/9787):
  @wcsiu)

### FEATURES

- `[rpc]` Add `match_event` query parameter to indicate to the RPC that it
  should match events _within_ attributes, not only within a height
  ([tendermint/tendermint\#9759](https://github.com/tendermint/tendermint/pull/9759))

### IMPROVEMENTS

- `[e2e]` Add functionality for uncoordinated (minor) upgrades
  ([\#56](https://github.com/tendermint/tendermint/pull/56))
- `[tools/tm-signer-harness]` Remove the folder as it is unused
  ([\#136](https://github.com/cometbft/cometbft/issues/136))
- Append the commit hash to the version of CometBFT being built
  ([\#204](https://github.com/cometbft/cometbft/pull/204))
- `[mempool/v1]` Suppress "rejected bad transaction" in priority mempool logs by
  reducing log level from info to debug
  ([\#314](https://github.com/cometbft/cometbft/pull/314): @JayT106)
- `[consensus]` Add `consensus_block_gossip_parts_received` and
  `consensus_step_duration_seconds` metrics in order to aid in investigating the
  impact of database compaction on consensus performance
  ([tendermint/tendermint\#9733](https://github.com/tendermint/tendermint/pull/9733))
- `[state/kvindexer]` Add `match.event` keyword to support condition evaluation
  based on the event the attributes belong to
  ([tendermint/tendermint\#9759](https://github.com/tendermint/tendermint/pull/9759))
- `[p2p]` Reduce log spam through reducing log level of "Dialing peer" and
  "Added peer" messages from info to debug
  ([tendermint/tendermint\#9764](https://github.com/tendermint/tendermint/pull/9764):
  @faddat)
- `[consensus]` Reduce bandwidth consumption of consensus votes by roughly 50%
  through fixing a small logic bug
  ([tendermint/tendermint\#9776](https://github.com/tendermint/tendermint/pull/9776))

---

CometBFT is a fork of [Tendermint Core](https://github.com/tendermint/tendermint) as of late December 2022.

## Bug bounty

Friendly reminder, we have a [bug bounty program](https://hackerone.com/cosmos).

## Previous changes

For changes released before the creation of CometBFT, please refer to the Tendermint Core [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/a9feb1c023e172b542c972605311af83b777855b/CHANGELOG.md).

