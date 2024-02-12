# Proposer-Based Timestamps (PBTS)

This document describes a version of the Tendermint consensus algorithm, adopted in CometBFT,
that uses proposer-based timestamps.

PBTS is a Byzantine fault-tolerant algorithm for computing [block times](../time.md).

## Overview

With PBTS, the timestamp of a block is assigned by its
proposer, according with its local clock.
In other words, the proposer of a block also *proposes* a timestamp for the block.
Validators can accept or reject a proposed block.
A block is only accepted if its timestamp is acceptable.
A proposed timestamp is acceptable if it is *received* within a certain time window,
determined by synchronous parameters.

The motivation for introducing this new method for assigning timestamps is
summarized in the [first draft proposal][main_v1].

### Synchronous Parameters

For validating timestamps, PBTS augments the system model considered by the
consensus algorithm with *synchronous assumptions*:

- **Synchronized clocks**: simultaneous clock reads at any two correct validators
differ by at most `PRECISION`;

- **Bounded message delays**: the end-to-end delay for delivering a message to all correct validators
is bounded by `MSGDELAY`.
This assumption is restricted to `Proposal` messages, broadcast by proposers.

`PRECISION` and `MSGDELAY` are consensus parameters, shared by all validators,
that define whether the timestamp of a block is acceptable,
according with the introduced `timely` predicate.

### Timestamp Validation

The `timely` predicate is defined as follows.
Let `proposalReceiveTime` be the time, read from its local clock, at
which a validator receives a `Proposal` message for a `block` with timestamp `ts = block.time`.
The proposed timestamp `ts` can be accepted if both:

 - `ts <= proposalReceiveTime + PRECISION`
 - `ts >= proposalReceiveTime - MSGDELAY - PRECISION`

The following diagram graphically represents the conditions for accepting a proposed timestamp:

![diagram](./diagram.png)

A more detailed and formalized description of the `timely` predicate is available in the
[System Model and Properties][sysmodel] document.

## Implementation

The implementation of PBTS requires some changes in Tendermint consensus algorithm,
summarized below:

- A proposer timestamps a block with the current time, read from its local clock.
The block's timestamp represents the time at which it was assembled
(after the `getValue()` call in line 18 of the [arXiv][arXiv] algorithm):

    - Block timestamps are definitive, meaning that the original timestamp
	is retained when a block is re-proposed (line 16);

    - To preserve monotonicity, a proposer might need to wait until its clock
	reads a time greater than the timestamp of the previous block;

- A validator only prevotes for a block if its timestamp is considered `timely`
(compared to the original algorithm, a check is added to line 23).
Otherwise, the validator prevotes `nil` (line 26):

    - Validators register the time at which they received `Proposal` messages,
	in order to evaluate the `timely` predicate;

    - Blocks that are re-proposed because they received `2f+1 Prevotes`
	in a previous round (line 28) are not subject to the `timely` predicate,
    as their timestamps have already been evaluated at a previous round.

The full solution is detailed and formalized in the [Algorithm Specification][algorithm] document.

## Further details

- [System Model and Properties][sysmodel]
- [Algorithm Specification][algorithm]
- [TLA+ Specification][proposertla]

<!---
## BFT Time

CometBFT provides a deterministic, Byzantine fault-tolerant, source of time,
defined by the `Time` field present in the headers of committed blocks.

In the current consensus implementation, the timestamp of a block is
computed by the [`BFT Time`][bfttime] algorithm:

- Validators include a timestamp in the `Precommit` messages they broadcast.
Timestamps are retrieved from the validators' local clocks,
with the only restriction that they must be **monotonic**:

    - The timestamp of a `Precommit` message voting for a block
	cannot be earlier than the `Time` field of that block;

- The timestamp of a block is deterministically computed from the timestamps of
a set of `Precommit` messages that certify the commit of the previous block.
This certificate, a set of `Precommit` messages from a round of the previous height,
is selected by the block's proposer and stored in the `Commit` field of the block:

    - The block timestamp is the *median* of the timestamps of the `Precommit` messages
	included in the `Commit` field, weighted by their voting power.
	Block timestamps are **monotonic** because
	timestamps of valid `Precommit` messages are monotonic;

Assuming that the voting power controlled by Byzantine validators is bounded by `f`,
the cumulative voting power of any valid `Commit` set must be at least `2f+1`.
As a result, the timestamp computed by `BFT Time` is not influenced by Byzantine validators,
as the weighted median of `Commit` timestamps comes from the clock of a non-faulty validator.

The consensus algorithm does not make any assumptions regarding the clocks of (correct) validators,
as block timestamps have no impact in its operation.
However, the `Time` field of committed blocks is used by other applications that integrate with CometBFT,
including IBC, as well as Cosmos SDK modules such as evidence, staking, and slashing.
And it is used based on the common belief that block timestamps
should bear some resemblance to real time, which is **not guaranteed**.

A more comprehensive discussion of the limitations of `BFT Time`
can be found in the [first draft][main_v1] of this proposal.
Of particular interest is to possibility of having validators equipped with "faulty" clocks,
not fairly accurate with real time, that control more than `f` voting power,
plus the proposer's flexibility when selecting a `Commit` set,
and thus determining the timestamp for a block.
--->

### Open issues

- [tendermint/spec#355: PBTS: evidence][issue355]: not really clear the context, probably not going to be solved.
- [tendermint/spec#372: PBTS: Treat proposal and block parts explicitly in the spec][issue372]
- [tendermint/spec#377: PBTS: margins for proposal times assigned by Byzantine proposers][issue377]

### Closed issues

- [tendermint/spec#353: Proposer time - fix message filter condition][issue353]
- [tendermint/spec#370: PBTS: association between timely predicate and timeout_commit][issue370]
- [tendermint/spec#371: PBTS: should synchrony parameters be adaptive?][issue371]

[main_v1]: ./v1/pbts_001_draft.md

[algorithm]: ./pbts-algorithm.md
[algorithm_v1]: ./v1/pbts-algorithm_001_draft.md

[sysmodel]: ./pbts-sysmodel.md
[sysmodel_v1]: ./v1/pbts-sysmodel_001_draft.md
[timely-predicate]: ./pbts-sysmodel.md#timely-predicate

[proposertla]: ./tla/README.md

[bfttime]: ../bft-time.md
[arXiv]: https://arxiv.org/pdf/1807.04938.pdf

[issue353]: https://github.com/tendermint/spec/issues/353
[issue355]: https://github.com/tendermint/spec/issues/355
[issue370]: https://github.com/tendermint/spec/issues/370
[issue371]: https://github.com/tendermint/spec/issues/371
[issue372]: https://github.com/tendermint/spec/issues/372
[issue377]: https://github.com/tendermint/spec/issues/377
