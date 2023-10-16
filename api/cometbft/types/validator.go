package types

import (
	v1beta1 "github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
)

type (
	BlockIDFlag     = v1beta1.BlockIDFlag
	SimpleValidator = v1beta1.SimpleValidator
	ValidatorSet    = v1beta1.ValidatorSet
)

const (
	BlockIDFlagUnknown BlockIDFlag = v1beta1.BlockIDFlagUnknown
	BlockIDFlagAbsent  BlockIDFlag = v1beta1.BlockIDFlagAbsent
	BlockIDFlagCommit  BlockIDFlag = v1beta1.BlockIDFlagCommit
	BlockIDFlagNil     BlockIDFlag = v1beta1.BlockIDFlagNil
)
