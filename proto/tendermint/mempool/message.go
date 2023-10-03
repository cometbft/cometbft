package mempool

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/p2p"
)

var (
	_ p2p.Wrapper   = &Txs{}
	_ p2p.Wrapper   = &SeenTx{}
	_ p2p.Wrapper   = &WantTx{}
	_ p2p.Unwrapper = &Message{}
)

// Wrap implements the p2p Wrapper interface and wraps a mempool message.
func (m *Txs) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_Txs{Txs: m}
	return mm
}

// Wrap implements the p2p Wrapper interface and wraps a mempool seen tx message.
func (m *SeenTx) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_SeenTx{SeenTx: m}
	return mm
}

// Wrap implements the p2p Wrapper interface and wraps a mempool want tx message.
func (m *WantTx) Wrap() proto.Message {
	mm := &Message{}
	mm.Sum = &Message_WantTx{WantTx: m}
	return mm
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped mempool
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_Txs:
		return m.GetTxs(), nil

	case *Message_SeenTx:
		return m.GetSeenTx(), nil

	case *Message_WantTx:
		return m.GetWantTx(), nil

	default:
		return nil, fmt.Errorf("unknown message: %T", msg)
	}
}
