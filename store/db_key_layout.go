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
	const (
		prefix    = "C:"
		prefixLen = len(prefix)
	)
	key := make([]byte, 0, prefixLen+20)

	key = append(key, prefix...)
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
	const (
		prefix    = "H:"
		prefixLen = len(prefix)
	)
	key := make([]byte, 0, prefixLen+20)

	key = append(key, prefix...)
	key = strconv.AppendInt(key, height, 10)

	return key
}

// CalcBlockPartKey implements BlockKeyLayout.
// It returns a database key of the form "P:<height>:<partIndex>" to store/retrieve a
// block part to/from the database.
func (*v1LegacyLayout) CalcBlockPartKey(height int64, partIndex int) []byte {
	const (
		prefix    = "P:"
		prefixLen = len(prefix)
	)

	// Here we have 2 ints, therefore 20+1 bytes.
	// 1 byte is for the partIndex should be sufficient. We have observed that most
	// chains have only a few parts per block. If things change, we can increment
	// this number. The theoretical max is 4 and comes from the following
	// calculation:
	// - the max configurable block size is 100MB (see types/params.go)
	// - a block part is 65KB (see types/params.go)
	// - the max number of parts that a block can be split into is therefore
	//   (max block size / block part size) + 1 = (100MB/65KB) + 1 = 1601
	// - the string representation of 1601 consists of 4 digits, therefore 4 bytes.
	//
	// The total size is : prefixLen + 20 + 1 (len(":")) + 1.
	key := make([]byte, 0, prefixLen+20+1+1)

	key = append(key, prefix...)
	key = strconv.AppendInt(key, height, 10)
	key = append(key, ':')
	key = strconv.AppendInt(key, int64(partIndex), 10)

	return key
}

// CalcExtCommitKey implements BlockKeyLayout.
// It returns a database key of the form "EC:<height>" to store/retrieve the
// ExtendedCommit for the given height to/from the database.
func (*v1LegacyLayout) CalcExtCommitKey(height int64) []byte {
	const (
		prefix    = "EC:"
		prefixLen = len(prefix)
	)
	key := make([]byte, 0, prefixLen+20)

	key = append(key, prefix...)
	key = strconv.AppendInt(key, height, 10)

	return key
}

// CalcSeenCommitKey implements BlockKeyLayout.
// It returns a database key of the form "SC:<height>" to store/retrieve a locally
// seen commit for the given height to/from the database.
func (*v1LegacyLayout) CalcSeenCommitKey(height int64) []byte {
	const (
		prefix    = "SC:"
		prefixLen = len(prefix)
	)
	key := make([]byte, 0, prefixLen+20)

	key = append(key, prefix...)
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
