package lp2p

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
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

		t.Cleanup(func() {
			hostB.Close()
			hostA.Close()
		})

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

		t.Cleanup(func() {
			hostC.Close()
			hostB.Close()
			hostA.Close()
		})

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

		t.Cleanup(func() {
			hostB.Close()
			hostA.Close()
		})

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
		transientErr := &ErrorTransient{Err: errors.New("something went wrong")}
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
