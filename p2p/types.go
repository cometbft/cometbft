package p2p

import (
	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/types"
)

type (
	// ConnState describes the state of a connection.
	ConnState = transport.ConnState
	// SendError is an error emitted by Peer#TrySend.
	//
	// If the send queue is full, Full() returns true.
	SendError = transport.WriteError
)

// Envelope contains a message with sender routing info.
type Envelope struct {
	Src       Peer          // sender (empty if outbound)
	Message   proto.Message // message payload
	ChannelID byte
}

var (
	_ types.Wrapper = &tmp2p.PexRequest{}
	_ types.Wrapper = &tmp2p.PexAddrs{}
)
