package store

import (
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

type v1LegacyLayout struct{}

// CalcBlockCommitKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcBlockCommitKey(height int64) []byte {
	return []byte("C:" + strconv.FormatInt(height, 10))
}

// CalcBlockHashKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcBlockHashKey(hash []byte) []byte {
	return append([]byte{'B', 'H', ':'}, hash...)
}

// CalcBlockMetaKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcBlockMetaKey(height int64) []byte {
	return []byte("H:" + strconv.FormatInt(height, 10))
}

// CalcBlockPartKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcBlockPartKey(height int64, partIndex int) []byte {
	var (
		keyPrefix = "P:"
		keySuffix = strconv.FormatInt(height, 10) + ":" + strconv.Itoa(partIndex)
		keyStr    = keyPrefix + keySuffix
	)
	return []byte(keyStr)
}

// CalcExtCommitKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcExtCommitKey(height int64) []byte {
	return []byte("EC:" + strconv.FormatInt(height, 10))
}

// CalcSeenCommitKey implements BlockKeyLayout.
func (*v1LegacyLayout) CalcSeenCommitKey(height int64) []byte {
	return []byte("SC:" + strconv.FormatInt(height, 10))
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
