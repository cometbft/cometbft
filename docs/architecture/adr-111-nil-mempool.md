# ADR 111: `nil` Mempool

## Changelog

- 2023-11-07: First version (@sergio-mena)

## Status

Accepted | Rejected | Deprecated | Superseded by

## Context

### In the Beginning...

Since inception, Tendermint Core and later on CometBFT have featured a mempool
as an internal piece of its architecture.
It is well understood that a dissemination mechanism
(sometimes using _Reliable Broadcast_ [\[HT94\]][HT94] but not necessarily),
is needed in a distributed system implementing State-Machine Replication (SMR).
This is also the case in blockchains.
Early designs such as Bitcoin or Ethereum include an _internal_ component,
responsible for dissemination, called mempool.
Tendermint Core chose to follow the same design given the success
of those early blockchains.

However, the design of ABCI clearly dividing the application logic (i.e., the appchain)
and the consensus logic that provides SMR semantics to the app is a unique innovation in Cosmos
that sets it apart from Bitcoin, Ethereum, and many others.
This clear separation of concerns entailed many consequences, mostly positive:
it allows CometBFT to be used underneath (currently) tens of different appchains in production
in the Cosmos ecosystem and elsewhere.
But there are other implications for having an internal mempool
in CometBFT: the interaction between the mempool, the application, and the network
becomes more indirect, and thus more complex and hard to understand and operate.

### ABCI++ Improvements and Remaining Shortcomings

Before the release of ABCI++, `CheckTx` was the main mechanism the app had at its disposal to influence
what made it to the mempool, and very indirectly what transactions got ultimately proposed in a block.
Since ABCI 1.0 (the first part of ABCI++, shipped in `v0.37.x`), the application now has
a say in what is proposed with `PrepareProposal` and `ProcessProposal`.

This has greatly improved the ability for appchains to influence the contents of proposed block.
Further, it has enabled many new use cases for appchains. However some issues remain with
the current model:

* We are using the same p2p network for disseminating transactions and consensus-related messages.
* Many mempool parameters are configured on a per-node basis by node operators,
  allowing the possibility of inconsistent mempool configuration across the network
  with potentially serious scalability effects
  (even causing unacceptable performance degradation in some extreme cases).
* The mempool is using a basic (robust but suboptimal) flood algorithm
  * the CometBFT team are working on improving it as one of our current priorities,
    but any improvement we come up with must address the needs of a vast spectrum of applications,
    as well as be heavily scaled-tested in various scenarios
    (in an attempt to cover the applications' wide spectrum)
  * a mempool designed specifically for one particular application
    would reduce the search space as its designers can devise it with just their application's
    needs in mind.
* The interaction with the application is still somewhat convoluted:
  * the app has to decide what logic to implement in `CheckTx`,
    what to do with the transaction list coming in `RequestPrepareProposal`,
    whether it wants to maintain an app-side mempool (more on this below), and whether or not
    to combine the transactions in the app-side mempool with those coming in `RequestPrepareProposal`
  * all those combinations are hard to fully understand, as the semantics and guarantees are
    often not clear
  * when using exclusively an app-mempool (the approach taken in the Cosmos SDK `v0.47.x`)
    for populating proposed blocks, with the aim of simplifying the app developers' life,
    we still have a suboptimal model where we need to continue using CometBFT's mempool
    in order to disseminate the transactions. So, we end up using twice as much memory,
    as in-transit transactions need to be kept in both mempools.

The approach presented in this ADR builds on the app-mempool design released in `v0.47.x`
of the Cosmos SDK,
and briefly discussed in the last bullet point above (see [SDK app-mempool][sdk-app-mempool] for further details of this model).

In the app-mempool design in Cosmos SDK `v0.47.x`
an unconfirmed transaction must be both in CometBFT's mempool for dissemination and
in the app's mempool so the application can decide how to manage the mempool.
The many advantages of this approach are beyond question. However it has some implications:

* Having every transaction both in CometBFT and in the application is suboptimal in terms of memory.
  Additonally, the app developer has to be careful
  that the contents of both mempools does not diverge over time
  (hence the crucial role `re-CheckTx` plays post-ABCI++).
* The main reason for a transaction needing to be in CometBFT's mempool is
  because the design in Cosmos SDK `v0.47.x` does not consider an application
  that has its own means of disseminating transactions.
  It reuses the peer to peer network underneath CometBFT reactors.
* So, if an app has an ad-hoc design of how to disseminate transactions,
  there is no point in having transactions in CometBFT's mempool.

This proposal targets this kind of applications:
those that have an ad-hoc (and likely more efficient) mechanism for transaction dissemination.
We propose to introduce a `nil` mempool which will effectively act as a stubbed object
internally:

* it will reject any transaction being locally submitted or gossipped by a peer
* when it is time to _reap_ (as it is currently called) the mempool the answer will always be empty
* the application running on the proposer validator will add transactions it received
  using the appchains's own mechanism via `PrepareProposal`.

## Alternative Approaches

These are the alternatives known to date:

1. Keep the current model. Useful for basic apps, but clearly suboptimal for applications
   with their own mechanism to disseminate transactions and particular performance requirements.
2. Provide more efficient general-purpose mempool implementations.
   This is an ongoing effort, but will take some time (and R&D effort) to come up with
   advanced mechanisms -- likely highly configurable -- which then will have to be thoroughly tested.
3. A similar approach to this one ([ADR110][adr-110]) whereby the application-specific
   mechanism directly interacts with CometBFT via a newly defined gRPC interface.

## Decision

TBD

## Detailed Design

What this ADR proposes can already be achieved with an unmodified CometBFT since `v0.37.x`,
albeit with a complex, poor UX.
This core of this proposal is to make some internal changes so it is clear an simple for app developers,
thus improving the UX.

#### `nil` Mempool

We propose a new mempool implementation, called `nil` Mempool, that effectively disables all mempool functionality
within CometBFT.
The `nil` Mempool implements the `Mempool` interface in a very simple manner:

*	`CheckTx(tx types.Tx) (*abcicli.ReqRes, error)`: returns `nil, ErrNotAllowed`
*	`RemoveTxByKey(txKey types.TxKey) error`: returns `ErrNotFound`
* `ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs`: returns `nil`
* `ReapMaxTxs(max int) types.Txs`: returns `nil`
*	`Lock()`: does nothing
* `Unlock()`: does nothing
*	`Update(...) error`: returns `nil`
* `FlushAppConn() error`: returns `nil`
*	`Flush()`: does nothing
* `TxsAvailable() <-chan struct{}`: returns a closed channel
* `EnableTxsAvailable()`: does nothing
* `SetTxRemovedCallback(cb func(types.TxKey))`: does nothing
* `Size() int` returns 0
* `SizeBytes() int64` returns 0

Upon startup, the `nil` mempool reactor will advertise no channels to the peer to peer layer.

### Configuration

We propose the following changes to the `config.toml` file:

```toml
[mempool]
# The type of mempool for this CometBFT node to use.
#
# Valid types of mempools supported by CometBFT:
# - "flood" : clist mempool with flooding gossip protocol (default)
# - "nil"   : nil-mempool (app needs an alternative tx dissemination mechanism)
type = "nil"
```

The config validation logic will be modified to add a new rule that rejects a configuration file if:

* the mempool is set to `nil`
* `create_empty_blocks`, in `consensus` section, is set to `false`.

### Impacted Workflows

* *Submitting a transaction*. Users are not to submit transactions via CometBFT's RPC.
  `BroadcastTx*` RPC methods will fail.
  The application running on a full node must offer an interface for users to submit new transactions.
  It could also be a distinct node (or set of nodes) in the network.
  These considerations are exclusively the application's concern in this approach.
* *Time to propose a block*. The consensus reactor will call `ReapMaxBytesMaxGas` which will return a `nil` slice.
  `RequestPrepareProposal` will thus contain no transactions.
* *Consensus waiting for transactions to become available*. `TxsAvailable()` returns a closed channel,
  so consensus doesn't block (potentially producing empty blocks).
  At any rate, a configuration with the `nil` mempool and `create_empty_blocks` set to `false`
  will be rejected at the first place.
* *A new block is decided*.
  * When `Update` is called, nothing is done (no decided transaction is removed).
  * Locking and unlocking the mempool has no effect.

### Impact on Current Release Plans

The changes needed for this approach, are fairly simple, and the logic is clear.
This might allow us to even deliver it as part of CometBFT `v1` (our next release)
even without a noticeable impact on `v1`'s delivery schedule.

The CometBFT team (learning from past dramatic events) usually takes a conservative approach
for backporting changes to release branches that have already undergone a full QA cycle
(and thus are in code-freeze mode).
For this reason, although the limited impact of these changes would limit the risks
of backporting to `v0.38.x` and `v0.37.x`, a careful risk/benefit evaluation will
have to be carried out.

Backporting to `v0.34.x` does not make sense as this version predates the release of `ABCI 1.0`,
so using the `nil` mempool renders CometBFT's operation useless.

>TODO: Need to add a small section here explaining the dilemma between a Consensus Param and a
>      `config.toml` field.

## Consequences

### Positive

- Applications can now find mempool mechanisms that fit better their particular needs:
  - Ad-hoc ways to add, remove, merge, reorder, modify, prioritize transactions according
    to application needs.
  - A way to disseminate transactions (gossip-based or other) to get the submitted transactions
    to proposers. The application developers can devise simpler, efficient mechanisms tailored
    to their application.
  - Back-pressure mechanisms to prevent malicious users from abusing the transaction
    dissemination mechanism.
- In this approach, CometBFT's peer-to-peer layer does not need to deal with transaction gossipping,
  and its resources can be used by other reactors such as consensus, evidence,
  block-sync, or state-sync.
- There is no risk for the operators of a network to provide inconsistent configurations
  for some mempool-related parameters. Some of those misconfigurations are known to have caused
  serious performance issues in CometBFT's peer to peer network.
  Unless, of course, the application-defined transaction dissemination mechanism ends up
  allowing similar configuration inconsistencies.
- The interaction between the application and CometBFT at `PrepareProposal` time
  is simplified. No transactions are ever provided by CometBFT,
  and no transactions can ever be left in the mempool when CometBFT calls `PrepareProposal`:
  the application trivially has all the information.
- UX is improved compared to how this can be done prior to this ADR.

### Negative

- With the `nil` mempool, it is up to the application to provide users with a way
  to submit transactions and deliver those transactions to validators.
  This is a considerable endeavor, and more basic appchains may consider it is not worth the hassle.


[sdk-app-mempool]: [https://docs.cosmos.network/v0.47/build/building-apps/app-mempool]
[adr-110]: [https://github.com/cometbft/cometbft/pull/1565]
[HT94]: [https://dl.acm.org/doi/book/10.5555/866693]