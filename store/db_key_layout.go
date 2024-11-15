package store

import (
	"encoding/hex"
	"strconv"

	"github.com/google/orderedcode"
)

type BlockKeyLayout interface {
	CalcBlockMetaKey(height int64) []byte

	CalcBlockPartKey(height int64, partIndex int) []byte

	CalcBlockCommitKey(height int64) []byte

	CalcSeenCommitKey(height int64) []byte

	CalcExtCommitKey(height int64) []byte

	CalcBlockHashKey(hash []byte) []byte
}

// v1LegacyLayout is a legacy implementation of BlockKeyLayout, kept for backwards
// compatibility. Newer code should use [v2Layout].
type v1LegacyLayout struct{}

// In the following [v1LegacyLayout] methods, we preallocate the key's slice to speed
// up append operations and avoid extra allocations.
// The size of the slice is the length of the prefix plus the length the string
// representation of a 64-bit integer. Namely, the longest 64-bit int has 19 digits,
// therefore its string representation is 20 bytes long (19 digits + 1 byte for the
// sign).

// CalcBlockCommitKey implements BlockKeyLayout.
// It returns a database key of the form "C:<height>" to store/retrieve the commit
// of the block at the given height to/from the database.
func (*v1LegacyLayout) CalcBlockCommitKey(height int64) []byte {
	const prefixLen = len("C:")

	key := make([]byte, prefixLen, prefixLen+20)

	key[0], key[1] = 'C', ':'
	key = strconv.AppendInt(key, height, 10)

	return key
}

// CalcBlockHashKey implements BlockKeyLayout.
// It returns a database key of the form "BH:hex(<hash>)" to store/retrieve a block
// to/from the database using its hash.
func (*v1LegacyLayout) CalcBlockHashKey(hash []byte) []byte {
	const prefixLen = len("BH:")

	key := make([]byte, prefixLen+hex.EncodedLen(len(hash)))

	key[0], key[1], key[2] = 'B', 'H', ':'
	hex.Encode(key[prefixLen:], hash)

	return key
}

// CalcBlockMetaKey implements BlockKeyLayout.
// It returns a database key of the form "H:<height>" to store/retrieve the metadata
// of the block at the given height to/from the database.
func (*v1LegacyLayout) CalcBlockMetaKey(height int64) []byte {
	const prefixLen = len("H:")

	key := make([]byte, prefixLen, prefixLen+20)

	key[0], key[1] = 'H', ':'
	key = strconv.AppendInt(key, height, 10)

	return key
}

// CalcBlockPartKey implements BlockKeyLayout.
// It returns a database key of the form "P:<height>:<partIndex>" to store/retrieve a
// block part to/from the database.
func (*v1LegacyLayout) CalcBlockPartKey(height int64, partIndex int) []byte {
	const prefixLen = len("P:")

	// Here we have 2 ints, therefore 20+20 bytes.
	// The total size is : 2 (prefixLen) + 20 + 1 (len(":")) + 20
	key := make([]byte, prefixLen, prefixLen+20+1+20)

	key[0], key[1] = 'P', ':'
	key = strconv.AppendInt(key, height, 10)
	key = append(key, ':')
	key = strconv.AppendInt(key, int64(partIndex), 10)

	return key
}

// CalcExtCommitKey implements BlockKeyLayout.
// It returns a database key of the form "EC:<height>" to store/retrieve the
// ExtendedCommit for the given height to/from the database.
func (*v1LegacyLayout) CalcExtCommitKey(height int64) []byte {
	const prefixLen = len("EC:")

	key := make([]byte, prefixLen, prefixLen+20)

	key[0], key[1], key[2] = 'E', 'C', ':'
	key = strconv.AppendInt(key, height, 10)

	return key
}

// CalcSeenCommitKey implements BlockKeyLayout.
// It returns a database key of the form "SC:<height>" to store/retrieve a locally
// seen commit for the given height to/from the database.
func (*v1LegacyLayout) CalcSeenCommitKey(height int64) []byte {
	const prefixLen = len("SC:")

	key := make([]byte, prefixLen, prefixLen+20)

	key[0], key[1], key[2] = 'S', 'C', ':'
	key = strconv.AppendInt(key, height, 10)

	return key
}

var _ BlockKeyLayout = (*v1LegacyLayout)(nil)

type v2Layout struct{}

// key prefixes.
const (
	// prefixes are unique across all tm db's.
	prefixBlockMeta   = int64(0)
	prefixBlockPart   = int64(1)
	prefixBlockCommit = int64(2)
	prefixSeenCommit  = int64(3)
	prefixExtCommit   = int64(4)
	prefixBlockHash   = int64(5)
)

// CalcBlockCommitKey implements BlockKeyLayout.
func (*v2Layout) CalcBlockCommitKey(height int64) []byte {
	key, err := orderedcode.Append(nil, prefixBlockCommit, height)
	if err != nil {
		panic(err)
	}
	return key
}

// CalcBlockHashKey implements BlockKeyLayout.
func (*v2Layout) CalcBlockHashKey(hash []byte) []byte {
	key, err := orderedcode.Append(nil, prefixBlockHash, string(hash))
	if err != nil {
		panic(err)
	}
	return key
}

// CalcBlockMetaKey implements BlockKeyLayout.
func (*v2Layout) CalcBlockMetaKey(height int64) []byte {
	key, err := orderedcode.Append(nil, prefixBlockMeta, height)
	if err != nil {
		panic(err)
	}
	return key
}

// CalcBlockPartKey implements BlockKeyLayout.
func (*v2Layout) CalcBlockPartKey(height int64, partIndex int) []byte {
	key, err := orderedcode.Append(nil, prefixBlockPart, height, int64(partIndex))
	if err != nil {
		panic(err)
	}
	return key
}

// CalcExtCommitKey implements BlockKeyLayout.
func (*v2Layout) CalcExtCommitKey(height int64) []byte {
	key, err := orderedcode.Append(nil, prefixExtCommit, height)
	if err != nil {
		panic(err)
	}
	return key
}

// CalcSeenCommitKey implements BlockKeyLayout.
func (*v2Layout) CalcSeenCommitKey(height int64) []byte {
	key, err := orderedcode.Append(nil, prefixSeenCommit, height)
	if err != nil {
		panic(err)
	}
	return key
}

var _ BlockKeyLayout = (*v2Layout)(nil)
