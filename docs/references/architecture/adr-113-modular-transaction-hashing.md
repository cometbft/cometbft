# ADR 113: Modular transaction hashing

## Changelog

- 2024-02-05: First version (@melekes)
- 2024-05-28: Complete refactor (@melekes)
- 2024-06-07: Limit the scope to transaction hashing (@melekes)

## Status

Proposed

## Context

Hashing in CometBFT is currently implemented using `crypto/tmhash`
package, which itself relies on [`sha256`](https://pkg.go.dev/crypto/sha256).

Among the things which are hashed are the block's header, evidence, consensus
params, commit, partset header, transactions and others.

### Transaction hashing

The transaction hash is used by:

- the built-in transaction indexer;
- the `/tx` and `/tx_search` RPC endpoints, which allow users
to search for a transaction using its hash;

The problem some application developers are facing is a mismatch between the
internal/app representation of transactions and the one employed by CometBFT. For
example, [Evmos](https://evmos.org/) wants transactions to be hashed using
the [RLP][rlp].

In order to be flexible, CometBFT needs to allow changing the transaction
hashing algorithm if desired by the app developers.

### General hashing

The suggested solution could've been used to change the hashing function for
all structs, not just transactions. But the implification of that change is
quite significant. For example, if the chain is using a different hashing
scheme, then it looses IBC-compatibility. The IBC modules assumes fixed hashing
scheme. The destination chain needs to know the hashing function of the source
chain in order to verify the validators hash.

## Alternative Approaches

1. Do nothing => not flexible.
2. Add `HashFn` argument to `NewNode` and pass this function down the stack =>
   complicates the code.
3. Allow changing the hashing function for all structs => breaks IBC
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

// SetFmtHash sets the function used to convert a checksum to a string.
func SetFmtHash(f func([]byte) string) {
    fmtHash = f
}

// SetTxHash sets the hash function used for transaction hashing.
func SetTxHash(h crypto.Hash) {
    txHash = h
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
with the hashing function of their choice.

`SetFmtHash` allows developers to swap the default string function
(`fmt.Sprintf("%X", bz)`) with their own implementation.

The above solution only changes transaction hashes (including the header's
`DataHash`).

The design in the current ADR only aims to support custom hash functions,
it does not support _changing_ the hash function for an existing chain.
If the application developer decides to change the default hashing scheme, they
can only do so once before launching their app. If they attempt to upgrade
after without a hard fork, the resulting hashes won't match. A hard fork would
work.

## Consequences

### Positive

- Modular transaction hashing

### Neutral

- App developers need to take performance into account when choosing custom
  hash function.

### Negative

- Global variables.

## References

- [Original issue](https://github.com/tendermint/tendermint/issues/6539)

[rlp]: https://ethereum.org/developers/docs/data-structures-and-encoding/rlp
