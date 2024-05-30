# ADR 113: Modular transaction hashing

## Changelog

- 2024-02-05: First version (@melekes)
- 2024-05-28: Complete refactor (@melekes)

## Status

Proposed

## Context

Transaction hashing in CometBFT is currently implemented using `crypto/tmhash`
package, which itself relies on [`sha256`](https://pkg.go.dev/crypto/sha256) to
calculate a transaction's hash.

The hash is then used by:

- the built-in transaction indexer;
- the `/tx` and `/tx_search` RPC endpoints, which allow users
to search for a transaction using its hash;
- `types` package to calculate the Merkle root of block's transactions.

`tmhash` library is also used to calculate various hashes in CometBFT repo (e.g.,
evidence, consensus params, commit, header, partset header).

The problem some application developers are facing is a mismatch between the
internal/app representation of transactions and the one employed by CometBFT. For
example, [Evmos](https://evmos.org/) wants transactions to be hashed using
the [RLP][rlp].

In order to be flexible, CometBFT needs to allow changing the transaction
hashing algorithm if desired by the app developers.

## Alternative Approaches

1. Do nothing => not flexible.
2. Add `HashFn` argument to `NewNode` and pass this function down the stack =>
   complicates the code.

## Decision

Give app developers a way to provide their own transaction hash function.

## Detailed Design

Use `sha256` by default, but give developers a way to change the hashing function:

```go
import (
	"hash"
	"crypto/sha256"
)

var (
	// The size of a checksum in bytes.
	Size = sha256.Size // all hashes
	// The truncated size of a checksum in bytes.
	TruncatedSize = 20 // except validators addresses

	hashFunc = func() hash.Hash {
		return sha256.New()
	}

	stringFunc = func(bz []byte) string {
		return fmt.Sprintf("%X", bz)
	}
)

// New returns a new hash.Hash calculating the given hash function.
func New() hash.Hash {
	return hashFunc()
}

// ReplaceHashFuncWith replaces the default SHA256 hashing function with a given one.
func ReplaceHashFuncWith(f func() hash.Hash, size, truncatedSize int) {
	hashFunc = f
	Size = size
	TruncatedSize = truncatedSize
}

// ReplaceStringFuncWith replaces the default string function (`%X`) with a given one.
func ReplaceStringFuncWith(f func([]byte) string) {
	stringFunc = f
}

// Sum returns the checksum of the data.
func Sum(bz []byte) []byte {
	return hashFunc().Sum(bz)
}

// TruncatedSum returns the truncated checksum of the data.
func TruncatedSum(bz []byte) []byte {
	return hashFunc().Sum(bz)[:TruncatedSize]
}
```

Let's break this down. By default, we use `sha256` standard crypto library.
`ReplaceHashFuncWith` allows developers to swap the default hashing function
with the hashing function of their choice. Not that `size` and `truncatedSize`
are also configurable, which doesn't limit us to 32 byte checksums.

`ReplaceStringFuncWith` allows developers to swap the default string function
(`fmt.Sprintf("%X", bz)`) with their own implementation.

The above solution changes hashes across the whole CometBFT, meaning header's
hash, evidence's hash, data's hash, etc.

If the application developer decides to change the default hashing scheme, they
can only do so once before launching their app. If they attempt to upgrade
after, the resulting hashes won't match. A hard fork is an option, but they
have to be careful with the past data.

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
