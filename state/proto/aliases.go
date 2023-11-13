package proto

import (
	"github.com/cometbft/cometbft/api/cometbft/state/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/state/v1beta3"
	"github.com/cometbft/cometbft/api/cometbft/state/v1beta4"
)

type (
	ABCIResponsesInfo   = v1beta4.ABCIResponsesInfo
	ConsensusParamsInfo = v1beta3.ConsensusParamsInfo
	LegacyABCIResponses = v1beta3.LegacyABCIResponses
	ResponseBeginBlock  = v1beta3.ResponseBeginBlock
	ResponseEndBlock    = v1beta3.ResponseEndBlock
	State               = v1beta3.State
	ValidatorsInfo      = v1beta1.ValidatorsInfo
	Version             = v1beta1.Version
)
