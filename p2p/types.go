package p2p

import (
	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
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

// LoadOrGenNodeKey loads a node key from the given path or generates a new one.
func LoadOrGenNodeKey(path string) (*nodekey.NodeKey, error) {
	return nodekey.LoadOrGen(path)
}

// LoadNodeKey loads a node key from the given path.
func LoadNodeKey(path string) (*nodekey.NodeKey, error) {
	return nodekey.Load(path)
}
