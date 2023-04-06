package types

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/types/v1"
	v3 "github.com/cometbft/cometbft/api/cometbft/types/v3"
)

type DuplicateVoteEvidence = v3.DuplicateVoteEvidence
type Evidence = v3.Evidence
type EvidenceList = v3.EvidenceList
type LightClientAttackEvidence = v1.LightClientAttackEvidence

type Evidence_DuplicateVoteEvidence = v3.Evidence_DuplicateVoteEvidence
type Evidence_LightClientAttackEvidence = v3.Evidence_LightClientAttackEvidence
