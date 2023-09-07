---
order: 1
parent:
    title: Quick Start
    order: 2
---


# Creating a Data Companion for CometBFT

## Fetching data

One of the most important steps to initiate a Data Companion workflow is to extract the necessary data from the node.
Fortunately, CometBFT provides new endpoints that allow you to access the required data, such as `version`, `block` and
`block results` services.

This documentation aims to provide a detailed explanation of CometBFT's latest gRPC services that can be used to retrieve
the data you need.

## Enabling the gRPC services

In order to be able to use the gRPC services, they should be enabled through CometBFT's configuration

In the `[gRPC]` section of the configuration
```
#######################################################
###       gRPC Server Configuration Options         ###
#######################################################

#
# Note that the gRPC server is exposed unauthenticated. It is critical that
# this server not be exposed directly to the public internet. If this service
# must be accessed via the public internet, please ensure that appropriate
# precautions are taken (e.g. fronting with a reverse proxy like nginx with TLS
# termination and authentication, using DDoS protection services like
# CloudFlare, etc.).
#

[grpc]
```

Add the address for the regular (non-privileged) services, for example:

```
# TCP or UNIX socket address for the RPC server to listen on. If not specified,
# the gRPC server will be disabled.
laddr = "tcp://0.0.0.0:26090"
```

The regular gRPC services are **enabled by default**. Each service has its own property to disable or enabled it. For example,
to enable the `Version` service, in the `[grpc.version_service]` section, ensure that the `enabled` property is set to `true`:

```
#
# Each gRPC service can be turned on/off, and in some cases configured,
# individually. If the gRPC server is not enabled, all individual services'
# configurations are ignored.
#

# The gRPC version service provides version information about the node and the
# protocols it uses.
[grpc.version_service]
enabled = true
```

Do the same thing for the `block_service` and the `block_results_service` to have them enabled.

```
# The gRPC block service returns block information
[grpc.block_service]
enabled = true

# The gRPC block results service returns block results for a given height. If no height
# is given, it will return the block results from the latest height.
[grpc.block_results_service]
enabled = true
```

## Privileged Services

CometBFT provides "privileged" services which are not intended to be exposed to the public-facing Internet.

The privileged services offered by CometBFT can modify the data stored in the node, and hence, it's essential to keep
them off by default to avoid any unintended modifications.

However, when required, these services can be activated
to set and retrieve a retained height, which can influence the pruning mechanism on the node.

Therefore, proper caution and care should be exercised while using these services.

To enable the privileged endpoint and services, you have to set the appropriate values in the configuration file.

The first step is to set the address for the privileged services, for example:
```
#
# Configuration for privileged gRPC endpoints, which should **never** be exposed
# to the public internet.
#
[grpc.privileged]
# The host/port on which to expose privileged gRPC endpoints.
laddr = "tcp://0.0.0.0:26091"
```

In the `[grpc.privileged.pruning_service]` section, ensure the value `enabled` is set to `true`

```
#
# Configuration specifically for the gRPC pruning service, which is considered a
# privileged service.
#
[grpc.privileged.pruning_service]

# Only controls whether the pruning service is accessible via the gRPC API - not
# whether a previously set pruning service retain height is honored by the
# node. See the [storage.pruning] section for control over pruning.
#
# Disabled by default.
enabled = true
```

## Fetching data

CometBFT has recently introduced new gRPC services to facilitate data retrieval via the gRPC endpoint services.
These new services, namely, `version, `block` and `block results`, offer a more efficient and faster means of fetching data
as compared to the RPC endpoint. While the RPC endpoint is still a viable option, gRPC will be the preferred choice going forward.

### Fetching **Block** data

In order to retrieve `block` data using the gRPC block service, you have to ensure the gRPC endpoint and the service is
enabled as described in the section above.

Once the service has been enabled, the Golang gRPC client provided by CometBFT can be utilized to retrieve data from the node,
such as retrieved a block by its height.

This client code is a convenient option for retrieving data, as it allows for requests to
be sent and responses to be managed in a more idiomatic manner. However, if necessary, the proto version can also be used directly.

Here is an example code to retrieve a block by height:
```
import (
     "github.com/cometbft/cometbft/rpc/grpc/client"
)

ctx := context.Background()

// Service Client
addr := "0.0.0.0:26090"
conn, err := client.New(ctx, addr, client.WithInsecure())
if err != nil {
    // Do something with the error
}

block, err := conn.GetBlockByHeight(ctx, height)
if err != nil {
    // Do something with the error
} else {
    // Do something with the `block`
}

```

### Fetching **Block Results** data

To fetch `block results` you can use a similar code as the previous one but just invoking the method to that retrieves
block results

Here's an example:
```
blockResults, err := conn.GetBlockResults(ctx, height)
if err != nil {
    // Do something with the error
} else {
    // Do something with the `blockResults`
}

```

### Stream for the latest height

[TODO]

## Storing the fetched data

In the Data Companion workflow, the second step involves saving the data retrieved from a blockchain onto an external
storage medium, such as a database. This external storage medium is important because it allows the data to be accessed
and utilized by custom web services that can serve the blockchain data in a more efficient way.

When choosing a database, evaluate your specific needs, including data size, user access, and budget.
For example, the [RPC Companion](https://github.com/cometbft/rpc-companion) uses Postgresql as a starting point, but there
are many other options to consider. Choose a database that meets your needs and helps you achieve your objectives.

Before proceeding to the next step, it is crucial to verify that the data has been correctly stored in the external database.
Once you have confirmed that the data has been successfully stored externally, and if you have evidence that the previous
data has also been stored externally, you can proceed to update the new "retain_height" information. This update will
allow the node to prune the information that is now stored externally.

## Pruning the node data

In order to successfully execute the Data Companion workflow, the third step entails utilizing the newly introduced
gRPC APIs to set the retain height value on the node. The pruning service allows the data companion to effectively influence
pruning on the node.

One important concept that can affect the pruning of nodes is the `retain_height`. The retain_height determines the specific
height from which the data can be safely deleted from the node's storage. By considering the retain_height,
nodes can effectively manage their storage usage and ensure that they are only retaining the data that is necessary for
their operations. This is important because storage space is a finite resource and nodes with limited storage space may
struggle to keep up with the growth of the blockchain.

For a comprehensive understanding of the pruning service for the data companion, please see the document
[Using a Data Companion to influence data pruning on a CometBFT node](./pruning.md).
