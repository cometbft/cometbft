package gossip

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestGossipBroadcast(t *testing.T) {

	t.Run("AllToAll", func(t *testing.T) {
		const (
			hostsCount     = 10
			testDuration   = 3 * time.Second
			iterationSleep = 5 * time.Millisecond
		)

		testGossipBroadcast(t, hostsCount, testDuration, iterationSleep, connectAllNodes)
	})

	t.Run("PartialGraph", func(t *testing.T) {
		const (
			hostsCount     = 10
			testDuration   = 3 * time.Second
			iterationSleep = 10 * time.Millisecond
		)

		testGossipBroadcast(t, hostsCount, testDuration, iterationSleep, connectPartialTopology)
	})
}

func testGossipBroadcast(
	t *testing.T,
	hostsCount int,
	testDuration time.Duration,
	iterationSleep time.Duration,
	connectNodes func(t *testing.T, ctx context.Context, nodes []testNode),
) {

	// ARRANGE
	ts := newTestSuite(t, false)
	ctx := context.Background()

	// Given many nodes
	ports := utils.GetFreePorts(t, hostsCount)

	nodes := make([]testNode, 0, hostsCount)
	for _, port := range ports {
		nodes = append(nodes, ts.newNode(ctx, port))
	}

	// Connect nodes
	connectNodes(t, ctx, nodes)

	// Given receive bucket
	var (
		broadcasts = atomic.Uint64{}
		receives   = atomic.Uint64{}

		bucket     = make([]testGossipMessage, 0, hostsCount*100)
		mu         = sync.Mutex{}
		protocolID = protocol.ID("/p2p/test-protocol")

		wgReceives = sync.WaitGroup{}
	)

	// Given message handlers
	handlerMaker := func(node testNode) Handler {
		return func(protocolID protocol.ID, msg *pubsub.Message) error {
			wgReceives.Add(1)
			defer wgReceives.Done()

			ts := time.UnixMicro(int64(binary.BigEndian.Uint64(msg.Data)))

			el := testGossipMessage{
				origin:         node.address.ID,
				receiver:       msg.ReceivedFrom,
				receiveLatency: time.Since(ts),
			}

			mu.Lock()
			bucket = append(bucket, el)
			mu.Unlock()

			receives.Add(1)

			return nil
		}
	}

	// Join all nodes to the protocol
	for _, node := range nodes {
		err := node.service.Join(protocolID, handlerMaker(node))
		require.NoError(t, err, "failed to join protocol %s on node %s", protocolID, node.address.ID)
	}

	// ACT
	t.Logf("Broadcasting messages for %s", testDuration.String())

	ctx, cancel := context.WithTimeout(ctx, testDuration)
	defer cancel()

	for ctx.Err() == nil {
		// Pick random node and broadcast message
		idx := rand.Intn(len(nodes))
		node := nodes[idx]

		// Payload is a timestamp in microseconds
		nowMicro := time.Now().UTC().UnixMicro()
		payload := make([]byte, 8)
		binary.BigEndian.PutUint64(payload, uint64(nowMicro))

		err := node.service.Broadcast(protocolID, payload)
		require.NoError(t, err, "failed to broadcast message on node %s", node.address.ID)

		broadcasts.Add(1)
		time.Sleep(iterationSleep)
	}

	ts.logger.Info("Waiting for receives to complete")
	wgReceives.Wait()

	// ASSERT
	n := uint64(hostsCount)
	b := broadcasts.Load()
	r := receives.Load()

	// s*(n-1)
	idealReceives := b * (n - 1)

	t.Logf("Nodes: %d, Broadcasted messages: %d, Received messages: %d (ideal %d)", n, b, r, idealReceives)

	printGossipStats(t, bucket)
}

type testSuite struct {
	t      *testing.T
	ctx    context.Context
	logger log.Logger
}

type testNode struct {
	host    host.Host
	service *Service
	address peer.AddrInfo
}

type testGossipMessage struct {
	origin         peer.ID
	receiver       peer.ID
	receiveLatency time.Duration
}

func newTestSuite(t *testing.T, enableLogging bool) *testSuite {
	ctx := context.Background()

	logger := log.TestingLogger()
	if !enableLogging {
		logger = log.NewNopLogger()
	}

	return &testSuite{
		ctx:    ctx,
		t:      t,
		logger: logger,
	}
}

func (ts *testSuite) newHost(port int) host.Host {
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/udp/%d/quic-v1", port))
	require.NoError(ts.t, err, "failed to create multiaddr")

	host, err := libp2p.New(
		libp2p.ListenAddrs(addr),
		libp2p.UserAgent("cometbft"),
		libp2p.Transport(quic.NewTransport),
		libp2p.ResourceManager(&network.NullResourceManager{}),
	)

	require.NoError(ts.t, err, "failed to create libp2p host")

	ts.t.Cleanup(func() {
		require.NoError(ts.t, host.Close(), "failed to close host %s", host.ID())
	})

	return host
}

func (ts *testSuite) newNode(ctx context.Context, port int) testNode {
	h := ts.newHost(port)
	logger := ts.logger.With("host", h.ID()).With("port", port)

	service, err := New(ctx, h, logger)
	require.NoError(ts.t, err, "failed to create gossip service")

	logger.Info("Created gossip service")
	ts.t.Cleanup(service.Close)

	return testNode{
		host:    h,
		service: service,
		address: peer.AddrInfo{
			ID:    h.ID(),
			Addrs: h.Addrs(),
		},
	}
}

// connectAllNodes connects all nodes to each other
func connectAllNodes(t *testing.T, ctx context.Context, nodes []testNode) {
	// should be N*(N-1)/2 connections
	var connectionsCount int

	for _, us := range nodes {
		for _, them := range nodes {
			// skip self
			if us.address.ID == them.address.ID {
				continue
			}

			// skip if already connected (by counterparty)
			if len(us.host.Network().ConnsToPeer(them.address.ID)) > 0 {
				continue
			}

			err := us.host.Connect(ctx, them.address)
			require.NoError(t, err, "failed to connect nodes (%s -> %s)", us.address.ID, them.address.ID)
			connectionsCount++
		}
	}

	t.Logf("Total nodes: %d, Total connections: %d", len(nodes), connectionsCount)
}

// connectPartialGraph connects some nodes to other nodes in a partial graph
func connectPartialTopology(t *testing.T, ctx context.Context, nodes []testNode) {
	var connectionsCount int

	require.Equal(t, 10, len(nodes), "expected 10 nodes for connectPartialGraph")

	connect := func(ourIndex int, theirIndices ...int) {
		us := nodes[ourIndex]

		for _, theirIndex := range theirIndices {
			them := nodes[theirIndex]

			if len(us.host.Network().ConnsToPeer(them.address.ID)) > 0 {
				t.Logf("Already connected node #%d to #%d", ourIndex, theirIndex)
				continue
			}

			err := us.host.Connect(ctx, them.address)
			require.NoError(t, err, "failed to connect nodes (%s -> %s)", us.address.ID, them.address.ID)

			connectionsCount++
			t.Logf("Connected node #%d to #%d", ourIndex, theirIndex)
		}
	}

	// connect in a chain (2 connections per node)
	// but also connect each node a third/fourth peer on the opposite side of the chain
	connect(0, 1, 4, 8)
	connect(1, 2, 5)
	connect(2, 3, 6)
	connect(3, 4, 7, 9)
	connect(4, 5)
	connect(5, 6)
	connect(6, 7)
	connect(7, 8)
	connect(8, 9)
	connect(9, 0)

	t.Logf("Total nodes: %d, Total connections: %d", len(nodes), connectionsCount)
}

func printGossipStats(t *testing.T, messages []testGossipMessage) {
	var (
		peers      = make(map[peer.ID]struct{})
		general    = make([]time.Duration, 0, len(messages))
		byReceiver = make(map[peer.ID][]time.Duration)
		byOrigin   = make(map[peer.ID][]time.Duration)
	)

	for _, message := range messages {
		peers[message.origin] = struct{}{}
		peers[message.receiver] = struct{}{}

		general = append(general, message.receiveLatency)
		byReceiver[message.receiver] = append(byReceiver[message.receiver], message.receiveLatency)
		byOrigin[message.origin] = append(byOrigin[message.origin], message.receiveLatency)
	}

	utils.LogDurationStats(t, "General receive latency:", general)
	t.Log("\n")

	for peer := range peers {
		t.Logf("Peer: %s", peer)
		utils.LogDurationStats(t, "Receive latency:", byReceiver[peer])
		utils.LogDurationStats(t, "Send latency:", byOrigin[peer])
		t.Log("\n")
	}
}
