package p2p

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/internal/cmap"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/types"
)

//go:generate ../scripts/mockery_generate.sh Peer

// Same as the default Prometheus scrape interval in order to not lose
// granularity.
const metricsTickerDuration = 1 * time.Second

// peerConfig is used to bundle data we need to fully setup a Peer.
type peerConfig struct {
	onPeerError func(Peer, any)
	outbound    bool
	// isPersistent allows you to set a function, which, given socket address
	// (for outbound peers) OR self-reported address (for inbound peers), tells
	// if the peer is persistent or not.
	isPersistent func(*na.NetAddr) bool
	// streamID -> streamInfo
	streamInfoByStreamID map[byte]streamInfo
	metrics              *Metrics
}

// Peer is an interface representing a peer connected on a reactor.
type Peer interface {
	service.Service
	FlushStop()

	ID() nodekey.ID       // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect

	NodeInfo() ni.NodeInfo          // peer's info
	ConnState() transport.ConnState // connection state
	SocketAddr() *na.NetAddr        // actual address of the socket

	HasChannel(chID byte) bool // Does the peer implement this channel?
	// Send a message to the peer, blocking version.
	Send(e Envelope) error
	// TrySend is a non-blocking version of Send.
	TrySend(e Envelope) error

	Set(key string, value any)
	Get(key string) any

	SetRemovalFailed()
	GetRemovalFailed() bool
}

// ----------------------------------------------------------

// peerConn contains the raw connection and its config.
type peerConn struct {
	outbound       bool
	persistent     bool
	transport.Conn // Source connection

	socketAddr *na.NetAddr

	// cached RemoteIP()
	ip net.IP
}

func newPeerConn(
	outbound, persistent bool,
	conn transport.Conn,
	socketAddr *na.NetAddr,
) peerConn {
	return peerConn{
		outbound:   outbound,
		persistent: persistent,
		Conn:       conn,
		socketAddr: socketAddr,
	}
}

// ID returns the peer's ID.
//
// Only used in tests.
func (pc peerConn) ID() nodekey.ID {
	return pc.socketAddr.ID
}

// Return the IP from the connection RemoteAddr.
func (pc peerConn) RemoteIP() net.IP {
	if pc.ip != nil {
		return pc.ip
	}

	host, _, err := net.SplitHostPort(pc.RemoteAddr().String())
	if err != nil {
		panic(err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}

	pc.ip = ips[0]

	return pc.ip
}

func (pc peerConn) String() string {
	return pc.socketAddr.String()
}

// peer implements Peer.
//
// Before using a peer, you will need to perform a handshake on connection.
type peer struct {
	service.BaseService

	peerConn

	// peer's node info and the channel it knows about
	// channels = nodeInfo.Channels
	// cached to avoid copying nodeInfo in HasChannel
	nodeInfo ni.NodeInfo
	channels []byte
	streams  map[byte]transport.Stream

	// User data
	Data *cmap.CMap

	metrics        *Metrics
	pendingMetrics *peerPendingMetricsCache

	// When removal of a peer fails, we set this flag
	removalAttemptFailed bool

	// streamID -> streamInfo
	streamInfoByStreamID map[byte]streamInfo
	onPeerError          func(Peer, any)
}

type PeerOption func(*peer)

type streamInfo struct {
	// Reactor associated with this stream.
	reactor Reactor
	// Message type for this stream.
	msgType proto.Message
}

func newPeer(
	pc peerConn,
	nodeInfo ni.NodeInfo,
	streamInfoByStreamID map[byte]streamInfo,
	onPeerError func(Peer, any),
	options ...PeerOption,
) *peer {
	p := &peer{
		peerConn:             pc,
		nodeInfo:             nodeInfo,
		channels:             nodeInfo.(ni.Default).Channels,
		Data:                 cmap.NewCMap(),
		metrics:              NopMetrics(),
		pendingMetrics:       newPeerPendingMetricsCache(),
		streamInfoByStreamID: streamInfoByStreamID,
		onPeerError:          onPeerError,
	}

	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	for _, option := range options {
		option(p)
	}

	if mconn, ok := p.peerConn.Conn.(*tcpconn.MConnection); ok {
		mconn.OnReceive(p.onReceive)
	}

	return p
}

func (p *peer) onReceive(streamID byte, bz []byte) {
	defer func() {
		if r := recover(); r != nil {
			p.Logger.Error("Peer panicked", "err", r, "stack", string(debug.Stack()))
			p.onPeerError(p, r)
		}
	}()

	var (
		reactor = p.streamInfoByStreamID[streamID].reactor
		msgType = p.streamInfoByStreamID[streamID].msgType
		logger  = p.Logger.With("stream", streamID)
	)

	if !p.IsRunning() {
		return
	}

	msg := proto.Clone(msgType)
	err := proto.Unmarshal(bz, msg)
	if err != nil {
		logger.Error("proto.Unmarshal", "as", reflect.TypeOf(msgType), "err", err)
		p.onPeerError(p, err)
		return
	}

	if w, ok := msg.(types.Unwrapper); ok {
		msg, err = w.Unwrap()
		if err != nil {
			logger.Error("proto.Unwrap", "err", err)
			p.onPeerError(p, err)
			return
		}
	}

	logger.Debug("Received message", "msgType", msgType)

	p.pendingMetrics.AddPendingRecvBytes(getMsgType(msg), len(bz))

	reactor.Receive(Envelope{
		ChannelID: streamID,
		Src:       p,
		Message:   msg,
	})
}

// String representation.
func (p *peer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v out}", p.peerConn)
	}

	return fmt.Sprintf("Peer{%v in}", p.peerConn)
}

// ---------------------------------------------------
// Implements service.Service

// SetLogger implements BaseService.
func (p *peer) SetLogger(l log.Logger) {
	p.Logger = l
}

// OnStart implements BaseService.
func (p *peer) OnStart() error {
	// Open streams for all reactors.
	p.streams = make(map[byte]transport.Stream)
	for streamID, info := range p.streamInfoByStreamID {
		var d transport.StreamDescriptor
		descs := info.reactor.StreamDescriptors()
		for _, desc := range descs {
			if desc.StreamID() == streamID {
				d = desc
				break
			}
		}
		stream, err := p.peerConn.OpenStream(streamID, d)
		if err != nil {
			return fmt.Errorf("opening stream %v: %w", streamID, err)
		}
		p.streams[streamID] = stream
	}

	// Start the MConnection if it's an MConnection.
	// NOTE: we do not start the MConnection until all the streams are registered.
	if mconn, ok := p.peerConn.Conn.(*tcpconn.MConnection); ok {
		if err := mconn.Start(); err != nil {
			return fmt.Errorf("starting MConnection: %w", err)
		}
	}

	go p.eventLoop()

	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// Send calls will get flushed before closing the connection.
func (p *peer) FlushStop() {
	if err := p.FlushAndClose("stopping peer"); err != nil {
		p.Logger.Error("Close", "err", err)
	}
}

// OnStop implements BaseService.
func (p *peer) OnStop() {
	if err := p.Close("stopping peer"); err != nil {
		p.Logger.Error("Close", "err", err)
	}
}

// ---------------------------------------------------
// Implements Peer

// ID returns the peer's ID - the hex encoded hash of its pubkey.
func (p *peer) ID() nodekey.ID {
	return p.nodeInfo.ID()
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *peer) IsOutbound() bool {
	return p.peerConn.outbound
}

// IsPersistent returns true if the peer is persistent, false otherwise.
func (p *peer) IsPersistent() bool {
	return p.peerConn.persistent
}

// NodeInfo returns a copy of the peer's NodeInfo.
func (p *peer) NodeInfo() ni.NodeInfo {
	return p.nodeInfo
}

// SocketAddr returns the address of the socket.
// For outbound peers, it's the address dialed (after DNS resolution).
// For inbound peers, it's the address returned by the underlying connection
// (not what's reported in the peer's NodeInfo).
func (p *peer) SocketAddr() *na.NetAddr {
	return p.peerConn.socketAddr
}

// Send sends the given envelope via the stream identified by e.ChannelID. It
// blocks until the message is sent.
//
// thread-safe.
func (p *peer) Send(e Envelope) error {
	stream, ok := p.streams[e.ChannelID]
	if !ok {
		// This should never happen.
		return fmt.Errorf("stream %d not found", e.ChannelID)
	}
	err := p.send(e.Message, stream.Write /* blocking */)
	if err != nil {
		p.Logger.Error("Send", "err", err)
		return err
	}
	p.Logger.Debug("Sent message", "streamID", e.ChannelID)
	return nil
}

// TrySend tries to send the given envelope via the stream identified by e.ChannelID.
// It's non-blocking and returns immediately if the send queue is full.
//
// thread-safe.
func (p *peer) TrySend(e Envelope) error {
	stream, ok := p.streams[e.ChannelID]
	if !ok {
		// This should never happen.
		return fmt.Errorf("stream %d not found", e.ChannelID)
	}

	err := p.send(e.Message, stream.TryWrite /* non-blocking */)
	if err != nil {
		if e, ok := err.(transport.WriteError); ok && e.Full() {
			p.Logger.Debug("Send", "err", err)
		} else {
			p.Logger.Error("Send", "err", err)
		}
		return err
	}
	p.Logger.Debug("Sent message", "streamID", e.ChannelID)
	return nil
}

func (p *peer) send(msg proto.Message, writeFn func([]byte) (int, error)) error {
	if !p.IsRunning() {
		return errors.New("peer not running")
	}

	msgType := getMsgType(msg)
	if w, ok := msg.(types.Wrapper); ok {
		msg = w.Wrap()
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		// This should never happen.
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	n, err := writeFn(msgBytes)
	if err != nil {
		return err
	} else if n != len(msgBytes) {
		// Should never happen in the current TCP implementation.
		return fmt.Errorf("incomplete write: got %d, wanted %d", n, len(msgBytes))
	}

	p.pendingMetrics.AddPendingSendBytes(msgType, n)
	return nil
}

// Get the data for a given key.
//
// thread safe.
func (p *peer) Get(key string) any {
	return p.Data.Get(key)
}

// Set sets the data for the given key.
//
// thread safe.
func (p *peer) Set(key string, data any) {
	p.Data.Set(key, data)
}

// HasChannel returns whether the peer reported implementing this channel.
func (p *peer) HasChannel(chID byte) bool {
	for _, ch := range p.channels {
		if ch == chID {
			return true
		}
	}
	return false
}

func (p *peer) SetRemovalFailed() {
	p.removalAttemptFailed = true
}

func (p *peer) GetRemovalFailed() bool {
	return p.removalAttemptFailed
}

// ---------------------------------------------------
// methods only used for testing
// TODO: can we remove these?

// RemoteAddr returns peer's remote network address.
func (p *peer) RemoteAddr() net.Addr {
	return p.Conn.RemoteAddr()
}

// ---------------------------------------------------

func PeerMetrics(metrics *Metrics) PeerOption {
	return func(p *peer) {
		p.metrics = metrics
	}
}

// report metrics + handle underlying connection errors.
func (p *peer) eventLoop() {
	metricsTicker := time.NewTicker(metricsTickerDuration)
	defer metricsTicker.Stop()

	for {
		select {
		case err := <-p.Conn.ErrorCh():
			p.Logger.Error("Connection error", "err", err)
			p.onPeerError(p, err)
			return
		case <-metricsTicker.C:
			state := p.ConnState()
			var totalSendQueueSize int
			for _, s := range state.StreamStates {
				totalSendQueueSize += s.SendQueueSize
			}
			p.metrics.RecvRateLimiterDelay.With("peer_id", string(p.ID())).
				Add(state.RecvRateLimiterDelay.Seconds())
			p.metrics.SendRateLimiterDelay.With("peer_id", string(p.ID())).
				Add(state.SendRateLimiterDelay.Seconds())
			p.metrics.PeerPendingSendBytes.With("peer_id", string(p.ID())).Set(float64(totalSendQueueSize))

			// Report per peer, per message total bytes, since the last interval
			func() {
				p.pendingMetrics.mtx.Lock()
				defer p.pendingMetrics.mtx.Unlock()
				for _, entry := range p.pendingMetrics.perMessageCache {
					if entry.pendingSendBytes > 0 {
						p.metrics.MessageSendBytesTotal.
							With("message_type", entry.label).
							Add(float64(entry.pendingSendBytes))
						entry.pendingSendBytes = 0
					}
					if entry.pendingRecvBytes > 0 {
						p.metrics.MessageReceiveBytesTotal.
							With("message_type", entry.label).
							Add(float64(entry.pendingRecvBytes))
						entry.pendingRecvBytes = 0
					}
				}
			}()

		case <-p.Quit():
			return
		}
	}
}

// ------------------------------------------------------------------
// helper funcs

func wrapPeer(c transport.Conn, ni ni.NodeInfo, cfg peerConfig, socketAddr *na.NetAddr) Peer {
	persistent := false
	if cfg.isPersistent != nil {
		if cfg.outbound {
			persistent = cfg.isPersistent(socketAddr)
		} else {
			selfReportedAddr, err := ni.NetAddr()
			if err == nil {
				persistent = cfg.isPersistent(selfReportedAddr)
			}
		}
	}

	peerConn := newPeerConn(
		cfg.outbound,
		persistent,
		c,
		socketAddr,
	)

	return newPeer(
		peerConn,
		ni,
		cfg.streamInfoByStreamID,
		cfg.onPeerError,
		PeerMetrics(cfg.metrics),
	)
}
