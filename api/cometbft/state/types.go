package state

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/state/v1"
	v3 "github.com/cometbft/cometbft/api/cometbft/state/v3"
)

type ABCIResponsesInfo = v3.ABCIResponsesInfo
type ConsensusParamsInfo = v3.ConsensusParamsInfo
type LegacyABCIResponses = v3.LegacyABCIResponses
type ResponseBeginBlock = v3.ResponseBeginBlock
type ResponseEndBlock = v3.ResponseEndBlock
type State = v3.State
type ValidatorsInfo = v1.ValidatorsInfo
type Version = v1.Version
