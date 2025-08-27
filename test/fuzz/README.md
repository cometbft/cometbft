# Fuzz Testing

This directory contains fuzz tests for CometBFT packages using Go's native fuzzing infrastructure (Go 1.20+).

## Overview

Fuzz testing helps identify security vulnerabilities and bugs by providing random, unexpected, or invalid inputs to functions. The tests in this directory target critical components of CometBFT:

- **Mempool**: Tests `CheckTx` functionality using the kvstore in-process ABCI app
- **P2P**: Tests `SecretConnection#Read` and `SecretConnection#Write` for network security
- **RPC**: Tests JSON-RPC server for API robustness

## Running Fuzz Tests

Use the Go toolchain to run fuzz tests:

```bash
# Test mempool CheckTx functionality
go test -fuzz Mempool ./tests

# Test P2P SecretConnection read/write operations
go test -fuzz P2PSecretConnection ./tests

# Test RPC JSON-RPC server
go test -fuzz RPCJSONRPCServer ./tests
```

## Test Files

- `tests/mempool_test.go` - Mempool transaction validation fuzzing
- `tests/p2p_secretconnection_test.go` - P2P encrypted connection fuzzing
- `tests/rpc_jsonrpc_server_test.go` - RPC server input validation fuzzing

## Continuous Fuzzing

This project also supports continuous fuzzing via OSS-Fuzz. See `oss-fuzz-build.sh` for build configuration.

## Resources

- [Go Fuzzing Documentation](https://go.dev/doc/fuzz/)
- [CometBFT Security Policy](../SECURITY.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
