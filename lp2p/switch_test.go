package lp2p

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/stretchr/testify/require"
)

func TestSwitch(t *testing.T) {
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
			hostA = makeTestHost(t, ports[0], []config.LibP2PBootstrapPeer{}, true)
			hostB = makeTestHost(t, ports[1], []config.LibP2PBootstrapPeer{}, true)
			hostC = makeTestHost(t, ports[2], []config.LibP2PBootstrapPeer{}, true)
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

		// Connect host A to B (non-persistent) and C (persistent)
		err = switchA.connectPeer(ctx, hostB.AddrInfo(), PeerAddOptions{
			Persistent: false,
		})
		require.NoError(t, err)

		err = switchA.connectPeer(ctx, hostC.AddrInfo(), PeerAddOptions{
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
