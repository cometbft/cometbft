---
order: 2
---
# BFT Time

BFT Time is a Byzantine fault-tolerant algorithm for computing [block times](./time.md).

> :warning:
> CometBFT `v1.x` introduced [Proposer-Based Timestamps (PBTS)][pbts-spec],
> intended to be a replacement for BFT Time.
> Users are strongly encouraged to adopt PBTS in new chains, or when upgrading
> existing chains.

## Overview

In order to commit a block, a node needs to receive `Precommit` messages for
the corresponding `BlockID` from validators whose cumulative voting power is
more than `2/3` of the total voting power.
The received `Precommit` messages should refer to the same round, the _commit round_.
A set of `Precommit` messages with the properties above mentioned is a `Commit`.
A `Commit` set of height `H` is included in blocks proposed in height `H+1`.

BFT Time computes the `Time` field of a block proposed in height `H` deterministically
from the `LastCommit` field of the block, which is a `Commit` set from
height `H-1`, using the `MedianTime` method defined as follows:

- `MedianTime`: the weighted median of `Timestamp` fields, of the previous block with
 heights defined by validators' voting powers or, in other words, the median of the
  `Timestamp` fields of the `Precommit` messages forming a `Commit`, where the value of
   each `Timestamp` field is counted a number of times proportional to the voting power
   of the validator that produced and signed that `Precommit` message.
The median of a set of values is one of the values of the set, so that the
`Time` of a proposed block is one of the `Timestamp` fields of the `Precommit`
messages included in the `LastCommit` set of that block.

### Example

Consider the following example:

- We have four validators p1, p2, p3 and p4, with voting power
  distribution: (p1, 23), (p2, 27), (p3, 10), (p4, 10).
  The total voting power is 70, so we assume that the faulty validators have at
most 23 of voting power (since `N = 3F + 1`, where `N` is the total voting
power and `F` is the maximum voting power of faulty validators).
- We have the following `Precommit` messages in some `LastCommit` field (we
ignore all fields except the `Timestamp` field): (p1, 100), (p2, 98), (p3, 1000), (p4, 500).
We assume that p3 and p4 are faulty validators.
- Let's assume that the `block.LastCommit` field contains `Precommit`s of
  validators p2, p3 and p4.
-  The `MedianTime` is then chosen the following way: the value 98 (p2) is
   counted 27 times, the value 1000 (p3) is counted 10 times and the value 500
(p4) is counted also 10 times.  The median value will be `98`.

Notice that, no matter what set of `Precommit` messages with at least `2/3` of
the total voting power we choose, the `MedianTime` value will always be a
value among the `Timestamp` values produced by correct validators.

## Operation

In order to implement BFT Time, validators need to set the `Timestamp` field of
`Precommit` messages they sign and broadcast, and block proposers need to
compute the block `Time` from the `LastCommit` set.

### Vote Time

When producing a `Precommit` message, a validator should set the `Timestamp` field as follows:

1. Let `now` be the clock time of the validator.
2. If `LockedBlock` is defined, set `Timestamp = max(now, LockedBlock.Time + 1ms)`.
3. Else if `ProposalBlock` is defined, set `Timestamp = max(now, ProposalBlock.Time + 1ms)`.
4. Otherwise, set `Timestamp = now`.

The `LockedBlock`, if set, is the block for which the validator is issuing a `Precommit`.
The `ProposalBlock` is the block proposed in that round; in favorable runs, it
matches the `LockedBlock`.


The validator in practice _proposes_ the `Time` for the next block when setting
the `Timestamp` of its `Precommit`.
The proposed `Time` is, by default, the validator's current clock time.
To ensure [Time Monotonicity](./time.md#properties), the `Time` of the next block should be 
higher than the `Time` of the block to be committed in the current height.
So if `now` is smaller than `Time`, the validator proposes the `Time` of the block to be committed
plus a small delta, set to `1ms`.

### Proposed Time

The proposer of a round of consensus produces a block to be proposed.
The proposed block must include a `Commit` set from the commit round of the
previous height, as the block's `LastCommit` field.

The `Time` for the proposed block is then set as `Block.Time = MedianTime(block.LastCommit)`.

Since the block `Time` is produced in a deterministic way, every node that
receives the proposed block, can validate `Block.Time` using the same
procedure.  Block with wrongly computed block times are rejected.

## Properties

BFT Time guarantees the two main [properties](./time.md#properties) for block times:

- **Time Monotonicity**: the [production](#vote-time) of `Timestamp` fields for
  `Precommit` messages at correct validators ensures that the `Time` proposed
   for the next block is higher than the `Time` of the current block.
   Since the `Time` of a block is retrieved from a `Precommit`
   produced by a correct validator, monotonicity is guaranteed.

- **Byzantine Fault Tolerance**: given a `Commit` set that forms the
  `LastCommit` field, a range of [valid values](#proposed-time) for the `Time` field of the
block is defined only by `Precommit` messages produced by correct validators,
i.e., faulty validators cannot arbitrarily influence (increase or decrease) the
`Time` value.  

Notice that the guarantees rely on the fact that the voting power owned by
Byzantine validators is limited, more specifically, is less than 1/3 of the
total voting power, which is also a requirement for the consensus algorithm.

[pbts-spec]: ./proposer-based-timestamp/README.md
