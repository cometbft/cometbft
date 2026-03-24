package lp2p

import (
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	tmp2p "github.com/cometbft/cometbft/proto/tendermint/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReactorSet(t *testing.T) {
	t.Run("reactors", func(t *testing.T) {
		// ARRANGE
		configOverride := func(cfg *config.LibP2PConfig) {
			cfg.Scaler.Overrides = []config.LibP2PScalerOverride{
				{
					Reactor:          "B",
					ThresholdLatency: 999 * time.Millisecond,
				},
			}
		}

		ts := newReactorSetTestSuite(t, withLogging(), withModifiedConfig(configOverride))
		rs := newReactorSet(ts.sw)

		var (
			reactorA = ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xA1}})
			reactorB = ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xB1}})
			reactorC = ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xA1}})
		)

		// ACT + ASSERT #1: add A
		require.NoError(t, rs.Add(reactorA, "A"))
		require.Len(t, rs.reactors, 1)

		// ACT + ASSERT #2: add A again
		err := rs.Add(ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xA2}}), "A")
		require.ErrorContains(t, err, `reactor "A" is already registered`)

		// ACT + ASSERT #3: add B
		require.NoError(t, rs.Add(reactorB, "B"))
		require.Len(t, rs.reactors, 2)

		// ACT + ASSERT #4: add C with same protocol as A
		err = rs.Add(reactorC, "C")
		require.ErrorContains(t, err, "is already registered")

		// ACT + ASSERT #5: get by name B
		byNameB, ok := rs.GetByName("B")
		require.True(t, ok)
		assert.Same(t, reactorB, byNameB)

		// check that B has a custom config
		reactorItemB, _ := rs.getByName("B")
		require.Contains(t, reactorItemB.consumerQueue.Scaler().String(), "999ms")

		// ACT + ASSERT #6: get by name D
		byNameD, ok := rs.GetByName("D")
		require.False(t, ok)
		require.Nil(t, byNameD)

		// ACT: start reactor set
		startedProtocols := []protocol.ID{}
		err = rs.Start(func(id protocol.ID) {
			startedProtocols = append(startedProtocols, id)
		})

		// ASSERT #7: Start -> ok
		require.NoError(t, err)
		require.True(t, reactorA.IsRunning())
		require.True(t, reactorB.IsRunning())
		require.Len(t, startedProtocols, 2)
		require.ElementsMatch(t, []protocol.ID{ProtocolID(0xA1), ProtocolID(0xB1)}, startedProtocols)

		// ACT: stop reactor set
		rs.Stop()

		// ASSERT #8: Stop -> ok
		require.False(t, reactorA.IsRunning())
		require.False(t, reactorB.IsRunning())
	})

	t.Run("peers", func(t *testing.T) {
		// ARRANGE
		ts := newReactorSetTestSuite(t)
		rs := newReactorSet(ts.sw)

		reactorA := ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xC1}})
		reactorB := ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xD1}})

		require.NoError(t, rs.Add(reactorA, "A"))
		require.NoError(t, rs.Add(reactorB, "B"))

		peer1 := &Peer{}
		peer2 := &Peer{}

		// ACT #1: add peers
		rs.AddPeer(peer1)
		rs.AddPeer(peer2)

		// ACT #2: init peers
		rs.InitPeer(peer1)
		rs.InitPeer(peer2)

		// ACT #3: remove peers
		rs.RemovePeer(peer1, "manual_disconnect_1")
		rs.RemovePeer(peer2, "manual_disconnect_2")

		// ASSERT
		for _, reactor := range []*reactorMock{reactorA, reactorB} {
			require.Len(t, reactor.addPeers, 2)
			require.Same(t, peer1, reactor.addPeers[0].(*Peer))
			require.Same(t, peer2, reactor.addPeers[1].(*Peer))

			require.Len(t, reactor.initPeers, 2)
			require.Same(t, peer1, reactor.initPeers[0].(*Peer))
			require.Same(t, peer2, reactor.initPeers[1].(*Peer))

			require.Len(t, reactor.removedPeers, 2)
			require.Same(t, peer1, reactor.removedPeers[0].(*Peer))
			require.Same(t, peer2, reactor.removedPeers[1].(*Peer))
		}
	})

	t.Run("receive", func(t *testing.T) {
		// ARRANGE
		ts := newReactorSetTestSuite(t)
		rs := newReactorSet(ts.sw)

		reactorA := ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xE1}})
		reactorB := ts.newReactor([]*conn.ChannelDescriptor{{ID: 0xE2}})

		require.NoError(t, rs.Add(reactorA, "A"))
		require.NoError(t, rs.Add(reactorB, "B"))
		require.NoError(t, rs.Start(func(protocol.ID) {}))
		t.Cleanup(rs.Stop)

		envelopeA := p2p.Envelope{
			ChannelID: 0xE1,
			Message:   &tmp2p.PexRequest{},
		}
		envelopeUnknown := p2p.Envelope{
			ChannelID: 0xEE,
			Message:   &tmp2p.PexRequest{},
		}

		// ACT #1: receive for known reactor A
		rs.Receive("A", "PexRequest", envelopeA, 5)

		// ASSERT #1: envelope is processed by reactor A
		require.Eventually(t, func() bool {
			return len(reactorA.receivedEnvelopes()) == 1
		}, 2*time.Second, 10*time.Millisecond)

		receivedA := reactorA.receivedEnvelopes()
		require.Len(t, receivedA, 1)
		assert.Equal(t, envelopeA.ChannelID, receivedA[0].ChannelID)
		assert.IsType(t, &tmp2p.PexRequest{}, receivedA[0].Message)
		assert.Len(t, reactorB.receivedEnvelopes(), 0)

		// ACT #2: receive for unknown reactor name
		rs.Receive("unknown", "FooBar", envelopeUnknown, 1)

		// ASSERT #2: no additional messages are processed
		assert.Len(t, reactorA.receivedEnvelopes(), 1)
		assert.Len(t, reactorB.receivedEnvelopes(), 0)
	})
}

type reactorSetTestSuite struct {
	t  *testing.T
	sw *Switch
}

type reactorMock struct {
	p2p.BaseReactor

	mu sync.Mutex

	channels     []*conn.ChannelDescriptor
	addPeers     []p2p.Peer
	initPeers    []p2p.Peer
	removedPeers []p2p.Peer
	received     []p2p.Envelope
}

func newReactorSetTestSuite(t *testing.T, opts ...testOption) *reactorSetTestSuite {
	var (
		logBuffer = &syncBuffer{}
		logger    = log.NewTMLogger(logBuffer)
	)

	ports := utils.GetFreePorts(t, 1)
	host := makeTestHost(t, ports[0], opts...)

	sw, err := NewSwitch(nil, host, nil, p2p.NopMetrics(), logger)
	require.NoError(t, err)

	return &reactorSetTestSuite{
		t:  t,
		sw: sw,
	}
}

func newReactorMock(channels []*conn.ChannelDescriptor, logger log.Logger) *reactorMock {
	r := &reactorMock{
		channels: channels,
	}
	r.BaseReactor = *p2p.NewBaseReactor("ReactorMock", r)
	r.SetLogger(logger)

	return r
}

func (ts *reactorSetTestSuite) newReactor(channels []*conn.ChannelDescriptor) *reactorMock {
	return newReactorMock(channels, ts.sw.Logger)
}

func (r *reactorMock) GetChannels() []*conn.ChannelDescriptor {
	return r.channels
}

func (r *reactorMock) InitPeer(peer p2p.Peer) p2p.Peer {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.initPeers = append(r.initPeers, peer)
	return peer
}

func (r *reactorMock) AddPeer(peer p2p.Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.addPeers = append(r.addPeers, peer)
}

func (r *reactorMock) RemovePeer(peer p2p.Peer, _ any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.removedPeers = append(r.removedPeers, peer)
}

func (r *reactorMock) Receive(e p2p.Envelope) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.received = append(r.received, e)
}

func (r *reactorMock) receivedEnvelopes() []p2p.Envelope {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := make([]p2p.Envelope, len(r.received))
	copy(cp, r.received)
	return cp
}
