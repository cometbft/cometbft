package kcp

import (
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	kcp "github.com/xtaci/kcp-go"
)

const (
	defaultDataShards      = 10
	defaultParityShards    = 3
	defaultMaxWindowSize   = 32768
	defaultReadBufferSize  = 4194304 // 4MB
	defaultWriteBufferSize = 4194304 // 4MB
	defaultReadTimeout     = 30 * time.Second
	defaultWriteTimeout    = 30 * time.Second
	defaultNoDelay         = 1     // Enable nodelay mode for faster transmission
	defaultFastResend      = 1     // Enable fast resend
	defaultCongestionCtrl  = false // Disable congestion control for consensus
	defaultRTO             = 200   // Lower RTO for faster retransmission
	defaultMTU             = 1400  // Typical MTU size
)

// Transport implements the transport.Transport interface using KCP
type Transport struct {
	listener *kcp.Listener
	logger   log.Logger
	metrics  *transport.MetricsCollector
	netAddr  *na.NetAddr

	// Connection management
	mtx         sync.RWMutex
	connections map[string]*kcp.UDPSession

	// Options
	dataShards    int
	parityShards  int
	maxWindowSize int
	readTimeout   time.Duration
	writeTimeout  time.Duration

	closed      chan struct{}
	isListening bool

	noDelay        int
	fastResend     int
	congestionCtrl bool
	rto            int
	mtu            int
}

// Options contains KCP-specific configuration
type Options struct {
	DataShards     int
	ParityShards   int
	MaxWindowSize  int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	NoDelay        int  // 0:disable(default), 1:enable
	FastResend     int  // 0:disable, 1:enable unconditional fast resend
	CongestionCtrl bool // Enable/disable congestion control
	RTO            int  // Retransmission timeout in milliseconds
	MTU            int  // Maximum transmission unit
}

// NewTransport creates a new KCP transport instance
func NewTransport(opts *Options) (*Transport, error) {
	if opts == nil {
		opts = &Options{}
	}

	if opts.DataShards == 0 {
		opts.DataShards = defaultDataShards
	}
	if opts.ParityShards == 0 {
		opts.ParityShards = defaultParityShards
	}
	if opts.MaxWindowSize == 0 {
		opts.MaxWindowSize = defaultMaxWindowSize
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = defaultReadTimeout
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = defaultWriteTimeout
	}

	// Set consensus-optimized defaults
	if opts.NoDelay == 0 {
		opts.NoDelay = defaultNoDelay
	}
	if opts.FastResend == 0 {
		opts.FastResend = defaultFastResend
	}
	if !opts.CongestionCtrl {
		opts.CongestionCtrl = defaultCongestionCtrl
	}
	if opts.RTO == 0 {
		opts.RTO = defaultRTO
	}
	if opts.MTU == 0 {
		opts.MTU = defaultMTU
	}

	return &Transport{
		connections:    make(map[string]*kcp.UDPSession),
		dataShards:     opts.DataShards,
		parityShards:   opts.ParityShards,
		maxWindowSize:  opts.MaxWindowSize,
		readTimeout:    opts.ReadTimeout,
		writeTimeout:   opts.WriteTimeout,
		closed:         make(chan struct{}),
		logger:         log.NewNopLogger(),
		noDelay:        opts.NoDelay,
		fastResend:     opts.FastResend,
		congestionCtrl: opts.CongestionCtrl,
		rto:            opts.RTO,
		mtu:            opts.MTU,
	}, nil
}

// Listen implements transport.Transport
func (t *Transport) Listen(addr na.NetAddr) error {
	block, _ := kcp.NewNoneBlockCrypt(nil)
	listener, err := kcp.ListenWithOptions(addr.DialString(), block, t.dataShards, t.parityShards)
	if err != nil {
		return err
	}
	t.listener = listener
	t.netAddr = &addr
	t.isListening = true
	return nil
}

// Accept implements transport.Transport
func (t *Transport) Accept() (transport.Conn, *na.NetAddr, error) {
	if !t.isListening {
		return nil, nil, ErrTransportNotListening
	}

	session, err := t.listener.AcceptKCP()
	if err != nil {
		return nil, nil, err
	}

	// Configure session with optimizations
	configureSession(session, t)

	wrapper := NewConn(session)
	netAddr := na.New("", session.RemoteAddr())

	return wrapper, netAddr, nil
}

func (t *Transport) Dial(addr na.NetAddr) (transport.Conn, error) {
	block, _ := kcp.NewNoneBlockCrypt(nil)

	session, err := kcp.DialWithOptions(addr.String(), block, t.dataShards, t.parityShards)
	if err != nil {
		return nil, err
	}

	// Configure session with optimizations
	configureSession(session, t)

	t.mtx.Lock()
	t.connections[session.RemoteAddr().String()] = session
	t.mtx.Unlock()

	return NewConn(session), nil
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
		if err := conn.Close(); err != nil {
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

// In both Accept and Dial methods, after creating the session:
// configureSession updates to match actual kcp-go API
// configureSession updates to match actual kcp-go API
func configureSession(session *kcp.UDPSession, t *Transport) {
	// SetNoDelay: ikcp_nodelay(kcp, nodelay, interval, resend, nc)
	congestionInt := 0
	if !t.congestionCtrl {
		congestionInt = 1
	}
	session.SetNoDelay(t.noDelay, 10, 2, congestionInt)

	// Set basic session parameters
	session.SetStreamMode(true)
	session.SetWriteDelay(false)
	session.SetACKNoDelay(true)

	// Set MTU
	session.SetMtu(t.mtu)

	// Set window size
	session.SetWindowSize(t.maxWindowSize, t.maxWindowSize)
}

func (t *Transport) Protocol() transport.Protocol {
	return transport.KCPProtocol
}
