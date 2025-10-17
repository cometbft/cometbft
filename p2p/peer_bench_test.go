package p2p

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

type p2pUnidirectionalConfig struct {
	duration        time.Duration
	sendConcurrency int
	receiveDelay    time.Duration
}

// TestBenchP2PUnidirectional measures message throughput and latency by sending messages
// from one peer to another in a single direction.
func TestBenchP2PUnidirectional(t *testing.T) {
	utils.GuardP2PBenchTest(t)

	for _, tt := range []p2pUnidirectionalConfig{
		{
			duration:        10 * time.Second,
			sendConcurrency: 1,
			receiveDelay:    0,
		},
		{
			duration:        10 * time.Second,
			sendConcurrency: 8,
			receiveDelay:    0,
		},
		{
			duration:        10 * time.Second,
			sendConcurrency: 1,
			receiveDelay:    10 * time.Millisecond,
		},
		{
			duration:        10 * time.Second,
			sendConcurrency: 8,
			receiveDelay:    10 * time.Millisecond,
		},
		{
			duration:        10 * time.Second,
			sendConcurrency: 16,
			receiveDelay:    10 * time.Millisecond,
		},
	} {
		name := fmt.Sprintf("duration-%s.concurrency-%d.delay-%s", tt.duration, tt.sendConcurrency, tt.receiveDelay)
		t.Run(name, func(t *testing.T) {
			runtime.GC()
			testBenchP2PUnidirectional(t, tt)
			t.Log("")
		})
	}
}

func testBenchP2PUnidirectional(t *testing.T, cfg p2pUnidirectionalConfig) {
	t.Logf("Test duration: %s", cfg.duration.String())
	t.Logf("Send concurrency: %d goroutines", cfg.sendConcurrency)
	t.Logf("Work imitation delay: %s", cfg.receiveDelay.String())

	// ARRANGE
	const (
		benchChanFoo = byte(0xaa)
		bandwidth    = 100 * (1 << 20) // 100 MB
	)

	// Given P2P and MUX connection configs
	p2pCfg := config.DefaultP2PConfig()
	p2pCfg.AddrBookStrict = true

	// todo
	// p2pCfg.SendRate = bandwidth
	// p2pCfg.RecvRate = bandwidth

	muxConnConfig := conn.DefaultMConnConfig()

	// todo
	// muxConnConfig := conn.MConnConfig{
	// SendRate:                bandwidth,
	// RecvRate:                bandwidth,
	// MaxPacketMsgPayloadSize: bandwidth,
	// FlushThrottle:           10 * time.Millisecond,
	// PingInterval:            5 * time.Second,
	// PongTimeout:             3 * time.Second,
	// }

	logger := log.NewNopLogger()

	// Given recipient reactor ...
	recipientReactor := NewBenchReactor(t, benchChanFoo, cfg.receiveDelay)

	// With relevant switch ...
	recipientSwitch := createSwitchWithReactor(t, p2pCfg, muxConnConfig, "recipient", recipientReactor, logger)
	require.NoError(t, recipientSwitch.Start())
	defer recipientSwitch.Stop()

	recipientAddr := recipientSwitch.NetAddress()
	t.Logf("Recipient listening at: %s", recipientAddr.String())

	// Given sender reactor ...
	stubReactor := NewTestReactor(recipientReactor.GetChannels(), false)
	stubReactor.SetLogger(logger)

	senderSwitch := createSwitchWithReactor(t, p2pCfg, muxConnConfig, "sender", stubReactor, logger)
	require.NoError(t, senderSwitch.Start())
	defer senderSwitch.Stop()

	t.Logf("Sender listening at: %s", senderSwitch.NetAddress().String())

	// Connect sender to recipient
	require.NoError(t, senderSwitch.DialPeerWithAddress(recipientAddr))
	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 1, recipientSwitch.Peers().Size(), "Recipient should have 2 peers")
	require.Equal(t, 1, senderSwitch.Peers().Size(), "Sender should have 1 peer")

	senderPeers := senderSwitch.Peers().Copy()
	require.Len(t, senderPeers, 1)

	recipientPeer := senderPeers[0]

	ctx, cancel := context.WithTimeout(context.Background(), cfg.duration)
	defer cancel()

	t.Log("Sending messages...")

	var (
		start = time.Now()

		sendSuccesses = atomic.Uint64{}
		sendFailures  = atomic.Uint64{}

		receiveSuccesses = atomic.Uint64{}
		receiveFailures  = atomic.Uint64{}

		receiveLatencies = make([]time.Duration, 0, 10_000)
		processLatencies = make([]time.Duration, 0, 10_000)

		waitProcessing = make(chan struct{})
	)

	sendFunc := func() {
		nowStr := strconv.FormatInt(time.Now().UnixMicro(), 10)
		msg := types.ToRequestEcho(nowStr)

		sent := recipientPeer.Send(Envelope{
			ChannelID: benchChanFoo,
			Message:   msg,
		})

		if sent {
			sendSuccesses.Add(1)
		} else {
			sendFailures.Add(1)
		}
	}

	finish := func() {
		t.Logf("Finished. Waiting for processing to complete. Current sink size: %d", len(recipientReactor.sink))
		close(recipientReactor.sink)
		<-waitProcessing
	}

	go func() {
		for record := range recipientReactor.sink {
			// todo
			// if record.err != nil {
			// receiveFailures.Add(1)
			// continue
			// }

			receiveSuccesses.Add(1)

			req, ok := record.payload.(*types.RequestEcho)
			require.True(t, ok)

			msg := strings.TrimLeft(req.Message, "\n\x10")
			i64, err := strconv.ParseInt(msg, 10, 64)
			require.NoError(t, err, "invalid i64: %s", string(msg))

			sentAt := time.UnixMicro(i64)

			receiveLatencies = append(receiveLatencies, record.receivedAt.Sub(sentAt))
			processLatencies = append(processLatencies, record.processedAt.Sub(sentAt))
		}

		close(waitProcessing)
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			finish()
			break LOOP
		default:
			// send sync
			if cfg.sendConcurrency < 2 {
				sendFunc()
				continue
			}

			// send async
			wg := sync.WaitGroup{}
			wg.Add(cfg.sendConcurrency)
			for i := 0; i < cfg.sendConcurrency; i++ {
				go func() {
					defer wg.Done()
					sendFunc()
				}()
			}
			wg.Wait()
		}
	}

	<-waitProcessing

	// ASSERT
	timeTaken := time.Since(start)

	t.Logf("Sent messages: %d", sendSuccesses.Load()+sendFailures.Load())
	t.Logf("  success: %d, failure %d", sendSuccesses.Load(), sendFailures.Load())
	t.Logf("  send RPS: %.0f", float64(sendSuccesses.Load())/timeTaken.Seconds())

	t.Logf("Received messages: %d", receiveSuccesses.Load()+receiveFailures.Load())
	t.Logf("  success: %d, failure: %d", receiveSuccesses.Load(), receiveFailures.Load())
	t.Logf("  receive RPS: %.0f", float64(receiveSuccesses.Load())/timeTaken.Seconds())

	utils.LogDurationStats(t, "Receive latency:", receiveLatencies)
	utils.LogDurationStats(t, "Process latency:", processLatencies)
}

// createSwitchWithReactor creates a new switch with the given reactor using real TCP
func createSwitchWithReactor(
	t *testing.T,
	cfg *config.P2PConfig,
	muxConnConfig conn.MConnConfig,
	name string,
	reactor Reactor,
	logger log.Logger,
) *Switch {
	t.Helper()

	nodeKey := NodeKey{
		PrivKey: ed25519.GenPrivKey(),
	}

	ports := utils.GetFreePorts(t, 1)
	addrStr := fmt.Sprintf("127.0.0.1:%d", ports[0])

	nodeInfo := DefaultNodeInfo{
		Moniker:         name,
		ProtocolVersion: defaultProtocolVersion,
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      addrStr,
		Network:         "testing",
		Version:         "1.2.3",
		Channels:        []byte{},
	}

	addr, err := NewNetAddressString(IDAddressString(nodeKey.ID(), addrStr))
	require.NoError(t, err)

	// Create transport
	transport := NewMultiplexTransport(nodeInfo, nodeKey, muxConnConfig)

	// Listen on the address
	require.NoError(t, transport.Listen(*addr))

	// Update the transport's netAddr with the actual listener address
	// This is necessary because we're using port 0 to get a free port
	actualAddr := NewNetAddress(nodeKey.ID(), transport.listener.Addr())
	transport.netAddr = *actualAddr

	// Create switch
	sw := NewSwitch(cfg, transport)
	sw.SetLogger(logger.With("switch", name))
	sw.SetNodeKey(&nodeKey)

	// Add the reactor
	sw.AddReactor(name, reactor)
	// Update node info with channels
	for ch := range sw.reactorsByCh {
		nodeInfo.Channels = append(nodeInfo.Channels, ch)
	}

	transport.nodeInfo = nodeInfo

	sw.SetNodeInfo(nodeInfo)

	return sw
}

type BenchReactor struct {
	BaseReactor

	t *testing.T

	channelID    byte
	receiveDelay time.Duration
	sink         chan benchRecord
}

type benchRecord struct {
	payload     proto.Message
	receivedAt  time.Time
	processedAt time.Time
}

func NewBenchReactor(t *testing.T, channelID byte, receiveDelay time.Duration) *BenchReactor {
	r := &BenchReactor{
		t:            t,
		channelID:    channelID,
		receiveDelay: receiveDelay,
		sink:         make(chan benchRecord, 1_000_000),
	}

	r.BaseReactor = *NewBaseReactor("BenchReactor", r)
	r.SetLogger(log.NewNopLogger())

	return r
}

func (r *BenchReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{
		{ID: r.channelID, Priority: 1, MessageType: &types.RequestEcho{}},
	}
}

func (r *BenchReactor) Receive(e Envelope) {
	receivedAt := time.Now()

	// "perform some work"
	if r.receiveDelay > 0 {
		time.Sleep(r.receiveDelay)
	}

	processedAt := time.Now()

	// we're sending raw payload to exclude parsing latency
	r.sink <- benchRecord{
		payload:     e.Message,
		receivedAt:  receivedAt,
		processedAt: processedAt,
	}
}
