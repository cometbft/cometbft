# ABCI Clients

## Overview

CometBFT communicates with the application through the Application Blockchain Interface (ABCI). The ABCI is an interface that defines the boundary between the replication engine (the blockchain), CometBFT, and the state machine (the application). By using different ABCI client implementations, CometBFT can connect to applications in different ways, each with specific advantages and trade-offs.

CometBFT provides several client implementations for connecting to ABCI applications:

1. **Local Clients** - for when the application runs in the same process
   - `LocalClient` - uses a mutex to ensure thread safety
   - `UnsyncLocalClient` - less synchronization overhead than `LocalClient`
   - `NoLockLocalClient` (in 1.0) - removes all mutual exclusion mechanisms

2. **Remote Clients** - for when the application runs in a separate process
   - `SocketClient` - communicates over a socket connection
   - `GRPCClient` - communicates using gRPC

This document provides details about these clients, focusing particularly on the local client variants and their concurrency models.

## Local Clients

When CometBFT and the application run in the same process, local clients are used to avoid the overhead of inter-process communication.

### LocalClient

The `LocalClient` is the traditional client that provides full mutex protection for all ABCI calls.

```go
// NewLocalClient creates a local client, which wraps the application interface
// that Comet as the client will call to the application as the server.
//
// Concurrency control in each client instance is enforced by way of a single
// mutex. If a mutex is not supplied (i.e. if mtx is nil), then one will be
// created.
func NewLocalClient(mtx *cmtsync.Mutex, app types.Application) Client
```

**Key characteristics:**
- Uses a mutex lock for every ABCI method call
- Provides thread safety for applications that are not thread-safe
- Higher overhead due to mutex acquisition/release for every call
- Appropriate for applications that are not designed for concurrent access

### UnsyncLocalClient

The `UnsyncLocalClient` reduces synchronization overhead by only maintaining a mutex over the callback mechanism, but not over application method calls.

```go
// NewUnsyncLocalClient creates a local client, which wraps the application
// interface that Comet as the client will call to the application as the
// server.
//
// This differs from NewLocalClient in that it returns a client that only
// maintains a mutex over the callback used by CheckTxAsync and not over the
// application, leaving it up to the proxy to handle all concurrency. If the
// proxy does not impose any concurrency restrictions, it is then left up to
// the application to implement its own concurrency for the relevant group of
// calls.
func NewUnsyncLocalClient(app types.Application) Client
```

**Key characteristics:**
- Only uses mutex protection for callback handling
- Application method calls are unprotected
- Lower overhead compared to `LocalClient`
- Requires the application or proxy to handle its own concurrency
- Best for applications that implement their own thread safety mechanisms

## When to Use UnsyncLocalClient

You should use `UnsyncLocalClient` when:

1. Your ABCI application handles concurrency internally
2. You want to reduce the overhead of mutex locking/unlocking
3. You're proxying to an application that already handles concurrency
4. You're implementing a proxy that manages concurrency itself

## Usage Examples

Here's an example of how to create an `UnsyncLocalClient`:

```go
import (
    abcicli "github.com/cometbft/cometbft/abci/client"
    "github.com/cometbft/cometbft/abci/types"
)

// Create your ABCI application
app := MyThreadSafeApplication{}

// Create an unsynchronized local client
client := abcicli.NewUnsyncLocalClient(app)

// Start the client
if err := client.Start(); err != nil {
    // Handle error
}

// Use the client to make ABCI calls
// Note: No mutex protection on these calls!
resp, err := client.Info(ctx, &types.InfoRequest{})
```

## Client Creation in CometBFT

CometBFT creates multiple ABCI clients for different purposes:

1. **Consensus Connection**: Used for consensus-related calls (`InitChain`, `FinalizeBlock`, etc.)
2. **Mempool Connection**: Used for transaction validation (`CheckTx`)
3. **Query Connection**: Used for queries to the application state (`Query`, `Info`)
4. **Snapshot Connection**: Used for state sync functionality

The type of client used for each connection is determined by the application proxy configuration.

## Performance Considerations

When choosing between different client implementations, consider:

- **Thread safety**: Does your application handle concurrent access?
- **Performance**: Local clients are generally faster than remote clients
- **Process isolation**: Remote clients provide better fault isolation
- **Synchronization overhead**: `UnsyncLocalClient` has less overhead than `LocalClient`

## See Also

- [ABCI Specification](../../../abci/README.md)
- [Application Development Guide](../../guides/app-dev/app-architecture.md)
