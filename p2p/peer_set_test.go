package p2p

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/service"
)

// mockPeer for testing the PeerSet
type mockPeer struct {
	service.BaseService
	ip net.IP
	id ID
}

func (mp *mockPeer) FlushStop()               { mp.Stop() } //nolint:errcheck // ignore error
func (mp *mockPeer) TrySend(Envelope) bool    { return true }
func (mp *mockPeer) Send(Envelope) bool       { return true }
func (mp *mockPeer) NodeInfo() NodeInfo       { return DefaultNodeInfo{} }
func (mp *mockPeer) Status() ConnectionStatus { return ConnectionStatus{} }
func (mp *mockPeer) ID() ID                   { return mp.id }
func (mp *mockPeer) IsOutbound() bool         { return false }
func (mp *mockPeer) IsPersistent() bool       { return true }
func (mp *mockPeer) Get(s string) any         { return s }
func (mp *mockPeer) Set(string, any)          {}
func (mp *mockPeer) RemoteIP() net.IP         { return mp.ip }
func (mp *mockPeer) SocketAddr() *NetAddress  { return nil }
func (mp *mockPeer) RemoteAddr() net.Addr     { return &net.TCPAddr{IP: mp.ip, Port: 8800} }
func (mp *mockPeer) CloseConn() error         { return nil }
func (mp *mockPeer) SetRemovalFailed()        {}
func (mp *mockPeer) GetRemovalFailed() bool   { return false }

// Returns a mock peer
func newMockPeer(ip net.IP) *mockPeer {
	if ip == nil {
		ip = net.IP{127, 0, 0, 1}
	}
	nodeKey := NodeKey{PrivKey: ed25519.GenPrivKey()}
	return &mockPeer{
		ip: ip,
		id: nodeKey.ID(),
	}
}

func TestPeerSetAddRemoveOne(t *testing.T) {
	t.Parallel()

	peerSet := NewPeerSet()

	var peerList []Peer
	for i := 0; i < 5; i++ {
		p := newMockPeer(net.IP{127, 0, 0, byte(i)})
		if err := peerSet.Add(p); err != nil {
			t.Error(err)
		}
		peerList = append(peerList, p)
	}

	n := len(peerList)
	// 1. Test removing from the front
	for i, peerAtFront := range peerList {
		removed := peerSet.Remove(peerAtFront)
		assert.True(t, removed)
		wantSize := n - i - 1
		for j := 0; j < 2; j++ {
			assert.Equal(t, false, peerSet.Has(peerAtFront.ID()), "#%d Run #%d: failed to remove peer", i, j)
			assert.Equal(t, wantSize, peerSet.Size(), "#%d Run #%d: failed to remove peer and decrement size", i, j)
			// Test the route of removing the now non-existent element
			removed := peerSet.Remove(peerAtFront)
			assert.False(t, removed)
		}
	}

	// 2. Next we are testing removing the peer at the end
	// a) Replenish the peerSet
	for _, peer := range peerList {
		if err := peerSet.Add(peer); err != nil {
			t.Error(err)
		}
	}

	// b) In reverse, remove each element
	for i := n - 1; i >= 0; i-- {
		peerAtEnd := peerList[i]
		removed := peerSet.Remove(peerAtEnd)
		assert.True(t, removed)
		assert.Equal(t, false, peerSet.Has(peerAtEnd.ID()), "#%d: failed to remove item at end", i)
		assert.Equal(t, i, peerSet.Size(), "#%d: differing sizes after peerSet.Remove(atEndPeer)", i)
	}
}

func TestPeerSetAddRemoveMany(t *testing.T) {
	t.Parallel()
	peerSet := NewPeerSet()

	peers := []Peer{}
	N := 100
	for i := 0; i < N; i++ {
		peer := newMockPeer(net.IP{127, 0, 0, byte(i)})
		if err := peerSet.Add(peer); err != nil {
			t.Errorf("failed to add new peer")
		}
		if peerSet.Size() != i+1 {
			t.Errorf("failed to add new peer and increment size")
		}
		peers = append(peers, peer)
	}

	for i, peer := range peers {
		removed := peerSet.Remove(peer)
		assert.True(t, removed)
		if peerSet.Has(peer.ID()) {
			t.Errorf("failed to remove peer")
		}
		if peerSet.Size() != len(peers)-i-1 {
			t.Errorf("failed to remove peer and decrement size")
		}
	}
}

func TestPeerSetAddDuplicate(t *testing.T) {
	t.Parallel()
	peerSet := NewPeerSet()
	peer := newMockPeer(nil)

	n := 20
	errsChan := make(chan error)
	// Add the same asynchronously to test the
	// concurrent guarantees of our APIs, and
	// our expectation in the end is that only
	// one addition succeeded, but the rest are
	// instances of ErrSwitchDuplicatePeer.
	for i := 0; i < n; i++ {
		go func() {
			errsChan <- peerSet.Add(peer)
		}()
	}

	// Now collect and tally the results
	errsTally := make(map[string]int)
	for i := 0; i < n; i++ {
		err := <-errsChan

		switch err.(type) {
		case ErrSwitchDuplicatePeerID:
			errsTally["duplicateID"]++
		default:
			errsTally["other"]++
		}
	}

	// Our next procedure is to ensure that only one addition
	// succeeded and that the rest are each ErrSwitchDuplicatePeer.
	wantErrCount, gotErrCount := n-1, errsTally["duplicateID"]
	assert.Equal(t, wantErrCount, gotErrCount, "invalid ErrSwitchDuplicatePeer count")

	wantNilErrCount, gotNilErrCount := 1, errsTally["other"]
	assert.Equal(t, wantNilErrCount, gotNilErrCount, "invalid nil errCount")
}

func TestPeerSetGet(t *testing.T) {
	t.Parallel()

	var (
		peerSet = NewPeerSet()
		peer    = newMockPeer(nil)
	)

	assert.Nil(t, peerSet.Get(peer.ID()), "expecting a nil lookup, before .Add")

	if err := peerSet.Add(peer); err != nil {
		t.Fatalf("Failed to add new peer: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		// Add them asynchronously to test the
		// concurrent guarantees of our APIs.
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			have, want := peerSet.Get(peer.ID()), peer
			assert.Equal(t, have, want, "%d: have %v, want %v", i, have, want)
		}(i)
	}
	wg.Wait()
}

func TestPeerSetForEachAllowsMutation(t *testing.T) {
	t.Parallel()

	ps := NewPeerSet()
	total := 10
	for i := 0; i < total; i++ {
		p := newMockPeer(net.IP{127, 0, 0, byte(i)})
		assert.NoError(t, ps.Add(p))
	}

	visited := make(map[ID]int)

	ps.ForEach(func(p Peer) {
		visited[p.ID()]++
		ps.Remove(p)
	})

	assert.Equal(t, total, len(visited), "expected to visit all peers once")
	assert.Equal(t, 0, ps.Size(), "removing peers during iteration should succeed")
}

func TestPeerSetForEachSnapshotUnderChurn(t *testing.T) {
	t.Parallel()

	ps := NewPeerSet()
	const basePeers = 32
	for i := 0; i < basePeers; i++ {
		assert.NoError(t, ps.Add(newMockPeer(net.IP{10, 0, 0, byte(i + 1)})))
	}

	errCh := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 500; i++ {
			p := newMockPeer(net.IP{192, 0, 2, byte(i % 250)})
			if err := ps.Add(p); err != nil {
				select {
				case errCh <- fmt.Errorf("add failed: %w", err):
				default:
				}
				return
			}
			if removed := ps.Remove(p); !removed {
				select {
				case errCh <- fmt.Errorf("remove failed for %s", p.ID()):
				default:
				}
				return
			}
			time.Sleep(100 * time.Microsecond)
		}
	}()

	for iteration := 0; iteration < 500; iteration++ {
		select {
		case err := <-errCh:
			t.Fatal(err)
		default:
		}

		seen := make(map[ID]struct{}, basePeers)
		ps.ForEach(func(peer Peer) {
			if peer == nil {
				select {
				case errCh <- fmt.Errorf("nil peer observed at iteration %d", iteration):
				default:
				}
				return
			}
			if _, ok := seen[peer.ID()]; ok {
				select {
				case errCh <- fmt.Errorf("duplicate peer %s observed", peer.ID()):
				default:
				}
				return
			}
			seen[peer.ID()] = struct{}{}
		})
	}

	<-done
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestPeerSetBroadcastThroughputUnderChurn(t *testing.T) {
	t.Parallel()

	ps := NewPeerSet()
	const basePeers = 64
	for i := 0; i < basePeers; i++ {
		assert.NoError(t, ps.Add(newMockPeer(net.IP{172, 16, 0, byte(i + 1)})))
	}

	var (
		addDurationNS       int64
		removeDurationNS    int64
		broadcastDurationNS int64
		addCount            int64
		removeCount         int64
		broadcastCount      int64
	)

	churnDone := make(chan struct{})
	go func() {
		defer close(churnDone)
		for i := 0; i < 200; i++ {
			p := newMockPeer(net.IP{198, 51, 100, byte(i % 250)})

			start := time.Now()
			err := ps.Add(p)
			atomic.AddInt64(&addDurationNS, time.Since(start).Nanoseconds())
			if err == nil {
				atomic.AddInt64(&addCount, 1)
			}

			start = time.Now()
			if ps.Remove(p) {
				atomic.AddInt64(&removeDurationNS, time.Since(start).Nanoseconds())
				atomic.AddInt64(&removeCount, 1)
			}

			time.Sleep(50 * time.Microsecond)
		}
	}()

	for i := 0; i < 200; i++ {
		start := time.Now()
		var visited int
		ps.ForEach(func(peer Peer) {
			if peer != nil {
				visited++
			}
		})
		atomic.AddInt64(&broadcastDurationNS, time.Since(start).Nanoseconds())
		atomic.AddInt64(&broadcastCount, 1)

		if visited == 0 {
			t.Fatalf("broadcast iteration %d visited zero peers", i)
		}

		time.Sleep(50 * time.Microsecond)
	}

	<-churnDone

	getAvg := func(totalNS, count int64) time.Duration {
		if count == 0 {
			return 0
		}
		return time.Duration(totalNS / count)
	}

	t.Logf("avg peer add latency: %s (%d ops)", getAvg(atomic.LoadInt64(&addDurationNS), atomic.LoadInt64(&addCount)), atomic.LoadInt64(&addCount))
	t.Logf("avg peer remove latency: %s (%d ops)", getAvg(atomic.LoadInt64(&removeDurationNS), atomic.LoadInt64(&removeCount)), atomic.LoadInt64(&removeCount))
	t.Logf("avg broadcast iteration latency: %s (%d ops)", getAvg(atomic.LoadInt64(&broadcastDurationNS), atomic.LoadInt64(&broadcastCount)), atomic.LoadInt64(&broadcastCount))

	if broadcastCount == 0 {
		t.Fatal("expected broadcast iterations to run")
	}
	if addCount == 0 || removeCount == 0 {
		t.Fatal("expected churn goroutine to add and remove peers")
	}
}
