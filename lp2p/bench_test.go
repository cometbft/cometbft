package lp2p

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/require"
)

const envBench = "LP2P_BENCH_TEST"

type lp2pUnidirectionalConfig struct {
	duration        time.Duration
	sendConcurrency int
	receiveDelay    time.Duration
}

// TestBenchLP2PUnidirectional this test measures messages libp2p throughput and latency by sending messages
// from peer A to peer B in a single direction. The payload only contains a timestamp,
// so we can measure e2e latency.
func TestBenchLP2PUnidirectional(t *testing.T) {
	guardBench(t)

	for _, tt := range []lp2pUnidirectionalConfig{
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
			testBenchLP2PUnidirectional(t, tt)
			t.Log("")
		})
	}
}

func testBenchLP2PUnidirectional(t *testing.T, cfg lp2pUnidirectionalConfig) {
	const channelFoo = byte(0xaa)

	t.Logf("Test duration: %s", cfg.duration.String())
	t.Logf("Send concurrency: %d goroutines", cfg.sendConcurrency)
	t.Logf("Work imitation delay: %s", cfg.receiveDelay.String())

	// ARRANGE
	ctx := context.Background()

	// Given 2 hosts
	ports := utils.GetFreePorts(t, 2)

	host1 := makeTestHost(t, ports[0], AddressBookConfig{}, false)
	host2 := makeTestHost(t, ports[1], AddressBookConfig{
		Peers: []PeerConfig{
			{
				Host: fmt.Sprintf("127.0.0.1:%d", ports[0]),
				ID:   host1.ID().String(),
			},
		},
	}, false)

	ConnectPeers(ctx, host2, host2.ConfigPeers())
	t.Cleanup(func() {
		host2.Close()
		host1.Close()
	})

	type record struct {
		payload     []byte
		err         error
		receivedAt  time.Time
		processedAt time.Time
	}

	// Given sink for records
	sink := make(chan record, 1_000_000)

	// Given host1 stream handler

	// Given host1 as peer inside host2
	host2peer1, err := NewPeer(host2, host1.AddrInfo(), p2p.NopMetrics())
	require.NoError(t, err)
	require.NoError(t, host2peer1.Start())

	// ACT
	// Send messages from host2 to host1
	ctx, cancel := context.WithTimeout(ctx, cfg.duration)
	defer cancel()

	var (
		sendSuccesses = atomic.Uint64{}
		sendFailures  = atomic.Uint64{}

		receiveSuccesses = atomic.Uint64{}
		receiveFailures  = atomic.Uint64{}

		receiveLatencies = make([]time.Duration, 0, 10_000)
		processLatencies = make([]time.Duration, 0, 10_000)

		waitProcessing = make(chan struct{})
	)

	protocolID := ProtocolID(channelFoo)
	host1.SetStreamHandler(protocolID, func(stream network.Stream) {
		defer func() {
			if r := recover(); r != nil {
				if strings.Contains(fmt.Sprintf("%+v", r), "closed chan") {
					// might be expected
					return
				}

				panic(r)
			}
		}()

		payload, err := StreamReadClose(stream)

		receivedAt := time.Now()

		// "perform some work"
		if err == nil && cfg.receiveDelay > 0 {
			time.Sleep(cfg.receiveDelay)
		}

		processedAt := time.Now()

		// we're sending raw payload to exclude parsing latency
		sink <- record{
			payload:     payload,
			err:         err,
			receivedAt:  receivedAt,
			processedAt: processedAt,
		}
	})

	// Run a routine to receive messages from host1
	go func() {
		for record := range sink {
			if record.err != nil {
				receiveFailures.Add(1)
				continue
			}

			receiveSuccesses.Add(1)

			msg := &types.Request{}
			require.NoError(t, msg.Unmarshal(record.payload))
			require.NotNil(t, msg.GetEcho())

			i64, err := strconv.ParseInt(msg.GetEcho().GetMessage(), 10, 64)
			require.NoError(t, err)

			sentAt := time.UnixMicro(i64)

			receiveLatencies = append(receiveLatencies, record.receivedAt.Sub(sentAt))
			processLatencies = append(processLatencies, record.processedAt.Sub(sentAt))
		}

		close(waitProcessing)
	}()

	sendFunc := func() {
		nowStr := strconv.FormatInt(time.Now().UnixMicro(), 10)
		msg := types.ToRequestEcho(nowStr)

		sent := host2peer1.Send(p2p.Envelope{
			ChannelID: channelFoo,
			Message:   msg,
		})

		if sent {
			sendSuccesses.Add(1)
		} else {
			sendFailures.Add(1)
		}
	}

	finish := func() {
		t.Logf("Finished. Waiting for processing to complete. Current sink size: %d", len(sink))
		close(sink)
		<-waitProcessing
	}

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
	t.Logf("Sent messages: %d", sendSuccesses.Load()+sendFailures.Load())
	t.Logf("  success: %d, failure %d", sendSuccesses.Load(), sendFailures.Load())
	t.Logf("  send RPS: %.0f", float64(sendSuccesses.Load())/cfg.duration.Seconds())

	t.Logf("Received messages: %d", receiveSuccesses.Load()+receiveFailures.Load())
	t.Logf("  success: %d, failure: %d", receiveSuccesses.Load(), receiveFailures.Load())
	t.Logf("  receive RPS: %.0f", float64(receiveSuccesses.Load())/cfg.duration.Seconds())

	require.NotEmpty(t, receiveLatencies)
	sort.Slice(receiveLatencies, func(i, j int) bool {
		return receiveLatencies[i] < receiveLatencies[j]
	})

	t.Log("Receive latency:")
	t.Logf(
		"  min: %s, p50: %s, p90: %s, p95: %s, p99: %s, max: %s",
		receiveLatencies[0].String(),
		percentile(receiveLatencies, 50).String(),
		percentile(receiveLatencies, 90).String(),
		percentile(receiveLatencies, 95).String(),
		percentile(receiveLatencies, 99).String(),
		receiveLatencies[len(receiveLatencies)-1].String(),
	)

	require.NotEmpty(t, processLatencies)
	sort.Slice(processLatencies, func(i, j int) bool {
		return processLatencies[i] < processLatencies[j]
	})

	t.Log("Process latency:")
	t.Logf(
		"  min: %s, p50: %s, p90: %s, p95: %s, p99: %s, max: %s",
		processLatencies[0].String(),
		percentile(processLatencies, 50).String(),
		percentile(processLatencies, 90).String(),
		percentile(processLatencies, 95).String(),
		percentile(processLatencies, 99).String(),
		processLatencies[len(processLatencies)-1].String(),
	)
}

func percentile(durations []time.Duration, p float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	idx := int(float64(len(durations)-1) * p / 100.0)

	return durations[idx]
}

func guardBench(t *testing.T) {
	if os.Getenv(envBench) == "" {
		t.Skip("LP2P_BENCH_TEST is not set")
	}
}
