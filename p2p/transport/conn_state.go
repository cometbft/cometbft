package transport

import "time"

// ConnState describes the state of a connection.
type ConnState struct {
	// ConnectedFor is the duration for which the connection has been connected.
	ConnectedFor time.Duration `json:"connected_for"`
	// StreamStates describes the state of streams.
	StreamStates map[byte]StreamState `json:"stream_states"`
	// SendRateLimiterDelay is the delay imposed by the send rate limiter.
	//
	// Only applies to TCP.
	SendRateLimiterDelay time.Duration `json:"send_rate_limiter_delay"`
	// RecvRateLimiterDelay is the delay imposed by the receive rate limiter.
	//
	// Only applies to TCP.
	RecvRateLimiterDelay time.Duration `json:"recv_rate_limiter_delay"`
}

// StreamState is the state of a stream.
type StreamState struct {
	// SendQueueSize is the size of the send queue.
	SendQueueSize int `json:"send_queue_size"`
	// SendQueueCapacity is the capacity of the send queue.
	SendQueueCapacity int `json:"send_queue_capacity"`
}
