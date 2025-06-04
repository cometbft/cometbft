package log

import (
	"fmt"

	cmtbytes "github.com/cometbft/cometbft/v2/libs/bytes"
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

// LazyHash is a wrapper around a hashable object that defers the Hash call
// until the Stringer interface is invoked.
// This is particularly useful for avoiding calling Sprintf when debugging is
// not active.
type LazyHash struct {
	inner hashable
}

type hashable interface {
	Hash() cmtbytes.HexBytes
}

// NewLazyHash defers calling `Hash()` until the Stringer interface is invoked.
// This is particularly useful for avoiding calling Sprintf when debugging is not
// active.
func NewLazyHash(inner hashable) *LazyHash {
	return &LazyHash{inner}
}

func (l *LazyHash) String() string {
	return l.inner.Hash().String()
}
