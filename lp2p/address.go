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

func AddrInfoFromHostAndID(host, id string) (peer.AddrInfo, error) {
	addr, err := AddressToMultiAddr(host, TransportQUIC)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to convert host to multiaddr: %w", err)
	}

	peerID, err := peer.Decode(id)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("failed to decode id: %w", err)
	}

	return peer.AddrInfo{ID: peerID, Addrs: []ma.Multiaddr{addr}}, nil
}

// addrToQuicMultiaddr converts a given address to a QUIC multiaddr
// example: "tcp://192.0.2.0:65432" -> "/ip4/192.0.2.0/udp/65432/quic-v1"
// example: "tcp://my-host.cluster.local:65432" -> "dns/my-host.cluster.local/udp/65432/quic-v1"
func addrToQuicMultiaddr(parts *url.URL, layer4 string) (ma.Multiaddr, error) {
	hostname := parts.Hostname()

	// Determine the network protocol prefix based on the hostname
	var networkProto string
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.To4() != nil {
			networkProto = "ip4"
		} else {
			networkProto = "ip6"
		}
	} else {
		// Not an IP address, treat as DNS hostname
		networkProto = "dns"
	}

	raw := fmt.Sprintf("/%s/%s/%s/%s/%s", networkProto, hostname, layer4, parts.Port(), TransportQUIC)

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
		// Not a literal IP — try resolving as a DNS hostname.
		ips, err := net.LookupIP(parts[0])
		if err != nil || len(ips) == 0 {
			return nil, fmt.Errorf("unable to resolve address %s", parts[0])
		}
		// Prefer IPv4 over IPv6, matching standard dual-stack behavior (RFC 6724).
		ip = preferIPv4(ips)
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

func preferIPv4(ips []net.IP) net.IP {
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip
		}
	}
	return ips[0]
}
