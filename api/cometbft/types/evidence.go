//nolint:revive,stylecheck
package types

import (
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta3"
)

type (
	DuplicateVoteEvidence     = v1beta3.DuplicateVoteEvidence
	Evidence                  = v1beta3.Evidence
	EvidenceList              = v1beta3.EvidenceList
	LightClientAttackEvidence = v1beta1.LightClientAttackEvidence
)

type (
	Evidence_DuplicateVoteEvidence     = v1beta3.Evidence_DuplicateVoteEvidence
	Evidence_LightClientAttackEvidence = v1beta3.Evidence_LightClientAttackEvidence
)
