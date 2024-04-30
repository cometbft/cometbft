# ADR 115: Predictable Block Times

## Changelog

 - April 30 2024: Created by @melekes

## Status

**Proposed**

## Context

Let's say you're an ABCI application developer and you want to have
predictable block times: 1 block each 6s. How would you do that?

You found out that proposing a block and voting on it takes roughly 2s in your
network. You change `timeout_commit` (_how long a node waits for additional
precommits after committing a block, before starting on the next height_) from
1s to 4s in the default config and ship your app with that modified config. Do
you have predictable block times now?

No - you don't. The block time will be around 6s, but it won't be exactly 6s. A
validator could set `timeout_commit`  to 0s because it's a local parameter,
which is not enforced by Comet. That means whenever this validator proposes a
block, the overall block time will be 2s (not 6s!).

There are other reasons why the block time might not be predictable:

1. CometBFT can go into multiple rounds of consensus, which will increase the
   block time.
2. Network latency.
3. Clock drifts.

By now, many teams (Celestia, Berachain, Osmosis, etc.) expressed their
interest in either a) having predictable block times OR b) having the ability
to alter the next block time.

In case of Osmosis, committing a big block might take longer than usual. They
could increase the block time for big blocks to give the state machine some
extra time to finish execution if that was possible.

In case of Berachain, they want to have a predictable block time to match
Ethereum's slots, which are equal to 12s.

## Decision

Remove `timeout_commit` in favor of a new field in `FinalizeBlockResponse`:
`next_block_delay`. This field's semantics stays essentially the same: delay
between this block and the time when the next block is proposed.

If Comet goes into multiple rounds of consensus, the ABCI application can react
by lowering `next_block_delay`. Of course, nobody could guarantee that there
won't be 100000 rounds of consensus, so **it's still best effort** when it
comes to predictable block times.

## Alternative Approaches

1. Do nothing. The block time will stay approximate.
2. Make `timeout_commit` global (= consensus parameter). It doesn't solve the
   problem of multiple rounds of consensus.
3. Add `next_block_delay`, but keep `timeout_commit`. Predictable block times
   are still not possible since `timeout_commit` is local.
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

Block proposals received before last block time + `next_block_delay` MUST be rejected.
A proposer MUST wait until last block time + `next_block_delay` to propose a block.

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

- Comet will not guarantee predictable block times to 100%.

### Negative

- ABCI application developers will need to set `next_block_delay` to `1s`
  to preserve the old behavior.

## References

* [cometbft/cometbft#2655](https://github.com/cometbft/cometbft/issues/2655)
* [tendermint/tendermint#5911](https://github.com/tendermint/tendermint/issues/5911)
* [discussion about timeout params](https://github.com/cometbft/cometbft/discussions/2266)

[spec-comment]: https://github.com/tendermint/tendermint/issues/5911#issuecomment-804889910
