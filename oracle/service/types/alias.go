package types

import (
	r "github.com/cometbft/cometbft/redis"
)

type (
	GenericValue = r.GenericValue
)

var (
	StringToGenericValue  = r.StringToGenericValue
	Float64ToGenericValue = r.Float64ToGenericValue
	Uint64ToGenericValue  = r.Uint64ToGenericValue
)
