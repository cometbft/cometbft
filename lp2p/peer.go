package lp2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
)

// Peer represents a remote node connected via libp2p.
// It implements p2p.Peer interface and wraps the libp2p connection
// with CometBFT-specific peer attributes and messaging capabilities.
type Peer struct {
	service.BaseService

	host *Host

	// addrInfo lp2p peer representation. Note that lp2p's addressbook CAN contain
	// different AddrInfo.Addrs for this peer: e.g. peer could announce different addresses in identity protocol.
	// Imagine peerA has p2p.ExternalAddress=<some_pub_ip>, but in our bootstrap_peers it exists under
	// <vpc_private_ip>. We want to use <vpc_private_ip> in this case regardless of what peerA tells us.
	// We might make this configurable and revisit if needed.
	addrInfo peer.AddrInfo

	netAddr *p2p.NetAddress

	// behavioral flags (are not mutually exclusive)
	isPrivate       bool
	isPersistent    bool
	isUnconditional bool

	metrics *p2p.Metrics
}

var _ p2p.Peer = (*Peer)(nil)

func NewPeer(
	host *Host,
	addrInfo peer.AddrInfo,
	metrics *p2p.Metrics,
	isPrivate, isPersistent, isUnconditional bool,
) (*Peer, error) {
	netAddr, err := netAddressFromPeer(addrInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse net address: %w", err)
	}

	p := &Peer{
		host:     host,
		addrInfo: addrInfo,
		netAddr:  netAddr,

		isPrivate:       isPrivate,
		isPersistent:    isPersistent,
		isUnconditional: isUnconditional,

		metrics: metrics,
	}

	logger := host.Logger().With("peer_id", addrInfo.ID.String())

	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	p.SetLogger(logger)

	return p, nil
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{%s}", p.ID())
}

func (p *Peer) ID() p2p.ID {
	return peerIDToKey(p.addrInfo.ID)
}

func (p *Peer) SocketAddr() *p2p.NetAddress {
	return p.netAddr
}

// AddrInfo returns original addr info.
// Note it might differ from host's peerstore
func (p *Peer) AddrInfo() peer.AddrInfo {
	return p.addrInfo
}

func (p *Peer) Get(key string) any {
	v, err := p.host.Peerstore().Get(p.addrInfo.ID, key)
	if err != nil {
		return nil
	}

	return v
}

func (p *Peer) Set(key string, value any) {
	//nolint:errcheck // always returns err=nil
	p.host.Peerstore().Put(p.addrInfo.ID, key, value)
}

func (p *Peer) IsPersistent() bool {
	return p.isPersistent
}

func (p *Peer) IsPrivate() bool {
	// todo: STACK-2089
	return p.isPrivate
}

func (p *Peer) IsUnconditional() bool {
	return p.isUnconditional
}

// Send implements p2p.Peer.
func (p *Peer) Send(e p2p.Envelope) bool {
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "Send", "err", err)
		p.handleSendErr(err)
		return false
	}

	return true
}

func (p *Peer) TrySend(e p2p.Envelope) bool {
	// todo same as SEND, but if current queue is full (its cap=1), immediately return FALSE
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "TrySend", "err", err)
		p.handleSendErr(err)
		return false
	}

	return true
}

func (p *Peer) CloseConn() error {
	return p.host.Network().ClosePeer(p.addrInfo.ID)
}

func (p *Peer) send(e p2p.Envelope) (err error) {
	var (
		peerID     = p.addrInfo.ID
		protocolID = ProtocolID(e.ChannelID)
	)

	payload, err := marshalProto(e.Message)
	if err != nil {
		return err
	}

	var (
		peerIDStr    = peerID.String()
		messageType  = protoTypeName(e.Message)
		payloadLen   = float64(len(payload))
		metricLabels = []string{
			"peer_id", peerIDStr,
			"chID", fmt.Sprintf("%#x", e.ChannelID),
		}

		// note metric's name is misleading, it's a counter, not sum(bytes_pending)
		pendingMessagesCounter = p.metrics.PeerPendingSendBytes.With("peer_id", peerIDStr)
	)

	pendingMessagesCounter.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutStream)
	defer cancel()

	start := time.Now()

	defer func() {
		pendingMessagesCounter.Add(-1)

		if err != nil {
			return
		}

		p.metrics.PeerSendBytesTotal.With(metricLabels...).Add(payloadLen)
		p.metrics.MessageSendBytesTotal.With("message_type", messageType).Add(payloadLen)

		p.Logger.Debug(
			"Sent envelope",
			"protocol", protocolID,
			"peer_id", peerIDStr,
			"send_dur", time.Since(start).String(),
		)
	}()

	// if no streams are available, it will block or return an error
	s, err := p.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return fmt.Errorf("failed to open stream %s: %w", protocolID, err)
	}

	return StreamWriteClose(s, payload)
}

func (p *Peer) handleSendErr(err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, swarm.ErrAllDialsFailed), errors.Is(err, swarm.ErrNoGoodAddresses):
		p.host.EmitPeerFailure(p.addrInfo.ID, err)
	}
}

// NodeInfo returns a DefaultNodeInfo populated with the peer's ID and address.
// Since libp2p does not perform a CometBFT-style handshake, only the fields
// derivable from the connection are filled in (ID, listen address).
func (p *Peer) NodeInfo() p2p.NodeInfo {
	return p2p.DefaultNodeInfo{
		DefaultNodeID: p.ID(),
		ListenAddr:    p.netAddr.DialString(),
	}
}

// RemoteIP returns the remote IP address of the peer derived from its address info.
func (p *Peer) RemoteIP() net.IP {
	return p.netAddr.IP
}

// RemoteAddr returns the remote address of the peer as a net.Addr.
func (p *Peer) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   p.netAddr.IP,
		Port: int(p.netAddr.Port),
	}
}

// IsOutbound returns true because all lp2p peers are bi-directional.
func (*Peer) IsOutbound() bool { return true }

// Status returns an empty ConnectionStatus. Per-channel send queue
// statistics are not available with the libp2p transport.
func (*Peer) Status() conn.ConnectionStatus { return conn.ConnectionStatus{} }

func (*Peer) FlushStop()             {}
func (*Peer) SetRemovalFailed()      {}
func (*Peer) GetRemovalFailed() bool { return false }
