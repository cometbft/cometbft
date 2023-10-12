---
order: 1
parent:
    title: Introduction
    order: 1
---

# Introduction

A proposal was made in
[ADR-101](https://github.com/cometbft/cometbft/blob/6f2590df767be4c1824f1cc4070a647c417e6e75/docs/architecture/adr-101-data-companion-pull-api.md)
to introduce new gRPC endpoints that can be used by an external application to fetch data from the node and to control
which data is pruned by the node.

The Data Companion pruning service allows users to keep only the necessary data on the node,
enabling more efficient storage management and improved performance of the node. With this new service, users can have
greater control over their pruning mechanism and therefore better ability to optimize the node's storage.

The new pruning service allows granular control of what can be pruned such as blocks and state, ABCI results (if enabled), block
indexer data and transaction indexer data.

By also using the new gRPC services, it's possible now to retrieve data from the node, such as `block` and `block results`
in a more efficient way.

The [gRPC services](./grpc.md) document provides practical information and insights that will guide you through the
process of using these services in order to create a Data Companion service.

Note that this version of CometBFT (v0.38) already includes a gRPC service
(`rpc/grpc/api.go`) that is considered legacy code and will be removed in future
releases. In case you need to use the legacy gRPC endpoints, make sure that they
have a different URL than the gRPC services described in this document.
