# CometBFT

[Byzantine-Fault Tolerant][bft] [State Machine Replication][smr]. Or
[Blockchain], for short.

[![Version][version-badge]][version-url]
[![API Reference][api-badge]][api-url]
[![Go version][go-badge]][go-url]
[![License][license-badge]][license-url]
[![Sourcegraph][sg-badge]][sg-url]

[![Discord chat][discord-badge]][discord-url]

| Branch  | Tests                                          | Linting                                     |
|---------|------------------------------------------------|---------------------------------------------|
| main    | [![Tests][tests-badge]][tests-url]             | [![Lint][lint-badge]][lint-url]             |
| v1.x    | [![Tests][tests-badge-v1x]][tests-url-v1x]     | [![Lint][lint-badge-v1x]][lint-url-v1x]     |
| v0.38.x | [![Tests][tests-badge-v038x]][tests-url-v038x] | [![Lint][lint-badge-v038x]][lint-url-v038x] |

CometBFT is a Byzantine Fault Tolerant (BFT) middleware that takes a
state transition machine - written in any programming language - and securely
replicates it on many machines. In modular blockchain terminology,
CometBFT can be thought of as a sequencer layer and is indeed used in
modern decentralized (shared) sequencer implementations.

CometBFT is the canonical implementation of the Tendermint consensus algorithm and is a
primary building block for the [Interchain Stack](https://interchain.io/). Historically,
CometBFT originated as a fork of [Tendermint Core][tm-core] in early 2023
(announcement [here][comet-announcement]) and since then it diverged significantly by adopting modern features such as [PBTS][pbts] or [ABCI v2][abci-v2]. CometBFT provides [optimistic responsiveness][optimistic-responsive] guarantees.

For protocol details, please take a look at the [CometBFT Specification](./spec/README.md).

For a detailed analysis of the Tendermint consensus protocol, including safety and liveness
proofs, read our paper, "[The latest gossip on BFT
consensus](https://arxiv.org/abs/1807.04938)".

For general links, including communications and announcements: [![Linktree][linktree-badge] linktr.ee/cometbft][linktree-url]

## Documentation

Complete documentation can be found on the
[website](https://docs.cometbft.com/).

## Releases

Please do not depend on `main` as your production branch, as it may receive
significant breaking changes at any time. Use
[releases](https://github.com/cometbft/cometbft/releases) instead.

If you intend to run CometBFT in production, we're happy to help. To contact us,
in order of preference:

- [Create a new discussion on
  GitHub](https://github.com/cometbft/cometbft/discussions)
- Reach out to us via [Telegram](https://t.me/CometBFT)
- [Join the Cosmos Network Discord](https://discord.gg/interchain) and
  discuss in
  [`#cometbft`](https://discord.com/channels/669268347736686612/1069933855307472906)

More on how releases are conducted can be found [here](./RELEASES.md).

## Support Policy

CometBFT aligns with other components of the [Interchain Stack](https://interchain.io/)
and we offer long-term support (LTS) guarantees for certain releases. The
complete End of Life (EOL) schedule, LTS plans, and the general support policy is
in documented and regularly updated in the
discussion [Support policy for CometBFT releases #590](https://github.com/cometbft/cometbft/discussions/590).

## Security

Please see [SECURITY.md](./SECURITY.md).

## Minimum requirements

| CometBFT version | Requirement | Version        | Tested with  |
|------------------|-------------|----------------|--------------|
| main             | Go version  | 1.23 or higher | up to 1.23.6 |
| v1.x             | Go version  | 1.23 or higher | up to 1.23.1 |
| v0.38.x          | Go version  | 1.22 or higher | up to 1.22   |

### Install

See the [install guide](docs/tutorials/install.md).

### Quick Start

- [Single node](docs/tutorials/quick-start.md)

## Contributing

Please abide by the [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

Before contributing to the project, please take a look at the [contributing
guidelines](CONTRIBUTING.md) and the [style guide](STYLE_GUIDE.md). You may also
find it helpful to read the [specifications](./spec/README.md), and familiarize
yourself with our [Architectural Decision Records
(ADRs)](docs/references/architecture/README.md) and [Request For Comments
(RFCs)](docs/references/rfc/README.md).

## Versioning

As of v1, CometBFT uses the following approach to versioning:

- **Major version** bumps, such as v1.0.0 to v2.0.0, would generally involve
  changes that _force_ users to perform a coordinated upgrade in order to use
  the new version, such as protocol-breaking changes (e.g. changes to how block
  hashes are computed and thus what the network considers to be "valid blocks",
  or how the consensus protocol works, or changes that affect network-level
  compatibility between nodes, etc.).
- **Minor version** bumps, such as v1.1.0 to v1.2.0, are reserved for rolling
  out new features or substantial changes that do not force a coordinated
  upgrade (i.e. not protocol-breaking), but could potentially break Go APIs.
- **Patch version** bumps, such as v1.0.0 to v1.0.1, are reserved for
  bug/security fixes that are not protocol- or Go API-breaking.

### Upgrades

We do not guarantee compatibility between major releases of CometBFT. Minor
releases of the same major release series (v1.1, v1.2, etc.) should, unless
otherwise specified, be compatible with each other. Patch releases of the same
minor release series (v1.0.1, v1.0.2, etc.) are guaranteed to be compatible with
each other.

For more detailed information on upgrading from one version to another, see
[UPGRADING.md](./UPGRADING.md).

### Supported Versions

Because we are a small core team, we have limited capacity to ship patch
updates, including security updates. Consequently, we strongly recommend keeping
CometBFT up-to-date. Upgrading instructions can be found in
[UPGRADING.md](./UPGRADING.md).

Currently supported versions include:

- v1.x: Currently in pre-release with no guarantees as to API stability until a
  release candidate is cut. See [RELEASES.md](./RELEASES.md) for details on our
  process as to API stability guarantees that can be expected of CometBFT
  pre-releases.
- v0.38.x: CometBFT v0.38 introduces ABCI 2.0, which implements the entirety of
  ABCI++

## Resources

### Libraries

- [Cosmos SDK](http://github.com/cosmos/cosmos-sdk): A framework for building
  high-value public blockchain applications in Go
- [Tendermint in Rust](https://github.com/informalsystems/tendermint-rs)
- [ABCI Tower](https://github.com/penumbra-zone/tower-abci)

### Applications

- [Cosmos Hub](https://hub.cosmos.network/)
- [Celestia](https://celestia.org/)
- [Anoma](https://anoma.network/)
- [Vocdoni](https://developer.vocdoni.io/)

### Research

Below are links to the original Tendermint consensus algorithm and relevant
whitepapers, which CometBFT will continue to build on.

- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.04938)
- [Master's Thesis on Tendermint](https://atrium.lib.uoguelph.ca/xmlui/handle/10214/9769)
- [Original Whitepaper: "Tendermint: Consensus Without Mining"](https://tendermint.com/static/docs/tendermint.pdf)

## Join us

CometBFT is currently maintained by [Interchain Inc.](https://medium.com/the-interchain-foundation/cosmos-is-expanding-skip-joins-the-interchain-foundation-cfd346551dda). 

Funding for CometBFT development comes primarily from the [Interchain
Foundation](https://interchain.io), a Swiss non-profit.

[bft]: https://en.wikipedia.org/wiki/Byzantine_fault_tolerance
[smr]: https://en.wikipedia.org/wiki/State_machine_replication
[optimistic-responsive]: https://informal.systems/blog/tendermint-responsiveness
[Blockchain]: https://en.wikipedia.org/wiki/Blockchain
[version-badge]: https://img.shields.io/github/v/release/cometbft/cometbft.svg
[version-url]: https://github.com/cometbft/cometbft/releases/latest
[api-badge]: https://pkg.go.dev/badge/github.com/cometbft/cometbft.svg
[api-url]: https://pkg.go.dev/github.com/cometbft/cometbft
[go-badge]: https://img.shields.io/badge/go-1.21-blue.svg
[go-url]: https://github.com/moovweb/gvm
[discord-badge]: https://img.shields.io/discord/669268347736686612.svg
[discord-url]: https://discord.gg/interchain
[license-badge]: https://img.shields.io/github/license/cometbft/cometbft.svg
[license-url]: https://github.com/cometbft/cometbft/blob/main/LICENSE
[sg-badge]: https://sourcegraph.com/github.com/cometbft/cometbft/-/badge.svg
[sg-url]: https://sourcegraph.com/github.com/cometbft/cometbft?badge
[tests-url]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml
[tests-url-v1x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av1.x
[tests-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.38.x
[tests-url-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.37.x
[tests-url-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml?query=branch%3Av0.34.x
[tests-badge]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=main
[tests-badge-v1x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v1.x
[tests-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.38.x
[tests-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.37.x
[tests-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/tests.yml/badge.svg?branch=v0.34.x
[lint-badge]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=main
[lint-badge-v034x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.34.x
[lint-badge-v037x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.37.x
[lint-badge-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v0.38.x
[lint-badge-v1x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml/badge.svg?branch=v1.x
[lint-url]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml
[lint-url-v038x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av0.38.x
[lint-url-v1x]: https://github.com/cometbft/cometbft/actions/workflows/lint.yml?query=branch%3Av1.x
[tm-core]: https://github.com/tendermint/tendermint
[pbts]: https://docs.cometbft.com/v1.0/explanation/core/proposer-based-timestamps
[abci-v2]: https://docs.cometbft.com/v1.0/spec/abci/
[comet-announcement]: https://informal.systems/blog/cosmos-meet-cometbft
[linktree-url]: https://linktr.ee/cometbft
[linktree-badge]: https://www.google.com/s2/favicons?domain=https://linktr.ee/
