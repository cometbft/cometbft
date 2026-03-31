package lp2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
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
	"github.com/cometbft/cometbft/test/utils"
	"github.com/cometbft/cometbft/version"
	"github.com/stretchr/testify/require"
)

type perfBench struct {
	// type of the test
	TestType string

	// type of the network (p2p / lib-p2p)
	NetworkType string

	// imitation of some processing operations in the receiver side
	ProcessingDelay time.Duration

	// size of the message to send
	MessageSize int

	// number of goroutines to send messages
	SendConcurrency int

	// duration of the benchmark
	BenchDuration time.Duration
}

const (
	perfBenchDuration = 10 * time.Second

	perfBenchTypeOneWay    = "one-way"
	perfBenchTypeReqRes    = "request-response"
	perfBenchTypeBroadcast = "broadcast"

	perfBenchNetworkCometP2P = "comet-p2p"
	perfBenchNetworkLP2P     = "lp2p"

	testChannelID = byte(0xaa)
)

func generatePerfBenchmarkMatrix() []perfBench {
	testTypes := []string{perfBenchTypeOneWay, perfBenchTypeReqRes, perfBenchTypeBroadcast}

	networkTypes := []string{perfBenchNetworkCometP2P, perfBenchNetworkLP2P}

	processingDelays := []time.Duration{
		0,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
	}

	msgSizes := []int{
		1024 / 2,   // 512b
		1024 * 1,   // 1kb
		1024 * 8,   // 8kb
		1024 * 64,  // 64kb
		1024 * 512, // 512kb
	}

	sendRoutines := []int{
		1, 8, 16,
	}

	product := len(testTypes) * len(networkTypes) * len(processingDelays) * len(msgSizes) * len(sendRoutines)
	cases := make([]perfBench, 0, product)

	for _, testType := range testTypes {
		for _, networkType := range networkTypes {
			for _, procDelay := range processingDelays {
				for _, msgSize := range msgSizes {
					for _, sendConcurrency := range sendRoutines {
						cases = append(cases, perfBench{
							TestType:        testType,
							NetworkType:     networkType,
							ProcessingDelay: procDelay,
							MessageSize:     msgSize,
							SendConcurrency: sendConcurrency,
							BenchDuration:   perfBenchDuration,
						})
					}
				}
			}
		}
	}

	return cases
}

func (b *perfBench) String() string {
	msgSize := fmt.Sprintf("%dkb", b.MessageSize/1024)
	if b.MessageSize < 1024 {
		msgSize = fmt.Sprintf("%db", b.MessageSize)
	}

	delay := "no"
	if b.ProcessingDelay > 0 {
		delay = b.ProcessingDelay.String()
	}

	// e.g. one-way-comet-p2p.msg-1kb.cc-8.delay-10ms.dur-10s
	return fmt.Sprintf(
		"%s-%s.msg-%s.cc-%d.delay-%s.dur-%s",
		b.TestType,
		b.NetworkType,
		msgSize,
		b.SendConcurrency,
		delay,
		b.BenchDuration.String(),
	)
}

func (b *perfBench) ChannelDescriptor() *conn.ChannelDescriptor {
	return &conn.ChannelDescriptor{
		ID:       testChannelID,
		Priority: 1,

		// only for comet-p2p
		SendQueueCapacity: 100_000,

		// 4MB, only for comet-p2p, bigger than ALL existing buffers
		RecvBufferCapacity: 4 * (1 << 20),

		// payload + 1kb overhead
		RecvMessageCapacity: b.MessageSize + 1024,
		MessageType:         &types.RequestEcho{},
	}
}

func (b *perfBench) SamplePayload() []byte {
	payload := make([]byte, b.MessageSize)
	_, err := rand.Read(payload)
	if err != nil {
		panic(err)
	}

	return payload
}

// ConfigLP2PModifier "maxed out" config for lp2p
func (b *perfBench) ConfigLP2PModifier(c *config.LibP2PConfig) {
	c.Limits.Mode = config.LibP2PLimitsModeDisabled
	c.Scaler.MinWorkers = 64
	c.Scaler.MaxWorkers = 128
	c.Scaler.ThresholdLatency = 500 * time.Millisecond
}

// CometP2PConfig "maxed out" config for comet-p2p
func (b *perfBench) CometP2PConfig() *config.P2PConfig {
	p2pConfig := config.DefaultP2PConfig()
	// 4MB, bigger than all perf. msgs
	p2pConfig.MaxPacketMsgPayloadSize = 4 * (1 << 20) // 4MB

	// 1 GB/s
	p2pConfig.SendRate = 1 << 30
	p2pConfig.RecvRate = 1 << 30

	return p2pConfig
}

func TestBench(t *testing.T) {
	utils.GuardP2PBenchTest(t)

	matrix := generatePerfBenchmarkMatrix()

	t.Logf("Total test cases: %d", len(matrix))

	for _, tt := range matrix {
		t.Run(tt.String(), func(t *testing.T) {
			runtime.GC()
			testPerformanceBenchmark(t, tt)
		})
	}
}

func testPerformanceBenchmark(t *testing.T, config perfBench) {
	switch {
	case config.TestType == perfBenchTypeOneWay && config.NetworkType == perfBenchNetworkLP2P:
		benchSetupAndRunOneWayLP2P(t, config)
	case config.TestType == perfBenchTypeOneWay && config.NetworkType == perfBenchNetworkCometP2P:
		benchSetupAndRunOneWayCometP2P(t, config)
	default:
		t.Skip("test type not yet supported")
	}

	// todo lp2p x req-res
	// todo lp2p x broadcast
	// todo comet-p2p x req-res
	// todo comet-p2p x broadcast
}

func benchSetupAndRunOneWayLP2P(t *testing.T, cfg perfBench) {
	// ARRANGE
	// Given 2 hosts without limits
	ports := utils.GetFreePorts(t, 2)
	host1 := makeTestHost(t, ports[0], withModifiedConfig(cfg.ConfigLP2PModifier))
	host2 := makeTestHost(t, ports[1], withModifiedConfig(cfg.ConfigLP2PModifier), withBootstrapPeers([]config.LibP2PBootstrapPeer{
		{
			Host: fmt.Sprintf("127.0.0.1:%d", ports[0]),
			ID:   host1.ID().String(),
		},
	}))

	// Given 2 connected p2p switches
	switch1, _ := newPerfSwitchLP2P(t, host1, cfg)
	switch2, reactor2 := newPerfSwitchLP2P(t, host2, cfg)

	// connect & start two switches to each other
	connectSwitches(t, []*Switch{switch1, switch2})

	// given peer2 inside peer1
	peer2 := switch1.Peers().Get(peerIDToKey(host2.ID()))
	require.NotNil(t, peer2)

	benchRunOneWay(t, cfg, peer2, reactor2)
}

func benchSetupAndRunOneWayCometP2P(t *testing.T, cfg perfBench) {
	// ARRANGE
	// Given 2 connected comet-p2p switches
	switch1, _ := newPerfSwitchCometP2P(t, "sender", cfg)
	switch2, reactor2 := newPerfSwitchCometP2P(t, "recipient", cfg)

	// connect sender -> recipient
	require.NoError(t, switch1.DialPeerWithAddress(switch2.NetAddress()))
	time.Sleep(100 * time.Millisecond)

	// given recipient peer inside sender
	peer2 := switch1.Peers().Get(switch2.NodeInfo().ID())
	require.NotNil(t, peer2)

	benchRunOneWay(t, cfg, peer2, reactor2)
}

func benchRunOneWay(t *testing.T, cfg perfBench, peer2 p2p.Peer, reactor2 *PerfReactor) {
	// ARRANGE
	// Given a peer
	require.NotNil(t, peer2)

	// given some metrics for processing
	var (
		sendSuccess  = atomic.Uint64{}
		sendFailures = atomic.Uint64{}
		sendBytes    = atomic.Uint64{}

		receiveSuccess = atomic.Uint64{}
		receiveBytes   = atomic.Uint64{}

		receiveLatencies = make([]time.Duration, 0, 10_000)

		sink         = make(chan utils.PerfRecord, 1_000_000)
		onSinkClosed = make(chan struct{})
	)

	// Given a limited test duration
	ctx, cancel := context.WithTimeout(context.Background(), cfg.BenchDuration)
	defer cancel()

	// ACT
	start := time.Now()

	// 1. receive messages on recipient (async)
	reactor2.OnReceive(func(e p2p.Envelope) {
		// drop
		if ctx.Err() != nil {
			return
		}

		var record utils.PerfRecord
		err := record.FromEcho(e.Message.(*types.RequestEcho))
		if err != nil {
			// should not fail
			panic(err)
		}

		record.ReceivedAt = time.Now()

		if cfg.ProcessingDelay > 0 {
			time.Sleep(cfg.ProcessingDelay)
		}

		// send non-blocking to the sink
		sink <- record
	})

	// 2. run sink routine to receive messages from recipient
	go func() {
		for record := range sink {
			receiveSuccess.Add(1)
			receiveBytes.Add(uint64(len(record.Payload)))

			// no need for a mutex here, we're not sharing this slice
			sentAt := record.SentAt
			receiveLatencies = append(receiveLatencies, record.ReceivedAt.Sub(sentAt))
		}

		t.Logf("Finished receiver goroutine")
		close(onSinkClosed)
	}()

	// 3. send messages from sender->recipient based on concurrency
	samplePayload := cfg.SamplePayload()

	sendFunc := func() {
		perfRecord := utils.PerfRecord{
			SentAt:  time.Now(),
			Payload: samplePayload,
		}

		sent := peer2.Send(p2p.Envelope{
			ChannelID: testChannelID,
			Message:   perfRecord.AsEcho(),
		})

		if sent {
			sendSuccess.Add(1)
			sendBytes.Add(uint64(len(samplePayload)))
		} else {
			sendFailures.Add(1)
		}
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			t.Logf("Finished. Waiting for processing to complete. Current sink size: %d", len(sink))
			close(sink)
			break LOOP
		default:
			// send sync
			if cfg.SendConcurrency < 2 {
				sendFunc()
				continue
			}

			// send async
			wg := sync.WaitGroup{}
			wg.Add(cfg.SendConcurrency)
			for i := 0; i < cfg.SendConcurrency; i++ {
				go func() {
					defer wg.Done()
					sendFunc()
				}()
			}
			wg.Wait()
		}
	}

	// ASSERT
	// wait for the sink to be closed
	<-onSinkClosed

	// log perf stats
	utils.LogPerformanceStatsOneWay(
		t,
		start,
		sendSuccess.Load(), sendFailures.Load(), receiveSuccess.Load(),
		sendBytes.Load(), receiveBytes.Load(),
		receiveLatencies,
	)
}

func newPerfSwitchLP2P(t *testing.T, host *Host, cfg perfBench) (*Switch, *PerfReactor) {
	t.Helper()

	reactor := NewPerfReactor(t, cfg.ChannelDescriptor())
	sw, err := NewSwitch(
		nil,
		host,
		[]SwitchReactor{
			{Name: "PerfReactor", Reactor: reactor},
		},
		p2p.NopMetrics(),
		host.Logger(),
	)
	require.NoError(t, err)

	return sw, reactor
}

func newPerfSwitchCometP2P(t *testing.T, name string, cfg perfBench) (*p2p.Switch, *PerfReactor) {
	t.Helper()

	p2pConfig := cfg.CometP2PConfig()
	chanDesc := cfg.ChannelDescriptor()

	nodeKey := p2p.NodeKey{
		PrivKey: ed25519.GenPrivKey(),
	}

	ports := utils.GetFreePorts(t, 1)
	addrStr := fmt.Sprintf("127.0.0.1:%d", ports[0])

	nodeInfo := p2p.DefaultNodeInfo{
		Moniker:         name,
		ProtocolVersion: p2p.NewProtocolVersion(version.P2PProtocol, version.BlockProtocol, 0),
		DefaultNodeID:   nodeKey.ID(),
		ListenAddr:      addrStr,
		Network:         "testing",
		Version:         "1.0.0",
		Channels:        []byte{chanDesc.ID},
	}

	addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.ID(), addrStr))
	require.NoError(t, err)

	transport := p2p.NewMultiplexTransport(nodeInfo, nodeKey, p2p.MConnConfig(p2pConfig))
	require.NoError(t, transport.Listen(*addr))

	reactor := NewPerfReactor(t, chanDesc)
	reactor.SetLogger(log.NewNopLogger().With("reactor", "PerfReactor", "switch", name))

	sw := p2p.NewSwitch(p2pConfig, transport)
	sw.SetLogger(log.NewNopLogger().With("switch", name))
	sw.SetNodeKey(&nodeKey)
	sw.AddReactor("PerfReactor", reactor)
	sw.SetNodeInfo(nodeInfo)

	require.NoError(t, sw.Start())
	t.Cleanup(func() { _ = sw.Stop() })

	return sw, reactor
}

// PerfReactor is a reactor that measures various performance metrics.
type PerfReactor struct {
	p2p.BaseReactor

	t         *testing.T
	desc      *conn.ChannelDescriptor
	onReceive func(p2p.Envelope)
}

var _ p2p.Reactor = &PerfReactor{}

func NewPerfReactor(t *testing.T, desc *conn.ChannelDescriptor) *PerfReactor {
	r := &PerfReactor{
		t:    t,
		desc: desc,
	}
	r.BaseReactor = *p2p.NewBaseReactor("PerfReactor", r)

	return r
}

func (p *PerfReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{p.desc}
}

func (p *PerfReactor) OnReceive(handler func(p2p.Envelope)) { p.onReceive = handler }

func (p *PerfReactor) Receive(e p2p.Envelope) {
	if p.onReceive == nil {
		p.t.Fatalf("onReceive is not set")
	}

	p.onReceive(e)
}
