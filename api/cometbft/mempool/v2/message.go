package v2

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
)

// Wrap implements the p2p Wrapper interface and wraps a mempool message.
func (m *Txs) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_Txs{Txs: m}
	return mm
}

func (m *HaveTx) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_HaveTx{HaveTx: m}
	return mm
}

func (m *ResetRoute) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_ResetRoute{ResetRoute: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_Txs:
		return m.GetTxs(), nil
	case *Message_HaveTx:
		return m.GetHaveTx(), nil
	case *Message_ResetRoute:
		return m.GetResetRoute(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
