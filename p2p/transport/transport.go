package transport

import (
	"github.com/cosmos/gogoproto/proto"

	na "github.com/cometbft/cometbft/p2p/netaddr"
)

// Transport connects the local node to the rest of the network.
type Transport interface {
	// NetAddr returns the network address of the local node.
	NetAddr() na.NetAddr

	// Accept waits for and returns the next connection to the local node.
	Accept() (Conn, *na.NetAddr, error)

	// Dial dials the given address and returns a connection.
	Dial(addr na.NetAddr) (Conn, error)

	// Listen starts listening on the specified address
	Listen(addr na.NetAddr) error
}

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}
