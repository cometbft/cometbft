//nolint:revive,stylecheck
package consensus

import (
	v1 "github.com/cometbft/cometbft/api/cometbft/consensus/v1"
	v2 "github.com/cometbft/cometbft/api/cometbft/consensus/v2"
)

type (
	Message                      = v2.Message
	Message_BlockPart            = v2.Message_BlockPart
	Message_HasVote              = v2.Message_HasVote
	Message_NewRoundStep         = v2.Message_NewRoundStep
	Message_NewValidBlock        = v2.Message_NewValidBlock
	Message_Proposal             = v2.Message_Proposal
	Message_ProposalPol          = v2.Message_ProposalPol
	Message_Vote                 = v2.Message_Vote
	Message_VoteSetBits          = v2.Message_VoteSetBits
	Message_VoteSetMaj23         = v2.Message_VoteSetMaj23
	Message_HasProposalBlockPart = v2.Message_HasProposalBlockPart
)

type (
	BlockPart            = v1.BlockPart
	HasProposalBlockPart = v2.HasProposalBlockPart
	HasVote              = v1.HasVote
	NewRoundStep         = v1.NewRoundStep
	NewValidBlock        = v1.NewValidBlock
	Proposal             = v1.Proposal
	ProposalPOL          = v1.ProposalPOL
	Vote                 = v2.Vote
	VoteSetBits          = v1.VoteSetBits
	VoteSetMaj23         = v1.VoteSetMaj23
)
