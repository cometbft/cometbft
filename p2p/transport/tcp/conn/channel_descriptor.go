package conn

import "github.com/cosmos/gogoproto/proto"

const (
	defaultRecvBufferCapacity  = 4096
	defaultRecvMessageCapacity = 22020096 // 21MB
)

type ChannelDescriptor struct {
	ID                  byte
	Priority            int
	SendQueueCapacity   int
	RecvBufferCapacity  int
	RecvMessageCapacity int
	MessageTypeI        proto.Message
}

// StreamID returns the channel ID. Implements p2p.StreamDescriptor.
func (d ChannelDescriptor) StreamID() byte {
	return d.ID
}

// MessageType returns the message type. Implements p2p.StreamDescriptor.
func (d ChannelDescriptor) MessageType() proto.Message {
	return d.MessageTypeI
}

func (chDesc ChannelDescriptor) FillDefaults() (filled ChannelDescriptor) {
	if chDesc.SendQueueCapacity == 0 {
		chDesc.SendQueueCapacity = defaultSendQueueCapacity
	}
	if chDesc.RecvBufferCapacity == 0 {
		chDesc.RecvBufferCapacity = defaultRecvBufferCapacity
	}
	if chDesc.RecvMessageCapacity == 0 {
		chDesc.RecvMessageCapacity = defaultRecvMessageCapacity
	}
	filled = chDesc
	return filled
}
