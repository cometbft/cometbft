package merkle

import (
	"hash"
)

// TODO: make these have a large predefined capacity.
var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

// returns h(<empty>).
func emptyHash(h hash.Hash) []byte {
	h.Reset()
	h.Write([]byte{})
	return h.Sum(nil)
}

// returns h(0x00 || leaf).
func leafHash(h hash.Hash, leaf []byte) []byte {
	h.Reset()
	h.Write(leafPrefix)
	h.Write(leaf)
	return h.Sum(nil)
}

// returns h(0x01 || left || right).
func innerHash(h hash.Hash, left []byte, right []byte) []byte {
	h.Reset()
	h.Write(innerPrefix)
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}
