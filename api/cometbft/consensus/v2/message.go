package v2

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/p2p"
)

var (
	_ p2p.Wrapper = &Vote{}
	_ p2p.Wrapper = &HasProposalBlockPart{}
)

func (m *Vote) Wrap() proto.Message {
	cm := &Message{}
	cm.Sum = &Message_Vote{Vote: m}
	return cm
}

func (m *HasProposalBlockPart) Wrap() proto.Message {
	cm := &Message{}
	cm.Sum = &Message_HasProposalBlockPart{HasProposalBlockPart: m}
	return cm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped consensus
// proto message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_NewRoundStep:
		return m.GetNewRoundStep(), nil

	case *Message_NewValidBlock:
		return m.GetNewValidBlock(), nil

	case *Message_Proposal:
		return m.GetProposal(), nil

	case *Message_ProposalPol:
		return m.GetProposalPol(), nil

	case *Message_BlockPart:
		return m.GetBlockPart(), nil

	case *Message_Vote:
		return m.GetVote(), nil

	case *Message_HasVote:
		return m.GetHasVote(), nil

	case *Message_HasProposalBlockPart:
		return m.GetHasProposalBlockPart(), nil

	case *Message_VoteSetMaj23:
		return m.GetVoteSetMaj23(), nil

	case *Message_VoteSetBits:
		return m.GetVoteSetBits(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
