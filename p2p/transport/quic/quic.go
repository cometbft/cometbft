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
func (t *Transport) Listen(addr na.NetAddr) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr.DialString())
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	listener, err := quic.Listen(udpConn, t.tlsConfig, t.quicConfig)
	if err != nil {
		return err
	}

	t.listener = *listener
	t.netAddr = &addr
	t.isListening = true
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

// Protocol implements transport.Transport
func (t *Transport) Protocol() transport.Protocol {
	return transport.QUICProtocol
}
