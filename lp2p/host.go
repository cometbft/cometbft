// Package lp2p implements auxiliary functions for go-libp2p integration in CometBFT.
// The name is chosen to avoid conflicts with the p2p package.
package lp2p

import (
	"context"
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
	logger      log.Logger
	configPeers []peer.AddrInfo
}

// TransportQUIC quic transport.
// @see https://docs.libp2p.io/concepts/transports/quic
const TransportQUIC = "quic-v1"

// NewHost Host constructor.
func NewHost(
	config *config.P2PConfig,
	nodeKey cmcrypto.PrivKey,
	addressBook AddressBookConfig,
	logger log.Logger,
) (*Host, error) {
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

	peers, err := addressBook.DecodePeers()
	if err != nil {
		return nil, fmt.Errorf("failed to decode peers from address book: %w", err)
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
		Host:        host,
		configPeers: peers,
		logger:      logger,
	}, nil
}

func (h *Host) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
}

func (h *Host) ConfigPeers() []peer.AddrInfo {
	return h.configPeers
}

func (h *Host) Logger() log.Logger {
	return h.logger
}

func ConnectPeers(ctx context.Context, h *Host, peers []peer.AddrInfo) {
	if len(peers) == 0 {
		h.logger.Info("No peers to connect to!")
		return
	}

	for _, peer := range peers {
		// dial to self
		if h.ID().String() == peer.ID.String() {
			continue
		}

		h.logger.Info("Connecting to peer", "peer", peer.String())

		if err := h.Connect(ctx, peer); err != nil {
			h.logger.Error("Failed to connect to peer", "peer", peer.String(), "err", err)
			continue
		}
	}
}
