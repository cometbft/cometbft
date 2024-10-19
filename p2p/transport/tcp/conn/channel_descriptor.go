package conn

import "github.com/cosmos/gogoproto/proto"

const (
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096 // 21MB
)

// ChannelDescriptor describes a channel.
type ChannelDescriptor struct {
	// ID is a unique identifier for the channel.
	ID byte
	// Priority is the priority of the channel.
	Priority int
	// SendQueueCapacity is the capacity of the send queue.
	SendQueueCapacity int
	// RecvBufferCapacity is the capacity of the receive buffer.
	RecvBufferCapacity int
	// RecvMessageCapacity is the capacity of the receive queue.
	RecvMessageCapacity int
	// MessageTypeI is the message type.
	MessageTypeI proto.Message
}

// StreamID returns the channel ID. Implements p2p.StreamDescriptor.
func (d ChannelDescriptor) StreamID() byte {
	return d.ID
}

// MessageType returns the message type. Implements p2p.StreamDescriptor.
func (d ChannelDescriptor) MessageType() proto.Message {
	return d.MessageTypeI
}

// FIllDefaults fills in default values for the channel descriptor.
func (d ChannelDescriptor) FillDefaults() (filled ChannelDescriptor) {
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
