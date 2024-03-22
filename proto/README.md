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

## Why does CometBFT use `tendermint` Protobuf definitions?

This is as a result of CometBFT being a fork of [Tendermint Core][tmcore] and
wanting to provide integrators with as painless a way as possible of
transitioning from Tendermint Core to CometBFT.

As of CometBFT v1, however, the project will transition to using and providing a
`cometbft` package of Protobuf definitions (see [\#1330]).

## How are `tendermint` Protobuf definitions versioned?

At present, the canonical source of Protobuf definitions for all CometBFT v0.x
releases is on each respective release branch. Each respective release's
Protobuf definitions are also, for convenience, published to a corresponding
branch in the `tendermint/tendermint` Buf repository.

| CometBFT version | Canonical Protobufs                         | Buf registry                              |
|------------------|---------------------------------------------|-------------------------------------------|
| v0.38.x          | [v0.38.x Protobuf definitions][v038-protos] | [Buf repository v0.38.x branch][v038-buf] |
| v0.37.x          | [v0.37.x Protobuf definitions][v037-protos] | [Buf repository v0.37.x branch][v037-buf] |
| v0.34.x          | [v0.34.x Protobuf definitions][v034-protos] | [Buf repository v0.34.x branch][v034-buf] |

[protobuf]: https://protobuf.dev/
[CometBFT]: https://github.com/cometbft/cometbft
[abci]: https://github.com/cometbft/cometbft/tree/main/spec/abci
[buf]: https://buf.build/tendermint/tendermint
[tmcore]: https://github.com/tendermint/tendermint
[\#1330]: https://github.com/cometbft/cometbft/issues/1330
[v034-protos]: https://github.com/cometbft/cometbft/tree/v0.34.x/proto
[v034-buf]: https://buf.build/tendermint/tendermint/docs/v0.34.x
[v037-protos]: https://github.com/cometbft/cometbft/tree/v0.37.x/proto
[v037-buf]: https://buf.build/tendermint/tendermint/docs/v0.37.x
[v038-protos]: https://github.com/cometbft/cometbft/tree/v0.38.x/proto
[v038-buf]: https://buf.build/tendermint/tendermint/docs/v0.38.x
