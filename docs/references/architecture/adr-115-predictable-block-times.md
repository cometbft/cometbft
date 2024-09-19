# ADR 115: Predictable Block Times

## Changelog

 - April 30 2024: Created by @melekes
 - May 13 2024: Updated by @melekes
 - June 11 2024: Mark as accepted @melekes

## Status

**Accepted**

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

A validator could also change `timeout_commit` to 0s in its node's config file.
That means whenever THIS validator proposes a block, the block time will be 2s
(not 6s!).

Note that the value of `timeout_commit` is static and can't be changed
without restarting the node.

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

Celestia's usecase here for the medium term is to be able to have a longer
proposal timeout so that they can spend a larger percentage of the block time
gossiping while having consistent block times.

## Proposal

Move `timeout_commit` into `FinalizeBlockResponse` as `next_block_delay`. This
field's semantics stays essentially the same: delay between the time when the
current block is committed and the next height is started. The idea is
literally to have the same behavior as `timeout_commit`, while allowing the
application to pick a different value for each height.

If Comet goes into multiple rounds of consensus, the ABCI application can react
by lowering `next_block_delay`. Of course, nobody could guarantee that there
won't be 100000 rounds of consensus, so **it's still best effort** when it
comes to individual block times.

## Alternative Approaches

1. Do nothing. The block time will stay largely unpredictable.
2. Make `timeout_commit` global (= consensus parameter). It doesn't solve the
   problem of multiple rounds of consensus + the execution is delayed by one
   block (as with other consensus parameters).
3. Add `next_block_delay`, but keep `timeout_commit`. Individual validators can
   still change `timeout_commit` in their node's config file and mess up the
   block times.
4. Add `next_block_delay`, but keep `timeout_commit` and make it global. It's
   confusing to have two parameters controlling the same behavior.

## Detailed Design

Move `timeout_commit` from the config into `FinalizeBlockResponse` as `next_block_delay`.

```protobuf
message FinalizeBlockResponse {
  // ...
  // The delay between this block and the time when the next block is proposed.
  google.protobuf.Duration next_block_delay = 6
      [(gogoproto.nullable) = false, (gogoproto.stdduration) = true];
}
```

A correct proposer MUST wait until the last block is committed + `next_block_delay` to propose a block.
A correct validator MUST wait until the last block is committed + `next_block_delay` to start the next height.

`next_block_delay` is a non-deterministic field (unlike most fields in
`FinalizeBlockResponse`), that is: it is not part of the replicated data. This
means that each node may provide a different value, which is supposed to depend
on how longs things are taking at the local node. Or it can replicate the
existing behavior (fixed `timeout_commit`) by always returning a constant value
(e.g. "3s").

### ABCI application

In order to leverage this feature most applications:

* need to use real --wallclock-- time;
* mandate it's nodes to have synchronized clocks (NTP, or other). This is
  not a big deal since PBTS also requires this;
* `time` field in `PrepareProposalRequest`, `ProcessProposalRequest` and
  `FinalizeBlockRequest` could be trusted when using PBTS.

### Specification

Timeout estimate in the spec should be updated to reflect `next_block_delay`:

```
block(i+1).ProposalTime > block(i).CommitTime + NEXTBLOCKDELAY
```

See [this comment][spec-comment] for more details.

### Upgrade path

* keep `timeout_commit` deprecated;
* if a value for `timeout_commit` is detected at process start-up, warn the
  user that they are using a deprecated field;
* if the app provides a value for `next_block_delay`, then `timeout_commit` is
  ignored;
* if the app does not provide a value for `next_block_delay`, then CometBFT falls
  back to `timeout_commit`.

## Consequences

### Positive

- ABCI application developers will have more control over block times.

### Neutral

### Negative

## References

* [cometbft/cometbft#2655](https://github.com/cometbft/cometbft/issues/2655)
* [tendermint/tendermint#5911](https://github.com/tendermint/tendermint/issues/5911)
* [discussion about timeout params](https://github.com/cometbft/cometbft/discussions/2266)

[spec-comment]: https://github.com/tendermint/tendermint/issues/5911#issuecomment-804889910
