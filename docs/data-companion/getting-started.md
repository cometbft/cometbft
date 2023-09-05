---
order: 1
parent:
    title: Creating a Data Companion
    order: 2
---


# Creating a Data Companion for CometBFT

## Fetching data

If you're planning to initiate a Data Companion workflow, the first crucial step is to extract the data that you need to
offload from the node. CometBFT offers a range of endpoints where you can access the data, including Blocks and Block Results.

This documentation aims to provide a detailed explanation of the latest gRPC services that CometBFT offers, which can
be used to retrieve the required data.

### gRPC services

The endpoints to control the pruning mechanism are exposed via the gRPC services. They are configured and need to be
enabled separately from the regular gRPC services. We call them "privileged" services since by invoking them the storage
on the node can be manipulated and only operators with privileged access to the server should be able to invoke them.

The gRPC services offer endpoints that provide control over the pruning mechanism. These endpoints are distinct from the
regular gRPC services and require separate configuration and activation. These endpoints are known as "privileged" services
because they have the ability to manipulate the storage on the node. Therefore, only operators who have privileged access
to the server should be allowed to use them.

The privileged services are designed to enable operators to control the pruning mechanism, which is a feature of the node
that allows for the removal of unnecessary data from the storage. By removing unnecessary data, the storage capacity
can be optimized, and the performance of the node can be enhanced. The pruning mechanism is critical for the efficient
functioning of the node, and it is essential that only authorized personnel have access to the privileged services that
control it. Unauthorized use of the privileged services can lead to data loss or corruption, which can have serious
consequences for the node and the network it is connected to. Therefore, the security and integrity of the privileged
services must be maintained at all times.

#### Enabling the gRPC services

In order to be able to use the gRPC service, they should be enabled through CometBFT's configuration

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

The regular gRPC services are enabled by default. Each service has its own property to disable or enabled it. For example,
to enable the `Version` service, the `[grpc.version_service]` section, ensure that the `enabled` property is set to `true`:

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

This is the same thing for the `block_service` and the `block_results_service`

```
# The gRPC block service returns block information
[grpc.block_service]
enabled = true

# The gRPC block results service returns block results for a given height. If no height
# is given, it will return the block results from the latest height.
[grpc.block_results_service]
enabled = true
```

##### Privileged Services

CometBFT provides "privileged" services which are not intended to be exposed to the public-facing Internet.
These services are designed to perform critical operations on the blockchain network, and therefore, it's crucial to
ensure that they are used only when necessary.

The privileged services offered by CometBFT can modify the data stored in the node, and hence, it's essential to keep
them off by default to avoid any unintended modifications. However, when required, these services can be activated
to set and retrieve a retained height, which can influence the pruning mechanism on the node.

It's worth noting that any modifications made by the privileged services can impact the overall node data. Therefore,
proper caution and care should be exercised while using these services.

To enable the privileged endpoint and services set the appropriate values in the configuration file.

Add the address for the regular (non-privileged) services, for example:
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

## Storing the fetched data

In the Data Companion workflow, the second step involves saving the data retrieved from a blockchain onto an external
storage medium, such as a database. This external storage medium is important because it allows the data to be accessed
and utilized by custom web services that can serve the blockchain data in a more efficient way. Once the data has been
successfully stored externally, the Data Companion can then inform the relevant node that the data is no longer required.

One important concept that can affect the pruning of nodes is the "retain_height". The retain_height determines the specific
point in time from which the data can be safely deleted from the node's storage. By considering the retain_height,
nodes can effectively manage their storage usage and ensure that they are only retaining the data that is necessary for their operations.
This is important because storage space is a finite resource and nodes with limited storage space may struggle to keep up with the growth of the blockchain.

## Pruning the node data

In order to successfully execute the Data Companion workflow, the third step entails utilizing the newly introduced
gRPC APIs to regulate the retained height value of the data companion. The pruning control mechanism, which has been
recently introduced, allows the data companion to effectively manage its operations. For a comprehensive understanding
of this mechanism, including its features and functionalities, please see the [Pruning](./pruning.md) section.
