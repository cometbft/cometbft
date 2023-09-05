---
order: 1
parent:
    title: Introduction
    order: 1
---

# Introduction

In [ADR-101](https://github.com/cometbft/cometbft/blob/thane/adr-084-data-companion-pull-api/docs/architecture/adr-101-data-companion-pull-api.md),
it is proposed that a Data Companion be introduced as a solution to enable external applications to have an impact on the
pruning of data on a node. The Data Companion works by leveraging the newly introduced gRPC APIs to control the [pruning](./pruning.md)
process on the node.

By using the Data Companion, users can ensure that only the necessary data is kept on the node, while unnecessary data is efficiently pruned.
This allows for more efficient storage management, as well as improved overall performance of the node.
With the Data Companion, users can have greater control over their data and can optimize their storage usage to meet their specific needs.

To assist you in implementing a Data Companion Pull API, we recommend checking out the informative and constructive
[Getting Started](./getting-started.md) document. This resource offers valuable insights and practical information that
will guide you through the process, from start to finish. By following the best practices outlined in the guide, you can
ensure that your implementation is secure, efficient, and meets all necessary requirements.
