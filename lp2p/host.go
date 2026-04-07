// Package lp2p implements auxiliary functions for go-libp2p integration in CometBFT.
// The name is chosen to avoid conflicts with the p2p package.
package lp2p

import (
	"context"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/config"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// Host is a wrapper around the libp2p host.
// Note that host should NOT be responsible for high-level peer management
// as it's Switch's responsibility.
type Host struct {
	host.Host

	config config.LibP2PConfig

	// bootstrapPeers are initial peers specified in the address book
	bootstrapPeers map[peer.ID]BootstrapPeer

	logger log.Logger

	peerFailureHandlers []func(id peer.ID, err error)
}

// BootstrapPeer initial peers to connect to
type BootstrapPeer struct {
	AddrInfo      peer.AddrInfo
	Private       bool
	Persistent    bool
	Unconditional bool
}

// TransportQUIC quic transport.
// @see https://docs.libp2p.io/concepts/transports/quic
const TransportQUIC = "quic-v1"

// NewHost Host constructor.
func NewHost(config *config.P2PConfig, nodeKey cmcrypto.PrivKey, logger log.Logger) (*Host, error) {
	if !config.LibP2PEnabled() {
		return nil, fmt.Errorf("libp2p is disabled")
	}

	if err := config.LibP2PConfig.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid libp2p config: %w", err)
	}

	privateKey, err := PrivateKeyFromCosmosKey(nodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	listenAddr, err := AddressToMultiAddr(config.ListenAddress, TransportQUIC)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ListenAddress, err)
	}

	bootstrapPeers, err := BootstrapPeersFromConfig(config.LibP2PConfig)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to decode bootstrap peers: %w", err)
	case len(bootstrapPeers) == 0:
		logger.Info("No bootstrap peers provided in the config")
	}

	// host will be set later
	connGater, connGaterEnabled := ConnectionGaterFromConfig(config.LibP2PConfig, nil)

	resourceManager, _, err := ResourceManagerFromConfig(config.LibP2PConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager: %w", err)
	}

	// todo: add support for libp2p.BandwidthReporter()
	opts := []libp2p.Option{
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(listenAddr),
		libp2p.UserAgent("cometbft"),
		libp2p.Ping(true),
		libp2p.Transport(quic.NewTransport),
		libp2p.ResourceManager(resourceManager),
	}

	if connGaterEnabled {
		opts = append(opts, libp2p.ConnectionGater(connGater))
	}

	// We listen on `listenAddr` but advertise `externalAddr` to peers
	if config.ExternalAddress != "" {
		externalAddr, err := AddressToMultiAddr(config.ExternalAddress, TransportQUIC)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ExternalAddress, err)
		}

		opts = append(opts, withAddressFactory(externalAddr))
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	h := &Host{
		Host:           host,
		config:         config.LibP2PConfig,
		bootstrapPeers: bootstrapPeers,
		logger:         logger,
	}

	if connGaterEnabled {
		connGater.SetHost(h)
	}

	return h, nil
}

func (h *Host) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
}

func (h *Host) BootstrapPeers() map[peer.ID]BootstrapPeer {
	return h.bootstrapPeers
}

func (h *Host) BootstrapPeer(id peer.ID) (BootstrapPeer, bool) {
	bp, ok := h.bootstrapPeers[id]
	return bp, ok
}

func (h *Host) Logger() log.Logger {
	return h.logger
}

// Ping pings peers and logs RTT latency (blocking)
// Keep in might that ping service might be disabled on the counterparty's side.
func (h *Host) Ping(ctx context.Context, addrInfo peer.AddrInfo) (time.Duration, error) {
	res := <-ping.Ping(ctx, h, addrInfo.ID)

	return res.RTT, res.Error
}

func (h *Host) AddPeerFailureHandler(handler func(id peer.ID, err error)) {
	h.peerFailureHandlers = append(h.peerFailureHandlers, handler)
}

// EmitPeerFailure emits a peer failure event to all registered handlers.
// This semantic is over host.eventBus for simplicity.
func (h *Host) EmitPeerFailure(id peer.ID, err error) {
	for _, handler := range h.peerFailureHandlers {
		go handler(id, err)
	}
}

func BootstrapPeersFromConfig(config config.LibP2PConfig) (map[peer.ID]BootstrapPeer, error) {
	peers := make(map[peer.ID]BootstrapPeer, len(config.BootstrapPeers))

	for _, bp := range config.BootstrapPeers {
		addr, err := AddrInfoFromHostAndID(bp.Host, bp.ID)
		if err != nil {
			return nil, fmt.Errorf("[%s, %s]: %w", bp.Host, bp.ID, err)
		}

		if _, ok := peers[addr.ID]; ok {
			continue
		}

		peers[addr.ID] = BootstrapPeer{
			AddrInfo:      addr,
			Private:       bp.Private,
			Persistent:    bp.Persistent,
			Unconditional: bp.Unconditional,
		}
	}

	return peers, nil
}

// ResourceManagerFromConfig creates a resource manager from the given config.
func ResourceManagerFromConfig(cfg config.LibP2PConfig) (network.ResourceManager, rcmgr.Limiter, error) {
	if cfg.Limits.Mode == config.LibP2PLimitsModeDisabled {
		return &network.NullResourceManager{}, nil, nil
	}

	// this is what lib-p2p does by default:
	// mem limit: 1/8th of total memory, max 128MB, min 1GB (see defaults.AutoScale())
	defaults := rcmgr.DefaultLimits

	// cap limits for default lib-p2p protocols (identity, ping, ...)
	libp2p.SetDefaultServiceLimits(&defaults)

	if cfg.Limits.Mode == config.LibP2PLimitsModeDefault {
		limiter := rcmgr.NewFixedLimiter(defaults.AutoScale())
		mgr, err := rcmgr.NewResourceManager(limiter)

		return mgr, limiter, err
	}

	if cfg.Limits.Mode == config.LibP2PLimitsModeCustom {
		var (
			partialDefaults = defaults.AutoScale().ToPartialLimitConfig()
			limits          = rcmgr.InfiniteLimits.ToPartialLimitConfig()
			maxPeerStreams  = rcmgr.LimitVal(cfg.Limits.MaxPeerStreams)
		)

		// 1. copy defaults for built-in services/protocols
		limits.Service = partialDefaults.Service
		limits.ServicePeer = partialDefaults.ServicePeer
		limits.Protocol = partialDefaults.Protocol
		limits.ProtocolPeer = partialDefaults.ProtocolPeer

		// 2. also copy sane default conns for peers
		limits.PeerDefault.Conns = partialDefaults.PeerDefault.Conns
		limits.PeerDefault.ConnsInbound = partialDefaults.PeerDefault.ConnsInbound
		limits.PeerDefault.ConnsOutbound = partialDefaults.PeerDefault.ConnsOutbound

		// 2.1 limit max system connections to (max conns per peer * max peers)
		limits.System.Conns = partialDefaults.PeerDefault.Conns * maxPeerStreams

		// 3. set max streams
		// https://github.com/libp2p/go-libp2p/blob/da810a1/p2p/host/resource-manager/scope.go#L168
		limits.PeerDefault.Streams = maxPeerStreams
		limits.PeerDefault.StreamsInbound = maxPeerStreams
		limits.PeerDefault.StreamsOutbound = maxPeerStreams

		limiter := rcmgr.NewFixedLimiter(limits.Build(rcmgr.InfiniteLimits))
		mgr, err := rcmgr.NewResourceManager(limiter)

		return mgr, limiter, err
	}

	return nil, nil, fmt.Errorf("unknown limits mode: %q", cfg.Limits.Mode)
}

// ConnGater limits the number of simultaneously connected peers.
// It is only enabled when `lp2p.limits.mode = "custom"` and uses
// `lp2p.limits.max_peers` as the cap.
//
// The host is injected after host creation because libp2p requires the
// connection gater option during `libp2p.New(...)`, before the host exists.
type ConnGater struct {
	host     *Host
	maxPeers int
}

var _ connmgr.ConnectionGater = (*ConnGater)(nil)

// ConnectionGaterFromConfig creates a connection gater from the given config or returns false if disabled.
func ConnectionGaterFromConfig(cfg config.LibP2PConfig, host *Host) (*ConnGater, bool) {
	if cfg.Limits.Mode != config.LibP2PLimitsModeCustom {
		return nil, false
	}

	return &ConnGater{
		host:     host,
		maxPeers: cfg.Limits.MaxPeers,
	}, true
}

// SetHost sets the host for the connection gater. The host is injected after creation
// because libp2p requires the connection gater option during libp2p.New, before the host exists.
func (c *ConnGater) SetHost(host *Host) { c.host = host }

// InterceptAccept is called when a peer attempts to connect. It returns false to reject the connection
// if the peer count has reached max_peers.
func (c *ConnGater) InterceptAccept(network.ConnMultiaddrs) bool {
	return c.allowMorePeers("caller", "InterceptAccept")
}

func (c *ConnGater) InterceptAddrDial(pid peer.ID, _ multiaddr.Multiaddr) bool {
	return c.allowMorePeers("caller", "InterceptAddrDial", "peer_id", pid.String())
}

func (c *ConnGater) InterceptPeerDial(pid peer.ID) bool {
	return c.allowMorePeers("caller", "InterceptPeerDial", "peer_id", pid.String())
}

func (c *ConnGater) InterceptSecured(network.Direction, peer.ID, network.ConnMultiaddrs) bool {
	return true
}

func (c *ConnGater) InterceptUpgraded(network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func (c *ConnGater) allowMorePeers(labels ...any) bool {
	if c.host == nil {
		return false
	}

	current := len(c.host.Network().Peers())

	if current < c.maxPeers {
		return true
	}

	labels = append(labels, "current_peers", current, "max_peers", c.maxPeers)

	c.host.logger.Info("Rejecting peer due to max peers limit", labels...)

	return false
}
