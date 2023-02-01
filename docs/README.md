---
title: CometBFT Documentation
description: CometBFT is a blockchain application platform.
footer:
  newsletter: false
---

# CometBFT

Welcome to the CometBFT documentation!

CometBFT is a blockchain application platform; it provides the equivalent
of a web-server, database, and supporting libraries for blockchain applications
written in any programming language. Like a web-server serving web applications,
CometBFT serves blockchain applications.

<<<<<<< HEAD
More formally, Tendermint Core performs Byzantine Fault Tolerant (BFT) State
Machine Replication (SMR) for arbitrary deterministic, finite state machines.
For more background, see [What is
Tendermint?](introduction/what-is-tendermint.md).
=======
More formally, CometBFT performs Byzantine Fault Tolerant (BFT)
State Machine Replication (SMR) for arbitrary deterministic, finite state machines.
For more background, see [What is CometBFT?](introduction/what-is-cometbft.md).
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))

To get started quickly with an example application, see the [quick start
guide](introduction/quick-start.md).

<<<<<<< HEAD
To learn about application development on Tendermint, see the [Application
Blockchain
Interface](https://github.com/tendermint/tendermint/tree/v0.34.x/spec/abci).

For more details on using Tendermint, see the respective documentation for
[Tendermint Core](tendermint-core/), [benchmarking and monitoring](tools/), and
[network deployments](networks/).

To find out about the Tendermint ecosystem you can go
[here](https://github.com/tendermint/awesome#ecosystem). If you are a project
that is using Tendermint you are welcome to make a PR to add your project to the
list.

## Contribute

To contribute to the documentation, see [this
file](https://github.com/tendermint/tendermint/blob/main/docs/DOCS_README.md)
for details of the build process and considerations when making changes.
=======
To learn about application development on CometBFT, see the [Application Blockchain Interface](https://github.com/cometbft/cometbft/tree/main/spec/abci).

For more details on using CometBFT, see the respective documentation for
[CometBFT internals](core/), [benchmarking and monitoring](tools/), and [network deployments](networks/).

## Contribute

To recommend a change to the documentation, please submit a PR. Each major
release's documentation is housed on the corresponding release branch, e.g. for
the v0.34 release series, the documentation is housed on the `v0.34.x` branch.

When submitting changes that affect all releases, please start by submitting a
PR to the docs on `main` - this will be backported to the relevant release
branches. If a change is exclusively relevant to a specific release, please
target that release branch with your PR.

Changes to the documentation will be reviewed by the team and, if accepted and
merged, published to <https://docs.cometbft.com> for the respective version(s).

The build process for the documentation is housed in the [CometBFT documentation
repository](https://github.com/cometbft/cometbft-docs).
>>>>>>> 98838143f (Rename Tendermint to CometBFT in /docs (#197))
