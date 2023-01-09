# ADR 084: Data Companion Pull API

## Changelog

- 2022-12-18: First draft (@thanethomson)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

Following from the discussion around the development of [ADR 082][adr-082], an
alternative model is proposed here for offloading certain data from nodes to a
"data companion". This alternative model inverts the control of the data
offloading process, when compared to ADR 082, from the node to the data
companion.

Overall, this approach provides slightly weaker guarantees than that of ADR 082,
but represents a simpler model to implement.

## Alternative Approaches

Other considered alternatives to this ADR are also outlined in
[ADR-082][adr-082].

## Decision

> This section records the decision that was made.

## Detailed Design

### Requirements

Similar requirements are proposed here as for [ADR-082][adr-082].

1. A node _must_ support at most one data companion.

2. All or part of the following data _must_ be obtainable by the companion, and
   as close to real-time as possible:
   1. Committed block data
   2. `FinalizeBlockResponse` data, but only for committed blocks

3. The companion _must_ be able to establish the earliest height for which the
   node has all of the requisite data.

4. The API _must_ be (or be able to be) appropriately shielded from untrusted
   consumers and abuse. Critical control facets of the API (e.g. those that
   influence the node's pruning mechanisms) _must_ be implemented in such a way
   as to eliminate the possibility of accidentally exposing those endpoints to
   the public internet unprotected.

5. The node _must_ know, by way of signals from the companion, which heights'
   associated data are safe to automatically prune.

6. The API _must_ be opt-in. When off or not in use, it _should_ have no impact
   on system performance.

7. It _must_ not cause back-pressure into consensus.

8. It _must_ not cause unbounded memory growth.

9. It _must_ provide one or more ways for operators to control storage growth.

10. It _must_ provide insight to operators (e.g. by way of logs/metrics) to
    assist in dealing with possible failure modes.

11. The solution _should_ be able to be backported to older versions of
    Tendermint (e.g. v0.34).

### Entity Relationships

The following model shows the proposed relationships between Tendermint, a
socket-based ABCI application, and the proposed data companion service.

```
     +----------+      +------------+      +----------------+
     | ABCI App | <--- | Tendermint | <--- | Data Companion |
     +----------+      +------------+      +----------------+
```

In this diagram, it is evident that Tendermint connects out to the ABCI
application, and the companion connects to the Tendermint node.

### gRPC API

At the time of this writing, it is proposed that Tendermint implement a full
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

With this in mind, the following gRPC API is proposed, where the Tendermint node
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
// functionality on the Tendermint node to help optimize node storage.
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
in the Tendermint configuration file.

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
  approach in [ADR 082][adr-082]
- Simpler implementation than [ADR 082][adr-082]

### Negative

- Increases system complexity slightly in the short-term
- If data companions are not correctly implemented and deployed (e.g. if a
  companion is attached to the same storage as the node, and/or if its retain
  height signaling is poorly handled), this could result in substantially
  increased storage usage

### Neutral

- Expands the overall API surface area of a node in the short-term

## References

- [ADR 082 - Data Companion Push API][adr-082]
- [\#81 - rpc: Add gRPC support][\#81]
- [`.htpasswd`][htpasswd]

[adr-082]: https://github.com/CometBFT/tendermint/pull/73
[\#81]: https://github.com/CometBFT/tendermint/issues/81
[htpasswd]: https://httpd.apache.org/docs/current/programs/htpasswd.html
