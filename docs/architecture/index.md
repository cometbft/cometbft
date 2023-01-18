---
---

# Architecture Decision Records (ADR)

This is a location to record all high-level architecture decisions in the tendermint project.

You can read more about the ADR concept in this [blog post](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t).

An ADR should provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References
- Changelog

Note the distinction between an ADR and a spec. The ADR provides the context, intuition, reasoning, and
justification for a change in architecture, or for the architecture of something
new. The spec is much more compressed and streamlined summary of everything as
it stands today.

If recorded decisions turned out to be lacking, convene a discussion, record the new decisions here, and then modify the code to match.

Note the context/background should be written in the present tense.

## Table of Contents

### Implemented

- [ADR-001: Logging](../adr-001-logging)
- [ADR-002: Event-Subscription](../adr-002-event-subscription)
- [ADR-003: ABCI-APP-RPC](../adr-003-abci-app-rpc)
- [ADR-004: Historical-Validators](../adr-004-historical-validators)
- [ADR-005: Consensus-Params](../adr-005-consensus-params)
- [ADR-008: Priv-Validator](../adr-008-priv-validator)
- [ADR-009: ABCI-Design](../adr-009-ABCI-design)
- [ADR-010: Crypto-Changes](../adr-010-crypto-changes)
- [ADR-011: Monitoring](../adr-011-monitoring)
- [ADR-014: Secp-Malleability](../adr-014-secp-malleability)
- [ADR-015: Crypto-Encoding](../adr-015-crypto-encoding)
- [ADR-016: Protocol-Versions](../adr-016-protocol-versions)
- [ADR-017: Chain-Versions](../adr-017-chain-versions)
- [ADR-018: ABCI-Validators](../adr-018-ABCI-Validators)
- [ADR-019: Multisigs](../adr-019-multisigs)
- [ADR-020: Block-Size](../adr-020-block-size)
- [ADR-021: ABCI-Events](../adr-021-abci-events)
- [ADR-025: Commit](../adr-025-commit)
- [ADR-026: General-Merkle-Proof](../adr-026-general-merkle-proof)
- [ADR-033: Pubsub](../adr-033-pubsub)
- [ADR-034: Priv-Validator-File-Structure](../adr-034-priv-validator-file-structure)
- [ADR-043: Blockchain-RiRi-Org](../adr-043-blockchain-riri-org)
- [ADR-044: Lite-Client-With-Weak-Subjectivity](../adr-044-lite-client-with-weak-subjectivity)
- [ADR-046: Light-Client-Implementation](../adr-046-light-client-implementation)
- [ADR-047: Handling-Evidence-From-Light-Client](../adr-047-handling-evidence-from-light-client)
- [ADR-051: Double-Signing-Risk-Reduction](../adr-051-double-signing-risk-reduction)
- [ADR-052: Tendermint-Mode](../adr-052-tendermint-mode)
- [ADR-053: State-Sync-Prototype](../adr-053-state-sync-prototype)
- [ADR-054: Crypto-Encoding-2](../adr-054-crypto-encoding-2)
- [ADR-055: Protobuf-Design](../adr-055-protobuf-design)
- [ADR-056: Light-Client-Amnesia-Attacks](../adr-056-light-client-amnesia-attacks)
- [ADR-059: Evidence-Composition-and-Lifecycle](../adr-059-evidence-composition-and-lifecycle)
- [ADR-065: Custom Event Indexing](../adr-065-custom-event-indexing)
- [ADR-066: E2E-Testing](../adr-066-e2e-testing)
- [ADR-072: Restore Requests for Comments](../adr-072-request-for-comments)
- [ADR-076: Combine Spec and Tendermint Repositories](../adr-076-combine-spec-repo)
- [ADR-077: Configurable Block Retention](../adr-077-block-retention)
- [ADR-078: Non-zero Genesis](../adr-078-nonzero-genesis)

### Accepted

- [ADR-006: Trust-Metric](../adr-006-trust-metric)
- [ADR-024: Sign-Bytes](../adr-024-sign-bytes)
- [ADR-035: Documentation](../adr-035-documentation)
- [ADR-039: Peer-Behaviour](../adr-039-peer-behaviour)
- [ADR-063: Privval-gRPC](../adr-063-privval-grpc)
- [ADR-067: Mempool Refactor](../adr-067-mempool-refactor)
- [ADR-071: Proposer-Based Timestamps](../adr-071-proposer-based-timestamps)
- [ADR-075: RPC Event Subscription Interface](../adr-075-rpc-subscription)
- [ADR-079: Ed25519 Verification](../adr-079-ed25519-verification)
- [ADR-081: Protocol Buffers Management](../adr-081-protobuf-mgmt)

### Deprecated

None

### Rejected

- [ADR-023: ABCI-Propose-tx](../adr-023-ABCI-propose-tx)
- [ADR-029: Check-Tx-Consensus](../adr-029-check-tx-consensus)
- [ADR-058: Event-Hashing](../adr-058-event-hashing)

### Proposed

- [ADR-007: Trust-Metric-Usage](../adr-007-trust-metric-usage)
- [ADR-012: Peer-Transport](../adr-012-peer-transport)
- [ADR-013: Symmetric-Crypto](../adr-013-symmetric-crypto)
- [ADR-022: ABCI-Errors](../adr-022-abci-errors)
- [ADR-030: Consensus-Refactor](../adr-030-consensus-refactor)
- [ADR-036: Empty Blocks via ABCI](../adr-036-empty-blocks-abci)
- [ADR-037: Deliver-Block](../adr-037-deliver-block)
- [ADR-038: Non-Zero-Start-Height](../adr-038-non-zero-start-height)
- [ADR-040: Blockchain Reactor Refactor](../adr-040-blockchain-reactor-refactor)
- [ADR-041: Proposer-Selection-via-ABCI](../adr-041-proposer-selection-via-abci)
- [ADR-042: State Sync Design](../adr-042-state-sync)
- [ADR-045: ABCI-Evidence](../adr-045-abci-evidence)
- [ADR-050: Improved Trusted Peering](../adr-050-improved-trusted-peering)
- [ADR-057: RPC](../adr-057-RPC)
- [ADR-060: Go-API-Stability](../adr-060-go-api-stability)
- [ADR-061: P2P-Refactor-Scope](../adr-061-p2p-refactor-scope)
- [ADR-062: P2P-Architecture](../adr-062-p2p-architecture)
- [ADR-064: Batch Verification](../adr-064-batch-verification)
- [ADR-068: Reverse-Sync](../adr-068-reverse-sync)
- [ADR-069: Node Initialization](../adr-069-flexible-node-initialization)
- [ADR-073: Adopt LibP2P](../adr-073-libp2p)
- [ADR-074: Migrate Timeout Parameters to Consensus Parameters](../adr-074-timeout-params)
- [ADR-080: Reverse Sync](../adr-080-reverse-sync)
