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
