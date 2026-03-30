package lp2p

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/cometbft/cometbft/test/utils"
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
	if config.NetworkType == perfBenchNetworkCometP2P {
		t.Skip("network type not yet supported")
		return
	}

	if config.TestType != perfBenchTypeOneWay {
		t.Skip("test type not yet supported")
		return
	}

	// todo
	testPerformanceBenchmarkOneWay(t, config)
}

func testPerformanceBenchmarkOneWay(t *testing.T, cfg perfBench) {
	// ARRANGE
	// Given a sample payload
	samplePayload := make([]byte, cfg.MessageSize)
	_, err := rand.Read(samplePayload)
	require.NoError(t, err)

	// Given 2 hosts without limits
	ports := utils.GetFreePorts(t, 2)

	modifyConfig := func(c *config.LibP2PConfig) {
		c.Limits.Mode = config.LibP2PLimitsModeDisabled
		c.Scaler.MinWorkers = 32
		c.Scaler.MaxWorkers = 128
		c.Scaler.ThresholdLatency = 500 * time.Millisecond
	}

	host1 := makeTestHost(t, ports[0], withModifiedConfig(modifyConfig))
	host2 := makeTestHost(t, ports[1], withModifiedConfig(modifyConfig), withBootstrapPeers([]config.LibP2PBootstrapPeer{
		{
			Host: fmt.Sprintf("127.0.0.1:%d", ports[0]),
			ID:   host1.ID().String(),
		},
	}))

	switchMaker := func(host *Host) (*Switch, *perfReactor) {
		reactor := newPerfReactor(t, cfg, testChannelID)
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

	switch1, _ := switchMaker(host1)
	switch2, reactor2 := switchMaker(host2)

	// connect & start two switches to each other
	connectSwitches(t, []*Switch{switch1, switch2})

	// given peer2 inside peer1
	peer2 := switch1.Peers().Get(peerIDToKey(host2.ID()))
	require.NotNil(t, peer2)

	// given some metrics for processing
	var (
		sendSuccess  = atomic.Uint64{}
		sendFailures = atomic.Uint64{}
		sendBytes    = atomic.Uint64{}

		receiveSuccess = atomic.Uint64{}
		receiveBytes   = atomic.Uint64{}

		receiveLatencies = make([]time.Duration, 0, 10_000)
		processLatencies = make([]time.Duration, 0, 10_000)

		sink         = make(chan perfRecord, 1_000_000)
		onSinkClosed = make(chan struct{})
	)

	// given receiver function that decodes and records latency samples
	reactor2.onReceive = func(e p2p.Envelope) {
		var record perfRecord
		err := record.FromEcho(e.Message.(*types.RequestEcho))
		if err != nil {
			// should not fail
			panic(err)
		}

		record.ReceivedAt = time.Now()

		if cfg.ProcessingDelay > 0 {
			time.Sleep(cfg.ProcessingDelay)
		}

		record.ProcessedAt = time.Now()

		// send non-blocking to the sink
		sink <- record
	}

	// Given a limited test duration
	ctx, cancel := context.WithTimeout(context.Background(), cfg.BenchDuration)
	defer cancel()

	// ACT
	start := time.Now()

	// 1. receive messages on host2 (async)
	reactor2.onReceive = func(e p2p.Envelope) {
		// drop
		if ctx.Err() != nil {
			return
		}

		var record perfRecord
		err := record.FromEcho(e.Message.(*types.RequestEcho))
		if err != nil {
			// should not fail
			panic(err)
		}

		record.ReceivedAt = time.Now()

		if cfg.ProcessingDelay > 0 {
			time.Sleep(cfg.ProcessingDelay)
		}

		record.ProcessedAt = time.Now()

		// send non-blocking to the sink
		sink <- record
	}

	// 2. run sink routine to receive messages from host2
	go func() {
		for record := range sink {
			receiveSuccess.Add(1)
			receiveBytes.Add(uint64(len(record.Payload)))

			// no need for a mutex here, we're not sharing this slice
			sentAt := record.SentAt
			receiveLatencies = append(receiveLatencies, record.ReceivedAt.Sub(sentAt))
			processLatencies = append(processLatencies, record.ProcessedAt.Sub(sentAt))
		}

		t.Logf("Finished receiver goroutine")
		close(onSinkClosed)
	}()

	// 3. send messages from host1->host2 based on concurrency
	sendFunc := func() {
		perfRecord := perfRecord{
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

	var (
		timeTaken = time.Since(start)

		sendOK     = sendSuccess.Load()
		sendFailed = sendFailures.Load()

		receivedOK = receiveSuccess.Load()

		// if sendFailed is low, then this diff indicates that messages are QUEUED in the priority queue
		// and NOT lost. Since we're benchmarking a concrete time frame, we don't wait for
		// all messages to be processed, so they'll lower the "throughput" value.
		inFlight           = sendOK - receivedOK
		inFlightPercentage = 100 * float64(inFlight) / float64(sendOK+sendFailed)
	)

	t.Logf("Sent messages: %d", sendOK+sendFailed)
	t.Logf("  success: %d, failure %d", sendOK, sendFailed)
	t.Logf("  send RPS: %.0f", float64(sendOK)/timeTaken.Seconds())

	t.Logf("Received messages: %d", receivedOK)
	t.Logf("  success: %d, in-flight: %d", receivedOK, inFlight)
	t.Logf("  receive RPS: %.0f", float64(receivedOK)/timeTaken.Seconds())
	t.Logf("  still in-flight: %d (%.3f%%)", int64(inFlight), inFlightPercentage)

	utils.LogBytesThroughputStats(t, "Send throughput:", sendBytes.Load(), timeTaken)
	utils.LogBytesThroughputStats(t, "Receive throughput:", receiveBytes.Load(), timeTaken)

	utils.LogDurationStats(t, "Receive latency:", receiveLatencies)
	utils.LogDurationStats(t, "Process latency:", processLatencies)
}

type perfReactor struct {
	p2p.BaseReactor

	t         *testing.T
	cfg       perfBench
	channelID byte
	onReceive func(p2p.Envelope)
}

var _ p2p.Reactor = &perfReactor{}

func newPerfReactor(t *testing.T, cfg perfBench, channelID byte) *perfReactor {
	r := &perfReactor{
		t:         t,
		cfg:       cfg,
		channelID: channelID,
	}
	r.BaseReactor = *p2p.NewBaseReactor("perfReactorLP2P", r)
	r.SetLogger(log.NewNopLogger())

	return r
}

func (p *perfReactor) GetChannels() []*conn.ChannelDescriptor {
	return []*conn.ChannelDescriptor{
		{
			ID:                  p.channelID,
			Priority:            1,
			SendQueueCapacity:   100_000,                  // only for comet-p2p
			RecvBufferCapacity:  4 * (1 << 20),            // 4MB, only for comet-p2p, bigger than ALL existing buffers
			RecvMessageCapacity: p.cfg.MessageSize + 1024, // payload + 1kb overhead
			MessageType:         &types.RequestEcho{},
		},
	}
}

func (p *perfReactor) OnReceive(handler func(p2p.Envelope)) { p.onReceive = handler }

func (p *perfReactor) Receive(e p2p.Envelope) {
	if p.onReceive == nil {
		p.t.Fatalf("onReceive is not set")
	}

	p.onReceive(e)
}

type perfRecord struct {
	Payload     []byte
	SentAt      time.Time
	ReceivedAt  time.Time
	ProcessedAt time.Time
}

func (r *perfRecord) AsEcho() *types.RequestEcho {
	msg := make([]byte, 8+len(r.Payload))
	binary.BigEndian.PutUint64(msg[:8], uint64(r.SentAt.UnixMicro()))
	copy(msg[8:], r.Payload)

	return &types.RequestEcho{Message: string(msg)}
}

func (r *perfRecord) FromEcho(echo *types.RequestEcho) error {
	raw := []byte(echo.Message)
	if len(raw) < 8 {
		return fmt.Errorf("invalid perf record: got %d bytes", len(raw))
	}

	tsMicros := int64(binary.BigEndian.Uint64(raw[:8]))
	r.SentAt = time.UnixMicro(tsMicros)
	r.Payload = append(make([]byte, 0, len(raw)-8), raw[8:]...)

	return nil
}
