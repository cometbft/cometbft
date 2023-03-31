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

This diagram shows all the required components for a full Standalone RPC solution. The solution implementation contains
many parts and each one is described below:

#### Full Node

A **full node** for a blockchain that runs a CometBFT process. The full node needs to expose the CometBFT RPC and accessible to the
ingestion service, so it can pull the data from the **full node**.

#### Ingest Service

The **ingest service** pulls the data from the full node via its RPC endpoints and store the information retrieved in
the database. In the future, if a gRPC interface is implemented in the full node this might be used to pull the data
from the server.

The **ingest service** can control the pruning on the full node via a mechanism to track a `retain height`. Once the ingest service
pulls the data from the full node and is able to process it and it gets an acknowledgement from the database that the data was inserted,
the **ingest service** can communicate with the full node notifying it that a specific height has been processed and set the processed
height as the `retain height` on the full node signaling this way to the node that this height can be pruned.

If the **ingest service** becomes unavailable (e.g. stops), then it should resume synchronization with the full node when it is back online.
The **ingest service** should query the full node for the last `retain height` and the **ingest service** should request
and process all the heights missing on the database until it catches up with the full node latest height.

In case the **ingest service** becomes unavailable for a long time and there are a lot of height be caught up, it is
important for the **ingest service** to do it in a throttled way in order not to stress the full node and cause issues in its consensus processing.

#### Database

The database stores the data retrieved from the full node and provide this data for the RPC server instance. Since the frequency
that blocks are generated on the chain are in the range from 5 seconds to 7 seconds on average, the _write_ back pressure is not
very high from a modern database perspective. While the frequency and number of requests for reading the data from the database will
be much larger due to the fact that the RPC service instance can be scaled. Therefore, a database that provides a high read
throughput should be favored.

For this is initial solution implementation it is proposed that a modern relational database should be used in order to support
the RPC scalability and this will also provide more flexibility when implementing the RPC v2 endpoints that can return data
in different schemas.

The data needs to be available both for the ingest service (writes) and the RPC server instance (reads) so an embedded key-value
store is not recommended in this case since accessing the data remotely might not be optimal for an embedded key-value database and
since the RPC might have many server instances running that will need to retrieve data concurrently it is recommended to use
a well-known robust database engine that can support such load such as Postgres.

Also, a database that can support ACID transactions is important to provide more guarantees that
the data was successfully inserted in the database and this acknowledgement can be used by the ingest service to notify the
full node to prune this inserted data.

#### RPC server instance

The **RPC server instance** is a node that runs the RPC API process for the data companion. This server instance provide an RPC API (v1) with
the same endpoints as the full node. The Standalone RPC service will expose the same endpoints and will accept the same request types and
return wire compatible responses (should match the same response as the equivalent full node RPC endpoint).

The **RPC server instance**, when serving a particular request, retrieves the required data from the database in order to
fulfill the request. The data should be serialized in a way that makes it wire compatible with the CometBFT RPC endpoint.

Identical requests should return idempotent responses, no side effects should cause the RPC service to return different responses.

These are the endpoints to be implemented for the Standalone RPC (v1)

  - /block
  - /block_results
  - ... (TBD)

These are some of the future new endpoints that could be implemented for the Standalone RPC (v2)
  - (TBD)

> NOTE: The Standalone RPC server instances should not implement endpoints that can modify state in the blockchain such as
  the `/broadcast_tx_*` endpoints. Since there might be many load balanced RPC server instances, this might cause issues with
transactions, for example sequential transactions might be relayed in the wrong order causing the full node to reject some of
the transactions with sequence mismatch errors. It is expected that RPC clients have logic to forward these requests directly
to the full node.

#### Load balancer

The RPC service endpoints should be exposeds through an external load-balancer service such as Cloudflare or AWS ELB, or
a server running its own load balancer mechanism (e.g. nginx).

The RPC clients should make requests to the Standalone RPC server instances through this load balancer.

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


