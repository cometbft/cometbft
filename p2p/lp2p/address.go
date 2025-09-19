package lp2p

import (
	"fmt"
	"net/url"
	"strings"

	ma "github.com/multiformats/go-multiaddr"
)

const (
	layer4UDP = "udp"
)

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
