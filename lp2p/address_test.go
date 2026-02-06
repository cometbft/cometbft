package lp2p

import (
	"net"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
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
