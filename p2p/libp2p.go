package p2p

import (
	"fmt"
	"net/url"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	corepeer "github.com/libp2p/go-libp2p/core/peer"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	ma "github.com/multiformats/go-multiaddr"
)

type LibP2PHost struct {
	host.Host
}

func NewLibP2P(config *config.P2PConfig, nodeKey *NodeKey) (*LibP2PHost, error) {
	if !config.UseLibP2P {
		return nil, fmt.Errorf("libp2p is disabled")
	}

	params, err := convertConfigToLibp2p(config, nodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config to libp2p: %w", err)
	}

	// Hard-coded opts for QUIC transport.
	// TODO: make configurable
	opts := []libp2p.Option{
		libp2p.Identity(params.privateKey),
		libp2p.ListenAddrs(params.listenAddr),
		libp2p.UserAgent("cometbft"),
		libp2p.Transport(quic.NewTransport),
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	return &LibP2PHost{
		Host: host,
	}, nil
}

func (h *LibP2PHost) AddrInfo() corepeer.AddrInfo {
	return corepeer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
}

type libp2pConfig struct {
	privateKey crypto.PrivKey
	listenAddr ma.Multiaddr
}

func convertConfigToLibp2p(config *config.P2PConfig, nodeKey *NodeKey) (*libp2pConfig, error) {
	if config == nil || nodeKey == nil {
		return nil, fmt.Errorf("config or node key is nil")
	}

	// 1. Parse node's private key
	keyType := nodeKey.PrivKey.Type()
	if keyType != ed25519.KeyType {
		return nil, fmt.Errorf("unsupported private key type (got %q, want %q)", keyType, ed25519.KeyType)
	}

	pk, err := crypto.UnmarshalEd25519PrivateKey(nodeKey.PrivKey.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	// 2. Parse address into multi addr (forces UDP & QUIC)
	listenAddr, err := addressToQuicMultiaddr(config.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to convert %q to multiaddr: %w", config.ListenAddress, err)
	}

	return &libp2pConfig{
		privateKey: pk,
		listenAddr: listenAddr,
	}, nil
}

// addressToQuicMultiaddr converts a given address to a QUIC multiaddr
// example: "tcp://192.0.2.0:65432" -> "/ip4/192.0.2.0/udp/65432/quic-v1"
func addressToQuicMultiaddr(addr string) (ma.Multiaddr, error) {
	parts, err := url.Parse(addr)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to parse address: %w", err)
	case parts.Hostname() == "":
		return nil, fmt.Errorf("host is empty")
	case parts.Port() == "":
		return nil, fmt.Errorf("port is empty")
	}

	// /ip4/192.0.2.0/udp/65432/quic-v1/
	// @see https://docs.libp2p.io/concepts/transports/quic
	raw := fmt.Sprintf("/ip4/%s/udp/%s/quic-v1/", parts.Hostname(), parts.Port())

	return ma.NewMultiaddr(raw)
}
