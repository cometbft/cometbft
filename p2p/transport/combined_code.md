# Combined Code Files

This document contains the consolidated code from multiple files in the current folder.
It is formatted in Markdown with syntax highlighting to aid analysis and understanding.

---

## Table of Contents

- [`base_transport.go`](#file-1)
- [`combined_code.md`](#file-2)
- [`conn.go`](#file-3)
- [`conn_state.go`](#file-4)
- [`quic/errors.go`](#file-5)
- [`quic/quic.go`](#file-6)
- [`quic/quic_test.go`](#file-7)
- [`quic/stream.go`](#file-8)
- [`quic/wrapper.go`](#file-9)
- [`tcp/conn/connection.go`](#file-10)
- [`tcp/conn/connection_test.go`](#file-11)
- [`tcp/conn/errors.go`](#file-12)
- [`tcp/conn/evil_secret_connection_test.go`](#file-13)
- [`tcp/conn/secret_connection.go`](#file-14)
- [`tcp/conn/secret_connection_test.go`](#file-15)
- [`tcp/conn/stream.go`](#file-16)
- [`tcp/conn/stream_descriptor.go`](#file-17)
- [`tcp/conn_set.go`](#file-18)
- [`tcp/errors.go`](#file-19)
- [`tcp/tcp.go`](#file-20)
- [`tcp/tcp_test.go`](#file-21)
- [`transport.go`](#file-22)

---

<a name="file-1"></a>

### File: `base_transport.go`

*Modified:* 2025-02-08 17:58:04 • *Size:* 3 KB

```go
package transport

import (
	"context"
	"net"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// Protocol represents different transport protocols
type Protocol string

const (
	ProtocolTCP  Protocol = "tcp"
	ProtocolQUIC Protocol = "quic"
	ProtocolKCP  Protocol = "kcp"
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

```

---

<a name="file-2"></a>

### File: `combined_code.md`

*Modified:* 2025-02-08 18:00:57 • *Size:* 224 KB

```markdown
# Combined Code Files

This document contains the consolidated code from multiple files in the current folder.
It is formatted in Markdown with syntax highlighting to aid analysis and understanding.

---

## Table of Contents

- [`base_transport.go`](#file-1)
- [`combined_code.md`](#file-2)
- [`conn.go`](#file-3)
- [`conn_state.go`](#file-4)
- [`quic/errors.go`](#file-5)
- [`quic/quic.go`](#file-6)
- [`quic/quic_test.go`](#file-7)
- [`quic/wrapper.go`](#file-8)
- [`tcp/conn/connection.go`](#file-9)
- [`tcp/conn/connection_test.go`](#file-10)
- [`tcp/conn/errors.go`](#file-11)
- [`tcp/conn/evil_secret_connection_test.go`](#file-12)
- [`tcp/conn/secret_connection.go`](#file-13)
- [`tcp/conn/secret_connection_test.go`](#file-14)
- [`tcp/conn/stream.go`](#file-15)
- [`tcp/conn/stream_descriptor.go`](#file-16)
- [`tcp/conn_set.go`](#file-17)
- [`tcp/errors.go`](#file-18)
- [`tcp/tcp.go`](#file-19)
- [`tcp/tcp_test.go`](#file-20)
- [`transport.go`](#file-21)

---

<a name="file-1"></a>

### File: `base_transport.go`

*Modified:* 2025-02-08 17:58:04 • *Size:* 3 KB

```go
package transport

import (
	"context"
	"net"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// Protocol represents different transport protocols
type Protocol string

const (
	ProtocolTCP  Protocol = "tcp"
	ProtocolQUIC Protocol = "quic"
	ProtocolKCP  Protocol = "kcp"
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

```

---

<a name="file-2"></a>

### File: `combined_code.md`

*Modified:* 2025-02-08 17:37:18 • *Size:* 106 KB

```markdown
# Combined Code Files

This document contains the consolidated code from multiple files in the current folder.
It is formatted in Markdown with syntax highlighting to aid analysis and understanding.

---

## Table of Contents

- [`conn.go`](#file-1)
- [`conn_state.go`](#file-2)
- [`tcp/conn/connection.go`](#file-3)
- [`tcp/conn/connection_test.go`](#file-4)
- [`tcp/conn/errors.go`](#file-5)
- [`tcp/conn/evil_secret_connection_test.go`](#file-6)
- [`tcp/conn/secret_connection.go`](#file-7)
- [`tcp/conn/secret_connection_test.go`](#file-8)
- [`tcp/conn/stream.go`](#file-9)
- [`tcp/conn/stream_descriptor.go`](#file-10)
- [`tcp/conn_set.go`](#file-11)
- [`tcp/errors.go`](#file-12)
- [`tcp/tcp.go`](#file-13)
- [`tcp/tcp_test.go`](#file-14)
- [`transport.go`](#file-15)

---

<a name="file-1"></a>

### File: `conn.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 3 KB

```go
package transport

import (
	"io"
	"net"
	"time"
)

// Conn is a multiplexed connection that can send and receive data
// on multiple streams.
type Conn interface {
	// OpenStream opens a new stream on the connection with an optional
	// description. If you're using tcp.MultiplexTransport, all streams must be
	// registered in advance.
	OpenStream(streamID byte, desc any) (Stream, error)

	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr

	// Close closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	Close(reason string) error

	// FlushAndClose flushes all the pending bytes and closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	FlushAndClose(reason string) error

	// ConnState returns basic details about the connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnState() ConnState

	// ErrorCh returns a channel that will receive errors from the connection.
	ErrorCh() <-chan error

	// HandshakeStream returns the stream to be used for the handshake.
	HandshakeStream() HandshakeStream
}

// Stream is the interface implemented by QUIC streams or multiplexed TCP connection.
type Stream interface {
	SendStream
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// Write writes data to the stream.
	// It blocks until data is sent or the stream is closed.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// TryWrite attempts to write data to the stream.
	// If the send queue is full, the error satisfies the WriteError interface, and Full() will be true.
	TryWrite(b []byte) (n int, err error)
}

// WriteError is returned by TryWrite when the send queue is full.
type WriteError interface {
	error
	Full() bool // Is the error due to the send queue being full?
}

// HandshakeStream is a stream that is used for the handshake.
type HandshakeStream interface {
	SetDeadline(t time.Time) error
	io.ReadWriter
}

```

---

<a name="file-2"></a>

### File: `conn_state.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 1 KB

```go
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

```

---

<a name="file-3"></a>

### File: `tcp/conn/connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 26 KB

```go
package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	flow "github.com/cometbft/cometbft/internal/flowrate"
	"github.com/cometbft/cometbft/internal/timer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/transport"
)

const (
	defaultMaxPacketMsgPayloadSize = 1024

	numBatchPacketMsgs = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	updateStats        = 2 * time.Second

	// some of these defaults are written in the user config
	// flushThrottle, sendRate, recvRate
	// TODO: remove values present in config.
	defaultFlushThrottle = 10 * time.Millisecond

	defaultSendRate     = int64(512000) // 500KB/s
	defaultRecvRate     = int64(512000) // 500KB/s
	defaultPingInterval = 60 * time.Second
	defaultPongTimeout  = 45 * time.Second
)

// OnReceiveFn is a callback func, which is called by the MConnection when a
// new message is received.
type OnReceiveFn = func(byte, []byte)

// MConnection is a multiplexed connection.
//
// __multiplex__ *noun* a system or signal involving simultaneous transmission
// of several messages along a single channel of communication.
//
// Each connection handles message transmission on multiple abstract
// communication streams. Each stream has a globally unique byte id. The byte
// id and the relative priorities of each stream are configured upon
// initialization of the connection.
//
// To open a stream, call OpenStream with the stream id. Remember that the
// stream id must be globally unique.
//
// Connection errors are communicated through the ErrorCh channel.
//
// Connection can be closed either by calling Close or FlushAndClose. If you
// need to flush all pending messages before closing the connection, call
// FlushAndClose. Otherwise, call Close.
type MConnection struct {
	service.BaseService

	conn          net.Conn
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flow.Monitor
	recvMonitor   *flow.Monitor
	send          chan struct{}
	pong          chan struct{}
	errorCh       chan error
	config        MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	quitSendRoutine chan struct{}
	doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	flushTimer *timer.ThrottleTimer // flush writes as necessary but throttled.
	pingTimer  *time.Ticker         // send pings periodically

	// close conn if pong is not received in pongTimeout
	pongTimer     *time.Timer
	pongTimeoutCh chan bool // true - timeout, false - peer sent pong

	chStatsTimer *time.Ticker // update channel stats periodically

	created time.Time // time of creation

	_maxPacketMsgSize int

	// streamID -> channel
	channelsIdx map[byte]*stream

	// A map which stores the received messages. Used in tests.
	msgsByStreamIDMap map[byte]chan []byte

	onReceiveFn OnReceiveFn
}

var _ transport.Conn = (*MConnection)(nil)

// MConnConfig is a MConnection configuration.
type MConnConfig struct {
	SendRate int64 `mapstructure:"send_rate"`
	RecvRate int64 `mapstructure:"recv_rate"`

	// Maximum payload size
	MaxPacketMsgPayloadSize int `mapstructure:"max_packet_msg_payload_size"`

	// Interval to flush writes (throttled)
	FlushThrottle time.Duration `mapstructure:"flush_throttle"`

	// Interval to send pings
	PingInterval time.Duration `mapstructure:"ping_interval"`

	// Maximum wait time for pongs
	PongTimeout time.Duration `mapstructure:"pong_timeout"`

	// Fuzz connection
	TestFuzz       bool                   `mapstructure:"test_fuzz"`
	TestFuzzConfig *config.FuzzConnConfig `mapstructure:"test_fuzz_config"`
}

// DefaultMConnConfig returns the default config.
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:                defaultSendRate,
		RecvRate:                defaultRecvRate,
		MaxPacketMsgPayloadSize: defaultMaxPacketMsgPayloadSize,
		FlushThrottle:           defaultFlushThrottle,
		PingInterval:            defaultPingInterval,
		PongTimeout:             defaultPongTimeout,
	}
}

// NewMConnection wraps net.Conn and creates multiplex connection.
func NewMConnection(conn net.Conn, config MConnConfig) *MConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}

	mconn := &MConnection{
		conn:              conn,
		bufConnReader:     bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter:     bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:       flow.New(0, 0),
		recvMonitor:       flow.New(0, 0),
		send:              make(chan struct{}, 1),
		pong:              make(chan struct{}, 1),
		errorCh:           make(chan error, 1),
		config:            config,
		created:           time.Now(),
		channelsIdx:       make(map[byte]*stream),
		msgsByStreamIDMap: make(map[byte]chan []byte),
	}

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

// OnReceive sets the callback function to be executed each time we read a message.
func (c *MConnection) OnReceive(fn OnReceiveFn) {
	c.onReceiveFn = fn
}

func (c *MConnection) SetLogger(l log.Logger) {
	c.BaseService.SetLogger(l)
}

// OnStart implements BaseService.
func (c *MConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

func (c *MConnection) Conn() net.Conn {
	return c.conn
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
func (c *MConnection) stopServices() (alreadyStopped bool) {
	select {
	case <-c.quitSendRoutine:
		// already quit
		return true
	default:
	}

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	close(c.quitSendRoutine)
	return false
}

// ErrorCh returns a channel that will receive errors from the connection.
func (c *MConnection) ErrorCh() <-chan error {
	return c.errorCh
}

func (c *MConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *MConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// OpenStream opens a new stream on the connection. Remember that the
// stream id must be globally unique.
//
// Panics if the connection is already running (i.e., all streams
// must be registered in advance).
func (c *MConnection) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	if c.IsRunning() {
		panic("MConnection is already running. Please register all streams in advance")
	}

	c.Logger.Debug("Opening stream", "streamID", streamID, "desc", desc)

	if _, ok := c.channelsIdx[streamID]; ok {
		return nil, fmt.Errorf("stream %X already exists", streamID)
	}

	d := StreamDescriptor{
		ID:       streamID,
		Priority: 1,
	}
	if desc, ok := desc.(StreamDescriptor); ok {
		d = desc
	}
	c.channelsIdx[streamID] = newChannel(c, d)
	c.channelsIdx[streamID].SetLogger(c.Logger.With("streamID", streamID))
	// Allocate some buffer, otherwise CI tests will fail.
	c.msgsByStreamIDMap[streamID] = make(chan []byte, 5)

	return &MConnectionStream{conn: c, streamID: streamID}, nil
}

// HandshakeStream returns the underlying net.Conn connection.
func (c *MConnection) HandshakeStream() transport.HandshakeStream {
	return c.conn
}

// Close closes the connection. It flushes all pending writes before closing.
func (c *MConnection) Close(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	return c.conn.Close()
}

func (c *MConnection) FlushAndClose(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	// flush all pending writes
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine
		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		w := protoio.NewDelimitedWriter(c.bufConnWriter)
		eof := c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		for !eof {
			eof = c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		}
		_ = c.flush()
	}

	return c.conn.Close()
}

func (c *MConnection) ConnState() (state transport.ConnState) {
	state.ConnectedFor = time.Since(c.created)
	state.SendRateLimiterDelay = c.sendMonitor.Status().SleepTime
	state.RecvRateLimiterDelay = c.recvMonitor.Status().SleepTime
	state.StreamStates = make(map[byte]transport.StreamState)

	for streamID, channel := range c.channelsIdx {
		state.StreamStates[streamID] = transport.StreamState{
			SendQueueSize:     channel.loadSendQueueSize(),
			SendQueueCapacity: cap(channel.sendQueue),
		}
	}

	return state
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() error {
	return c.bufConnWriter.Flush()
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.Close(fmt.Sprintf("recovered from panic: %v", r))
	}
}

// thread-safe.
func (c *MConnection) sendBytes(chID byte, msgBytes []byte, blocking bool) error {
	if !c.IsRunning() {
		return nil
	}

	// Uncomment in you need to see raw bytes.
	// c.Logger.Debug("Send",
	// 	"streamID", chID,
	// 	"msgBytes", log.NewLazySprintf("%X", msgBytes),
	// 	"timeout", timeout)

	channel, ok := c.channelsIdx[chID]
	if !ok {
		panic(fmt.Sprintf("Unknown channel %X. Forgot to register?", chID))
	}
	if err := channel.sendBytes(msgBytes, blocking); err != nil {
		return err
	}

	// Wake up sendRoutine if necessary
	select {
	case c.send <- struct{}{}:
	default:
	}
	return nil
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
//
// thread-safe.
func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Unknown channel %X", chID))
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

	protoWriter := protoio.NewDelimitedWriter(c.bufConnWriter)

FOR_LOOP:
	for {
		var _n int
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.chStatsTimer.C:
			for _, channel := range c.channelsIdx {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPing", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
				select {
				case c.pongTimeoutCh <- true:
				default:
				}
			})
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPong", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs(protoWriter)
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomePacketMsgs(w protoio.Writer) bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, c.config.SendRate, true)

	// Now send some PacketMsgs.
	return c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendBatchPacketMsgs(w protoio.Writer, batchSize int) bool {
	// Send a batch of PacketMsgs.
	totalBytesWritten := 0
	defer func() {
		if totalBytesWritten > 0 {
			c.sendMonitor.Update(totalBytesWritten)
		}
	}()
	for i := 0; i < batchSize; i++ {
		channel := c.selectChannel()
		// nothing to send across any channel.
		if channel == nil {
			return true
		}
		bytesWritten, err := c.sendPacketMsgOnChannel(w, channel)
		if err {
			return true
		}
		totalBytesWritten += bytesWritten
	}
	return false
}

// selects a channel to gossip our next message on.
// TODO: Make "batchChannelToGossipOn", so we can do our proto marshaling overheads in parallel,
// and we can avoid re-checking for `isSendPending`.
// We can easily mock the recentlySent differences for the batch choosing.
func (c *MConnection) selectChannel() *stream {
	// Choose a channel to create a PacketMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *stream
	for _, channel := range c.channelsIdx {
		// If nothing to send, skip this channel
		// TODO: Skip continually looking for isSendPending on channels we've already skipped in this batch-send.
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		// TODO: RecentlySent right now is bytes. This should be refactored to num messages to fix
		// gossip prioritization bugs.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}
	return leastChannel
}

// returns (num_bytes_written, error_occurred).
func (c *MConnection) sendPacketMsgOnChannel(w protoio.Writer, sendChannel *stream) (int, bool) {
	// Make & send a PacketMsg from this channel
	n, err := sendChannel.writePacketMsgTo(w)
	if err != nil {
		c.Logger.Error("Failed to write PacketMsg", "err", err)
		c.Close(err.Error())
		return n, true
	}
	// TODO: Change this to only add flush signals at the start and end of the batch.
	c.flushTimer.Set()
	return n, false
}

// recvRoutine reads PacketMsgs and reconstructs the message using the
// channels' "recving" buffer. After a whole message has been assembled, it's
// pushed to an internal queue, which is accessible via Read. Blocks depending
// on how the connection is throttled. Otherwise, it never blocks.
func (c *MConnection) recvRoutine() {
	defer c._recover()

	protoReader := protoio.NewDelimitedReader(c.bufConnReader, c._maxPacketMsgSize)

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(cmtmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet tmp2p.Packet

		_n, err := protoReader.ReadMsg(&packet)
		c.recvMonitor.Update(_n)
		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is expected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if errors.Is(err, io.EOF) {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)")
				} else {
					c.Logger.Debug("Connection failed @ recvRoutine (reading byte)", "err", err)
				}
				c.Close(err.Error())
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.Sum.(type) {
		case *tmp2p.Packet_PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketPong:
			c.Logger.Debug("Receive Pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketMsg:
			channelID := byte(pkt.PacketMsg.ChannelID)
			channel, ok := c.channelsIdx[channelID]
			if !ok || pkt.PacketMsg.ChannelID < 0 || pkt.PacketMsg.ChannelID > math.MaxUint8 {
				err := fmt.Errorf("unknown channel %X", pkt.PacketMsg.ChannelID)
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}

			msgBytes, err := channel.recvPacketMsg(*pkt.PacketMsg)
			if err != nil {
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}
			if msgBytes != nil {
				// Uncomment in you need to see raw bytes.
				// c.Logger.Debug("Received", "streamID", channelID, "msgBytes", log.NewLazySprintf("%X", msgBytes))
				if c.onReceiveFn != nil {
					c.onReceiveFn(channelID, msgBytes)
				} else {
					bz := make([]byte, len(msgBytes))
					copy(bz, msgBytes)
					c.msgsByStreamIDMap[channelID] <- bz
				}
			}
		default:
			err := fmt.Errorf("unknown message type %v", reflect.TypeOf(packet))
			c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	close(c.pong)
}

// Used in tests.
func (c *MConnection) readBytes(streamID byte, b []byte, timeout time.Duration) (n int, err error) {
	select {
	case msgBytes := <-c.msgsByStreamIDMap[streamID]:
		n = copy(b, msgBytes)
		if n < len(msgBytes) {
			err = errors.New("short buffer")
			return 0, err
		}
		return n, nil
	case <-time.After(timeout):
		return 0, errors.New("read timeout")
	}
}

// not goroutine-safe.
func (c *MConnection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize returns a maximum size of PacketMsg.
func (c *MConnection) maxPacketMsgSize() int {
	bz, err := proto.Marshal(mustWrapPacket(&tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, c.config.MaxPacketMsgPayloadSize),
	}))
	if err != nil {
		panic(err)
	}
	return len(bz)
}

// -----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type stream struct {
	conn          *MConnection
	desc          StreamDescriptor
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	recentlySent  int64 // exponential moving average

	nextPacketMsg           *tmp2p.PacketMsg
	nextP2pWrapperPacketMsg *tmp2p.Packet_PacketMsg
	nextPacket              *tmp2p.Packet

	maxPacketMsgPayloadSize int

	Logger log.Logger
}

func newChannel(conn *MConnection, desc StreamDescriptor) *stream {
	desc = desc.FillDefaults()
	if desc.Priority <= 0 {
		panic("Channel default priority must be a positive integer")
	}
	return &stream{
		conn:                    conn,
		desc:                    desc,
		sendQueue:               make(chan []byte, desc.SendQueueCapacity),
		recving:                 make([]byte, 0, desc.RecvBufferCapacity),
		nextPacketMsg:           &tmp2p.PacketMsg{ChannelID: int32(desc.ID)},
		nextP2pWrapperPacketMsg: &tmp2p.Packet_PacketMsg{},
		nextPacket:              &tmp2p.Packet{},
		maxPacketMsgPayloadSize: conn.config.MaxPacketMsgPayloadSize,
	}
}

func (ch *stream) SetLogger(l log.Logger) {
	ch.Logger = l
}

// Queues message to send to this channel. Blocks if blocking is true.
// thread-safe.
func (ch *stream) sendBytes(bytes []byte, blocking bool) error {
	if blocking {
		select {
		case ch.sendQueue <- bytes:
			atomic.AddInt32(&ch.sendQueueSize, 1)
			return nil
		case <-ch.conn.Quit():
			return nil
		}
	}

	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return nil
	default:
		return ErrWriteQueueFull{}
	case <-ch.conn.Quit():
		return nil
	}
}

// Goroutine-safe.
func (ch *stream) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *stream) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any PacketMsgs are pending to be sent.
// Call before calling updateNextPacket
// Goroutine-safe.
func (ch *stream) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Updates the nextPacket proto message for us to send.
// Not goroutine-safe.
func (ch *stream) updateNextPacket() {
	maxSize := ch.maxPacketMsgPayloadSize
	if len(ch.sending) <= maxSize {
		ch.nextPacketMsg.Data = ch.sending
		ch.nextPacketMsg.EOF = true
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		ch.nextPacketMsg.Data = ch.sending[:maxSize]
		ch.nextPacketMsg.EOF = false
		ch.sending = ch.sending[maxSize:]
	}

	ch.nextP2pWrapperPacketMsg.PacketMsg = ch.nextPacketMsg
	ch.nextPacket.Sum = ch.nextP2pWrapperPacketMsg
}

// Writes next PacketMsg to w and updates c.recentlySent.
// Not goroutine-safe.
func (ch *stream) writePacketMsgTo(w protoio.Writer) (n int, err error) {
	ch.updateNextPacket()
	n, err = w.WriteMsg(ch.nextPacket)
	if err != nil {
		err = ErrPacketWrite{Source: err}
	}

	atomic.AddInt64(&ch.recentlySent, int64(n))
	return n, err
}

// Handles incoming PacketMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe.
func (ch *stream) recvPacketMsg(packet tmp2p.PacketMsg) ([]byte, error) {
	recvCap, recvReceived := ch.desc.RecvMessageCapacity, len(ch.recving)+len(packet.Data)
	if recvCap < recvReceived {
		return nil, ErrPacketTooBig{Max: recvCap, Received: recvReceived}
	}

	ch.recving = append(ch.recving, packet.Data...)
	if packet.EOF {
		msgBytes := ch.recving

		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// thread-safe.
func (ch *stream) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

// ----------------------------------------
// Packet

// mustWrapPacket takes a packet kind (oneof) and wraps it in a tmp2p.Packet message.
func mustWrapPacket(pb proto.Message) *tmp2p.Packet {
	msg := &tmp2p.Packet{}
	mustWrapPacketInto(pb, msg)
	return msg
}

func mustWrapPacketInto(pb proto.Message, dst *tmp2p.Packet) {
	switch pb := pb.(type) {
	case *tmp2p.PacketPing:
		dst.Sum = &tmp2p.Packet_PacketPing{
			PacketPing: pb,
		}
	case *tmp2p.PacketPong:
		dst.Sum = &tmp2p.Packet_PacketPong{
			PacketPong: pb,
		}
	case *tmp2p.PacketMsg:
		dst.Sum = &tmp2p.Packet_PacketMsg{
			PacketMsg: pb,
		}
	default:
		panic(fmt.Errorf("unknown packet type %T", pb))
	}
}

```

---

<a name="file-4"></a>

### File: `tcp/conn/connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	pbtypes "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
)

const (
	maxPingPongPacketSize = 1024 // bytes
	testStreamID          = 0x01
)

func createMConnectionWithSingleStream(t *testing.T, conn net.Conn) (*MConnection, *MConnectionStream) {
	t.Helper()

	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	c := NewMConnection(conn, cfg)
	c.SetLogger(log.TestingLogger())

	stream, err := c.OpenStream(testStreamID, nil)
	require.NoError(t, err)

	return c, stream.(*MConnectionStream)
}

func TestMConnection_Flush(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	clientConn, clientStream := createMConnectionWithSingleStream(t, client)
	err := clientConn.Start()
	require.NoError(t, err)

	msg := []byte("abc")
	n, err := clientStream.Write(msg)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	// start the reader in a new routine, so we can flush
	errCh := make(chan error)
	go func() {
		buf := make([]byte, 100) // msg + ping
		_, err := server.Read(buf)
		errCh <- err
	}()

	// stop the conn - it should flush all conns
	err = clientConn.FlushAndClose("test")
	require.NoError(t, err)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Error reading from server: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for msgs to be read")
	}
}

func TestMConnection_StreamWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, clientStream := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	msg := []byte("Ant-Man")
	_, err = clientStream.Write(msg)
	require.NoError(t, err)
	// NOTE: subsequent writes could pass because we are reading from
	// the send queue in a separate goroutine.
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
	assert.True(t, mconn.CanSend(testStreamID))

	msg = []byte("Spider-Man")
	require.NoError(t, err)
	_, err = clientStream.TryWrite(msg)
	require.NoError(t, err)
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
}

func TestMConnection_StreamReadWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn1, stream1 := createMConnectionWithSingleStream(t, client)
	err := mconn1.Start()
	require.NoError(t, err)
	defer mconn1.Close("normal")

	mconn2, stream2 := createMConnectionWithSingleStream(t, server)
	err = mconn2.Start()
	require.NoError(t, err)
	defer mconn2.Close("normal")

	// => write
	msg := []byte("Cyclops")
	_, err = stream1.Write(msg)
	require.NoError(t, err)

	// => read
	buf := make([]byte, len(msg))
	n, err := stream2.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.Equal(t, msg, buf)
}

func TestMConnectionStatus(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	state := mconn.ConnState()
	assert.NotNil(t, state)
	assert.Zero(t, state.StreamStates[testStreamID].SendQueueSize)
}

func TestMConnection_PongTimeoutResultsInError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		// read ping
		var pkt tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 200*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
	case <-time.After(pongTimerExpired):
		t.Fatalf("Expected to receive error after %v", pongTimerExpired)
	}
}

func TestMConnection_MultiplePongsInTheBeginning(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pongs in a row (abuse)
	protoWriter := protoio.NewDelimitedWriter(server)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (one byte)
		var packet tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&packet)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 20*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_MultiplePings(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pings in a row (abuse)
	// see https://github.com/tendermint/tendermint/issues/1190
	protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
	protoWriter := protoio.NewDelimitedWriter(server)
	var pkt tmp2p.Packet

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	assert.True(t, mconn.IsRunning())
}

func TestMConnection_PingPongs(t *testing.T) {
	// check that we are not leaking any go-routines
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	server, client := net.Pipe()

	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
		protoWriter := protoio.NewDelimitedWriter(server)
		var pkt tmp2p.PacketPing

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)

		time.Sleep(mconn.config.PingInterval)

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing
	<-serverGotPing

	pongTimerExpired := (mconn.config.PongTimeout + 20*time.Millisecond) * 2
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(2 * pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_StopsAndReturnsError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	if err := client.Close(); err != nil {
		t.Error(err)
	}

	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
		assert.False(t, mconn.IsRunning())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not receive error in 500ms")
	}
}

//nolint:unparam
func newClientAndServerConnsForReadErrors(t *testing.T) (*MConnection, *MConnectionStream, *MConnection, *MConnectionStream) {
	t.Helper()
	server, client := net.Pipe()

	// create client conn with two channels
	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	mconnClient := NewMConnection(client, cfg)
	clientStream, err := mconnClient.OpenStream(testStreamID, StreamDescriptor{ID: testStreamID, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	// create another channel
	_, err = mconnClient.OpenStream(0x02, StreamDescriptor{ID: 0x02, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	mconnClient.SetLogger(log.TestingLogger().With("module", "client"))
	err = mconnClient.Start()
	require.NoError(t, err)

	// create server conn with 1 channel
	// it fires on chOnErr when there's an error
	serverLogger := log.TestingLogger().With("module", "server")
	mconnServer, serverStream := createMConnectionWithSingleStream(t, server)
	mconnServer.SetLogger(serverLogger)
	err = mconnServer.Start()
	require.NoError(t, err)

	return mconnClient, clientStream.(*MConnectionStream), mconnServer, serverStream
}

func assertBytes(t *testing.T, s *MConnectionStream, want []byte) {
	t.Helper()

	buf := make([]byte, len(want))
	n, err := s.Read(buf)
	require.NoError(t, err)
	if assert.Equal(t, len(want), n) {
		assert.Equal(t, want, buf)
	}
}

func gotError(ch <-chan error) bool {
	after := time.After(time.Second * 5)
	select {
	case <-ch:
		return true
	case <-after:
		return false
	}
}

func TestMConnection_ReadErrorBadEncoding(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send badly encoded data
	client := mconnClient.conn
	_, err := client.Write([]byte{1, 2, 3, 4, 5})
	require.NoError(t, err)

	assert.True(t, gotError(mconnServer.ErrorCh()), "badly encoded msgPacket")
}

// func TestMConnection_ReadErrorUnknownChannel(t *testing.T) {
// 	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
// 	defer mconnClient.Close("normal")
// 	defer mconnServer.Close("normal")

// 	msg := []byte("Ant-Man")

// 	// send msg that has unknown channel
// 	client := mconnClient.conn
// 	protoWriter := protoio.NewDelimitedWriter(client)
// 	packet := tmp2p.PacketMsg{
// 		ChannelID: 0x03,
// 		EOF:       true,
// 		Data:      msg,
// 	}
// 	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
// 	require.NoError(t, err)

// 	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown channel")
// }

func TestMConnection_ReadErrorLongMessage(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	msg := make([]byte, mconnClient.config.MaxPacketMsgPayloadSize)
	packet := tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      msg,
	}

	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, msg)

	// send msg that's too long
	packet = tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, mconnClient.config.MaxPacketMsgPayloadSize+100),
	}

	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.Error(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "msg too long")
}

func TestMConnection_ReadErrorUnknownMsgType(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send msg with unknown msg type
	_, err := protoio.NewDelimitedWriter(mconnClient.conn).WriteMsg(&pbtypes.Header{ChainID: "x"})
	require.NoError(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown msg type")
}

//nolint:lll //ignore line length for tests
func TestConnVectors(t *testing.T) {
	testCases := []struct {
		testName string
		msg      proto.Message
		expBytes string
	}{
		{"PacketPing", &tmp2p.PacketPing{}, "0a00"},
		{"PacketPong", &tmp2p.PacketPong{}, "1200"},
		{"PacketMsg", &tmp2p.PacketMsg{ChannelID: 1, EOF: false, Data: []byte("data transmitted over the wire")}, "1a2208011a1e64617461207472616e736d6974746564206f766572207468652077697265"},
	}

	for _, tc := range testCases {
		pm := mustWrapPacket(tc.msg)
		bz, err := pm.Marshal()
		require.NoError(t, err, tc.testName)

		require.Equal(t, tc.expBytes, hex.EncodeToString(bz), tc.testName)
	}
}

func TestMConnection_ChannelOverflow(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	packet := tmp2p.PacketMsg{
		ChannelID: testStreamID,
		EOF:       true,
		Data:      []byte(`42`),
	}
	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, []byte(`42`))

	// channel ID that's too large
	packet.ChannelID = int32(1025)
	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
}

```

---

<a name="file-5"></a>

### File: `tcp/conn/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/p2p/transport"
)

var (
	ErrInvalidSecretConnKeySend = errors.New("send invalid secret connection key")
	ErrInvalidSecretConnKeyRecv = errors.New("invalid receive SecretConnection Key")
	ErrChallengeVerification    = errors.New("challenge verification failed")

	// ErrTimeout is returned when a read or write operation times out.
	ErrTimeout = errors.New("read/write timeout")
)

// ErrWriteQueueFull is returned when the write queue is full.
type ErrWriteQueueFull struct{}

var _ transport.WriteError = ErrWriteQueueFull{}

func (ErrWriteQueueFull) Error() string {
	return "write queue is full"
}

func (ErrWriteQueueFull) Full() bool {
	return true
}

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	Source error
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet message: %v", e.Source)
}

func (e ErrPacketWrite) Unwrap() error {
	return e.Source
}

type ErrUnexpectedPubKeyType struct {
	Expected string
	Got      string
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptFrame struct {
	Source error
}

func (e ErrDecryptFrame) Error() string {
	return fmt.Sprintf("SecretConnection: failed to decrypt the frame: %v", e.Source)
}

func (e ErrDecryptFrame) Unwrap() error {
	return e.Source
}

type ErrPacketTooBig struct {
	Received int
	Max      int
}

func (e ErrPacketTooBig) Error() string {
	return fmt.Sprintf("received message exceeds available capacity (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkTooBig struct {
	Received int
	Max      int
}

func (e ErrChunkTooBig) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}

```

---

<a name="file-6"></a>

### File: `tcp/conn/evil_secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 8 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
)

type buffer struct {
	next bytes.Buffer
}

func (b *buffer) Read(data []byte) (n int, err error) {
	return b.next.Read(data)
}

func (b *buffer) Write(data []byte) (n int, err error) {
	return b.next.Write(data)
}

func (b *buffer) Bytes() []byte {
	return b.next.Bytes()
}

func (*buffer) Close() error {
	return nil
}

type evilConn struct {
	secretConn *SecretConnection
	buffer     *buffer

	locEphPub  *[32]byte
	locEphPriv *[32]byte
	remEphPub  *[32]byte
	privKey    crypto.PrivKey

	readStep   int
	writeStep  int
	readOffset int

	shareEphKey        bool
	badEphKey          bool
	shareAuthSignature bool
	badAuthSignature   bool
}

func newEvilConn(shareEphKey, badEphKey, shareAuthSignature, badAuthSignature bool) *evilConn {
	privKey := ed25519.GenPrivKey()
	locEphPub, locEphPriv := genEphKeys()
	var rep [32]byte
	c := &evilConn{
		locEphPub:  locEphPub,
		locEphPriv: locEphPriv,
		remEphPub:  &rep,
		privKey:    privKey,

		shareEphKey:        shareEphKey,
		badEphKey:          badEphKey,
		shareAuthSignature: shareAuthSignature,
		badAuthSignature:   badAuthSignature,
	}

	return c
}

func (c *evilConn) Read(data []byte) (n int, err error) {
	if !c.shareEphKey {
		return 0, io.EOF
	}

	switch c.readStep {
	case 0:
		if !c.badEphKey {
			lc := *c.locEphPub
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				panic(err)
			}
			copy(data, bz[c.readOffset:])
			n = len(data)
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("drop users;")})
			if err != nil {
				panic(err)
			}
			copy(data, bz)
			n = len(data)
		}
		c.readOffset += n

		if n >= 32 {
			c.readOffset = 0
			c.readStep = 1
			if !c.shareAuthSignature {
				c.readStep = 2
			}
		}

		return n, nil
	case 1:
		signature := c.signChallenge()
		if !c.badAuthSignature {
			pkpb, err := cryptoenc.PubKeyToProto(c.privKey.PubKey())
			if err != nil {
				panic(err)
			}
			bz, err := protoio.MarshalDelimited(&tmp2p.AuthSigMessage{PubKey: pkpb, Sig: signature})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes()[c.readOffset:])
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("select * from users;")})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes())
		}
		c.readOffset += len(data)
		return n, nil
	default:
		return 0, io.EOF
	}
}

func (c *evilConn) Write(data []byte) (n int, err error) {
	switch c.writeStep {
	case 0:
		var (
			bytes     gogotypes.BytesValue
			remEphPub [32]byte
		)
		err := protoio.UnmarshalDelimited(data, &bytes)
		if err != nil {
			panic(err)
		}
		copy(remEphPub[:], bytes.Value)
		c.remEphPub = &remEphPub
		c.writeStep = 1
		if !c.shareAuthSignature {
			c.writeStep = 2
		}
		return len(data), nil
	case 1:
		// Signature is not needed, therefore skipped.
		return len(data), nil
	default:
		return 0, io.EOF
	}
}

func (*evilConn) Close() error {
	return nil
}

func (c *evilConn) signChallenge() []byte {
	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(c.locEphPub, c.remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least, lexicographically
	// sorted.
	locIsLeast := bytes.Equal(c.locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(c.remEphPub, c.locEphPriv)
	if err != nil {
		panic(err)
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via HKDF-SHA2
	// on the transcript state (which itself also uses HKDF-SHA2 to derive a key
	// from the dhSecret).
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		panic(errors.New("invalid send SecretConnection Key"))
	}
	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		panic(errors.New("invalid receive SecretConnection Key"))
	}

	b := &buffer{}
	c.secretConn = &SecretConnection{
		conn:            b,
		connWriter:      bufio.NewWriterSize(b, defaultWriteBufferSize),
		connReader:      b,
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
	}
	c.buffer = b

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, c.privKey)
	if err != nil {
		panic(err)
	}

	return locSignature
}

// TestMakeSecretConnection creates an evil connection and tests that
// MakeSecretConnection errors at different stages.
func TestMakeSecretConnection(t *testing.T) {
	testCases := []struct {
		name       string
		conn       *evilConn
		checkError func(error) bool // Function to check if the error matches the expectation
	}{
		{"refuse to share ethimeral key", newEvilConn(false, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad ethimeral key", newEvilConn(true, true, false, false), func(err error) bool { return assert.Contains(t, err.Error(), "wrong wireType") }},
		{"refuse to share auth signature", newEvilConn(true, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad auth signature", newEvilConn(true, false, true, true), func(err error) bool { return errors.As(err, &ErrDecryptFrame{}) }},
		// fails with the introduction of changes PR #3419
		// {"all good", newEvilConn(true, false, true, false), func(err error) bool { return err == nil }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privKey := ed25519.GenPrivKey()
			_, err := MakeSecretConnection(tc.conn, privKey)
			if tc.checkError != nil {
				assert.True(t, tc.checkError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

```

---

<a name="file-7"></a>

### File: `tcp/conn/secret_connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/internal/async"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// 4 + 1024 == 1028 total frame size.
const (
	dataLenSize      = 4
	dataMaxSize      = 1024
	totalFrameSize   = dataMaxSize + dataLenSize
	aeadSizeOverhead = 16 // overhead of poly 1305 authentication tag
	aeadKeySize      = chacha20poly1305.KeySize
	aeadNonceSize    = chacha20poly1305.NonceSize

	labelEphemeralLowerPublicKey = "EPHEMERAL_LOWER_PUBLIC_KEY"
	labelEphemeralUpperPublicKey = "EPHEMERAL_UPPER_PUBLIC_KEY"
	labelDHSecret                = "DH_SECRET"
	labelSecretConnectionMac     = "SECRET_CONNECTION_MAC"

	defaultWriteBufferSize = 128 * 1024
	// try to read the biggest logical packet we can get, in one read.
	// biggest logical packet is encoding_overhead(64kb).
	defaultReadBufferSize = 65 * 1024
)

var (
	ErrSmallOrderRemotePubKey    = errors.New("detected low order point from remote peer")
	secretConnKeyAndChallengeGen = []byte("TENDERMINT_SECRET_CONNECTION_KEY_AND_CHALLENGE_GEN")
)

// SecretConnection implements net.Conn.
// It is an implementation of the STS protocol.
// For more details regarding this implementation of the STS protocol, please refer to:
// https://github.com/cometbft/cometbft/blob/main/spec/p2p/legacy-docs/peer.md#authenticated-encryption-handshake.
//
// The original STS protocol, which inspired this implementation:
// https://citeseerx.ist.psu.edu/document?rapid=rep1&type=pdf&doi=b852bc961328ce74f7231a4b569eec1ab6c3cf50. # codespell:ignore
//
// Consumers of the SecretConnection are responsible for authenticating
// the remote peer's pubkey against known information, like a nodeID.
type SecretConnection struct {
	// immutable
	recvAead cipher.AEAD
	sendAead cipher.AEAD

	remPubKey crypto.PubKey

	conn       io.ReadWriteCloser
	connWriter *bufio.Writer
	connReader io.Reader

	// net.Conn must be thread safe:
	// https://golang.org/pkg/net/#Conn.
	// Since we have internal mutable state,
	// we need mtxs. But recv and send states
	// are independent, so we can use two mtxs.
	// All .Read are covered by recvMtx,
	// all .Write are covered by sendMtx.
	recvMtx         cmtsync.Mutex
	recvBuffer      []byte
	recvNonce       *[aeadNonceSize]byte
	recvFrame       []byte
	recvSealedFrame []byte

	sendMtx         cmtsync.Mutex
	sendNonce       *[aeadNonceSize]byte
	sendFrame       []byte
	sendSealedFrame []byte
}

// MakeSecretConnection performs handshake and returns a new authenticated
// SecretConnection.
// Returns nil if there is an error in handshake.
// Caller should call conn.Close().
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	// NOTE: every 32-byte string is accepted as a Curve25519 public key (see
	// DJB's Curve25519 paper: http://cr.yp.to/ecdh/curve25519-20060209.pdf)
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, err
	}

	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least,
	// lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, err
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via
	// HKDF-SHA2 on the dhSecret.
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeySend
	}

	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeyRecv
	}

	sc := &SecretConnection{
		conn:            conn,
		connWriter:      bufio.NewWriterSize(conn, defaultWriteBufferSize),
		connReader:      bufio.NewReaderSize(conn, defaultReadBufferSize),
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
	}

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, locPrivKey)
	if err != nil {
		return nil, err
	}

	// Share (in secret) each other's pubkey & challenge signature
	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, err
	}

	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	// Usage in your function
	if _, ok := remPubKey.(ed25519.PubKey); !ok {
		return nil, ErrUnexpectedPubKeyType{
			Expected: ed25519.KeyType,
			Got:      remPubKey.Type(),
		}
	}

	if !remPubKey.VerifySignature(challenge[:], remSignature) {
		return nil, ErrChallengeVerification
	}

	// We've authorized.
	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey returns authenticated remote pubkey.
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Writes encrypted frames of `totalFrameSize + aeadSizeOverhead`.
// CONTRACT: data smaller than dataMaxSize is written atomically.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	sc.sendMtx.Lock()
	defer sc.sendMtx.Unlock()
	sealedFrame, frame := sc.sendSealedFrame, sc.sendFrame

	for 0 < len(data) {
		if err := func() error {
			var chunk []byte
			if dataMaxSize < len(data) {
				chunk = data[:dataMaxSize]
				data = data[dataMaxSize:]
			} else {
				chunk = data
				data = nil
			}
			chunkLength := len(chunk)
			binary.LittleEndian.PutUint32(frame, uint32(chunkLength))
			copy(frame[dataLenSize:], chunk)

			// encrypt the frame
			sc.sendAead.Seal(sealedFrame[:0], sc.sendNonce[:], frame, nil)
			incrNonce(sc.sendNonce)
			// end encryption

			_, err = sc.connWriter.Write(sealedFrame)
			if err != nil {
				return err
			}

			n += len(chunk)
			return nil
		}(); err != nil {
			return n, err
		}
	}
	sc.connWriter.Flush()
	return n, err
}

// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	sc.recvMtx.Lock()
	defer sc.recvMtx.Unlock()

	// read off and update the recvBuffer, if non-empty
	if 0 < len(sc.recvBuffer) {
		n = copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[n:]
		return n, err
	}

	// read off the conn
	sealedFrame := sc.recvSealedFrame
	_, err = io.ReadFull(sc.connReader, sealedFrame)
	if err != nil {
		return n, err
	}

	// decrypt the frame.
	// reads and updates the sc.recvNonce
	frame := sc.recvFrame
	_, err = sc.recvAead.Open(frame[:0], sc.recvNonce[:], sealedFrame, nil)
	if err != nil {
		return n, ErrDecryptFrame{Source: err}
	}

	incrNonce(sc.recvNonce)
	// end decryption

	// copy checkLength worth into data,
	// set recvBuffer to the rest.
	chunkLength := binary.LittleEndian.Uint32(frame) // read the first four bytes
	if chunkLength > dataMaxSize {
		return 0, ErrChunkTooBig{
			Received: int(chunkLength),
			Max:      dataMaxSize,
		}
	}

	chunk := frame[dataLenSize : dataLenSize+chunkLength]
	n = copy(data, chunk)
	if n < len(chunk) {
		sc.recvBuffer = make([]byte, len(chunk)-n)
		copy(sc.recvBuffer, chunk[n:])
	}
	return n, err
}

// Implements net.Conn.
func (sc *SecretConnection) Close() error                  { return sc.conn.Close() }
func (sc *SecretConnection) LocalAddr() net.Addr           { return sc.conn.(net.Conn).LocalAddr() }
func (sc *SecretConnection) RemoteAddr() net.Addr          { return sc.conn.(net.Conn).RemoteAddr() }
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		panic("failed to generate ephemeral key-pair")
	}
	return ephPub, ephPriv
}

func shareEphPubKey(conn io.ReadWriter, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	// Send our pubkey and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			lc := *locEphPub
			_, err = protoio.NewDelimitedWriter(conn).WriteMsg(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var bytes gogotypes.BytesValue
			_, err = protoio.NewDelimitedReader(conn, 1024*1024).ReadMsg(&bytes)
			if err != nil {
				return nil, true, err // abort
			}

			var _remEphPub [32]byte
			copy(_remEphPub[:], bytes.Value)
			return _remEphPub, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return remEphPub, err
	}

	// Otherwise:
	_remEphPub := trs.FirstValue().([32]byte)
	return &_remEphPub, nil
}

func deriveSecrets(
	dhSecret *[32]byte,
	locIsLeast bool,
) (recvSecret, sendSecret *[aeadKeySize]byte) {
	hash := sha256.New
	hkdf := hkdf.New(hash, dhSecret[:], nil, secretConnKeyAndChallengeGen)
	// get enough data for 2 aead keys, and a 32 byte challenge
	res := new([2*aeadKeySize + 32]byte)
	_, err := io.ReadFull(hkdf, res[:])
	if err != nil {
		panic(err)
	}

	recvSecret = new([aeadKeySize]byte)
	sendSecret = new([aeadKeySize]byte)

	// bytes 0 through aeadKeySize - 1 are one aead key.
	// bytes aeadKeySize through 2*aeadKeySize -1 are another aead key.
	// which key corresponds to sending and receiving key depends on whether
	// the local key is less than the remote key.
	if locIsLeast {
		copy(recvSecret[:], res[0:aeadKeySize])
		copy(sendSecret[:], res[aeadKeySize:aeadKeySize*2])
	} else {
		copy(sendSecret[:], res[0:aeadKeySize])
		copy(recvSecret[:], res[aeadKeySize:aeadKeySize*2])
	}

	return recvSecret, sendSecret
}

// computeDHSecret computes a Diffie-Hellman shared secret key
// from our own local private key and the other's public key.
func computeDHSecret(remPubKey, locPrivKey *[32]byte) (*[32]byte, error) {
	shrKey, err := curve25519.X25519(locPrivKey[:], remPubKey[:])
	if err != nil {
		return nil, err
	}
	var shrKeyArray [32]byte
	copy(shrKeyArray[:], shrKey)
	return &shrKeyArray, nil
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		lo = foo
		hi = bar
	} else {
		lo = bar
		hi = foo
	}
	return lo, hi
}

func signChallenge(challenge *[32]byte, locPrivKey crypto.PrivKey) ([]byte, error) {
	signature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig []byte
}

func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature []byte) (recvMsg authSigMessage, err error) {
	// Send our info and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			pbpk, err := cryptoenc.PubKeyToProto(pubKey)
			if err != nil {
				return nil, true, err
			}
			_, err = protoio.NewDelimitedWriter(sc).WriteMsg(&tmp2p.AuthSigMessage{PubKey: pbpk, Sig: signature})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var pba tmp2p.AuthSigMessage
			_, err = protoio.NewDelimitedReader(sc, 1024*1024).ReadMsg(&pba)
			if err != nil {
				return nil, true, err // abort
			}

			pk, err := cryptoenc.PubKeyFromProto(pba.PubKey)
			if err != nil {
				return nil, true, err // abort
			}

			_recvMsg := authSigMessage{
				Key: pk,
				Sig: pba.Sig,
			}
			return _recvMsg, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return recvMsg, err
	}

	_recvMsg := trs.FirstValue().(authSigMessage)
	return _recvMsg, nil
}

// --------------------------------------------------------------------------------

// Increment nonce little-endian by 1 with wraparound.
// Due to chacha20poly1305 expecting a 12 byte nonce we do not use the first four
// bytes. We only increment a 64 bit unsigned int in the remaining 8 bytes
// (little-endian in nonce[4:]).
func incrNonce(nonce *[aeadNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		// Terminates the session and makes sure the nonce would not re-used.
		// See https://github.com/tendermint/tendermint/issues/3531
		panic("can't increase nonce without overflow")
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}

```

---

<a name="file-8"></a>

### File: `tcp/conn/secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 13 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/internal/async"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
)

// Run go test -update from within this module
// to update the golden test vector file.
var update = flag.Bool("update", false, "update .golden files")

type kvstoreConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (drw kvstoreConn) Close() (err error) {
	err2 := drw.PipeWriter.CloseWithError(io.EOF)
	err1 := drw.PipeReader.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

type privKeyWithNilPubKey struct {
	orig crypto.PrivKey
}

func (pk privKeyWithNilPubKey) Bytes() []byte                   { return pk.orig.Bytes() }
func (pk privKeyWithNilPubKey) Sign(msg []byte) ([]byte, error) { return pk.orig.Sign(msg) }
func (privKeyWithNilPubKey) PubKey() crypto.PubKey              { return nil }
func (privKeyWithNilPubKey) Type() string                       { return "privKeyWithNilPubKey" }

func TestSecretConnectionHandshake(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
	if err := barSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentWrite(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)

	// write from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	n := 100
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)

	// Consume reads from bar's reader
	readLots(t, wg, barSecConn, n*2)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentRead(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)
	n := 100

	// read from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go readLots(t, wg, fooSecConn, n/2)
	go readLots(t, wg, fooSecConn, n/2)

	// write to bar
	writeLots(t, wg, barSecConn, fooWriteText, n)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestSecretConnectionReadWrite(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	fooWrites, barWrites := []string{}, []string{}
	fooReads, barReads := []string{}, []string{}

	// Pre-generate the things to write (for foo & bar)
	for i := 0; i < 100; i++ {
		fooWrites = append(fooWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
		barWrites = append(barWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
	}

	// A helper that will run with (fooConn, fooWrites, fooReads) and vice versa
	genNodeRunner := func(nodeConn kvstoreConn, nodeWrites []string, nodeReads *[]string) async.Task {
		return func(_ int) (any, bool, error) {
			// Initiate cryptographic private key and secret connection through nodeConn.
			nodePrvKey := ed25519.GenPrivKey()
			nodeSecretConn, err := MakeSecretConnection(nodeConn, nodePrvKey)
			if err != nil {
				t.Errorf("failed to establish SecretConnection for node: %v", err)
				return nil, true, err
			}
			// In parallel, handle some reads and writes.
			trs, ok := async.Parallel(
				func(_ int) (any, bool, error) {
					// Node writes:
					for _, nodeWrite := range nodeWrites {
						n, err := nodeSecretConn.Write([]byte(nodeWrite))
						if err != nil {
							t.Errorf("failed to write to nodeSecretConn: %v", err)
							return nil, true, err
						}
						if n != len(nodeWrite) {
							err = fmt.Errorf("failed to write all bytes. Expected %v, wrote %v", len(nodeWrite), n)
							t.Error(err)
							return nil, true, err
						}
					}
					if err := nodeConn.PipeWriter.Close(); err != nil {
						t.Error(err)
						return nil, true, err
					}
					return nil, false, nil
				},
				func(_ int) (any, bool, error) {
					// Node reads:
					readBuffer := make([]byte, dataMaxSize)
					for {
						n, err := nodeSecretConn.Read(readBuffer)
						if errors.Is(err, io.EOF) {
							if err := nodeConn.PipeReader.Close(); err != nil {
								t.Error(err)
								return nil, true, err
							}
							return nil, false, nil
						} else if err != nil {
							t.Errorf("failed to read from nodeSecretConn: %v", err)
							return nil, true, err
						}
						*nodeReads = append(*nodeReads, string(readBuffer[:n]))
					}
				},
			)
			assert.True(t, ok, "Unexpected task abortion")

			// If error:
			if trs.FirstError() != nil {
				return nil, true, trs.FirstError()
			}

			// Otherwise:
			return nil, false, nil
		}
	}

	// Run foo & bar in parallel
	trs, ok := async.Parallel(
		genNodeRunner(fooConn, fooWrites, &fooReads),
		genNodeRunner(barConn, barWrites, &barReads),
	)
	require.NoError(t, trs.FirstError())
	require.True(t, ok, "unexpected task abortion")

	// A helper to ensure that the writes and reads match.
	// Additionally, small writes (<= dataMaxSize) must be atomically read.
	compareWritesReads := func(writes []string, reads []string) {
		for {
			// Pop next write & corresponding reads
			read := ""
			write := writes[0]
			readCount := 0
			for _, readChunk := range reads {
				read += readChunk
				readCount++
				if len(write) <= len(read) {
					break
				}
				if len(write) <= dataMaxSize {
					break // atomicity of small writes
				}
			}
			// Compare
			if write != read {
				t.Errorf("expected to read %X, got %X", write, read)
			}
			// Iterate
			writes = writes[1:]
			reads = reads[readCount:]
			if len(writes) == 0 {
				break
			}
		}
	}

	compareWritesReads(fooWrites, barReads)
	compareWritesReads(barWrites, fooReads)
}

func TestDeriveSecretsAndChallengeGolden(t *testing.T) {
	goldenFilepath := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("Updating golden test vector file %s", goldenFilepath)
		data := createGoldenTestVectors(t)
		err := cmtos.WriteFile(goldenFilepath, []byte(data), 0o644)
		require.NoError(t, err)
	}
	f, err := os.Open(goldenFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		randSecretVector, err := hex.DecodeString(params[0])
		require.NoError(t, err)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		locIsLeast, err := strconv.ParseBool(params[1])
		require.NoError(t, err)
		expectedRecvSecret, err := hex.DecodeString(params[2])
		require.NoError(t, err)
		expectedSendSecret, err := hex.DecodeString(params[3])
		require.NoError(t, err)

		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		require.Equal(t, expectedRecvSecret, (*recvSecret)[:], "Recv Secrets aren't equal")
		require.Equal(t, expectedSendSecret, (*sendSecret)[:], "Send Secrets aren't equal")
	}
}

func TestNilPubkey(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	defer fooConn.Close()
	defer barConn.Close()
	fooPrvKey := ed25519.GenPrivKey()
	barPrvKey := privKeyWithNilPubKey{ed25519.GenPrivKey()}

	go MakeSecretConnection(fooConn, fooPrvKey) //nolint:errcheck // ignore for tests

	_, err := MakeSecretConnection(barConn, barPrvKey)
	require.Error(t, err)
	assert.Equal(t, "encoding: unsupported key <nil>", err.Error())
}

func writeLots(t *testing.T, wg *sync.WaitGroup, conn io.Writer, txt string, n int) {
	t.Helper()
	defer wg.Done()
	for i := 0; i < n; i++ {
		_, err := conn.Write([]byte(txt))
		if err != nil {
			t.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
}

func readLots(t *testing.T, wg *sync.WaitGroup, conn io.Reader, n int) {
	t.Helper()
	readBuffer := make([]byte, dataMaxSize)
	for i := 0; i < n; i++ {
		_, err := conn.Read(readBuffer)
		require.NoError(t, err)
	}
	wg.Done()
}

// Creates the data for a test vector file.
// The file format is:
// Hex(diffie_hellman_secret), loc_is_least, Hex(recvSecret), Hex(sendSecret), Hex(challenge).
func createGoldenTestVectors(*testing.T) string {
	data := ""
	for i := 0; i < 32; i++ {
		randSecretVector := cmtrand.Bytes(32)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		data += hex.EncodeToString((*randSecret)[:]) + ","
		locIsLeast := cmtrand.Bool()
		data += strconv.FormatBool(locIsLeast) + ","
		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		data += hex.EncodeToString((*recvSecret)[:]) + ","
		data += hex.EncodeToString((*sendSecret)[:]) + ","
	}
	return data
}

// Each returned ReadWriteCloser is akin to a net.Connection.
func makeKVStoreConnPair() (fooConn, barConn kvstoreConn) {
	barReader, fooWriter := io.Pipe()
	fooReader, barWriter := io.Pipe()
	return kvstoreConn{fooReader, fooWriter}, kvstoreConn{barReader, barWriter}
}

func makeSecretConnPair(tb testing.TB) (fooSecConn, barSecConn *SecretConnection) {
	tb.Helper()
	var (
		fooConn, barConn = makeKVStoreConnPair()
		fooPrvKey        = ed25519.GenPrivKey()
		fooPubKey        = fooPrvKey.PubKey()
		barPrvKey        = ed25519.GenPrivKey()
		barPubKey        = barPrvKey.PubKey()
	)

	// Make connections from both sides in parallel.
	trs, ok := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			fooSecConn, err = MakeSecretConnection(fooConn, fooPrvKey)
			if err != nil {
				tb.Errorf("failed to establish SecretConnection for foo: %v", err)
				return nil, true, err
			}
			remotePubBytes := fooSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), barPubKey.Bytes()) {
				err = fmt.Errorf("unexpected fooSecConn.RemotePubKey.  Expected %v, got %v",
					barPubKey, fooSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			barSecConn, err = MakeSecretConnection(barConn, barPrvKey)
			if barSecConn == nil {
				tb.Errorf("failed to establish SecretConnection for bar: %v", err)
				return nil, true, err
			}
			remotePubBytes := barSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), fooPubKey.Bytes()) {
				err = fmt.Errorf("unexpected barSecConn.RemotePubKey.  Expected %v, got %v",
					fooPubKey, barSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
	)

	require.NoError(tb, trs.FirstError())
	require.True(tb, ok, "Unexpected task abortion")

	return fooSecConn, barSecConn
}

// Benchmarks

func BenchmarkWriteSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	// Consume reads from bar's reader
	go func() {
		readBuffer := make([]byte, dataMaxSize)
		for {
			_, err := barSecConn.Read(readBuffer)
			if errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				b.Errorf("failed to read from barSecConn: %v", err)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := cmtrand.Intn(len(fooWriteBytes))
		_, err := fooSecConn.Write(fooWriteBytes[idx])
		if err != nil {
			b.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
	b.StopTimer()

	if err := fooSecConn.Close(); err != nil {
		b.Error(err)
	}
	// barSecConn.Close() race condition
}

func BenchmarkReadSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	go func() {
		for i := 0; i < b.N; i++ {
			idx := cmtrand.Intn(len(fooWriteBytes))
			_, err := fooSecConn.Write(fooWriteBytes[idx])
			if err != nil {
				b.Errorf("failed to write to fooSecConn: %v, %v,%v", err, i, b.N)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		readBuffer := make([]byte, dataMaxSize)
		_, err := barSecConn.Read(readBuffer)

		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			b.Fatalf("Failed to read from barSecConn: %v", err)
		}
	}
	b.StopTimer()
}

```

---

<a name="file-9"></a>

### File: `tcp/conn/stream.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import "time"

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn     *MConnection
	streamID byte
}

// Read reads bytes for the given stream from the internal read queue. Used in
// tests. Production code should use MConnection.OnReceive to avoid copying the
// data.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	return s.conn.readBytes(s.streamID, b, 5*time.Second)
}

// Write queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) Write(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, true /* blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// TryWrite queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) TryWrite(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, false /* non-blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Close closes the stream.
// thread-safe.
func (s *MConnectionStream) Close() error {
	delete(s.conn.channelsIdx, s.streamID)
	return nil
}

```

---

<a name="file-10"></a>

### File: `tcp/conn/stream_descriptor.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
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

```

---

<a name="file-11"></a>

### File: `tcp/conn_set.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"net"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(conn net.Conn) bool
	HasIP(ip net.IP) bool
	Set(conn net.Conn, ip []net.IP)
	Remove(conn net.Conn)
	RemoveAddr(addr net.Addr)
}

type connSetItem struct {
	conn net.Conn
	ips  []net.IP
}

type connSet struct {
	cmtsync.RWMutex

	conns map[string]connSetItem
}

// NewConnSet returns a ConnSet implementation.
func NewConnSet() ConnSet {
	return &connSet{
		conns: map[string]connSetItem{},
	}
}

func (cs *connSet) Has(c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[c.RemoteAddr().String()]

	return ok
}

func (cs *connSet) HasIP(ip net.IP) bool {
	cs.RLock()
	defer cs.RUnlock()

	for _, c := range cs.conns {
		for _, known := range c.ips {
			if known.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func (cs *connSet) Remove(c net.Conn) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, c.RemoteAddr().String())
}

func (cs *connSet) RemoveAddr(addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, addr.String())
}

func (cs *connSet) Set(c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[c.RemoteAddr().String()] = connSetItem{
		conn: c,
		ips:  ips,
	}
}

```

---

<a name="file-12"></a>

### File: `tcp/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
)

// ErrTransportClosed is raised when the Transport has been closed.
type ErrTransportClosed struct{}

func (ErrTransportClosed) Error() string {
	return "transport has been closed"
}

// ErrFilterTimeout indicates that a filter operation timed out.
type ErrFilterTimeout struct{}

func (ErrFilterTimeout) Error() string {
	return "filter timed out"
}

// ErrRejected indicates that a Peer was rejected carrying additional
// information as to the reason.
type ErrRejected struct {
	addr          na.NetAddr
	conn          net.Conn
	err           error
	id            nodekey.ID
	isAuthFailure bool
	isDuplicate   bool
	isFiltered    bool
}

// Addr returns the network address for the rejected Peer.
func (e ErrRejected) Addr() na.NetAddr {
	return e.addr
}

func (e ErrRejected) Error() string {
	if e.isAuthFailure {
		return fmt.Sprintf("auth failure: %s", e.err)
	}

	if e.isDuplicate {
		if e.conn != nil {
			return fmt.Sprintf(
				"duplicate CONN<%s>",
				e.conn.RemoteAddr().String(),
			)
		}
		if e.id != "" {
			return fmt.Sprintf("duplicate ID<%v>", e.id)
		}
	}

	if e.isFiltered {
		if e.conn != nil {
			return fmt.Sprintf(
				"filtered CONN<%s>: %s",
				e.conn.RemoteAddr().String(),
				e.err,
			)
		}

		if e.id != "" {
			return fmt.Sprintf("filtered ID<%v>: %s", e.id, e.err)
		}
	}

	return e.err.Error()
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e ErrRejected) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e ErrRejected) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e ErrRejected) IsFiltered() bool { return e.isFiltered }

```

---

<a name="file-13"></a>

### File: `tcp/tcp.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 11 KB

```go
package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/netutil"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/fuzz"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

const (
	defaultDialTimeout      = time.Second
	defaultFilterTimeout    = 5 * time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// IPResolver is a behavior subset of net.Resolver.
type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// accept is the container to carry the upgraded connection from an
// asynchronously running routine to the Accept method.
type accept struct {
	netAddr *na.NetAddr
	conn    *conn.MConnection
	err     error
}

// ConnFilterFunc to be implemented by filter hooks after a new connection has
// been established. The set of existing connections is passed along together
// with all resolved IPs for the new connection.
type ConnFilterFunc func(ConnSet, net.Conn, []net.IP) error

// ConnDuplicateIPFilter resolves and keeps all ips for an incoming connection
// and refuses new ones if they come from a known ip.
func ConnDuplicateIPFilter() ConnFilterFunc {
	return func(cs ConnSet, c net.Conn, ips []net.IP) error {
		for _, ip := range ips {
			if cs.HasIP(ip) {
				return ErrRejected{
					conn:        c,
					err:         fmt.Errorf("ip<%v> already connected", ip),
					isDuplicate: true,
				}
			}
		}

		return nil
	}
}

// MultiplexTransportOption sets an optional parameter on the
// MultiplexTransport.
type MultiplexTransportOption func(*MultiplexTransport)

// MultiplexTransportConnFilters sets the filters for rejection new connections.
func MultiplexTransportConnFilters(
	filters ...ConnFilterFunc,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.connFilters = filters }
}

// MultiplexTransportFilterTimeout sets the timeout waited for filter calls to
// return.
func MultiplexTransportFilterTimeout(
	timeout time.Duration,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.filterTimeout = timeout }
}

// MultiplexTransportResolver sets the Resolver used for ip lokkups, defaults to
// net.DefaultResolver.
func MultiplexTransportResolver(resolver IPResolver) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.resolver = resolver }
}

// MultiplexTransportMaxIncomingConnections sets the maximum number of
// simultaneous connections (incoming). Default: 0 (unlimited).
func MultiplexTransportMaxIncomingConnections(n int) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.maxIncomingConnections = n }
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	netAddr                na.NetAddr
	listener               net.Listener
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	conns       ConnSet
	connFilters []ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeKey          nodekey.NodeKey
	resolver         IPResolver

	// TODO(xla): This config is still needed as we parameterise peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig *conn.MConnConfig
	logger  log.Logger
}

// Test multiplexTransport for interface completeness.
var (
	_ transport.Transport = (*MultiplexTransport)(nil)
)

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(nodeKey nodekey.NodeKey, mConfig conn.MConnConfig) *MultiplexTransport {
	return &MultiplexTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      defaultDialTimeout,
		filterTimeout:    defaultFilterTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          &mConfig,
		nodeKey:          nodeKey,
		conns:            NewConnSet(),
		resolver:         net.DefaultResolver,
		logger:           log.NewNopLogger(),
	}
}

// SetLogger sets the logger for the transport.
func (mt *MultiplexTransport) SetLogger(l log.Logger) {
	mt.logger = l
}

// NetAddr implements Transport.
func (mt *MultiplexTransport) NetAddr() na.NetAddr {
	return mt.netAddr
}

// Accept implements Transport.
func (mt *MultiplexTransport) Accept() (transport.Conn, *na.NetAddr, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, nil, a.err
		}

		return a.conn, a.netAddr, nil
	case <-mt.closec:
		return nil, nil, ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *MultiplexTransport) Dial(addr na.NetAddr) (transport.Conn, error) {
	c, err := addr.DialTimeout(mt.dialTimeout)
	if err != nil {
		return nil, err
	}

	if mt.mConfig.TestFuzz {
		// so we have time to do peer handshakes and get set up.
		c = fuzz.ConnAfterFromConfig(c, 10*time.Second, mt.mConfig.TestFuzzConfig)
	}

	// TODO(xla): Evaluate if we should apply filters if we explicitly dial.
	if err := mt.filterConn(c); err != nil {
		return nil, err
	}

	mconn, _, err := mt.upgrade(c, &addr)
	if err != nil {
		return nil, err
	}
	mconn.SetLogger(mt.logger.With("remote", addr))

	go mt.cleanupConn(c.RemoteAddr(), mconn.Quit())

	return mconn, nil
}

func (mt *MultiplexTransport) Close() error {
	close(mt.closec)

	if mt.listener != nil {
		return mt.listener.Close()
	}

	return nil
}

func (mt *MultiplexTransport) Listen(addr na.NetAddr) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return err
	}

	if mt.maxIncomingConnections > 0 {
		ln = netutil.LimitListener(ln, mt.maxIncomingConnections)
	}

	mt.netAddr = *na.New(addr.ID, ln.Addr())
	mt.listener = ln

	go mt.acceptPeers()

	return nil
}

func (mt *MultiplexTransport) cleanupConn(netAddr net.Addr, quitCh <-chan struct{}) {
	select {
	case <-quitCh:
		mt.conns.RemoveAddr(netAddr)
	case <-mt.closec:
		return
	}
}

func (mt *MultiplexTransport) acceptPeers() {
	for {
		c, err := mt.listener.Accept()
		if err != nil {
			// If Close() has been called, silently exit.
			select {
			case _, ok := <-mt.closec:
				if !ok {
					return
				}
			default:
				// Transport is not closed
			}

			mt.acceptc <- accept{err: err}
			return
		}

		// Connection upgrade and filtering should be asynchronous to avoid
		// Head-of-line blocking[0].
		// Reference:  https://github.com/tendermint/tendermint/issues/2047
		//
		// [0] https://en.wikipedia.org/wiki/Head-of-line_blocking
		go func(c net.Conn) {
			defer func() {
				if r := recover(); r != nil {
					err := ErrRejected{
						conn:          c,
						err:           fmt.Errorf("recovered from panic: %v", r),
						isAuthFailure: true,
					}
					select {
					case mt.acceptc <- accept{err: err}:
					case <-mt.closec:
						// Give up if the transport was closed.
						_ = c.Close()
						return
					}
				}
			}()

			var (
				mconn        *conn.MConnection
				remotePubKey crypto.PubKey
				netAddr      *na.NetAddr
			)

			err := mt.filterConn(c)
			if err == nil {
				mconn, remotePubKey, err = mt.upgrade(c, nil)
				if err == nil {
					addr := c.RemoteAddr()
					id := nodekey.PubKeyToID(remotePubKey)
					netAddr = na.New(id, addr)
					mconn.SetLogger(mt.logger.With("remote", netAddr))
					go mt.cleanupConn(addr, mconn.Quit())
				}
			}

			select {
			case mt.acceptc <- accept{netAddr, mconn, err}:
				// Make the upgraded peer available.
			case <-mt.closec:
				// Give up if the transport was closed.
				_ = c.Close()
				return
			}
		}(c)
	}
}

func (mt *MultiplexTransport) filterConn(c net.Conn) (err error) {
	defer func() {
		if err != nil {
			_ = c.Close()
		}
	}()

	// Reject if connection is already present.
	if mt.conns.Has(c) {
		return ErrRejected{conn: c, isDuplicate: true}
	}

	// Resolve ips for incoming conn.
	ips, err := resolveIPs(mt.resolver, c)
	if err != nil {
		return err
	}

	errc := make(chan error, len(mt.connFilters))

	for _, f := range mt.connFilters {
		go func(f ConnFilterFunc, c net.Conn, ips []net.IP, errc chan<- error) {
			errc <- f(mt.conns, c, ips)
		}(f, c, ips, errc)
	}

	for i := 0; i < cap(errc); i++ {
		select {
		case err := <-errc:
			if err != nil {
				return ErrRejected{conn: c, err: err, isFiltered: true}
			}
		case <-time.After(mt.filterTimeout):
			return ErrFilterTimeout{}
		}
	}

	mt.conns.Set(c, ips)

	return nil
}

func (mt *MultiplexTransport) upgrade(
	c net.Conn,
	dialedAddr *na.NetAddr,
) (*conn.MConnection, crypto.PubKey, error) {
	var err error
	defer func() {
		if err != nil {
			mt.conns.Remove(c)
			_ = c.Close()
		}
	}()

	secretConn, err := upgradeSecretConn(c, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	if err != nil {
		return nil, nil, ErrRejected{
			conn:          c,
			err:           fmt.Errorf("secret conn failed: %w", err),
			isAuthFailure: true,
		}
	}

	// For outgoing conns, ensure connection key matches dialed key.
	remotePubKey := secretConn.RemotePubKey()
	connID := nodekey.PubKeyToID(remotePubKey)
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID != dialedID {
			return nil, nil, ErrRejected{
				conn: c,
				id:   connID,
				err: fmt.Errorf(
					"conn.ID (%v) dialed ID (%v) mismatch",
					connID,
					dialedID,
				),
				isAuthFailure: true,
			}
		}
	}

	// Copy MConnConfig to avoid it being modified by the transport.
	return conn.NewMConnection(secretConn, *mt.mConfig), remotePubKey, nil
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func resolveIPs(resolver IPResolver, c net.Conn) ([]net.IP, error) {
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	addrs, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}

	ips := []net.IP{}

	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}

	return ips, nil
}

```

---

<a name="file-14"></a>

### File: `tcp/tcp_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 10 KB

```go
package tcp

import (
	"errors"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

// newMultiplexTransport returns a tcp connected multiplexed peer
// using the default MConnConfig. It's a convenience function used
// for testing.
func newMultiplexTransport(
	nodeKey nodekey.NodeKey,
) *MultiplexTransport {
	return NewMultiplexTransport(
		nodeKey, conn.DefaultMConnConfig(),
	)
}

func TestTransportMultiplex_ConnFilter(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			return errors.New("rejected")
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if e, ok := err.(ErrRejected); ok {
		if !e.IsFiltered() {
			t.Errorf("expected peer to be filtered, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplex_ConnFilterTimeout(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)
	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if _, ok := err.(ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout, got %v", err)
	}
}

func TestTransportMultiplex_MaxIncomingConnections(t *testing.T) {
	pv := ed25519.GenPrivKey()
	id := nodekey.PubKeyToID(pv.PubKey())
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: pv,
		},
	)

	MultiplexTransportMaxIncomingConnections(0)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	const maxIncomingConns = 2
	MultiplexTransportMaxIncomingConnections(maxIncomingConns)(mt)
	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	// Connect more peers than max
	for i := 0; i <= maxIncomingConns; i++ {
		errc := make(chan error)
		go testDialer(*laddr, errc)

		err = <-errc
		if i < maxIncomingConns {
			if err != nil {
				t.Errorf("dialer connection failed: %v", err)
			}
			_, _, err = mt.Accept()
			if err != nil {
				t.Errorf("connection failed: %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "i/o timeout") {
			// mt actually blocks forever on trying to accept a new peer into a full channel so
			// expect the dialer to encounter a timeout error. Calling mt.Accept will block until
			// mt is closed.
			t.Errorf("expected i/o timeout error, got %v", err)
		}
	}
}

func TestTransportMultiplex_AcceptMultiple(t *testing.T) {
	mt := testSetupMultiplexTransport(t)
	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	var (
		seed     = rand.New(rand.NewSource(time.Now().UnixNano()))
		nDialers = seed.Intn(64) + 64
		errc     = make(chan error, nDialers)
	)

	// Setup dialers.
	for i := 0; i < nDialers; i++ {
		go testDialer(*laddr, errc)
	}

	// Catch connection errors.
	for i := 0; i < nDialers; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}

	conns := []transport.Conn{}

	// Accept all connections.
	for i := 0; i < cap(errc); i++ {
		c, _, err := mt.Accept()
		if err != nil {
			t.Fatal(err)
		}

		conns = append(conns, c)
	}

	if have, want := len(conns), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func testDialer(dialAddr na.NetAddr, errc chan error) {
	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	_, err := dialer.Dial(dialAddr)
	if err != nil {
		errc <- err
		return
	}

	// Signal that the connection was established.
	errc <- nil
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		fastNodePV = ed25519.GenPrivKey()
		errc       = make(chan error)
		fastc      = make(chan struct{})
		slowc      = make(chan struct{})
		slowdonec  = make(chan struct{})
	)

	// Simulate slow Peer.
	go func() {
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		c, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(slowc)
		defer func() {
			close(slowdonec)
		}()

		// Make sure we switch to fast peer goroutine.
		runtime.Gosched()

		select {
		case <-fastc:
			// Fast peer connected.
		case <-time.After(200 * time.Millisecond):
			// We error if the fast peer didn't succeed.
			errc <- errors.New("fast peer timed out")
		}

		_, err = upgradeSecretConn(c, 200*time.Millisecond, ed25519.GenPrivKey())
		if err != nil {
			errc <- err
			return
		}
	}()

	// Simulate fast Peer.
	go func() {
		<-slowc

		dialer := newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: fastNodePV,
			},
		)
		dialer.SetLogger(log.TestingLogger())
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr)
		if err != nil {
			errc <- err
			return
		}

		close(fastc)
		<-slowdonec
		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Logf("connection failed: %v", err)
	}

	_, _, err := mt.Accept()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransportMultiplexDialRejectWrongID(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	wrongID := nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
	addr := na.New(wrongID, mt.listener.Addr())

	_, err := dialer.Dial(*addr)
	if err != nil {
		t.Logf("connection failed: %v", err)
		if e, ok := err.(ErrRejected); ok {
			if !e.IsAuthFailure() {
				t.Errorf("expected auth failure, got %v", e)
			}
		} else {
			t.Errorf("expected ErrRejected, got %v", err)
		}
	}
}

func TestTransportConnDuplicateIPFilter(t *testing.T) {
	filter := ConnDuplicateIPFilter()

	if err := filter(nil, &testTransportConn{}, nil); err != nil {
		t.Fatal(err)
	}

	var (
		c  = &testTransportConn{}
		cs = NewConnSet()
	)

	cs.Set(c, []net.IP{
		{10, 0, 10, 1},
		{10, 0, 10, 2},
		{10, 0, 10, 3},
	})

	if err := filter(cs, c, []net.IP{
		{10, 0, 10, 2},
	}); err == nil {
		t.Errorf("expected Peer to be rejected as duplicate")
	}
}

// create listener.
func testSetupMultiplexTransport(t *testing.T) *MultiplexTransport {
	t.Helper()

	var (
		pv = ed25519.GenPrivKey()
		id = nodekey.PubKeyToID(pv.PubKey())
		mt = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	mt.SetLogger(log.TestingLogger())

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	// give the listener some time to get ready
	time.Sleep(20 * time.Millisecond)

	return mt
}

type testTransportAddr struct{}

func (*testTransportAddr) Network() string { return "tcp" }
func (*testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (*testTransportConn) Close() error {
	return errors.New("close() not implemented")
}

func (*testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) Read(_ []byte) (int, error) {
	return -1, errors.New("read() not implemented")
}

func (*testTransportConn) SetDeadline(_ time.Time) error {
	return errors.New("setDeadline() not implemented")
}

func (*testTransportConn) SetReadDeadline(_ time.Time) error {
	return errors.New("setReadDeadline() not implemented")
}

func (*testTransportConn) SetWriteDeadline(_ time.Time) error {
	return errors.New("setWriteDeadline() not implemented")
}

func (*testTransportConn) Write(_ []byte) (int, error) {
	return -1, errors.New("write() not implemented")
}

```

---

<a name="file-15"></a>

### File: `transport.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 1 KB

```go
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
}

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}

```

---

## Summary

- **Total files processed:** 15
- **Total combined size:** 111 KB

## Breakdown of File Sizes by Type

- **go**: 111 KB
```

---

<a name="file-3"></a>

### File: `conn.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 3 KB

```go
package transport

import (
	"io"
	"net"
	"time"
)

// Conn is a multiplexed connection that can send and receive data
// on multiple streams.
type Conn interface {
	// OpenStream opens a new stream on the connection with an optional
	// description. If you're using tcp.MultiplexTransport, all streams must be
	// registered in advance.
	OpenStream(streamID byte, desc any) (Stream, error)

	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr

	// Close closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	Close(reason string) error

	// FlushAndClose flushes all the pending bytes and closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	FlushAndClose(reason string) error

	// ConnState returns basic details about the connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnState() ConnState

	// ErrorCh returns a channel that will receive errors from the connection.
	ErrorCh() <-chan error

	// HandshakeStream returns the stream to be used for the handshake.
	HandshakeStream() HandshakeStream
}

// Stream is the interface implemented by QUIC streams or multiplexed TCP connection.
type Stream interface {
	SendStream
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// Write writes data to the stream.
	// It blocks until data is sent or the stream is closed.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// TryWrite attempts to write data to the stream.
	// If the send queue is full, the error satisfies the WriteError interface, and Full() will be true.
	TryWrite(b []byte) (n int, err error)
}

// WriteError is returned by TryWrite when the send queue is full.
type WriteError interface {
	error
	Full() bool // Is the error due to the send queue being full?
}

// HandshakeStream is a stream that is used for the handshake.
type HandshakeStream interface {
	SetDeadline(t time.Time) error
	io.ReadWriter
}

```

---

<a name="file-4"></a>

### File: `conn_state.go`

*Modified:* 2025-02-08 17:49:58 • *Size:* 1 KB

```go
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

```

---

<a name="file-5"></a>

### File: `quic/errors.go`

*Modified:* 2025-02-08 17:55:08 • *Size:* 1 KB

```go
package quic

import "errors"

var (
	// ErrTransportNotListening is returned when trying to accept connections before listening
	ErrTransportNotListening = errors.New("transport not listening")

	// ErrTransportClosed is returned when the transport has been closed
	ErrTransportClosed = errors.New("transport closed")

	// ErrInvalidAddress is returned when an invalid address is provided
	ErrInvalidAddress = errors.New("invalid address")
)

```

---

<a name="file-6"></a>

### File: `quic/quic.go`

*Modified:* 2025-02-08 17:58:32 • *Size:* 3 KB

```go
package quic

import (
	"crypto/tls"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/quic-go/quic-go"
)

// Default configuration values
const (
	defaultMaxIncomingStreams = 100
	defaultKeepAlivePeriod    = 30 * time.Second
	defaultIdleTimeout        = 5 * time.Minute
	defaultHandshakeTimeout   = 10 * time.Second
)

// Transport implements the transport.BaseTransport interface using QUIC
type Transport struct {
	listener   *quic.Listener
	tlsConfig  *tls.Config
	quicConfig *quic.Config
	logger     log.Logger
	metrics    *transport.MetricsCollector

	// Connection management
	mtx         sync.RWMutex
	connections map[string]quic.Connection

	// Options
	maxStreams  int
	keepAlive   time.Duration
	idleTimeout time.Duration

	closed      chan struct{}
	isListening bool
}

// Options contains QUIC-specific configuration
type Options struct {
	TLSConfig          *tls.Config
	MaxIncomingStreams int
	KeepAlivePeriod    time.Duration
	IdleTimeout        time.Duration
}

// NewTransport creates a new QUIC transport instance
func NewTransport(opts *Options) (*Transport, error) {
	if opts == nil {
		opts = &Options{}
	}

	if opts.MaxIncomingStreams == 0 {
		opts.MaxIncomingStreams = defaultMaxIncomingStreams
	}
	if opts.KeepAlivePeriod == 0 {
		opts.KeepAlivePeriod = defaultKeepAlivePeriod
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = defaultIdleTimeout
	}

	quicConfig := &quic.Config{
		MaxIncomingStreams: int64(opts.MaxIncomingStreams),
		MaxIdleTimeout:     opts.IdleTimeout,
		KeepAlivePeriod:    opts.KeepAlivePeriod,
	}

	return &Transport{
		tlsConfig:   opts.TLSConfig,
		quicConfig:  quicConfig,
		connections: make(map[string]quic.Connection),
		maxStreams:  opts.MaxIncomingStreams,
		keepAlive:   opts.KeepAlivePeriod,
		idleTimeout: opts.IdleTimeout,
		closed:      make(chan struct{}),
		logger:      log.NewNopLogger(),
	}, nil
}

// Protocol implements transport.BaseTransport
func (*Transport) Protocol() transport.Protocol {
	return transport.ProtocolQUIC
}

// ... [rest of the implementation remains the same, just remove 'QUIC' prefix from references] ...

// GetMetrics implements transport.BaseTransport
func (*Transport) GetMetrics() *transport.Metrics {
	// TODO: Implement metrics collection using QUIC's connection monitoring
	return &transport.Metrics{}
}

```

---

<a name="file-7"></a>

### File: `quic/quic_test.go`

*Modified:* 2025-02-08 18:00:13 • *Size:* 5 KB

```go
package quic

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func generateTestTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-test"},
	}
}

func TestQUICTransportBasics(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	// Create transport with options
	opts := &Options{
		TLSConfig:          tlsConfig,
		MaxIncomingStreams: 10,
		KeepAlivePeriod:    time.Second,
		IdleTimeout:        time.Minute,
	}

	transport, err := NewTransport(opts)
	require.NoError(t, err)

	// Listen on a random port
	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	// Get the assigned address
	listener := *transport.listener
	addr := listener.Addr().String()

	// Try to connect
	ctx := context.Background()
	conn, err := transport.Dial(ctx, addr)
	require.NoError(t, err)

	// Write some data
	testData := []byte("hello world")
	n, err := conn.Write(testData)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)

	// Accept the connection on the server side
	serverConn, err := transport.Accept()
	require.NoError(t, err)

	// Read the data
	buf := make([]byte, len(testData))
	n, err = serverConn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)
	require.Equal(t, testData, buf)

	// Close connections
	require.NoError(t, conn.Close())
	require.NoError(t, serverConn.Close())
	require.NoError(t, transport.Close())
}

func TestQUICTransportConcurrent(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewQUICTransport(&QUICTransportOptions{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	listener := *transport.listener
	addr := listener.Addr().String()

	// Launch multiple concurrent connections
	const numConns = 10
	done := make(chan struct{})

	for i := 0; i < numConns; i++ {
		go func() {
			ctx := context.Background()
			conn, err := transport.Dial(ctx, addr)
			require.NoError(t, err)

			data := []byte("test data")
			_, err = conn.Write(data)
			require.NoError(t, err)

			require.NoError(t, conn.Close())
			done <- struct{}{}
		}()
	}

	// Accept and handle all connections
	for i := 0; i < numConns; i++ {
		conn, err := transport.Accept()
		require.NoError(t, err)

		go func(c net.Conn) {
			buf := make([]byte, 1024)
			_, err := c.Read(buf)
			require.NoError(t, err)
			require.NoError(t, c.Close())
		}(conn)
	}

	// Wait for all clients to finish
	for i := 0; i < numConns; i++ {
		<-done
	}

	require.NoError(t, transport.Close())
}

func TestQUICTransportTimeout(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewQUICTransport(&QUICTransportOptions{
		TLSConfig:   tlsConfig,
		IdleTimeout: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	listener := *transport.listener
	addr := listener.Addr().String()

	// Connect and let the connection idle
	ctx := context.Background()
	conn, err := transport.Dial(ctx, addr)
	require.NoError(t, err)

	// Accept the connection
	serverConn, err := transport.Accept()
	require.NoError(t, err)

	// Wait for idle timeout
	time.Sleep(200 * time.Millisecond)

	// Verify connections are closed
	_, err = conn.Write([]byte("test"))
	require.Error(t, err)

	_, err = serverConn.Write([]byte("test"))
	require.Error(t, err)

	require.NoError(t, transport.Close())
}

func TestQUICTransportError(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewQUICTransport(&QUICTransportOptions{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	// Try to accept before listening
	_, err = transport.Accept()
	require.Equal(t, ErrTransportNotListening, err)

	// Try to listen on invalid address
	err = transport.Listen("invalid-addr")
	require.Error(t, err)

	// Try to dial invalid address
	ctx := context.Background()
	_, err = transport.Dial(ctx, "invalid-addr")
	require.Error(t, err)

	require.NoError(t, transport.Close())
}

func TestQUICTransportMetrics(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewQUICTransport(&QUICTransportOptions{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	// Verify initial metrics
	metrics := transport.GetMetrics()
	require.NotNil(t, metrics)

	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	listener := *transport.listener
	addr := listener.Addr().String()

	// Create some traffic
	ctx := context.Background()
	conn, err := transport.Dial(ctx, addr)
	require.NoError(t, err)

	serverConn, err := transport.Accept()
	require.NoError(t, err)

	data := make([]byte, 1024)
	for i := 0; i < 10; i++ {
		_, err = conn.Write(data)
		require.NoError(t, err)

		_, err = serverConn.Read(make([]byte, len(data)))
		require.NoError(t, err)
	}

	// Verify updated metrics
	metrics = transport.GetMetrics()
	require.NotNil(t, metrics)

	require.NoError(t, conn.Close())
	require.NoError(t, serverConn.Close())
	require.NoError(t, transport.Close())
}

```

---

<a name="file-8"></a>

### File: `quic/wrapper.go`

*Modified:* 2025-02-08 17:58:56 • *Size:* 2 KB

```go
package quic

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// Conn wraps a QUIC connection and stream to implement net.Conn
type Conn struct {
	conn   quic.Connection
	stream quic.Stream
}

// Read implements net.Conn
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write implements net.Conn
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close implements net.Conn
func (c *Conn) Close() error {
	// Close both the stream and the connection
	if err := c.stream.Close(); err != nil {
		return err
	}
	return c.conn.CloseWithError(0, "connection closed")
}

// LocalAddr implements net.Conn
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr implements net.Conn
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements net.Conn
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline implements net.Conn
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// GetConnectionState returns the underlying QUIC connection state
func (c *Conn) GetConnectionState() quic.ConnectionState {
	return c.conn.ConnectionState()
}

```

---

<a name="file-9"></a>

### File: `tcp/conn/connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 26 KB

```go
package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	flow "github.com/cometbft/cometbft/internal/flowrate"
	"github.com/cometbft/cometbft/internal/timer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/transport"
)

const (
	defaultMaxPacketMsgPayloadSize = 1024

	numBatchPacketMsgs = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	updateStats        = 2 * time.Second

	// some of these defaults are written in the user config
	// flushThrottle, sendRate, recvRate
	// TODO: remove values present in config.
	defaultFlushThrottle = 10 * time.Millisecond

	defaultSendRate     = int64(512000) // 500KB/s
	defaultRecvRate     = int64(512000) // 500KB/s
	defaultPingInterval = 60 * time.Second
	defaultPongTimeout  = 45 * time.Second
)

// OnReceiveFn is a callback func, which is called by the MConnection when a
// new message is received.
type OnReceiveFn = func(byte, []byte)

// MConnection is a multiplexed connection.
//
// __multiplex__ *noun* a system or signal involving simultaneous transmission
// of several messages along a single channel of communication.
//
// Each connection handles message transmission on multiple abstract
// communication streams. Each stream has a globally unique byte id. The byte
// id and the relative priorities of each stream are configured upon
// initialization of the connection.
//
// To open a stream, call OpenStream with the stream id. Remember that the
// stream id must be globally unique.
//
// Connection errors are communicated through the ErrorCh channel.
//
// Connection can be closed either by calling Close or FlushAndClose. If you
// need to flush all pending messages before closing the connection, call
// FlushAndClose. Otherwise, call Close.
type MConnection struct {
	service.BaseService

	conn          net.Conn
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flow.Monitor
	recvMonitor   *flow.Monitor
	send          chan struct{}
	pong          chan struct{}
	errorCh       chan error
	config        MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	quitSendRoutine chan struct{}
	doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	flushTimer *timer.ThrottleTimer // flush writes as necessary but throttled.
	pingTimer  *time.Ticker         // send pings periodically

	// close conn if pong is not received in pongTimeout
	pongTimer     *time.Timer
	pongTimeoutCh chan bool // true - timeout, false - peer sent pong

	chStatsTimer *time.Ticker // update channel stats periodically

	created time.Time // time of creation

	_maxPacketMsgSize int

	// streamID -> channel
	channelsIdx map[byte]*stream

	// A map which stores the received messages. Used in tests.
	msgsByStreamIDMap map[byte]chan []byte

	onReceiveFn OnReceiveFn
}

var _ transport.Conn = (*MConnection)(nil)

// MConnConfig is a MConnection configuration.
type MConnConfig struct {
	SendRate int64 `mapstructure:"send_rate"`
	RecvRate int64 `mapstructure:"recv_rate"`

	// Maximum payload size
	MaxPacketMsgPayloadSize int `mapstructure:"max_packet_msg_payload_size"`

	// Interval to flush writes (throttled)
	FlushThrottle time.Duration `mapstructure:"flush_throttle"`

	// Interval to send pings
	PingInterval time.Duration `mapstructure:"ping_interval"`

	// Maximum wait time for pongs
	PongTimeout time.Duration `mapstructure:"pong_timeout"`

	// Fuzz connection
	TestFuzz       bool                   `mapstructure:"test_fuzz"`
	TestFuzzConfig *config.FuzzConnConfig `mapstructure:"test_fuzz_config"`
}

// DefaultMConnConfig returns the default config.
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:                defaultSendRate,
		RecvRate:                defaultRecvRate,
		MaxPacketMsgPayloadSize: defaultMaxPacketMsgPayloadSize,
		FlushThrottle:           defaultFlushThrottle,
		PingInterval:            defaultPingInterval,
		PongTimeout:             defaultPongTimeout,
	}
}

// NewMConnection wraps net.Conn and creates multiplex connection.
func NewMConnection(conn net.Conn, config MConnConfig) *MConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}

	mconn := &MConnection{
		conn:              conn,
		bufConnReader:     bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter:     bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:       flow.New(0, 0),
		recvMonitor:       flow.New(0, 0),
		send:              make(chan struct{}, 1),
		pong:              make(chan struct{}, 1),
		errorCh:           make(chan error, 1),
		config:            config,
		created:           time.Now(),
		channelsIdx:       make(map[byte]*stream),
		msgsByStreamIDMap: make(map[byte]chan []byte),
	}

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

// OnReceive sets the callback function to be executed each time we read a message.
func (c *MConnection) OnReceive(fn OnReceiveFn) {
	c.onReceiveFn = fn
}

func (c *MConnection) SetLogger(l log.Logger) {
	c.BaseService.SetLogger(l)
}

// OnStart implements BaseService.
func (c *MConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

func (c *MConnection) Conn() net.Conn {
	return c.conn
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
func (c *MConnection) stopServices() (alreadyStopped bool) {
	select {
	case <-c.quitSendRoutine:
		// already quit
		return true
	default:
	}

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	close(c.quitSendRoutine)
	return false
}

// ErrorCh returns a channel that will receive errors from the connection.
func (c *MConnection) ErrorCh() <-chan error {
	return c.errorCh
}

func (c *MConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *MConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// OpenStream opens a new stream on the connection. Remember that the
// stream id must be globally unique.
//
// Panics if the connection is already running (i.e., all streams
// must be registered in advance).
func (c *MConnection) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	if c.IsRunning() {
		panic("MConnection is already running. Please register all streams in advance")
	}

	c.Logger.Debug("Opening stream", "streamID", streamID, "desc", desc)

	if _, ok := c.channelsIdx[streamID]; ok {
		return nil, fmt.Errorf("stream %X already exists", streamID)
	}

	d := StreamDescriptor{
		ID:       streamID,
		Priority: 1,
	}
	if desc, ok := desc.(StreamDescriptor); ok {
		d = desc
	}
	c.channelsIdx[streamID] = newChannel(c, d)
	c.channelsIdx[streamID].SetLogger(c.Logger.With("streamID", streamID))
	// Allocate some buffer, otherwise CI tests will fail.
	c.msgsByStreamIDMap[streamID] = make(chan []byte, 5)

	return &MConnectionStream{conn: c, streamID: streamID}, nil
}

// HandshakeStream returns the underlying net.Conn connection.
func (c *MConnection) HandshakeStream() transport.HandshakeStream {
	return c.conn
}

// Close closes the connection. It flushes all pending writes before closing.
func (c *MConnection) Close(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	return c.conn.Close()
}

func (c *MConnection) FlushAndClose(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	// flush all pending writes
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine
		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		w := protoio.NewDelimitedWriter(c.bufConnWriter)
		eof := c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		for !eof {
			eof = c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		}
		_ = c.flush()
	}

	return c.conn.Close()
}

func (c *MConnection) ConnState() (state transport.ConnState) {
	state.ConnectedFor = time.Since(c.created)
	state.SendRateLimiterDelay = c.sendMonitor.Status().SleepTime
	state.RecvRateLimiterDelay = c.recvMonitor.Status().SleepTime
	state.StreamStates = make(map[byte]transport.StreamState)

	for streamID, channel := range c.channelsIdx {
		state.StreamStates[streamID] = transport.StreamState{
			SendQueueSize:     channel.loadSendQueueSize(),
			SendQueueCapacity: cap(channel.sendQueue),
		}
	}

	return state
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() error {
	return c.bufConnWriter.Flush()
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.Close(fmt.Sprintf("recovered from panic: %v", r))
	}
}

// thread-safe.
func (c *MConnection) sendBytes(chID byte, msgBytes []byte, blocking bool) error {
	if !c.IsRunning() {
		return nil
	}

	// Uncomment in you need to see raw bytes.
	// c.Logger.Debug("Send",
	// 	"streamID", chID,
	// 	"msgBytes", log.NewLazySprintf("%X", msgBytes),
	// 	"timeout", timeout)

	channel, ok := c.channelsIdx[chID]
	if !ok {
		panic(fmt.Sprintf("Unknown channel %X. Forgot to register?", chID))
	}
	if err := channel.sendBytes(msgBytes, blocking); err != nil {
		return err
	}

	// Wake up sendRoutine if necessary
	select {
	case c.send <- struct{}{}:
	default:
	}
	return nil
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
//
// thread-safe.
func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Unknown channel %X", chID))
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

	protoWriter := protoio.NewDelimitedWriter(c.bufConnWriter)

FOR_LOOP:
	for {
		var _n int
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.chStatsTimer.C:
			for _, channel := range c.channelsIdx {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPing", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
				select {
				case c.pongTimeoutCh <- true:
				default:
				}
			})
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPong", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs(protoWriter)
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomePacketMsgs(w protoio.Writer) bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, c.config.SendRate, true)

	// Now send some PacketMsgs.
	return c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendBatchPacketMsgs(w protoio.Writer, batchSize int) bool {
	// Send a batch of PacketMsgs.
	totalBytesWritten := 0
	defer func() {
		if totalBytesWritten > 0 {
			c.sendMonitor.Update(totalBytesWritten)
		}
	}()
	for i := 0; i < batchSize; i++ {
		channel := c.selectChannel()
		// nothing to send across any channel.
		if channel == nil {
			return true
		}
		bytesWritten, err := c.sendPacketMsgOnChannel(w, channel)
		if err {
			return true
		}
		totalBytesWritten += bytesWritten
	}
	return false
}

// selects a channel to gossip our next message on.
// TODO: Make "batchChannelToGossipOn", so we can do our proto marshaling overheads in parallel,
// and we can avoid re-checking for `isSendPending`.
// We can easily mock the recentlySent differences for the batch choosing.
func (c *MConnection) selectChannel() *stream {
	// Choose a channel to create a PacketMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *stream
	for _, channel := range c.channelsIdx {
		// If nothing to send, skip this channel
		// TODO: Skip continually looking for isSendPending on channels we've already skipped in this batch-send.
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		// TODO: RecentlySent right now is bytes. This should be refactored to num messages to fix
		// gossip prioritization bugs.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}
	return leastChannel
}

// returns (num_bytes_written, error_occurred).
func (c *MConnection) sendPacketMsgOnChannel(w protoio.Writer, sendChannel *stream) (int, bool) {
	// Make & send a PacketMsg from this channel
	n, err := sendChannel.writePacketMsgTo(w)
	if err != nil {
		c.Logger.Error("Failed to write PacketMsg", "err", err)
		c.Close(err.Error())
		return n, true
	}
	// TODO: Change this to only add flush signals at the start and end of the batch.
	c.flushTimer.Set()
	return n, false
}

// recvRoutine reads PacketMsgs and reconstructs the message using the
// channels' "recving" buffer. After a whole message has been assembled, it's
// pushed to an internal queue, which is accessible via Read. Blocks depending
// on how the connection is throttled. Otherwise, it never blocks.
func (c *MConnection) recvRoutine() {
	defer c._recover()

	protoReader := protoio.NewDelimitedReader(c.bufConnReader, c._maxPacketMsgSize)

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(cmtmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet tmp2p.Packet

		_n, err := protoReader.ReadMsg(&packet)
		c.recvMonitor.Update(_n)
		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is expected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if errors.Is(err, io.EOF) {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)")
				} else {
					c.Logger.Debug("Connection failed @ recvRoutine (reading byte)", "err", err)
				}
				c.Close(err.Error())
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.Sum.(type) {
		case *tmp2p.Packet_PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketPong:
			c.Logger.Debug("Receive Pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketMsg:
			channelID := byte(pkt.PacketMsg.ChannelID)
			channel, ok := c.channelsIdx[channelID]
			if !ok || pkt.PacketMsg.ChannelID < 0 || pkt.PacketMsg.ChannelID > math.MaxUint8 {
				err := fmt.Errorf("unknown channel %X", pkt.PacketMsg.ChannelID)
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}

			msgBytes, err := channel.recvPacketMsg(*pkt.PacketMsg)
			if err != nil {
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}
			if msgBytes != nil {
				// Uncomment in you need to see raw bytes.
				// c.Logger.Debug("Received", "streamID", channelID, "msgBytes", log.NewLazySprintf("%X", msgBytes))
				if c.onReceiveFn != nil {
					c.onReceiveFn(channelID, msgBytes)
				} else {
					bz := make([]byte, len(msgBytes))
					copy(bz, msgBytes)
					c.msgsByStreamIDMap[channelID] <- bz
				}
			}
		default:
			err := fmt.Errorf("unknown message type %v", reflect.TypeOf(packet))
			c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	close(c.pong)
}

// Used in tests.
func (c *MConnection) readBytes(streamID byte, b []byte, timeout time.Duration) (n int, err error) {
	select {
	case msgBytes := <-c.msgsByStreamIDMap[streamID]:
		n = copy(b, msgBytes)
		if n < len(msgBytes) {
			err = errors.New("short buffer")
			return 0, err
		}
		return n, nil
	case <-time.After(timeout):
		return 0, errors.New("read timeout")
	}
}

// not goroutine-safe.
func (c *MConnection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize returns a maximum size of PacketMsg.
func (c *MConnection) maxPacketMsgSize() int {
	bz, err := proto.Marshal(mustWrapPacket(&tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, c.config.MaxPacketMsgPayloadSize),
	}))
	if err != nil {
		panic(err)
	}
	return len(bz)
}

// -----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type stream struct {
	conn          *MConnection
	desc          StreamDescriptor
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	recentlySent  int64 // exponential moving average

	nextPacketMsg           *tmp2p.PacketMsg
	nextP2pWrapperPacketMsg *tmp2p.Packet_PacketMsg
	nextPacket              *tmp2p.Packet

	maxPacketMsgPayloadSize int

	Logger log.Logger
}

func newChannel(conn *MConnection, desc StreamDescriptor) *stream {
	desc = desc.FillDefaults()
	if desc.Priority <= 0 {
		panic("Channel default priority must be a positive integer")
	}
	return &stream{
		conn:                    conn,
		desc:                    desc,
		sendQueue:               make(chan []byte, desc.SendQueueCapacity),
		recving:                 make([]byte, 0, desc.RecvBufferCapacity),
		nextPacketMsg:           &tmp2p.PacketMsg{ChannelID: int32(desc.ID)},
		nextP2pWrapperPacketMsg: &tmp2p.Packet_PacketMsg{},
		nextPacket:              &tmp2p.Packet{},
		maxPacketMsgPayloadSize: conn.config.MaxPacketMsgPayloadSize,
	}
}

func (ch *stream) SetLogger(l log.Logger) {
	ch.Logger = l
}

// Queues message to send to this channel. Blocks if blocking is true.
// thread-safe.
func (ch *stream) sendBytes(bytes []byte, blocking bool) error {
	if blocking {
		select {
		case ch.sendQueue <- bytes:
			atomic.AddInt32(&ch.sendQueueSize, 1)
			return nil
		case <-ch.conn.Quit():
			return nil
		}
	}

	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return nil
	default:
		return ErrWriteQueueFull{}
	case <-ch.conn.Quit():
		return nil
	}
}

// Goroutine-safe.
func (ch *stream) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *stream) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any PacketMsgs are pending to be sent.
// Call before calling updateNextPacket
// Goroutine-safe.
func (ch *stream) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Updates the nextPacket proto message for us to send.
// Not goroutine-safe.
func (ch *stream) updateNextPacket() {
	maxSize := ch.maxPacketMsgPayloadSize
	if len(ch.sending) <= maxSize {
		ch.nextPacketMsg.Data = ch.sending
		ch.nextPacketMsg.EOF = true
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		ch.nextPacketMsg.Data = ch.sending[:maxSize]
		ch.nextPacketMsg.EOF = false
		ch.sending = ch.sending[maxSize:]
	}

	ch.nextP2pWrapperPacketMsg.PacketMsg = ch.nextPacketMsg
	ch.nextPacket.Sum = ch.nextP2pWrapperPacketMsg
}

// Writes next PacketMsg to w and updates c.recentlySent.
// Not goroutine-safe.
func (ch *stream) writePacketMsgTo(w protoio.Writer) (n int, err error) {
	ch.updateNextPacket()
	n, err = w.WriteMsg(ch.nextPacket)
	if err != nil {
		err = ErrPacketWrite{Source: err}
	}

	atomic.AddInt64(&ch.recentlySent, int64(n))
	return n, err
}

// Handles incoming PacketMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe.
func (ch *stream) recvPacketMsg(packet tmp2p.PacketMsg) ([]byte, error) {
	recvCap, recvReceived := ch.desc.RecvMessageCapacity, len(ch.recving)+len(packet.Data)
	if recvCap < recvReceived {
		return nil, ErrPacketTooBig{Max: recvCap, Received: recvReceived}
	}

	ch.recving = append(ch.recving, packet.Data...)
	if packet.EOF {
		msgBytes := ch.recving

		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// thread-safe.
func (ch *stream) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

// ----------------------------------------
// Packet

// mustWrapPacket takes a packet kind (oneof) and wraps it in a tmp2p.Packet message.
func mustWrapPacket(pb proto.Message) *tmp2p.Packet {
	msg := &tmp2p.Packet{}
	mustWrapPacketInto(pb, msg)
	return msg
}

func mustWrapPacketInto(pb proto.Message, dst *tmp2p.Packet) {
	switch pb := pb.(type) {
	case *tmp2p.PacketPing:
		dst.Sum = &tmp2p.Packet_PacketPing{
			PacketPing: pb,
		}
	case *tmp2p.PacketPong:
		dst.Sum = &tmp2p.Packet_PacketPong{
			PacketPong: pb,
		}
	case *tmp2p.PacketMsg:
		dst.Sum = &tmp2p.Packet_PacketMsg{
			PacketMsg: pb,
		}
	default:
		panic(fmt.Errorf("unknown packet type %T", pb))
	}
}

```

---

<a name="file-10"></a>

### File: `tcp/conn/connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	pbtypes "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
)

const (
	maxPingPongPacketSize = 1024 // bytes
	testStreamID          = 0x01
)

func createMConnectionWithSingleStream(t *testing.T, conn net.Conn) (*MConnection, *MConnectionStream) {
	t.Helper()

	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	c := NewMConnection(conn, cfg)
	c.SetLogger(log.TestingLogger())

	stream, err := c.OpenStream(testStreamID, nil)
	require.NoError(t, err)

	return c, stream.(*MConnectionStream)
}

func TestMConnection_Flush(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	clientConn, clientStream := createMConnectionWithSingleStream(t, client)
	err := clientConn.Start()
	require.NoError(t, err)

	msg := []byte("abc")
	n, err := clientStream.Write(msg)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	// start the reader in a new routine, so we can flush
	errCh := make(chan error)
	go func() {
		buf := make([]byte, 100) // msg + ping
		_, err := server.Read(buf)
		errCh <- err
	}()

	// stop the conn - it should flush all conns
	err = clientConn.FlushAndClose("test")
	require.NoError(t, err)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Error reading from server: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for msgs to be read")
	}
}

func TestMConnection_StreamWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, clientStream := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	msg := []byte("Ant-Man")
	_, err = clientStream.Write(msg)
	require.NoError(t, err)
	// NOTE: subsequent writes could pass because we are reading from
	// the send queue in a separate goroutine.
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
	assert.True(t, mconn.CanSend(testStreamID))

	msg = []byte("Spider-Man")
	require.NoError(t, err)
	_, err = clientStream.TryWrite(msg)
	require.NoError(t, err)
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
}

func TestMConnection_StreamReadWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn1, stream1 := createMConnectionWithSingleStream(t, client)
	err := mconn1.Start()
	require.NoError(t, err)
	defer mconn1.Close("normal")

	mconn2, stream2 := createMConnectionWithSingleStream(t, server)
	err = mconn2.Start()
	require.NoError(t, err)
	defer mconn2.Close("normal")

	// => write
	msg := []byte("Cyclops")
	_, err = stream1.Write(msg)
	require.NoError(t, err)

	// => read
	buf := make([]byte, len(msg))
	n, err := stream2.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.Equal(t, msg, buf)
}

func TestMConnectionStatus(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	state := mconn.ConnState()
	assert.NotNil(t, state)
	assert.Zero(t, state.StreamStates[testStreamID].SendQueueSize)
}

func TestMConnection_PongTimeoutResultsInError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		// read ping
		var pkt tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 200*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
	case <-time.After(pongTimerExpired):
		t.Fatalf("Expected to receive error after %v", pongTimerExpired)
	}
}

func TestMConnection_MultiplePongsInTheBeginning(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pongs in a row (abuse)
	protoWriter := protoio.NewDelimitedWriter(server)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (one byte)
		var packet tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&packet)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 20*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_MultiplePings(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pings in a row (abuse)
	// see https://github.com/tendermint/tendermint/issues/1190
	protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
	protoWriter := protoio.NewDelimitedWriter(server)
	var pkt tmp2p.Packet

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	assert.True(t, mconn.IsRunning())
}

func TestMConnection_PingPongs(t *testing.T) {
	// check that we are not leaking any go-routines
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	server, client := net.Pipe()

	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
		protoWriter := protoio.NewDelimitedWriter(server)
		var pkt tmp2p.PacketPing

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)

		time.Sleep(mconn.config.PingInterval)

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing
	<-serverGotPing

	pongTimerExpired := (mconn.config.PongTimeout + 20*time.Millisecond) * 2
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(2 * pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_StopsAndReturnsError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	if err := client.Close(); err != nil {
		t.Error(err)
	}

	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
		assert.False(t, mconn.IsRunning())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not receive error in 500ms")
	}
}

//nolint:unparam
func newClientAndServerConnsForReadErrors(t *testing.T) (*MConnection, *MConnectionStream, *MConnection, *MConnectionStream) {
	t.Helper()
	server, client := net.Pipe()

	// create client conn with two channels
	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	mconnClient := NewMConnection(client, cfg)
	clientStream, err := mconnClient.OpenStream(testStreamID, StreamDescriptor{ID: testStreamID, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	// create another channel
	_, err = mconnClient.OpenStream(0x02, StreamDescriptor{ID: 0x02, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	mconnClient.SetLogger(log.TestingLogger().With("module", "client"))
	err = mconnClient.Start()
	require.NoError(t, err)

	// create server conn with 1 channel
	// it fires on chOnErr when there's an error
	serverLogger := log.TestingLogger().With("module", "server")
	mconnServer, serverStream := createMConnectionWithSingleStream(t, server)
	mconnServer.SetLogger(serverLogger)
	err = mconnServer.Start()
	require.NoError(t, err)

	return mconnClient, clientStream.(*MConnectionStream), mconnServer, serverStream
}

func assertBytes(t *testing.T, s *MConnectionStream, want []byte) {
	t.Helper()

	buf := make([]byte, len(want))
	n, err := s.Read(buf)
	require.NoError(t, err)
	if assert.Equal(t, len(want), n) {
		assert.Equal(t, want, buf)
	}
}

func gotError(ch <-chan error) bool {
	after := time.After(time.Second * 5)
	select {
	case <-ch:
		return true
	case <-after:
		return false
	}
}

func TestMConnection_ReadErrorBadEncoding(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send badly encoded data
	client := mconnClient.conn
	_, err := client.Write([]byte{1, 2, 3, 4, 5})
	require.NoError(t, err)

	assert.True(t, gotError(mconnServer.ErrorCh()), "badly encoded msgPacket")
}

// func TestMConnection_ReadErrorUnknownChannel(t *testing.T) {
// 	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
// 	defer mconnClient.Close("normal")
// 	defer mconnServer.Close("normal")

// 	msg := []byte("Ant-Man")

// 	// send msg that has unknown channel
// 	client := mconnClient.conn
// 	protoWriter := protoio.NewDelimitedWriter(client)
// 	packet := tmp2p.PacketMsg{
// 		ChannelID: 0x03,
// 		EOF:       true,
// 		Data:      msg,
// 	}
// 	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
// 	require.NoError(t, err)

// 	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown channel")
// }

func TestMConnection_ReadErrorLongMessage(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	msg := make([]byte, mconnClient.config.MaxPacketMsgPayloadSize)
	packet := tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      msg,
	}

	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, msg)

	// send msg that's too long
	packet = tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, mconnClient.config.MaxPacketMsgPayloadSize+100),
	}

	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.Error(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "msg too long")
}

func TestMConnection_ReadErrorUnknownMsgType(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send msg with unknown msg type
	_, err := protoio.NewDelimitedWriter(mconnClient.conn).WriteMsg(&pbtypes.Header{ChainID: "x"})
	require.NoError(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown msg type")
}

//nolint:lll //ignore line length for tests
func TestConnVectors(t *testing.T) {
	testCases := []struct {
		testName string
		msg      proto.Message
		expBytes string
	}{
		{"PacketPing", &tmp2p.PacketPing{}, "0a00"},
		{"PacketPong", &tmp2p.PacketPong{}, "1200"},
		{"PacketMsg", &tmp2p.PacketMsg{ChannelID: 1, EOF: false, Data: []byte("data transmitted over the wire")}, "1a2208011a1e64617461207472616e736d6974746564206f766572207468652077697265"},
	}

	for _, tc := range testCases {
		pm := mustWrapPacket(tc.msg)
		bz, err := pm.Marshal()
		require.NoError(t, err, tc.testName)

		require.Equal(t, tc.expBytes, hex.EncodeToString(bz), tc.testName)
	}
}

func TestMConnection_ChannelOverflow(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	packet := tmp2p.PacketMsg{
		ChannelID: testStreamID,
		EOF:       true,
		Data:      []byte(`42`),
	}
	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, []byte(`42`))

	// channel ID that's too large
	packet.ChannelID = int32(1025)
	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
}

```

---

<a name="file-11"></a>

### File: `tcp/conn/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/p2p/transport"
)

var (
	ErrInvalidSecretConnKeySend = errors.New("send invalid secret connection key")
	ErrInvalidSecretConnKeyRecv = errors.New("invalid receive SecretConnection Key")
	ErrChallengeVerification    = errors.New("challenge verification failed")

	// ErrTimeout is returned when a read or write operation times out.
	ErrTimeout = errors.New("read/write timeout")
)

// ErrWriteQueueFull is returned when the write queue is full.
type ErrWriteQueueFull struct{}

var _ transport.WriteError = ErrWriteQueueFull{}

func (ErrWriteQueueFull) Error() string {
	return "write queue is full"
}

func (ErrWriteQueueFull) Full() bool {
	return true
}

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	Source error
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet message: %v", e.Source)
}

func (e ErrPacketWrite) Unwrap() error {
	return e.Source
}

type ErrUnexpectedPubKeyType struct {
	Expected string
	Got      string
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptFrame struct {
	Source error
}

func (e ErrDecryptFrame) Error() string {
	return fmt.Sprintf("SecretConnection: failed to decrypt the frame: %v", e.Source)
}

func (e ErrDecryptFrame) Unwrap() error {
	return e.Source
}

type ErrPacketTooBig struct {
	Received int
	Max      int
}

func (e ErrPacketTooBig) Error() string {
	return fmt.Sprintf("received message exceeds available capacity (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkTooBig struct {
	Received int
	Max      int
}

func (e ErrChunkTooBig) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}

```

---

<a name="file-12"></a>

### File: `tcp/conn/evil_secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 8 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
)

type buffer struct {
	next bytes.Buffer
}

func (b *buffer) Read(data []byte) (n int, err error) {
	return b.next.Read(data)
}

func (b *buffer) Write(data []byte) (n int, err error) {
	return b.next.Write(data)
}

func (b *buffer) Bytes() []byte {
	return b.next.Bytes()
}

func (*buffer) Close() error {
	return nil
}

type evilConn struct {
	secretConn *SecretConnection
	buffer     *buffer

	locEphPub  *[32]byte
	locEphPriv *[32]byte
	remEphPub  *[32]byte
	privKey    crypto.PrivKey

	readStep   int
	writeStep  int
	readOffset int

	shareEphKey        bool
	badEphKey          bool
	shareAuthSignature bool
	badAuthSignature   bool
}

func newEvilConn(shareEphKey, badEphKey, shareAuthSignature, badAuthSignature bool) *evilConn {
	privKey := ed25519.GenPrivKey()
	locEphPub, locEphPriv := genEphKeys()
	var rep [32]byte
	c := &evilConn{
		locEphPub:  locEphPub,
		locEphPriv: locEphPriv,
		remEphPub:  &rep,
		privKey:    privKey,

		shareEphKey:        shareEphKey,
		badEphKey:          badEphKey,
		shareAuthSignature: shareAuthSignature,
		badAuthSignature:   badAuthSignature,
	}

	return c
}

func (c *evilConn) Read(data []byte) (n int, err error) {
	if !c.shareEphKey {
		return 0, io.EOF
	}

	switch c.readStep {
	case 0:
		if !c.badEphKey {
			lc := *c.locEphPub
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				panic(err)
			}
			copy(data, bz[c.readOffset:])
			n = len(data)
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("drop users;")})
			if err != nil {
				panic(err)
			}
			copy(data, bz)
			n = len(data)
		}
		c.readOffset += n

		if n >= 32 {
			c.readOffset = 0
			c.readStep = 1
			if !c.shareAuthSignature {
				c.readStep = 2
			}
		}

		return n, nil
	case 1:
		signature := c.signChallenge()
		if !c.badAuthSignature {
			pkpb, err := cryptoenc.PubKeyToProto(c.privKey.PubKey())
			if err != nil {
				panic(err)
			}
			bz, err := protoio.MarshalDelimited(&tmp2p.AuthSigMessage{PubKey: pkpb, Sig: signature})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes()[c.readOffset:])
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("select * from users;")})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes())
		}
		c.readOffset += len(data)
		return n, nil
	default:
		return 0, io.EOF
	}
}

func (c *evilConn) Write(data []byte) (n int, err error) {
	switch c.writeStep {
	case 0:
		var (
			bytes     gogotypes.BytesValue
			remEphPub [32]byte
		)
		err := protoio.UnmarshalDelimited(data, &bytes)
		if err != nil {
			panic(err)
		}
		copy(remEphPub[:], bytes.Value)
		c.remEphPub = &remEphPub
		c.writeStep = 1
		if !c.shareAuthSignature {
			c.writeStep = 2
		}
		return len(data), nil
	case 1:
		// Signature is not needed, therefore skipped.
		return len(data), nil
	default:
		return 0, io.EOF
	}
}

func (*evilConn) Close() error {
	return nil
}

func (c *evilConn) signChallenge() []byte {
	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(c.locEphPub, c.remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least, lexicographically
	// sorted.
	locIsLeast := bytes.Equal(c.locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(c.remEphPub, c.locEphPriv)
	if err != nil {
		panic(err)
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via HKDF-SHA2
	// on the transcript state (which itself also uses HKDF-SHA2 to derive a key
	// from the dhSecret).
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		panic(errors.New("invalid send SecretConnection Key"))
	}
	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		panic(errors.New("invalid receive SecretConnection Key"))
	}

	b := &buffer{}
	c.secretConn = &SecretConnection{
		conn:            b,
		connWriter:      bufio.NewWriterSize(b, defaultWriteBufferSize),
		connReader:      b,
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
	}
	c.buffer = b

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, c.privKey)
	if err != nil {
		panic(err)
	}

	return locSignature
}

// TestMakeSecretConnection creates an evil connection and tests that
// MakeSecretConnection errors at different stages.
func TestMakeSecretConnection(t *testing.T) {
	testCases := []struct {
		name       string
		conn       *evilConn
		checkError func(error) bool // Function to check if the error matches the expectation
	}{
		{"refuse to share ethimeral key", newEvilConn(false, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad ethimeral key", newEvilConn(true, true, false, false), func(err error) bool { return assert.Contains(t, err.Error(), "wrong wireType") }},
		{"refuse to share auth signature", newEvilConn(true, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad auth signature", newEvilConn(true, false, true, true), func(err error) bool { return errors.As(err, &ErrDecryptFrame{}) }},
		// fails with the introduction of changes PR #3419
		// {"all good", newEvilConn(true, false, true, false), func(err error) bool { return err == nil }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privKey := ed25519.GenPrivKey()
			_, err := MakeSecretConnection(tc.conn, privKey)
			if tc.checkError != nil {
				assert.True(t, tc.checkError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

```

---

<a name="file-13"></a>

### File: `tcp/conn/secret_connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/internal/async"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// 4 + 1024 == 1028 total frame size.
const (
	dataLenSize      = 4
	dataMaxSize      = 1024
	totalFrameSize   = dataMaxSize + dataLenSize
	aeadSizeOverhead = 16 // overhead of poly 1305 authentication tag
	aeadKeySize      = chacha20poly1305.KeySize
	aeadNonceSize    = chacha20poly1305.NonceSize

	labelEphemeralLowerPublicKey = "EPHEMERAL_LOWER_PUBLIC_KEY"
	labelEphemeralUpperPublicKey = "EPHEMERAL_UPPER_PUBLIC_KEY"
	labelDHSecret                = "DH_SECRET"
	labelSecretConnectionMac     = "SECRET_CONNECTION_MAC"

	defaultWriteBufferSize = 128 * 1024
	// try to read the biggest logical packet we can get, in one read.
	// biggest logical packet is encoding_overhead(64kb).
	defaultReadBufferSize = 65 * 1024
)

var (
	ErrSmallOrderRemotePubKey    = errors.New("detected low order point from remote peer")
	secretConnKeyAndChallengeGen = []byte("TENDERMINT_SECRET_CONNECTION_KEY_AND_CHALLENGE_GEN")
)

// SecretConnection implements net.Conn.
// It is an implementation of the STS protocol.
// For more details regarding this implementation of the STS protocol, please refer to:
// https://github.com/cometbft/cometbft/blob/main/spec/p2p/legacy-docs/peer.md#authenticated-encryption-handshake.
//
// The original STS protocol, which inspired this implementation:
// https://citeseerx.ist.psu.edu/document?rapid=rep1&type=pdf&doi=b852bc961328ce74f7231a4b569eec1ab6c3cf50. # codespell:ignore
//
// Consumers of the SecretConnection are responsible for authenticating
// the remote peer's pubkey against known information, like a nodeID.
type SecretConnection struct {
	// immutable
	recvAead cipher.AEAD
	sendAead cipher.AEAD

	remPubKey crypto.PubKey

	conn       io.ReadWriteCloser
	connWriter *bufio.Writer
	connReader io.Reader

	// net.Conn must be thread safe:
	// https://golang.org/pkg/net/#Conn.
	// Since we have internal mutable state,
	// we need mtxs. But recv and send states
	// are independent, so we can use two mtxs.
	// All .Read are covered by recvMtx,
	// all .Write are covered by sendMtx.
	recvMtx         cmtsync.Mutex
	recvBuffer      []byte
	recvNonce       *[aeadNonceSize]byte
	recvFrame       []byte
	recvSealedFrame []byte

	sendMtx         cmtsync.Mutex
	sendNonce       *[aeadNonceSize]byte
	sendFrame       []byte
	sendSealedFrame []byte
}

// MakeSecretConnection performs handshake and returns a new authenticated
// SecretConnection.
// Returns nil if there is an error in handshake.
// Caller should call conn.Close().
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	// NOTE: every 32-byte string is accepted as a Curve25519 public key (see
	// DJB's Curve25519 paper: http://cr.yp.to/ecdh/curve25519-20060209.pdf)
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, err
	}

	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least,
	// lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, err
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via
	// HKDF-SHA2 on the dhSecret.
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeySend
	}

	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeyRecv
	}

	sc := &SecretConnection{
		conn:            conn,
		connWriter:      bufio.NewWriterSize(conn, defaultWriteBufferSize),
		connReader:      bufio.NewReaderSize(conn, defaultReadBufferSize),
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
	}

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, locPrivKey)
	if err != nil {
		return nil, err
	}

	// Share (in secret) each other's pubkey & challenge signature
	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, err
	}

	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	// Usage in your function
	if _, ok := remPubKey.(ed25519.PubKey); !ok {
		return nil, ErrUnexpectedPubKeyType{
			Expected: ed25519.KeyType,
			Got:      remPubKey.Type(),
		}
	}

	if !remPubKey.VerifySignature(challenge[:], remSignature) {
		return nil, ErrChallengeVerification
	}

	// We've authorized.
	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey returns authenticated remote pubkey.
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Writes encrypted frames of `totalFrameSize + aeadSizeOverhead`.
// CONTRACT: data smaller than dataMaxSize is written atomically.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	sc.sendMtx.Lock()
	defer sc.sendMtx.Unlock()
	sealedFrame, frame := sc.sendSealedFrame, sc.sendFrame

	for 0 < len(data) {
		if err := func() error {
			var chunk []byte
			if dataMaxSize < len(data) {
				chunk = data[:dataMaxSize]
				data = data[dataMaxSize:]
			} else {
				chunk = data
				data = nil
			}
			chunkLength := len(chunk)
			binary.LittleEndian.PutUint32(frame, uint32(chunkLength))
			copy(frame[dataLenSize:], chunk)

			// encrypt the frame
			sc.sendAead.Seal(sealedFrame[:0], sc.sendNonce[:], frame, nil)
			incrNonce(sc.sendNonce)
			// end encryption

			_, err = sc.connWriter.Write(sealedFrame)
			if err != nil {
				return err
			}

			n += len(chunk)
			return nil
		}(); err != nil {
			return n, err
		}
	}
	sc.connWriter.Flush()
	return n, err
}

// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	sc.recvMtx.Lock()
	defer sc.recvMtx.Unlock()

	// read off and update the recvBuffer, if non-empty
	if 0 < len(sc.recvBuffer) {
		n = copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[n:]
		return n, err
	}

	// read off the conn
	sealedFrame := sc.recvSealedFrame
	_, err = io.ReadFull(sc.connReader, sealedFrame)
	if err != nil {
		return n, err
	}

	// decrypt the frame.
	// reads and updates the sc.recvNonce
	frame := sc.recvFrame
	_, err = sc.recvAead.Open(frame[:0], sc.recvNonce[:], sealedFrame, nil)
	if err != nil {
		return n, ErrDecryptFrame{Source: err}
	}

	incrNonce(sc.recvNonce)
	// end decryption

	// copy checkLength worth into data,
	// set recvBuffer to the rest.
	chunkLength := binary.LittleEndian.Uint32(frame) // read the first four bytes
	if chunkLength > dataMaxSize {
		return 0, ErrChunkTooBig{
			Received: int(chunkLength),
			Max:      dataMaxSize,
		}
	}

	chunk := frame[dataLenSize : dataLenSize+chunkLength]
	n = copy(data, chunk)
	if n < len(chunk) {
		sc.recvBuffer = make([]byte, len(chunk)-n)
		copy(sc.recvBuffer, chunk[n:])
	}
	return n, err
}

// Implements net.Conn.
func (sc *SecretConnection) Close() error                  { return sc.conn.Close() }
func (sc *SecretConnection) LocalAddr() net.Addr           { return sc.conn.(net.Conn).LocalAddr() }
func (sc *SecretConnection) RemoteAddr() net.Addr          { return sc.conn.(net.Conn).RemoteAddr() }
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		panic("failed to generate ephemeral key-pair")
	}
	return ephPub, ephPriv
}

func shareEphPubKey(conn io.ReadWriter, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	// Send our pubkey and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			lc := *locEphPub
			_, err = protoio.NewDelimitedWriter(conn).WriteMsg(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var bytes gogotypes.BytesValue
			_, err = protoio.NewDelimitedReader(conn, 1024*1024).ReadMsg(&bytes)
			if err != nil {
				return nil, true, err // abort
			}

			var _remEphPub [32]byte
			copy(_remEphPub[:], bytes.Value)
			return _remEphPub, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return remEphPub, err
	}

	// Otherwise:
	_remEphPub := trs.FirstValue().([32]byte)
	return &_remEphPub, nil
}

func deriveSecrets(
	dhSecret *[32]byte,
	locIsLeast bool,
) (recvSecret, sendSecret *[aeadKeySize]byte) {
	hash := sha256.New
	hkdf := hkdf.New(hash, dhSecret[:], nil, secretConnKeyAndChallengeGen)
	// get enough data for 2 aead keys, and a 32 byte challenge
	res := new([2*aeadKeySize + 32]byte)
	_, err := io.ReadFull(hkdf, res[:])
	if err != nil {
		panic(err)
	}

	recvSecret = new([aeadKeySize]byte)
	sendSecret = new([aeadKeySize]byte)

	// bytes 0 through aeadKeySize - 1 are one aead key.
	// bytes aeadKeySize through 2*aeadKeySize -1 are another aead key.
	// which key corresponds to sending and receiving key depends on whether
	// the local key is less than the remote key.
	if locIsLeast {
		copy(recvSecret[:], res[0:aeadKeySize])
		copy(sendSecret[:], res[aeadKeySize:aeadKeySize*2])
	} else {
		copy(sendSecret[:], res[0:aeadKeySize])
		copy(recvSecret[:], res[aeadKeySize:aeadKeySize*2])
	}

	return recvSecret, sendSecret
}

// computeDHSecret computes a Diffie-Hellman shared secret key
// from our own local private key and the other's public key.
func computeDHSecret(remPubKey, locPrivKey *[32]byte) (*[32]byte, error) {
	shrKey, err := curve25519.X25519(locPrivKey[:], remPubKey[:])
	if err != nil {
		return nil, err
	}
	var shrKeyArray [32]byte
	copy(shrKeyArray[:], shrKey)
	return &shrKeyArray, nil
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		lo = foo
		hi = bar
	} else {
		lo = bar
		hi = foo
	}
	return lo, hi
}

func signChallenge(challenge *[32]byte, locPrivKey crypto.PrivKey) ([]byte, error) {
	signature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig []byte
}

func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature []byte) (recvMsg authSigMessage, err error) {
	// Send our info and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			pbpk, err := cryptoenc.PubKeyToProto(pubKey)
			if err != nil {
				return nil, true, err
			}
			_, err = protoio.NewDelimitedWriter(sc).WriteMsg(&tmp2p.AuthSigMessage{PubKey: pbpk, Sig: signature})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var pba tmp2p.AuthSigMessage
			_, err = protoio.NewDelimitedReader(sc, 1024*1024).ReadMsg(&pba)
			if err != nil {
				return nil, true, err // abort
			}

			pk, err := cryptoenc.PubKeyFromProto(pba.PubKey)
			if err != nil {
				return nil, true, err // abort
			}

			_recvMsg := authSigMessage{
				Key: pk,
				Sig: pba.Sig,
			}
			return _recvMsg, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return recvMsg, err
	}

	_recvMsg := trs.FirstValue().(authSigMessage)
	return _recvMsg, nil
}

// --------------------------------------------------------------------------------

// Increment nonce little-endian by 1 with wraparound.
// Due to chacha20poly1305 expecting a 12 byte nonce we do not use the first four
// bytes. We only increment a 64 bit unsigned int in the remaining 8 bytes
// (little-endian in nonce[4:]).
func incrNonce(nonce *[aeadNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		// Terminates the session and makes sure the nonce would not re-used.
		// See https://github.com/tendermint/tendermint/issues/3531
		panic("can't increase nonce without overflow")
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}

```

---

<a name="file-14"></a>

### File: `tcp/conn/secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 13 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/internal/async"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
)

// Run go test -update from within this module
// to update the golden test vector file.
var update = flag.Bool("update", false, "update .golden files")

type kvstoreConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (drw kvstoreConn) Close() (err error) {
	err2 := drw.PipeWriter.CloseWithError(io.EOF)
	err1 := drw.PipeReader.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

type privKeyWithNilPubKey struct {
	orig crypto.PrivKey
}

func (pk privKeyWithNilPubKey) Bytes() []byte                   { return pk.orig.Bytes() }
func (pk privKeyWithNilPubKey) Sign(msg []byte) ([]byte, error) { return pk.orig.Sign(msg) }
func (privKeyWithNilPubKey) PubKey() crypto.PubKey              { return nil }
func (privKeyWithNilPubKey) Type() string                       { return "privKeyWithNilPubKey" }

func TestSecretConnectionHandshake(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
	if err := barSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentWrite(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)

	// write from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	n := 100
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)

	// Consume reads from bar's reader
	readLots(t, wg, barSecConn, n*2)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentRead(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)
	n := 100

	// read from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go readLots(t, wg, fooSecConn, n/2)
	go readLots(t, wg, fooSecConn, n/2)

	// write to bar
	writeLots(t, wg, barSecConn, fooWriteText, n)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestSecretConnectionReadWrite(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	fooWrites, barWrites := []string{}, []string{}
	fooReads, barReads := []string{}, []string{}

	// Pre-generate the things to write (for foo & bar)
	for i := 0; i < 100; i++ {
		fooWrites = append(fooWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
		barWrites = append(barWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
	}

	// A helper that will run with (fooConn, fooWrites, fooReads) and vice versa
	genNodeRunner := func(nodeConn kvstoreConn, nodeWrites []string, nodeReads *[]string) async.Task {
		return func(_ int) (any, bool, error) {
			// Initiate cryptographic private key and secret connection through nodeConn.
			nodePrvKey := ed25519.GenPrivKey()
			nodeSecretConn, err := MakeSecretConnection(nodeConn, nodePrvKey)
			if err != nil {
				t.Errorf("failed to establish SecretConnection for node: %v", err)
				return nil, true, err
			}
			// In parallel, handle some reads and writes.
			trs, ok := async.Parallel(
				func(_ int) (any, bool, error) {
					// Node writes:
					for _, nodeWrite := range nodeWrites {
						n, err := nodeSecretConn.Write([]byte(nodeWrite))
						if err != nil {
							t.Errorf("failed to write to nodeSecretConn: %v", err)
							return nil, true, err
						}
						if n != len(nodeWrite) {
							err = fmt.Errorf("failed to write all bytes. Expected %v, wrote %v", len(nodeWrite), n)
							t.Error(err)
							return nil, true, err
						}
					}
					if err := nodeConn.PipeWriter.Close(); err != nil {
						t.Error(err)
						return nil, true, err
					}
					return nil, false, nil
				},
				func(_ int) (any, bool, error) {
					// Node reads:
					readBuffer := make([]byte, dataMaxSize)
					for {
						n, err := nodeSecretConn.Read(readBuffer)
						if errors.Is(err, io.EOF) {
							if err := nodeConn.PipeReader.Close(); err != nil {
								t.Error(err)
								return nil, true, err
							}
							return nil, false, nil
						} else if err != nil {
							t.Errorf("failed to read from nodeSecretConn: %v", err)
							return nil, true, err
						}
						*nodeReads = append(*nodeReads, string(readBuffer[:n]))
					}
				},
			)
			assert.True(t, ok, "Unexpected task abortion")

			// If error:
			if trs.FirstError() != nil {
				return nil, true, trs.FirstError()
			}

			// Otherwise:
			return nil, false, nil
		}
	}

	// Run foo & bar in parallel
	trs, ok := async.Parallel(
		genNodeRunner(fooConn, fooWrites, &fooReads),
		genNodeRunner(barConn, barWrites, &barReads),
	)
	require.NoError(t, trs.FirstError())
	require.True(t, ok, "unexpected task abortion")

	// A helper to ensure that the writes and reads match.
	// Additionally, small writes (<= dataMaxSize) must be atomically read.
	compareWritesReads := func(writes []string, reads []string) {
		for {
			// Pop next write & corresponding reads
			read := ""
			write := writes[0]
			readCount := 0
			for _, readChunk := range reads {
				read += readChunk
				readCount++
				if len(write) <= len(read) {
					break
				}
				if len(write) <= dataMaxSize {
					break // atomicity of small writes
				}
			}
			// Compare
			if write != read {
				t.Errorf("expected to read %X, got %X", write, read)
			}
			// Iterate
			writes = writes[1:]
			reads = reads[readCount:]
			if len(writes) == 0 {
				break
			}
		}
	}

	compareWritesReads(fooWrites, barReads)
	compareWritesReads(barWrites, fooReads)
}

func TestDeriveSecretsAndChallengeGolden(t *testing.T) {
	goldenFilepath := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("Updating golden test vector file %s", goldenFilepath)
		data := createGoldenTestVectors(t)
		err := cmtos.WriteFile(goldenFilepath, []byte(data), 0o644)
		require.NoError(t, err)
	}
	f, err := os.Open(goldenFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		randSecretVector, err := hex.DecodeString(params[0])
		require.NoError(t, err)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		locIsLeast, err := strconv.ParseBool(params[1])
		require.NoError(t, err)
		expectedRecvSecret, err := hex.DecodeString(params[2])
		require.NoError(t, err)
		expectedSendSecret, err := hex.DecodeString(params[3])
		require.NoError(t, err)

		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		require.Equal(t, expectedRecvSecret, (*recvSecret)[:], "Recv Secrets aren't equal")
		require.Equal(t, expectedSendSecret, (*sendSecret)[:], "Send Secrets aren't equal")
	}
}

func TestNilPubkey(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	defer fooConn.Close()
	defer barConn.Close()
	fooPrvKey := ed25519.GenPrivKey()
	barPrvKey := privKeyWithNilPubKey{ed25519.GenPrivKey()}

	go MakeSecretConnection(fooConn, fooPrvKey) //nolint:errcheck // ignore for tests

	_, err := MakeSecretConnection(barConn, barPrvKey)
	require.Error(t, err)
	assert.Equal(t, "encoding: unsupported key <nil>", err.Error())
}

func writeLots(t *testing.T, wg *sync.WaitGroup, conn io.Writer, txt string, n int) {
	t.Helper()
	defer wg.Done()
	for i := 0; i < n; i++ {
		_, err := conn.Write([]byte(txt))
		if err != nil {
			t.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
}

func readLots(t *testing.T, wg *sync.WaitGroup, conn io.Reader, n int) {
	t.Helper()
	readBuffer := make([]byte, dataMaxSize)
	for i := 0; i < n; i++ {
		_, err := conn.Read(readBuffer)
		require.NoError(t, err)
	}
	wg.Done()
}

// Creates the data for a test vector file.
// The file format is:
// Hex(diffie_hellman_secret), loc_is_least, Hex(recvSecret), Hex(sendSecret), Hex(challenge).
func createGoldenTestVectors(*testing.T) string {
	data := ""
	for i := 0; i < 32; i++ {
		randSecretVector := cmtrand.Bytes(32)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		data += hex.EncodeToString((*randSecret)[:]) + ","
		locIsLeast := cmtrand.Bool()
		data += strconv.FormatBool(locIsLeast) + ","
		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		data += hex.EncodeToString((*recvSecret)[:]) + ","
		data += hex.EncodeToString((*sendSecret)[:]) + ","
	}
	return data
}

// Each returned ReadWriteCloser is akin to a net.Connection.
func makeKVStoreConnPair() (fooConn, barConn kvstoreConn) {
	barReader, fooWriter := io.Pipe()
	fooReader, barWriter := io.Pipe()
	return kvstoreConn{fooReader, fooWriter}, kvstoreConn{barReader, barWriter}
}

func makeSecretConnPair(tb testing.TB) (fooSecConn, barSecConn *SecretConnection) {
	tb.Helper()
	var (
		fooConn, barConn = makeKVStoreConnPair()
		fooPrvKey        = ed25519.GenPrivKey()
		fooPubKey        = fooPrvKey.PubKey()
		barPrvKey        = ed25519.GenPrivKey()
		barPubKey        = barPrvKey.PubKey()
	)

	// Make connections from both sides in parallel.
	trs, ok := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			fooSecConn, err = MakeSecretConnection(fooConn, fooPrvKey)
			if err != nil {
				tb.Errorf("failed to establish SecretConnection for foo: %v", err)
				return nil, true, err
			}
			remotePubBytes := fooSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), barPubKey.Bytes()) {
				err = fmt.Errorf("unexpected fooSecConn.RemotePubKey.  Expected %v, got %v",
					barPubKey, fooSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			barSecConn, err = MakeSecretConnection(barConn, barPrvKey)
			if barSecConn == nil {
				tb.Errorf("failed to establish SecretConnection for bar: %v", err)
				return nil, true, err
			}
			remotePubBytes := barSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), fooPubKey.Bytes()) {
				err = fmt.Errorf("unexpected barSecConn.RemotePubKey.  Expected %v, got %v",
					fooPubKey, barSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
	)

	require.NoError(tb, trs.FirstError())
	require.True(tb, ok, "Unexpected task abortion")

	return fooSecConn, barSecConn
}

// Benchmarks

func BenchmarkWriteSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	// Consume reads from bar's reader
	go func() {
		readBuffer := make([]byte, dataMaxSize)
		for {
			_, err := barSecConn.Read(readBuffer)
			if errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				b.Errorf("failed to read from barSecConn: %v", err)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := cmtrand.Intn(len(fooWriteBytes))
		_, err := fooSecConn.Write(fooWriteBytes[idx])
		if err != nil {
			b.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
	b.StopTimer()

	if err := fooSecConn.Close(); err != nil {
		b.Error(err)
	}
	// barSecConn.Close() race condition
}

func BenchmarkReadSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	go func() {
		for i := 0; i < b.N; i++ {
			idx := cmtrand.Intn(len(fooWriteBytes))
			_, err := fooSecConn.Write(fooWriteBytes[idx])
			if err != nil {
				b.Errorf("failed to write to fooSecConn: %v, %v,%v", err, i, b.N)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		readBuffer := make([]byte, dataMaxSize)
		_, err := barSecConn.Read(readBuffer)

		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			b.Fatalf("Failed to read from barSecConn: %v", err)
		}
	}
	b.StopTimer()
}

```

---

<a name="file-15"></a>

### File: `tcp/conn/stream.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import "time"

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn     *MConnection
	streamID byte
}

// Read reads bytes for the given stream from the internal read queue. Used in
// tests. Production code should use MConnection.OnReceive to avoid copying the
// data.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	return s.conn.readBytes(s.streamID, b, 5*time.Second)
}

// Write queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) Write(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, true /* blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// TryWrite queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) TryWrite(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, false /* non-blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Close closes the stream.
// thread-safe.
func (s *MConnectionStream) Close() error {
	delete(s.conn.channelsIdx, s.streamID)
	return nil
}

```

---

<a name="file-16"></a>

### File: `tcp/conn/stream_descriptor.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
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

```

---

<a name="file-17"></a>

### File: `tcp/conn_set.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"net"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(conn net.Conn) bool
	HasIP(ip net.IP) bool
	Set(conn net.Conn, ip []net.IP)
	Remove(conn net.Conn)
	RemoveAddr(addr net.Addr)
}

type connSetItem struct {
	conn net.Conn
	ips  []net.IP
}

type connSet struct {
	cmtsync.RWMutex

	conns map[string]connSetItem
}

// NewConnSet returns a ConnSet implementation.
func NewConnSet() ConnSet {
	return &connSet{
		conns: map[string]connSetItem{},
	}
}

func (cs *connSet) Has(c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[c.RemoteAddr().String()]

	return ok
}

func (cs *connSet) HasIP(ip net.IP) bool {
	cs.RLock()
	defer cs.RUnlock()

	for _, c := range cs.conns {
		for _, known := range c.ips {
			if known.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func (cs *connSet) Remove(c net.Conn) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, c.RemoteAddr().String())
}

func (cs *connSet) RemoveAddr(addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, addr.String())
}

func (cs *connSet) Set(c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[c.RemoteAddr().String()] = connSetItem{
		conn: c,
		ips:  ips,
	}
}

```

---

<a name="file-18"></a>

### File: `tcp/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
)

// ErrTransportClosed is raised when the Transport has been closed.
type ErrTransportClosed struct{}

func (ErrTransportClosed) Error() string {
	return "transport has been closed"
}

// ErrFilterTimeout indicates that a filter operation timed out.
type ErrFilterTimeout struct{}

func (ErrFilterTimeout) Error() string {
	return "filter timed out"
}

// ErrRejected indicates that a Peer was rejected carrying additional
// information as to the reason.
type ErrRejected struct {
	addr          na.NetAddr
	conn          net.Conn
	err           error
	id            nodekey.ID
	isAuthFailure bool
	isDuplicate   bool
	isFiltered    bool
}

// Addr returns the network address for the rejected Peer.
func (e ErrRejected) Addr() na.NetAddr {
	return e.addr
}

func (e ErrRejected) Error() string {
	if e.isAuthFailure {
		return fmt.Sprintf("auth failure: %s", e.err)
	}

	if e.isDuplicate {
		if e.conn != nil {
			return fmt.Sprintf(
				"duplicate CONN<%s>",
				e.conn.RemoteAddr().String(),
			)
		}
		if e.id != "" {
			return fmt.Sprintf("duplicate ID<%v>", e.id)
		}
	}

	if e.isFiltered {
		if e.conn != nil {
			return fmt.Sprintf(
				"filtered CONN<%s>: %s",
				e.conn.RemoteAddr().String(),
				e.err,
			)
		}

		if e.id != "" {
			return fmt.Sprintf("filtered ID<%v>: %s", e.id, e.err)
		}
	}

	return e.err.Error()
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e ErrRejected) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e ErrRejected) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e ErrRejected) IsFiltered() bool { return e.isFiltered }

```

---

<a name="file-19"></a>

### File: `tcp/tcp.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 11 KB

```go
package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/netutil"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/fuzz"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

const (
	defaultDialTimeout      = time.Second
	defaultFilterTimeout    = 5 * time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// IPResolver is a behavior subset of net.Resolver.
type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// accept is the container to carry the upgraded connection from an
// asynchronously running routine to the Accept method.
type accept struct {
	netAddr *na.NetAddr
	conn    *conn.MConnection
	err     error
}

// ConnFilterFunc to be implemented by filter hooks after a new connection has
// been established. The set of existing connections is passed along together
// with all resolved IPs for the new connection.
type ConnFilterFunc func(ConnSet, net.Conn, []net.IP) error

// ConnDuplicateIPFilter resolves and keeps all ips for an incoming connection
// and refuses new ones if they come from a known ip.
func ConnDuplicateIPFilter() ConnFilterFunc {
	return func(cs ConnSet, c net.Conn, ips []net.IP) error {
		for _, ip := range ips {
			if cs.HasIP(ip) {
				return ErrRejected{
					conn:        c,
					err:         fmt.Errorf("ip<%v> already connected", ip),
					isDuplicate: true,
				}
			}
		}

		return nil
	}
}

// MultiplexTransportOption sets an optional parameter on the
// MultiplexTransport.
type MultiplexTransportOption func(*MultiplexTransport)

// MultiplexTransportConnFilters sets the filters for rejection new connections.
func MultiplexTransportConnFilters(
	filters ...ConnFilterFunc,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.connFilters = filters }
}

// MultiplexTransportFilterTimeout sets the timeout waited for filter calls to
// return.
func MultiplexTransportFilterTimeout(
	timeout time.Duration,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.filterTimeout = timeout }
}

// MultiplexTransportResolver sets the Resolver used for ip lokkups, defaults to
// net.DefaultResolver.
func MultiplexTransportResolver(resolver IPResolver) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.resolver = resolver }
}

// MultiplexTransportMaxIncomingConnections sets the maximum number of
// simultaneous connections (incoming). Default: 0 (unlimited).
func MultiplexTransportMaxIncomingConnections(n int) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.maxIncomingConnections = n }
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	netAddr                na.NetAddr
	listener               net.Listener
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	conns       ConnSet
	connFilters []ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeKey          nodekey.NodeKey
	resolver         IPResolver

	// TODO(xla): This config is still needed as we parameterise peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig *conn.MConnConfig
	logger  log.Logger
}

// Test multiplexTransport for interface completeness.
var (
	_ transport.Transport = (*MultiplexTransport)(nil)
)

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(nodeKey nodekey.NodeKey, mConfig conn.MConnConfig) *MultiplexTransport {
	return &MultiplexTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      defaultDialTimeout,
		filterTimeout:    defaultFilterTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          &mConfig,
		nodeKey:          nodeKey,
		conns:            NewConnSet(),
		resolver:         net.DefaultResolver,
		logger:           log.NewNopLogger(),
	}
}

// SetLogger sets the logger for the transport.
func (mt *MultiplexTransport) SetLogger(l log.Logger) {
	mt.logger = l
}

// NetAddr implements Transport.
func (mt *MultiplexTransport) NetAddr() na.NetAddr {
	return mt.netAddr
}

// Accept implements Transport.
func (mt *MultiplexTransport) Accept() (transport.Conn, *na.NetAddr, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, nil, a.err
		}

		return a.conn, a.netAddr, nil
	case <-mt.closec:
		return nil, nil, ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *MultiplexTransport) Dial(addr na.NetAddr) (transport.Conn, error) {
	c, err := addr.DialTimeout(mt.dialTimeout)
	if err != nil {
		return nil, err
	}

	if mt.mConfig.TestFuzz {
		// so we have time to do peer handshakes and get set up.
		c = fuzz.ConnAfterFromConfig(c, 10*time.Second, mt.mConfig.TestFuzzConfig)
	}

	// TODO(xla): Evaluate if we should apply filters if we explicitly dial.
	if err := mt.filterConn(c); err != nil {
		return nil, err
	}

	mconn, _, err := mt.upgrade(c, &addr)
	if err != nil {
		return nil, err
	}
	mconn.SetLogger(mt.logger.With("remote", addr))

	go mt.cleanupConn(c.RemoteAddr(), mconn.Quit())

	return mconn, nil
}

func (mt *MultiplexTransport) Close() error {
	close(mt.closec)

	if mt.listener != nil {
		return mt.listener.Close()
	}

	return nil
}

func (mt *MultiplexTransport) Listen(addr na.NetAddr) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return err
	}

	if mt.maxIncomingConnections > 0 {
		ln = netutil.LimitListener(ln, mt.maxIncomingConnections)
	}

	mt.netAddr = *na.New(addr.ID, ln.Addr())
	mt.listener = ln

	go mt.acceptPeers()

	return nil
}

func (mt *MultiplexTransport) cleanupConn(netAddr net.Addr, quitCh <-chan struct{}) {
	select {
	case <-quitCh:
		mt.conns.RemoveAddr(netAddr)
	case <-mt.closec:
		return
	}
}

func (mt *MultiplexTransport) acceptPeers() {
	for {
		c, err := mt.listener.Accept()
		if err != nil {
			// If Close() has been called, silently exit.
			select {
			case _, ok := <-mt.closec:
				if !ok {
					return
				}
			default:
				// Transport is not closed
			}

			mt.acceptc <- accept{err: err}
			return
		}

		// Connection upgrade and filtering should be asynchronous to avoid
		// Head-of-line blocking[0].
		// Reference:  https://github.com/tendermint/tendermint/issues/2047
		//
		// [0] https://en.wikipedia.org/wiki/Head-of-line_blocking
		go func(c net.Conn) {
			defer func() {
				if r := recover(); r != nil {
					err := ErrRejected{
						conn:          c,
						err:           fmt.Errorf("recovered from panic: %v", r),
						isAuthFailure: true,
					}
					select {
					case mt.acceptc <- accept{err: err}:
					case <-mt.closec:
						// Give up if the transport was closed.
						_ = c.Close()
						return
					}
				}
			}()

			var (
				mconn        *conn.MConnection
				remotePubKey crypto.PubKey
				netAddr      *na.NetAddr
			)

			err := mt.filterConn(c)
			if err == nil {
				mconn, remotePubKey, err = mt.upgrade(c, nil)
				if err == nil {
					addr := c.RemoteAddr()
					id := nodekey.PubKeyToID(remotePubKey)
					netAddr = na.New(id, addr)
					mconn.SetLogger(mt.logger.With("remote", netAddr))
					go mt.cleanupConn(addr, mconn.Quit())
				}
			}

			select {
			case mt.acceptc <- accept{netAddr, mconn, err}:
				// Make the upgraded peer available.
			case <-mt.closec:
				// Give up if the transport was closed.
				_ = c.Close()
				return
			}
		}(c)
	}
}

func (mt *MultiplexTransport) filterConn(c net.Conn) (err error) {
	defer func() {
		if err != nil {
			_ = c.Close()
		}
	}()

	// Reject if connection is already present.
	if mt.conns.Has(c) {
		return ErrRejected{conn: c, isDuplicate: true}
	}

	// Resolve ips for incoming conn.
	ips, err := resolveIPs(mt.resolver, c)
	if err != nil {
		return err
	}

	errc := make(chan error, len(mt.connFilters))

	for _, f := range mt.connFilters {
		go func(f ConnFilterFunc, c net.Conn, ips []net.IP, errc chan<- error) {
			errc <- f(mt.conns, c, ips)
		}(f, c, ips, errc)
	}

	for i := 0; i < cap(errc); i++ {
		select {
		case err := <-errc:
			if err != nil {
				return ErrRejected{conn: c, err: err, isFiltered: true}
			}
		case <-time.After(mt.filterTimeout):
			return ErrFilterTimeout{}
		}
	}

	mt.conns.Set(c, ips)

	return nil
}

func (mt *MultiplexTransport) upgrade(
	c net.Conn,
	dialedAddr *na.NetAddr,
) (*conn.MConnection, crypto.PubKey, error) {
	var err error
	defer func() {
		if err != nil {
			mt.conns.Remove(c)
			_ = c.Close()
		}
	}()

	secretConn, err := upgradeSecretConn(c, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	if err != nil {
		return nil, nil, ErrRejected{
			conn:          c,
			err:           fmt.Errorf("secret conn failed: %w", err),
			isAuthFailure: true,
		}
	}

	// For outgoing conns, ensure connection key matches dialed key.
	remotePubKey := secretConn.RemotePubKey()
	connID := nodekey.PubKeyToID(remotePubKey)
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID != dialedID {
			return nil, nil, ErrRejected{
				conn: c,
				id:   connID,
				err: fmt.Errorf(
					"conn.ID (%v) dialed ID (%v) mismatch",
					connID,
					dialedID,
				),
				isAuthFailure: true,
			}
		}
	}

	// Copy MConnConfig to avoid it being modified by the transport.
	return conn.NewMConnection(secretConn, *mt.mConfig), remotePubKey, nil
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func resolveIPs(resolver IPResolver, c net.Conn) ([]net.IP, error) {
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	addrs, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}

	ips := []net.IP{}

	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}

	return ips, nil
}

```

---

<a name="file-20"></a>

### File: `tcp/tcp_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 10 KB

```go
package tcp

import (
	"errors"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

// newMultiplexTransport returns a tcp connected multiplexed peer
// using the default MConnConfig. It's a convenience function used
// for testing.
func newMultiplexTransport(
	nodeKey nodekey.NodeKey,
) *MultiplexTransport {
	return NewMultiplexTransport(
		nodeKey, conn.DefaultMConnConfig(),
	)
}

func TestTransportMultiplex_ConnFilter(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			return errors.New("rejected")
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if e, ok := err.(ErrRejected); ok {
		if !e.IsFiltered() {
			t.Errorf("expected peer to be filtered, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplex_ConnFilterTimeout(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)
	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if _, ok := err.(ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout, got %v", err)
	}
}

func TestTransportMultiplex_MaxIncomingConnections(t *testing.T) {
	pv := ed25519.GenPrivKey()
	id := nodekey.PubKeyToID(pv.PubKey())
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: pv,
		},
	)

	MultiplexTransportMaxIncomingConnections(0)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	const maxIncomingConns = 2
	MultiplexTransportMaxIncomingConnections(maxIncomingConns)(mt)
	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	// Connect more peers than max
	for i := 0; i <= maxIncomingConns; i++ {
		errc := make(chan error)
		go testDialer(*laddr, errc)

		err = <-errc
		if i < maxIncomingConns {
			if err != nil {
				t.Errorf("dialer connection failed: %v", err)
			}
			_, _, err = mt.Accept()
			if err != nil {
				t.Errorf("connection failed: %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "i/o timeout") {
			// mt actually blocks forever on trying to accept a new peer into a full channel so
			// expect the dialer to encounter a timeout error. Calling mt.Accept will block until
			// mt is closed.
			t.Errorf("expected i/o timeout error, got %v", err)
		}
	}
}

func TestTransportMultiplex_AcceptMultiple(t *testing.T) {
	mt := testSetupMultiplexTransport(t)
	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	var (
		seed     = rand.New(rand.NewSource(time.Now().UnixNano()))
		nDialers = seed.Intn(64) + 64
		errc     = make(chan error, nDialers)
	)

	// Setup dialers.
	for i := 0; i < nDialers; i++ {
		go testDialer(*laddr, errc)
	}

	// Catch connection errors.
	for i := 0; i < nDialers; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}

	conns := []transport.Conn{}

	// Accept all connections.
	for i := 0; i < cap(errc); i++ {
		c, _, err := mt.Accept()
		if err != nil {
			t.Fatal(err)
		}

		conns = append(conns, c)
	}

	if have, want := len(conns), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func testDialer(dialAddr na.NetAddr, errc chan error) {
	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	_, err := dialer.Dial(dialAddr)
	if err != nil {
		errc <- err
		return
	}

	// Signal that the connection was established.
	errc <- nil
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		fastNodePV = ed25519.GenPrivKey()
		errc       = make(chan error)
		fastc      = make(chan struct{})
		slowc      = make(chan struct{})
		slowdonec  = make(chan struct{})
	)

	// Simulate slow Peer.
	go func() {
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		c, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(slowc)
		defer func() {
			close(slowdonec)
		}()

		// Make sure we switch to fast peer goroutine.
		runtime.Gosched()

		select {
		case <-fastc:
			// Fast peer connected.
		case <-time.After(200 * time.Millisecond):
			// We error if the fast peer didn't succeed.
			errc <- errors.New("fast peer timed out")
		}

		_, err = upgradeSecretConn(c, 200*time.Millisecond, ed25519.GenPrivKey())
		if err != nil {
			errc <- err
			return
		}
	}()

	// Simulate fast Peer.
	go func() {
		<-slowc

		dialer := newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: fastNodePV,
			},
		)
		dialer.SetLogger(log.TestingLogger())
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr)
		if err != nil {
			errc <- err
			return
		}

		close(fastc)
		<-slowdonec
		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Logf("connection failed: %v", err)
	}

	_, _, err := mt.Accept()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransportMultiplexDialRejectWrongID(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	wrongID := nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
	addr := na.New(wrongID, mt.listener.Addr())

	_, err := dialer.Dial(*addr)
	if err != nil {
		t.Logf("connection failed: %v", err)
		if e, ok := err.(ErrRejected); ok {
			if !e.IsAuthFailure() {
				t.Errorf("expected auth failure, got %v", e)
			}
		} else {
			t.Errorf("expected ErrRejected, got %v", err)
		}
	}
}

func TestTransportConnDuplicateIPFilter(t *testing.T) {
	filter := ConnDuplicateIPFilter()

	if err := filter(nil, &testTransportConn{}, nil); err != nil {
		t.Fatal(err)
	}

	var (
		c  = &testTransportConn{}
		cs = NewConnSet()
	)

	cs.Set(c, []net.IP{
		{10, 0, 10, 1},
		{10, 0, 10, 2},
		{10, 0, 10, 3},
	})

	if err := filter(cs, c, []net.IP{
		{10, 0, 10, 2},
	}); err == nil {
		t.Errorf("expected Peer to be rejected as duplicate")
	}
}

// create listener.
func testSetupMultiplexTransport(t *testing.T) *MultiplexTransport {
	t.Helper()

	var (
		pv = ed25519.GenPrivKey()
		id = nodekey.PubKeyToID(pv.PubKey())
		mt = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	mt.SetLogger(log.TestingLogger())

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	// give the listener some time to get ready
	time.Sleep(20 * time.Millisecond)

	return mt
}

type testTransportAddr struct{}

func (*testTransportAddr) Network() string { return "tcp" }
func (*testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (*testTransportConn) Close() error {
	return errors.New("close() not implemented")
}

func (*testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) Read(_ []byte) (int, error) {
	return -1, errors.New("read() not implemented")
}

func (*testTransportConn) SetDeadline(_ time.Time) error {
	return errors.New("setDeadline() not implemented")
}

func (*testTransportConn) SetReadDeadline(_ time.Time) error {
	return errors.New("setReadDeadline() not implemented")
}

func (*testTransportConn) SetWriteDeadline(_ time.Time) error {
	return errors.New("setWriteDeadline() not implemented")
}

func (*testTransportConn) Write(_ []byte) (int, error) {
	return -1, errors.New("write() not implemented")
}

```

---

<a name="file-21"></a>

### File: `transport.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 1 KB

```go
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
}

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}

```

---

## Summary

- **Total files processed:** 21
- **Total combined size:** 231 KB

## Breakdown of File Sizes by Type

- **md**: 106 KB
- **go**: 125 KB
```

---

<a name="file-3"></a>

### File: `conn.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 3 KB

```go
package transport

import (
	"io"
	"net"
	"time"
)

// Conn is a multiplexed connection that can send and receive data
// on multiple streams.
type Conn interface {
	// OpenStream opens a new stream on the connection with an optional
	// description. If you're using tcp.MultiplexTransport, all streams must be
	// registered in advance.
	OpenStream(streamID byte, desc any) (Stream, error)

	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr

	// Close closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	Close(reason string) error

	// FlushAndClose flushes all the pending bytes and closes the connection.
	// If the protocol supports it, a reason will be sent to the remote.
	// Any blocked Read operations will be unblocked and return errors.
	FlushAndClose(reason string) error

	// ConnState returns basic details about the connection.
	// Warning: This API should not be considered stable and might change soon.
	ConnState() ConnState

	// ErrorCh returns a channel that will receive errors from the connection.
	ErrorCh() <-chan error

	// HandshakeStream returns the stream to be used for the handshake.
	HandshakeStream() HandshakeStream
}

// Stream is the interface implemented by QUIC streams or multiplexed TCP connection.
type Stream interface {
	SendStream
}

// A SendStream is a unidirectional Send Stream.
type SendStream interface {
	// Write writes data to the stream.
	// It blocks until data is sent or the stream is closed.
	io.Writer
	// Close closes the write-direction of the stream.
	// Future calls to Write are not permitted after calling Close.
	// It must not be called concurrently with Write.
	// It must not be called after calling CancelWrite.
	io.Closer
	// TryWrite attempts to write data to the stream.
	// If the send queue is full, the error satisfies the WriteError interface, and Full() will be true.
	TryWrite(b []byte) (n int, err error)
}

// WriteError is returned by TryWrite when the send queue is full.
type WriteError interface {
	error
	Full() bool // Is the error due to the send queue being full?
}

// HandshakeStream is a stream that is used for the handshake.
type HandshakeStream interface {
	SetDeadline(t time.Time) error
	io.ReadWriter
}

```

---

<a name="file-4"></a>

### File: `conn_state.go`

*Modified:* 2025-02-08 17:49:58 • *Size:* 1 KB

```go
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

```

---

<a name="file-5"></a>

### File: `quic/errors.go`

*Modified:* 2025-02-08 17:55:08 • *Size:* 1 KB

```go
package quic

import "errors"

var (
	// ErrTransportNotListening is returned when trying to accept connections before listening
	ErrTransportNotListening = errors.New("transport not listening")

	// ErrTransportClosed is returned when the transport has been closed
	ErrTransportClosed = errors.New("transport closed")

	// ErrInvalidAddress is returned when an invalid address is provided
	ErrInvalidAddress = errors.New("invalid address")
)

```

---

<a name="file-6"></a>

### File: `quic/quic.go`

*Modified:* 2025-02-08 18:05:49 • *Size:* 5 KB

```go
package quic

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/quic-go/quic-go"
)

// Default configuration values
const (
	defaultMaxIncomingStreams = 100
	defaultKeepAlivePeriod    = 30 * time.Second
	defaultIdleTimeout        = 5 * time.Minute
	defaultHandshakeTimeout   = 10 * time.Second
)

// Transport implements the transport.Transport interface using QUIC
type Transport struct {
	listener   quic.Listener
	tlsConfig  *tls.Config
	quicConfig *quic.Config
	logger     log.Logger
	metrics    *transport.MetricsCollector
	netAddr    *na.NetAddr

	// Connection management
	mtx         sync.RWMutex
	connections map[string]quic.Connection

	// Options
	maxStreams  int
	keepAlive   time.Duration
	idleTimeout time.Duration

	closed      chan struct{}
	isListening bool
}

// Options contains QUIC-specific configuration
type Options struct {
	TLSConfig          *tls.Config
	MaxIncomingStreams int
	KeepAlivePeriod    time.Duration
	IdleTimeout        time.Duration
}

// NewTransport creates a new QUIC transport instance
func NewTransport(opts *Options) (*Transport, error) {
	if opts == nil {
		opts = &Options{}
	}

	if opts.MaxIncomingStreams == 0 {
		opts.MaxIncomingStreams = defaultMaxIncomingStreams
	}
	if opts.KeepAlivePeriod == 0 {
		opts.KeepAlivePeriod = defaultKeepAlivePeriod
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = defaultIdleTimeout
	}

	quicConfig := &quic.Config{
		MaxIncomingStreams: int64(opts.MaxIncomingStreams),
		MaxIdleTimeout:     opts.IdleTimeout,
		KeepAlivePeriod:    opts.KeepAlivePeriod,
	}

	return &Transport{
		tlsConfig:   opts.TLSConfig,
		quicConfig:  quicConfig,
		connections: make(map[string]quic.Connection),
		maxStreams:  opts.MaxIncomingStreams,
		keepAlive:   opts.KeepAlivePeriod,
		idleTimeout: opts.IdleTimeout,
		closed:      make(chan struct{}),
		logger:      log.NewNopLogger(),
	}, nil
}

// Listen implements transport.Transport
func (t *Transport) Listen(laddr string) error {
	addr, err := net.ResolveUDPAddr("udp", laddr)
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	listener, err := quic.Listen(udpConn, t.tlsConfig, t.quicConfig)
	if err != nil {
		return err
	}

	t.listener = *listener
	t.isListening = true
	t.netAddr = na.New("", listener.Addr())

	return nil
}

// Dial implements transport.Transport
func (t *Transport) Dial(addr na.NetAddr) (transport.Conn, error) {
	ctx := context.Background()
	conn, err := quic.DialAddr(ctx, addr.String(), t.tlsConfig, t.quicConfig)
	if err != nil {
		return nil, err
	}

	t.mtx.Lock()
	t.connections[conn.RemoteAddr().String()] = conn
	t.mtx.Unlock()

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}

	return NewConn(conn, stream), nil
}

// Accept implements transport.Transport
func (t *Transport) Accept() (transport.Conn, *na.NetAddr, error) {
	if !t.isListening {
		return nil, nil, ErrTransportNotListening
	}

	conn, err := t.listener.Accept(context.Background())
	if err != nil {
		return nil, nil, err
	}

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, nil, err
	}

	// Create wrapper and address
	wrapper := NewConn(conn, stream)
	netAddr := na.New("", conn.RemoteAddr())

	return wrapper, netAddr, nil
}

// NetAddr implements transport.Transport
func (t *Transport) NetAddr() na.NetAddr {
	if t.netAddr == nil {
		return na.NetAddr{}
	}
	return *t.netAddr
}

// Close implements transport.Transport
func (t *Transport) Close() error {
	close(t.closed)

	t.mtx.Lock()
	defer t.mtx.Unlock()

	if t.isListening {
		if err := t.listener.Close(); err != nil {
			return err
		}
	}

	for _, conn := range t.connections {
		if err := conn.CloseWithError(0, "transport closed"); err != nil {
			t.logger.Error("Error closing connection", "err", err)
		}
	}

	t.isListening = false
	return nil
}

// SetLogger sets the logger
func (t *Transport) SetLogger(l log.Logger) {
	t.logger = l
}

```

---

<a name="file-7"></a>

### File: `quic/quic_test.go`

*Modified:* 2025-02-08 18:06:59 • *Size:* 4 KB

```go
package quic

import (
	"crypto/tls"
	"io"
	"testing"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/stretchr/testify/require"
)

func generateTestTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-test"},
	}
}

func TestQUICTransportBasics(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	// Create transport with options
	opts := &Options{
		TLSConfig:          tlsConfig,
		MaxIncomingStreams: 10,
		KeepAlivePeriod:    time.Second,
		IdleTimeout:        time.Minute,
	}

	transport, err := NewTransport(opts)
	require.NoError(t, err)

	// Listen on a random port
	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	// Get the assigned address
	addr := transport.NetAddr()

	// Try to connect
	clientTransport, err := NewTransport(opts)
	require.NoError(t, err)

	conn, err := clientTransport.Dial(addr)
	require.NoError(t, err)

	// Write some data using the handshake stream
	testData := []byte("hello world")
	hstream := conn.HandshakeStream()
	n, err := hstream.Write(testData)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)

	// Accept the connection on the server side
	serverConn, _, err := transport.Accept()
	require.NoError(t, err)

	// Read the data from the handshake stream
	buf := make([]byte, len(testData))
	hstream = serverConn.HandshakeStream()
	n, err = io.ReadFull(hstream, buf)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)
	require.Equal(t, testData, buf)

	// Close connections
	require.NoError(t, conn.Close("test done"))
	require.NoError(t, serverConn.Close("test done"))
	require.NoError(t, transport.Close())
	require.NoError(t, clientTransport.Close())
}

func TestQUICTransportConcurrent(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewTransport(&Options{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	addr := transport.NetAddr()

	// Launch multiple concurrent connections
	const numConns = 10
	done := make(chan struct{})

	for i := 0; i < numConns; i++ {
		go func() {
			clientTransport, err := NewTransport(&Options{
				TLSConfig: tlsConfig,
			})
			require.NoError(t, err)

			conn, err := clientTransport.Dial(addr)
			require.NoError(t, err)

			data := []byte("test data")
			hstream := conn.HandshakeStream()
			_, err = hstream.Write(data)
			require.NoError(t, err)

			require.NoError(t, conn.Close("done"))
			require.NoError(t, clientTransport.Close())
			done <- struct{}{}
		}()
	}

	// Accept and handle all connections
	for i := 0; i < numConns; i++ {
		conn, _, err := transport.Accept()
		require.NoError(t, err)

		go func(c transport.Conn) {
			buf := make([]byte, 1024)
			hstream := c.HandshakeStream()
			_, err := io.ReadFull(hstream, buf[:9]) // len("test data") = 9
			require.NoError(t, err)
			require.NoError(t, c.Close("done"))
		}(conn)
	}

	// Wait for all clients to finish
	for i := 0; i < numConns; i++ {
		<-done
	}

	require.NoError(t, transport.Close())
}

func TestQUICTransportError(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewTransport(&Options{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	// Try to accept before listening
	_, _, err = transport.Accept()
	require.Equal(t, ErrTransportNotListening, err)

	// Try to listen on invalid address
	err = transport.Listen("invalid-addr")
	require.Error(t, err)

	// Try to dial invalid address
	invalidAddr := na.NetAddr{}
	_, err = transport.Dial(invalidAddr)
	require.Error(t, err)

	require.NoError(t, transport.Close())
}

```

---

<a name="file-8"></a>

### File: `quic/stream.go`

*Modified:* 2025-02-08 18:05:16 • *Size:* 1 KB

```go
package quic

import (
	"github.com/quic-go/quic-go"
)

// Stream implements the transport.Stream interface
type Stream struct {
	stream quic.Stream
}

func (s *Stream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *Stream) Close() error {
	return s.stream.Close()
}

func (s *Stream) TryWrite(b []byte) (n int, err error) {
	// QUIC streams don't have a non-blocking write, so we just do a regular write
	return s.Write(b)
}

```

---

<a name="file-9"></a>

### File: `quic/wrapper.go`

*Modified:* 2025-02-08 18:04:53 • *Size:* 3 KB

```go
package quic

import (
	"net"
	"time"

	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/quic-go/quic-go"
)

// Conn wraps a QUIC connection and stream to implement transport.Conn
type Conn struct {
	conn    quic.Connection
	stream  quic.Stream
	errorCh chan error
}

func NewConn(conn quic.Connection, stream quic.Stream) *Conn {
	return &Conn{
		conn:    conn,
		stream:  stream,
		errorCh: make(chan error, 1),
	}
}

// Read implements io.Reader
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write implements io.Writer
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close implements transport.Conn
func (c *Conn) Close(reason string) error {
	if err := c.stream.Close(); err != nil {
		return err
	}
	return c.conn.CloseWithError(0, reason)
}

// LocalAddr implements net.Conn
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr implements net.Conn
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements net.Conn
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline implements net.Conn
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// HandshakeStream returns the underlying stream for the handshake
func (c *Conn) HandshakeStream() transport.HandshakeStream {
	return c.stream
}

// ErrorCh returns a channel that will receive errors from the connection
func (c *Conn) ErrorCh() <-chan error {
	return c.errorCh
}

// OpenStream opens a new stream on the connection
func (c *Conn) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	stream, err := c.conn.OpenStream()
	if err != nil {
		return nil, err
	}
	return &Stream{stream: stream}, nil
}

// ConnState returns basic details about the connection
func (c *Conn) ConnState() transport.ConnState {
	return transport.ConnState{
		StreamStates: make(map[byte]transport.StreamState),
	}
}

// FlushAndClose flushes all pending writes and closes the connection
func (c *Conn) FlushAndClose(reason string) error {
	return c.Close(reason)
}

```

---

<a name="file-10"></a>

### File: `tcp/conn/connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 26 KB

```go
package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/cosmos/gogoproto/proto"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	flow "github.com/cometbft/cometbft/internal/flowrate"
	"github.com/cometbft/cometbft/internal/timer"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/transport"
)

const (
	defaultMaxPacketMsgPayloadSize = 1024

	numBatchPacketMsgs = 10
	minReadBufferSize  = 1024
	minWriteBufferSize = 65536
	updateStats        = 2 * time.Second

	// some of these defaults are written in the user config
	// flushThrottle, sendRate, recvRate
	// TODO: remove values present in config.
	defaultFlushThrottle = 10 * time.Millisecond

	defaultSendRate     = int64(512000) // 500KB/s
	defaultRecvRate     = int64(512000) // 500KB/s
	defaultPingInterval = 60 * time.Second
	defaultPongTimeout  = 45 * time.Second
)

// OnReceiveFn is a callback func, which is called by the MConnection when a
// new message is received.
type OnReceiveFn = func(byte, []byte)

// MConnection is a multiplexed connection.
//
// __multiplex__ *noun* a system or signal involving simultaneous transmission
// of several messages along a single channel of communication.
//
// Each connection handles message transmission on multiple abstract
// communication streams. Each stream has a globally unique byte id. The byte
// id and the relative priorities of each stream are configured upon
// initialization of the connection.
//
// To open a stream, call OpenStream with the stream id. Remember that the
// stream id must be globally unique.
//
// Connection errors are communicated through the ErrorCh channel.
//
// Connection can be closed either by calling Close or FlushAndClose. If you
// need to flush all pending messages before closing the connection, call
// FlushAndClose. Otherwise, call Close.
type MConnection struct {
	service.BaseService

	conn          net.Conn
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flow.Monitor
	recvMonitor   *flow.Monitor
	send          chan struct{}
	pong          chan struct{}
	errorCh       chan error
	config        MConnConfig

	// Closing quitSendRoutine will cause the sendRoutine to eventually quit.
	// doneSendRoutine is closed when the sendRoutine actually quits.
	quitSendRoutine chan struct{}
	doneSendRoutine chan struct{}

	// Closing quitRecvRouting will cause the recvRouting to eventually quit.
	quitRecvRoutine chan struct{}

	flushTimer *timer.ThrottleTimer // flush writes as necessary but throttled.
	pingTimer  *time.Ticker         // send pings periodically

	// close conn if pong is not received in pongTimeout
	pongTimer     *time.Timer
	pongTimeoutCh chan bool // true - timeout, false - peer sent pong

	chStatsTimer *time.Ticker // update channel stats periodically

	created time.Time // time of creation

	_maxPacketMsgSize int

	// streamID -> channel
	channelsIdx map[byte]*stream

	// A map which stores the received messages. Used in tests.
	msgsByStreamIDMap map[byte]chan []byte

	onReceiveFn OnReceiveFn
}

var _ transport.Conn = (*MConnection)(nil)

// MConnConfig is a MConnection configuration.
type MConnConfig struct {
	SendRate int64 `mapstructure:"send_rate"`
	RecvRate int64 `mapstructure:"recv_rate"`

	// Maximum payload size
	MaxPacketMsgPayloadSize int `mapstructure:"max_packet_msg_payload_size"`

	// Interval to flush writes (throttled)
	FlushThrottle time.Duration `mapstructure:"flush_throttle"`

	// Interval to send pings
	PingInterval time.Duration `mapstructure:"ping_interval"`

	// Maximum wait time for pongs
	PongTimeout time.Duration `mapstructure:"pong_timeout"`

	// Fuzz connection
	TestFuzz       bool                   `mapstructure:"test_fuzz"`
	TestFuzzConfig *config.FuzzConnConfig `mapstructure:"test_fuzz_config"`
}

// DefaultMConnConfig returns the default config.
func DefaultMConnConfig() MConnConfig {
	return MConnConfig{
		SendRate:                defaultSendRate,
		RecvRate:                defaultRecvRate,
		MaxPacketMsgPayloadSize: defaultMaxPacketMsgPayloadSize,
		FlushThrottle:           defaultFlushThrottle,
		PingInterval:            defaultPingInterval,
		PongTimeout:             defaultPongTimeout,
	}
}

// NewMConnection wraps net.Conn and creates multiplex connection.
func NewMConnection(conn net.Conn, config MConnConfig) *MConnection {
	if config.PongTimeout >= config.PingInterval {
		panic("pongTimeout must be less than pingInterval (otherwise, next ping will reset pong timer)")
	}

	mconn := &MConnection{
		conn:              conn,
		bufConnReader:     bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter:     bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:       flow.New(0, 0),
		recvMonitor:       flow.New(0, 0),
		send:              make(chan struct{}, 1),
		pong:              make(chan struct{}, 1),
		errorCh:           make(chan error, 1),
		config:            config,
		created:           time.Now(),
		channelsIdx:       make(map[byte]*stream),
		msgsByStreamIDMap: make(map[byte]chan []byte),
	}

	mconn.BaseService = *service.NewBaseService(nil, "MConnection", mconn)

	// maxPacketMsgSize() is a bit heavy, so call just once
	mconn._maxPacketMsgSize = mconn.maxPacketMsgSize()

	return mconn
}

// OnReceive sets the callback function to be executed each time we read a message.
func (c *MConnection) OnReceive(fn OnReceiveFn) {
	c.onReceiveFn = fn
}

func (c *MConnection) SetLogger(l log.Logger) {
	c.BaseService.SetLogger(l)
}

// OnStart implements BaseService.
func (c *MConnection) OnStart() error {
	if err := c.BaseService.OnStart(); err != nil {
		return err
	}
	c.flushTimer = timer.NewThrottleTimer("flush", c.config.FlushThrottle)
	c.pingTimer = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTimer = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return nil
}

func (c *MConnection) Conn() net.Conn {
	return c.conn
}

// stopServices stops the BaseService and timers and closes the quitSendRoutine.
// if the quitSendRoutine was already closed, it returns true, otherwise it returns false.
func (c *MConnection) stopServices() (alreadyStopped bool) {
	select {
	case <-c.quitSendRoutine:
		// already quit
		return true
	default:
	}

	select {
	case <-c.quitRecvRoutine:
		// already quit
		return true
	default:
	}

	c.flushTimer.Stop()
	c.pingTimer.Stop()
	c.chStatsTimer.Stop()

	// inform the recvRouting that we are shutting down
	close(c.quitRecvRoutine)
	close(c.quitSendRoutine)
	return false
}

// ErrorCh returns a channel that will receive errors from the connection.
func (c *MConnection) ErrorCh() <-chan error {
	return c.errorCh
}

func (c *MConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *MConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// OpenStream opens a new stream on the connection. Remember that the
// stream id must be globally unique.
//
// Panics if the connection is already running (i.e., all streams
// must be registered in advance).
func (c *MConnection) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	if c.IsRunning() {
		panic("MConnection is already running. Please register all streams in advance")
	}

	c.Logger.Debug("Opening stream", "streamID", streamID, "desc", desc)

	if _, ok := c.channelsIdx[streamID]; ok {
		return nil, fmt.Errorf("stream %X already exists", streamID)
	}

	d := StreamDescriptor{
		ID:       streamID,
		Priority: 1,
	}
	if desc, ok := desc.(StreamDescriptor); ok {
		d = desc
	}
	c.channelsIdx[streamID] = newChannel(c, d)
	c.channelsIdx[streamID].SetLogger(c.Logger.With("streamID", streamID))
	// Allocate some buffer, otherwise CI tests will fail.
	c.msgsByStreamIDMap[streamID] = make(chan []byte, 5)

	return &MConnectionStream{conn: c, streamID: streamID}, nil
}

// HandshakeStream returns the underlying net.Conn connection.
func (c *MConnection) HandshakeStream() transport.HandshakeStream {
	return c.conn
}

// Close closes the connection. It flushes all pending writes before closing.
func (c *MConnection) Close(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	return c.conn.Close()
}

func (c *MConnection) FlushAndClose(reason string) error {
	if err := c.Stop(); err != nil {
		// If the connection was not fully started (an error occurred before the
		// peer was started), close the underlying connection.
		if errors.Is(err, service.ErrNotStarted) {
			return c.conn.Close()
		}
		return err
	}

	if c.stopServices() {
		return nil
	}

	// inform the error channel that we are shutting down.
	select {
	case c.errorCh <- errors.New(reason):
	default:
	}

	// flush all pending writes
	{
		// wait until the sendRoutine exits
		// so we dont race on calling sendSomePacketMsgs
		<-c.doneSendRoutine
		// Send and flush all pending msgs.
		// Since sendRoutine has exited, we can call this
		// safely
		w := protoio.NewDelimitedWriter(c.bufConnWriter)
		eof := c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		for !eof {
			eof = c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
		}
		_ = c.flush()
	}

	return c.conn.Close()
}

func (c *MConnection) ConnState() (state transport.ConnState) {
	state.ConnectedFor = time.Since(c.created)
	state.SendRateLimiterDelay = c.sendMonitor.Status().SleepTime
	state.RecvRateLimiterDelay = c.recvMonitor.Status().SleepTime
	state.StreamStates = make(map[byte]transport.StreamState)

	for streamID, channel := range c.channelsIdx {
		state.StreamStates[streamID] = transport.StreamState{
			SendQueueSize:     channel.loadSendQueueSize(),
			SendQueueCapacity: cap(channel.sendQueue),
		}
	}

	return state
}

func (c *MConnection) String() string {
	return fmt.Sprintf("MConn{%v}", c.conn.RemoteAddr())
}

func (c *MConnection) flush() error {
	return c.bufConnWriter.Flush()
}

// Catch panics, usually caused by remote disconnects.
func (c *MConnection) _recover() {
	if r := recover(); r != nil {
		c.Logger.Error("MConnection panicked", "err", r, "stack", string(debug.Stack()))
		c.Close(fmt.Sprintf("recovered from panic: %v", r))
	}
}

// thread-safe.
func (c *MConnection) sendBytes(chID byte, msgBytes []byte, blocking bool) error {
	if !c.IsRunning() {
		return nil
	}

	// Uncomment in you need to see raw bytes.
	// c.Logger.Debug("Send",
	// 	"streamID", chID,
	// 	"msgBytes", log.NewLazySprintf("%X", msgBytes),
	// 	"timeout", timeout)

	channel, ok := c.channelsIdx[chID]
	if !ok {
		panic(fmt.Sprintf("Unknown channel %X. Forgot to register?", chID))
	}
	if err := channel.sendBytes(msgBytes, blocking); err != nil {
		return err
	}

	// Wake up sendRoutine if necessary
	select {
	case c.send <- struct{}{}:
	default:
	}
	return nil
}

// CanSend returns true if you can send more data onto the chID, false
// otherwise. Use only as a heuristic.
//
// thread-safe.
func (c *MConnection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}

	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error(fmt.Sprintf("Unknown channel %X", chID))
		return false
	}
	return channel.canSend()
}

// sendRoutine polls for packets to send from channels.
func (c *MConnection) sendRoutine() {
	defer c._recover()

	protoWriter := protoio.NewDelimitedWriter(c.bufConnWriter)

FOR_LOOP:
	for {
		var _n int
		var err error
	SELECTION:
		select {
		case <-c.flushTimer.Ch:
			// NOTE: flushTimer.Set() must be called every time
			// something is written to .bufConnWriter.
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.chStatsTimer.C:
			for _, channel := range c.channelsIdx {
				channel.updateStats()
			}
		case <-c.pingTimer.C:
			c.Logger.Debug("Send Ping")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPing", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			c.Logger.Debug("Starting pong timer", "dur", c.config.PongTimeout)
			c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
				select {
				case c.pongTimeoutCh <- true:
				default:
				}
			})
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Debug("Pong timeout")
				err = errors.New("pong timeout")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			c.Logger.Debug("Send Pong")
			_n, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
			if err != nil {
				c.Logger.Error("Failed to send PacketPong", "err", err)
				break SELECTION
			}
			c.sendMonitor.Update(_n)
			if fErr := c.flush(); fErr != nil {
				c.Logger.Error("Failed to flush", "err", fErr)
			}
		case <-c.quitSendRoutine:
			break FOR_LOOP
		case <-c.send:
			// Send some PacketMsgs
			eof := c.sendSomePacketMsgs(protoWriter)
			if !eof {
				// Keep sendRoutine awake.
				select {
				case c.send <- struct{}{}:
				default:
				}
			}
		}

		if !c.IsRunning() {
			break FOR_LOOP
		}
		if err != nil {
			c.Logger.Error("Connection failed @ sendRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

// Returns true if messages from channels were exhausted.
// Blocks in accordance to .sendMonitor throttling.
func (c *MConnection) sendSomePacketMsgs(w protoio.Writer) bool {
	// Block until .sendMonitor says we can write.
	// Once we're ready we send more than we asked for,
	// but amortized it should even out.
	c.sendMonitor.Limit(c._maxPacketMsgSize, c.config.SendRate, true)

	// Now send some PacketMsgs.
	return c.sendBatchPacketMsgs(w, numBatchPacketMsgs)
}

// Returns true if messages from channels were exhausted.
func (c *MConnection) sendBatchPacketMsgs(w protoio.Writer, batchSize int) bool {
	// Send a batch of PacketMsgs.
	totalBytesWritten := 0
	defer func() {
		if totalBytesWritten > 0 {
			c.sendMonitor.Update(totalBytesWritten)
		}
	}()
	for i := 0; i < batchSize; i++ {
		channel := c.selectChannel()
		// nothing to send across any channel.
		if channel == nil {
			return true
		}
		bytesWritten, err := c.sendPacketMsgOnChannel(w, channel)
		if err {
			return true
		}
		totalBytesWritten += bytesWritten
	}
	return false
}

// selects a channel to gossip our next message on.
// TODO: Make "batchChannelToGossipOn", so we can do our proto marshaling overheads in parallel,
// and we can avoid re-checking for `isSendPending`.
// We can easily mock the recentlySent differences for the batch choosing.
func (c *MConnection) selectChannel() *stream {
	// Choose a channel to create a PacketMsg from.
	// The chosen channel will be the one whose recentlySent/priority is the least.
	var leastRatio float32 = math.MaxFloat32
	var leastChannel *stream
	for _, channel := range c.channelsIdx {
		// If nothing to send, skip this channel
		// TODO: Skip continually looking for isSendPending on channels we've already skipped in this batch-send.
		if !channel.isSendPending() {
			continue
		}
		// Get ratio, and keep track of lowest ratio.
		// TODO: RecentlySent right now is bytes. This should be refactored to num messages to fix
		// gossip prioritization bugs.
		ratio := float32(channel.recentlySent) / float32(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			leastChannel = channel
		}
	}
	return leastChannel
}

// returns (num_bytes_written, error_occurred).
func (c *MConnection) sendPacketMsgOnChannel(w protoio.Writer, sendChannel *stream) (int, bool) {
	// Make & send a PacketMsg from this channel
	n, err := sendChannel.writePacketMsgTo(w)
	if err != nil {
		c.Logger.Error("Failed to write PacketMsg", "err", err)
		c.Close(err.Error())
		return n, true
	}
	// TODO: Change this to only add flush signals at the start and end of the batch.
	c.flushTimer.Set()
	return n, false
}

// recvRoutine reads PacketMsgs and reconstructs the message using the
// channels' "recving" buffer. After a whole message has been assembled, it's
// pushed to an internal queue, which is accessible via Read. Blocks depending
// on how the connection is throttled. Otherwise, it never blocks.
func (c *MConnection) recvRoutine() {
	defer c._recover()

	protoReader := protoio.NewDelimitedReader(c.bufConnReader, c._maxPacketMsgSize)

FOR_LOOP:
	for {
		// Block until .recvMonitor says we can read.
		c.recvMonitor.Limit(c._maxPacketMsgSize, atomic.LoadInt64(&c.config.RecvRate), true)

		// Peek into bufConnReader for debugging
		/*
			if numBytes := c.bufConnReader.Buffered(); numBytes > 0 {
				bz, err := c.bufConnReader.Peek(cmtmath.MinInt(numBytes, 100))
				if err == nil {
					// return
				} else {
					c.Logger.Debug("Error peeking connection buffer", "err", err)
					// return nil
				}
				c.Logger.Info("Peek connection buffer", "numBytes", numBytes, "bz", bz)
			}
		*/

		// Read packet type
		var packet tmp2p.Packet

		_n, err := protoReader.ReadMsg(&packet)
		c.recvMonitor.Update(_n)
		if err != nil {
			// stopServices was invoked and we are shutting down
			// receiving is expected to fail since we will close the connection
			select {
			case <-c.quitRecvRoutine:
				break FOR_LOOP
			default:
			}

			if c.IsRunning() {
				if errors.Is(err, io.EOF) {
					c.Logger.Info("Connection is closed @ recvRoutine (likely by the other side)")
				} else {
					c.Logger.Debug("Connection failed @ recvRoutine (reading byte)", "err", err)
				}
				c.Close(err.Error())
			}
			break FOR_LOOP
		}

		// Read more depending on packet type.
		switch pkt := packet.Sum.(type) {
		case *tmp2p.Packet_PacketPing:
			// TODO: prevent abuse, as they cause flush()'s.
			// https://github.com/tendermint/tendermint/issues/1190
			c.Logger.Debug("Receive Ping")
			select {
			case c.pong <- struct{}{}:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketPong:
			c.Logger.Debug("Receive Pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
				// never block
			}
		case *tmp2p.Packet_PacketMsg:
			channelID := byte(pkt.PacketMsg.ChannelID)
			channel, ok := c.channelsIdx[channelID]
			if !ok || pkt.PacketMsg.ChannelID < 0 || pkt.PacketMsg.ChannelID > math.MaxUint8 {
				err := fmt.Errorf("unknown channel %X", pkt.PacketMsg.ChannelID)
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}

			msgBytes, err := channel.recvPacketMsg(*pkt.PacketMsg)
			if err != nil {
				c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
				c.Close(err.Error())
				break FOR_LOOP
			}
			if msgBytes != nil {
				// Uncomment in you need to see raw bytes.
				// c.Logger.Debug("Received", "streamID", channelID, "msgBytes", log.NewLazySprintf("%X", msgBytes))
				if c.onReceiveFn != nil {
					c.onReceiveFn(channelID, msgBytes)
				} else {
					bz := make([]byte, len(msgBytes))
					copy(bz, msgBytes)
					c.msgsByStreamIDMap[channelID] <- bz
				}
			}
		default:
			err := fmt.Errorf("unknown message type %v", reflect.TypeOf(packet))
			c.Logger.Debug("Connection failed @ recvRoutine", "err", err)
			c.Close(err.Error())
			break FOR_LOOP
		}
	}

	// Cleanup
	close(c.pong)
}

// Used in tests.
func (c *MConnection) readBytes(streamID byte, b []byte, timeout time.Duration) (n int, err error) {
	select {
	case msgBytes := <-c.msgsByStreamIDMap[streamID]:
		n = copy(b, msgBytes)
		if n < len(msgBytes) {
			err = errors.New("short buffer")
			return 0, err
		}
		return n, nil
	case <-time.After(timeout):
		return 0, errors.New("read timeout")
	}
}

// not goroutine-safe.
func (c *MConnection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize returns a maximum size of PacketMsg.
func (c *MConnection) maxPacketMsgSize() int {
	bz, err := proto.Marshal(mustWrapPacket(&tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, c.config.MaxPacketMsgPayloadSize),
	}))
	if err != nil {
		panic(err)
	}
	return len(bz)
}

// -----------------------------------------------------------------------------

// NOTE: not goroutine-safe.
type stream struct {
	conn          *MConnection
	desc          StreamDescriptor
	sendQueue     chan []byte
	sendQueueSize int32 // atomic.
	recving       []byte
	sending       []byte
	recentlySent  int64 // exponential moving average

	nextPacketMsg           *tmp2p.PacketMsg
	nextP2pWrapperPacketMsg *tmp2p.Packet_PacketMsg
	nextPacket              *tmp2p.Packet

	maxPacketMsgPayloadSize int

	Logger log.Logger
}

func newChannel(conn *MConnection, desc StreamDescriptor) *stream {
	desc = desc.FillDefaults()
	if desc.Priority <= 0 {
		panic("Channel default priority must be a positive integer")
	}
	return &stream{
		conn:                    conn,
		desc:                    desc,
		sendQueue:               make(chan []byte, desc.SendQueueCapacity),
		recving:                 make([]byte, 0, desc.RecvBufferCapacity),
		nextPacketMsg:           &tmp2p.PacketMsg{ChannelID: int32(desc.ID)},
		nextP2pWrapperPacketMsg: &tmp2p.Packet_PacketMsg{},
		nextPacket:              &tmp2p.Packet{},
		maxPacketMsgPayloadSize: conn.config.MaxPacketMsgPayloadSize,
	}
}

func (ch *stream) SetLogger(l log.Logger) {
	ch.Logger = l
}

// Queues message to send to this channel. Blocks if blocking is true.
// thread-safe.
func (ch *stream) sendBytes(bytes []byte, blocking bool) error {
	if blocking {
		select {
		case ch.sendQueue <- bytes:
			atomic.AddInt32(&ch.sendQueueSize, 1)
			return nil
		case <-ch.conn.Quit():
			return nil
		}
	}

	select {
	case ch.sendQueue <- bytes:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return nil
	default:
		return ErrWriteQueueFull{}
	case <-ch.conn.Quit():
		return nil
	}
}

// Goroutine-safe.
func (ch *stream) loadSendQueueSize() (size int) {
	return int(atomic.LoadInt32(&ch.sendQueueSize))
}

// Goroutine-safe
// Use only as a heuristic.
func (ch *stream) canSend() bool {
	return ch.loadSendQueueSize() < defaultSendQueueCapacity
}

// Returns true if any PacketMsgs are pending to be sent.
// Call before calling updateNextPacket
// Goroutine-safe.
func (ch *stream) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// Updates the nextPacket proto message for us to send.
// Not goroutine-safe.
func (ch *stream) updateNextPacket() {
	maxSize := ch.maxPacketMsgPayloadSize
	if len(ch.sending) <= maxSize {
		ch.nextPacketMsg.Data = ch.sending
		ch.nextPacketMsg.EOF = true
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // decrement sendQueueSize
	} else {
		ch.nextPacketMsg.Data = ch.sending[:maxSize]
		ch.nextPacketMsg.EOF = false
		ch.sending = ch.sending[maxSize:]
	}

	ch.nextP2pWrapperPacketMsg.PacketMsg = ch.nextPacketMsg
	ch.nextPacket.Sum = ch.nextP2pWrapperPacketMsg
}

// Writes next PacketMsg to w and updates c.recentlySent.
// Not goroutine-safe.
func (ch *stream) writePacketMsgTo(w protoio.Writer) (n int, err error) {
	ch.updateNextPacket()
	n, err = w.WriteMsg(ch.nextPacket)
	if err != nil {
		err = ErrPacketWrite{Source: err}
	}

	atomic.AddInt64(&ch.recentlySent, int64(n))
	return n, err
}

// Handles incoming PacketMsgs. It returns a message bytes if message is
// complete. NOTE message bytes may change on next call to recvPacketMsg.
// Not goroutine-safe.
func (ch *stream) recvPacketMsg(packet tmp2p.PacketMsg) ([]byte, error) {
	recvCap, recvReceived := ch.desc.RecvMessageCapacity, len(ch.recving)+len(packet.Data)
	if recvCap < recvReceived {
		return nil, ErrPacketTooBig{Max: recvCap, Received: recvReceived}
	}

	ch.recving = append(ch.recving, packet.Data...)
	if packet.EOF {
		msgBytes := ch.recving

		// clear the slice without re-allocating.
		// http://stackoverflow.com/questions/16971741/how-do-you-clear-a-slice-in-go
		//   suggests this could be a memory leak, but we might as well keep the memory for the channel until it closes,
		//	at which point the recving slice stops being used and should be garbage collected
		ch.recving = ch.recving[:0] // make([]byte, 0, ch.desc.RecvBufferCapacity)
		return msgBytes, nil
	}
	return nil, nil
}

// Call this periodically to update stats for throttling purposes.
// thread-safe.
func (ch *stream) updateStats() {
	// Exponential decay of stats.
	// TODO: optimize.
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

// ----------------------------------------
// Packet

// mustWrapPacket takes a packet kind (oneof) and wraps it in a tmp2p.Packet message.
func mustWrapPacket(pb proto.Message) *tmp2p.Packet {
	msg := &tmp2p.Packet{}
	mustWrapPacketInto(pb, msg)
	return msg
}

func mustWrapPacketInto(pb proto.Message, dst *tmp2p.Packet) {
	switch pb := pb.(type) {
	case *tmp2p.PacketPing:
		dst.Sum = &tmp2p.Packet_PacketPing{
			PacketPing: pb,
		}
	case *tmp2p.PacketPong:
		dst.Sum = &tmp2p.Packet_PacketPong{
			PacketPong: pb,
		}
	case *tmp2p.PacketMsg:
		dst.Sum = &tmp2p.Packet_PacketMsg{
			PacketMsg: pb,
		}
	default:
		panic(fmt.Errorf("unknown packet type %T", pb))
	}
}

```

---

<a name="file-11"></a>

### File: `tcp/conn/connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	pbtypes "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/protoio"
)

const (
	maxPingPongPacketSize = 1024 // bytes
	testStreamID          = 0x01
)

func createMConnectionWithSingleStream(t *testing.T, conn net.Conn) (*MConnection, *MConnectionStream) {
	t.Helper()

	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	c := NewMConnection(conn, cfg)
	c.SetLogger(log.TestingLogger())

	stream, err := c.OpenStream(testStreamID, nil)
	require.NoError(t, err)

	return c, stream.(*MConnectionStream)
}

func TestMConnection_Flush(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	clientConn, clientStream := createMConnectionWithSingleStream(t, client)
	err := clientConn.Start()
	require.NoError(t, err)

	msg := []byte("abc")
	n, err := clientStream.Write(msg)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	// start the reader in a new routine, so we can flush
	errCh := make(chan error)
	go func() {
		buf := make([]byte, 100) // msg + ping
		_, err := server.Read(buf)
		errCh <- err
	}()

	// stop the conn - it should flush all conns
	err = clientConn.FlushAndClose("test")
	require.NoError(t, err)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Error reading from server: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for msgs to be read")
	}
}

func TestMConnection_StreamWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, clientStream := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	msg := []byte("Ant-Man")
	_, err = clientStream.Write(msg)
	require.NoError(t, err)
	// NOTE: subsequent writes could pass because we are reading from
	// the send queue in a separate goroutine.
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
	assert.True(t, mconn.CanSend(testStreamID))

	msg = []byte("Spider-Man")
	require.NoError(t, err)
	_, err = clientStream.TryWrite(msg)
	require.NoError(t, err)
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
}

func TestMConnection_StreamReadWrite(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn1, stream1 := createMConnectionWithSingleStream(t, client)
	err := mconn1.Start()
	require.NoError(t, err)
	defer mconn1.Close("normal")

	mconn2, stream2 := createMConnectionWithSingleStream(t, server)
	err = mconn2.Start()
	require.NoError(t, err)
	defer mconn2.Close("normal")

	// => write
	msg := []byte("Cyclops")
	_, err = stream1.Write(msg)
	require.NoError(t, err)

	// => read
	buf := make([]byte, len(msg))
	n, err := stream2.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)
	assert.Equal(t, msg, buf)
}

func TestMConnectionStatus(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	state := mconn.ConnState()
	assert.NotNil(t, state)
	assert.Zero(t, state.StreamStates[testStreamID].SendQueueSize)
}

func TestMConnection_PongTimeoutResultsInError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		// read ping
		var pkt tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 200*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
	case <-time.After(pongTimerExpired):
		t.Fatalf("Expected to receive error after %v", pongTimerExpired)
	}
}

func TestMConnection_MultiplePongsInTheBeginning(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pongs in a row (abuse)
	protoWriter := protoio.NewDelimitedWriter(server)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
	require.NoError(t, err)

	serverGotPing := make(chan struct{})
	go func() {
		// read ping (one byte)
		var packet tmp2p.Packet
		_, err := protoio.NewDelimitedReader(server, maxPingPongPacketSize).ReadMsg(&packet)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing

	pongTimerExpired := mconn.config.PongTimeout + 20*time.Millisecond
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_MultiplePings(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	// sending 3 pings in a row (abuse)
	// see https://github.com/tendermint/tendermint/issues/1190
	protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
	protoWriter := protoio.NewDelimitedWriter(server)
	var pkt tmp2p.Packet

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPing{}))
	require.NoError(t, err)

	_, err = protoReader.ReadMsg(&pkt)
	require.NoError(t, err)

	assert.True(t, mconn.IsRunning())
}

func TestMConnection_PingPongs(t *testing.T) {
	// check that we are not leaking any go-routines
	defer leaktest.CheckTimeout(t, 10*time.Second)()

	server, client := net.Pipe()

	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	serverGotPing := make(chan struct{})
	go func() {
		protoReader := protoio.NewDelimitedReader(server, maxPingPongPacketSize)
		protoWriter := protoio.NewDelimitedWriter(server)
		var pkt tmp2p.PacketPing

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)

		time.Sleep(mconn.config.PingInterval)

		// read ping
		_, err = protoReader.ReadMsg(&pkt)
		require.NoError(t, err)
		serverGotPing <- struct{}{}

		// respond with pong
		_, err = protoWriter.WriteMsg(mustWrapPacket(&tmp2p.PacketPong{}))
		require.NoError(t, err)
	}()
	<-serverGotPing
	<-serverGotPing

	pongTimerExpired := (mconn.config.PongTimeout + 20*time.Millisecond) * 2
	select {
	case err := <-mconn.ErrorCh():
		t.Fatalf("Expected no error, but got %v", err)
	case <-time.After(2 * pongTimerExpired):
		assert.True(t, mconn.IsRunning())
	}
}

func TestMConnection_StopsAndReturnsError(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	mconn, _ := createMConnectionWithSingleStream(t, client)
	err := mconn.Start()
	require.NoError(t, err)
	defer mconn.Close("normal")

	if err := client.Close(); err != nil {
		t.Error(err)
	}

	select {
	case err := <-mconn.ErrorCh():
		assert.NotNil(t, err)
		assert.False(t, mconn.IsRunning())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Did not receive error in 500ms")
	}
}

//nolint:unparam
func newClientAndServerConnsForReadErrors(t *testing.T) (*MConnection, *MConnectionStream, *MConnection, *MConnectionStream) {
	t.Helper()
	server, client := net.Pipe()

	// create client conn with two channels
	cfg := DefaultMConnConfig()
	cfg.PingInterval = 90 * time.Millisecond
	cfg.PongTimeout = 45 * time.Millisecond
	mconnClient := NewMConnection(client, cfg)
	clientStream, err := mconnClient.OpenStream(testStreamID, StreamDescriptor{ID: testStreamID, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	// create another channel
	_, err = mconnClient.OpenStream(0x02, StreamDescriptor{ID: 0x02, Priority: 1, SendQueueCapacity: 1})
	require.NoError(t, err)
	mconnClient.SetLogger(log.TestingLogger().With("module", "client"))
	err = mconnClient.Start()
	require.NoError(t, err)

	// create server conn with 1 channel
	// it fires on chOnErr when there's an error
	serverLogger := log.TestingLogger().With("module", "server")
	mconnServer, serverStream := createMConnectionWithSingleStream(t, server)
	mconnServer.SetLogger(serverLogger)
	err = mconnServer.Start()
	require.NoError(t, err)

	return mconnClient, clientStream.(*MConnectionStream), mconnServer, serverStream
}

func assertBytes(t *testing.T, s *MConnectionStream, want []byte) {
	t.Helper()

	buf := make([]byte, len(want))
	n, err := s.Read(buf)
	require.NoError(t, err)
	if assert.Equal(t, len(want), n) {
		assert.Equal(t, want, buf)
	}
}

func gotError(ch <-chan error) bool {
	after := time.After(time.Second * 5)
	select {
	case <-ch:
		return true
	case <-after:
		return false
	}
}

func TestMConnection_ReadErrorBadEncoding(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send badly encoded data
	client := mconnClient.conn
	_, err := client.Write([]byte{1, 2, 3, 4, 5})
	require.NoError(t, err)

	assert.True(t, gotError(mconnServer.ErrorCh()), "badly encoded msgPacket")
}

// func TestMConnection_ReadErrorUnknownChannel(t *testing.T) {
// 	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
// 	defer mconnClient.Close("normal")
// 	defer mconnServer.Close("normal")

// 	msg := []byte("Ant-Man")

// 	// send msg that has unknown channel
// 	client := mconnClient.conn
// 	protoWriter := protoio.NewDelimitedWriter(client)
// 	packet := tmp2p.PacketMsg{
// 		ChannelID: 0x03,
// 		EOF:       true,
// 		Data:      msg,
// 	}
// 	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
// 	require.NoError(t, err)

// 	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown channel")
// }

func TestMConnection_ReadErrorLongMessage(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	msg := make([]byte, mconnClient.config.MaxPacketMsgPayloadSize)
	packet := tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      msg,
	}

	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, msg)

	// send msg that's too long
	packet = tmp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, mconnClient.config.MaxPacketMsgPayloadSize+100),
	}

	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.Error(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "msg too long")
}

func TestMConnection_ReadErrorUnknownMsgType(t *testing.T) {
	mconnClient, _, mconnServer, _ := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	// send msg with unknown msg type
	_, err := protoio.NewDelimitedWriter(mconnClient.conn).WriteMsg(&pbtypes.Header{ChainID: "x"})
	require.NoError(t, err)
	assert.True(t, gotError(mconnServer.ErrorCh()), "unknown msg type")
}

//nolint:lll //ignore line length for tests
func TestConnVectors(t *testing.T) {
	testCases := []struct {
		testName string
		msg      proto.Message
		expBytes string
	}{
		{"PacketPing", &tmp2p.PacketPing{}, "0a00"},
		{"PacketPong", &tmp2p.PacketPong{}, "1200"},
		{"PacketMsg", &tmp2p.PacketMsg{ChannelID: 1, EOF: false, Data: []byte("data transmitted over the wire")}, "1a2208011a1e64617461207472616e736d6974746564206f766572207468652077697265"},
	}

	for _, tc := range testCases {
		pm := mustWrapPacket(tc.msg)
		bz, err := pm.Marshal()
		require.NoError(t, err, tc.testName)

		require.Equal(t, tc.expBytes, hex.EncodeToString(bz), tc.testName)
	}
}

func TestMConnection_ChannelOverflow(t *testing.T) {
	mconnClient, _, mconnServer, serverStream := newClientAndServerConnsForReadErrors(t)
	defer mconnClient.Close("normal")
	defer mconnServer.Close("normal")

	client := mconnClient.conn
	protoWriter := protoio.NewDelimitedWriter(client)

	// send msg that's just right
	packet := tmp2p.PacketMsg{
		ChannelID: testStreamID,
		EOF:       true,
		Data:      []byte(`42`),
	}
	_, err := protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
	assertBytes(t, serverStream, []byte(`42`))

	// channel ID that's too large
	packet.ChannelID = int32(1025)
	_, err = protoWriter.WriteMsg(mustWrapPacket(&packet))
	require.NoError(t, err)
}

```

---

<a name="file-12"></a>

### File: `tcp/conn/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/p2p/transport"
)

var (
	ErrInvalidSecretConnKeySend = errors.New("send invalid secret connection key")
	ErrInvalidSecretConnKeyRecv = errors.New("invalid receive SecretConnection Key")
	ErrChallengeVerification    = errors.New("challenge verification failed")

	// ErrTimeout is returned when a read or write operation times out.
	ErrTimeout = errors.New("read/write timeout")
)

// ErrWriteQueueFull is returned when the write queue is full.
type ErrWriteQueueFull struct{}

var _ transport.WriteError = ErrWriteQueueFull{}

func (ErrWriteQueueFull) Error() string {
	return "write queue is full"
}

func (ErrWriteQueueFull) Full() bool {
	return true
}

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	Source error
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet message: %v", e.Source)
}

func (e ErrPacketWrite) Unwrap() error {
	return e.Source
}

type ErrUnexpectedPubKeyType struct {
	Expected string
	Got      string
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptFrame struct {
	Source error
}

func (e ErrDecryptFrame) Error() string {
	return fmt.Sprintf("SecretConnection: failed to decrypt the frame: %v", e.Source)
}

func (e ErrDecryptFrame) Unwrap() error {
	return e.Source
}

type ErrPacketTooBig struct {
	Received int
	Max      int
}

func (e ErrPacketTooBig) Error() string {
	return fmt.Sprintf("received message exceeds available capacity (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkTooBig struct {
	Received int
	Max      int
}

func (e ErrChunkTooBig) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}

```

---

<a name="file-13"></a>

### File: `tcp/conn/evil_secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 8 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/protoio"
)

type buffer struct {
	next bytes.Buffer
}

func (b *buffer) Read(data []byte) (n int, err error) {
	return b.next.Read(data)
}

func (b *buffer) Write(data []byte) (n int, err error) {
	return b.next.Write(data)
}

func (b *buffer) Bytes() []byte {
	return b.next.Bytes()
}

func (*buffer) Close() error {
	return nil
}

type evilConn struct {
	secretConn *SecretConnection
	buffer     *buffer

	locEphPub  *[32]byte
	locEphPriv *[32]byte
	remEphPub  *[32]byte
	privKey    crypto.PrivKey

	readStep   int
	writeStep  int
	readOffset int

	shareEphKey        bool
	badEphKey          bool
	shareAuthSignature bool
	badAuthSignature   bool
}

func newEvilConn(shareEphKey, badEphKey, shareAuthSignature, badAuthSignature bool) *evilConn {
	privKey := ed25519.GenPrivKey()
	locEphPub, locEphPriv := genEphKeys()
	var rep [32]byte
	c := &evilConn{
		locEphPub:  locEphPub,
		locEphPriv: locEphPriv,
		remEphPub:  &rep,
		privKey:    privKey,

		shareEphKey:        shareEphKey,
		badEphKey:          badEphKey,
		shareAuthSignature: shareAuthSignature,
		badAuthSignature:   badAuthSignature,
	}

	return c
}

func (c *evilConn) Read(data []byte) (n int, err error) {
	if !c.shareEphKey {
		return 0, io.EOF
	}

	switch c.readStep {
	case 0:
		if !c.badEphKey {
			lc := *c.locEphPub
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				panic(err)
			}
			copy(data, bz[c.readOffset:])
			n = len(data)
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("drop users;")})
			if err != nil {
				panic(err)
			}
			copy(data, bz)
			n = len(data)
		}
		c.readOffset += n

		if n >= 32 {
			c.readOffset = 0
			c.readStep = 1
			if !c.shareAuthSignature {
				c.readStep = 2
			}
		}

		return n, nil
	case 1:
		signature := c.signChallenge()
		if !c.badAuthSignature {
			pkpb, err := cryptoenc.PubKeyToProto(c.privKey.PubKey())
			if err != nil {
				panic(err)
			}
			bz, err := protoio.MarshalDelimited(&tmp2p.AuthSigMessage{PubKey: pkpb, Sig: signature})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes()[c.readOffset:])
		} else {
			bz, err := protoio.MarshalDelimited(&gogotypes.BytesValue{Value: []byte("select * from users;")})
			if err != nil {
				panic(err)
			}
			n, err = c.secretConn.Write(bz)
			if err != nil {
				panic(err)
			}
			if c.readOffset > len(c.buffer.Bytes()) {
				return len(data), nil
			}
			copy(data, c.buffer.Bytes())
		}
		c.readOffset += len(data)
		return n, nil
	default:
		return 0, io.EOF
	}
}

func (c *evilConn) Write(data []byte) (n int, err error) {
	switch c.writeStep {
	case 0:
		var (
			bytes     gogotypes.BytesValue
			remEphPub [32]byte
		)
		err := protoio.UnmarshalDelimited(data, &bytes)
		if err != nil {
			panic(err)
		}
		copy(remEphPub[:], bytes.Value)
		c.remEphPub = &remEphPub
		c.writeStep = 1
		if !c.shareAuthSignature {
			c.writeStep = 2
		}
		return len(data), nil
	case 1:
		// Signature is not needed, therefore skipped.
		return len(data), nil
	default:
		return 0, io.EOF
	}
}

func (*evilConn) Close() error {
	return nil
}

func (c *evilConn) signChallenge() []byte {
	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(c.locEphPub, c.remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least, lexicographically
	// sorted.
	locIsLeast := bytes.Equal(c.locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(c.remEphPub, c.locEphPriv)
	if err != nil {
		panic(err)
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via HKDF-SHA2
	// on the transcript state (which itself also uses HKDF-SHA2 to derive a key
	// from the dhSecret).
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		panic(errors.New("invalid send SecretConnection Key"))
	}
	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		panic(errors.New("invalid receive SecretConnection Key"))
	}

	b := &buffer{}
	c.secretConn = &SecretConnection{
		conn:            b,
		connWriter:      bufio.NewWriterSize(b, defaultWriteBufferSize),
		connReader:      b,
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, totalFrameSize+aeadSizeOverhead),
	}
	c.buffer = b

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, c.privKey)
	if err != nil {
		panic(err)
	}

	return locSignature
}

// TestMakeSecretConnection creates an evil connection and tests that
// MakeSecretConnection errors at different stages.
func TestMakeSecretConnection(t *testing.T) {
	testCases := []struct {
		name       string
		conn       *evilConn
		checkError func(error) bool // Function to check if the error matches the expectation
	}{
		{"refuse to share ethimeral key", newEvilConn(false, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad ethimeral key", newEvilConn(true, true, false, false), func(err error) bool { return assert.Contains(t, err.Error(), "wrong wireType") }},
		{"refuse to share auth signature", newEvilConn(true, false, false, false), func(err error) bool { return errors.Is(err, io.EOF) }},
		{"share bad auth signature", newEvilConn(true, false, true, true), func(err error) bool { return errors.As(err, &ErrDecryptFrame{}) }},
		// fails with the introduction of changes PR #3419
		// {"all good", newEvilConn(true, false, true, false), func(err error) bool { return err == nil }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privKey := ed25519.GenPrivKey()
			_, err := MakeSecretConnection(tc.conn, privKey)
			if tc.checkError != nil {
				assert.True(t, tc.checkError(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

```

---

<a name="file-14"></a>

### File: `tcp/conn/secret_connection.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 14 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/oasisprotocol/curve25519-voi/primitives/merlin"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/nacl/box"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/internal/async"
	"github.com/cometbft/cometbft/libs/protoio"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// 4 + 1024 == 1028 total frame size.
const (
	dataLenSize      = 4
	dataMaxSize      = 1024
	totalFrameSize   = dataMaxSize + dataLenSize
	aeadSizeOverhead = 16 // overhead of poly 1305 authentication tag
	aeadKeySize      = chacha20poly1305.KeySize
	aeadNonceSize    = chacha20poly1305.NonceSize

	labelEphemeralLowerPublicKey = "EPHEMERAL_LOWER_PUBLIC_KEY"
	labelEphemeralUpperPublicKey = "EPHEMERAL_UPPER_PUBLIC_KEY"
	labelDHSecret                = "DH_SECRET"
	labelSecretConnectionMac     = "SECRET_CONNECTION_MAC"

	defaultWriteBufferSize = 128 * 1024
	// try to read the biggest logical packet we can get, in one read.
	// biggest logical packet is encoding_overhead(64kb).
	defaultReadBufferSize = 65 * 1024
)

var (
	ErrSmallOrderRemotePubKey    = errors.New("detected low order point from remote peer")
	secretConnKeyAndChallengeGen = []byte("TENDERMINT_SECRET_CONNECTION_KEY_AND_CHALLENGE_GEN")
)

// SecretConnection implements net.Conn.
// It is an implementation of the STS protocol.
// For more details regarding this implementation of the STS protocol, please refer to:
// https://github.com/cometbft/cometbft/blob/main/spec/p2p/legacy-docs/peer.md#authenticated-encryption-handshake.
//
// The original STS protocol, which inspired this implementation:
// https://citeseerx.ist.psu.edu/document?rapid=rep1&type=pdf&doi=b852bc961328ce74f7231a4b569eec1ab6c3cf50. # codespell:ignore
//
// Consumers of the SecretConnection are responsible for authenticating
// the remote peer's pubkey against known information, like a nodeID.
type SecretConnection struct {
	// immutable
	recvAead cipher.AEAD
	sendAead cipher.AEAD

	remPubKey crypto.PubKey

	conn       io.ReadWriteCloser
	connWriter *bufio.Writer
	connReader io.Reader

	// net.Conn must be thread safe:
	// https://golang.org/pkg/net/#Conn.
	// Since we have internal mutable state,
	// we need mtxs. But recv and send states
	// are independent, so we can use two mtxs.
	// All .Read are covered by recvMtx,
	// all .Write are covered by sendMtx.
	recvMtx         cmtsync.Mutex
	recvBuffer      []byte
	recvNonce       *[aeadNonceSize]byte
	recvFrame       []byte
	recvSealedFrame []byte

	sendMtx         cmtsync.Mutex
	sendNonce       *[aeadNonceSize]byte
	sendFrame       []byte
	sendSealedFrame []byte
}

// MakeSecretConnection performs handshake and returns a new authenticated
// SecretConnection.
// Returns nil if there is an error in handshake.
// Caller should call conn.Close().
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {
	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	// NOTE: every 32-byte string is accepted as a Curve25519 public key (see
	// DJB's Curve25519 paper: http://cr.yp.to/ecdh/curve25519-20060209.pdf)
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, err
	}

	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	transcript := merlin.NewTranscript("TENDERMINT_SECRET_CONNECTION_TRANSCRIPT_HASH")

	transcript.AppendMessage(labelEphemeralLowerPublicKey, loEphPub[:])
	transcript.AppendMessage(labelEphemeralUpperPublicKey, hiEphPub[:])

	// Check if the local ephemeral public key was the least,
	// lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Compute common diffie hellman secret using X25519.
	dhSecret, err := computeDHSecret(remEphPub, locEphPriv)
	if err != nil {
		return nil, err
	}

	transcript.AppendMessage(labelDHSecret, dhSecret[:])

	// Generate the secret used for receiving, sending, challenge via
	// HKDF-SHA2 on the dhSecret.
	recvSecret, sendSecret := deriveSecrets(dhSecret, locIsLeast)

	const challengeSize = 32
	var challenge [challengeSize]byte
	transcript.ExtractBytes(challenge[:], labelSecretConnectionMac)

	sendAead, err := chacha20poly1305.New(sendSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeySend
	}

	recvAead, err := chacha20poly1305.New(recvSecret[:])
	if err != nil {
		return nil, ErrInvalidSecretConnKeyRecv
	}

	sc := &SecretConnection{
		conn:            conn,
		connWriter:      bufio.NewWriterSize(conn, defaultWriteBufferSize),
		connReader:      bufio.NewReaderSize(conn, defaultReadBufferSize),
		recvBuffer:      nil,
		recvNonce:       new([aeadNonceSize]byte),
		sendNonce:       new([aeadNonceSize]byte),
		recvAead:        recvAead,
		sendAead:        sendAead,
		recvFrame:       make([]byte, totalFrameSize),
		recvSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
		sendFrame:       make([]byte, totalFrameSize),
		sendSealedFrame: make([]byte, aeadSizeOverhead+totalFrameSize),
	}

	// Sign the challenge bytes for authentication.
	locSignature, err := signChallenge(&challenge, locPrivKey)
	if err != nil {
		return nil, err
	}

	// Share (in secret) each other's pubkey & challenge signature
	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, err
	}

	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	// Usage in your function
	if _, ok := remPubKey.(ed25519.PubKey); !ok {
		return nil, ErrUnexpectedPubKeyType{
			Expected: ed25519.KeyType,
			Got:      remPubKey.Type(),
		}
	}

	if !remPubKey.VerifySignature(challenge[:], remSignature) {
		return nil, ErrChallengeVerification
	}

	// We've authorized.
	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey returns authenticated remote pubkey.
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Writes encrypted frames of `totalFrameSize + aeadSizeOverhead`.
// CONTRACT: data smaller than dataMaxSize is written atomically.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	sc.sendMtx.Lock()
	defer sc.sendMtx.Unlock()
	sealedFrame, frame := sc.sendSealedFrame, sc.sendFrame

	for 0 < len(data) {
		if err := func() error {
			var chunk []byte
			if dataMaxSize < len(data) {
				chunk = data[:dataMaxSize]
				data = data[dataMaxSize:]
			} else {
				chunk = data
				data = nil
			}
			chunkLength := len(chunk)
			binary.LittleEndian.PutUint32(frame, uint32(chunkLength))
			copy(frame[dataLenSize:], chunk)

			// encrypt the frame
			sc.sendAead.Seal(sealedFrame[:0], sc.sendNonce[:], frame, nil)
			incrNonce(sc.sendNonce)
			// end encryption

			_, err = sc.connWriter.Write(sealedFrame)
			if err != nil {
				return err
			}

			n += len(chunk)
			return nil
		}(); err != nil {
			return n, err
		}
	}
	sc.connWriter.Flush()
	return n, err
}

// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	sc.recvMtx.Lock()
	defer sc.recvMtx.Unlock()

	// read off and update the recvBuffer, if non-empty
	if 0 < len(sc.recvBuffer) {
		n = copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[n:]
		return n, err
	}

	// read off the conn
	sealedFrame := sc.recvSealedFrame
	_, err = io.ReadFull(sc.connReader, sealedFrame)
	if err != nil {
		return n, err
	}

	// decrypt the frame.
	// reads and updates the sc.recvNonce
	frame := sc.recvFrame
	_, err = sc.recvAead.Open(frame[:0], sc.recvNonce[:], sealedFrame, nil)
	if err != nil {
		return n, ErrDecryptFrame{Source: err}
	}

	incrNonce(sc.recvNonce)
	// end decryption

	// copy checkLength worth into data,
	// set recvBuffer to the rest.
	chunkLength := binary.LittleEndian.Uint32(frame) // read the first four bytes
	if chunkLength > dataMaxSize {
		return 0, ErrChunkTooBig{
			Received: int(chunkLength),
			Max:      dataMaxSize,
		}
	}

	chunk := frame[dataLenSize : dataLenSize+chunkLength]
	n = copy(data, chunk)
	if n < len(chunk) {
		sc.recvBuffer = make([]byte, len(chunk)-n)
		copy(sc.recvBuffer, chunk[n:])
	}
	return n, err
}

// Implements net.Conn.
func (sc *SecretConnection) Close() error                  { return sc.conn.Close() }
func (sc *SecretConnection) LocalAddr() net.Addr           { return sc.conn.(net.Conn).LocalAddr() }
func (sc *SecretConnection) RemoteAddr() net.Addr          { return sc.conn.(net.Conn).RemoteAddr() }
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		panic("failed to generate ephemeral key-pair")
	}
	return ephPub, ephPriv
}

func shareEphPubKey(conn io.ReadWriter, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	// Send our pubkey and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			lc := *locEphPub
			_, err = protoio.NewDelimitedWriter(conn).WriteMsg(&gogotypes.BytesValue{Value: lc[:]})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var bytes gogotypes.BytesValue
			_, err = protoio.NewDelimitedReader(conn, 1024*1024).ReadMsg(&bytes)
			if err != nil {
				return nil, true, err // abort
			}

			var _remEphPub [32]byte
			copy(_remEphPub[:], bytes.Value)
			return _remEphPub, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return remEphPub, err
	}

	// Otherwise:
	_remEphPub := trs.FirstValue().([32]byte)
	return &_remEphPub, nil
}

func deriveSecrets(
	dhSecret *[32]byte,
	locIsLeast bool,
) (recvSecret, sendSecret *[aeadKeySize]byte) {
	hash := sha256.New
	hkdf := hkdf.New(hash, dhSecret[:], nil, secretConnKeyAndChallengeGen)
	// get enough data for 2 aead keys, and a 32 byte challenge
	res := new([2*aeadKeySize + 32]byte)
	_, err := io.ReadFull(hkdf, res[:])
	if err != nil {
		panic(err)
	}

	recvSecret = new([aeadKeySize]byte)
	sendSecret = new([aeadKeySize]byte)

	// bytes 0 through aeadKeySize - 1 are one aead key.
	// bytes aeadKeySize through 2*aeadKeySize -1 are another aead key.
	// which key corresponds to sending and receiving key depends on whether
	// the local key is less than the remote key.
	if locIsLeast {
		copy(recvSecret[:], res[0:aeadKeySize])
		copy(sendSecret[:], res[aeadKeySize:aeadKeySize*2])
	} else {
		copy(sendSecret[:], res[0:aeadKeySize])
		copy(recvSecret[:], res[aeadKeySize:aeadKeySize*2])
	}

	return recvSecret, sendSecret
}

// computeDHSecret computes a Diffie-Hellman shared secret key
// from our own local private key and the other's public key.
func computeDHSecret(remPubKey, locPrivKey *[32]byte) (*[32]byte, error) {
	shrKey, err := curve25519.X25519(locPrivKey[:], remPubKey[:])
	if err != nil {
		return nil, err
	}
	var shrKeyArray [32]byte
	copy(shrKeyArray[:], shrKey)
	return &shrKeyArray, nil
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		lo = foo
		hi = bar
	} else {
		lo = bar
		hi = foo
	}
	return lo, hi
}

func signChallenge(challenge *[32]byte, locPrivKey crypto.PrivKey) ([]byte, error) {
	signature, err := locPrivKey.Sign(challenge[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig []byte
}

func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature []byte) (recvMsg authSigMessage, err error) {
	// Send our info and receive theirs in tandem.
	trs, _ := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			pbpk, err := cryptoenc.PubKeyToProto(pubKey)
			if err != nil {
				return nil, true, err
			}
			_, err = protoio.NewDelimitedWriter(sc).WriteMsg(&tmp2p.AuthSigMessage{PubKey: pbpk, Sig: signature})
			if err != nil {
				return nil, true, err // abort
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			var pba tmp2p.AuthSigMessage
			_, err = protoio.NewDelimitedReader(sc, 1024*1024).ReadMsg(&pba)
			if err != nil {
				return nil, true, err // abort
			}

			pk, err := cryptoenc.PubKeyFromProto(pba.PubKey)
			if err != nil {
				return nil, true, err // abort
			}

			_recvMsg := authSigMessage{
				Key: pk,
				Sig: pba.Sig,
			}
			return _recvMsg, false, nil
		},
	)

	// If error:
	if trs.FirstError() != nil {
		err = trs.FirstError()
		return recvMsg, err
	}

	_recvMsg := trs.FirstValue().(authSigMessage)
	return _recvMsg, nil
}

// --------------------------------------------------------------------------------

// Increment nonce little-endian by 1 with wraparound.
// Due to chacha20poly1305 expecting a 12 byte nonce we do not use the first four
// bytes. We only increment a 64 bit unsigned int in the remaining 8 bytes
// (little-endian in nonce[4:]).
func incrNonce(nonce *[aeadNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		// Terminates the session and makes sure the nonce would not re-used.
		// See https://github.com/tendermint/tendermint/issues/3531
		panic("can't increase nonce without overflow")
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}

```

---

<a name="file-15"></a>

### File: `tcp/conn/secret_connection_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 13 KB

```go
package conn

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/internal/async"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
)

// Run go test -update from within this module
// to update the golden test vector file.
var update = flag.Bool("update", false, "update .golden files")

type kvstoreConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (drw kvstoreConn) Close() (err error) {
	err2 := drw.PipeWriter.CloseWithError(io.EOF)
	err1 := drw.PipeReader.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

type privKeyWithNilPubKey struct {
	orig crypto.PrivKey
}

func (pk privKeyWithNilPubKey) Bytes() []byte                   { return pk.orig.Bytes() }
func (pk privKeyWithNilPubKey) Sign(msg []byte) ([]byte, error) { return pk.orig.Sign(msg) }
func (privKeyWithNilPubKey) PubKey() crypto.PubKey              { return nil }
func (privKeyWithNilPubKey) Type() string                       { return "privKeyWithNilPubKey" }

func TestSecretConnectionHandshake(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
	if err := barSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentWrite(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)

	// write from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	n := 100
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)
	go writeLots(t, wg, fooSecConn, fooWriteText, n)

	// Consume reads from bar's reader
	readLots(t, wg, barSecConn, n*2)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestConcurrentRead(t *testing.T) {
	fooSecConn, barSecConn := makeSecretConnPair(t)
	fooWriteText := cmtrand.Str(dataMaxSize)
	n := 100

	// read from two routines.
	// should be safe from race according to net.Conn:
	// https://golang.org/pkg/net/#Conn
	wg := new(sync.WaitGroup)
	wg.Add(3)
	go readLots(t, wg, fooSecConn, n/2)
	go readLots(t, wg, fooSecConn, n/2)

	// write to bar
	writeLots(t, wg, barSecConn, fooWriteText, n)
	wg.Wait()

	if err := fooSecConn.Close(); err != nil {
		t.Error(err)
	}
}

func TestSecretConnectionReadWrite(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	fooWrites, barWrites := []string{}, []string{}
	fooReads, barReads := []string{}, []string{}

	// Pre-generate the things to write (for foo & bar)
	for i := 0; i < 100; i++ {
		fooWrites = append(fooWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
		barWrites = append(barWrites, cmtrand.Str((cmtrand.Int()%(dataMaxSize*5))+1))
	}

	// A helper that will run with (fooConn, fooWrites, fooReads) and vice versa
	genNodeRunner := func(nodeConn kvstoreConn, nodeWrites []string, nodeReads *[]string) async.Task {
		return func(_ int) (any, bool, error) {
			// Initiate cryptographic private key and secret connection through nodeConn.
			nodePrvKey := ed25519.GenPrivKey()
			nodeSecretConn, err := MakeSecretConnection(nodeConn, nodePrvKey)
			if err != nil {
				t.Errorf("failed to establish SecretConnection for node: %v", err)
				return nil, true, err
			}
			// In parallel, handle some reads and writes.
			trs, ok := async.Parallel(
				func(_ int) (any, bool, error) {
					// Node writes:
					for _, nodeWrite := range nodeWrites {
						n, err := nodeSecretConn.Write([]byte(nodeWrite))
						if err != nil {
							t.Errorf("failed to write to nodeSecretConn: %v", err)
							return nil, true, err
						}
						if n != len(nodeWrite) {
							err = fmt.Errorf("failed to write all bytes. Expected %v, wrote %v", len(nodeWrite), n)
							t.Error(err)
							return nil, true, err
						}
					}
					if err := nodeConn.PipeWriter.Close(); err != nil {
						t.Error(err)
						return nil, true, err
					}
					return nil, false, nil
				},
				func(_ int) (any, bool, error) {
					// Node reads:
					readBuffer := make([]byte, dataMaxSize)
					for {
						n, err := nodeSecretConn.Read(readBuffer)
						if errors.Is(err, io.EOF) {
							if err := nodeConn.PipeReader.Close(); err != nil {
								t.Error(err)
								return nil, true, err
							}
							return nil, false, nil
						} else if err != nil {
							t.Errorf("failed to read from nodeSecretConn: %v", err)
							return nil, true, err
						}
						*nodeReads = append(*nodeReads, string(readBuffer[:n]))
					}
				},
			)
			assert.True(t, ok, "Unexpected task abortion")

			// If error:
			if trs.FirstError() != nil {
				return nil, true, trs.FirstError()
			}

			// Otherwise:
			return nil, false, nil
		}
	}

	// Run foo & bar in parallel
	trs, ok := async.Parallel(
		genNodeRunner(fooConn, fooWrites, &fooReads),
		genNodeRunner(barConn, barWrites, &barReads),
	)
	require.NoError(t, trs.FirstError())
	require.True(t, ok, "unexpected task abortion")

	// A helper to ensure that the writes and reads match.
	// Additionally, small writes (<= dataMaxSize) must be atomically read.
	compareWritesReads := func(writes []string, reads []string) {
		for {
			// Pop next write & corresponding reads
			read := ""
			write := writes[0]
			readCount := 0
			for _, readChunk := range reads {
				read += readChunk
				readCount++
				if len(write) <= len(read) {
					break
				}
				if len(write) <= dataMaxSize {
					break // atomicity of small writes
				}
			}
			// Compare
			if write != read {
				t.Errorf("expected to read %X, got %X", write, read)
			}
			// Iterate
			writes = writes[1:]
			reads = reads[readCount:]
			if len(writes) == 0 {
				break
			}
		}
	}

	compareWritesReads(fooWrites, barReads)
	compareWritesReads(barWrites, fooReads)
}

func TestDeriveSecretsAndChallengeGolden(t *testing.T) {
	goldenFilepath := filepath.Join("testdata", t.Name()+".golden")
	if *update {
		t.Logf("Updating golden test vector file %s", goldenFilepath)
		data := createGoldenTestVectors(t)
		err := cmtos.WriteFile(goldenFilepath, []byte(data), 0o644)
		require.NoError(t, err)
	}
	f, err := os.Open(goldenFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		params := strings.Split(line, ",")
		randSecretVector, err := hex.DecodeString(params[0])
		require.NoError(t, err)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		locIsLeast, err := strconv.ParseBool(params[1])
		require.NoError(t, err)
		expectedRecvSecret, err := hex.DecodeString(params[2])
		require.NoError(t, err)
		expectedSendSecret, err := hex.DecodeString(params[3])
		require.NoError(t, err)

		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		require.Equal(t, expectedRecvSecret, (*recvSecret)[:], "Recv Secrets aren't equal")
		require.Equal(t, expectedSendSecret, (*sendSecret)[:], "Send Secrets aren't equal")
	}
}

func TestNilPubkey(t *testing.T) {
	fooConn, barConn := makeKVStoreConnPair()
	defer fooConn.Close()
	defer barConn.Close()
	fooPrvKey := ed25519.GenPrivKey()
	barPrvKey := privKeyWithNilPubKey{ed25519.GenPrivKey()}

	go MakeSecretConnection(fooConn, fooPrvKey) //nolint:errcheck // ignore for tests

	_, err := MakeSecretConnection(barConn, barPrvKey)
	require.Error(t, err)
	assert.Equal(t, "encoding: unsupported key <nil>", err.Error())
}

func writeLots(t *testing.T, wg *sync.WaitGroup, conn io.Writer, txt string, n int) {
	t.Helper()
	defer wg.Done()
	for i := 0; i < n; i++ {
		_, err := conn.Write([]byte(txt))
		if err != nil {
			t.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
}

func readLots(t *testing.T, wg *sync.WaitGroup, conn io.Reader, n int) {
	t.Helper()
	readBuffer := make([]byte, dataMaxSize)
	for i := 0; i < n; i++ {
		_, err := conn.Read(readBuffer)
		require.NoError(t, err)
	}
	wg.Done()
}

// Creates the data for a test vector file.
// The file format is:
// Hex(diffie_hellman_secret), loc_is_least, Hex(recvSecret), Hex(sendSecret), Hex(challenge).
func createGoldenTestVectors(*testing.T) string {
	data := ""
	for i := 0; i < 32; i++ {
		randSecretVector := cmtrand.Bytes(32)
		randSecret := new([32]byte)
		copy((*randSecret)[:], randSecretVector)
		data += hex.EncodeToString((*randSecret)[:]) + ","
		locIsLeast := cmtrand.Bool()
		data += strconv.FormatBool(locIsLeast) + ","
		recvSecret, sendSecret := deriveSecrets(randSecret, locIsLeast)
		data += hex.EncodeToString((*recvSecret)[:]) + ","
		data += hex.EncodeToString((*sendSecret)[:]) + ","
	}
	return data
}

// Each returned ReadWriteCloser is akin to a net.Connection.
func makeKVStoreConnPair() (fooConn, barConn kvstoreConn) {
	barReader, fooWriter := io.Pipe()
	fooReader, barWriter := io.Pipe()
	return kvstoreConn{fooReader, fooWriter}, kvstoreConn{barReader, barWriter}
}

func makeSecretConnPair(tb testing.TB) (fooSecConn, barSecConn *SecretConnection) {
	tb.Helper()
	var (
		fooConn, barConn = makeKVStoreConnPair()
		fooPrvKey        = ed25519.GenPrivKey()
		fooPubKey        = fooPrvKey.PubKey()
		barPrvKey        = ed25519.GenPrivKey()
		barPubKey        = barPrvKey.PubKey()
	)

	// Make connections from both sides in parallel.
	trs, ok := async.Parallel(
		func(_ int) (val any, abort bool, err error) {
			fooSecConn, err = MakeSecretConnection(fooConn, fooPrvKey)
			if err != nil {
				tb.Errorf("failed to establish SecretConnection for foo: %v", err)
				return nil, true, err
			}
			remotePubBytes := fooSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), barPubKey.Bytes()) {
				err = fmt.Errorf("unexpected fooSecConn.RemotePubKey.  Expected %v, got %v",
					barPubKey, fooSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
		func(_ int) (val any, abort bool, err error) {
			barSecConn, err = MakeSecretConnection(barConn, barPrvKey)
			if barSecConn == nil {
				tb.Errorf("failed to establish SecretConnection for bar: %v", err)
				return nil, true, err
			}
			remotePubBytes := barSecConn.RemotePubKey()
			if !bytes.Equal(remotePubBytes.Bytes(), fooPubKey.Bytes()) {
				err = fmt.Errorf("unexpected barSecConn.RemotePubKey.  Expected %v, got %v",
					fooPubKey, barSecConn.RemotePubKey())
				tb.Error(err)
				return nil, true, err
			}
			return nil, false, nil
		},
	)

	require.NoError(tb, trs.FirstError())
	require.True(tb, ok, "Unexpected task abortion")

	return fooSecConn, barSecConn
}

// Benchmarks

func BenchmarkWriteSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	// Consume reads from bar's reader
	go func() {
		readBuffer := make([]byte, dataMaxSize)
		for {
			_, err := barSecConn.Read(readBuffer)
			if errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				b.Errorf("failed to read from barSecConn: %v", err)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		idx := cmtrand.Intn(len(fooWriteBytes))
		_, err := fooSecConn.Write(fooWriteBytes[idx])
		if err != nil {
			b.Errorf("failed to write to fooSecConn: %v", err)
			return
		}
	}
	b.StopTimer()

	if err := fooSecConn.Close(); err != nil {
		b.Error(err)
	}
	// barSecConn.Close() race condition
}

func BenchmarkReadSecretConnection(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()
	fooSecConn, barSecConn := makeSecretConnPair(b)
	randomMsgSizes := []int{
		dataMaxSize / 10,
		dataMaxSize / 3,
		dataMaxSize / 2,
		dataMaxSize,
		dataMaxSize * 3 / 2,
		dataMaxSize * 2,
		dataMaxSize * 7 / 2,
	}
	fooWriteBytes := make([][]byte, 0, len(randomMsgSizes))
	for _, size := range randomMsgSizes {
		fooWriteBytes = append(fooWriteBytes, cmtrand.Bytes(size))
	}
	go func() {
		for i := 0; i < b.N; i++ {
			idx := cmtrand.Intn(len(fooWriteBytes))
			_, err := fooSecConn.Write(fooWriteBytes[idx])
			if err != nil {
				b.Errorf("failed to write to fooSecConn: %v, %v,%v", err, i, b.N)
				return
			}
		}
	}()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		readBuffer := make([]byte, dataMaxSize)
		_, err := barSecConn.Read(readBuffer)

		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			b.Fatalf("Failed to read from barSecConn: %v", err)
		}
	}
	b.StopTimer()
}

```

---

<a name="file-16"></a>

### File: `tcp/conn/stream.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package conn

import "time"

// MCConnectionStream is just a wrapper around the original net.Conn.
type MConnectionStream struct {
	conn     *MConnection
	streamID byte
}

// Read reads bytes for the given stream from the internal read queue. Used in
// tests. Production code should use MConnection.OnReceive to avoid copying the
// data.
func (s *MConnectionStream) Read(b []byte) (n int, err error) {
	return s.conn.readBytes(s.streamID, b, 5*time.Second)
}

// Write queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) Write(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, true /* blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// TryWrite queues bytes to be sent onto the internal write queue.
// thread-safe.
func (s *MConnectionStream) TryWrite(b []byte) (n int, err error) {
	if err := s.conn.sendBytes(s.streamID, b, false /* non-blocking */); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Close closes the stream.
// thread-safe.
func (s *MConnectionStream) Close() error {
	delete(s.conn.channelsIdx, s.streamID)
	return nil
}

```

---

<a name="file-17"></a>

### File: `tcp/conn/stream_descriptor.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
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

```

---

<a name="file-18"></a>

### File: `tcp/conn_set.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"net"

	cmtsync "github.com/cometbft/cometbft/libs/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(conn net.Conn) bool
	HasIP(ip net.IP) bool
	Set(conn net.Conn, ip []net.IP)
	Remove(conn net.Conn)
	RemoveAddr(addr net.Addr)
}

type connSetItem struct {
	conn net.Conn
	ips  []net.IP
}

type connSet struct {
	cmtsync.RWMutex

	conns map[string]connSetItem
}

// NewConnSet returns a ConnSet implementation.
func NewConnSet() ConnSet {
	return &connSet{
		conns: map[string]connSetItem{},
	}
}

func (cs *connSet) Has(c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[c.RemoteAddr().String()]

	return ok
}

func (cs *connSet) HasIP(ip net.IP) bool {
	cs.RLock()
	defer cs.RUnlock()

	for _, c := range cs.conns {
		for _, known := range c.ips {
			if known.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func (cs *connSet) Remove(c net.Conn) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, c.RemoteAddr().String())
}

func (cs *connSet) RemoveAddr(addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, addr.String())
}

func (cs *connSet) Set(c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[c.RemoteAddr().String()] = connSetItem{
		conn: c,
		ips:  ips,
	}
}

```

---

<a name="file-19"></a>

### File: `tcp/errors.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 2 KB

```go
package tcp

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
)

// ErrTransportClosed is raised when the Transport has been closed.
type ErrTransportClosed struct{}

func (ErrTransportClosed) Error() string {
	return "transport has been closed"
}

// ErrFilterTimeout indicates that a filter operation timed out.
type ErrFilterTimeout struct{}

func (ErrFilterTimeout) Error() string {
	return "filter timed out"
}

// ErrRejected indicates that a Peer was rejected carrying additional
// information as to the reason.
type ErrRejected struct {
	addr          na.NetAddr
	conn          net.Conn
	err           error
	id            nodekey.ID
	isAuthFailure bool
	isDuplicate   bool
	isFiltered    bool
}

// Addr returns the network address for the rejected Peer.
func (e ErrRejected) Addr() na.NetAddr {
	return e.addr
}

func (e ErrRejected) Error() string {
	if e.isAuthFailure {
		return fmt.Sprintf("auth failure: %s", e.err)
	}

	if e.isDuplicate {
		if e.conn != nil {
			return fmt.Sprintf(
				"duplicate CONN<%s>",
				e.conn.RemoteAddr().String(),
			)
		}
		if e.id != "" {
			return fmt.Sprintf("duplicate ID<%v>", e.id)
		}
	}

	if e.isFiltered {
		if e.conn != nil {
			return fmt.Sprintf(
				"filtered CONN<%s>: %s",
				e.conn.RemoteAddr().String(),
				e.err,
			)
		}

		if e.id != "" {
			return fmt.Sprintf("filtered ID<%v>: %s", e.id, e.err)
		}
	}

	return e.err.Error()
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e ErrRejected) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e ErrRejected) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e ErrRejected) IsFiltered() bool { return e.isFiltered }

```

---

<a name="file-20"></a>

### File: `tcp/tcp.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 11 KB

```go
package tcp

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/netutil"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/fuzz"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

const (
	defaultDialTimeout      = time.Second
	defaultFilterTimeout    = 5 * time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// IPResolver is a behavior subset of net.Resolver.
type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// accept is the container to carry the upgraded connection from an
// asynchronously running routine to the Accept method.
type accept struct {
	netAddr *na.NetAddr
	conn    *conn.MConnection
	err     error
}

// ConnFilterFunc to be implemented by filter hooks after a new connection has
// been established. The set of existing connections is passed along together
// with all resolved IPs for the new connection.
type ConnFilterFunc func(ConnSet, net.Conn, []net.IP) error

// ConnDuplicateIPFilter resolves and keeps all ips for an incoming connection
// and refuses new ones if they come from a known ip.
func ConnDuplicateIPFilter() ConnFilterFunc {
	return func(cs ConnSet, c net.Conn, ips []net.IP) error {
		for _, ip := range ips {
			if cs.HasIP(ip) {
				return ErrRejected{
					conn:        c,
					err:         fmt.Errorf("ip<%v> already connected", ip),
					isDuplicate: true,
				}
			}
		}

		return nil
	}
}

// MultiplexTransportOption sets an optional parameter on the
// MultiplexTransport.
type MultiplexTransportOption func(*MultiplexTransport)

// MultiplexTransportConnFilters sets the filters for rejection new connections.
func MultiplexTransportConnFilters(
	filters ...ConnFilterFunc,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.connFilters = filters }
}

// MultiplexTransportFilterTimeout sets the timeout waited for filter calls to
// return.
func MultiplexTransportFilterTimeout(
	timeout time.Duration,
) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.filterTimeout = timeout }
}

// MultiplexTransportResolver sets the Resolver used for ip lokkups, defaults to
// net.DefaultResolver.
func MultiplexTransportResolver(resolver IPResolver) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.resolver = resolver }
}

// MultiplexTransportMaxIncomingConnections sets the maximum number of
// simultaneous connections (incoming). Default: 0 (unlimited).
func MultiplexTransportMaxIncomingConnections(n int) MultiplexTransportOption {
	return func(mt *MultiplexTransport) { mt.maxIncomingConnections = n }
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	netAddr                na.NetAddr
	listener               net.Listener
	maxIncomingConnections int // see MaxIncomingConnections

	acceptc chan accept
	closec  chan struct{}

	// Lookup table for duplicate ip and id checks.
	conns       ConnSet
	connFilters []ConnFilterFunc

	dialTimeout      time.Duration
	filterTimeout    time.Duration
	handshakeTimeout time.Duration
	nodeKey          nodekey.NodeKey
	resolver         IPResolver

	// TODO(xla): This config is still needed as we parameterise peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig *conn.MConnConfig
	logger  log.Logger
}

// Test multiplexTransport for interface completeness.
var (
	_ transport.Transport = (*MultiplexTransport)(nil)
)

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(nodeKey nodekey.NodeKey, mConfig conn.MConnConfig) *MultiplexTransport {
	return &MultiplexTransport{
		acceptc:          make(chan accept),
		closec:           make(chan struct{}),
		dialTimeout:      defaultDialTimeout,
		filterTimeout:    defaultFilterTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          &mConfig,
		nodeKey:          nodeKey,
		conns:            NewConnSet(),
		resolver:         net.DefaultResolver,
		logger:           log.NewNopLogger(),
	}
}

// SetLogger sets the logger for the transport.
func (mt *MultiplexTransport) SetLogger(l log.Logger) {
	mt.logger = l
}

// NetAddr implements Transport.
func (mt *MultiplexTransport) NetAddr() na.NetAddr {
	return mt.netAddr
}

// Accept implements Transport.
func (mt *MultiplexTransport) Accept() (transport.Conn, *na.NetAddr, error) {
	select {
	// This case should never have any side-effectful/blocking operations to
	// ensure that quality peers are ready to be used.
	case a := <-mt.acceptc:
		if a.err != nil {
			return nil, nil, a.err
		}

		return a.conn, a.netAddr, nil
	case <-mt.closec:
		return nil, nil, ErrTransportClosed{}
	}
}

// Dial implements Transport.
func (mt *MultiplexTransport) Dial(addr na.NetAddr) (transport.Conn, error) {
	c, err := addr.DialTimeout(mt.dialTimeout)
	if err != nil {
		return nil, err
	}

	if mt.mConfig.TestFuzz {
		// so we have time to do peer handshakes and get set up.
		c = fuzz.ConnAfterFromConfig(c, 10*time.Second, mt.mConfig.TestFuzzConfig)
	}

	// TODO(xla): Evaluate if we should apply filters if we explicitly dial.
	if err := mt.filterConn(c); err != nil {
		return nil, err
	}

	mconn, _, err := mt.upgrade(c, &addr)
	if err != nil {
		return nil, err
	}
	mconn.SetLogger(mt.logger.With("remote", addr))

	go mt.cleanupConn(c.RemoteAddr(), mconn.Quit())

	return mconn, nil
}

func (mt *MultiplexTransport) Close() error {
	close(mt.closec)

	if mt.listener != nil {
		return mt.listener.Close()
	}

	return nil
}

func (mt *MultiplexTransport) Listen(addr na.NetAddr) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return err
	}

	if mt.maxIncomingConnections > 0 {
		ln = netutil.LimitListener(ln, mt.maxIncomingConnections)
	}

	mt.netAddr = *na.New(addr.ID, ln.Addr())
	mt.listener = ln

	go mt.acceptPeers()

	return nil
}

func (mt *MultiplexTransport) cleanupConn(netAddr net.Addr, quitCh <-chan struct{}) {
	select {
	case <-quitCh:
		mt.conns.RemoveAddr(netAddr)
	case <-mt.closec:
		return
	}
}

func (mt *MultiplexTransport) acceptPeers() {
	for {
		c, err := mt.listener.Accept()
		if err != nil {
			// If Close() has been called, silently exit.
			select {
			case _, ok := <-mt.closec:
				if !ok {
					return
				}
			default:
				// Transport is not closed
			}

			mt.acceptc <- accept{err: err}
			return
		}

		// Connection upgrade and filtering should be asynchronous to avoid
		// Head-of-line blocking[0].
		// Reference:  https://github.com/tendermint/tendermint/issues/2047
		//
		// [0] https://en.wikipedia.org/wiki/Head-of-line_blocking
		go func(c net.Conn) {
			defer func() {
				if r := recover(); r != nil {
					err := ErrRejected{
						conn:          c,
						err:           fmt.Errorf("recovered from panic: %v", r),
						isAuthFailure: true,
					}
					select {
					case mt.acceptc <- accept{err: err}:
					case <-mt.closec:
						// Give up if the transport was closed.
						_ = c.Close()
						return
					}
				}
			}()

			var (
				mconn        *conn.MConnection
				remotePubKey crypto.PubKey
				netAddr      *na.NetAddr
			)

			err := mt.filterConn(c)
			if err == nil {
				mconn, remotePubKey, err = mt.upgrade(c, nil)
				if err == nil {
					addr := c.RemoteAddr()
					id := nodekey.PubKeyToID(remotePubKey)
					netAddr = na.New(id, addr)
					mconn.SetLogger(mt.logger.With("remote", netAddr))
					go mt.cleanupConn(addr, mconn.Quit())
				}
			}

			select {
			case mt.acceptc <- accept{netAddr, mconn, err}:
				// Make the upgraded peer available.
			case <-mt.closec:
				// Give up if the transport was closed.
				_ = c.Close()
				return
			}
		}(c)
	}
}

func (mt *MultiplexTransport) filterConn(c net.Conn) (err error) {
	defer func() {
		if err != nil {
			_ = c.Close()
		}
	}()

	// Reject if connection is already present.
	if mt.conns.Has(c) {
		return ErrRejected{conn: c, isDuplicate: true}
	}

	// Resolve ips for incoming conn.
	ips, err := resolveIPs(mt.resolver, c)
	if err != nil {
		return err
	}

	errc := make(chan error, len(mt.connFilters))

	for _, f := range mt.connFilters {
		go func(f ConnFilterFunc, c net.Conn, ips []net.IP, errc chan<- error) {
			errc <- f(mt.conns, c, ips)
		}(f, c, ips, errc)
	}

	for i := 0; i < cap(errc); i++ {
		select {
		case err := <-errc:
			if err != nil {
				return ErrRejected{conn: c, err: err, isFiltered: true}
			}
		case <-time.After(mt.filterTimeout):
			return ErrFilterTimeout{}
		}
	}

	mt.conns.Set(c, ips)

	return nil
}

func (mt *MultiplexTransport) upgrade(
	c net.Conn,
	dialedAddr *na.NetAddr,
) (*conn.MConnection, crypto.PubKey, error) {
	var err error
	defer func() {
		if err != nil {
			mt.conns.Remove(c)
			_ = c.Close()
		}
	}()

	secretConn, err := upgradeSecretConn(c, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	if err != nil {
		return nil, nil, ErrRejected{
			conn:          c,
			err:           fmt.Errorf("secret conn failed: %w", err),
			isAuthFailure: true,
		}
	}

	// For outgoing conns, ensure connection key matches dialed key.
	remotePubKey := secretConn.RemotePubKey()
	connID := nodekey.PubKeyToID(remotePubKey)
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID != dialedID {
			return nil, nil, ErrRejected{
				conn: c,
				id:   connID,
				err: fmt.Errorf(
					"conn.ID (%v) dialed ID (%v) mismatch",
					connID,
					dialedID,
				),
				isAuthFailure: true,
			}
		}
	}

	// Copy MConnConfig to avoid it being modified by the transport.
	return conn.NewMConnection(secretConn, *mt.mConfig), remotePubKey, nil
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}

func resolveIPs(resolver IPResolver, c net.Conn) ([]net.IP, error) {
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	addrs, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}

	ips := []net.IP{}

	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}

	return ips, nil
}

```

---

<a name="file-21"></a>

### File: `tcp/tcp_test.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 10 KB

```go
package tcp

import (
	"errors"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

// newMultiplexTransport returns a tcp connected multiplexed peer
// using the default MConnConfig. It's a convenience function used
// for testing.
func newMultiplexTransport(
	nodeKey nodekey.NodeKey,
) *MultiplexTransport {
	return NewMultiplexTransport(
		nodeKey, conn.DefaultMConnConfig(),
	)
}

func TestTransportMultiplex_ConnFilter(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			return errors.New("rejected")
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if e, ok := err.(ErrRejected); ok {
		if !e.IsFiltered() {
			t.Errorf("expected peer to be filtered, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplex_ConnFilterTimeout(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)
	go func() {
		addr := na.New(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, _, err = mt.Accept()
	if _, ok := err.(ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout, got %v", err)
	}
}

func TestTransportMultiplex_MaxIncomingConnections(t *testing.T) {
	pv := ed25519.GenPrivKey()
	id := nodekey.PubKeyToID(pv.PubKey())
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: pv,
		},
	)

	MultiplexTransportMaxIncomingConnections(0)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	const maxIncomingConns = 2
	MultiplexTransportMaxIncomingConnections(maxIncomingConns)(mt)
	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	// Connect more peers than max
	for i := 0; i <= maxIncomingConns; i++ {
		errc := make(chan error)
		go testDialer(*laddr, errc)

		err = <-errc
		if i < maxIncomingConns {
			if err != nil {
				t.Errorf("dialer connection failed: %v", err)
			}
			_, _, err = mt.Accept()
			if err != nil {
				t.Errorf("connection failed: %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "i/o timeout") {
			// mt actually blocks forever on trying to accept a new peer into a full channel so
			// expect the dialer to encounter a timeout error. Calling mt.Accept will block until
			// mt is closed.
			t.Errorf("expected i/o timeout error, got %v", err)
		}
	}
}

func TestTransportMultiplex_AcceptMultiple(t *testing.T) {
	mt := testSetupMultiplexTransport(t)
	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	var (
		seed     = rand.New(rand.NewSource(time.Now().UnixNano()))
		nDialers = seed.Intn(64) + 64
		errc     = make(chan error, nDialers)
	)

	// Setup dialers.
	for i := 0; i < nDialers; i++ {
		go testDialer(*laddr, errc)
	}

	// Catch connection errors.
	for i := 0; i < nDialers; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}

	conns := []transport.Conn{}

	// Accept all connections.
	for i := 0; i < cap(errc); i++ {
		c, _, err := mt.Accept()
		if err != nil {
			t.Fatal(err)
		}

		conns = append(conns, c)
	}

	if have, want := len(conns), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func testDialer(dialAddr na.NetAddr, errc chan error) {
	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	_, err := dialer.Dial(dialAddr)
	if err != nil {
		errc <- err
		return
	}

	// Signal that the connection was established.
	errc <- nil
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		fastNodePV = ed25519.GenPrivKey()
		errc       = make(chan error)
		fastc      = make(chan struct{})
		slowc      = make(chan struct{})
		slowdonec  = make(chan struct{})
	)

	// Simulate slow Peer.
	go func() {
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		c, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(slowc)
		defer func() {
			close(slowdonec)
		}()

		// Make sure we switch to fast peer goroutine.
		runtime.Gosched()

		select {
		case <-fastc:
			// Fast peer connected.
		case <-time.After(200 * time.Millisecond):
			// We error if the fast peer didn't succeed.
			errc <- errors.New("fast peer timed out")
		}

		_, err = upgradeSecretConn(c, 200*time.Millisecond, ed25519.GenPrivKey())
		if err != nil {
			errc <- err
			return
		}
	}()

	// Simulate fast Peer.
	go func() {
		<-slowc

		dialer := newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: fastNodePV,
			},
		)
		dialer.SetLogger(log.TestingLogger())
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr)
		if err != nil {
			errc <- err
			return
		}

		close(fastc)
		<-slowdonec
		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Logf("connection failed: %v", err)
	}

	_, _, err := mt.Accept()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransportMultiplexDialRejectWrongID(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	wrongID := nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
	addr := na.New(wrongID, mt.listener.Addr())

	_, err := dialer.Dial(*addr)
	if err != nil {
		t.Logf("connection failed: %v", err)
		if e, ok := err.(ErrRejected); ok {
			if !e.IsAuthFailure() {
				t.Errorf("expected auth failure, got %v", e)
			}
		} else {
			t.Errorf("expected ErrRejected, got %v", err)
		}
	}
}

func TestTransportConnDuplicateIPFilter(t *testing.T) {
	filter := ConnDuplicateIPFilter()

	if err := filter(nil, &testTransportConn{}, nil); err != nil {
		t.Fatal(err)
	}

	var (
		c  = &testTransportConn{}
		cs = NewConnSet()
	)

	cs.Set(c, []net.IP{
		{10, 0, 10, 1},
		{10, 0, 10, 2},
		{10, 0, 10, 3},
	})

	if err := filter(cs, c, []net.IP{
		{10, 0, 10, 2},
	}); err == nil {
		t.Errorf("expected Peer to be rejected as duplicate")
	}
}

// create listener.
func testSetupMultiplexTransport(t *testing.T) *MultiplexTransport {
	t.Helper()

	var (
		pv = ed25519.GenPrivKey()
		id = nodekey.PubKeyToID(pv.PubKey())
		mt = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	mt.SetLogger(log.TestingLogger())

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	// give the listener some time to get ready
	time.Sleep(20 * time.Millisecond)

	return mt
}

type testTransportAddr struct{}

func (*testTransportAddr) Network() string { return "tcp" }
func (*testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (*testTransportConn) Close() error {
	return errors.New("close() not implemented")
}

func (*testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) Read(_ []byte) (int, error) {
	return -1, errors.New("read() not implemented")
}

func (*testTransportConn) SetDeadline(_ time.Time) error {
	return errors.New("setDeadline() not implemented")
}

func (*testTransportConn) SetReadDeadline(_ time.Time) error {
	return errors.New("setReadDeadline() not implemented")
}

func (*testTransportConn) SetWriteDeadline(_ time.Time) error {
	return errors.New("setWriteDeadline() not implemented")
}

func (*testTransportConn) Write(_ []byte) (int, error) {
	return -1, errors.New("write() not implemented")
}

```

---

<a name="file-22"></a>

### File: `transport.go`

*Modified:* 2025-02-08 11:20:25 • *Size:* 1 KB

```go
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
}

// StreamDescriptor describes a data stream. This could be a substream within a
// multiplexed TCP connection, QUIC stream, etc.
type StreamDescriptor interface {
	// StreamID returns the ID of the stream.
	StreamID() byte
	// MessageType returns the type of the message sent/received on this stream.
	MessageType() proto.Message
}

```

---

## Summary

- **Total files processed:** 22
- **Total combined size:** 352 KB

## Breakdown of File Sizes by Type

- **md**: 224 KB
- **go**: 128 KB