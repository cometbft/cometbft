# ADR-102: Standalone RPC (WIP)

## Changelog

- 2022-03-27: First draft (@andynog)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

This ADR proposes an architecture of a ***Standalone RPC*** solution implemented based on a Data Companion Pull API (proposed
in the ADR-101 [TODO: add reference]). This solution can run as a sidecar concurrently with the full node, and it is optional.

This ADR provides a reference implementation of a system that can be used to offload queryable data from a CometBFT
full node to a data companion that exposes the same endpoints as the regular RPC endpoints of a full CometBFT node,
which makes it easier for integrators of RPC clients such as client libraries and applications to switch to this
***Standalone RPC*** with as minimum effort as possible.

This architecture also make it possible to scale horizontally the querying capacity of a full node by running multiple
copies of the ***Standalone RPC*** server instances that can be behind a scalable load-balancer (e.g. Cloudflare)  and can serve
the data in a more reliable way.

## Alternative Approaches

Currently, there aren't any alternative solutions that are based on a data companion architecture. It is
expected that integrators and operators propose and design their own solutions based on their specific use-cases.

This ADR provides a reference implementation that can be used and adapted for individual use-cases.

## Decision

TBD

## Detailed Design

### Requirements

The target audience for this solution are operators and integrators that want to alleviate the load on their nodes by offloading
the queryable data requests to the **Standalone RPC**.

This solution shall meet the following requirements in order to provide real benefits to these users.

The **Standalone RPC** solution shall:

- Provide an ingestion service implemented as a data companion that can pull data from the node and store it on
its own storage (database)
- Provide its own storage (database) that can handle a high-load of reads and also have a good performance on inserting data.
- Implement it based on a schema that needs to be backwards compatible with the current CometBFT RPC. For the RPC v2, new schemas can
be defined to cater different future use cases.
- Leverage the existing data types from the CometBFT in order to support backwards compatibility
- Provide a scalable RPC that is backwards compatible with the existing CometBFT RPC.
- Do not enforce any breaking changes to the existing RPC.
- Ensure the responses returned by the RPC v1 endpoints are wire compatible with the existing CometBFT endpoints.
- Users should be able to only replace the URL for queries and no modifications will be needed for client libraries and
applications (at least for the RPC v1).
- Implement tests to verify backwards compatibility.
- Be able to handle multiple concurrent requests and return idempotent responses. If the number of requests increase,
the solution should provide a mechanism to handle the load (e.g. allow linear scaling of the server instances serving the requests).
- Provide good performance even when querying for larger datasets.
- Do not crash or panic if the querying demand is very high or large responses are being returned.
- Implement mechanisms to ensure large responses can be returned (e.g. pagination) by the RPC.
- Implement mechanisms to prevent and mitigate DDoS attacks, such as rate-limiting (this is not mandatory since it can
be offered by a load balancer).
- Optimized storage for an access pattern of higher read throughput than write throughput.
- Provide metrics to allow operators to properly monitor the services and infrastructure.
- Be implemented on its own repository and be self-contained (the CometBFT repository should not depend on the
Standalone RPC repository).
- Do not break any major CometBFT release and be compatible with older CometBFT releases such as v0.34.x or v0.37.x.
- Do not require changes on the Cosmos SDK repository or related projects (e.g. IBC).


It is *NOT* in the scope of the **Standalone RPC**:

- To provide RPC endpoints that are of a "write" nature, for example `broadcast_tx_*` endpoints should not be supported due to the complexity
  of the race conditions that might occur in a load balanced Standalone RPC environment.
- Provide an authentication mechanism for the RPC query endpoints

### Proposal

#### High-level architecture

![High-level architecture](/home/andy/go/src/github.com/cometbft/cometbft/docs/architecture/images/adr-102-architecture.png)

#### Full Node

A full node for a blockchain based on CometBFT

#### Ingest Service

Pulls data from the full node and stores in the database

#### Database

Stores the data fetched from the full node and provide data for the RPC server instance

#### RPC server instance

Serves the RPC requests and provide responses with data retrieved from its own local storage

- Planned Standalone RPC v1
  - /block
  - /block_results
  - ...


- Future Standalone RPC v2 - new endpoints
  - (TBD)

- Transactions that modify the state
  - /broadcast_tx_*

#### Load balancer

An external load-balancer service such as Cloudflare or AWS ELB, or a server running its own load balancer mechanism (e.g. nginx).

## Consequences

### Positive

- Alternative and optional standalone RPC that is more scalable and reliable with a higher query throughput.
- Less backpressure on the full node that is running consensus.
- Possibility for future additional endpoints (v2).
- Allow users to create better and faster indexers and analytics solutions.

### Negative

- Additional infrastructure complexity to set up and maintain.
- Additional infrastructure cost.

### Neutral

- Optional feature, users will only use it if needed.
- No privacy / security issues should arise since the data returned by the Standalone RPC will be the same
as the current RPC.

## References

- [Improve experience for integrators](https://github.com/cometbft/cometbft/issues/40)
- [ADR-101: Data Companions Pull API (tracking issue)](https://github.com/cometbft/cometbft/issues/574)
- [ADR-101: Data Companions Pull API (PR)](https://github.com/cometbft/cometbft/pull/82)
- [CometBFT documentation - RPC](https://docs.cometbft.com/v0.37/rpc/)


