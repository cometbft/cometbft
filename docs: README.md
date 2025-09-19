# CometBFT
[![Version][version-badge]][version-url]
[![Go version][go-badge]][go-url]
[![Discord chat][discord-badge]][discord-url]
[![License][license-badge]][license-url]
[![Sourcegraph][sg-badge]][sg-url]

CometBFT is the most widely-adopted, battle-tested consensus engine in blockchain today. It is a [Byzantine Fault Tolerant (BFT)](https://en.wikipedia.org/wiki/Byzantine_fault) middleware that takes a state transition machine - written in any programming language - and securely replicates it on many machines.

CometBFT is highly performant and achieves speeds of up to 10k transactions per second (TPS). Its flagship feature, ABCI++ enables developers to add programmability and customization to every step of the consensus engine.

Developers can use CometBFT for BFT state machine replication of applications written in any programming language and development environment. This modularity gives developers flexibility to choose tools and technologies best suited for specific projects, improves maintainability, and delivers the scalability required for large-scale decentralized applications.

CometBFT is a fork of [Tendermint Core][tm-core] and implements the Tendermint consensus algorithm.

## Releases

Please do not depend on `main` as your production branch. Use [releases](https://github.com/cometbft/cometbft/releases) instead.

More on how releases are conducted can be found [here](./RELEASES.md).


## Minimum requirements

| CometBFT version | Requirement | Notes             |
|------------------|-------------|-------------------|
| main             | Go version  | Go 1.23 or higher |
| v0.38.x          | Go version  | Go 1.22 or higher |

### Install

See the [install guide](./docs/guides/install.md).

### Quick Start

- [Single node](./docs/guides/quick-start.md)
- [Local cluster using docker-compose](./docs/networks/docker-compose.md)

## Versioning

### Semantic Versioning

CometBFT uses [Semantic Versioning](http://semver.org/) to determine when and
how the version changes. 

To provide some stability to users of 0.X.X versions of CometBFT, the MINOR
version is used to signal breaking changes across CometBFT's API. This API
includes all publicly exposed types, functions, and methods in non-internal Go
packages as well as the types and methods accessible via the CometBFT RPC
interface.

Breaking changes to these public APIs will be documented in the CHANGELOG.

### Upgrades

In an effort to avoid accumulating technical debt, we do not
guarantee that breaking changes (i.e. bumps in the MINOR version) will work with
existing CometBFT blockchains. In these cases you will have to start a new
blockchain, or write something custom to get the old data into the new chain.
However, any bump in the PATCH version should be compatible with existing
blockchain histories.

For more information on upgrading, see [UPGRADING.md](./UPGRADING.md).

### Supported Versions

Currently supported versions include:

- v0.38.x: CometBFT v0.38 introduces ABCI 2.0, which implements the entirety of
  ABCI++


## Developer Community and Support

The issue list of this repo is exclusively for bug reports and feature requests. We have active, helpful communities on Discord, Telegram, and Slack.

**| Need Help? | Support & Community: [Discord](https://discord.com/invite/interchain) - [Telegram](https://t.me/CosmosOG) - [Talk to an Expert](https://cosmos.network/interest-form) - [Join the #Cosmos-tech Slack Channel](https://forms.gle/A8jawLgB8zuL1FN36) |**

## Security

To report a security vulnerability, see the Cosmos [bug bounty program](https://hackerone.com/cosmos). For examples of the kinds of bugs we're looking for, see [our security policy](SECURITY.md).

## Maintainers
[Cosmos Labs](https://cosmoslabs.io/) maintains the core components of the stack: Cosmos SDK, CometBFT, IBC, Cosmos EVM, and various developer tools and frameworks. In addition to developing and maintaining the Cosmos Stack, Cosmos Labs provides advisory and engineering services for blockchain solutions. [Get in touch with Cosmos Labs](https://www.cosmoslabs.io/contact).

Cosmos Labs is a wholly-owned subsidiary of the [Interchain Foundation](https://interchain.io/), the Swiss nonprofit responsible for treasury management, funding public goods, and supporting governance for Cosmos. 

The Cosmos Stack is supported by a robust community of open-source contributors. 

## Contributing

If you are interested in working on an issue, please comment on it, and take a look at the [contributing guidelines](./CONTRIBUTING.md). We welcome and appreciate community contributions! 

## Documentation and Resources

### Documentation
- [CometBFT Documentation](https://docs.cometbft.com/v0.38/)
- [CometBFT Specification](./spec/README.md)
- [Documentation](./docs/docs/01-ibc/01-overview.md)

### Cosmos Stack Libraries

- [Cosmos SDK](http://github.com/cosmos/cosmos-sdk) - A framework for building
  applications in Golang
- [The Inter-Blockchain Communication Protocol (IBC)](https://github.com/cosmos/ibc-go/) - A blockchain interoperability protocol that allows blockchains to transfer any type of data encoded in bytes.
- [Cosmos EVM](https://github.com/cosmos/evm) - Native EVM layer for Cosmos SDK chains. 

### Research

Below are links to the original Tendermint consensus algorithm and relevant
whitepapers which CometBFT will continue to build on.

- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.04938)
- [Master's Thesis on Tendermint](https://atrium.lib.uoguelph.ca/xmlui/handle/10214/9769)
- [Original Whitepaper: "Tendermint: Consensus Without Mining"](https://tendermint.com/static/docs/tendermint.pdf)



[bft]: https://en.wikipedia.org/wiki/Byzantine_fault_tolerance
[smr]: https://en.wikipedia.org/wiki/State_machine_replication
[Blockchain]: https://en.wikipedia.org/wiki/Blockchain
[version-badge]: https://img.shields.io/github/v/release/cometbft/cometbft.svg
[version-url]: https://github.com/cometbft/cometbft/releases/latest
[api-badge]: https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
[api-url]: https://pkg.go.dev/github.com/cometbft/cometbft
[go-badge]: https://img.shields.io/badge/go-1.22-blue.svg
[go-url]: https://github.com/moovweb/gvm
[discord-badge]: https://img.shields.io/discord/669268347736686612.svg
[discord-url]: https://discord.gg/interchain
[license-badge]: https://img.shields.io/github/license/cometbft/cometbft.svg
[license-url]: https://github.com/cometbft/cometbft/blob/main/LICENSE
[sg-badge]: https://sourcegraph.com/github.com/cometbft/cometbft/-/badge.svg
[sg-url]: https://sourcegraph.com/github.com/cometbft/cometbft?badge
[tests-url]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml
[tests-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.38.x
[tests-url-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.37.x
[tests-url-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.34.x
[tests-badge]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=main
[tests-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.38.x
[tests-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.37.x
[tests-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.34.x
[lint-badge]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=main
[lint-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.34.x
[lint-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.37.x
[lint-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.38.x
[lint-url]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml
[lint-url-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.34.x
[lint-url-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.37.x
[lint-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.38.x
[tm-core]: https://github.com/tendermint/tendermint
