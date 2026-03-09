package lp2p

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestAddressToMultiAddr(t *testing.T) {
	for _, tt := range []struct {
		name        string
		addr        string
		transport   string
		want        string
		errContains string
	}{
		{
			name:      "tcp to quic",
			addr:      "tcp://1.1.1.1:5678",
			transport: TransportQUIC,
			want:      "/ip4/1.1.1.1/udp/5678/quic-v1",
		},
		{
			name:      "just ip and port to quic",
			addr:      "1.1.1.1:5678",
			transport: TransportQUIC,
			want:      "/ip4/1.1.1.1/udp/5678/quic-v1",
		},
		{
			name:        "no port provided",
			addr:        "1.1.1.1",
			transport:   TransportQUIC,
			errContains: "port is empty",
		},
		{
			name:      "hostname",
			addr:      "my-app-7d9c6f7c9f-2xk8m.default.pod.cluster.local:5678",
			transport: TransportQUIC,
			want:      "/dns/my-app-7d9c6f7c9f-2xk8m.default.pod.cluster.local/udp/5678/quic-v1",
		},
		{
			name:      "hostname2",
			addr:      "my-app-7d9c6f7c9f-2xk8m:5678",
			transport: TransportQUIC,
			want:      "/dns/my-app-7d9c6f7c9f-2xk8m/udp/5678/quic-v1",
		},
		{
			name:      "localhost",
			addr:      "localhost:5678",
			transport: TransportQUIC,
			want:      "/dns/localhost/udp/5678/quic-v1",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddressToMultiAddr(tt.addr, tt.transport)
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestAddrInfoFromHostAndID(t *testing.T) {
	// Generate valid peer IDs for test cases
	genPeerID := func(t *testing.T) string {
		t.Helper()
		pk := ed25519.GenPrivKey()
		id, err := IDFromPrivateKey(pk)
		require.NoError(t, err)
		return id.String()
	}

	staticString := func(s string) func(*testing.T) string {
		return func(*testing.T) string { return s }
	}

	for _, tt := range []struct {
		name        string
		host        string
		id          func(*testing.T) string
		errContains string
		assert      func(t *testing.T, addrInfo peer.AddrInfo)
	}{
		{
			name: "valid host and peer ID",
			host: "127.0.0.1:26656",
			id:   genPeerID,
			assert: func(t *testing.T, addrInfo peer.AddrInfo) {
				require.NotEmpty(t, addrInfo.ID)
				require.Len(t, addrInfo.Addrs, 1)
				require.Equal(t, "/ip4/127.0.0.1/udp/26656/quic-v1", addrInfo.Addrs[0].String())
			},
		},
		{
			name: "valid host with tcp protocol and peer ID",
			host: "tcp://192.0.2.0:65432",
			id:   genPeerID,
			assert: func(t *testing.T, addrInfo peer.AddrInfo) {
				require.NotEmpty(t, addrInfo.ID)
				require.Len(t, addrInfo.Addrs, 1)
				require.Equal(t, "/ip4/192.0.2.0/udp/65432/quic-v1", addrInfo.Addrs[0].String())
			},
		},
		{
			name:        "invalid host format - no port",
			host:        "127.0.0.1",
			id:          staticString("12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"),
			errContains: "failed to convert host to multiaddr",
		},
		{
			name:        "invalid host format - empty host",
			host:        "",
			id:          staticString("12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"),
			errContains: "failed to convert host to multiaddr",
		},
		{
			name:        "invalid peer ID",
			host:        "127.0.0.1:26656",
			id:          staticString("invalid-peer-id"),
			errContains: "failed to decode id",
		},
		{
			name:        "empty peer ID",
			host:        "127.0.0.1:26656",
			id:          staticString(""),
			errContains: "failed to decode id",
		},
		{
			name:        "invalid host format - malformed address",
			host:        "not-an-address",
			id:          staticString("12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"),
			errContains: "failed to convert host to multiaddr",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			peerID := tt.id(t)

			// ACT
			addrInfo, err := AddrInfoFromHostAndID(tt.host, peerID)

			// ASSERT
			if tt.errContains != "" {
				require.ErrorContains(t, err, tt.errContains)
				require.Empty(t, addrInfo.ID)
				require.Empty(t, addrInfo.Addrs)
				return
			}

			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, addrInfo)
			}
		})
	}
}

func TestNetAddressFromPeer(t *testing.T) {
	peerID, err := IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)

	t.Run("localhost", func(t *testing.T) {
		// "localhost" is resolvable on all platforms.
		addr, err := AddressToMultiAddr("localhost:5678", TransportQUIC)
		require.NoError(t, err)
		require.Equal(t, "/dns/localhost/udp/5678/quic-v1", addr.String())

		addrInfo := peer.AddrInfo{ID: peerID, Addrs: []ma.Multiaddr{addr}}

		netAddr, err := netAddressFromPeer(addrInfo)
		require.NoError(t, err)
		require.NotNil(t, netAddr.IP)
		require.Equal(t, uint16(5678), netAddr.Port)
	})

	t.Run("unresolvable", func(t *testing.T) {
		addr, err := AddressToMultiAddr("this-host-does-not-exist.invalid:5678", TransportQUIC)
		require.NoError(t, err)

		addrInfo := peer.AddrInfo{ID: peerID, Addrs: []ma.Multiaddr{addr}}

		_, err = netAddressFromPeer(addrInfo)
		require.Error(t, err)
		require.ErrorContains(t, err, "unable to resolve address")
	})
}

func TestPreferIPv4(t *testing.T) {
	v4 := net.ParseIP("1.2.3.4")
	v6 := net.ParseIP("::1")

	require.Equal(t, v4, preferIPv4([]net.IP{v4}), "single v4")
	require.Equal(t, v6, preferIPv4([]net.IP{v6}), "single v6 falls back")
	require.Equal(t, v4, preferIPv4([]net.IP{v6, v4}), "v4 preferred over v6")
	require.Equal(t, v4, preferIPv4([]net.IP{v4, v6}), "v4 first stays first")
}

func TestMultiAddrStr(t *testing.T) {
	for _, tc := range []struct {
		name  string
		addrs []ma.Multiaddr
		want  string
	}{
		{
			name:  "empty",
			addrs: nil,
			want:  "<empty>",
		},
		{
			name:  "single",
			addrs: mustMultiaddrs(t, "/ip4/127.0.0.1/udp/26656/quic-v1"),
			want:  "/ip4/127.0.0.1/udp/26656/quic-v1",
		},
		{
			name:  "multiple",
			addrs: mustMultiaddrs(t, "/ip4/127.0.0.1/udp/26656/quic-v1", "/ip6/::1/udp/26656/quic-v1"),
			want:  "/ip4/127.0.0.1/udp/26656/quic-v1, /ip6/::1/udp/26656/quic-v1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := multiAddrStr(tc.addrs)
			require.Equal(t, tc.want, got)
		})
	}
}

func mustMultiaddrs(t *testing.T, raw ...string) []ma.Multiaddr {
	t.Helper()
	addrs := make([]ma.Multiaddr, len(raw))
	for i, r := range raw {
		a, err := ma.NewMultiaddr(r)
		require.NoError(t, err)
		addrs[i] = a
	}
	return addrs
}

func TestMultiAddrStrByID(t *testing.T) {
	// Create a host and add a peer with addresses to the peerstore
	ports := utils.GetFreePorts(t, 1)
	host := makeTestHostForAddress(t, ports[0])
	peerID, err := IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)

	addr, err := ma.NewMultiaddr("/ip4/192.0.2.1/udp/26656/quic-v1")
	require.NoError(t, err)
	host.Peerstore().AddAddrs(peerID, []ma.Multiaddr{addr}, 24*time.Hour)

	got := host.multiAddrStrByID(peerID)
	require.Equal(t, "/ip4/192.0.2.1/udp/26656/quic-v1", got)

	// Unknown peer returns empty
	unknownID, _ := IDFromPrivateKey(ed25519.GenPrivKey())
	gotUnknown := host.multiAddrStrByID(unknownID)
	require.Equal(t, "<empty>", gotUnknown)
}

func makeTestHostForAddress(t *testing.T, port int) *Host {
	t.Helper()
	cfg := config.DefaultP2PConfig()
	cfg.RootDir = t.TempDir()
	cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.ExternalAddress = cfg.ListenAddress
	cfg.LibP2PConfig.Enabled = true

	host, err := NewHost(cfg, ed25519.GenPrivKey(), log.NewNopLogger())
	require.NoError(t, err)
	t.Cleanup(func() { host.Close() })
	return host
}

func TestIsDNSAddr(t *testing.T) {
	for _, tc := range []struct {
		name   string
		raw    string
		expect bool
	}{
		{
			name:   "not a dns ipv4 tcp",
			raw:    "/ip4/127.0.0.1/tcp/26656",
			expect: false,
		},
		{
			name:   "not a dns ipv6 udp quic",
			raw:    "/ip6/::1/udp/26656/quic-v1",
			expect: false,
		},
		{
			name:   "not a dns p2p only",
			raw:    "/p2p/12D3KooWB4s8mXvuQKrW8P3QnQfYeyH6Ydb8hQm6mdHh8W9c5QkA",
			expect: false,
		},
		{
			name:   "dns protocol only",
			raw:    "/dns/seed.cometbft.com",
			expect: true,
		},
		{
			name:   "dns4 protocol",
			raw:    "/dns4/seed.cometbft.com/tcp/26656",
			expect: true,
		},
		{
			name:   "dns6 protocol",
			raw:    "/dns6/seed.cometbft.com/udp/26656/quic-v1",
			expect: true,
		},
		{
			name:   "dnsaddr protocol",
			raw:    "/dnsaddr/bootstrap.cometbft.com",
			expect: true,
		},
		{
			name:   "dns with quic and p2p",
			raw:    "/dns4/bootstrap.cometbft.com/udp/26656/quic-v1/p2p/12D3KooWB4s8mXvuQKrW8P3QnQfYeyH6Ydb8hQm6mdHh8W9c5QkA",
			expect: true,
		},
		{
			name:   "dns protocol appears later in path",
			raw:    "/ip4/127.0.0.1/tcp/26656/dnsaddr/bootstrap.cometbft.com",
			expect: true,
		},
		{
			name:   "dns appears before ip protocol",
			raw:    "/dns4/bootstrap.cometbft.com/ip4/127.0.0.1/tcp/26656",
			expect: true,
		},
		{
			name:   "localhost dns host",
			raw:    "/dns/localhost/tcp/26656",
			expect: true,
		},
		{
			name:   "loopback not dns even with quic",
			raw:    "/ip4/127.0.0.1/udp/26656/quic-v1",
			expect: false,
		},
		{
			name:   "not dns with circuit relay",
			raw:    "/ip4/127.0.0.1/tcp/26656/p2p-circuit/p2p/12D3KooWB4s8mXvuQKrW8P3QnQfYeyH6Ydb8hQm6mdHh8W9c5QkA",
			expect: false,
		},
		{
			name:   "dns with websocket",
			raw:    "/dns4/rpc.cometbft.com/tcp/443/tls/ws",
			expect: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE
			addr, err := ma.NewMultiaddr(tc.raw)
			require.NoError(t, err)

			// ACT
			got := IsDNSAddr(addr)

			// ASSERT
			require.Equal(t, tc.expect, got)
		})
	}
}
