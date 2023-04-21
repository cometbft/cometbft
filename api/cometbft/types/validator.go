package types

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"
)

type BlockIDFlag = v1.BlockIDFlag
type SimpleValidator = v1.SimpleValidator
type ValidatorSet = v1.ValidatorSet

const (
	BlockIDFlagUnknown BlockIDFlag = v1.BlockIDFlagUnknown
	BlockIDFlagAbsent  BlockIDFlag = v1.BlockIDFlagAbsent
	BlockIDFlagCommit  BlockIDFlag = v1.BlockIDFlagCommit
	BlockIDFlagNil     BlockIDFlag = v1.BlockIDFlagNil
)
