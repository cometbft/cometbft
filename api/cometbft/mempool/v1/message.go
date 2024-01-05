package v1

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
)

type Unwrapper interface {
	Unwrap() (proto.Message, error)
}

func (m *Message_Txs) Unwrap() (proto.Message, error) {
	return m.Txs, nil
}

func (m *Message_SeenTx) Unwrap() (proto.Message, error) {
	return m.SeenTx, nil
}

func (m *Message_WantTx) Unwrap() (proto.Message, error) {
	return m.WantTx, nil
}

func (m *Message) Unwrap() (proto.Message, error) {
	if unwrapper, ok := m.Sum.(Unwrapper); ok {
		return unwrapper.Unwrap()
	}
	return nil, fmt.Errorf("unknown message: %T", m.Sum)
}
