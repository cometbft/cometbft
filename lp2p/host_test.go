package lp2p

import (
	"context"
	"fmt"
	"math"
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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

type connMultiaddrsMock struct {
	local, remote ma.Multiaddr
}

func (c *connMultiaddrsMock) LocalMultiaddr() ma.Multiaddr  { return c.local }
func (c *connMultiaddrsMock) RemoteMultiaddr() ma.Multiaddr { return c.remote }

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

	// Given sample envelope
	type envelope struct {
		protocol protocol.ID
		sender   peer.ID
		receiver peer.ID
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

func TestHostConnGater(t *testing.T) {
	t.Run("rejectThirdPeer", func(t *testing.T) {
		// ARRANGE
		const (
			waitTimeout  = 2 * time.Second
			waitInterval = 50 * time.Millisecond
		)

		var (
			ctx   = context.Background()
			ports = utils.GetFreePorts(t, 4)
			cfg1  = func(cfg *config.LibP2PConfig) {
				cfg.Limits.Mode = config.LibP2PLimitsModeCustom
				cfg.Limits.MaxPeers = 2
				cfg.Limits.MaxPeerStreams = 2
			}

			// given 4 hosts, and host1 has custom max peers config
			host1 = makeTestHost(t, ports[0], withLogging(), withModifiedConfig(cfg1))
			host2 = makeTestHost(t, ports[1], withLogging())
			host3 = makeTestHost(t, ports[2], withLogging())
			host4 = makeTestHost(t, ports[3], withLogging())

			network1 = host1.Network()
		)

		require.NoError(t, host2.Connect(ctx, host1.AddrInfo()))
		require.NoError(t, host3.Connect(ctx, host1.AddrInfo()))

		require.Eventually(t, func() bool {
			return network1.Connectedness(host2.ID()) == network.Connected &&
				network1.Connectedness(host3.ID()) == network.Connected
		}, waitTimeout, waitInterval)

		// ACT
		// try to connect host4->host1
		// note the err is most likely nil because the connection is established, but will be quickly rejected.
		_ = host4.Connect(ctx, host1.AddrInfo())

		// ASSERT
		checkNotConnected := func() bool {
			return network1.Connectedness(host4.ID()) == network.NotConnected && len(network1.Peers()) == 2
		}

		require.Eventually(t, checkNotConnected, waitTimeout, waitInterval)
		require.ElementsMatch(t, []peer.ID{host2.ID(), host3.ID()}, host1.Network().Peers())
	})

	t.Run("rejectWhenHostNil", func(t *testing.T) {
		// ConnGater rejects all connections when host is not yet set (allowMorePeers returns false)
		cg := &ConnGater{host: nil, maxPeers: 10}
		addr, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/26656/quic-v1")
		require.NoError(t, err)
		cm := &connMultiaddrsMock{local: addr, remote: addr}
		require.False(t, cg.InterceptAccept(cm))
		require.False(t, cg.InterceptAddrDial(peer.ID(""), addr))
		require.False(t, cg.InterceptPeerDial(peer.ID("")))
	})
}

func TestResourceManager(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig().LibP2PConfig
		cfg.Limits.Mode = config.LibP2PLimitsModeDisabled

		// ACT
		rm, _, err := ResourceManagerFromConfig(cfg)

		// ASSERT
		require.NoError(t, err)
		_, ok := rm.(*network.NullResourceManager)
		require.True(t, ok)
	})

	t.Run("default", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig().LibP2PConfig
		cfg.Limits.Mode = config.LibP2PLimitsModeDefault

		// ACT
		rm, limiter, err := ResourceManagerFromConfig(cfg)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, rm.Close()) })

		// ASSERT
		require.NotNil(t, rm)
		_, isNull := rm.(*network.NullResourceManager)
		require.False(t, isNull)

		// there should be some limits base on defaults and some params related to the current machine
		systemLimits := limiter.GetSystemLimits()
		t.Logf("systemLimits: %T: %+v", systemLimits, systemLimits)

		require.Less(t, systemLimits.GetStreamTotalLimit(), 1_000_000)
	})

	t.Run("custom", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig().LibP2PConfig
		cfg.Limits.Mode = config.LibP2PLimitsModeCustom
		cfg.Limits.MaxPeerStreams = 50
		cfg.Limits.MaxPeers = 10

		// ACT
		rm, limiter, err := ResourceManagerFromConfig(cfg)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, rm.Close()) })

		// ASSERT
		require.NotNil(t, rm)
		_, isNull := rm.(*network.NullResourceManager)
		require.False(t, isNull)

		var (
			systemLimits  = limiter.GetSystemLimits()
			peerLimits    = limiter.GetPeerLimits("")
			serviceLimits = limiter.GetServiceLimits(identify.ServiceName)
		)

		t.Logf("systemLimits: %T: %+v", systemLimits, systemLimits)
		t.Logf("peerLimits: %T: %+v", peerLimits, peerLimits)
		t.Logf("serviceLimits(identify): %T: %+v", serviceLimits, serviceLimits)

		// no limits on "system" scope...
		require.Equal(t, math.MaxInt64, systemLimits.GetStreamTotalLimit())

		// ...but strict limits on "peer" scope
		require.Equal(t, cfg.Limits.MaxPeerStreams, peerLimits.GetStreamTotalLimit())

		// and also default limits on built-in "service" scope
		require.NotEqual(t, math.MaxInt64, serviceLimits.GetStreamTotalLimit())
	})
}

func TestBootstrapPeers(t *testing.T) {
	t.Run("valid config with peers", func(t *testing.T) {
		// ARRANGE
		pkID := func(pk ed25519.PrivKey) peer.ID {
			id, err := IDFromPrivateKey(pk)
			require.NoError(t, err)
			return id
		}

		// Given 2 private keys
		pk1 := ed25519.GenPrivKey()
		pk2 := ed25519.GenPrivKey()
		pkID1 := pkID(pk1)
		pkID2 := pkID(pk2)

		// Given a P2P config with libp2p enabled and address book peers
		cfg := config.DefaultP2PConfig()
		cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "127.0.0.1:26656", ID: pkID1.String(), Private: true, Persistent: false, Unconditional: true},
			{Host: "127.0.0.1:26657", ID: pkID2.String(), Private: false, Persistent: true, Unconditional: false},
			// duplicate will be ignored
			{Host: "127.0.0.1:26657", ID: pkID2.String(), Private: false, Persistent: true, Unconditional: false},
		}

		// ACT
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg.LibP2PConfig)

		// ASSERT
		require.NoError(t, err)
		require.Len(t, bootstrapPeers, 2)

		// Check first peer
		bp1, ok := bootstrapPeers[pkID1]
		require.True(t, ok)
		require.Equal(t, pkID1, bp1.AddrInfo.ID)
		require.Len(t, bp1.AddrInfo.Addrs, 1)
		require.True(t, bp1.Private)
		require.False(t, bp1.Persistent)
		require.True(t, bp1.Unconditional)

		// Check second peer
		bp2, ok := bootstrapPeers[pkID2]
		require.True(t, ok)
		require.Equal(t, pkID2, bp2.AddrInfo.ID)
		require.Len(t, bp2.AddrInfo.Addrs, 1)
		require.False(t, bp2.Private)
		require.True(t, bp2.Persistent)
		require.False(t, bp2.Unconditional)
	})

	t.Run("invalid host format", func(t *testing.T) {
		// ARRANGE
		cfg := config.DefaultP2PConfig()
		cfg.LibP2PConfig.BootstrapPeers = []config.LibP2PBootstrapPeer{
			{Host: "invalid-host", ID: "12D3KooWRqqKwyNnjwukrxXTUXLiNK838WN5tc8Nk2DnMVPbpVPV"},
		}

		// ACT
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg.LibP2PConfig)

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
		bootstrapPeers, err := BootstrapPeersFromConfig(cfg.LibP2PConfig)

		// ASSERT
		require.Error(t, err)
		require.Nil(t, bootstrapPeers)
	})
}

type testOpts struct {
	pk             ed25519.PrivKey
	bootstrapPeers []config.LibP2PBootstrapPeer
	enableLogging  bool
	modifyConfig   func(*config.LibP2PConfig)
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

func withModifiedConfig(modifier func(*config.LibP2PConfig)) testOption {
	return func(opts *testOpts) { opts.modifyConfig = modifier }
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
	cfg := config.DefaultP2PConfig()
	cfg.RootDir = t.TempDir()
	cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.ExternalAddress = fmt.Sprintf("127.0.0.1:%d", port)

	cfg.LibP2PConfig.Enabled = true
	cfg.LibP2PConfig.BootstrapPeers = optsVal.bootstrapPeers
	if optsVal.modifyConfig != nil {
		optsVal.modifyConfig(&cfg.LibP2PConfig)
	}

	logger := log.NewNopLogger()
	if optsVal.enableLogging {
		logger = log.TestingLogger()
	}

	host, err := NewHost(cfg, optsVal.pk, logger)
	require.NoError(t, err)

	t.Cleanup(func() { host.Close() })

	return host
}

func connectBootstrapPeers(t *testing.T, ctx context.Context, h *Host, peers map[peer.ID]BootstrapPeer) {
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

	return hosts
}
