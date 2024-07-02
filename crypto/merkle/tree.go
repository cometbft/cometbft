package merkle

import (
	"crypto/sha256"
	"hash"
	"math/bits"
)

// HashFromByteSlices computes a Merkle tree where the leaves are the byte slice,
// in the provided order. It follows RFC-6962. It uses SHA256 as the hash function.
func HashFromByteSlices(items [][]byte) []byte {
	return HashFromByteSlicesWithHash(sha256.New(), items)
}

// HashFromByteSlicesWithHash computes a Merkle tree where the leaves are the byte slice,
// in the provided order. It follows RFC-6962. It uses the provided hash function.
func HashFromByteSlicesWithHash(h hash.Hash, items [][]byte) []byte {
	switch len(items) {
	case 0:
		return emptyHash(h)
	case 1:
		return leafHash(h, items[0])
	default:
		k := getSplitPoint(int64(len(items)))
		left := HashFromByteSlicesWithHash(h, items[:k])
		right := HashFromByteSlicesWithHash(h, items[k:])
		return innerHash(h, left, right)
	}
}

// getSplitPoint returns the largest power of 2 less than length.
func getSplitPoint(length int64) int64 {
	if length < 1 {
		panic("Trying to split a tree with size < 1")
	}
	uLength := uint(length)
	bitlen := bits.Len(uLength)
	k := int64(1 << uint(bitlen-1))
	if k == length {
		k >>= 1
	}
	return k
}
