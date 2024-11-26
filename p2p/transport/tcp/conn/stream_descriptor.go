package conn

import "github.com/cosmos/gogoproto/proto"

const (
	defaultSendQueueCapacity   = 1
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096 // 21MB
)

// StreamDescriptor describes a logical stream within a session.
type StreamDescriptor struct {
	// ID is a unique identifier.
	ID byte
	// Priority is integer priority (higher means more priority).
	Priority int
	// SendQueueCapacity is the capacity of the send queue.
	// Default: 1
	SendQueueCapacity int
	// RecvBufferCapacity is the capacity of the receive buffer.
	// Default: 4KB
	RecvBufferCapacity int
	// RecvMessageCapacity is the capacity of the receive queue.
	// Default: 21MB
	RecvMessageCapacity int
	// MessageTypeI is the message type.
	MessageTypeI proto.Message
}

// StreamID returns the stream ID. Implements transport.StreamDescriptor.
func (d StreamDescriptor) StreamID() byte {
	return d.ID
}

// MessageType returns the message type. Implements transport.StreamDescriptor.
func (d StreamDescriptor) MessageType() proto.Message {
	return d.MessageTypeI
}

// FIllDefaults fills in default values for the channel descriptor.
func (d StreamDescriptor) FillDefaults() (filled StreamDescriptor) {
	if d.SendQueueCapacity == 0 {
		d.SendQueueCapacity = defaultSendQueueCapacity
	}
	if d.RecvBufferCapacity == 0 {
		d.RecvBufferCapacity = defaultRecvBufferCapacity
	}
	if d.RecvMessageCapacity == 0 {
		d.RecvMessageCapacity = defaultRecvMessageCapacity
	}
	filled = d
	return filled
}
