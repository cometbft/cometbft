# Remote Signer Specification

## Overview

A remote signer is a component that securely stores and uses a validator's private keys outside of the main node process. This increases the security and flexibility of validator infrastructure.

## Architecture

- **Node (CometBFT)** — the client that sends signing requests.
- **Remote Signer** — the server that holds the private key and performs signing upon request.

## Communication Protocol

- Uses a TCP connection with message serialization via Protocol Buffers (protobuf).
- In the future — gRPC (see [ADR-063](../../docs/references/architecture/tendermint-core/adr-063-privval-grpc.md)).

## Message Formats

### Main request types:

- **GetPubKey** — retrieve the validator's public key.
- **SignVote** — sign a vote.
- **SignProposal** — sign a proposal.
- **SignBytes** — sign arbitrary bytes.
- **Ping** — check service availability.

### Example protobuf messages (simplified):

```proto
message PubKeyRequest { string chain_id = 1; }
message PubKeyResponse { bytes pub_key = 1; string error = 2; }

message SignVoteRequest { Vote vote = 1; string chain_id = 2; bool skip_extension_signing = 3; }
message SignedVoteResponse { Vote vote = 1; string error = 2; }

message SignProposalRequest { Proposal proposal = 1; string chain_id = 2; }
message SignedProposalResponse { Proposal proposal = 1; string error = 2; }

message SignBytesRequest { bytes value = 1; }
message SignBytesResponse { bytes signature = 1; string error = 2; }

message PingRequest {}
message PingResponse {}
```

## Exchange Examples

- Node sends a `SignVoteRequest`, remote signer returns a `SignedVoteResponse` with the signature.
- Node sends a `GetPubKeyRequest`, remote signer returns a `PubKeyResponse` with the public key.

## Security

- It is recommended to use TLS to secure the connection.
- Mutual authentication (client/server certificates) is possible.
- The private key must be protected from unauthorized access.

## Errors

- All responses may contain an `error` field with a description.
- In the gRPC version, it is recommended to use standard gRPC error codes.

## Integration Example

- The node (CometBFT) is configured to work with a remote signer via configuration.
- The remote signer runs as a separate process/service and accepts connections from the node.

## Compatibility and Migration

- Current protocol: TCP + protobuf.
- Future: gRPC (see [ADR-063](../../docs/references/architecture/tendermint-core/adr-063-privval-grpc.md)).
- Migration will require updating both the node and the remote signer.
