package lp2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/libp2p/go-libp2p/core/peer"
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

	// Failure tracking for automatic peer removal
	sendFailures    atomic.Int32       // consecutive send failures
	maxSendFailures int32              // threshold for removal (0 = disabled)
	onRemovalError  func(*Peer, error) // callback to trigger removal
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

		// Initialize with zero failures, will be configured by Switch
		sendFailures:    atomic.Int32{},
		maxSendFailures: 0, // 0 = disabled, will be set by Switch
		onRemovalError:  nil,
	}

	logger := host.Logger().With("peer_id", addrInfo.ID.String())

	p.BaseService = *service.NewBaseService(nil, "Peer", p)
	p.SetLogger(logger)

	return p, nil
}

// ConfigureFailureTracking sets up automatic peer removal on send failures.
// Must be called before the peer starts handling messages.
func (p *Peer) ConfigureFailureTracking(maxFailures int32, onRemovalError func(*Peer, error)) {
	p.maxSendFailures = maxFailures
	p.onRemovalError = onRemovalError
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

// isPeerFailure determines if an error represents a peer-related failure
// (network/connection issues) vs a local resource issue (should not count
// against the peer).
func isPeerFailure(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Local resource exhaustion - not the peer's fault
	if strings.Contains(errMsg, "resource limit exceeded") ||
		strings.Contains(errMsg, "too many open files") ||
		strings.Contains(errMsg, "cannot allocate memory") {
		return false
	}

	// Marshal/validation errors - not a peer connection issue
	if strings.Contains(errMsg, "failed to marshal") ||
		strings.Contains(errMsg, "proto") {
		return false
	}

	// Context cancellation from our side - not a peer issue
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Peer-related failures (network, connection, stream issues)
	// These indicate the peer is unreachable or having problems
	if strings.Contains(errMsg, "failed to open stream") ||
		strings.Contains(errMsg, "failed to dial") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no route to host") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "protocol not supported") ||
		strings.Contains(errMsg, "stream reset") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "send failed") ||
		errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Default: treat as peer failure for safety
	// Better to remove a problematic peer than keep a bad one
	return true
}

func (p *Peer) handleSendFailure(err error) {
	// Only track peer-related failures, not local resource issues
	if !isPeerFailure(err) {
		p.Logger.Debug(
			"send failed with non-peer error, not tracking",
			"peer_id", p.ID(),
			"err", err,
		)
		return
	}

	// Track consecutive failures
	if p.maxSendFailures > 0 {
		failures := p.sendFailures.Add(1)

		// Check if threshold exceeded
		if failures >= p.maxSendFailures {
			p.Logger.Error(
				"peer exceeded max send failures, triggering removal",
				"peer_id", p.ID(),
				"consecutive_failures", failures,
				"threshold", p.maxSendFailures,
			)

			// Trigger removal via callback (if configured)
			if p.onRemovalError != nil {
				removalErr := fmt.Errorf("exceeded max send failures: %d/%d", failures, p.maxSendFailures)
				go p.onRemovalError(p, removalErr)
			}
		}
	}
}

// Send implements p2p.Peer.
func (p *Peer) Send(e p2p.Envelope) bool {
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "Send", "err", err)

		p.handleSendFailure(err)

		return false
	}

	// Reset counter on success
	if p.maxSendFailures > 0 {
		p.sendFailures.Store(0)
	}

	return true
}

func (p *Peer) TrySend(e p2p.Envelope) bool {
	// todo same as SEND, but if current queue is full (its cap=1), immediately return FALSE
	if err := p.send(e); err != nil {
		p.Logger.Error("failed to send message", "channel", e.ChannelID, "method", "TrySend", "err", err)
		p.handleSendFailure(err)
		return false
	}

	// Reset counter on success
	if p.maxSendFailures > 0 {
		p.sendFailures.Store(0)
	}

	return true
}

func (p *Peer) CloseConn() error {
	// Clear cached addresses to force DNS re-resolution on next connection attempt
	p.host.Peerstore().ClearAddrs(p.addrInfo.ID)
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

// These methods are not implemented as they're not used by reactors
// (only by PEX/p2p-transport which is not used with go-libp2p)

func (*Peer) Status() conn.ConnectionStatus { return conn.ConnectionStatus{} }
func (*Peer) NodeInfo() p2p.NodeInfo        { return nil }
func (*Peer) RemoteIP() net.IP              { return nil }
func (*Peer) RemoteAddr() net.Addr          { return nil }
func (*Peer) IsOutbound() bool              { return false }
func (*Peer) FlushStop()                    {}
func (*Peer) SetRemovalFailed()             {}
func (*Peer) GetRemovalFailed() bool        { return false }
