# ADR 113: Modular transaction hashing

## Changelog

- 2024-02-05: First version (@melekes)

## Status

TBD

## Context

Transaction hashing in CometBFT is currently implemented using `crypto/tmhash`
package, which itself relies on [`sha256`](https://pkg.go.dev/crypto/sha256) to calculate a transaction's hash.

The hash is then used by the built-in indexer (to index this
transaction) and by the RPC `tx` and `tx_search` endpoints, which allow users
to search for a transaction using a hash.

`tmhash` is also used to calculate various hashes in CometBFT (e.g.,
evidence, consensus params, commit, header, partset header).

The problem some application developers are facing is a mismatch between the
internal/app representation of transactions and the one employed by CometBFT. For
example, [Evmos](https://evmos.org/) wants transactions to be hashed using
the [RLP][rlp].

In order to be flexible, CometBFT needs to allow changing the transaction hashing algorithm
if desired by the app developers.

## Alternative Approaches

None.

## Decision

Give app developers a way to provide their own transaction hash function.

## Detailed Design

Add `hashFn HashFn` option to `NewNode` in `node.go`.

```go
// HashFn defines an interface for the hashing function.
type HashFn interface {
    // New returns a new hash.Hash calculating the given hash function.
    New() hash.Hash
    // Size returns the length, in bytes, of a digest resulting from the given hash function.
    Size() int
}

// Use CustomHashFn to supply custom hashing function.
func CustomHashFn(hashFn HashFn) Option {
	return func(n *Node) {
		n.hashFn = hashFn
	}
}

// DefaultHashFn is a TMHash.
func DefaultHashFn() HashFn {
    return TMHash{}
}

// TMHash uses tmhash.
type TMHash struct {}
func (TMHash) New() hash.Hash {
    return tmhash.New()
}
func (TMHash) Size(bz []byte) {
    return tmhash.Size
}
```

And then use it to calculate a transaction's hash.

## Consequences

### Positive

- Modular transaction hashing
- Paving a way for the pluggable cryptography. hashFn can later be used to
  calculate header's hash, evidence hash.

### Negative

- Potentially adds complexity for developers.

### Neutral

- App developers need to take performance into account when choosing custom
  hash function.

## References

- [Original issue](https://github.com/tendermint/tendermint/issues/6539)

[rlp]: https://ethereum.org/developers/docs/data-structures-and-encoding/rlp
