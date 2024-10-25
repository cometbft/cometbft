package log

import (
	"fmt"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
)

type LazySprintf struct {
	format string
	args   []any
}

// NewLazySprintf defers fmt.Sprintf until the Stringer interface is invoked.
// This is particularly useful for avoiding calling Sprintf when debugging is not
// active.
func NewLazySprintf(format string, args ...any) *LazySprintf {
	return &LazySprintf{format, args}
}

func (l *LazySprintf) String() string {
	return fmt.Sprintf(l.format, l.args...)
}

type lazyHash struct {
	inner hashable
}

type hashable interface {
	Hash() cmtbytes.HexBytes
}

// NewLazyHash defers Hash until the Stringer interface is invoked. This is
// particularly useful for avoiding calling Sprintf when debugging is not
// active.
func NewLazyHash(inner hashable) *lazyHash {
	return &lazyHash{inner}
}

func (l *lazyHash) String() string {
	return l.inner.Hash().String()
}
