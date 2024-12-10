package p2p

import (
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/internal/cmap"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/types"
)

//go:generate ../scripts/mockery_generate.sh Peer

// Same as the default Prometheus scrape interval in order to not lose
// granularity.
const metricsTickerDuration = 1 * time.Second

// Peer is an interface representing a peer connected on a reactor.
type Peer interface {
	service.Service
	FlushStop()

	ID() nodekey.ID       // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect

	// Conn returns the underlying connection.
	Conn() net.Conn

	NodeInfo() ni.NodeInfo // peer's info
	Status() tcpconn.ConnectionStatus
	SocketAddr() *na.NetAddr // actual address of the socket

	HasChannel(chID byte) bool // Does the peer implement this channel?
	Send(e Envelope) bool      // Send a message to the peer, blocking version
	TrySend(e Envelope) bool   // Send a message to the peer, non-blocking version

	Set(key string, value any)
	Get(key string) any

	SetRemovalFailed()
	GetRemovalFailed() bool
}

// ----------------------------------------------------------

// peerConn contains the raw connection and its config.
type peerConn struct {
	outbound   bool
	persistent bool
	conn       net.Conn // Source connection

	socketAddr *na.NetAddr

	// cached RemoteIP()
	ip net.IP
}

func newPeerConn(
	outbound, persistent bool,
	conn net.Conn,
	socketAddr *na.NetAddr,
) peerConn {
	return peerConn{
		outbound:   outbound,
		persistent: persistent,
		conn:       conn,
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

	host, _, err := net.SplitHostPort(pc.conn.RemoteAddr().String())
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

// peer implements Peer.
//
// Before using a peer, you will need to perform a handshake on connection.
type peer struct {
	service.BaseService

	// raw peerConn and the multiplex connection
	peerConn
	mconn *tcpconn.MConnection

	// peer's node info and the channel it knows about
	// channels = nodeInfo.Channels
	// cached to avoid copying nodeInfo in HasChannel
	nodeInfo ni.NodeInfo
	channels []byte

	// User data
	Data *cmap.CMap

	metrics        *Metrics
	pendingMetrics *peerPendingMetricsCache

	// When removal of a peer fails, we set this flag
	removalAttemptFailed bool
}

type PeerOption func(*peer)

func newPeer(
	pc peerConn,
	mConfig tcpconn.MConnConfig,
	nodeInfo ni.NodeInfo,
	reactorsByCh map[byte]Reactor,
	msgTypeByChID map[byte]proto.Message,
	streams []StreamDescriptor,
	onPeerError func(Peer, any),
	options ...PeerOption,
) *peer {
	p := &peer{
		peerConn:       pc,
		nodeInfo:       nodeInfo,
		channels:       nodeInfo.(ni.Default).Channels,
		Data:           cmap.NewCMap(),
		metrics:        NopMetrics(),
		pendingMetrics: newPeerPendingMetricsCache(),
	}

	// TODO: rip this out from the peer
	p.mconn = createMConnection(
		pc.conn,
		p,
		reactorsByCh,
		msgTypeByChID,
		streams,
		onPeerError,
		mConfig,
	)
	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	for _, option := range options {
		option(p)
	}

	return p
}

// String representation.
func (p *peer) String() string {
	if p.outbound {
		return fmt.Sprintf("Peer{%v %v out}", p.mconn, p.ID())
	}

	return fmt.Sprintf("Peer{%v %v in}", p.mconn, p.ID())
}

// ---------------------------------------------------
// Implements service.Service

// SetLogger implements BaseService.
func (p *peer) SetLogger(l log.Logger) {
	p.Logger = l
	p.mconn.SetLogger(l)
}

// OnStart implements BaseService.
func (p *peer) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}

	if err := p.mconn.Start(); err != nil {
		return err
	}

	go p.metricsReporter()
	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// .Send() calls will get flushed before closing the connection.
//
// NOTE: it is not safe to call this method more than once.
func (p *peer) FlushStop() {
	p.mconn.FlushStop() // stop everything and close the conn
}

// OnStop implements BaseService.
func (p *peer) OnStop() {
	if err := p.mconn.Stop(); err != nil { // stop everything and close the conn
		p.Logger.Debug("Error while stopping peer", "err", err)
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

// Status returns the peer's ConnectionStatus.
func (p *peer) Status() tcpconn.ConnectionStatus {
	return p.mconn.Status()
}

// Send msg bytes to the channel identified by chID byte. Returns false if the
// send queue is full after timeout, specified by MConnection.
//
// thread safe.
func (p *peer) Send(e Envelope) bool {
	return p.send(e.ChannelID, e.Message, p.mconn.Send)
}

// TrySend msg bytes to the channel identified by chID byte. Immediately returns
// false if the send queue is full.
//
// thread safe.
func (p *peer) TrySend(e Envelope) bool {
	return p.send(e.ChannelID, e.Message, p.mconn.TrySend)
}

func (p *peer) send(chID byte, msg proto.Message, sendFunc func(byte, []byte) bool) bool {
	if !p.IsRunning() {
		return false
	} else if !p.HasChannel(chID) {
		return false
	}
	msgType := getMsgType(msg)
	if w, ok := msg.(types.Wrapper); ok {
		msg = w.Wrap()
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		p.Logger.Error("marshaling message to send", "error", err)
		return false
	}
	res := sendFunc(chID, msgBytes)
	if res {
		p.pendingMetrics.AddPendingSendBytes(msgType, len(msgBytes))
	}
	return res
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

// Conn returns the underlying peer source connection.
func (p *peer) Conn() net.Conn {
	return p.peerConn.conn
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
	return p.peerConn.conn.RemoteAddr()
}

// CanSend returns true if the send queue is not full, false otherwise.
func (p *peer) CanSend(chID byte) bool {
	if !p.IsRunning() {
		return false
	}
	return p.mconn.CanSend(chID)
}

// ---------------------------------------------------

func PeerMetrics(metrics *Metrics) PeerOption {
	return func(p *peer) {
		p.metrics = metrics
	}
}

func (p *peer) metricsReporter() {
	metricsTicker := time.NewTicker(metricsTickerDuration)
	defer metricsTicker.Stop()

	for {
		select {
		case <-metricsTicker.C:
			status := p.mconn.Status()
			var sendQueueSize float64
			for _, chStatus := range status.Channels {
				sendQueueSize += float64(chStatus.SendQueueSize)
			}

			p.metrics.RecvRateLimiterDelay.With("peer_id", string(p.ID())).
				Add(status.RecvMonitor.SleepTime.Seconds())
			p.metrics.SendRateLimiterDelay.With("peer_id", string(p.ID())).
				Add(status.SendMonitor.SleepTime.Seconds())

			p.metrics.PeerPendingSendBytes.With("peer_id", string(p.ID())).Set(sendQueueSize)
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

func createMConnection(
	conn net.Conn,
	p *peer,
	reactorsByCh map[byte]Reactor,
	msgTypeByChID map[byte]proto.Message,
	streamDescs []StreamDescriptor,
	onPeerError func(Peer, any),
	config tcpconn.MConnConfig,
) *tcpconn.MConnection {
	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			// Note that its ok to panic here as it's caught in the conn._recover,
			// which does onPeerError.
			panic(fmt.Sprintf("Unknown channel %X", chID))
		}
		mt := msgTypeByChID[chID]
		msg := proto.Clone(mt)
		err := proto.Unmarshal(msgBytes, msg)
		if err != nil {
			panic(fmt.Sprintf("unmarshaling message: %v into type: %s", err, reflect.TypeOf(mt)))
		}
		if w, ok := msg.(types.Unwrapper); ok {
			msg, err = w.Unwrap()
			if err != nil {
				panic(fmt.Sprintf("unwrapping message: %v", err))
			}
		}
		p.pendingMetrics.AddPendingRecvBytes(getMsgType(msg), len(msgBytes))
		reactor.Receive(Envelope{
			ChannelID: chID,
			Src:       p,
			Message:   msg,
		})
	}

	onError := func(r any) {
		onPeerError(p, r)
	}

	// filter out non-tcpconn.ChannelDescriptor streams
	tcpDescs := make([]*tcpconn.ChannelDescriptor, 0, len(streamDescs))
	for _, stream := range streamDescs {
		var ok bool
		d, ok := stream.(*tcpconn.ChannelDescriptor)
		if !ok {
			continue
		}
		tcpDescs = append(tcpDescs, d)
	}

	return tcpconn.NewMConnectionWithConfig(
		conn,
		tcpDescs,
		onReceive,
		onError,
		config,
	)
}

func wrapPeer(c net.Conn, ni ni.NodeInfo, cfg peerConfig, socketAddr *na.NetAddr, mConfig tcpconn.MConnConfig) Peer {
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

	p := newPeer(
		peerConn,
		mConfig,
		ni,
		cfg.reactorsByCh,
		cfg.msgTypeByChID,
		cfg.streamDescs,
		cfg.onPeerError,
		PeerMetrics(cfg.metrics),
	)

	return p
}
