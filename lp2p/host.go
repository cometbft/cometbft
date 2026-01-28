// Package lp2p implements auxiliary functions for go-libp2p integration in CometBFT.
// The name is chosen to avoid conflicts with the p2p package.
package lp2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
)

// Host is a wrapper around the libp2p host.
// Note that host should NOT be responsible for high-level peer management
// as it's Switch's responsibility.
type Host struct {
	host.Host

	// bootstrapPeers are initial peers specified in the address book
	bootstrapPeers []BootstrapPeer

	logger log.Logger
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

	privateKey, err := privateKeyFromCosmosKey(nodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	listenAddr, err := AddressToMultiAddr(config.ListenAddress, TransportQUIC)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ListenAddress, err)
	}

	bootstrapPeers, err := BootstrapPeersFromConfig(config)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to decode bootstrap peers: %w", err)
	case len(bootstrapPeers) == 0:
		logger.Info("No bootstrap peers provided in the config")
	}

	// todo: add support for libp2p.ResourceManager() based on p2p.lp2p toml config
	// todo: add support for libp2p.BandwidthReporter()
	opts := []libp2p.Option{
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(listenAddr),
		libp2p.UserAgent("cometbft"),
		libp2p.Transport(quic.NewTransport),
	}

	if config.LibP2PConfig.DisableResourceManager {
		opts = append(opts, libp2p.ResourceManager(&network.NullResourceManager{}))
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

	return &Host{
		Host:           host,
		bootstrapPeers: bootstrapPeers,
		logger:         logger,
	}, nil
}

func (h *Host) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
}

func (h *Host) BootstrapPeers() []BootstrapPeer {
	return h.bootstrapPeers
}

func (h *Host) Logger() log.Logger {
	return h.logger
}

func BootstrapPeersFromConfig(config *config.P2PConfig) ([]BootstrapPeer, error) {
	peers := make([]BootstrapPeer, 0, len(config.LibP2PConfig.BootstrapPeers))

	// dedup
	cache := make(map[peer.ID]struct{})

	for _, bp := range config.LibP2PConfig.BootstrapPeers {
		addr, err := AddrInfoFromHostAndID(bp.Host, bp.ID)
		if err != nil {
			return nil, fmt.Errorf("[%s, %s]: %w", bp.Host, bp.ID, err)
		}

		if _, ok := cache[addr.ID]; ok {
			continue
		}

		peers = append(peers, BootstrapPeer{
			AddrInfo:      addr,
			Private:       bp.Private,
			Persistent:    bp.Persistent,
			Unconditional: bp.Unconditional,
		})

		cache[addr.ID] = struct{}{}
	}

	return peers, nil
}
