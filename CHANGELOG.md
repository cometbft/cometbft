# CHANGELOG

## Unreleased

### BREAKING CHANGES

- `[proto]` Renamed the packages from `tendermint.*` to `cometbft.*`
  and introduced versioned packages to distinguish between proto definitions
  released in 0.34.x, 0.37.x, 0.38.x, and 1.0.x versions.
  Prior to the 1.0 release, the versioned packages are suffixed with
  `.v1beta1`, `.v1beta2`, and so on; all definitions describing the protocols
  as per the 1.0.0 release are in packages suffixed with `.v1`.
  Relocated generated Go code into a new `api` folder and changed the import
  paths accordingly.
  ([\#495](https://github.com/cometbft/cometbft/pull/495)
  [\#1504](https://github.com/cometbft/cometbft/issues/1504))
- `[crypto/merkle]` The public `Proof.ComputeRootHash` function has been deleted.
   ([\#558](https://github.com/cometbft/cometbft/issues/558))
- `[rpc/grpc]` Remove the deprecated gRPC broadcast API
  ([\#650](https://github.com/cometbft/cometbft/issues/650))
- `[proto]` The names in the `cometbft.abci.v1` versioned proto package
  are changed to satisfy the
  [buf guidelines](https://buf.build/docs/best-practices/style-guide/)
  ([#736](https://github.com/cometbft/cometbft/issues/736),
   [#1504](https://github.com/cometbft/cometbft/issues/1504),
   [#1530](https://github.com/cometbft/cometbft/issues/1530)):
  * Names of request and response types used in gRPC changed by making
    `Request`/`Response` the suffix instead of the prefix, e.g.
    `RequestCheckTx` ⭢ `CheckTxRequest`.
  * The `Request` and `Response` multiplex messages are redefined accordingly.
  * `CheckTxType` values renamed with the `CHECK_TX_TYPE_` prefix.
  * `MisbehaviorType` values renamed with the `MISBEHAVIOR_TYPE_` prefix.
  * `Result` enum formerly nested in `ResponseOfferSnapshot` replaced with the package-level
    `OfferSnapshotResult`, its values named with the
    `OFFER_SNAPSHOT_RESULT_` prefix.
  * `Result` enum formerly nested in `ResponseApplyShapshotChunk` replaced with the package-level
    `ApplySnapshotChunkResult`, its values named with the
    `APPLY_SNAPSHOT_CHUNK_RESULT_` prefix.
  * `Status` enum formerly nested in `ResponseProcessProposal` replaced with the package-level
    `ProcessProposalStatus`, its values named with the
    `PROCESS_PROPOSAL_STATUS_` prefix.
  * `Status` enum formerly nested in `ResponseVerifyVoteExtension` replaced with the package-level
    `VerifyVoteExtensionStatus`, its values named with the
    `VERIFY_VOTE_EXTENSION_STATUS_` prefix.
  * New definition of `Misbehavior` using the changed `MisbehaviorType`.
  * The gRPC service is renamed `ABCIService` and defined using the types listed above.
- `[proto]` In the `cometbft.state.v1` package, the definition for `ABCIResponsesInfo`
  is changed, renaming `response_finalize_block` field to `finalize_block`.
- `[abci]` Changed the proto-derived enum type and constant aliases to the
  buf-recommended naming conventions adopted in the `abci/v1` proto package.
  For example, `ResponseProcessProposal_ACCEPT` is renamed to `PROCESS_PROPOSAL_STATUS_ACCEPT`
  ([\#736](https://github.com/cometbft/cometbft/issues/736)).
- `[abci]` The `Type` enum field is now required to be set to a value other
  than the default `CHECK_TX_TYPE_UNKNOWN` for a valid `CheckTxRequest`
  ([\#736](https://github.com/cometbft/cometbft/issues/736)).
- `[rpc]` The endpoints `broadcast_tx_*` now return an error when the node is
  performing block sync or state sync.
  ([\#785](https://github.com/cometbft/cometbft/issues/785))
- `[mempool]` When the node is performing block sync or state sync, the mempool
  reactor now discards incoming transactions from peers, and does not propagate
  transactions to peers.
  ([\#785](https://github.com/cometbft/cometbft/issues/785))
- `[consensus]` `Handshaker.Handshake` now requires `context.Context` ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
- `[node]` `NewNode` now requires `context.Context` as the first parameter ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
`[mempool]` Change the signature of `CheckTx` in the `Mempool` interface to
`CheckTx(tx types.Tx) (*abcicli.ReqRes, error)`. Also, add new method
`SetTxRemovedCallback`.
([\#1010](https://github.com/cometbft/cometbft/issues/1010))
- `[state]` The `state.Store` interface has been expanded
  to accommodate the data pull companion API of ADR 101
  ([\#1096](https://github.com/cometbft/cometbft/issues/1096))
- `[proxy]` Expand `ClientCreator` interface to allow
  for per-"connection" control of client creation
  ([\#1141](https://github.com/cometbft/cometbft/pull/1141))
- `[mempool]` Remove `mempoolIDs` for internally storing peer ids as `p2p.ID`
  instead of `uint16`.
  ([\#1146](https://github.com/cometbft/cometbft/pull/1146))
- `[cmd]` Remove `replay` and `replay-console` subcommands
  and corresponding consensus file replay code, such as
  `consensus.RunReplayFile`, and `consensus.State.ReplayFile`
  ([\#1170](https://github.com/cometbft/cometbft/pull/1170))
- `[state/indexer/block]` BlockIndexer now has additional method `Prune`, `GetRetainHeight`, `SetRetainHeight` ([\#1176](https://github.com/cometbft/cometbft/pull/1176))
- `[state/txindex]` TxIndexer now has additional methods: `Prune`, `GetRetainHeight`, `SetRetainHeight` ([\#1176](https://github.com/cometbft/cometbft/pull/1176))
- `[node]` Change the signature of `GenesisDocProvider` to
  return the checksum of JSON content alongside the parsed genesis data
  ([\#1287](https://github.com/cometbft/cometbft/issues/1287)).
- `[node]` Go-API breaking: Change the signature of `LoadStateFromDBOrGenesisDocProvider`
   to accept an optional operator provided hash of the genesis file
  ([\#1324](https://github.com/cometbft/cometbft/pull/1324)).
- `[version]` Bumped the P2P version from 8 to 9, as this release contains new P2P messages.
   ([\#1411](https://github.com/cometbft/cometbft/pull/1411))
- `[rpc/client]` Hard-code the `/websocket` endpoint path such that it is
  no longer configurable, removing the related client constructor parameter
  ([\#1412](https://github.com/cometbft/cometbft/pull/1412))
- `[libs/os]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[store]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/autofile]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/tempfile]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/clist]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/net]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/strings]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[state]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/rand]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/flowrate]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/async]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/timer]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/sync]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[statesync]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/bits]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/progressbar]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[blocksync]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/protoio]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/events]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/cmap]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[inspect]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/service]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[evidence]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/pubsub]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[libs/fail]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[consensus]` Move to `internal`
  ([\#1485](https://github.com/cometbft/cometbft/pull/1485))
- `[abci]` Renamed the alias types for gRPC requests, responses, and service
  instances to follow the naming changes in the proto-derived
  `api/cometbft/abci/v1` package
  ([\#1533](https://github.com/cometbft/cometbft/pull/1533)):
  * The prefixed naming pattern `RequestFoo`, `ReponseFoo` changed to
    suffixed `FooRequest`, `FooResponse`.
  * Each method gets its own unique request and response type to allow for
    independent evolution with backward compatibility.
  * `ABCIClient` renamed to `ABCIServiceClient`.
  * `ABCIServer` renamed to `ABCIServiceServer`.
- `[store]` Make the `LoadBlock` method also return block metadata
  ([\#1556](https://github.com/cometbft/cometbft/issues/1556))
- Made `/api` a standalone Go module with its own `go.mod`
  ([\#1561](https://github.com/cometbft/cometbft/issues/1561))
- `[comet]` Version variables, in `version/version.go`, have been renamed to reflect the CometBFT rebranding.
   ([cometbft/cometbft\#1621](https://github.com/cometbft/cometbft/pull/1621))
- `[state/store]` go-API breaking change in `PruneStates`: added parameter to pass the number of pruned states and return pruned entries in current pruning iteration. ([\#1972](https://github.com/cometbft/cometbft/pull/1972))
- `[state/store]` go-API breaking change in `PruneABCIResponses`: added parameter to force compaction. ([\#1972](https://github.com/cometbft/cometbft/pull/1972))
- `[proto]` Remove stateful block data retrieval methods from the
  data companion gRPC API as per
  [RFC 106](https://github.com/cometbft/cometbft/blob/main/docs/references/rfc/rfc-106-separate-stateful-methods.md)
  ([\#2230](https://github.com/cometbft/cometbft/issues/2230)):
  * `GetLatest` from `cometbft.services.block.v1.BlockService`;
  * `GetLatestBlockResults` from `cometbft.services.block_results.v1.BlockResultsService`.
- `[rpc/grpc]` Remove support for stateful block data retrieval methods from the
  data companion APIs as per [RFC 106](https://github.com/cometbft/cometbft/blob/main/docs/references/rfc/rfc-106-separate-stateful-methods.md)
  * `GetLatestBlock` method removed from the `BlockServiceClient` interface.
  * `GetLatestBlockResults` method removed from the `BlockResultServiceClient` interface.
  * `GetLatest` endpoint is no longer served by `BlockServiceServer` instances.
  * `GetLatestBlockResults` endpoint is no longer served by `BlockResultServiceServer` instances.
- `[p2p]` Rename `IPeerSet#List` to `Copy`, add `Random`, `ForEach` methods.
   Rename `PeerSet#List` to `Copy`, add `Random`, `ForEach` methods.
   ([\#2246](https://github.com/cometbft/cometbft/pull/2246))
- `[abci]` Deprecates `ABCIParams` field of `ConsensusParam` and
  introduces replacement in `FeatureParams` to enable Vote Extensions.
  ([\#2322](https://github.com/cometbft/cometbft/pull/2322))
- `[internal/state]` Moved function `MedianTime` to package `types`,
  and made it a method of `Commit` so it can be used by external packages.
  ([\#2397](https://github.com/cometbft/cometbft/pull/2397))

### BUG FIXES

`[consensus]` Fix for "Validation of `VoteExtensionsEnableHeight` can cause chain halt"
  ([ASA-2024-001](https://github.com/cometbft/cometbft/security/advisories/GHSA-qr8r-m495-7hc4))
- `[mempool]` Fix data races in `CListMempool` by making atomic the types of `height`, `txsBytes`, and
  `notifiedTxsAvailable`. ([\#642](https://github.com/cometbft/cometbft/pull/642))
- `[consensus]` Consensus now prevotes `nil` when the proposed value does not
  match the value the local validator has locked on
  ([\#1203](https://github.com/cometbft/cometbft/pull/1203))
- `[consensus]` Remove logic to unlock block on +2/3 prevote for nil
  ([\#1175](https://github.com/cometbft/cometbft/pull/1175): @BrendanChou)
- `[state/indexer]` Respect both height params while querying for events
   ([\#1529](https://github.com/cometbft/cometbft/pull/1529))
- `[state/pruning]` When no blocks are pruned, do not attempt to prune statestore
   ([\#1616](https://github.com/cometbft/cometbft/pull/1616))
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
- `[p2p/pex]` gracefully shutdown Reactor ([\#2010](https://github.com/cometbft/cometbft/pull/2010))
- `[privval]` Retry accepting a connection ([\#2047](https://github.com/cometbft/cometbft/pull/2047))
- `[state]` Fix rollback to a specific height
  ([\#2136](https://github.com/cometbft/cometbft/pull/2136))

### DEPENDENCIES

- Bump cometbft-db to v0.9.0, providing support for RocksDB v8
  ([\#1725](https://github.com/cometbft/cometbft/pull/1725))

### FEATURES

- `[rpc/grpc]` Add gRPC server to the node, configurable
  via a new `[grpc]` section in the configuration file
  ([\#816](https://github.com/cometbft/cometbft/issues/816))
- `[rpc/grpc]` Add gRPC version service to allow clients to
  establish the software and protocol versions of the node
  ([\#816](https://github.com/cometbft/cometbft/issues/816))
- `[rpc/grpc]` Add gRPC client with support for version service
  ([\#816](https://github.com/cometbft/cometbft/issues/816))
- `[grpc]` Add `BlockService` with client to facilitate fetching of blocks and
  streaming of the latest committed block height
  ([\#1094](https://github.com/cometbft/cometbft/issues/1094))
- `[config]` Add `[grpc.block_service]` section to configure gRPC `BlockService`
  ([\#1094](https://github.com/cometbft/cometbft/issues/1094))
- `[grpc]` Add `BlockResultsService` with client to fetch BlockResults
  for a given height, or latest.
  ([\#1095](https://github.com/cometbft/cometbft/issues/1095))
- `[config]` Add `[grpc.block_results_service]` gRPC configuration `BlockResultsService`
  ([\#1095](https://github.com/cometbft/cometbft/issues/1095))
- `[proto]` Add definitions and generated code for
  [ADR-101](./docs/architecture/adr-101-data-companion-pull-api.md)
  `PruningService` in the `cometbft.services.pruning.v1` proto package
  ([\#1097](https://github.com/cometbft/cometbft/issues/1097)).
- `[rpc/grpc]` Add privileged gRPC server and client facilities, in
  `server/privileged` and `client/privileged` packages respectively, to
  enable a separate API server within the node which serves trusted clients
  without authentication and should never be exposed to public internet
  ([\#1097](https://github.com/cometbft/cometbft/issues/1097)).
- `[rpc/grpc]` Add a pruning service adding on the privileged gRPC server API to
  give an [ADR-101](./docs/architecture/adr-101-data-companion-pull-api.md) data
  companion control over block data retained by the node. The
  `WithPruningService` option method in `server/privileged` is provided to
  configure the pruning service
  ([\#1097](https://github.com/cometbft/cometbft/issues/1097)).
- `[rpc/grpc]` Add `PruningServiceClient` interface
  for the gRPC client in `client/privileged` along with a configuration option
  to enable it
  ([\#1097](https://github.com/cometbft/cometbft/issues/1097)).
- `[config]` Add `[grpc.privileged]` section to configure the privileged
  gRPC server for the node, and `[grpc.privileged.pruning_service]` section
  to control the pruning service
  ([\#1097](https://github.com/cometbft/cometbft/issues/1097)).
- `[metrics]` Add metrics to monitor pruning and current available data in stores: `PruningServiceBlockRetainHeight`, `PruningServiceBlockResultsRetainHeight`, `ApplicationBlockRetainHeight`, `BlockStoreBaseHeight`, `ABCIResultsBaseHeight`.
  ([\#1234](https://github.com/cometbft/cometbft/pull/1234))
`[rpc/grpc]` Add gRPC endpoint for pruning the block and transaction indexes
([\#1327](https://github.com/cometbft/cometbft/pull/1327))
- `[state]` Add TxIndexer and BlockIndexer pruning metrics
  ([\#1334](https://github.com/cometbft/cometbft/issues/1334))
- `[metrics]` Add metric for mempool size in bytes `SizeBytes`.
  ([\#1512](https://github.com/cometbft/cometbft/pull/1512))
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
- `[config]` Add configuration parameters to tweak forced compaction. ([\#1972](https://github.com/cometbft/cometbft/pull/1972))
- `[store]` When pruning force compaction of the database. ([\#1972](https://github.com/cometbft/cometbft/pull/1972))
- `[metrics]` Added metrics to monitor state store access. ([\#1974](https://github.com/cometbft/cometbft/pull/1974))
- `[metrics]` Added metrics to monitor block store access. ([\#1974](https://github.com/cometbft/cometbft/pull/1974))
- `[test]` Added monitoring tools and dashboards for local testing with `localnet`. ([\#2107](https://github.com/cometbft/cometbft/issues/2107))
- Add [`pebbledb`](https://github.com/cockroachdb/pebble). To use, build with
  `pebbledb` tag (`go build -tags pebbledb`) ([\#2132](https://github.com/cometbft/cometbft/pull/2132/))
- [light/store] Added support for a different DB key representation within the light block store ([\#2327](https://github.com/cometbft/cometbft/pull/2327/))
- [store] Added support for a different DB key representation to state and block store ([\#2327](https://github.com/cometbft/cometbft/pull/2327/))
- [evidence/store] Added support for a different DB key representation within the evidence store ([\#2327](https://github.com/cometbft/cometbft/pull/2327/))
- `[e2e]` Add `block_max_bytes` option to the manifest file.
  ([\#2362](https://github.com/cometbft/cometbft/pull/2362))
- `[e2e]` Add new `--testnet-dir` parameter to set a custom directory for the generated testnet files.
  ([\#2433](https://github.com/cometbft/cometbft/pull/2433))
- `[docs]` Add report on storage improvements and findings. ([\#2569](https://github.com/cometbft/cometbft/pull/2569))
- `[consensus]` add a new `synchrony` field to the `ConsensusParameter` struct
  for controlling the parameters of the proposer-based timestamp algorithm. (@williambanfield)
  ([tendermint/tendermint\#7354](https://github.com/tendermint/tendermint/pull/7354))
- `[consensus]` Update the proposal logic per the Propose-based timestamps specification
  so that the proposer will wait for the previous block time to occur
  before proposing the next block. (@williambanfield)
  ([tendermint/tendermint\#7376](https://github.com/tendermint/tendermint/pull/7376))
- `[consensus]` Update block validation to no longer require the block timestamp
  to be the median of the timestamps of the previous commit. (@anca)
  ([tendermint/tendermint\#7382](https://github.com/tendermint/tendermint/pull/7382))
- `[consensus]` Use the proposed block timestamp as the proposal timestamp.
  Update the block validation logic to ensure that the proposed block's timestamp
  matches the timestamp in the proposal message. (@williambanfield)
  ([tendermint/tendermint\#7391](https://github.com/tendermint/tendermint/pull/7391))
- `[consensus]` Update proposal validation logic to Prevote nil
  if a proposal does not meet the conditions for Timeliness
  per the proposer-based timestamp specification. (@anca)
  ([tendermint/tendermint\#7415](https://github.com/tendermint/tendermint/pull/7415))
- `[consensus]` Use the proposer timestamp for the first height instead of the genesis time.
  Chains will still start consensus at the genesis time. (@anca)
  ([tendermint/tendermint\#7711](https://github.com/tendermint/tendermint/pull/7711))

### IMPROVEMENTS

- `[state/indexer]` Add transaction and block index pruning
  ([\#1176](https://github.com/cometbft/cometbft/pull/1176))
- `[mempool]` Add a metric (a counter) to measure whether a tx was received more than once.
  ([\#634](https://github.com/cometbft/cometbft/pull/634))
- `[docs/references]` Added ADR-102: RPC Companion.
  ([\#658](https://github.com/cometbft/cometbft/pull/658))
- `[consensus]` New metrics (counters) to track duplicate votes and block parts.
  ([\#896](https://github.com/cometbft/cometbft/pull/896))
- `[consensus]` Optimize vote and block part gossip with new message `HasProposalBlockPartMessage`,
  which is similar to `HasVoteMessage`; and random sleep in the loop broadcasting those messages.
  The sleep can be configured with new config `peer_gossip_intraloop_sleep_duration`, which is set to 0
  by default as this is experimental.
  Our scale tests show substantial bandwidth improvement with a value of 50 ms.
  ([\#904](https://github.com/cometbft/cometbft/pull/904))
- Update Apalache type annotations in the light client spec ([#955](https://github.com/cometbft/cometbft/pull/955))
- `[node]` Remove genesis persistence in state db, replaced by a hash
  ([cometbft/cometbft\#1017](https://github.com/cometbft/cometbft/pull/1017),
  [cometbft/cometbft\#1295](https://github.com/cometbft/cometbft/pull/1295))
- `[consensus]` Log vote validation failures at info level
  ([\#1022](https://github.com/cometbft/cometbft/pull/1022))
- `[config]` Added `[storage.pruning]` and `[storage.pruning.data_companion]`
  sections to facilitate background pruning and data companion (ADR 101)
  operations ([\#1096](https://github.com/cometbft/cometbft/issues/1096))
- `[state]` ABCI response pruning has been added for use by the data companion
  ([\#1096](https://github.com/cometbft/cometbft/issues/1096))
- `[state]` Block pruning has been moved from the block executor into a
  background process ([\#1096](https://github.com/cometbft/cometbft/issues/1096))
- `[node]` The `node.Node` struct now manages a
  `state.Pruner` service to facilitate background pruning
  ([\#1096](https://github.com/cometbft/cometbft/issues/1096))
- `[abci/client]` Add fully unsynchronized local client creator, which
  imposes no mutexes on the application, leaving all handling of concurrency up
  to the application ([\#1141](https://github.com/cometbft/cometbft/pull/1141))
- `[abci/client]` Add consensus-synchronized local client creator,
  which only imposes a mutex on the consensus "connection", leaving
  the concurrency of all other "connections" up to the application
  ([\#1141](https://github.com/cometbft/cometbft/pull/1141))
- `[consensus]` When prevoting, avoid calling PropocessProposal when we know the
  proposal was already validated by correct nodes.
  ([\#1230](https://github.com/cometbft/cometbft/pull/1230))
- `[node]` On upgrade, after [\#1296](https://github.com/cometbft/cometbft/pull/1296), delete the genesis file existing in the DB.
  ([cometbft/cometbft\#1297](https://github.com/cometbft/cometbft/pull/1297)
- `[cli/node]` The genesis hash provided with the `--genesis-hash` is now
   forwarded to the node, instead of reading the file.
  ([\#1324](https://github.com/cometbft/cometbft/pull/1324)).
- `[rpc]` The RPC API is now versioned, with all existing endpoints accessible
  via `/v1/*` as well as `/*`
  ([\#1412](https://github.com/cometbft/cometbft/pull/1412))
- `[consensus]` Reduce the default MaxBytes to 4MB and increase MaxGas to 10 million
  ([\#1518](https://github.com/cometbft/cometbft/pull/1518))
- `[mempool]` Add experimental feature to limit the number of persistent peers and non-persistent
  peers to which the node gossip transactions.
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[config]` Add mempool parameters `experimental_max_gossip_connections_to_persistent_peers` and
  `experimental_max_gossip_connections_to_non_persistent_peers` for limiting the number of peers to
  which the node gossip transactions.
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[e2e]` Allow latency emulation between nodes.
  ([\#1559](https://github.com/cometbft/cometbft/pull/1559))
- `[e2e]` Allow latency emulation between nodes.
  ([\#1560](https://github.com/cometbft/cometbft/pull/1560))
- `[e2e]` Allow disabling the PEX reactor on all nodes in the testnet
  ([\#1579](https://github.com/cometbft/cometbft/pull/1579))
- `[rpc]` Export `MakeHTTPDialer` to allow HTTP client constructors more flexibility.
  ([\#1594](https://github.com/cometbft/cometbft/pull/1594))
- `[types]` Validate `Validator#Address` in `ValidateBasic` ([\#1715](https://github.com/cometbft/cometbft/pull/1715))
- `[abci]` Increase ABCI socket message size limit to 2GB ([\#1730](https://github.com/cometbft/cometbft/pull/1730): @troykessler)
- `[state]` Save the state using a single DB batch ([\#1735](https://github.com/cometbft/cometbft/pull/1735))
- `[store]` Save block using a single DB batch if block is less than 640kB, otherwise each block part is saved individually
  ([\#1755](https://github.com/cometbft/cometbft/pull/1755))
- `[rpc]` Support setting proxy from env to `DefaultHttpClient`.
  ([\#1900](https://github.com/cometbft/cometbft/pull/1900))
- `[p2p]` Export p2p package errors ([\#1901](https://github.com/cometbft/cometbft/pull/1901)) (contributes to [\#1140](https://github.com/cometbft/cometbft/issues/1140))
- `[rpc]` Use default port for HTTP(S) URLs when there is no explicit port ([\#1903](https://github.com/cometbft/cometbft/pull/1903))
- `[light]` Export light package errors ([\#1904](https://github.com/cometbft/cometbft/pull/1904)) (contributes to [\#1140](https://github.com/cometbft/cometbft/issues/1140))
- `[crypto/merkle]` faster calculation of hashes ([#1921](https://github.com/cometbft/cometbft/pull/1921))
- Removed undesired linting from `Makefile` and added dependency check for `codespell`.
- `[blocksync]` Avoid double-calling `types.BlockFromProto` for performance
  reasons ([\#2016](https://github.com/cometbft/cometbft/pull/2016))
- `[state]` avoid double-saving `FinalizeBlockResponse` for performance reasons
  ([\#2017](https://github.com/cometbft/cometbft/pull/2017))
- `[e2e]` Add manifest option `VoteExtensionsUpdateHeight` to test
  vote extension activation via `InitChain` and `FinalizeBlock`.
  Also, extend the manifest generator to produce different values
  of this new option
  ([\#2065](https://github.com/cometbft/cometbft/pull/2065))
- `[consensus]` Add `chain_size_bytes` metric for measuring the size of the blockchain in bytes
  ([\#2093](https://github.com/cometbft/cometbft/pull/2093))
- `[e2e]` Add manifest option `load_max_txs` to limit the number of transactions generated by the
  `load` command. ([\#2094](https://github.com/cometbft/cometbft/pull/2094))
- Optimized the PSQL indexer
  ([\#2142](https://github.com/cometbft/cometbft/pull/2142)) thanks to external contributor @k0marov !
- `[e2e]` Add new targets `fast` and `clean` to Makefile.
  ([\#2192](https://github.com/cometbft/cometbft/pull/2192))
- `[rpc]` Export RPC package errors ([\#2200](https://github.com/cometbft/cometbft/pull/2200)) (contributes to [\#1140](https://github.com/cometbft/cometbft/issues/1140))
- `[p2p]` make `PeerSet.Remove` more efficient (Author: @odeke-em) [\#2246](https://github.com/cometbft/cometbft/pull/2246)
- `[jsonrpc]` enable HTTP basic auth in websocket client ([#2434](https://github.com/cometbft/cometbft/pull/2434))
- `[e2e]` Introduce the possibility in the manifest for some nodes
  to run in a preconfigured clock skew.
  ([\#2453](https://github.com/cometbft/cometbft/pull/2453))
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
- `[privval]` DO NOT require extension signature from privval if vote
  extensions are disabled. Remote signers can skip signing the extension if
  `skip_extension_signing` flag in `SignVoteRequest` is true.
  [\#2496](https://github.com/cometbft/cometbft/pull/2496)
- `[proto]` Add `skip_extension_signing` field to the `SignVoteRequest` message
  in `cometbft.privval.v1` ([\#2522](https://github.com/cometbft/cometbft/pull/2522)).
  The `cometbft.privval.v1beta2` package is added to capture the protocol as it was
  released in CometBFT 0.38.x
  ([\#2529](https://github.com/cometbft/cometbft/pull/2529)).

### MINIMUM GO VERSION

- Bump minimum Go version to v1.21
  ([\#1244](https://github.com/cometbft/cometbft/pull/1244))

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

## v0.34.31

*November 27, 2023*

Fixes a small bug in the mempool for an experimental feature.

### BUG FIXES

- `[mempool]` Avoid infinite wait in transaction sending routine when
  using experimental parameters to limiting transaction gossiping to peers
  ([\#1654](https://github.com/cometbft/cometbft/pull/1654))

## v0.34.30

*November 17, 2023*

This release contains, among other things, an opt-in, experimental feature to
help reduce the bandwidth consumption associated with the mempool's transaction
gossip.

### BUILD

- Bump Go version used to v1.20 since v1.19 has reached EOL
  ([\#1351](https://github.com/cometbft/cometbft/pull/1351))

### FEATURES

- `[metrics]` Add metric for mempool size in bytes `SizeBytes`.
  ([\#1512](https://github.com/cometbft/cometbft/pull/1512))

### IMPROVEMENTS

- `[node]` Make handshake cancelable ([cometbft/cometbft\#857](https://github.com/cometbft/cometbft/pull/857))
- `[node]` Close evidence.db OnStop ([cometbft/cometbft\#1210](https://github.com/cometbft/cometbft/pull/1210): @chillyvee)
- `[mempool]` Add experimental feature to limit the number of persistent peers and non-persistent
  peers to which the node gossip transactions (only for "v0" mempool).
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558),
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))
- `[config]` Add mempool parameters `experimental_max_gossip_connections_to_persistent_peers` and
  `experimental_max_gossip_connections_to_non_persistent_peers` for limiting the number of peers to
  which the node gossip transactions.
  ([\#1558](https://github.com/cometbft/cometbft/pull/1558))
  ([\#1584](https://github.com/cometbft/cometbft/pull/1584))

## v0.34.29

*June 14, 2023*

Provides several minor bug fixes, as well as fixes for several low-severity
security issues.

### BUG FIXES

- `[state/kvindex]` Querying event attributes that are bigger than int64 is now
  enabled. ([\#771](https://github.com/cometbft/cometbft/pull/771))
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
  ([\#788](https://github.com/cometbft/cometbft/pull/788))
- `[cmd/cometbft/commands/debug/kill]` **Low severity** - Fix unsafe int cast in
  `debug kill` command ([\#794](https://github.com/cometbft/cometbft/pull/794))
- `[consensus]` **Low severity** - Avoid recursive call after rename to
  `(*PeerState).MarshalJSON`
  ([\#863](https://github.com/cometbft/cometbft/pull/863))
- `[mempool/clist_mempool]` **Low severity** - Prevent a transaction from
  appearing twice in the mempool
  ([\#890](https://github.com/cometbft/cometbft/pull/890): @otrack)

## v0.34.28

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

- `[crypto/sr25519]` Upgrade to go-schnorrkel@v1.0.0 ([\#475](https://github.com/cometbft/cometbft/issues/475))
- `[jsonrpc/client]` Improve the error message for client errors stemming from
  bad HTTP responses.
  ([cometbft/cometbft\#638](https://github.com/cometbft/cometbft/pull/638))

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

