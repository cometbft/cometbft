---
order: 2
---
# Time

CometBFT provides a Byzantine fault-tolerant source of time.

Time in CometBFT is defined with the [`Time`][spec-time] field of the
block [`Header`][spec-header].

## Properties

The Time produced by CometBFT satisfies the following properties:

- **Time Monotonicity**: time is monotonically increasing.  More precisely, given
  two block headers `H1` of height `h1` and `H2` of height `h2`,
  it is guaranteed that if `h2 > h1` then `H2.Time > H1.Time`.

- **Byzantine Fault Tolerance**: faulty or malicious processes should not be able
  to arbitrarily influence (increase or decrease) the block Time.
  In other words, the Time of blocks should be defined by correct processes.

In addition, the Time produced by CometBFT is expected, by external observers, to provide:

- **Relation to real time**: block times bear some resemblance to real time.
  In other words, block times should represent, within some reasonable accuracy,
  the actual time at which blocks were produced.

## Implementations

CometBFT implements two algorithms for computing block times:

- [BFT Time][bft-time]: the algorithm adopted in versions up to `v0.38.x`;
  available, in legacy mode, in version `v1.x`.

- [Proposer-Based Timestamps][pbts-spec]: introduced in version `v1.x`,
  as a replacement for BFT Time.

Users are strongly encouraged to adopt PBTS in new chains, or to switch to PBTS
when upgrading existing chains.

The table below compares BFT Time and PBTS algorithms in terms of properties:

| Algorithm | Time Monotonicity | Byzantine Fault Tolerance   | Relation to real time |
------------|-------------------|-----------------------------|-----------------------|
| BFT Time  | Guaranteed        | Tolerates `< 1/3` Byzantine nodes     | Best effort and **not** guaranteed |
| PBTS      | Guaranteed        | Tolerates `< 2/3` Byzantine nodes     | Guaranteed within configured synchronous parameters: `PRECISION` and `MSGDELAY` |

[spec-time]: ../core/data_structures.md#time
[spec-header]: ../core/data_structures.md#header
[bft-time]: ./bft-time.md
[pbts-spec]: ./proposer-based-timestamp/README.md
