package types

import (
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta2"
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta3"
)

type (
	ABCIParams      = v1beta3.ABCIParams
	ConsensusParams = v1beta3.ConsensusParams
	BlockParams     = v1beta2.BlockParams
	EvidenceParams  = v1beta1.EvidenceParams
	HashedParams    = v1beta1.HashedParams
	ValidatorParams = v1beta1.ValidatorParams
	VersionParams   = v1beta1.VersionParams
)
