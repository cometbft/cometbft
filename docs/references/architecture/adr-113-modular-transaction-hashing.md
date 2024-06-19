# ADR 113: Modular transaction hashing

## Changelog

- 2024-02-05: First version (@melekes)
- 2024-05-28: Complete refactor (@melekes)
- 2024-06-07: Limit the scope to transaction hashing (@melekes)
- 2024-06-19: Explain why we don't expose this functionality in the CLI (@melekes)

## Status

Proposed

## Context

Hashing in CometBFT is currently implemented using `crypto/tmhash`
package, which itself relies on [`sha256`](https://pkg.go.dev/crypto/sha256).

Among the things which are hashed are the block's header, evidence, consensus
params, commit, partset header, transactions.

### Transaction hashing

The transaction hash is used by:

- the built-in transaction indexer;
- the `/tx` and `/tx_search` RPC endpoints, which allow users
to search for a transaction using its hash;
- mempool to identify transactions.

The problem some application developers are facing is a mismatch between the
internal/app representation of transactions and the one employed by CometBFT. For
example, [Evmos](https://evmos.org/) wants transactions to be hashed using
the [RLP][rlp].

In order to be flexible, CometBFT needs to allow changing the transaction
hashing algorithm if desired by the app developers.

### General hashing

The suggested solution could be used to change the hashing function for all
structs, not just transactions. But the result of such a change is quite
significant. If the chain is using a different hashing scheme, then it looses
IBC-compatibility. The IBC modules assumes fixed hashing scheme. The
destination chain needs to know the hashing function of the source chain in
order to verify the validators hash. So, this remains a future work for now.

## Alternative Approaches

1. Add `TxHashFunc` (transaction hashing function) to `NewNode` as an option
   and pass this function down the stack => avoids gloval variables, but leads
   to a massive API breakage. The problem is we're not 100% sure this will be a
   final solution. So every time we decide to change it, we will be breaking
   tons of API. The suggested solution allows us to be more flexible.
2. Allow changing the hashing function for all structs => breaks IBC
   compatibility (see 'General hashing' above).

## Decision

Give app developers a way to provide their own hash function.

## Detailed Design

Use `sha256` by default, but give developers a way to change the hashing function:

```go
import (
	"crypto"
	"hash"
	"crypto/sha256"
)

var (
    // Hash function used for transaction hashing.
    txHash = crypto.SHA256

    // fmtHash is a function that converts a byte slice to a string.
    fmtHash = func(bz []byte) string {
        return fmt.Sprintf("%X", bz)
    }
)

// SetTxHash sets the hash function used for transaction hashing.
//
// Call this function before starting the node. Changing the hashing function
// after the chain has started can ONLY be done with a hard fork.
func SetTxHash(h crypto.Hash) {
    txHash = h
}

// SetFmtHash sets the function used to convert a checksum to a string.
func SetFmtHash(f func([]byte) string) {
    fmtHash = f
}

// Bytes is a wrapper around a byte slice that implements the fmt.Stringer.
type Bytes []byte

func (bz Bytes) String() string {
    return fmtHash(bz)
}

func (bz Bytes) Bytes() []byte {
    return bz
}

// Sum returns the checksum of the data as Bytes.
func Sum(bz []byte) Bytes {
	return Bytes(TxHash.Hash.Sum(bz))
}
```

Let's break this down. By default, we use `sha256` standard crypto library.
`SetTxHash` allows developers to swap the default hashing function
with the hashing function of their choice. It will be used in:

- mempool;
- transaction indexer;
- `/tx` and `/tx_search` RPC endpoints.

Note the Header's `data_hash` will be different if the default hashing function
is changed.

`SetFmtHash` allows developers to swap the default string function
(`fmt.Sprintf("%X", bz)`) with their own implementation.

The design in the current ADR only aims to support custom hash functions,
it does not support _changing_ the hash function for an existing chain.
If the application developer decides to change the default hashing scheme, they
can only do so once before launching their app. If they attempt to upgrade
after without a hard fork, the resulting hashes won't match. A hard fork would
work.

The majority of chains should still use the default hashing function. That's
why we don't expose this functionality in the CLI or anything like that
(`TxHashFunc` in `NewNode`). Even though the number of chains using a different
hashing function can be significant, it's not the use-case we're optimizing
for. It's good to support it, but it's not the primary goal. Similarly, it's
good to support different p2p protocols, but we're optimizing for the default
one.

## Consequences

### Positive

- Modular transaction hashing

### Neutral

- App developers need to take performance into account when choosing custom
  hash function.

### Negative

- Global variables.

## References

- [tendermint#6539](https://github.com/tendermint/tendermint/issues/6539)
- [tendermint#6773](https://github.com/tendermint/tendermint/pull/6773)

[rlp]: https://ethereum.org/developers/docs/data-structures-and-encoding/rlp
