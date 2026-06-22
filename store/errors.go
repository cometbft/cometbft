package store

import "errors"

var (
	// ErrPruneHeightMustBePositive is returned when pruning to a non-positive height.
	ErrPruneHeightMustBePositive = errors.New("height must be greater than 0")

	// ErrPruneHeightBeyondLatest is returned when pruning beyond the latest stored height.
	ErrPruneHeightBeyondLatest = errors.New("cannot prune beyond the latest height")

	// ErrPruneHeightBelowBase is returned when pruning below the current base height.
	ErrPruneHeightBelowBase = errors.New("cannot prune below base height")
)
