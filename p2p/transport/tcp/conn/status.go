package conn

import (
	"time"

	flow "github.com/cometbft/cometbft/internal/flowrate"
)

// ConnectionStatus represents the status of a connection.
type ConnectionStatus struct {
	// Duration is the duration for which the connection has been connected.
	Duration time.Duration
	// SendMonitor shows the status of the outflow IO.
	SendMonitor flow.Status
	// RecvMonitor shows the status of the inflow IO.
	RecvMonitor flow.Status
	// Channels shows the status of each channel.
	Channels []ChannelStatus
}

// ConnectedFor returns the duration for which the connection has been connected.
func (cs ConnectionStatus) ConnectedFor() time.Duration {
	return cs.Duration
}

// ChannelStatus represents the status of a channel.
type ChannelStatus struct {
	// StreamID.
	ID byte
	// SendQueueCapacity is the capacity of the send queue.
	SendQueueCapacity int
	// SendQueueSize is the size of the send queue.
	SendQueueSize int
	// Priority is the priority of the channel.
	Priority int
	// RecentlySent is the number of messages sent recently.
	RecentlySent int64
}
