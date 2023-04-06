package types

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/types/v2"
	v3 "github.com/cometbft/cometbft/api/cometbft/types/v3"
)

type ABCIParams = v3.ABCIParams
type ConsensusParams = v3.ConsensusParams
type BlockParams = v2.BlockParams
type EvidenceParams = v1.EvidenceParams
type HashedParams = v1.HashedParams
type ValidatorParams = v1.ValidatorParams
type VersionParams = v2.VersionParams
