package transport

import (
	"context"
	"net"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// MetricsCollector defines the interface for collecting transport metrics
type MetricsCollector interface {
	// RecordLatency records round-trip latency
	RecordLatency(duration time.Duration)
	// RecordBandwidth records bandwidth usage
	RecordBandwidth(bytes int64, direction string)
	// RecordRetransmits records packet retransmissions (for reliable protocols)
	RecordRetransmits(count int64)
}

// Metrics contains real-time metrics about the connection
type Metrics struct {
	// Round trip time using EMA
	RTT time.Duration
	// Current bandwidth usage in bytes/sec
	Bandwidth int64
	// Protocol-specific metrics
	ProtocolMetrics any
}

// BaseTransport defines the core interface that all transport implementations must satisfy
type BaseTransport interface {
	// Protocol returns the transport protocol type
	Protocol() Protocol

	// Listen starts listening on the specified address
	Listen(laddr string) error

	// Dial establishes a connection to the specified address
	Dial(ctx context.Context, raddr string) (net.Conn, error)

	// Accept accepts an incoming connection
	Accept() (net.Conn, error)

	// GetMetrics returns current transport metrics
	GetMetrics() *Metrics

	// SetMetricsCollector sets the metrics collector
	SetMetricsCollector(collector MetricsCollector)

	// SetLogger sets the logger
	SetLogger(logger log.Logger)

	// Close closes the transport
	Close() error
}

// Options contains configuration options for transports
type Options struct {
	// Maximum packet size
	MaxPacketSize int

	// Read/Write buffer sizes
	ReadBufferSize  int
	WriteBufferSize int

	// Timeouts
	DialTimeout      time.Duration
	HandshakeTimeout time.Duration

	// Protocol-specific options stored as any
	ProtocolOptions any
}

// BaseTransportBuilder defines an interface for constructing transports
type BaseTransportBuilder interface {
	// WithOptions sets transport options
	WithOptions(opts Options) BaseTransportBuilder

	// WithLogger sets the logger
	WithLogger(logger log.Logger) BaseTransportBuilder

	// WithMetricsCollector sets the metrics collector
	WithMetricsCollector(collector MetricsCollector) BaseTransportBuilder

	// Build creates the transport instance
	Build() (BaseTransport, error)
}
