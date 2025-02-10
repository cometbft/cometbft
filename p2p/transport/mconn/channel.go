package mconn

import (
	"bytes"
	"sync"
)

// Channel represents a channel within an MConnection
type Channel struct {
	id               byte
	sendQueue        chan []byte
	recving          []byte
	recentlySent     int64 // exponential moving average
	maxPacketMsgSize int

	mtx sync.Mutex
	buf *bytes.Buffer
}

func newChannel(id byte, maxPacketMsgSize int) *Channel {
	return &Channel{
		id:               id,
		sendQueue:        make(chan []byte, defaultSendQueueCapacity),
		maxPacketMsgSize: maxPacketMsgSize,
		buf:              new(bytes.Buffer),
	}
}
