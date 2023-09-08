---
order: 1
parent:
    title: Introduction
    order: 1
---

# Introduction

A proposal was made in
[ADR-101](https://github.com/cometbft/cometbft/blob/thane/adr-084-data-companion-pull-api/docs/architecture/adr-101-data-companion-pull-api.md)
to introduce a new gRPC API that can be used by an external application to control which data is pruned by the node.

The Data Companion Pull API allows users to keep only the necessary data on the node,
enabling more efficient storage management and improved performance of the node. With this new API, users can have
greater control over their pruning mechanism and optimize their storage usage to meet their specific needs.

The new services allow granular control of what can be pruned such as blocks and state, ABCI results (if enabled), block
indexer data and transaction indexer data.

If you need help implementing a Data Companion that uses the Data Companion Pull API, you can refer to the guide titled
[Quick Start - Creating a Data Companion for CometBFT](./quick-start.md). This resource provides practical information
and insights that will guide you through the process of creating a Data Companion.
