package lp2p

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

const (
	layer4UDP = "udp"
)

func IDFromPrivateKey(cosmosPK cmcrypto.PrivKey) (peer.ID, error) {
	pk, err := privateKeyFromCosmosKey(cosmosPK)
	if err != nil {
		return "", fmt.Errorf("failed to convert private key to libp2p: %w", err)
	}

	return peer.IDFromPrivateKey(pk)
}

// AddressToMultiAddr converts a `listenAddress` to a multiaddr for the given transport
// Currently, only QUIC is supported. Example:
// "tcp://1.1.1.1:5678" yields to "/ip4/1.1.1.1/udp/5678/quic-v1"
func AddressToMultiAddr(addr string, transport string) (ma.Multiaddr, error) {
	if !strings.Contains(addr, "://") {
		addr = "tcp://" + addr
	}

	parts, err := url.Parse(addr)
	switch {
	case err != nil:
		return nil, fmt.Errorf("failed to parse address: %w", err)
	case parts.Hostname() == "":
		return nil, fmt.Errorf("host is empty")
	case parts.Port() == "":
		return nil, fmt.Errorf("port is empty")
	case transport == TransportQUIC:
		return addrToQuicMultiaddr(parts, layer4UDP)
	}

	return nil, fmt.Errorf("unsupported transport: %s", transport)
}

// addrToQuicMultiaddr converts a given address to a QUIC multiaddr
// example: "tcp://192.0.2.0:65432" -> "/ip4/192.0.2.0/udp/65432/quic-v1"
func addrToQuicMultiaddr(parts *url.URL, layer4 string) (ma.Multiaddr, error) {
	raw := fmt.Sprintf("/ip4/%s/%s/%s/%s/", parts.Hostname(), layer4, parts.Port(), TransportQUIC)

	return ma.NewMultiaddr(raw)
}

// netAddressFromPeer converts a peer.AddrInfo to a p2p.NetAddress
func netAddressFromPeer(addrInfo peer.AddrInfo) (*p2p.NetAddress, error) {
	if len(addrInfo.Addrs) == 0 {
		return nil, fmt.Errorf("no addresses")
	}

	// addStr ~ "1.2.3.4:1234"
	_, ipPort, err := manet.DialArgs(addrInfo.Addrs[0])
	if err != nil {
		return nil, err
	}

	parts := strings.Split(ipPort, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid address %s", ipPort)
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address %s", parts[0])
	}

	port, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port %s", parts[1])
	}

	return &p2p.NetAddress{
		ID:   peerIDToKey(addrInfo.ID),
		IP:   ip,
		Port: uint16(port),
	}, nil
}
