package types

import (
	v1beta1 "github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	v1beta3 "github.com/cometbft/cometbft/api/cometbft/types/v1beta3"
)

type (
	BlockID           = v1beta1.BlockID
	BlockMeta         = v1beta1.BlockMeta
	Commit            = v1beta1.Commit
	CommitSig         = v1beta1.CommitSig
	ExtendedCommit    = v1beta3.ExtendedCommit
	ExtendedCommitSig = v1beta3.ExtendedCommitSig
	Header            = v1beta1.Header
	LightBlock        = v1beta1.LightBlock
	Part              = v1beta1.Part
	PartSetHeader     = v1beta1.PartSetHeader
	Proposal          = v1beta1.Proposal
	SignedHeader      = v1beta1.SignedHeader
	TxProof           = v1beta1.TxProof
	Validator         = v1beta1.Validator
	Vote              = v1beta3.Vote
)
