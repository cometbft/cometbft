---
order: 1
parent:
    title: Pruning
    order: 3
---

# Pruning data from CometBFT using a Data Companion

CometBFT employs a sophisticated pruning logic to eliminate unnecessary data and reduce storage requirements. This
comprehensive document offers a detailed explanation of how CometBFT's pruning logic works.

It covers various use cases where the data companion can influence the pruning process and explains how the application
can control the pruning logic on the node. Our aim is to provide you with a thorough understanding of the intricate
workings of CometBFT's pruning system.

### Retain Height parameter

The pruning API uses two crucial parameters that allow for efficient data storage management. The "block retain height"
pruning service parameter specifies the height to which the node will preserve blocks. It is important to note that this
parameter differs from the application block retain height, which the application sets in response to ABCI commit messages.

The node will preserve blocks up to the lowest value between the pruning service and the application block retain height.
This feature helps manage storage efficiently and ensures that the node's storage is optimized.

The "block results retain height" pruning service parameter is the second crucial aspect of the pruning API.
This parameter determines the height up to which the node will keep block results. By retaining block results to a
certain height, the node can efficiently manage its storage and optimize its performance.

With these two parameters, the pruning API ensures that the node efficiently manages its data storage.

### Block Retain Height

In order to set the block retain height on the node, you have to enable the privileged services endpoint and the block service in the
configuration as described in the `Privileged Services` section of the Getting started guide.

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
configuration as described in the `Privileged Services` section of the Getting started guide.

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
