# ADR 105: Refactor list of senders in mempool

## Changelog

- 2023-06-26: Initial draft (@hvanz)

## Status

Proposed

## Context

Before adding a transaction to the mempool, we need to send a `CheckTx` message
to the application for validating the transaction. There are two variants of
this message, distinghished by the value in message field `type`:
- `CheckTxType_New` is for transactions that need to be validated before adding
it to the mempool for the first time.
- `CheckTxType_Recheck` is for transactions that are in the mempool and need to
be re-validated after committing a block and advancing to the next height.

The mempool communicates with the ABCI server (that is, the application) by
sending `abci.Request`s through the proxy `AppConnMempool`. The proxy provides a
callback mechanism for handling `abci.Response`s. The current mempool
implemention `CListMempool` (also called v0) utilizes this mechanism for
`Recheck` transactions but not for `New` transactions. Instead `New`
transactions require an ad-hoc mechanism for each request. 

The reason behind this difference is that for `New` transactions we need to
record the ID of the peer that sent the transaction. However, this information
is not included in `RequestCheckTx` messages. Recording the sender's ID is
necessary for the transaction propagation protocol, which uses the recorded list
of senders to prevent sending the transaction back to these peers, thus avoiding
duplicated messages. More importantly, this mechanism serves as the only means
to stop propagating transactions.

There are two design problems with this implementation. First, there is a
complex pattern for handling callbacks on `New` requests. The following code
snippet at the end of the `CheckTx` method, where transactions received for the
first time are processed, demonstrates the issue:
``` golang
	reqRes, err := mem.proxyAppConn.CheckTxAsync(context.TODO(), &abci.RequestCheckTx{Tx: tx})
	reqRes.SetCallback(mem.reqResCb(tx, txInfo, cb))
```
When we send a request for validating a transaction via `CheckTxAsync`, it
returns a `ReqRes` object. To handle the response asynchronously, we set an
ad-hoc callback `reqResCb` on `ReqRes`. This callback is different for each
transaction `tx` because it's parameterized by `tx`, `txInfo` (which essentially
contains the sender ID), and another callback function `cb` provided by the
caller of `CheckTx` to be applied on the response.

Secondly, the list of senders for each transaction is recorded directly in the
mempool's `txs` data structure. However, the list of senders is an important
component of the propagation protocol and it should be part of the reactor,
while `txs` is part of the implementation of this specific version (v0) of the
mempool.

In this document, we propose a solution that involves moving the list of senders
from the mempool implementation to the reactor. This change will simplify the
code, establish a more clear separation of the propagation protocol and the data
structure, and allow for future improvements to the mempool as a whole.

## Alternative Approaches

> This section contains information around alternative options that are considered
> before making a decision. It should contain a explanation on why the alternative
> approach(es) were not chosen.

## Decision

> This section records the decision that was made.
> It is best to record as much info as possible from the discussion that happened.
> This aids in not having to go back to the Pull Request to get the needed information.

## Detailed Design
 
We propose the following changes to the mempool's reactor and data structure.

- In the `Mempool` interface, change the signature of `CheckTx` from
    ``` golang
    CheckTx(tx types.Tx, cb func(*abci.ResponseCheckTx), txInfo TxInfo) error
    ```
    to
    ``` golang
    CheckTx(tx types.Tx) (abcicli.ReqRes, error)
    ```
  - The returning `ReqRes` object can be used to set and invoke a callback to
    handle the response, if needed.
  - The callback parameter `cb` is no longer needed. Currently, this is mainly
    used by the RPC endpoints `broadcast_tx_sync` and `broadcast_tx_commit`, but
    the same functionality can be obtained with `ReqRes`.
  - `txInfo` contains information about the sender, which is also no longer
    needed, as justified by the next point.
- The list of senders for each transaction is stored in `mempoolTx`, the entries
  for `txs`. Move the senders out of `mempoolTx` to a new map `txSenders` of
  type `map[types.TxKey]map[uint16]bool` in the mempool reactor. `txSenders`
  would map transaction keys to a set (a map to booleans) of peer ids. Add also
  a `cmtsync.RWMutex` lock to handle concurrent accesses to the map.
  - This refactoring should not change the fact that the list of senders live as
    long as the transactions are in the mempool. When a transaction is received
    by the reactor (either via RPC or P2P), we call `CheckTx`. We know whether a
    transaction is valid and was included in the mempool by reading the `ReqRes`
    response. If this is the case, record the list of senders in `txSenders`.
    When a transaction is removed from the mempool, notify the reactor to remove
    the list of senders for that transaction.
- In `CListMempool`, introduce a new channel `txsRemoved` of type `chan
  types.TxKey` to notify the mempool reactor that a transaction was removed from
  the mempool.
- In the mempool reactor, introduce a new goroutine to handle incoming
  transaction keys in the `txsRemoved` channel. For each key received, update
  `txSenders`.
- In `CListMempool`, `resCbFirstTime` is the function that handles responses of
  type `CheckTxType_New`. Instead of setting it as an ad-hoc callback on each
  transaction, we could now call it directly from `globalCb`, where responses of
  type `CheckTxType_Recheck` are already being handled.

## Consequences

The refactoring proposed here would not affect how users and other peers
interact with the mempool. It will only change how transaction metadata is
stored internally.

### Positive

- Get rid of the complex design pattern of callbacks for handling
  `CheckTxType_New` responses.
- Clear separation of propagation protocol in the reactor and the mempool
  implementation.
- Allow for future improvements to both the propagation protocol and the mempool
  implementation.

### Negative

None

### Neutral

- A new channel in `CListMempool` for notifying removed transactions. This is
  the same mechanism used to notify the consensus reactor when there available
  transactions to include in a block. A new goroutine in the reactor handling
  incoming messages in the new channel. This will be used for updating the list
  of senders when a transaction is removed from the mempool.

## References

> Are there any relevant PR comments, issues that led up to this, or articles
> referenced for why we made the given design choice? If so link them here!

- {reference link}
