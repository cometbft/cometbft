# dYdX Fork of CometBFT

This is a lightweight fork of CometBFT. The current version of the forked code resides on the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch).

## Making Changes to the Fork

1. Open a PR against the current default branch (i.e. `dydx-fork-v0.37.0`).
2. Get approval, and merge.
3. After merging, update the `v4` repository's `go.mod`, and `go.sum` files with your merged `$COMMIT_HASH`.
4. (In `dydxprotocol/v4`) `go mod edit -replace github.com/cometbft/cometbft=github.com/dydxprotocol/cometbft@$COMMIT_HASH`
5. (In `dydxprotocol/v4`) `go mod tidy`
6. (In `dydxprotocol/v4`) update package references in `mocks/Makefile`. See [here](https://github.com/dydxprotocol/v4/pull/848) for an example.
7. Open a PR in `dydxprotocol/v4` to bump the version of the fork.

## Fork maintenance

We'd like to keep the `main` branch up to date with `cometbft/cometbft`. You can utilize GitHub's [sync fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork) button to accomplish this. ⚠️ Please only use this on the `main` branch, not on the fork branches as it will discard our commits.⚠️

Note that this doesn't pull in upstream tags, so in order to do this follow these steps:
1. `git fetch upstream`
2. `git push --tags`

## dYdX Proto maintenance

In order to support some of our custom functionality, we require some dydx protobuf files to be copied into this repository. Currently, the source of truth for protos is in `dydxprotocol/v4`, and any changes that require updates to any of the protos in this repository should be sync'd over as well. Here are steps for updating and compiling the protos here.

1. Modify the protos in `proto/dydxcometbft`.
2. `make proto-gen`

Note that the protos cannot be copied over directly. golang protobufs share a global namespace, and we have changed the package name slightly to avoid a name clash.

We've also included a new dependency in the `buf.yaml` file for `"cosmos_proto/cosmos.proto"`. If this needs to be updated, run `buf build`. For more information, read [here](https://github.com/dydxprotocol/v4/tree/main/proto#update-protos).

In the future, we will aim to have a single source of truth for protos.

## Updating CometBFT to new versions

When a new version of CometBFT is published, we may want to adopt the changes in our fork. This process can be somewhat tedious, but below are the recommended steps to accomplish this.

1. Ensure the `main` branch and all tags are up to date by following the steps above in "Fork maintenance".
2. Create a new branch off the desired CometBFT commit using tags. `git checkout -b dydx-fork-$VERSION <CometBFT repo's tag name>`. The new branch should be named something like `dydx-fork-$VERSION` where `$VERSION` is the version of CometBFT being forked (should match the CometBFT repo's tag name). i.e. `dydx-fork-v0.37.0`.
3. Push the new branch.
4. Open a PR which cherry-picks each commit in the current default branch, in order, on to the new `dydx-fork-$VERSION` branch (note: you may want to consider creating multiple PRs for this process if there are difficulties or merge conflicts). For example, `git cherry-pick <commit hash>`.
5. Get approval, and merge.
6. Update `dydxprotocol/v4` by following the steps in "Making Changes to the fork" above.
7. Set `dydx-fork-$VERSION` as the [default branch](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-branches-in-your-repository/changing-the-default-branch) in this repository.

# CometBFT

[Byzantine-Fault Tolerant][bft] [State Machine Replication][smr]. Or
[Blockchain], for short.

[![Version][version-badge]][version-url]
[![API Reference][api-badge]][api-url]
[![Go version][go-badge]][go-url]
[![Discord chat][discord-badge]][discord-url]
[![License][license-badge]][license-url]
[![Sourcegraph][sg-badge]][sg-url]

| Branch  | Tests                                    | Linting                               |
|---------|------------------------------------------|---------------------------------------|
| main    | [![Tests][tests-badge]][tests-url]       | [![Lint][lint-badge]][lint-url]       |
| v0.37.x | [![Tests][tests-badge-v037x]][tests-url] | [![Lint][lint-badge-v037x]][lint-url] |
| v0.34.x | [![Tests][tests-badge-v034x]][tests-url] | [![Lint][lint-badge-v034x]][lint-url] |

CometBFT is a Byzantine Fault Tolerant (BFT) middleware that takes a
state transition machine - written in any programming language - and securely
replicates it on many machines.

It is a fork of [Tendermint Core][tm-core] and implements the Tendermint
consensus algorithm.

For protocol details, refer to the [CometBFT Specification](./spec/README.md).

For detailed analysis of the consensus protocol, including safety and liveness
proofs, read our paper, "[The latest gossip on BFT
consensus](https://arxiv.org/abs/1807.04938)".

## Documentation

Complete documentation can be found on the
[website](https://docs.cometbft.com/).

## Releases

Please do not depend on `main` as your production branch. Use
[releases](https://github.com/cometbft/cometbft/releases) instead.

We haven't released v1.0 yet
since we are making breaking changes to the protocol and the APIs. See below for
more details about [versioning](#versioning).

In any case, if you intend to run CometBFT in production, we're happy to help.

To contact us, you can also
[join the chat](https://discord.com/channels/669268347736686612/669283915743232011).

More on how releases are conducted can be found [here](./RELEASES.md).

## Security

To report a security vulnerability, see our [bug bounty
program](https://hackerone.com/cosmos). For examples of the kinds of bugs we're
looking for, see [our security policy](SECURITY.md).

## Minimum requirements

| CometBFT version | Requirement | Notes             |
|------------------|-------------|-------------------|
| v0.34.x          | Go version  | Go 1.19 or higher |
| v0.37.x          | Go version  | Go 1.20 or higher |
| main             | Go version  | Go 1.20 or higher |

### Install

See the [install guide](./docs/guides/install.md).

### Quick Start

- [Single node](./docs/guides/quick-start.md)
- [Local cluster using docker-compose](./docs/networks/docker-compose.md)

## Contributing

Please abide by the [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

Before contributing to the project, please take a look at the [contributing
guidelines](CONTRIBUTING.md) and the [style guide](STYLE_GUIDE.md). You may also
find it helpful to read the [specifications](./spec/README.md), and familiarize
yourself with our [Architectural Decision Records
(ADRs)](./docs/architecture/README.md) and [Request For Comments
(RFCs)](./docs/rfc/README.md).

## Versioning

### Semantic Versioning

CometBFT uses [Semantic Versioning](http://semver.org/) to determine when and
how the version changes. According to SemVer, anything in the public API can
change at any time before version 1.0.0

To provide some stability to users of 0.X.X versions of CometBFT, the MINOR
version is used to signal breaking changes across CometBFT's API. This API
includes all publicly exposed types, functions, and methods in non-internal Go
packages as well as the types and methods accessible via the CometBFT RPC
interface.

Breaking changes to these public APIs will be documented in the CHANGELOG.

### Upgrades

In an effort to avoid accumulating technical debt prior to 1.0.0, we do not
guarantee that breaking changes (i.e. bumps in the MINOR version) will work with
existing CometBFT blockchains. In these cases you will have to start a new
blockchain, or write something custom to get the old data into the new chain.
However, any bump in the PATCH version should be compatible with existing
blockchain histories.

For more information on upgrading, see [UPGRADING.md](./UPGRADING.md).

### Supported Versions

Because we are a small core team, we have limited capacity to ship patch
updates, including security updates. Consequently, we strongly recommend keeping
CometBFT up-to-date. Upgrading instructions can be found in
[UPGRADING.md](./UPGRADING.md).

Currently supported versions include:

- v0.34.x: The CometBFT v0.34 series is compatible with the Tendermint Core
  v0.34 series
- v0.37.x: (release candidate)

## Resources

### Libraries

- [Cosmos SDK](http://github.com/cosmos/cosmos-sdk); A framework for building
  applications in Golang
- [Tendermint in Rust](https://github.com/informalsystems/tendermint-rs)
- [ABCI Tower](https://github.com/penumbra-zone/tower-abci)

### Applications

- [Cosmos Hub](https://hub.cosmos.network/)
- [Terra](https://www.terra.money/)
- [Celestia](https://celestia.org/)
- [Anoma](https://anoma.network/)
- [Vocdoni](https://docs.vocdoni.io/)

### Research

Below are links to the original Tendermint consensus algorithm and relevant
whitepapers which CosmosBFT will continue to build on.

- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.04938)
- [Master's Thesis on Tendermint](https://atrium.lib.uoguelph.ca/xmlui/handle/10214/9769)
- [Original Whitepaper: "Tendermint: Consensus Without Mining"](https://tendermint.com/static/docs/tendermint.pdf)

## Join us

CometBFT is currently maintained by [Informal
Systems](https://informal.systems). If you'd like to work full-time on CometBFT,
[we're hiring](https://informal.systems/careers)!

Funding for CometBFT development comes primarily from the [Interchain
Foundation](https://interchain.io), a Swiss non-profit. Informal Systems also
maintains [cometbft.com](https://cometbft.com).

[bft]: https://en.wikipedia.org/wiki/Byzantine_fault_tolerance
[smr]: https://en.wikipedia.org/wiki/State_machine_replication
[Blockchain]: https://en.wikipedia.org/wiki/Blockchain
[version-badge]: https://img.shields.io/github/v/release/cometbft/cometbft.svg
[version-url]: https://github.com/cometbft/cometbft/releases/latest
[api-badge]: https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
[api-url]: https://pkg.go.dev/github.com/cometbft/cometbft
[go-badge]: https://img.shields.io/badge/go-1.20-blue.svg
[go-url]: https://github.com/moovweb/gvm
[discord-badge]: https://img.shields.io/discord/669268347736686612.svg
[discord-url]: https://discord.gg/cosmosnetwork
[license-badge]: https://img.shields.io/github/license/cometbft/cometbft.svg
[license-url]: https://github.com/cometbft/cometbft/blob/main/LICENSE
[sg-badge]: https://sourcegraph.com/github.com/cometbft/cometbft/-/badge.svg
[sg-url]: https://sourcegraph.com/github.com/cometbft/cometbft?badge
[tests-url]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml
[tests-badge]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=main
[tests-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.37.x
[tests-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.34.x
[lint-badge]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=main
[lint-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.34.x
[lint-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.37.x
[lint-url]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml
[tm-core]: https://github.com/tendermint/tendermint
