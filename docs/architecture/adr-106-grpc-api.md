# ADR 106: gRPC API

## Changelog

- 2024-03-27: Minor updates based on user feedback and ADR 101 implementation (@andynog)
- 2023-07-04: Expand available endpoints based on user feedback (@thanethomson)
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

Some services for this gRPC API have already been implemented as part of the [Data Companion Pull API implementation][adr-101-poc],
such as `Block`, `BlockResults` and `Version` services. Also the existing gRPC API (which only provides a
`BroadcastService` with a single method) was removed. These services will be available starting with the CometBFT`v1` release
(there was also a backport to an experimental `v0.38` release)

It is also envisaged that once it is
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
  - `GetHealth` - Lightweight mechanism to query the health of the node.
- `TransactionService` - Facilitates broadcasting and querying of transactions.
  - `BroadcastAsync` - Broadcast a transaction asynchronously. Does not wait for
    the transaction to be validated via `CheckTx`, nor does it wait for the
    transaction to be committed.
  - `BroadcastSync` - Broadcast a transaction, but only return once `CheckTx`
    has been called on the transaction. Does not wait for the transaction to be
    committed.
  - `GetByHash` - Fetch a transaction by way of its hash.
  - `Search` - Search for transactions with their results.
- `ApplicationService` - Provides a proxy interface through which to access the
  application being run by the node (via ABCI).
  - `Query` - Submit a query directly to the application via ABCI.
- `BlockService` - Provides information about blocks.
  - `GetLatestHeight` - Return a stream of latest block heights as new blocks
    are committed to the blockchain.
  - `GetByHeight` - Fetch the block associated with a particular height.
  - `GetHeaderByHeight` - Fetch the header associated with the block at a
    particular height.
  - `Search` - Search for blocks by way of block events.
- `BlockResultsService` - Provides information about block execution results.
  - `GetBlockResults` - Fetch the block results associated with a particular height.
- `ConsensusService` - Provides information about consensus.
  - `GetParams` - Fetch the consensus parameters for a particular height.
- `NetworkService` - Provides information about the blockchain network.
  - `GetGenesis` - Fetch paginated genesis data.
  - `GetPeers` - Fetch information about the peers to which the node is
    connected.

### Service Versioning

Every service will be versioned, for example:

- `VersionService` will have its Protobuf definition under
  `cometbft.services.version.v1`
- `NodeService` will have its Protobuf definition under `cometbft.services.node.v1`
- `TransactionService` will have its Protobuf definition under
  `cometbft.services.transaction.v1`
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

// Option is any function that allows for configuration of the gRPC server
// during its creation.
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
// via its gRPC.
type Client interface {
    ApplicationServiceClient
    BlockResultsServiceClient
    BlockServiceClient
    NodeServiceClient
    TransactionServiceClient
    VersionServiceClient

	// Close the connection to the server. Any subsequent requests will fail.
	Close() error
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
- Increases complexity maintaining multiple APIs (gRPC and JSON-RPC) in the short-term (until the JSON-RPC API is definitively extracted and moved outside the node).

[\#81]: https://github.com/cometbft/cometbft/issues/81
[\#94]: https://github.com/cometbft/cometbft/issues/94
[adr-057]: ./tendermint-core/adr-057-RPC.md
[tendermint/tendermint\#7121]: https://github.com/tendermint/tendermint/pull/7121
[tendermint/tendermint\#9683]: https://github.com/tendermint/tendermint/pull/9683
[adr-101]: https://github.com/cometbft/cometbft/pull/82
[adr-101-poc]: https://github.com/cometbft/cometbft/issues/816
[adr-102]: adr-102-rpc-companion.md
[adr-103]: ./adr-103-proto-versioning.md
[adr-075]: ./tendermint-core/adr-075-rpc-subscription.md
[rpc-docs]: https://docs.cometbft.com/v0.37/rpc/
[penumbra-proxy-svc]: https://buf.build/penumbra-zone/penumbra/docs/main:penumbra.util.tendermint_proxy.v1
[bsr]: https://buf.build/explore
