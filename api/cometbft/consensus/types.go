package consensus

import (
	"github.com/cometbft/cometbft/api/cometbft/consensus/v1"
	"github.com/cometbft/cometbft/api/cometbft/consensus/v2"
)

type Message = v2.Message
type Message_BlockPart = v2.Message_BlockPart
type Message_HasVote = v2.Message_HasVote
type Message_NewRoundStep = v2.Message_NewRoundStep
type Message_NewValidBlock = v2.Message_NewValidBlock
type Message_Proposal = v2.Message_Proposal
type Message_ProposalPol = v2.Message_ProposalPol
type Message_Vote = v2.Message_Vote
type Message_VoteSetBits = v2.Message_VoteSetBits
type Message_VoteSetMaj23 = v2.Message_VoteSetMaj23

type BlockPart = v1.BlockPart
type HasVote = v1.HasVote
type NewRoundStep = v1.NewRoundStep
type NewValidBlock = v1.NewValidBlock
type Proposal = v1.Proposal
type ProposalPOL = v1.ProposalPOL
type Vote = v2.Vote
type VoteSetBits = v1.VoteSetBits
type VoteSetMaj23 = v1.VoteSetMaj23

