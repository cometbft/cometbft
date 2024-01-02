package merkle

import (
	"encoding/binary"
	"io"
)

// Tree is a Merkle tree interface.
type Tree interface {
	Size() (size int)
	Height() (height int8)
	Has(key []byte) (has bool)
	Proof(key []byte) (value, proof []byte, exists bool) // TODO make it return an index
	Get(key []byte) (index int, value []byte, exists bool)
	GetByIndex(index int) (key, value []byte)
	Set(key, value []byte) (updated bool)
	Remove(key []byte) (value []byte, removed bool)
	HashWithCount() (hash []byte, count int)
	Hash() (hash []byte)
	Save() (hash []byte)
	Load(hash []byte)
	Copy() Tree
	Iterate(fx func(key, value []byte) (stop bool)) (stopped bool)
	IterateRange(start, end []byte, ascending bool, fx func(key, value []byte) (stop bool)) (stopped bool)
}

// -----------------------------------------------------------------------

// Uvarint length prefixed byteslice.
func encodeByteSlice(w io.Writer, bz []byte) (err error) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	_, err = w.Write(buf[0:n])
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}
