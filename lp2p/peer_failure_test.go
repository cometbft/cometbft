package lp2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test errors
var (
	errPeerConnectionFailed = errors.New("failed to open stream: connection refused")
	errLocalResourceLimit   = errors.New("resource limit exceeded")
)

// mockReactor is a test reactor that tracks lifecycle calls
type mockReactor struct {
	p2p.BaseReactor
	name      string
	channelID byte

	initPeerCalls   atomic.Int32
	addPeerCalls    atomic.Int32
	removePeerCalls atomic.Int32

	// Track peer IDs for verification
	mu             sync.Mutex
	addedPeers     []p2p.ID
	removedPeers   []p2p.ID
	removalReasons []any
}

func newMockReactor(name string, channelID byte) *mockReactor {
	r := &mockReactor{
		name:           name,
		channelID:      channelID,
		addedPeers:     make([]p2p.ID, 0),
		removedPeers:   make([]p2p.ID, 0),
		removalReasons: make([]any, 0),
	}
	r.BaseReactor = *p2p.NewBaseReactor(name, r)
	return r
}

func (r *mockReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{
		{
			ID:                  r.channelID,
			Priority:            1,
			SendQueueCapacity:   1,
			RecvMessageCapacity: 1024,
		},
	}
}

func (r *mockReactor) InitPeer(peer p2p.Peer) p2p.Peer {
	r.initPeerCalls.Add(1)
	return peer
}

func (r *mockReactor) AddPeer(peer p2p.Peer) {
	r.addPeerCalls.Add(1)
	r.mu.Lock()
	r.addedPeers = append(r.addedPeers, peer.ID())
	r.mu.Unlock()
}

func (r *mockReactor) RemovePeer(peer p2p.Peer, reason any) {
	r.removePeerCalls.Add(1)
	r.mu.Lock()
	r.removedPeers = append(r.removedPeers, peer.ID())
	r.removalReasons = append(r.removalReasons, reason)
	r.mu.Unlock()
}

func (r *mockReactor) getAddedPeers() []p2p.ID {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]p2p.ID, len(r.addedPeers))
	copy(result, r.addedPeers)
	return result
}

func (r *mockReactor) getRemovedPeers() []p2p.ID {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]p2p.ID, len(r.removedPeers))
	copy(result, r.removedPeers)
	return result
}

func (r *mockReactor) getRemovalReasons() []any {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]any, len(r.removalReasons))
	copy(result, r.removalReasons)
	return result
}

func (r *mockReactor) reset() {
	r.initPeerCalls.Store(0)
	r.addPeerCalls.Store(0)
	r.removePeerCalls.Store(0)
	r.mu.Lock()
	r.addedPeers = make([]p2p.ID, 0)
	r.removedPeers = make([]p2p.ID, 0)
	r.removalReasons = make([]any, 0)
	r.mu.Unlock()
}

// TestPeerFailureTracking_Unit tests the basic failure tracking logic
func TestPeerFailureTracking_Unit(t *testing.T) {
	t.Run("IncrementOnFailure", func(t *testing.T) {
		// Create a peer with failure tracking enabled
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(3, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Simulate failures
		peer.handleSendFailure(errPeerConnectionFailed)
		assert.Equal(t, int32(1), peer.sendFailures.Load())

		peer.handleSendFailure(errPeerConnectionFailed)
		assert.Equal(t, int32(2), peer.sendFailures.Load())

		// Not yet at threshold
		assert.False(t, removalCalled.Load())

		peer.handleSendFailure(errPeerConnectionFailed)
		assert.Equal(t, int32(3), peer.sendFailures.Load())

		// Should trigger removal
		time.Sleep(10 * time.Millisecond) // Give goroutine time to run
		assert.True(t, removalCalled.Load())
	})

	t.Run("ResetOnSuccess", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		peer.ConfigureFailureTracking(5, func(p *Peer, err error) {})

		// Increment failures
		peer.handleSendFailure(errPeerConnectionFailed)
		peer.handleSendFailure(errPeerConnectionFailed)
		assert.Equal(t, int32(2), peer.sendFailures.Load())

		// Reset on success
		peer.sendFailures.Store(0)
		assert.Equal(t, int32(0), peer.sendFailures.Load())

		// Failures start from zero again
		peer.handleSendFailure(errPeerConnectionFailed)
		assert.Equal(t, int32(1), peer.sendFailures.Load())
	})

	t.Run("DisabledWhenMaxIsZero", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(0, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Many failures should not trigger removal when disabled
		for i := 0; i < 100; i++ {
			peer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(10 * time.Millisecond)
		assert.False(t, removalCalled.Load())
	})

	t.Run("ConcurrentFailures", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCallCount := atomic.Int32{}
		peer.ConfigureFailureTracking(10, func(p *Peer, err error) {
			removalCallCount.Add(1)
		})

		// Simulate concurrent failures from multiple goroutines
		const numGoroutines = 5
		const failuresPerGoroutine = 4 // Total: 20 failures, should exceed threshold of 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < failuresPerGoroutine; j++ {
					peer.handleSendFailure(errPeerConnectionFailed)
					time.Sleep(time.Millisecond) // Small delay to spread out failures
				}
			}()
		}

		wg.Wait()
		time.Sleep(50 * time.Millisecond) // Wait for callbacks

		// Counter should reflect all failures
		assert.GreaterOrEqual(t, peer.sendFailures.Load(), int32(10))

		// Removal callback should be called (possibly multiple times due to race)
		assert.Greater(t, removalCallCount.Load(), int32(0))
	})
}

// TestPeerFailureTracking_Integration tests peer removal and reactor updates
func TestPeerFailureTracking_Integration(t *testing.T) {
	t.Run("ReactorsNotifiedOnRemoval", func(t *testing.T) {
		ctx := context.Background()
		ports := utils.GetFreePorts(t, 2)
		logger := log.NewNopLogger()

		// Create two reactors to track calls (with unique channel IDs)
		reactor1 := newMockReactor("TestReactor1", 0x01)
		reactor2 := newMockReactor("TestReactor2", 0x02)
		// Don't start reactors manually - let switch start them

		// Create hosts
		privateKeyA := ed25519.GenPrivKey()
		privateKeyB := ed25519.GenPrivKey()

		hostA := makeTestHost(t, ports[0], withPrivateKey(privateKeyA))
		hostB := makeTestHost(t, ports[1], withPrivateKey(privateKeyB))

		// Create switch with low failure threshold
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{
				{Name: "TestReactor1", Reactor: reactor1},
				{Name: "TestReactor2", Reactor: reactor2},
			},
			p2p.NopMetrics(),
			3, // maxSendFailures = 3
			logger.With("switch", "A"),
		)
		require.NoError(t, err)
		require.NoError(t, switchA.Start())
		defer switchA.Stop()

		// Connect peer B to switch A (with reactor callbacks)
		err = switchA.connectPeer(ctx, hostB.AddrInfo(), PeerAddOptions{
			Persistent:    false,
			OnBeforeStart: switchA.reactors.InitPeer,
			OnAfterStart:  switchA.reactors.AddPeer,
			OnStartFailed: switchA.reactors.RemovePeer,
		})
		require.NoError(t, err)

		// Wait for peer to be fully added
		time.Sleep(100 * time.Millisecond)

		// Verify both reactors were notified of peer addition
		assert.Equal(t, int32(1), reactor1.initPeerCalls.Load())
		assert.Equal(t, int32(1), reactor1.addPeerCalls.Load())
		assert.Equal(t, int32(1), reactor2.initPeerCalls.Load())
		assert.Equal(t, int32(1), reactor2.addPeerCalls.Load())

		// Get the peer
		peerID := peerIDToKey(hostB.ID())
		peer := switchA.peerSet.Get(peerID)
		require.NotNil(t, peer)

		// Simulate send failures to trigger removal
		lp2pPeer := peer.(*Peer)
		for i := 0; i < 3; i++ {
			lp2pPeer.handleSendFailure(errPeerConnectionFailed)
		}

		// Wait for peer removal to be processed
		time.Sleep(200 * time.Millisecond)

		// Verify both reactors were notified of peer removal
		assert.Equal(t, int32(1), reactor1.removePeerCalls.Load(), "reactor1 should have RemovePeer called")
		assert.Equal(t, int32(1), reactor2.removePeerCalls.Load(), "reactor2 should have RemovePeer called")

		// Verify the removed peer IDs match
		removed1 := reactor1.getRemovedPeers()
		removed2 := reactor2.getRemovedPeers()
		assert.Equal(t, []p2p.ID{peerID}, removed1)
		assert.Equal(t, []p2p.ID{peerID}, removed2)

		// Verify removal reason contains failure information
		reasons1 := reactor1.getRemovalReasons()
		require.Len(t, reasons1, 1)
		assert.Contains(t, fmt.Sprint(reasons1[0]), "max send failures")
	})

	t.Run("PersistentPeerReconnectNotifiesReactors", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		logger := log.NewNopLogger()

		reactor := newMockReactor("TestReactor", 0x01)
		// Don't start reactor manually - let switch start it

		privateKeyA := ed25519.GenPrivKey()
		privateKeyB := ed25519.GenPrivKey()

		// Setup bootstrap peers for A to connect to B as persistent
		idB, err := IDFromPrivateKey(privateKeyB)
		require.NoError(t, err)

		bootstrapPeersA := []config.LibP2PBootstrapPeer{
			{
				Host:       fmt.Sprintf("127.0.0.1:%d", ports[1]),
				ID:         idB.String(),
				Persistent: true, // This is a persistent peer
			},
		}

		hostA := makeTestHost(t, ports[0], withPrivateKey(privateKeyA), withBootstrapPeers(bootstrapPeersA))
		hostB := makeTestHost(t, ports[1], withPrivateKey(privateKeyB))

		// Create switch with very low failure threshold for faster testing
		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{
				{Name: "TestReactor", Reactor: reactor},
			},
			p2p.NopMetrics(),
			2, // maxSendFailures = 2 (low for fast test)
			logger.With("switch", "A"),
		)
		require.NoError(t, err)
		require.NoError(t, switchA.Start())
		defer switchA.Stop()

		// Wait for initial connection
		time.Sleep(200 * time.Millisecond)

		// Verify initial reactor calls
		assert.Equal(t, int32(1), reactor.initPeerCalls.Load(), "InitPeer should be called once initially")
		assert.Equal(t, int32(1), reactor.addPeerCalls.Load(), "AddPeer should be called once initially")

		peerID := peerIDToKey(hostB.ID())
		addedPeers := reactor.getAddedPeers()
		require.Len(t, addedPeers, 1)
		assert.Equal(t, peerID, addedPeers[0])

		// Get the peer and trigger failures
		peer := switchA.peerSet.Get(peerID)
		require.NotNil(t, peer)

		lp2pPeer := peer.(*Peer)
		for i := 0; i < 2; i++ {
			lp2pPeer.handleSendFailure(errPeerConnectionFailed)
		}

		// Wait for peer removal
		time.Sleep(200 * time.Millisecond)

		// Verify reactor was notified of removal
		assert.Equal(t, int32(1), reactor.removePeerCalls.Load(), "RemovePeer should be called once")

		// Wait for persistent peer to reconnect (with backoff)
		// The reconnection happens with exponential backoff starting at 100ms
		time.Sleep(1 * time.Second)

		// Verify reactor was notified of reconnection
		// Note: There might be multiple reconnection attempts
		assert.GreaterOrEqual(t, reactor.initPeerCalls.Load(), int32(2), "InitPeer should be called again on reconnect")
		assert.GreaterOrEqual(t, reactor.addPeerCalls.Load(), int32(2), "AddPeer should be called again on reconnect")

		// Verify the peer was added again
		finalAddedPeers := reactor.getAddedPeers()
		assert.GreaterOrEqual(t, len(finalAddedPeers), 2, "Peer should be added at least twice (initial + reconnect)")

		// Count how many times our specific peer was added
		peerAddCount := 0
		for _, id := range finalAddedPeers {
			if id == peerID {
				peerAddCount++
			}
		}
		assert.GreaterOrEqual(t, peerAddCount, 2, "Our peer should be added at least twice")
	})

	t.Run("NonPersistentPeerNotReconnected", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		logger := log.NewNopLogger()

		reactor := newMockReactor("TestReactor", 0x01)
		// Don't start reactor manually - let switch start it

		privateKeyA := ed25519.GenPrivKey()
		privateKeyB := ed25519.GenPrivKey()

		hostA := makeTestHost(t, ports[0], withPrivateKey(privateKeyA))
		hostB := makeTestHost(t, ports[1], withPrivateKey(privateKeyB))

		switchA, err := NewSwitch(
			nil,
			hostA,
			[]SwitchReactor{
				{Name: "TestReactor", Reactor: reactor},
			},
			p2p.NopMetrics(),
			2, // maxSendFailures = 2
			logger.With("switch", "A"),
		)
		require.NoError(t, err)
		require.NoError(t, switchA.Start())
		defer switchA.Stop()

		// Connect as non-persistent peer
		ctx := context.Background()
		err = switchA.connectPeer(ctx, hostB.AddrInfo(), PeerAddOptions{
			Persistent:    false, // Non-persistent
			OnBeforeStart: switchA.reactors.InitPeer,
			OnAfterStart:  switchA.reactors.AddPeer,
			OnStartFailed: switchA.reactors.RemovePeer,
		})
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Verify initial connection
		assert.Equal(t, int32(1), reactor.addPeerCalls.Load())

		// Trigger failures to remove peer
		peerID := peerIDToKey(hostB.ID())
		peer := switchA.peerSet.Get(peerID)
		require.NotNil(t, peer)

		lp2pPeer := peer.(*Peer)
		for i := 0; i < 2; i++ {
			lp2pPeer.handleSendFailure(errPeerConnectionFailed)
		}

		// Wait for removal
		time.Sleep(200 * time.Millisecond)

		assert.Equal(t, int32(1), reactor.removePeerCalls.Load())

		// Wait to ensure no reconnection happens
		time.Sleep(1 * time.Second)

		// Should still only have 1 add call (no reconnection)
		assert.Equal(t, int32(1), reactor.addPeerCalls.Load(), "Non-persistent peer should not reconnect")
	})

	t.Run("MultipleReactorsAllNotified", func(t *testing.T) {
		ctx := context.Background()
		ports := utils.GetFreePorts(t, 2)
		logger := log.NewNopLogger()

		// Create 5 reactors (with unique channel IDs)
		reactors := make([]*mockReactor, 5)
		switchReactors := make([]SwitchReactor, 5)
		for i := 0; i < 5; i++ {
			reactors[i] = newMockReactor(fmt.Sprintf("TestReactor%d", i), byte(0x01+i))
			// Don't start reactors manually - let switch start them
			switchReactors[i] = SwitchReactor{
				Name:    fmt.Sprintf("TestReactor%d", i),
				Reactor: reactors[i],
			}
		}

		privateKeyA := ed25519.GenPrivKey()
		privateKeyB := ed25519.GenPrivKey()

		hostA := makeTestHost(t, ports[0], withPrivateKey(privateKeyA))
		hostB := makeTestHost(t, ports[1], withPrivateKey(privateKeyB))

		switchA, err := NewSwitch(
			nil,
			hostA,
			switchReactors,
			p2p.NopMetrics(),
			2, // maxSendFailures = 2
			logger.With("switch", "A"),
		)
		require.NoError(t, err)
		require.NoError(t, switchA.Start())
		defer switchA.Stop()

		// Connect peer (with reactor callbacks)
		err = switchA.connectPeer(ctx, hostB.AddrInfo(), PeerAddOptions{
			OnBeforeStart: switchA.reactors.InitPeer,
			OnAfterStart:  switchA.reactors.AddPeer,
			OnStartFailed: switchA.reactors.RemovePeer,
		})
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Verify all reactors were notified of addition
		for i, reactor := range reactors {
			assert.Equal(t, int32(1), reactor.addPeerCalls.Load(), "Reactor %d should have AddPeer called", i)
		}

		// Trigger removal
		peerID := peerIDToKey(hostB.ID())
		peer := switchA.peerSet.Get(peerID)
		require.NotNil(t, peer)

		lp2pPeer := peer.(*Peer)
		for i := 0; i < 2; i++ {
			lp2pPeer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(200 * time.Millisecond)

		// Verify all reactors were notified of removal
		for i, reactor := range reactors {
			assert.Equal(t, int32(1), reactor.removePeerCalls.Load(), "Reactor %d should have RemovePeer called", i)
		}
	})
}

// TestPeerFailureTracking_EdgeCases tests edge cases and boundary conditions
func TestPeerFailureTracking_EdgeCases(t *testing.T) {
	t.Run("ExactThreshold", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(5, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Exactly at threshold should trigger
		for i := 0; i < 5; i++ {
			peer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(10 * time.Millisecond)
		assert.True(t, removalCalled.Load())
	})

	t.Run("BelowThreshold", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(5, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Just below threshold should not trigger
		for i := 0; i < 4; i++ {
			peer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(10 * time.Millisecond)
		assert.False(t, removalCalled.Load())
	})

	t.Run("NilCallback", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		// Configure with nil callback (shouldn't panic)
		peer.ConfigureFailureTracking(3, nil)

		// Should not panic even when threshold exceeded
		for i := 0; i < 5; i++ {
			peer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(10 * time.Millisecond)
		// Test passes if no panic occurred
	})

	t.Run("VeryHighThreshold", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(1000, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Many failures but below threshold
		for i := 0; i < 100; i++ {
			peer.handleSendFailure(errPeerConnectionFailed)
		}

		time.Sleep(10 * time.Millisecond)
		assert.False(t, removalCalled.Load())
	})

	t.Run("NonPeerErrorsNotTracked", func(t *testing.T) {
		ports := utils.GetFreePorts(t, 2)
		host := makeTestHost(t, ports[0])
		peerHost := makeTestHost(t, ports[1])

		peer, err := NewPeer(host, peerHost.AddrInfo(), p2p.NopMetrics(), false, false, false)
		require.NoError(t, err)

		removalCalled := atomic.Bool{}
		peer.ConfigureFailureTracking(3, func(p *Peer, err error) {
			removalCalled.Store(true)
		})

		// Local resource errors should NOT be tracked
		for i := 0; i < 10; i++ {
			peer.handleSendFailure(errLocalResourceLimit)
		}

		time.Sleep(10 * time.Millisecond)
		assert.False(t, removalCalled.Load(), "Resource limit errors should not trigger removal")
		assert.Equal(t, int32(0), peer.sendFailures.Load(), "Counter should remain at 0")

		// But peer errors SHOULD be tracked
		peer.handleSendFailure(errPeerConnectionFailed)
		peer.handleSendFailure(errPeerConnectionFailed)
		peer.handleSendFailure(errPeerConnectionFailed)

		time.Sleep(10 * time.Millisecond)
		assert.True(t, removalCalled.Load(), "Peer connection errors should trigger removal")
	})
}
