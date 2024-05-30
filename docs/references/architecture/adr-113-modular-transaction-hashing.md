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
const (
    Size          = 32 // all hashes
    TruncatedSize = 20 // except validator's address
)

var (
    hashFunc = func() Hash256 {
        return sha256.New()
    }
)

type Hash256 interface {
    hash.Hash
    // Fix the size where it's needed: block hash, evidence hash, ...
    Sum256(bz []byte) []byte
    // Allow custom representation of a resulting bytes
    String(bz []byte) string
}

func New() Hash256 {
    return hashFunc()
}

func ReplaceDefaultWith(f func() Hash256) {
    hashFunc = f
}

// For fixed size hashes.
func Sum256(bz []byte) [32]byte {
    return hashFunc().Sum256(bz)
}

// For validators addresses.
func Sum160(bz []byte) [20]byte {
    return hashFunc().Sum256(bz)[:TruncatedSize]
}
```

```go
import (
    gosha256 "crypto/sha256"
)

type sha256 struct {
   h hash.Hash
}

func New() Hash256 {
    return sha256{
        h: gosha256.New(),
    }
}

func (h sha256) Write(p []byte) (n int, err error) {
	return h.h.Write(p)
}

func (h sha256) Sum(b []byte) []byte {
	return h.h.Sum(b)
}

func (h sha256) Reset() {
	h.h.Reset()
}

func (sha256) Size() int {
	return gosha256.Size
}

func (h sha256) BlockSize() int {
	return h.h.BlockSize()
}

func (h sha256) Sum256(bz []byte) []byte {
	return h.h.Sum256(bz)
}

func (h sha256) String(bz []byte) string {
	return fmt.Sprintf("%X", bz)
}
```


## Consequences

### Positive

- Modular transaction hashing

### Neutral

- App developers need to take performance into account when choosing custom
  hash function.

### Negative

## References

- [Original issue](https://github.com/tendermint/tendermint/issues/6539)

[rlp]: https://ethereum.org/developers/docs/data-structures-and-encoding/rlp
