---
order: 1
parent:
  order: false
---

# Architecture Decision Records (ADR)

This is a location to record all high-level architecture decisions in the
CometBFT project.

You can read more about the ADR concept in this
[blog post](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t).

An ADR should, with a strong focus on the impact on _users_ of the system,
provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References
- Changelog

To create a new ADR, please use the [ADR template](./adr-template.md).

Note the distinction between an ADR and a spec. An ADR provides the context,
intuition, reasoning, and justification for a change in architecture, or for the
architecture of something new. A spec is more compressed and streamlined
summary of everything as it stands today.

If recorded decisions turned out to be lacking, convene a discussion, record the
new decisions here, and then modify the code to match.

Note the context/background should be written in the present tense.

## Table of Contents

The following ADRs are exclusively relevant to CometBFT. For historical ADRs
relevant to Tendermint Core as well, please see [this list](./tendermint-core/).
To distinguish CometBFT ADRs from historical ones from Tendermint Core, we start
numbering our ADRs from 100 onwards.

### Proposed

### Accepted

- [ADR-101: Data companion pull API](adr-101-data-companion-pull-api.md)
- [ADR-102: RPC Companion](adr-102-rpc-companion.md)
- [ADR-103: Protobuf definition versioning](adr-103-proto-versioning.md)
- [ADR-104: State sync from local snapshot](adr-104-out-of-band-state-sync.md)
- [ADR-105: Refactor list of senders in mempool](adr-105-refactor-mempool-senders.md)
- [ADR-106: gRPC API](adr-106-grpc-api)
- [ADR-107: Rename protobuf versions of 0.x releases to pre-v1 betas](adr-107-betaize-proto-versions.md)
- [ADR-109: Reduce CometBFT Go API Surface Area](adr-109-reduce-go-api-surface.md)
- [ADR-111: `nop` Mempool](adr-111-nop-mempool.md)
- [ADR-112: Proposer-Based Timestamps](adr-112-proposer-based-timestamps.md)

### Implemented

### Deprecated

### Rejected

