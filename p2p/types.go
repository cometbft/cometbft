package p2p

import (
	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/types"
)

type (
	// ConnState describes the state of a connection.
	ConnState = transport.ConnState
	// ID is the unique identifier for a peer.
	ID = nodekey.ID

	// NodeInfo is the information about a peer.
	NodeInfo = ni.NodeInfo
	// NodeInfoDefault is the default implementation of NodeInfo.
	NodeInfoDefault = ni.Default
	// NodeInfoDefaultOther is the default implementation of NodeInfo for other peers.
	NodeInfoDefaultOther = ni.DefaultOther
	// ProtocolVersion is the protocol version for the software.
	ProtocolVersion = ni.ProtocolVersion
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
