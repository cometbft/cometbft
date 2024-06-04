# ADR 113: Modular hashing

## Changelog

- 2024-02-05: First version (@melekes)
- 2024-05-28: Complete refactor (@melekes)

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

It's up for a debate whether the hashing function should be changed for the
whole CometBFT or just transactions.

## Alternative Approaches

1. Do nothing => not flexible.
2. Add `HashFn` argument to `NewNode` and pass this function down the stack =>
   complicates the code.
3. Limit the scope of the solution described below to transaction hashing
   (do not change header's hash, evidence's hash, etc.) => ?

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

const (
	// The truncated size of a checksum in bytes.
	TruncatedSize = 20 // validators addresses
)

var (

    // Hash used
    Hash = crypto.SHA256

	stringFunc = func(bz []byte) string {
		return fmt.Sprintf("%X", bz)
	}
)

// New returns a new hash.Hash calculating the given hash function.
func New() hash.Hash {
	return Hash.New()
}

// ReplaceDefaultHashWith replaces the default SHA256 hashing function with a given one.
func ReplaceDefaultHashWith(h crypto.Hash) {
	Hash = h
}

// ReplaceStringFuncWith replaces the default string function (`%X`) with a given one.
func ReplaceStringFuncWith(f func([]byte) string) {
	stringFunc = f
}

// Sum returns the checksum of the data.
func Sum(bz []byte) []byte {
	return New().Sum(bz)
}

// TruncatedSum returns the truncated checksum of the data.
// The checksum is only trimmed when its size is greater than TruncatedSize.
func TruncatedSum(bz []byte) []byte {
    sume := New().Sum(bz)
    if len(sum) < TruncatedSize {
        return sum
    }
	return sum[:TruncatedSize]
}

// FmtHash returns the checksum of the data as a string.
func FmtHash(bz []byte) string {
    return stringFunc(bz)
}
```

Let's break this down. By default, we use `sha256` standard crypto library.
`ReplaceDefaultHashWith` allows developers to swap the default hashing function
with the hashing function of their choice.

`ReplaceStringFuncWith` allows developers to swap the default string function
(`fmt.Sprintf("%X", bz)`) with their own implementation.

The above solution changes hashes across the whole CometBFT, meaning header's
hash, evidence's hash, data's hash, etc.

If the application developer decides to change the default hashing scheme, they
can only do so once before launching their app. If they attempt to upgrade
after without a hard fork, the resulting hashes won't match. A hard fork would
work.

The hashing function needs to be added to `Header`, so that light clients are aware of
the function to use for verification.

```go
type Header struct {
    // ...
	Hash crypto.Hash `json:"hash"`
}
```

Light clients would then use ^ when calculating validators hash or header's hash.

## Consequences

### Positive

- Modular hashing

### Neutral

- App developers need to take performance into account when choosing custom
  hash function.
- IBC is not affected by this change since the proof contains `Hash` (hashing
  function used) and, most importantly, `app_hash` is controlled by the app,
  not CometBFT.

### Negative

- Global variables.

## References

- [Original issue](https://github.com/tendermint/tendermint/issues/6539)

[rlp]: https://ethereum.org/developers/docs/data-structures-and-encoding/rlp
