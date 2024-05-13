# ADR 115: Predictable Block Times

## Changelog

 - April 30 2024: Created by @melekes
 - May 13 2024: Updated by @melekes

## Status

**Proposed**

## Context

Let's say you're an ABCI application developer and you want to have
constant block times: 1 block each 6s. How would you do that?

You found out that proposing a block and voting on it takes roughly 2s in your
network. You instruct validators in your network to change `timeout_commit`
(_how long a node waits for additional precommits after committing a block,
before starting on the next height_) from 1s to 4s in the node's config file.

Do you have predictable block times now?

No - you don't. The expected block time will be around 6s, but even in
favorable runs it will drift far apart from 6s due to:

1. CometBFT going into multiple rounds of consensus (it happens rarely but
   still).
2. Network latency.
3. Clock drifts.
4. The delay for processing a block.

A validator could also change `timeout_commit`  to 0s in its node's config file.
That means whenever this validator proposes a block, the block time will be 2s
(not 6s!).

Because 1-3 are out of your (and our) control, **we can't have constant block
times**. But we can design a mechanism so that the medium-to-long term average
block time converges to a desired value.

To achieve that, we need to define a form of variable block time. Namely, if a
block takes longer than expected, we should be able to render the next block(s)
faster in order to converge to the desired (average) block time.

### Use Cases

In case of Osmosis, committing a big block is expected to take longer than
usual. If we know the approximate size of the next block, we could increase
`timeout_commit` dynamically (if such feature existed) to give the state
machine some extra time to finish execution.

In case of Berachain, they want to have a constant block time to match
Ethereum's slots, which are equal to 12s.

## Proposal

Remove `timeout_commit` in favor of a new field in `FinalizeBlockResponse`:
`next_block_delay`. This field's semantics stays essentially the same: delay
between the time when the current block is committed and the next block is
proposed.

If Comet goes into multiple rounds of consensus, the ABCI application can react
by lowering `next_block_delay`. Of course, nobody could guarantee that there
won't be 100000 rounds of consensus, so **it's still best effort** when it
comes to average block times.

## Alternative Approaches

1. Do nothing. The block time will stay approximate.
2. Make `timeout_commit` global (= consensus parameter). It doesn't solve the
   problem of multiple rounds of consensus.
3. Add `next_block_delay`, but keep `timeout_commit`. Validators have control
   over block times.
4. Add `next_block_delay`, but keep `timeout_commit` and make it global. It's
   confusing to have two parameters controlling the same behavior.

## Detailed Design

Remove `timeout_commit` from the config and add `next_block_delay` to `FinalizeBlockResponse`.

```protobuf
message FinalizeBlockResponse {
  // ...
  // The delay between this block and the time when the next block is proposed.
  google.protobuf.Duration next_block_delay = 6
      [(gogoproto.nullable) = false, (gogoproto.stdduration) = true];
}
```

Block proposals received before the last block is committed + `next_block_delay` MUST be rejected.
A proposer MUST wait until the last block is committed + `next_block_delay` to propose a block.

### Specification

Timeout estimate in the spec should be updated to reflect `next_block_delay`:

```
block(i+1).Time > block(i).Time + NEXTBLOCKDELAY
```

See [this comment][spec-comment] for more details.

## Consequences

### Positive

- ABCI application developers will have more control over block times.

### Neutral

- Comet will not guarantee constant block times.

### Negative

- ABCI application developers will need to set `next_block_delay` to `1s`
  to preserve the old behavior if we remove `timeout_commit` from the node's
  config file.

## References

* [cometbft/cometbft#2655](https://github.com/cometbft/cometbft/issues/2655)
* [tendermint/tendermint#5911](https://github.com/tendermint/tendermint/issues/5911)
* [discussion about timeout params](https://github.com/cometbft/cometbft/discussions/2266)

[spec-comment]: https://github.com/tendermint/tendermint/issues/5911#issuecomment-804889910
