package p2p

import "github.com/cosmos/gogoproto/proto"

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplex TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	StreamID() byte
	MessageType() proto.Message
}
