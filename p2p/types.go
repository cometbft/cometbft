package p2p

import (
	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/types"
)

type (
	ConnectionStatus = conn.ConnectionStatus
	// ID is the unique identifier for a peer.
	ID = nodekey.ID
	// NodeKey is the node key.
	NodeKey = nodekey.NodeKey

	// NodeInfo is the information about a peer.
	NodeInfo = ni.NodeInfo
	// NodeInfoDefault is the default implementation of NodeInfo.
	NodeInfoDefault = ni.Default
	// NodeInfoDefaultOther is the default implementation of NodeInfo for other peers.
	NodeInfoDefaultOther = ni.DefaultOther
	// ProtocolVersion is the protocol version for the software.
	ProtocolVersion = ni.ProtocolVersion
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

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}

// LoadOrGenNodeKey loads a node key from the given path or generates a new one.
func LoadOrGenNodeKey(path string) (*nodekey.NodeKey, error) {
	return nodekey.LoadOrGen(path)
}

// LoadNodeKey loads a node key from the given path.
func LoadNodeKey(path string) (*nodekey.NodeKey, error) {
	return nodekey.Load(path)
}
