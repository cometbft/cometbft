# ADR 105: gRPC API

## Changelog

- 2023-05-16: First draft (@thanethomson)

## Status

Accepted | Rejected | Deprecated | Superseded by

Tracking issue: [\#81]

## Context

There has been discussion over the years as to which type of RPC interface would
be preferable for Tendermint Core, and now CometBFT, to offer to integrators.
[ADR 057][adr-057] captures some pros and cons of continuing to support the
JSON-RPC API versus implementing a gRPC API. Previously it was decided to remove
the gRPC API from Tendermint Core (see [tendermint/tendermint\#7121] and
[tendermint/tendermint\#9683]).

After discussion with users, and in considering the implementation of [ADR
101][adr-101] (the data companion pull API), a decision has been taken to
implement a gRPC API _in addition to_ the JSON-RPC API.

The current plan is to implement this gRPC API after v0.38 after deprecating and
subsequently removing the existing gRPC API (which only provides a
`BroadcastService` with a single method). It is also envisaged that, once it is
feasible to provide the RPC service independently of the node itself (see [ADR
102][adr-102]), the JSON-RPC API on the node itself could eventually be
deprecated and removed.

## Alternative Approaches

The primary alternative approach involves continuing to only provide support
for, and potentially evolve, the JSON-RPC API. This API currently exposes many
data types in non-standard and rather complex ways, making it difficult to
implement clients. As per [ADR 075][adr-075], it also does not conform fully to
the JSON-RPC 2.0 specification, further increasing client implementation
complexity.

## Decision

Implement gRPC services corresponding to a minimal subset of the currently
exposed [JSON-RPC endpoints][rpc-docs]. This set of services can always be
expanded over time according to user needs, but once released it is hard to
deprecate and remove such APIs.

## Detailed Design

### Services

The initial services to be exposed via gRPC are informed by [Penumbra's
`TendermintProxyService`][penumbra-proxy-svc], as well as the needs of the data
companion API proposed in [ADR 101][adr-101]. Additional services can be rolled
out additively in subsequent releases of CometBFT.

Services are roughly organized by way of their respective domain. The details of
each service, e.g. request/response types and precise Protobuf definitions, will
be worked out in the implementation.

- `VersionService` - A simple service that aims to be quite stable over time in
  order to be utilized by clients to establish the version of the software with
  which they are interacting (e.g. to pre-emptively determine compatibility).
  This could technically be part of the `NodeService`, but if the `NodeService`
  interface were to be modified, a new version of the service would need to be
  created, and all old versions would need to be maintained, since the
  `GetVersion` method needs to be quite stable.
  - `GetVersion` - Query the version of the software and protocols employed by
    the node (e.g. CometBFT, ABCI, block, P2P and application version).
- `NodeService` - Provides information about the node providing the gRPC
  interface.
  - `GetStatus` - Query the current node status, including node info, public
    key, latest block hash, app hash, block height and time.
- `TransactionService` - Facilitates broadcasting and querying of transactions.
  - `BroadcastAsync` - Broadcast a transaction asynchronously. Does not wait
    for the transaction to be validated via `CheckTx`, nor does it wait for the
    transaction to be committed.
  - `BroadcastSync` - Broadcast a transaction, but only return once `CheckTx`
    has been called on the transaction. Does not wait for the transaction to be
    committed.
  - `GetByHash` - Fetch a transaction by way of its hash.
- `ApplicationService` - Provides a proxy interface through which to access the
  application being run by the node (via ABCI).
  - `Query` - Submit a query directly to the application via ABCI.
- `BlockService` - Provides information about blocks.
  - `GetByHeight` - Fetch the block associated with a particular height.
- `BlockResultsService` - Provides information about block execution results.
  - `GetByHeight` - Fetch the block results associated with a particular height.

### Service Versioning

Every service will be versioned, for example:

- `VersionService` will have its Protobuf definition under
  `tendermint.services.version.v1` (or `cometbft.services.version.v1` after
  [\#94] has been implemented)
- `NodeService` will have its definitions under `tendermint.services.node.v1`
- `TransactionService` will have its definitions under
  `tendermint.services.transaction.v1`
- etc.

The general approach to versioning our Protobuf definitions is captured in [ADR
103][adr-103].

### Go API

#### Server

The following Go API is proposed for constructing the gRPC server to allow for
ease of construction within the node, and configurability for users who have
forked CometBFT.

```go
package server

type Option func(*serverBuilder)

// WithVersionService enables the version service on the CometBFT gRPC server.
//
// (Similar methods should be provided for every other service that can be
// exposed via the gRPC interface)
func WithVersionService() Option {
    // ...
}

// WithGRPCOption allows one to specify Google gRPC server options during the
// construction of the CometBFT gRPC server.
func WithGRPCOption(opt grpc.ServerOption) Option {
    // ...
}

// Serve constructs and runs a CometBFT gRPC server using the given listener and
// options.
func Serve(listener net.Listener, opts ...Option) error {
    // ...
}
```

#### Client

For convenience, a Go client API should be provided for use within the E2E
testing framework.

```go
package client

type Option func(*clientBuilder)

// Client defines the full client interface for interacting with a CometBFT node
// via its gRPC interface.
type Client interface {
    ApplicationServiceClient
    BlockResultsServiceClient
    BlockServiceClient
    NodeServiceClient
    TransactionServiceClient
    VersionServiceClient
}

// WithInsecure disables transport security for the underlying client
// connection.
//
// A shortcut for using grpc.WithTransportCredentials and
// insecure.NewCredentials from google.golang.org/grpc.
func WithInsecure() Option {
    // ...
}

// WithGRPCDialOption allows passing lower-level gRPC dial options through to
// the gRPC dialer when creating the client.
func WithGRPCDialOption(opt ggrpc.DialOption) Option {
    // ...
}

// New constructs a client for interacting with a CometBFT node via gRPC.
//
// Makes no assumptions about whether or not to use TLS to connect to the given
// address. To connect to a gRPC server without using TLS, use the WithInsecure
// option.
//
// To connect to a gRPC server with TLS, use the WithGRPCDialOption option with
// the appropriate gRPC credentials configuration. See
// https://pkg.go.dev/google.golang.org/grpc#WithTransportCredentials
func New(ctx context.Context, addr string, opts ...Option) (Client, error) {
    // ...
}
```

## Consequences

### Positive

- Protocol buffers provide a relatively simple, standard way of defining RPC
  interfaces across languages.
- gRPC service definitions can be versioned and published via the [Buf Schema
  Registry][bsr] (BSR) for easy consumption by integrators.

### Negative

- Only programming languages with reasonable gRPC support will be able to
  integrate with the gRPC API (although most major languages do have such
  support).
- Increases node complexity in the short-term (until such time that the JSON-RPC
  API is definitively extracted and moved outside of the node).

<!--
TODO: Replace ADR 101/102-related PR links with direct links to the ADRs once
      merged.
-->
[\#81]: https://github.com/cometbft/cometbft/issues/81
[\#94]: https://github.com/cometbft/cometbft/issues/94
[adr-057]: ./tendermint-core/adr-057-RPC.md
[tendermint/tendermint\#7121]: https://github.com/tendermint/tendermint/pull/7121
[tendermint/tendermint\#9683]: https://github.com/tendermint/tendermint/pull/9683
[adr-101]: https://github.com/cometbft/cometbft/pull/82
[adr-102]: https://github.com/cometbft/cometbft/pull/658
[adr-103]: ./adr-103-proto-versioning.md
[adr-075]: ./tendermint-core/adr-075-rpc-subscription.md
[rpc-docs]: https://docs.cometbft.com/v0.37/rpc/
[penumbra-proxy-svc]: https://buf.build/penumbra-zone/penumbra/docs/main:penumbra.client.v1alpha1#penumbra.client.v1alpha1.TendermintProxyService
[bsr]: https://buf.build/explore
