# ADR 101: Data Companion Pull API

## Changelog

- 2023-04-04: Update based on review feedback (@thanethomson)
- 2023-02-28: Renumber from 084 to 101 (@thanethomson)
- 2022-12-18: First draft (@thanethomson)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

Following from the discussion around the development of [ADR 100][adr-100], an
alternative model is proposed here for offloading certain data from nodes to a
"data companion". This alternative model inverts the control of the data
offloading process, when compared to ADR 100, from the node to the data
companion.

Overall, this approach provides slightly weaker guarantees than that of ADR 100,
but represents a simpler model to implement.

## Alternative Approaches

Other considered alternatives to this ADR are also outlined in
[ADR-100][adr-100].

## Decision

> This section records the decision that was made.

## Detailed Design

### Requirements

Similar requirements are proposed here as for [ADR-100][adr-100].

1. A node _must_ support at most one data companion.

2. All or part of the following data _must_ be obtainable by the companion, and
   as close to real-time as possible:
   1. Committed block data
   2. `FinalizeBlockResponse` data, but only for committed blocks

3. The companion _must_ be able to establish the earliest height for which the
   node has all of the associated data.

4. The API _must_ be (or be able to be) appropriately shielded from untrusted
   consumers and abuse. Critical control facets of the API (e.g. those that
   influence the node's pruning mechanisms) _must_ be implemented in such a way
   as to eliminate the possibility of accidentally exposing those endpoints to
   the public internet unprotected.

5. The node _must_ know, by way of signals from the companion, which heights'
   associated data are safe to prune.

6. The companion _must_ be able to handle the possibility that a node might
   start from a non-zero height (i.e. that the node may state sync from a
   specific height beyond genesis).

7. The companion _must_ be able to handle the possibility that a node may
   disable and then later re-enable the companion interface, potentially causing
   the companion to have missing data in between those two heights.

8. The API _must_ be opt-in. When off or not in use, it _should_ have no impact
   on system performance.

9. The API _must_ not cause back-pressure into consensus.

10. It _must_ not cause unbounded memory growth.

11. It _must_ provide one or more ways for operators to control storage growth.

12. It _must_ provide insight to operators (e.g. by way of logs/metrics) to
    assist in dealing with possible failure modes.

13. The solution _should_ be able to be backported to older versions of
    CometBFT (e.g. v0.34).

### Entity Relationships

The following model shows the proposed relationships between CometBFT, a
socket-based ABCI application, and the proposed data companion service.

```
     +----------+      +------------+      +----------------+
     | ABCI App | <--- |  CometBFT  | <--- | Data Companion |
     +----------+      +------------+      +----------------+
```

In this diagram, it is evident that CometBFT connects out to the ABCI
application, and the companion connects to the CometBFT node.

### Pruning Behaviour

There are two "modes" of pruning that need to be supported by the node:

1. **Data companion disabled**: Here, the node prunes as per normal based on the
   `retain_height` parameter supplied by the application via the
   [`Commit`][abci-commit] ABCI response.
2. **Data companion enabled**: Here, the node prunes _blocks_ based on the minimum
   among the retain heights specified by the application and the data companion,
   and the node prunes _block results_ based on the retain height set by the
   data companion.

Enabling the `discard_abci_responses` flag under the `[storage]` section in the
configuration is incompatible with enabling a data companion. If
`storage.discard_abci_responses` and `data_companion.enabled` are both `true`,
then the node _must_ fail to start.

### gRPC API

At the time of this writing, it is proposed that CometBFT implement a full
gRPC interface ([\#81]). As such, we have several options when it comes to
implementing the data companion pull API:

1. Extend the existing RPC API to simply provide the additional data
   companion-specific endpoints. In order to meet the
   [requirements](#requirements), however, some of the endpoints will have to be
   protected by default. This is simpler for clients to interact with though,
   because they only need to interact with a single endpoint for all of their
   needs.
2. Implement a separate RPC API on a different port to the standard gRPC
   interface. This allows for a clearer distinction between the standard and
   data companion-specific gRPC interfaces, but complicates the server and
   client interaction models.

Due to the poorer operator experience in option 2, it would be preferable to
implement option 1, but have certain endpoints be
[access-controlled](#access-control) by default.

With this in mind, the following gRPC API is proposed, where the CometBFT node
will implement these services.

#### Block Service

When implementing gRPC support for a node, this service (or at least a slight
variation of it) would be applicable anyways, regardless of whether this ADR is
implemented.

```protobuf
syntax = "proto3";

package tendermint.block_service.v1;

import "tendermint/abci/types.proto";
import "tendermint/types/types.proto";
import "tendermint/types/block.proto";

// BlockService provides information about blocks.
service BlockService {
    // GetLatestBlockID returns a stream of the latest block IDs as they are
    // committed by the network.
    rpc GetLatestBlockID(GetLatestBlockIDRequest) returns (stream GetLatestBlockIDResponse) {}

    // GetBlock attempts to retrieve the block at a particular height.
    rpc GetBlock(GetBlockRequest) returns (GetBlockResponse) {}

    // GetBlockResults attempts to retrieve the results of block execution for a
    // particular height.
    rpc GetBlockResults(GetBlockResultsRequest) returns (GetBlockResultsResponse) {}
}

message GetLatestBlockIDRequest {}

// GetLatestBlockIDResponse is a lightweight reference to the latest committed
// block.
message GetLatestBlockIDResponse {
    // The height of the latest committed block.
    int64 height = 1;
    // The ID of the latest committed block.
    tendermint.types.BlockID block_id = 2;
}

message GetBlockRequest {
    // The height of the block to get.
    int64 height = 1;
}

message GetBlockResponse {
    // Block data for the requested height.
    tendermint.types.Block block = 1;
}

message GetBlockResultsRequest {
    // The height of the block results to get.
    int64 height = 1;
}

message GetBlockResultsResponse {
    // All events produced by the ABCI BeginBlock call for the block.
    repeated tendermint.abci.Event begin_block_events = 1;

    // All transaction results produced by block execution.
    repeated tendermint.abci.ExecTxResult tx_results = 2;

    // Validator updates during block execution.
    repeated tendermint.abci.ValidatorUpdate validator_updates = 3;

    // Consensus parameter updates during block execution.
    tendermint.types.ConsensusParams consensus_param_updates = 4;

    // All events produced by the ABCI EndBlock call.
    // NB: This should be called finalize_block_events when ABCI 2.0 lands.
    repeated tendermint.abci.Event end_block_events = 5;
}
```

#### Data Companion Service

```protobuf
syntax = "proto3";

package tendermint.data_companion_service.v1;

// DataCompanionService provides privileged access to specialized pruning
// functionality on the CometBFT node to help optimize node storage.
service DataCompanionService {
    // SetRetainHeight notifies the node of the minimum height whose data must
    // be retained by the node. This data includes block data and block
    // execution results.
    //
    // Setting a retain height lower than a previous setting will result in an
    // error.
    //
    // The lower of this retain height and that set by the application in its
    // Commit response will be used by the node to determine which heights' data
    // can be pruned.
    rpc SetRetainHeight(SetRetainHeightRequest) returns (SetRetainHeightResponse) {}

    // GetRetainHeight returns the retain height set by the companion and that
    // set by the application. This can give the companion an indication as to
    // which heights' data are currently available.
    rpc GetRetainHeight(GetRetainHeightRequest) returns (GetRetainHeightResponse) {}
}

message SetRetainHeightRequest {
    int64 height = 1;
}

message SetRetainHeightResponse {}

message GetRetainHeightRequest {}

message GetRetainHeightResponse {
    // The retain height as set by the data companion.
    int64 data_companion_retain_height = 1;
    // The retain height as set by the ABCI application.
    int64 app_retain_height = 2;
}
```

With this API design, it is technically possible for an integrator to attach
multiple data companions to the node, but only one of their retain heights will
be applied.

### Access Control

As covered in the [gRPC API section](#grpc-api), it would be preferable to
implement some form of access control for sensitive, data companion-specific
APIs. At least **basic HTTP authentication** should be implemented for these
endpoints, where credentials should be obtained from an `.htpasswd` file, using
the same format as [Apache `.htpasswd` files][htpasswd], whose location is set
in the CometBFT configuration file.

`.htpasswd` files are relatively standard and are supported by many web servers.
The format is relatively straightforward to parse and interpret.

We should strongly consider only supporting the bcrypt encryption option for
passwords stored in `.htpasswd` files.

### Configuration

The following configuration file update is proposed to support the data
companion API.

```toml
# A data companion, if enabled, is intended to offload the storage of certain
# types of data from the node. Specifically:
# 1. Block data, including transactions.
# 2. Block results, including events, transaction results, validator and
#    consensus parameter updates.
#
# A data companion can influence the pruning height of the node, and therefore
# the data companion gRPC service is considered to be a sensitive endpoint that
# is password-protected by default.
[data_companion]

# Is the data companion gRPC API enabled at all? Default: false
enabled = false

# The maximum number of blocks to allow the data companion's retain height to
# lag behind the current height. The node will block all other operations until
# the companion's retain height is less than `max_retain_lag` blocks behind the
# current height. Default: 100
max_retain_lag = 100

# Authentication configuration for the data companion.
[data_companion.authentication]

# The authentication method to use. At present, the only supported method is
# "basic" (i.e. basic HTTP authentication).
method = "basic"

# Path to the file containing basic authentication credentials to access the
# data companion service. If the data companion is enabled and this password
# file is not supplied, or the file does not exist, or the file is of an invalid
# format, the node will fail to start.
#
# See https://httpd.apache.org/docs/current/programs/htpasswd.html for details
# on how to create/configure this file.
password_file = "/path/to/.htpasswd"
```

### Metrics

The following metrics are proposed to be added to monitor the health of the
interaction between a node and its data companion:

- `data_companion_retain_height` - The current retain height as requested by the
  data companion. This can give operators insight into whether the companion is
  lagging significantly behind the current network height.

## Consequences

### Positive

- Facilitates offloading of data to an external service, which can be scaled
  independently of the node
  - Potentially reduces load on the node itself
  - Paves the way for eventually reducing the surface area of a node's exposed
    APIs
- Allows the data companion more leeway in reading the data it needs than the
  approach in [ADR 100][adr-100]
- Simpler implementation than [ADR 100][adr-100]

### Negative

- Increases system complexity slightly in the short-term
- If data companions are not correctly implemented and deployed (e.g. if a
  companion is attached to the same storage as the node, and/or if its retain
  height signaling is poorly handled), this could result in substantially
  increased storage usage

### Neutral

- Expands the overall API surface area of a node in the short-term

## References

- [ADR 100 - Data Companion Push API][adr-100]
- [\#81 - rpc: Add gRPC support][\#81]
- [`.htpasswd`][htpasswd]

[adr-100]: https://github.com/cometbft/cometbft/pull/73
[\#81]: https://github.com/cometbft/cometbft/issues/81
[htpasswd]: https://httpd.apache.org/docs/current/programs/htpasswd.html
[abci-commit]: ../../spec/abci/abci++_methods.md#commit
