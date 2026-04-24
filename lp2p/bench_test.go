package lp2p

import (
	"context"
	"crypto/rand"
	"encoding/json"
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
	TestType string `json:"testType"`

	// type of the network (p2p / lib-p2p)
	NetworkType string `json:"networkType"`

	// imitation of some processing operations in the receiver side
	ProcessingDelay time.Duration `json:"processingDelay"`

	// size of the message to send
	MessageSize int `json:"messageSize"`

	// number of goroutines to send messages
	SendConcurrency int `json:"sendConcurrency"`

	// timeframe tests: duration of the benchmark
	BenchDuration time.Duration `json:"benchDuration"`

	// drain tests: total bandwidth of the benchmark
	BenchBandwidthBytes int `json:"benchBandwidthBytes"`

	// broadcast tests: number of peers in the network
	BroadcastPeers int `json:"broadcastPeers"`
}

const (
	perfBenchNetworkCometP2P = "comet-p2p"
	perfBenchNetworkLP2P     = "lp2p"

	// send messages from peer1 to peer2 for a given duration
	perfBenchTypeSendTimeframe = "send-timeframe"

	// send X messages from peer1 to peer2 and wait for peer2 to receive them all
	perfBenchTypeSendDrain = "send-drain"

	// todo: send X messages from peer1 to [peer2, peer3, peer4] and wait for peers to receive them all
	perfBenchTypeBroadcast = "broadcast-drain"

	// for timeframe tests: send messages for a given duration
	// and then cut.
	perfBenchTimeframe = 10 * time.Second

	// for drain tests: divide 1GB by msgs size to get the number of messages to send,
	// eg: for 64kb msgs: 1GB / 64kb = 15625 messages
	perfBenchDrainBandwidthBytes = 1 << 30

	perfBenchBroadcastPeers = 8

	testChannelID = byte(0xaa)
)

func generatePerfBenchmarkMatrix() []perfBench {
	testTypes := []string{perfBenchTypeSendTimeframe, perfBenchTypeSendDrain, perfBenchTypeBroadcast}

	networkTypes := []string{perfBenchNetworkCometP2P, perfBenchNetworkLP2P}

	processingDelays := []time.Duration{
		0,
		20 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
	}

	msgSizes := []int{
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
						perfBench := perfBench{
							TestType:        testType,
							NetworkType:     networkType,
							ProcessingDelay: procDelay,
							MessageSize:     msgSize,
							SendConcurrency: sendConcurrency,

							BenchDuration:       perfBenchTimeframe,
							BenchBandwidthBytes: perfBenchDrainBandwidthBytes,
							BroadcastPeers:      perfBenchBroadcastPeers,
						}

						// broadcast has no concurrency
						if testType == perfBenchTypeBroadcast && sendConcurrency > 1 {
							continue
						}

						cases = append(cases, perfBench)
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

	var suffix string
	switch b.TestType {
	case perfBenchTypeSendTimeframe:
		suffix = fmt.Sprintf("dur-%s", b.BenchDuration.String())
	case perfBenchTypeSendDrain:
		suffix = fmt.Sprintf("bw-%dmb", b.BenchBandwidthBytes/(1024*1024))
	case perfBenchTypeBroadcast:
		suffix = fmt.Sprintf("peers-%d", b.BroadcastPeers)
	}

	// e.g. one-way-comet-p2p.msg-1kb.cc-8.delay-10ms.dur-10s
	return fmt.Sprintf(
		"%s-%s.msg-%s.cc-%d.delay-%s.%s",
		b.TestType,
		b.NetworkType,
		msgSize,
		b.SendConcurrency,
		delay,
		suffix,
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
	p2pConfig.AllowDuplicateIP = true

	// 4MB, bigger than all perf. msgs
	p2pConfig.MaxPacketMsgPayloadSize = 4 * (1 << 20) // 4MB

	// 1 GB/s
	p2pConfig.SendRate = 1 << 30
	p2pConfig.RecvRate = 1 << 30

	return p2pConfig
}

// TestBench runs the performance benchmark for all test cases.
// `P2P_BENCH_TEST=matrix-only go test -run TestBench ./lp2p/... -v` - list available test cases
// `P2P_BENCH_TEST=1 go test TestBench/send-drain-lp2p.msg-64kb.cc-1.delay-20ms.dur-10` - run a specific test case
//
// @see scripts/bench/p2p_bench.py for a complete suite.
func TestBench(t *testing.T) {
	utils.GuardP2PBenchTest(t)

	matrix := generatePerfBenchmarkMatrix()

	if utils.PrintMatrixOnly() {
		for _, testCase := range matrix {
			fmt.Println("testCase:", testCase.String())
		}
		return
	}

	for _, testCase := range matrix {
		t.Run(testCase.String(), func(t *testing.T) {
			runtime.GC()
			t.Logf("[INPUT]: %s", jsonLine(t, testCase))
			out := testPerformanceBenchmark(t, testCase)

			t.Logf("[OUTPUT]: %s", jsonLine(t, out))
		})
	}
}

func testPerformanceBenchmark(t *testing.T, config perfBench) (output any) {
	switch config.NetworkType {
	case perfBenchNetworkLP2P:
		switch config.TestType {
		case perfBenchTypeSendTimeframe:
			// type1: lp2p x send-timeframe
			peer2, reactor2 := benchSetupSendLP2P(t, config)
			return benchRunSendTimeframe(t, config, peer2, reactor2)
		case perfBenchTypeSendDrain:
			// type2: lp2p x send-drain
			peer2, reactor2 := benchSetupSendLP2P(t, config)
			return benchRunSendDrain(t, config, peer2, reactor2)
		case perfBenchTypeBroadcast:
			// type3: lp2p x broadcast
			switches, reactors := benchSetupBroadcastLP2P(t, config)
			return benchRunBroadcast(t, config, switches, reactors)
		default:
			t.Fatalf("Invalid test config: %s", config.String())
			return nil
		}
	case perfBenchNetworkCometP2P:
		switch config.TestType {
		case perfBenchTypeSendTimeframe:
			// type4: comet-p2p x send-timeframe
			peer2, reactor2 := benchSetupSendCometP2P(t, config)
			return benchRunSendTimeframe(t, config, peer2, reactor2)
		case perfBenchTypeSendDrain:
			// type5: comet-p2p x send-drain
			peer2, reactor2 := benchSetupSendCometP2P(t, config)
			return benchRunSendDrain(t, config, peer2, reactor2)
		case perfBenchTypeBroadcast:
			// type6: comet-p2p x broadcast
			switches, reactors := benchSetupBroadcastCometP2P(t, config)
			return benchRunBroadcast(t, config, switches, reactors)
		default:
			t.Fatalf("Invalid test config: %s", config.String())
			return nil
		}
	default:
		t.Fatalf("Invalid test config: %+v", config)
		return nil
	}
}

func benchSetupSendLP2P(t *testing.T, cfg perfBench) (p2p.Peer, *perfReactor) {
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

	return peer2, reactor2
}

func benchSetupSendCometP2P(t *testing.T, cfg perfBench) (p2p.Peer, *perfReactor) {
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

	return peer2, reactor2
}

func benchSetupBroadcastLP2P(t *testing.T, cfg perfBench) ([]p2p.Switcher, []*perfReactor) {
	// ARRANGE
	// Given N hosts with switches and reactors
	var (
		hosts    = makeTestHosts(t, cfg.BroadcastPeers, withModifiedConfig(cfg.ConfigLP2PModifier))
		switches = make([]*Switch, cfg.BroadcastPeers)
		reactors = make([]*perfReactor, cfg.BroadcastPeers)
	)

	for i, host := range hosts {
		switches[i], reactors[i] = newPerfSwitchLP2P(t, host, cfg)
	}

	// connect switches to each other
	connectSwitches(t, switches)

	typed := make([]p2p.Switcher, 0, cfg.BroadcastPeers)
	for _, sw := range switches {
		typed = append(typed, sw)
	}

	return typed, reactors
}

func benchSetupBroadcastCometP2P(t *testing.T, cfg perfBench) ([]p2p.Switcher, []*perfReactor) {
	var (
		switches = make([]*p2p.Switch, cfg.BroadcastPeers)
		reactors = make([]*perfReactor, cfg.BroadcastPeers)
	)

	for i := 0; i < cfg.BroadcastPeers; i++ {
		switches[i], reactors[i] = newPerfSwitchCometP2P(t, fmt.Sprintf("peer-%d", i+1), cfg)
		if i > 0 {
			errDial := switches[0].DialPeerWithAddress(switches[i].NetAddress())
			require.NoError(t, errDial, "failed to dial peer #%d", i+1)
		}
	}

	// wait for switches to be connected
	time.Sleep(100 * time.Millisecond)

	typed := make([]p2p.Switcher, 0, cfg.BroadcastPeers)
	for _, sw := range switches {
		typed = append(typed, sw)
	}

	return typed, reactors
}

// send messages from peer1 to peer2 for cfg.BenchDuration.
// Doesn't wait for ALL messages to be received, just cut after the deadline
func benchRunSendTimeframe(t *testing.T, cfg perfBench, peer2 p2p.Peer, reactor2 *perfReactor) any {
	// ARRANGE
	// Given node1.peer2
	require.NotNil(t, peer2)

	// Given some metrics for processing
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

		// send non-blocking to the sink
		sink <- record

		if cfg.ProcessingDelay > 0 {
			time.Sleep(cfg.ProcessingDelay)
		}
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
	start := time.Now()

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
	time.Sleep(100 * time.Millisecond)
	return utils.LogPerformanceStatsSend(
		t,
		start,
		sendSuccess.Load(), sendFailures.Load(), receiveSuccess.Load(),
		sendBytes.Load(), receiveBytes.Load(),
		receiveLatencies,
	)
}

// send messages N from peer1 to peer2 and wait for peer2 to receive them.
// N is determined as cfg.BenchBandwidthBytes / cfg.MessageSize
func benchRunSendDrain(t *testing.T, cfg perfBench, peer2 p2p.Peer, reactor2 *perfReactor) any {
	// ARRANGE
	// Given node1.peer2
	require.NotNil(t, peer2)

	// Given number of messages to send
	numMessages := cfg.BenchBandwidthBytes / cfg.MessageSize

	t.Logf("Sending %d messages", numMessages)

	// Given some metrics for processing
	var (
		sendSuccess  = atomic.Uint64{}
		sendFailures = atomic.Uint64{}
		sendBytes    = atomic.Uint64{}

		receiveSuccess = atomic.Uint64{}
		receiveBytes   = atomic.Uint64{}

		receiveLatencies = make([]time.Duration, 0, 10_000)

		sink = make(chan utils.PerfRecord, 1_000_000)
	)

	// Given a timeout deadline for the benchmark
	ctx, cancel := context.WithTimeout(context.Background(), 3*cfg.BenchDuration)
	defer cancel()

	// ACT
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

		// send non-blocking to the sink
		sink <- record

		if cfg.ProcessingDelay > 0 {
			time.Sleep(cfg.ProcessingDelay)
		}
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
	start := time.Now()

	var sendIterations int
	for sendIterations < numMessages {
		if ctx.Err() != nil {
			t.Logf("Timed out. Current sink size: %d", len(sink))
			break
		}

		// single-threaded send
		if cfg.SendConcurrency == 1 {
			sendFunc()
			sendIterations++
			continue
		}

		// multi-threaded send
		concurrency := cfg.SendConcurrency
		if sendIterations+concurrency > numMessages {
			concurrency = numMessages - sendIterations
		}

		wg := sync.WaitGroup{}
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				sendFunc()
			}()
		}

		wg.Wait()
		sendIterations += concurrency
	}

	t.Logf(
		"Messages sent (%d, elapsed %s). Waiting for processing to complete. Current sink size: %d",
		sendIterations,
		time.Since(start).String(),
		len(sink),
	)

	// ASSERT
	// Wait for processing to complete or ctx.timeout or some idle state
	const maxIdleWait = 2 * time.Second
	completed := utils.WaitForProcessing(t, ctx, "send", &sendSuccess, &receiveSuccess, maxIdleWait)
	cancel()
	close(sink)

	if !completed {
		t.Logf("Processing not completed. Expected: %d, Actual: %d", sendSuccess.Load(), receiveSuccess.Load())
	}

	// log perf stats
	time.Sleep(100 * time.Millisecond)
	return utils.LogPerformanceStatsSend(
		t,
		start,
		sendSuccess.Load(), sendFailures.Load(), receiveSuccess.Load(),
		sendBytes.Load(), receiveBytes.Load(),
		receiveLatencies,
	)
}

func benchRunBroadcast(t *testing.T, cfg perfBench, switches []p2p.Switcher, reactors []*perfReactor) any {
	// ARRANGE
	// Given sender and receivers
	sender := switches[0]
	switches = switches[1:]
	reactors = reactors[1:]

	require.Equal(t, sender.Peers().Size(), len(switches))

	// Given N of messages that is cfg / message size / N of peers
	numPeers := len(switches)
	numMessages := cfg.BenchBandwidthBytes / cfg.MessageSize / numPeers
	t.Logf("Sending %d messages to %d receivers", numMessages, numPeers)

	// Given some metrics for processing
	var (
		receiveSuccess   = make([]atomic.Uint64, numPeers)
		receiveBytes     = make([]atomic.Uint64, numPeers)
		receiveLatencies = make([][]time.Duration, numPeers)
		timeTaken        = make([]time.Duration, numPeers)
		sinks            = make([]chan utils.PerfRecord, numPeers)
	)

	for i := range numPeers {
		receiveLatencies[i] = make([]time.Duration, 0, 10_000)
		sinks[i] = make(chan utils.PerfRecord, 1_000_000)
	}

	// Given a timeout deadline for the benchmark
	ctx, cancel := context.WithTimeout(context.Background(), 3*cfg.BenchDuration)
	defer cancel()

	// ACT
	for idx, reactor := range reactors {
		sink := sinks[idx]

		// 1. receive messages on recipient (async)
		reactor.OnReceive(func(e p2p.Envelope) {
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

			// send non-blocking to the sink
			sink <- record

			if cfg.ProcessingDelay > 0 {
				time.Sleep(cfg.ProcessingDelay)
			}
		})

		// 2. run sink routine to receive messages from recipient
		go func() {
			for record := range sink {
				receiveSuccess[idx].Add(1)
				receiveBytes[idx].Add(uint64(len(record.Payload)))

				// no need for a mutex here, we're not sharing this slice
				sentAt := record.SentAt
				receiveLatencies[idx] = append(receiveLatencies[idx], record.ReceivedAt.Sub(sentAt))
			}

			t.Logf("Finished receiver goroutine (%d)", idx)
		}()
	}

	// 3. send messages from sender->receivers
	samplePayload := cfg.SamplePayload()
	start := time.Now()

	for i := 0; i < numMessages; i++ {
		if ctx.Err() != nil {
			t.Logf("Broadcast: Timed out")
			break
		}

		perfRecord := utils.PerfRecord{
			SentAt:  time.Now(),
			Payload: samplePayload,
		}

		sender.TryBroadcast(p2p.Envelope{
			ChannelID: testChannelID,
			Message:   perfRecord.AsEcho(),
		})
	}

	// ASSERT
	// Wait for processing to complete or ctx.timeout or some idle state
	var wg sync.WaitGroup
	var expected atomic.Uint64

	wg.Add(numPeers)
	expected.Store(uint64(numMessages))

	for idx := range numPeers {
		name := fmt.Sprintf("peer-%d", idx+1)

		go func(idx int) {
			defer wg.Done()

			const maxIdleWait = 5 * time.Second

			if !utils.WaitForProcessing(t, ctx, name, &expected, &receiveSuccess[idx], maxIdleWait) {
				actual := receiveSuccess[idx].Load()
				t.Logf("%s: Processing not completed. Expected: %d, Actual: %d", name, numMessages, actual)
			}

			timeTaken[idx] = time.Since(start)
			close(sinks[idx])
		}(idx)
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)
	return utils.LogPerformanceStatsBroadcast(
		t,
		start,
		numMessages,
		receiveSuccess,
		receiveBytes,
		receiveLatencies,
		timeTaken,
	)
}

func newPerfSwitchLP2P(t *testing.T, host *Host, cfg perfBench) (*Switch, *perfReactor) {
	t.Helper()

	reactor := newPerfReactor(t, cfg.ChannelDescriptor())
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

func newPerfSwitchCometP2P(t *testing.T, name string, cfg perfBench) (*p2p.Switch, *perfReactor) {
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

	reactor := newPerfReactor(t, chanDesc)
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

// perfReactor is a reactor that measures various performance metrics.
type perfReactor struct {
	p2p.BaseReactor

	t         *testing.T
	desc      *conn.ChannelDescriptor
	onReceive func(p2p.Envelope)
}

var _ p2p.Reactor = &perfReactor{}

func newPerfReactor(t *testing.T, desc *conn.ChannelDescriptor) *perfReactor {
	r := &perfReactor{
		t:    t,
		desc: desc,
	}
	r.BaseReactor = *p2p.NewBaseReactor("PerfReactor", r)

	return r
}

func (p *perfReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{p.desc}
}

func (p *perfReactor) OnReceive(handler func(p2p.Envelope)) { p.onReceive = handler }

func (p *perfReactor) Receive(e p2p.Envelope) {
	if p.onReceive == nil {
		p.t.Fatalf("onReceive is not set")
	}

	p.onReceive(e)
}

func jsonLine(t *testing.T, obj any) string {
	t.Helper()

	if obj == nil {
		return "{}"
	}

	json, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("Failed to marshal object: %v", err)
	}

	return string(json)
}
