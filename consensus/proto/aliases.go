//nolint:revive,stylecheck
package proto

import (
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1beta2"
)

type (
	Message                      = v1beta2.Message
	Message_BlockPart            = v1beta2.Message_BlockPart
	Message_HasVote              = v1beta2.Message_HasVote
	Message_NewRoundStep         = v1beta2.Message_NewRoundStep
	Message_NewValidBlock        = v1beta2.Message_NewValidBlock
	Message_Proposal             = v1beta2.Message_Proposal
	Message_ProposalPol          = v1beta2.Message_ProposalPol
	Message_Vote                 = v1beta2.Message_Vote
	Message_VoteSetBits          = v1beta2.Message_VoteSetBits
	Message_VoteSetMaj23         = v1beta2.Message_VoteSetMaj23
	Message_HasProposalBlockPart = v1beta2.Message_HasProposalBlockPart
)

type (
	BlockPart            = v1beta1.BlockPart
	HasProposalBlockPart = v1beta2.HasProposalBlockPart
	HasVote              = v1beta1.HasVote
	NewRoundStep         = v1beta1.NewRoundStep
	NewValidBlock        = v1beta1.NewValidBlock
	Proposal             = v1beta1.Proposal
	ProposalPOL          = v1beta1.ProposalPOL
	Vote                 = v1beta2.Vote
	VoteSetBits          = v1beta1.VoteSetBits
	VoteSetMaj23         = v1beta1.VoteSetMaj23
)
