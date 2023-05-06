# ADR 101: Data Companion Pull API

## Changelog

- 2023-05-03: Update based on synchronous feedback from team (@thanethomson)
- 2023-04-04: Update based on review feedback (@thanethomson)
- 2023-02-28: Renumber from 084 to 101 (@thanethomson)
- 2022-12-18: First draft (@thanethomson)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

See the [context for ADR-100][adr-100-context].

The primary novelty introduced in this ADR is effectively just a new gRPC API
that allows an external application to influence which data the node prunes.
Otherwise, existing and planned RPC interfaces (such as the planned gRPC
interface) in addition to this pruning API should, in theory, deliver the same
kind of value as the solution proposed in ADR-100.

Even though the pruning API could be useful to operators outside the context of
the usage of a data companion (e.g. it could provide operators with more control
over node pruning behaviour), it is presented as part of the data companion
discussion to illustrate its initial intended use case.

## Alternative Approaches

[ADR-100][adr-100], as well as the [alternatives][adr-100-alt] outlined in
ADR-100, are all alternative approaches.

## Decision

To implement ADR-101 instead of ADR-100.

## Detailed Design

The model proposed in this ADR inverts that proposed in ADR-100, with the node
being the server and the data companion being the client. Here, the companion
"pulls" data from the node.

This provides much weaker data delivery guarantees than the "push" model of
ADR-100. In this "pull" model, the companion can lag behind consensus, but the
node does not crash if the companion is unavailable.

### Requirements

The requirements for ADR-101 are the same as the [requirements for
ADR-100][adr-100-req].

### Entity Relationships

The following model shows the proposed relationships between CometBFT, a
socket-based ABCI application, and the proposed data companion service.

```mermaid
flowchart RL
    comet[CometBFT]
    companion[Data Companion]
    app[ABCI App]

    comet --> app
    companion --> comet
```

In this diagram, it is evident that CometBFT (as a client) connects out to the
ABCI application (a server), and the companion (a client) connects to the
CometBFT node (a server).

### Pruning Behaviour

Two parameters are proposed as a necessary part of the pruning API:

- **Pruning service block retain height**, which influences the height to which
  the node will retain blocks. This is different to the **application block
  retain height**, which is set by the application in its response to each ABCI
  `commit` message.

  The node will prune blocks to whichever is lower between the pruning service
  and application block retain heights.

  The default value is 0, which indicates that only the application block retain
  height must be taken into consideration.

- **Pruning service block results retain height**, which influences the height
  to which the node will retain block results.

  The default value is 0, which indicates that all block results will be
  retained unless the `storage.discard_abci_responses` parameter is enabled, in
  which case no block results will be stored except those that are necessary to
  facilitate consensus.

These parameters need to be durable (i.e. stored on disk). Setting either of
these two values to 0 after they were previously non-zero will effectively
disable that specific facet of the pruning behaviour.

### gRPC API

At the time of this writing, it is proposed that CometBFT implement a full gRPC
interface ([\#81]). As such, we have several options when it comes to
implementing the data companion pull API:

1. Extend the proposed gRPC API from [\#81] to simply provide the additional
   data companion-specific endpoints. In order to meet the
   [requirements](#requirements), however, some of the endpoints will have to be
   protected by default. This is simpler for clients to interact with though,
   because they only need to interact with a single endpoint for all of their
   needs.
2. Implement a separate gRPC API on a different port to the standard gRPC
   interface. This allows for a clearer distinction between the standard and
   data companion-specific gRPC interfaces, but complicates the server and
   client interaction models.

Due to past experience of operators exposing _all_ RPC endpoints on a specific
port to the public internet, option 2 will be chosen here to minimize the
chances of this happening in future, even though it offers a slightly more
complicated experience for operators.

#### Block Service

The following `BlockService` will be implemented as part of [\#81], regardless
of whether or not this ADR is implemented. This API, therefore, needs to be more
generally useful than just for the purposes of the data companion. The minimal
API to support a data companion, however, is presented in this ADR.

```protobuf
syntax = "proto3";

package tendermint.services.block.v1;

import "tendermint/abci/types.proto";
import "tendermint/types/types.proto";
import "tendermint/types/block.proto";

// BlockService provides information about blocks.
service BlockService {
    // GetLatestHeight returns a stream of the latest block heights committed by
    // the network. This is a long-lived stream that is only terminated by the
    // server if an error occurs. The caller is expected to handle such
    // disconnections and automatically reconnect.
    rpc GetLatestHeight(GetLatestHeightRequest) returns (stream GetLatestHeightResponse) {}

    // GetBlockByHeight attempts to retrieve the block at a particular height.
    rpc GetBlockByHeight(GetBlockByHeightRequest) returns (GetBlockByHeightResponse) {}
}

message GetLatestHeightRequest {}

// GetLatestHeightResponse provides the height of the latest committed block.
message GetLatestHeightResponse {
    // The height of the latest committed block. Will be 0 if no data has been
    // committed yet.
    uint64 height = 1;
}

message GetBlockByHeightRequest {
    // The height of the block to get. Set to 0 to return the latest block.
    uint64 height = 1;
}

message GetBlockByHeightResponse {
    // Block data for the requested height.
    tendermint.types.Block block = 1;
}
```

#### Block Results Service

The following `BlockResultsService` service is proposed _separately_ to the
`BlockService`. There are several reasons as to why there are two separate gRPC
services to meet the companion's needs as opposed to just one:

1. The quantity of data stored by each is application-dependent, and coalescing
   the two types of data could impose significant overhead in some cases (this
   is primarily a justification for having separate RPC calls for each type of
   data).
2. Operators can enable/disable these services independently of one another.
3. The existing JSON-RPC API distinguishes between endpoints providing these two
   types of data (`/block` and `/block_results`), so users are already
   accustomed to this distinction.
4. Eventually, when we no longer need to store block results at all, we can
   simply deprecate the `BlockResultsService` without affecting clients who rely
   on `BlockService`.

```protobuf
syntax = "proto3";

package tendermint.services.block_results.v1;

// BlockResultsService provides information about the execution results for
// specific blocks.
service BlockResultsService {
    // GetBlockResults attempts to retrieve the execution results associated
    // with a block of a certain height.
    rpc GetBlockResults(GetBlockResultsRequest) returns (GetBlockResultsResponse)
}

message GetBlockResultsRequest {
    // The height of the block whose results are to be retrieved. Set to 0 to
    // return the latest block's results.
    uint64 height = 1;
}

message GetBlockResultsResponse {
    // The height associated with the block results.
    uint64 height = 1;

    // The contents of the FinalizeBlock response, which contain block execution
    // results.
    tendermint.abci.ResponseFinalizeBlock finalize_block_response = 2;
}
```

#### Pruning Service

This gRPC service is the only novel service proposed in this ADR, and
effectively gives a single external caller (e.g. a data companion) a say in how
the node prunes its data.

```protobuf
syntax = "proto3";

package tendermint.services.pruning.v1;

// PruningService provides privileged access to specialized pruning
// functionality on the CometBFT node to help control node storage.
service PruningService {
    // SetBlockRetainHeightRequest indicates to the node that it can safely
    // prune all block data up to the specified retain height.
    //
    // The lower of this retain height and that set by the application in its
    // Commit response will be used by the node to determine which heights' data
    // can be pruned.
    rpc SetBlockRetainHeight(SetBlockRetainHeightRequest) returns (SetBlockRetainHeightResponse)

    // GetBlockRetainHeight returns information about the retain height
    // parameters used by the node to influence block retention/pruning.
    rpc GetBlockRetainHeight(GetBlockRetainHeightRequest) returns (GetBlockRetainHeightResponse)

    // SetBlockResultsRetainHeightRequest indicates to the node that it can
    // safely prune all block results data up to the specified height.
    //
    // The node will always store the block results for the latest height to
    // help facilitate crash recovery.
    rpc SetBlockResultsRetainHeight(SetBlockResultsRetainHeightRequest) returns (SetBlockResultsRetainHeightResponse)

    // GetBlockResultsRetainHeight returns information about the retain height
    // parameters used by the node to influence block results retention/pruning.
    rpc GetBlockResultsRetainHeight(GetBlockResultsRetainHeightRequest) returns (GetBlockResultsRetainHeightResponse)
}

message SetBlockRetainHeightRequest {
    uint64 height = 1;
}

message SetBlockRetainHeightResponse {}

message GetBlockRetainHeightRequest {}

message GetBlockRetainHeightResponse {
    // The retain height set by the application.
    uint64 app_retain_height = 1;

    // The retain height set via the pruning service (e.g. by the data
    // companion) specifically for blocks.
    uint64 pruning_service_retain_height = 2;
}

message SetBlockResultsRetainHeightRequest {
    uint64 height = 1;
}

message SetBlockResultsRetainHeightResponse {}

message GetBlockResultsRetainHeightRequest {}

message GetBlockResultsRetainHeightResponse {
    // The retain height set by the pruning service (e.g. by the data
    // companion) specifically for block results.
    uint64 pruning_service_retain_height = 1;
}
```

With this API design, it is technically possible for an integrator to attach
multiple data companions to the node, but only one of their retain heights will
be considered by the node.

### Configuration

The following configuration file update is proposed to support the data
companion API.

```toml
#
# This is the envisaged configuration section for the gRPC API that will be
# introduced as part of https://github.com/cometbft/cometbft/issues/81
# (Still a WIP)
#
[grpc]

# The host/port on which to expose non-sensitive gRPC endpoints.
laddr = "tcp://localhost:26654"

#
# Configuration for sensitive gRPC endpoints, which should **never** be exposed
# to the public internet.
#
[grpc.sensitive]
# The host/port on which to expose sensitive gRPC endpoints.
laddr = "tcp://localhost:26655"

#
# Configuration specifically for the gRPC pruning service, which is considered a
# sensitive service.
#
[grpc.sensitive.pruning_service]

# Only controls whether the pruning service is accessible via the gRPC API - not
# whether a previously set pruning service retain height is honoured by the
# node.
#
# To disable the influence of previously set pruning service retain height(s) on
# node pruning, this endpoint should be enabled and the relevant pruning service
# retain heights should be set to 0.
#
# Disabled by default.
enabled = false
```

### Metrics

The following metrics are proposed to be added to monitor the health of the
interaction between a node and its data companion:

- `grpc_pruning_service_retain_height` - The current retain height as requested
  by the pruning service. This can give operators insight into how the data
  companion is affecting pruning.

Other metrics may be proposed as part of the non-sensitive gRPC API that could
assist operators in understanding the health of the interaction with the data
companion, but only if the data companion is the exclusive user of those APIs.

## Consequences

### Positive

- Facilitates offloading of data to an external service, which can be scaled
  independently of the node
  - Potentially reduces load on the node itself
  - Paves the way for eventually reducing the surface area of a node's exposed
    APIs
- Allows the data companion more leeway in reading the data it needs than the
  approach in [ADR 100][adr-100]
- Simpler implementation and fewer changes within the node than [ADR
  100][adr-100]

### Negative

- Increases system complexity slightly in the short-term
- If data companions are not correctly implemented and deployed (e.g. if a
  companion is attached to the same storage as the node, and/or if its retain
  height signalling is poorly handled), this could result in substantially
  increased storage usage

### Neutral

- Expands the overall API surface area of a node in the short-term

## References

- [ADR 100 - Data Companion Push API][adr-100]
- [\#81 - rpc: Add gRPC support][\#81]

<!--
TODO(thane): Replace GitHub links with relative Markdown file links once ADR-100 is merged.
-->
[adr-100-context]: https://github.com/cometbft/cometbft/blob/thane/adr-082-data-companion-api/docs/architecture/adr-100-data-companion-push-api.md#context
[adr-100]: https://github.com/cometbft/cometbft/blob/thane/adr-082-data-companion-api/docs/architecture/adr-100-data-companion-push-api.md
[adr-100-req]: https://github.com/cometbft/cometbft/blob/thane/adr-082-data-companion-api/docs/architecture/adr-100-data-companion-push-api.md#requirements
[adr-100-alt]: https://github.com/cometbft/cometbft/blob/thane/adr-082-data-companion-api/docs/architecture/adr-100-data-companion-push-api.md#alternative-approaches
[\#81]: https://github.com/cometbft/cometbft/issues/81
[abci-commit]: ../../spec/abci/abci++_methods.md#commit
