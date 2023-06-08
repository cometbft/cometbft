# ADR-102: RPC Companion

## Changelog

- 2022-03-27: First draft (@andynog)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

This ADR proposes an architecture of an ***RPC Companion*** solution implemented based on the proposed [ADR-101 Data Companion Pull API](https://github.com/cometbft/cometbft/blob/thane/adr-084-data-companion-pull-api/docs/architecture/adr-101-data-companion-pull-api.md). This solution can run as a sidecar concurrently with the full node, and it is optional.

This ADR provides a reference implementation of a system that can be used to offload queryable data from a CometBFT
full node to a database and offering a service that exposes the same JSONRPC methods on a endpoint as the regular JSONRPC methods of an CometBFT node endpoint,
which makes it easier for integrators of RPC clients such as client libraries and applications to switch to this
***RPC Companion*** with as minimum effort as possible.

This architecture also makes it possible to scale horizontally the querying capacity of a full node by running multiple
copies of the ***RPC Companion*** server instances that can be behind a scalable load-balancer (e.g. Cloudflare)  and can serve
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
the queryable data requests to the **RPC Companion**.

This solution shall meet the following requirements in order to provide real benefits to these users.

The **RPC Companion** solution shall:

1. Provide an **[Ingest Service](#ingest-service)** implemented as a data companion that can pull data from the node and store it on
its own storage (database)
2. Provide its own storage ([Database](#database)) that can handle a high-load of reads and also have a good performance on inserting data.
3. Implement a [database schema](#database-schema) that needs to be backwards compatible with the current CometBFT RPC.
4. Do not enforce any breaking changes to the existing RPC.
5. Ensure the responses returned by the [RPC Companion v1 endpoint](#rpc-endpoint) is wire compatible with the existing CometBFT endpoint.
6. Implement tests to verify backwards compatibility.
7. Be able to handle multiple concurrent requests and return idempotent responses.
8. Do not crash or panic if the querying demand is very high or large responses are returned.
9. Storage optimized for an access pattern of faster read throughput than write.
10. Provide metrics to allow operators to properly monitor the services and infrastructure.

It is ***NOT*** in the scope of the **RPC Companion**:

1. Provide an authentication mechanism for the RPC Companion endpoint.

### [RPC Endpoint](#rpc-endpoint)

The RPC Companion endpoint will be the same as the CometBFT endpoint but with a `/v1` appended to it. The RPC Companion endpoint
might also use a different port than the default CometBFT RPC port (e.g. `26657`).

For example, suppose these are the URLs for each RPC endpoint:

CometRPC -> `http://cosmos.host:26657`

RPC Companion -> `http://rpc-companion.host:8080/v1`

To make a request for a `block` at height `5` using the CometBFT endpoint:

`curl --header "Content-Type: application/json" --request POST --data '{"method": "block", "params": ["5"], "id": 1}' http://cosmos.host:26657`

The make the same request to the RPC Companion endpoint:

`curl --header "Content-Type: application/json" --request POST --data '{"method": "block", "params": ["5"], "id": 1}' http://rpc-companion.host:8080/v1`

> Note that only the URL changes between these two `curl` commands

The RPC Companion will accept JSONRPC requests, the same way as the CometBFT RPC endpoint does.

The methods for the RPC Companion endpoint on the following table are the ones that can be implemented first since it seems it will be straightforward and less complex

| **JSONRPC method**   | **JSONRPC Parameters**                 | **Description**                                 | **Notes**                                                                                                                                                                                                                                                                               |
|----------------------|----------------------------------------|-------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `abci_info`          |                                        | Get information about the application           | This companion endpoint will return the same response structure as the CometBFT endpoint. The companion endpoint will return the latest information stored in its database that was retrieved from the full node.                                                                       |
| `block`              | * height                               | Get block at a specified height                 | This companion endpoint will return the same response structure as the CometBFT endpoint. The data retrieved from the companion database for a particular block will have to be properly serialized into the `block` struct in order to be returned as a response.                      |
| `block_by_hash`      | * hash                                 | Get block by its hash                           | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `block_results`      | * height                               | Get block results at a specified height         | This companion endpoint will return the same response structure as the CometBFT endpoint. The data retrieved from the companion database for a particular block result will have to be properly serialized into the `ResultsBlockResults` struct in order to be returned as a response. |
| `blockchain`         | * minHeight <br/> * maxHeight          | Get blocks in a specified height range          | This companion endpoint will return the same response structure as the CometBFT endpoint. The data retrieved from the companion database will include one or more blocks.                                                                                                               |
| `commit`             | * height                               | Get commit results at a specified height        | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `consensus_params`   | * height                               | Get consensus parameters at a specified height  | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `header`             | * height                               | Get header at a specified height                | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `header_by_hash`     | * hash                                 | Get header by its hash                          | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `health`             |                                        | Get node health                                 | This endpoint basically only returns an empty response. This can be used to test if the server RPC is up.  While this on CometBFT is used to return a response if the full node is up, when using the companion service this will return an `OK` status if the companion service is up. |
| `tx`                 | * hash <br/> * prove                   | Get a transaction by its hash                   | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |
| `validators`         | * height <br/> * page <br/> * per_page | Get validator set at a specified height         | This companion endpoint will return the same response structure as the CometBFT endpoint.                                                                                                                                                                                               |

The following endpoints can also be implemented, but might require some additional effort and complexity to be implemented. These are mostly the `search` and `query` endpoints.

| **JSONRPC method**   | **JSONRPC Parameters**                                               | **Description**                        | **Notes**                                                                                                                                                                                    |
|----------------------|----------------------------------------------------------------------|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `abci_query`         | * path <br/> * data <br/> * height <br/> * prove                     | Query information from the application | This companion endpoint will return the same response structure as the CometBFT endpoint. The RPC companion service will have to implement a proper abci parameter to sql query translation  |
| `block_search`       | * query <br/> * page <br/> * per_page <br/> * order_by               | Query information about a block        | This companion endpoint will return the same response structure as the CometBFT endpoint. The RPC companion service will have to implement a proper query parameter to sql query translation |
| `tx_search`          | * query <br/> * page <br/> * per_page <br/> * prove <br/> * order_by | Query information about transactions   | This companion endpoint will return the same response structure as the CometBFT endpoint. The RPC companion service will have to implement a proper query parameter to sql query translation |

These endpoints will proxy the request to the full node. Since they are not dependent on data retrieval from the RPC Companion database they should just act as proxies to the full node.

Making these proxied methods available throught the RPC Companion endpoint will ensure that clients don't need to implement a routing logic for methods that would not be available in the RPC Companion endpoint.

> The `/broadcast_tx_*` endpoints might need some additional logic for proxying since some of them have different asynchronicity patterns.

| **JSONRPC method**     | *JSONRPC Parameters** | **Description**                           | **Notes**                                                                                                                                                                                                                                       | Proxy |
|------------------------|-----------------------|-------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------|
| `broadcast_evidence`   | * evidence            | Broadcast evidence of the misbehavior     | The evidence parameter is in JSON format                                                                                                                                                                                                        | yes   |
| `broadcast_tx_async`   | * tx                  | Broadcast a transaction                   | Returns right away with no response                                                                                                                                                                                                             | yes   |
| `broadcast_tx_sync`    | * tx                  | Broadcast a transaction                   | Returns with the response from CheckTx                                                                                                                                                                                                          | yes   |
| `broadcast_tx_commit`  | * tx                  | Broadcast a transaction                   | Returns with the responses from CheckTx and DeliverTx                                                                                                                                                                                           | yes   |
| `check_tx`             | * tx                  | Check a transaction                       | Checks a transaction without executing it                                                                                                                                                                                                       | yes   |
| `consensus_state`      |                       | Gets consensus state                      | The consensus state will not be stored in the RPC companion database so it should proxy the request to the full node                                                                                                                            | yes   |
| `dump_consensus_state` |                       | Gets the full consensus state             | The consensus state will not be stored in the RPC companion database so it should proxy the request to the full node                                                                                                                            | yes   |
| `genesis`              |                       | Gets the genesis information              | The RPC companion service can proxy the genesis request to the full node. If there are use cases that serving the genesis from the RPC companion service (no proxy) is desirable then it can be implemented as a RPC companion service endpoint | yes   |
| `net_info`             |                       | Gets network information                  | The request should proxy to the full node since the RPC companion database will not store network information                                                                                                                                   | yes   |
| `unconfirmed_txs`      | * limit               | Gets the list of unconfirmed transactions | The request should proxy to the full node since the RPC companion database will not store unconfirmed transactions information                                                                                                                  | yes   |
| `num_unconfirmed_txs`  |                       | Gets data about unconfirmed transactions  | The request should proxy to the full node since the RPC companion database will not store unconfirmed transactions information                                                                                                                  | yes   |
| `status`               |                       | Gets node status                          | The request should proxy to the full node since the RPC companion database will not store node status information                                                                                                                               | yes   |

> NOTE: The RPC Companion should not implement logic to store data in its database that can modify state in the blockchain such as
the `broadcast_tx_*` methods.  These requests will proxy to the full node as outlined above.

### High-level architecture

![High-level architecture](images/adr-102-architecture.jpeg)

This diagram shows all the required components for a full RPC Companion solution. The solution implementation contains
many parts and each one is described below:

### [Ingest Service](#ingest-service)

The **Ingest Service** pulls the data from the full node RPC endpoint and stores the information retrieved in
the RPC Companion database.

> In the future, if a gRPC interface is implemented in the full node this might be used to pull the data
from the server.

The **Ingest Service** can control the pruning on the full node via a mechanism to [pruning service](https://github.com/cometbft/cometbft/blob/thane/adr-084-data-companion-pull-api/docs/architecture/adr-101-data-companion-pull-api.md#pruning-service).
Once the Ingest Service pulls the data from the full node and is able to process it, and it gets an acknowledgement from the database that the data was inserted,
the **Ingest Service** can communicate with the full node notifying it that a specific height has been processed and set the processed
height as the `retain height` on the full node signaling this way to the node that this height can be pruned.

If the **Ingest Service** becomes unavailable (e.g. stops), then it should resume synchronization with the full node when it is back online.
The **Ingest Service** should query the full node for the last `retain height` and the **Ingest Service** should request
and process all the heights missing on the database until it catches up with the full node latest height.

In case the **Ingest Service** becomes unavailable for a long time and there are several heights to be synchronized, it is
important for the **Ingest Service** to do it in a throttled way in order not to stress the full node and cause issues in its consensus processing.

### [Database](#database)

The database stores the data retrieved from the full node and provide this data for the RPC server instance. Since the frequency
that blocks are generated on the chain are in the range from 5 seconds to 7 seconds on average, the _write_ back pressure is not
very high from a modern database perspective. While the frequency and number of requests for reading the data from the database will
be much larger due to the fact that the RPC service instance can be scaled. Therefore, a database that provides a high read
throughput should be favored.

For this is initial solution implementation it is proposed that the relational database [Postgresql](https://www.postgresql.org/) should be used in order to support
the RPC scalability and this will also provide more flexibility when implementing the RPC v2 endpoints that can return data
in different forms and database indexes might also be leveraged in order to boost the query responses performance.

The data needs to be available both for the Ingest Service (writes) and the RPC server instance (reads) and these service might be running from different machines so an embedded database
is not recommended in this case since accessing the data remotely might not be optimal for an embedded key-value database. Also
since the RPC might have many server instances (or processes) running that will need to retrieve data concurrently it is recommended to use
a well-known robust database engine that can support such a load.

Also, a database that can support ACID transactions is important to provide more guarantees that
the data was successfully inserted in the database and that an acknowledgement can be sent back to the Ingest Service to notify the
full node to prune the inserted data.

#### [Database Schema](#database-schema)

One of the challenges when implementing this solution is how to design a database schema that can be suitable to return responses that are
equivalent to the existing endpoints but at the same time offers flexibility in returning customized responses in the future.

Currently, CometBFT uses an abstraction layer from [cometbft-db](https://github.com/cometbft/cometbft-db) in order to support multiple embedded databases. These databases store
data as key-value pairs using a byte array datatype, for example to set a value for a key:

```go
func (db *GoLevelDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	if err := db.db.Put(key, value, nil); err != nil {
		return err
	}
	return nil
}
```

Since the RPC Companion stores the information in a relational database, there are opportunities to better structure and normalize the data. Here is the schema definition for a table to persist a normalized `ResultBlock` data structure in the database (Postgresql)

```sql
-- Table: comet.result_block

CREATE TABLE IF NOT EXISTS comet.result_block
(
    block_id_hash bytea NOT NULL,
    block_id_parts_hash bytea NOT NULL,
    block_id_parts_total comet.uint32 NOT NULL,
    block_header_height bigint NOT NULL,
    block_header_block_time timestamp with time zone NOT NULL,
    block_header_chain_id text NOT NULL,
    block_header_version_block comet.uint64 NOT NULL,
    block_header_version_app comet.uint64 NOT NULL,
    block_header_data_hash bytea NOT NULL,
    block_header_last_commit_hash bytea NOT NULL,
    block_header_validators_hash bytea NOT NULL,
    block_header_next_validators_hash bytea NOT NULL,
    block_header_consensus_hash bytea NOT NULL,
    block_header_app_hash bytea NOT NULL,
    block_header_last_results_hash bytea NOT NULL,
    block_header_evidence_hash bytea NOT NULL,
    block_header_proposer_address bytea NOT NULL,
    block_header_last_block_id_hash bytea NOT NULL,
    block_header_last_block_id_parts_hash bytea NOT NULL,
    block_header_last_block_id_part_total comet.uint32 NOT NULL,
    block_last_commit_height comet.uint64 NOT NULL,
    block_last_commit_round comet.uint32 NOT NULL,
    block_last_commit_block_id_hash bytea NOT NULL,
    block_last_commit_block_id_parts_total comet.uint32 NOT NULL,
    block_last_commit_block_id_parts_hash bytea NOT NULL,
    CONSTRAINT block_pkey PRIMARY KEY (block_header_height),
    CONSTRAINT last_commit_height_unique UNIQUE (block_last_commit_height),
    CONSTRAINT height_positive CHECK (block_header_height >= 0)
) TABLESPACE pg_default;
```

Additional tables will be required to persist normalized data related to the block structure. All database schema will be available in the `rpc-companion` Github repository under the `/database/schema` folder.

##### Data types

This solution will implement a data schema in the database using Postgresql built-in data types. By using a relational database, there's a possibility
to better normalize the data structures, this might provide savings in storage but might also add to the complexity of returning a particular
dataset because the data joins that will be required. Also, it would be important ensure the referential integrity is not violated since
this can cause issues to the clients consuming the data.

In order to accurately ensure the database is storing the full node data types properly, the database might implement custom data types (`domains` in Postgresql).

For example, Postgresql doesn't have an unsigned `uint64` datatype, therefore in order to support this in the database, you can use a `domain`, which is a base types with additional constrains. For example, this is the definition of a `uint64` domain:

```sql
-- DOMAIN: comet.uint64

CREATE DOMAIN comet.uint64
    AS numeric;

ALTER DOMAIN comet.uint64
    ADD CONSTRAINT value_max CHECK (VALUE <= '18446744073709551615'::numeric);

ALTER DOMAIN comet.uint64
    ADD CONSTRAINT value_positive CHECK (VALUE >= 0::numeric);
```
Additional domains for `uint32` and `uint8` will also be implemented.

##### Schema migration

Another point to consider is when the data structures change across CometBFT releases. There are a couple of ways to provide a solution to this
problem.

One way is to support a mechanism to migrate the old data to the new data structures. This would require additional logic for the migration process
that would need to be run before ingesting data from an upgraded full node that contains the new data structures. But this approach might also
cause some issues, for example, if the new data structure has a new field that there's no corresponding value in the "old" data structure, probably
the value would need to be set to `[null]` but this can have unintended consequences.

Another potential solution for this scenario is to find a way in the database that can support "versioning" of data structures. For example, let's
assume there's a `Block` structure, let's call it `v1`. If in the future there's a need to modify this structure that is not compatible with the
previous data structure, then the database would support a `v2` schema for `ResultBlock` and an `index` table could determine the criteria on which
data structure should be used for inserting or querying data.

There's also a possibility that the data structure could be stored in parallel with the normalized data but in one blob, e.g. in a `jsonb` field
that contains all the information needed to return the client, this could be an optimized way to serve the data but could add to the storage
requirements. This could improve the performance for reading data from the database. An alternative approach to this problem that can offer
similar results would be to use a "caching" layer.


### RPC server instance

The **RPC server instance** is a node that runs the RPC API process for the data companion. This server instance provides an RPC API (v1) with
the same JSONRPC methods in the full node endpoint. The RPC Companion service will expose the same JSONRPC methods and will accept the same request types and
return wire compatible responses (should match the same response as the equivalent full node RPC endpoint).

The **RPC server instance**, when serving a particular request, retrieves the required data from the database in order to
fulfill the request. The data should be serialized in a way that makes it wire compatible with the CometBFT RPC endpoint.

Identical requests should return idempotent responses, no side effects should cause the RPC service to return different responses.

The RPC service endpoints should be exposed through an external load-balancer service such as Cloudflare or AWS ELB, or
a server running its own load balancer mechanism (e.g. nginx).

The RPC clients should make requests to the **RPC Companion** server instances through this load balancer.

The **RPC server instance** will also implement logic for the proxied requests to the full node. It should properly handle the proxy requests and responses from the full node.

The **RPC Companion** should support the `https` protocol in order to support secure endpoints access. It's recommended that
the `https` support is provided by the load balancer but in case there's a single **RPC Companion** server instance, having an
option to use `https` is important and it's backwards compatible with the existing CometBFT RPC that supports that.


## Consequences

### Positive

- Alternative and optional **RPC Companion** that is more scalable and reliable with a higher query throughput.
- Less backpressure on the full node that is running consensus.
- Possibility for future additional endpoints (v2).
- Allow users to create better and faster indexers and analytics solutions.

### Negative

- Additional infrastructure complexity to set up and maintain.
- Additional infrastructure cost.

### Neutral

- Optional feature, users will only use it if needed.
- No privacy / security issues should arise since the data returned by the **RPC Companion** will be the same
as the current RPC.

## References

- [Improve experience for integrators](https://github.com/cometbft/cometbft/issues/40)
- [ADR-101: Data Companions Pull API (tracking issue)](https://github.com/cometbft/cometbft/issues/574)
- [ADR-101: Data Companions Pull API (PR)](https://github.com/cometbft/cometbft/pull/82)
- [CometBFT documentation - RPC](https://docs.cometbft.com/v0.37/rpc/)


