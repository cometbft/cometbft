package lp2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	corepeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	// ARRANGE
	ctx := context.Background()

	// Given sample protocol ids
	var (
		channelFoo = byte(0xaa)
		channelBar = byte(0xbb)
		protoFoo   = ProtocolID(channelFoo)
		protoBar   = ProtocolID(channelBar)
	)

	// Given 2 available ports
	ports := utils.GetFreePorts(t, 2)

	// Given two hosts that are connected to each other
	host1 := makeTestHost(t, ports[0], withLogging())
	host2 := makeTestHost(t, ports[1], withLogging(), withBootstrapPeers([]config.LibP2PBootstrapPeer{
		{
			// resolve host via hostname
			Host: fmt.Sprintf("localhost:%d", ports[0]),
			ID:   host1.ID().String(),
		},
	}))

	connectBootstrapPeers(t, ctx, host2, host2.BootstrapPeers())

	t.Logf("host1: %+v", host1.AddrInfo())
	t.Logf("host2: %+v", host2.AddrInfo())

	t.Cleanup(func() {
		host2.Close()
		host1.Close()
	})

	// Given sample envelope
	type envelope struct {
		protocol protocol.ID
		sender   corepeer.ID
		receiver corepeer.ID
		message  string
	}

	envelopes := []envelope{}
	mu := sync.Mutex{}

	// Given sample handler for both hosts
	handler := func(stream network.Stream) {
		var (
			conn     = stream.Conn()
			receiver = conn.LocalPeer()
			sender   = conn.RemotePeer()
		)

		if conn.ConnState().Transport != TransportQUIC {
			t.Fatalf("unexpected transport: %s", conn.ConnState().Transport)
			return
		}

		payload, err := StreamReadClose(stream)
		if err != nil {
			t.Fatalf("failed to read from stream originating from %s: %v", sender, err)
			return
		}

		msg := &types.Request{}
		require.NoError(t, msg.Unmarshal(payload))
		require.NotNil(t, msg.GetEcho())

		e := envelope{
			protocol: stream.Protocol(),
			sender:   sender,
			receiver: receiver,
			message:  msg.GetEcho().GetMessage(),
		}

		logMessage := e.message
		if len(logMessage) > 64 {
			logMessage = logMessage[:64] + "..."
		}

		t.Logf(
			"Received envelope: %s -> %s (proto %s): %s",
			e.sender.String(),
			e.receiver.String(),
			e.protocol,
			logMessage,
		)

		mu.Lock()
		defer mu.Unlock()

		envelopes = append(envelopes, e)
	}

	host1.SetStreamHandler(protoFoo, handler)
	host1.SetStreamHandler(protoBar, handler)

	host2.SetStreamHandler(protoFoo, handler)
	host2.SetStreamHandler(protoBar, handler)

	// Given counter peers
	host1Peer2, err := NewPeer(host1, host2.AddrInfo(), p2p.NopMetrics(), false, false, false)
	require.NoError(t, err, "failed to create peer 1->2")
	require.NoError(t, host1Peer2.Start(), "failed to start peer 1->2")

	host2Peer1, err := NewPeer(host2, host1.AddrInfo(), p2p.NopMetrics(), false, false, false)
	require.NoError(t, err, "failed to create peer 2->1")
	require.NoError(t, host2Peer1.Start(), "failed to start peer 2->1")

	t.Logf("host1Peer2: %+v", host1Peer2.ID())
	t.Logf("host2Peer1: %+v", host2Peer1.ID())

	// Given a long string
	// 300kb
	longStr := strings.Repeat("a", 300*1024)

	// ACT
	send1 := host1Peer2.Send(p2p.Envelope{
		ChannelID: channelFoo,
		Message:   types.ToRequestEcho("one two"),
	})

	send2 := host2Peer1.Send(p2p.Envelope{
		ChannelID: channelBar,
		Message:   types.ToRequestEcho("three four"),
	})

	send3 := host1Peer2.TrySend(p2p.Envelope{
		ChannelID: channelBar,
		Message:   types.ToRequestEcho(longStr),
	})

	// ASSERT
	// Ensure we've written to both streams
	require.True(t, send1, "failed to send message 1->2")
	require.True(t, send2, "failed to send message 2->1")
	require.True(t, send3, "failed to send message 1->2")

	// Ensure two envelopes have been received
	wait := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(envelopes) == 3
	}

	require.Eventually(t, wait, 500*time.Millisecond, 50*time.Millisecond)

	// Ensure both envelopes match the expected ones
	expectedEnvelopes := []envelope{
		{
			protocol: protoFoo,
			sender:   host1.ID(),
			receiver: host2.ID(),
			message:  "one two",
		},
		{
			protocol: protoBar,
			sender:   host2.ID(),
			receiver: host1.ID(),
			message:  "three four",
		},
		{
			protocol: protoBar,
			sender:   host1.ID(),
			receiver: host2.ID(),
			message:  longStr,
		},
	}

	require.ElementsMatch(t, expectedEnvelopes, envelopes)

	t.Run("Ping", func(t *testing.T) {
		// ARRANGE
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// ACT
		rtt, err := host1.Ping(ctx, host2.AddrInfo())

		// ASSERT
		require.NoError(t, err)
		require.NotZero(t, rtt)

		t.Logf("host1 -> host2 RTT: %s", rtt.String())
	})
}

type testOpts struct {
	pk             ed25519.PrivKey
	bootstrapPeers []config.LibP2PBootstrapPeer
	enableLogging  bool
}

type testOption func(*testOpts)

func withLogging() testOption {
	return func(opts *testOpts) { opts.enableLogging = true }
}

func withBootstrapPeers(bootstrapPeers []config.LibP2PBootstrapPeer) testOption {
	return func(opts *testOpts) { opts.bootstrapPeers = bootstrapPeers }
}

func withPrivateKey(pk ed25519.PrivKey) testOption {
	return func(opts *testOpts) { opts.pk = pk }
}

func makeTestHost(t *testing.T, port int, opts ...testOption) *Host {
	t.Helper()

	optsVal := &testOpts{
		pk:             ed25519.GenPrivKey(),
		bootstrapPeers: []config.LibP2PBootstrapPeer{},
		enableLogging:  false,
	}

	for _, opt := range opts {
		opt(optsVal)
	}

	// config
	config := config.DefaultP2PConfig()
	config.RootDir = t.TempDir()
	config.ListenAddress = fmt.Sprintf("127.0.0.1:%d", port)
	config.ExternalAddress = fmt.Sprintf("127.0.0.1:%d", port)

	config.LibP2PConfig.Enabled = true
	config.LibP2PConfig.DisableResourceManager = true
	config.LibP2PConfig.BootstrapPeers = optsVal.bootstrapPeers

	logger := log.NewNopLogger()
	if optsVal.enableLogging {
		logger = log.TestingLogger()
	}

	host, err := NewHost(config, optsVal.pk, logger)
	require.NoError(t, err)

	return host
}

func connectBootstrapPeers(t *testing.T, ctx context.Context, h *Host, peers []BootstrapPeer) {
	require.NotEmpty(t, peers, "no peers to connect to")

	for _, peer := range peers {
		// dial to self
		if h.ID().String() == peer.AddrInfo.ID.String() {
			continue
		}

		h.logger.Info("Connecting to peer", "peer_id", peer.AddrInfo.ID.String())

		err := h.Connect(ctx, peer.AddrInfo)
		require.NoError(t, err, "failed to connect to peer", "peer_id", peer.AddrInfo.ID.String())
	}
}

func makeTestHosts(t *testing.T, numHosts int, opts ...testOption) []*Host {
	ports := utils.GetFreePorts(t, numHosts)

	hosts := make([]*Host, len(ports))
	for i, port := range ports {
		hosts[i] = makeTestHost(t, port, opts...)
	}

	t.Cleanup(func() {
		for _, host := range hosts {
			host.Close()
		}
	})

	return hosts
}

func TestBootstrapPeers(t *testing.T) {
	t.Run("valid config with peers", func(t *testing.T) {
		// ARRANGE
		// Given 2 private keys
		pk1 := ed25519.GenPrivKey()
		pk2 := ed25519.GenPrivKey()

		pkID := func(pk ed25519.PrivKey) string {
			id, err := IDFromPrivateKey(pk)
			require.NoError(t, err)
			return id.String()
		}

		// Given a P2P config with libp2p enabled and address book peers
		cfg := config.DefaultP2PConfig()
		cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "127.0.0.1:26656", ID: pkID(pk1), Private: true, Persistent: false, Unconditional: true},
			{Host: "127.0.0.1:26657", ID: pkID(pk2), Private: false, Persistent: true, Unconditional: false},
			// duplicate will be ignored
			{Host: "127.0.0.1:26657", ID: pkID(pk2), Private: false, Persistent: true, Unconditional: false},
		}

		// ACT
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg)

		// ASSERT
		require.NoError(t, err)
		require.Len(t, bootstrapPeers, 2)

		// Check first peer
		require.Equal(t, pkID(pk1), bootstrapPeers[0].AddrInfo.ID.String())
		require.Len(t, bootstrapPeers[0].AddrInfo.Addrs, 1)
		require.True(t, bootstrapPeers[0].Private)
		require.False(t, bootstrapPeers[0].Persistent)
		require.True(t, bootstrapPeers[0].Unconditional)

		// Check second peer
		require.Equal(t, pkID(pk2), bootstrapPeers[1].AddrInfo.ID.String())
		require.Len(t, bootstrapPeers[1].AddrInfo.Addrs, 1)
		require.False(t, bootstrapPeers[1].Private)
		require.True(t, bootstrapPeers[1].Persistent)
		require.False(t, bootstrapPeers[1].Unconditional)
	})

	t.Run("invalid host format", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig()
		cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "invalid-host", ID: "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"},
		}

		// ACT
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg)

		// ASSERT
		require.Error(t, err)
		require.Nil(t, bootstrapPeers)
	})

	t.Run("invalid peer ID", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig()
		cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "127.0.0.1:26656", ID: "invalid-id"},
		}

		// ACT
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg)

		// ASSERT
		require.Error(t, err)
		require.Nil(t, bootstrapPeers)
	})
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
