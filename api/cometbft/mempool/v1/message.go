package v1

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

func (m *MempoolIsFull) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_MempoolIsFull{MempoolIsFull: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_Txs:
		return m.GetTxs(), nil
	case *Message_MempoolIsFull:
		return m.GetMempoolIsFull(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
