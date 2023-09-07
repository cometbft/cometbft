---
order: 1
parent:
    title: Pruning
    order: 3
---

# Using a Data Companion to influence data pruning on a CometBFT node

CometBFT employs a sophisticated pruning logic to eliminate unnecessary data and reduce storage requirements.

This document covers use cases where the data companion can influence the pruning process and explains how the data companion
can influence the pruning logic on the node.

CometBFT provides a privileged gRPC endpoint for the pruning service. This privileged endpoint is distinct from the
regular gRPC endpoint and require separate configuration and activation. These "privileged" services
have the ability to manipulate the storage on the node. Therefore, only operators who have privileged access
to the server should be allowed to use them.

## Pruning configuration

You need to ensure that the pruning for data companion is enabled in the configuration to allow the data companion
influence the node pruning mechanism.

In order to do that, please ensure that in the `[storage.pruning.data_companion]` section of the CometBFT's configuration
file, the property `enabled` is set to `true`:

```
[storage.pruning.data_companion]

# Whether automatic pruning respects values set by the data companion. Disabled
# by default. All other parameters in this section are ignored when this is
# disabled.
#
# If disabled, only the application retain height will influence block pruning
# (but not block results pruning). Only enabling this at a later stage will
# potentially mean that blocks below the application-set retain height at the
# time will not be available to the data companion.
enabled = true
```

In order to avoid unwanted pruning of data when the data companion is activated, it is possible to define the initial
retain height for block and block results in the configuration file. This configuration ensures that the necessary data
is retained and not removed in an undesirable way.

For example, you can change the values in this part of the configuration:
```
# The initial value for the data companion block retain height if the data
# companion has not yet explicitly set one. If the data companion has already
# set a block retain height, this is ignored.
initial_block_retain_height = 0

# The initial value for the data companion block results retain height if the
# data companion has not yet explicitly set one. If the data companion has
# already set a block results retain height, this is ignored.
initial_block_results_retain_height = 10
```

## Retain Height parameter

The pruning API uses two crucial parameters that allow for efficient data storage management. The "block retain height"
pruning parameter specifies the height to which the node will preserve blocks. It is important to note that this
parameter differs from the application block retain height, which the application sets in response to ABCI commit messages.

The node will preserve blocks up to the lowest value between the data companion retain height and the application block retain height.

The "block results retain height" pruning parameter is the second crucial aspect of the pruning API.
This parameter determines the height up to which the node will keep block results. By retaining block results to a
certain height, the node can efficiently manage its storage and optimize its performance.

With these two parameters, the pruning service ensures that the node efficiently manages its data storage.

### Block Retain Height

In order to set the block retain height on the node, you have to enable the privileged services endpoint and the block service in the
configuration as described in the `Privileged Services` section of the [Creating a Data Companion for CometBFT](./quick-start.md).

Once the services are enabled, you can use the Golang client provided by CometBFT to invoke the method that sets the block retain height.

Here is an example code:
```

import (
    "github.com/cometbft/cometbft/rpc/grpc/client/privileged"
)

ctx := context.Background()

// Privileged Service Client
addr := "0.0.0.0:26091"
conn, err := privileged.New(ctx, addr, privileged.WithInsecure())
if err != nil {
    // Do something with the error
}

err := conn.SetBlockRetainHeight(ctx, height)
if err != nil {
    // Do something with the error
}
```

> NOTE: If you try to set the `Block Retain Height` to a value that is lower to what is currently stored in the node, an error will
be returned informing that.

If you need to check what is the current value for the `Block Retain Height` you can use another service.

Here's an example:
```
retainHeight, err := conn.GetBlockRetainHeight(ctx)
if err != nil {
    // Do something with the error
} else {
    // Do something with the `retainHeight` value
}
```

### Block Results Retain Height

In order to set the block results retain height on the node, you have to enable the privileged services endpoint and the block results service in the
configuration as described in the `Privileged Services` section of the [Creating a Data Companion for CometBFT](./quick-start.md).

Once the services are enabled, you can use the Golang client provided by CometBFT to invoke the method that sets the block results retain height.

Here is an example code:
```

import (
    "github.com/cometbft/cometbft/rpc/grpc/client/privileged"
)

ctx := context.Background()

// Privileged Service Client
addr := "0.0.0.0:26091"
conn, err := privileged.New(ctx, addr, privileged.WithInsecure())
if err != nil {
    // Do something with the error
}

err := conn.SetBlockResultsRetainHeight(ctx, height)
if err != nil {
    // Do something with the error
}

```

> NOTE: If you try to set the `Block Results Retain Height` to a value that is lower to what is currently stored in the node, an error will
be returned informing that.

If you need to check what is the current value for the `Block Results Retain Height` you can use another service.

Here's an example:
```
retainHeight, err := conn.GetBlockResultsRetainHeight(f.context)
if err != nil {
    // Do something with the error
} else {
    // Do something with the `retainHeight` value
}

```
