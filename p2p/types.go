package p2p

import (
	"fmt"

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

	// StreamDescriptor describes a data stream. This could be a substream within
	// a multiplexed TCP connection, QUIC stream, etc.
	StreamDescriptor = transport.StreamDescriptor
)

// Envelope contains a message with sender routing info.
type Envelope struct {
	Src       Peer          // sender (empty if outbound)
	Message   proto.Message // message payload
	ChannelID byte

	// messageBytes are stored after first call to marshalMessage()
	messageBytes []byte
}

// marshalMessage marshals the Message field and stores it in a private field
// for future use.
// It returns the marshaled Message.
func (e *Envelope) marshalMessage() ([]byte, error) {
	if e.messageBytes == nil {
		msg := e.Message
		if w, ok := msg.(types.Wrapper); ok {
			msg = w.Wrap()
		}
		msgBytes, err := proto.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("proto.Marshal: %w", err)
		}
		e.messageBytes = msgBytes
	}
	return e.messageBytes, nil
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
