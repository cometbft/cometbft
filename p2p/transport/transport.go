package transport

import (
	"github.com/cosmos/gogoproto/proto"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport/types"
)

// Protocol is a string type for transport protocols
type Protocol string

// Define protocol constants
const (
	TCPProtocol  Protocol = "tcp"
	QUICProtocol Protocol = "quic"
	KCPProtocol  Protocol = "kcp"
)

// Transport defines the interface for network transports
type Transport interface {
	// NetAddr returns the network address of the local node.
	NetAddr() na.NetAddr

	// Accept waits for and returns the next connection to the local node.
	Accept() (types.Conn, *na.NetAddr, error)

	// Dial dials the given address and returns a connection.
	Dial(addr na.NetAddr) (types.Conn, error)

	// Listen starts listening on the specified address
	Listen(addr na.NetAddr) error

	// Close closes the transport
	Close() error

	// Protocol returns the transport protocol type
	Protocol() Protocol
}

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}
