// Package lp2p implements auxiliary functions for go-libp2p integration in CometBFT.
// The name is chosen to avoid conflicts with the p2p package.
package lp2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
)

type Host struct {
	host.Host
}

// TransportQUIC quic transport.
// @see https://docs.libp2p.io/concepts/transports/quic
const TransportQUIC = "quic-v1"

func NewHost(config *config.P2PConfig, nodeKey cmcrypto.PrivKey) (*Host, error) {
	if !config.UseLibP2P {
		return nil, fmt.Errorf("libp2p is disabled")
	}

	listenAddr, err := AddressToMultiAddr(config.ListenAddress, TransportQUIC)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ListenAddress, err)
	}

	privateKey, err := privateKeyFromCosmosKey(nodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	// Hard-coded opts for QUIC transport.
	// TODO: make configurable
	opts := []libp2p.Option{
		libp2p.Identity(privateKey),
		libp2p.ListenAddrs(listenAddr),
		libp2p.UserAgent("cometbft"),
		libp2p.Transport(quic.NewTransport),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	return &Host{
		Host: host,
	}, nil
}

func (h *Host) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{ID: h.ID(), Addrs: h.Addrs()}
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
