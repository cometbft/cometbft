// Package lp2p implements auxiliary functions for go-libp2p integration in CometBFT.
// The name is chosen to avoid conflicts with the p2p package.
package lp2p

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cometbft/cometbft/config"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	ma "github.com/multiformats/go-multiaddr"
)

type Host struct {
	host.Host
	logger      log.Logger
	configPeers []peer.AddrInfo
}

// TransportQUIC quic transport.
// @see https://docs.libp2p.io/concepts/transports/quic
const TransportQUIC = "quic-v1"

type Option func(*options)

func WithAddressBookConfig(ab *AddressBookConfig) Option {
	return func(o *options) { o.addressBook = ab }
}

type options struct {
	addressBook *AddressBookConfig
}

// NewHost creates a new host & connects to the peers in the address book.
func NewHost(
	config *config.P2PConfig,
	nodeKey cmcrypto.PrivKey,
	logger log.Logger,
	option ...Option,
) (*Host, error) {
	if !config.LibP2PEnabled() {
		return nil, fmt.Errorf("libp2p is disabled")
	}

	constructorOptions := &options{}
	for _, opt := range option {
		opt(constructorOptions)
	}

	privateKey, err := privateKeyFromCosmosKey(nodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	listenAddr, err := AddressToMultiAddr(config.ListenAddress, TransportQUIC)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ListenAddress, err)
	}

	addressBook, err := AddressBookFromFilePath(config.LibP2PAddressBookFile())
	switch {
	case constructorOptions.addressBook != nil:
		// override
		addressBook = constructorOptions.addressBook
	case errors.Is(err, os.ErrNotExist):
		logger.Info("Address book file does not exist!")
		addressBook = &AddressBookConfig{}
	case err != nil:
		return nil, fmt.Errorf("address book: %w", err)
	}

	peers, err := addressBook.DecodePeers()
	if err != nil {
		return nil, fmt.Errorf("failed to decode peers from address book: %w", err)
	}

	// Hard-coded opts for QUIC transport.
	// TODO: make configurable
	opts := []libp2p.Option{
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(listenAddr),
		libp2p.UserAgent("cometbft"),
		libp2p.Transport(quic.NewTransport),
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

func (h *Host) InitialConnect(ctx context.Context) {
	if len(h.configPeers) == 0 {
		h.logger.Info("No peers in the address book!")
		return
	}

	for _, peer := range h.configPeers {
		h.logger.Info("Connecting to peer", "peer", peer.String())
		if err := h.Connect(ctx, peer); err != nil {
			h.logger.Error("Failed to connect to peer", "peer", peer.String(), "err", err)
		}
	}
}

func privateKeyFromCosmosKey(key cmcrypto.PrivKey) (crypto.PrivKey, error) {
	keyType := key.Type()

	switch keyType {
	case ed25519.KeyType:
		return crypto.UnmarshalEd25519PrivateKey(key.Bytes())
	case secp256k1.KeyType:
		return crypto.UnmarshalSecp256k1PrivateKey(key.Bytes())
	default:
		return nil, fmt.Errorf("unsupported private key type %q", keyType)
	}
}

func withAddressFactory(addr ma.Multiaddr) libp2p.Option {
	fn := func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return []ma.Multiaddr{addr}
	}

	return libp2p.AddrsFactory(fn)
}
