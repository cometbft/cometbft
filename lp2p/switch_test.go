package lp2p

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	p2pmock "github.com/cometbft/cometbft/p2p/mock"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestSwitch(t *testing.T) {
	t.Run("BootstrapPeers", func(t *testing.T) {
		// ARRANGE
		var (
			ctx       = context.Background()
			ports     = utils.GetFreePorts(t, 2)
			logBuffer = &syncBuffer{}
			logger    = log.NewTMLogger(logBuffer)
		)

		// Given 2 hosts: A and B
		var (
			privateKeyA = ed25519.GenPrivKey()
			privateKeyB = ed25519.GenPrivKey()
		)

		pkToID := func(pk ed25519.PrivKey) string {
			id, err := IDFromPrivateKey(pk)
			require.NoError(t, err)
			return id.String()
		}

		// Given bootstrap peers for host A
		// host A has self as bootstrap peer -> should be ignored
		bootstrapPeersA := []config.LibP2PBootstrapPeer{
			{
				Host: fmt.Sprintf("127.0.0.1:%d", ports[0]),
				ID:   pkToID(privateKeyA),
			},
			{
				Host: fmt.Sprintf("127.0.0.1:%d", ports[1]),
				ID:   pkToID(privateKeyB),
			},
		}

		var (
			hostA = makeTestHost(t, ports[0], withLogging(), withPrivateKey(privateKeyA), withBootstrapPeers(bootstrapPeersA))
			hostB = makeTestHost(t, ports[1], withLogging(), withPrivateKey(privateKeyB))
		)

		// Given switch A (NOT started yet)
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{},
			p2p.NopMetrics(),
			logger.With("switch", "A"),
		)
		require.NoError(t, err)

		// ACT #1: Host B sends a stream to host A BEFORE switch A is started.
		// This simulates an incoming message arriving BEFORE bootstrap is complete.
		err = hostB.Connect(ctx, hostA.AddrInfo())
		require.NoError(t, err)

		// Create a test stream from B to A
		protoID := ProtocolID(0xBB)
		hostA.SetStreamHandler(protoID, switchA.handleStream)

		// Give some time for the stream handler to be set
		time.Sleep(50 * time.Millisecond)

		stream, err := hostB.NewStream(ctx, hostA.ID(), protoID)
		require.NoError(t, err)

		// If stream was created, write and close (errors expected)
		_, _ = stream.Write([]byte("test message"))
		_ = stream.Close()

		// Give some time for the stream to be processed
		time.Sleep(50 * time.Millisecond)

		// ASSERT #1: No peer should be added since switch is not active
		require.Equal(t, 0, switchA.Peers().Size())
		require.Contains(t, logBuffer.String(), "Ignoring stream from inactive switch")
		require.False(t, switchA.isActive())

		// ACT #2: Start switch A
		require.NoError(t, switchA.Start())
		t.Cleanup(func() {
			_ = switchA.Stop()
		})

		// ASSERT #2: Still no peer added (no new streams sent yet)
		require.Contains(t, logBuffer.String(), "Ignoring connection to self")
		require.Equal(t, 1, switchA.Peers().Size())
		require.True(t, switchA.isActive())

		// ASSERT #3: A has pinged B
		hasPingedB := func() bool {
			str := logBuffer.String()

			return strings.Contains(str, "Ping") && strings.Contains(str, "rtt")
		}

		require.Eventually(t, hasPingedB, time.Second, 50*time.Millisecond)
	})

	t.Run("PersistentPeers", func(t *testing.T) {
		// ARRANGE
		var (
			ctx       = context.Background()
			ports     = utils.GetFreePorts(t, 3)
			logBuffer = &syncBuffer{}
			logger    = log.NewTMLogger(logBuffer)
		)

		// Given 3 hosts: A, B, C
		// Hosts start with NO bootstrap peers
		var (
			hostA = makeTestHost(t, ports[0], withLogging())
			hostB = makeTestHost(t, ports[1], withLogging())
			hostC = makeTestHost(t, ports[2], withLogging())
		)

		// Given switch A with hosts B and C as bootstrap peers
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{},
			p2p.NopMetrics(),
			logger.With("switch", "A"),
		)
		require.NoError(t, err)

		// Connect host to itself should result in no-op
		err = switchA.bootstrapPeer(ctx, hostA.AddrInfo(), PeerAddOptions{
			Persistent: false,
		})
		require.NoError(t, err)

		// Connect host A to B (non-persistent) and C (persistent), and use DNS instead of ip!
		err = switchA.bootstrapPeer(ctx, patchAddrInfoIPToDNS(t, hostB.AddrInfo()), PeerAddOptions{
			Persistent: false,
		})
		require.NoError(t, err)

		err = switchA.bootstrapPeer(ctx, patchAddrInfoIPToDNS(t, hostC.AddrInfo()), PeerAddOptions{
			Persistent: true,
			// it doesn't influence the test, but let's also
			// verify that unconditional checks work
			Unconditional: true,
		})
		require.NoError(t, err)

		peerC := switchA.Peers().Get(peerIDToKey(hostC.ID()))
		require.NotNil(t, peerC, "peer C should be added")
		require.True(t, switchA.IsPeerPersistent(peerC.SocketAddr()))
		require.True(t, switchA.IsPeerUnconditional(peerC.ID()))

		// ACT #1: Start switch A
		require.NoError(t, switchA.Start())
		t.Cleanup(func() {
			_ = switchA.Stop()
		})

		// ASSERT #1: Switch A has 2 peers
		require.Equal(t, 2, switchA.Peers().Size())

		// peerstore & logs contain both DNS and IP addresses
		require.True(t, logBuffer.HasMatchingLine("Connected to peer", "/dns/localhost/udp", "/ip4/127.0.0.1/udp"))
		require.Len(t, hostA.Peerstore().Addrs(hostB.ID()), 2)
		require.Len(t, hostA.Peerstore().Addrs(hostC.ID()), 2)

		// ACT #2: Stop peer C for error (simulates disconnection)
		switchA.StopPeerForError(peerC, "simulated error")

		// ASSERT #2: Peer C is removed initially
		validatePeerRemoved := func() bool {
			return switchA.Peers().Size() == 1 && peerC.IsRunning() == false
		}

		require.Eventually(t, validatePeerRemoved, time.Second, 50*time.Millisecond, "C should be removed")

		validatePeerReconnected := func() bool {
			return switchA.Peers().Size() == 2
		}

		// ASSERT #3: Peer C is reconnected (persistent peer reconnection)
		require.Eventually(t, validatePeerReconnected, 10*time.Second, 100*time.Millisecond, "C should be reconnected")

		// Verify the reconnected peer is C
		reconnectedPeer := switchA.Peers().Get(peerIDToKey(hostC.ID()))
		require.NotNil(t, reconnectedPeer)
		require.True(t, reconnectedPeer.(*Peer).IsPersistent())

		// Check for expected log messages
		require.Contains(t, logBuffer.String(), "Removing peer")
		require.Contains(t, logBuffer.String(), "Reconnected to peer")

		fmt.Println(logBuffer.String())
	})

	t.Run("ErrorTransientReconnect", func(t *testing.T) {
		// ARRANGE
		var (
			ctx       = context.Background()
			ports     = utils.GetFreePorts(t, 2)
			logBuffer = &syncBuffer{}
			logger    = log.NewTMLogger(logBuffer)
		)

		// Given 2 hosts: A and B
		var (
			hostA = makeTestHost(t, ports[0], withLogging())
			hostB = makeTestHost(t, ports[1], withLogging())
		)

		// Given switch A
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{},
			p2p.NopMetrics(),
			logger.With("switch", "A"),
		)
		require.NoError(t, err)

		// Connect A to B as a NON-persistent peer
		err = switchA.bootstrapPeer(ctx, hostB.AddrInfo(), PeerAddOptions{})
		require.NoError(t, err)

		peerB := switchA.Peers().Get(peerIDToKey(hostB.ID()))
		require.NotNil(t, peerB)
		require.False(t, peerB.(*Peer).IsPersistent())

		// ACT #1: Start switch A
		require.NoError(t, switchA.Start())
		t.Cleanup(func() {
			_ = switchA.Stop()
		})

		// ASSERT #1: Switch A has 1 peer
		require.Equal(t, 1, switchA.Peers().Size())

		// ACT #2: Stop peer B with a TRANSIENT error (should trigger reconnection)
		transientErr := &p2p.ErrorTransient{Err: errors.New("something went wrong")}
		switchA.StopPeerForError(peerB, transientErr)

		// ASSERT #2: Peer B is removed initially
		validatePeerRemoved := func() bool {
			return switchA.Peers().Size() == 0 && peerB.IsRunning() == false
		}
		require.Eventually(t, validatePeerRemoved, time.Second, 50*time.Millisecond, "B should be removed")

		// ASSERT #3: Peer B is reconnected (transient error triggers reconnection)
		validatePeerReconnected := func() bool {
			return switchA.Peers().Size() == 1
		}
		require.Eventually(t, validatePeerReconnected, 10*time.Second, 100*time.Millisecond, "B should be reconnected")

		// Verify the reconnected peer is B
		reconnectedPeer := switchA.Peers().Get(peerIDToKey(hostB.ID()))
		require.NotNil(t, reconnectedPeer)

		// Check for expected log messages
		require.Contains(t, logBuffer.String(), "Removing peer")
		require.Contains(t, logBuffer.String(), "Will reconnect to peer after transient error")
		require.Contains(t, logBuffer.String(), "Reconnected to peer")
	})

	t.Run("EnsureScalers", func(t *testing.T) {
		// ARRANGE
		var (
			port      = utils.GetFreePorts(t, 1)[0]
			logBuffer = &syncBuffer{}
			logger    = log.NewTMLogger(logBuffer)
		)

		// Given default P2P config with lp2p enabled
		cfg := config.DefaultP2PConfig()
		cfg.RootDir = t.TempDir()
		cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", port)
		cfg.ExternalAddress = fmt.Sprintf("127.0.0.1:%d", port)
		cfg.LibP2PConfig.Enabled = true

		// Given a new host
		host, err := NewHost(cfg, ed25519.GenPrivKey(), logger)
		require.NoError(t, err)

		t.Cleanup(func() { _ = host.Close() })

		// Given two dummy reactors
		consensusReactor := p2pmock.NewReactor()
		consensusReactor.Channels = []*conn.ChannelDescriptor{
			{ID: 0x01, Priority: 1},
		}

		mempoolReactor := p2pmock.NewReactor()
		mempoolReactor.Channels = []*conn.ChannelDescriptor{
			{ID: 0x02, Priority: 1},
		}

		// ACT
		// Create switch which should add reactors with scalers
		sw, err := NewSwitch(
			nil,
			host,
			[]SwitchReactor{
				{Name: "CONSENSUS", Reactor: consensusReactor},
				{Name: "MEMPOOL", Reactor: mempoolReactor},
			},
			p2p.NopMetrics(),
			logger,
		)

		// ASSERT
		require.NoError(t, err)

		r1 := sw.reactors.reactors[0]
		require.Equal(t, "CONSENSUS", r1.name)

		r2 := sw.reactors.reactors[1]
		require.Equal(t, "MEMPOOL", r2.name)

		// Check logs
		// see config.DefaultLibP2PConfig()
		require.True(t, logBuffer.HasMatchingLine(
			"Added reactor", "reactor=CONSENSUS",
			"Workers[min:4, max:32]",
			"Threshold=100ms",
		))
		require.True(t, logBuffer.HasMatchingLine(
			"Added reactor", "reactor=MEMPOOL",
			"Workers[min:8, max:512]", "Threshold=500ms",
		))
	})

	t.Run("EndToEndFlow", func(t *testing.T) {
		// ARRANGE
		const channelID = 0xF1

		// Given 4 hosts
		hosts := makeTestHosts(t, 4, withLogging())
		hostA, hostB, hostC, hostD := hosts[0], hosts[1], hosts[2], hosts[3]

		// Given a common channel descriptor
		channelDescriptor := &conn.ChannelDescriptor{
			ID:                  channelID,
			Priority:            1,
			RecvMessageCapacity: 1024,
			MessageType:         &types.RequestEcho{},
		}

		switchMaker := func(host *Host) (*Switch, *reactorMock) {
			reactor := newReactorMock([]*conn.ChannelDescriptor{channelDescriptor}, host.Logger())
			sw, err := NewSwitch(
				nil,
				host,
				[]SwitchReactor{
					{Name: "echoReactor", Reactor: reactor},
				},
				p2p.NopMetrics(),
				host.Logger(),
			)
			require.NoError(t, err)

			return sw, reactor
		}

		// Given 4 switches
		switchA, reactorA := switchMaker(hostA)
		switchB, reactorB := switchMaker(hostB)
		switchC, reactorC := switchMaker(hostC)
		switchD, reactorD := switchMaker(hostD)

		// Connect all switches to each other.
		connectSwitches(t, []*Switch{switchA, switchB, switchC, switchD})

		// Given D is disconnected from A.
		peerDInA := switchA.Peers().Get(peerIDToKey(hostD.ID()))
		require.NotNil(t, peerDInA)
		switchA.StopPeerForError(peerDInA, "disconnect D from A")
		require.Eventually(t, func() bool {
			return switchA.Peers().Get(peerIDToKey(hostD.ID())) == nil
		}, time.Second, 20*time.Millisecond)

		// ACT
		// Broadcast message from A to everyone.
		switchA.BroadcastAsync(p2p.Envelope{
			ChannelID: channelID,
			Message:   &types.RequestEcho{Message: "to infinity and beyond"},
		})

		// ASSERT
		// All switches use the same protocol derived from channelID.
		msgsChecker := func(reactor *reactorMock) func() bool {
			return func() bool {
				msgs := reactor.receivedEnvelopes()
				return len(msgs) == 1 && msgs[0].Message.(*types.RequestEcho).Message == "to infinity and beyond"
			}
		}

		require.Eventually(t, msgsChecker(reactorB), time.Second, 20*time.Millisecond)
		require.Eventually(t, msgsChecker(reactorC), time.Second, 20*time.Millisecond)

		// A should not receive own broadcast and D is disconnected from A.
		require.Len(t, reactorA.receivedEnvelopes(), 0)
		require.Len(t, reactorD.receivedEnvelopes(), 0)
	})

	t.Run("MsgBytesFilterRejects", func(t *testing.T) {
		// ARRANGE
		const channelID = 0xF2

		// Given 2 hosts: A and B
		hosts := makeTestHosts(t, 2, withLogging())
		hostA, hostB := hosts[0], hosts[1]

		// Given a common channel descriptor
		channelDescriptor := &conn.ChannelDescriptor{
			ID:                  channelID,
			Priority:            1,
			RecvMessageCapacity: 1024,
			MessageType:         &types.RequestEcho{},
		}

		// Given switch A with a vanilla reactor
		reactorA := newReactorMock([]*conn.ChannelDescriptor{channelDescriptor}, hostA.Logger())
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{
				{Name: "echoReactor", Reactor: reactorA},
			},
			p2p.NopMetrics(),
			hostA.Logger(),
		)
		require.NoError(t, err)

		// Given switch B with a reactor that rejects all bytes on channelID
		reactorB := newFilteringReactor(
			[]*conn.ChannelDescriptor{channelDescriptor},
			hostB.Logger(),
			channelID,
			errors.New("rejected by filter for test"),
		)
		switchB, err := NewSwitch(
			nil,
			hostB,
			[]SwitchReactor{
				{Name: "echoReactor", Reactor: reactorB},
			},
			p2p.NopMetrics(),
			hostB.Logger(),
		)
		require.NoError(t, err)

		// Connect A and B
		connectSwitches(t, []*Switch{switchA, switchB})

		// Pre-condition: A has B as a peer (A bootstrapped to B).
		require.Eventually(t, func() bool {
			return switchA.Peers().Size() == 1
		}, time.Second, 20*time.Millisecond, "A should see B")

		// ACT: Broadcast message from A to B. B's filter should reject before
		// proto.Unmarshal runs and before reactor.Receive is called.
		switchA.BroadcastAsync(p2p.Envelope{
			ChannelID: channelID,
			Message:   &types.RequestEcho{Message: "should be filtered"},
		})

		// ASSERT #1: B's FilterMsgBytes was invoked
		require.Eventually(t, func() bool {
			return reactorB.filterCalls.Load() >= 1
		}, 2*time.Second, 20*time.Millisecond, "filter should have been called")

		// ASSERT #2: B's Receive was never invoked (rejection happened pre-unmarshal)
		require.Empty(t, reactorB.receivedEnvelopes(),
			"Receive must not be called when FilterMsgBytes rejects bytes")

		// ASSERT #3: B disconnected A via StopPeerForError
		require.Eventually(t, func() bool {
			return switchB.Peers().Size() == 0
		}, 2*time.Second, 20*time.Millisecond, "B should have disconnected A after filter rejection")
	})

	t.Run("MsgBytesFilterAllows", func(t *testing.T) {
		// ARRANGE
		const channelID = 0xF3

		// Given 2 hosts: A and B
		hosts := makeTestHosts(t, 2, withLogging())
		hostA, hostB := hosts[0], hosts[1]

		// Given a common channel descriptor
		channelDescriptor := &conn.ChannelDescriptor{
			ID:                  channelID,
			Priority:            1,
			RecvMessageCapacity: 1024,
			MessageType:         &types.RequestEcho{},
		}

		// Given switch A with a vanilla reactor
		reactorA := newReactorMock([]*conn.ChannelDescriptor{channelDescriptor}, hostA.Logger())
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{
				{Name: "echoReactor", Reactor: reactorA},
			},
			p2p.NopMetrics(),
			hostA.Logger(),
		)
		require.NoError(t, err)

		// Given switch B with a filtering reactor that allows everything (rejectErr == nil)
		reactorB := newFilteringReactor(
			[]*conn.ChannelDescriptor{channelDescriptor},
			hostB.Logger(),
			channelID,
			nil,
		)
		switchB, err := NewSwitch(
			nil,
			hostB,
			[]SwitchReactor{
				{Name: "echoReactor", Reactor: reactorB},
			},
			p2p.NopMetrics(),
			hostB.Logger(),
		)
		require.NoError(t, err)

		// Connect A and B
		connectSwitches(t, []*Switch{switchA, switchB})

		// Pre-condition: A has B as a peer (A bootstrapped to B). Without
		// this wait the broadcast can fire before the connection is fully
		// established and the message is dropped silently.
		require.Eventually(t, func() bool {
			return switchA.Peers().Size() == 1
		}, time.Second, 20*time.Millisecond, "A should see B")

		// ACT: Broadcast message from A to B
		switchA.BroadcastAsync(p2p.Envelope{
			ChannelID: channelID,
			Message:   &types.RequestEcho{Message: "passes filter"},
		})

		// ASSERT #1: B's Receive fires with the original message
		require.Eventually(t, func() bool {
			envs := reactorB.receivedEnvelopes()
			return len(envs) == 1 &&
				envs[0].Message.(*types.RequestEcho).Message == "passes filter"
		}, 2*time.Second, 20*time.Millisecond, "Receive should fire when filter allows")

		// ASSERT #2: B consulted the filter at least once
		require.GreaterOrEqual(t, reactorB.filterCalls.Load(), int32(1),
			"filter should have been consulted")

		// ASSERT #3: A is still B's peer
		require.Equal(t, 1, switchB.Peers().Size(),
			"B should still be connected to A on filter pass")
	})
}

// filteringReactor is a mock reactor that optionally filters messages via the
// MsgByteFilter implementation
type filteringReactor struct {
	*reactorMock
	filterCalls atomic.Int32
	rejectErr   error // if non-nil, FilterMsgBytes returns this for matching channel
	channelID   byte
}

func newFilteringReactor(
	channels []*conn.ChannelDescriptor,
	logger log.Logger,
	channelID byte,
	rejectErr error,
) *filteringReactor {
	return &filteringReactor{
		reactorMock: newReactorMock(channels, logger),
		channelID:   channelID,
		rejectErr:   rejectErr,
	}
}

func (r *filteringReactor) FilterMsgBytes(chID byte, _ p2p.Peer, _ []byte) error {
	if chID != r.channelID {
		return nil
	}
	r.filterCalls.Add(1)
	return r.rejectErr
}

// syncBuffer is a thread-safe bytes.Buffer.
type syncBuffer struct {
	buf bytes.Buffer
	mu  sync.RWMutex
}

func (b *syncBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.String()
}

func (b *syncBuffer) HasMatchingLine(conditions ...string) bool {
	lines := strings.Split(b.String(), "\n")

	matcher := func(line string) bool {
		for _, condition := range conditions {
			if !strings.Contains(line, condition) {
				return false
			}
		}

		return true
	}

	for _, line := range lines {
		if matcher(line) {
			return true
		}
	}

	return false
}

func patchAddrInfoIPToDNS(t *testing.T, addrInfo peer.AddrInfo) peer.AddrInfo {
	const (
		expect  = "/ip4/127.0.0.1"
		replace = "/dns/localhost"
	)

	require.Len(t, addrInfo.Addrs, 1)
	addr := addrInfo.Addrs[0]

	require.True(t, strings.HasPrefix(addr.String(), expect))

	addrNewRaw := strings.Replace(addr.String(), expect, replace, 1)
	addrNew, err := ma.NewMultiaddr(addrNewRaw)
	require.NoError(t, err)

	return peer.AddrInfo{
		ID:    addrInfo.ID,
		Addrs: []ma.Multiaddr{addrNew},
	}
}

func connectSwitches(t *testing.T, switches []*Switch) {
	t.Helper()

	ctx := context.Background()

	for i, sw1 := range switches {
		for j := i + 1; j < len(switches); j++ {
			sw2 := switches[j]

			err := sw1.bootstrapPeer(ctx, sw2.host.AddrInfo(), PeerAddOptions{
				Persistent: true,
			})

			require.NoError(t, err)
		}

		require.NoError(t, sw1.Start())
		t.Cleanup(func() { _ = sw1.Stop() })
	}
}
