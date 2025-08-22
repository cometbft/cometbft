<!-- NB: Ensure that all hyperlinks in this doc are absolute URLs, not relative
ones, as this doc gets published to the Buf registry and relative URLs will fail
to resolve. -->
# CometBFT v0.38.x Protocol Buffers Definitions

This is the set of [Protobuf][protobuf] definitions of types used by various
parts of [CometBFT]:

- The [Application Blockchain Interface][abci] (ABCI), especially in the context
  of _remote_ applications.
- The P2P layer, in how CometBFT nodes interact with each other over the
  network.
- In interaction with remote signers ("privval").
- The RPC, in that the native JSON serialization of certain Protobuf types is
  used when accepting and responding to RPC requests.
- The storage layer, in how data is serialized to and deserialized from on-disk
  storage.

The canonical Protobuf definitions live in the `proto` folder of the relevant
release branch of CometBFT. These definitions are published to the [Buf
registry][buf] for integrators' convenience.

## How are `cometbft` Protobuf definitions versioned?

At present, the canonical source of Protobuf definitions for all CometBFT v0.x
releases is on each respective release branch. Each respective release's
Protobuf definitions are also, for convenience, published to a corresponding
branch in the `cometbft/cometbft` Buf repository.

| CometBFT version | Canonical Protobufs                         | Buf registry                              |
|------------------|---------------------------------------------|-------------------------------------------|
| v0.39.x          | [v0.39.x Protobuf definitions][v039-protos] | [Buf repository v0.39.x branch][v039-buf] |

[protobuf]: https://protobuf.dev/
[CometBFT]: https://github.com/cometbft/cometbft
[abci]: https://github.com/cometbft/cometbft/tree/main/spec/abci
[buf]: https://buf.build/cometbft/cometbft
[\#1330]: https://github.com/cometbft/cometbft/issues/1330
[v039-protos]: https://github.com/cometbft/cometbft/tree/v0.39.x/proto
[v039-buf]: https://buf.build/cometbft/cometbft/docs/v0.39.x
