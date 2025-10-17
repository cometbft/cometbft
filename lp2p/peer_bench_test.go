package lp2p

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
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/test/utils"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/require"
)

type lp2pUnidirectionalConfig struct {
	duration        time.Duration
	sendConcurrency int
	receiveDelay    time.Duration
}

// TestBenchLP2PUnidirectional this test measures messages libp2p throughput and latency by sending messages
// from peer A to peer B in a single direction. The payload only contains a timestamp,
// so we can measure e2e latency.
func TestBenchLP2PUnidirectional(t *testing.T) {
	utils.GuardP2PBenchTest(t)

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

	t.Logf("host1: %+v", host1.AddrInfo().String())
	t.Logf("host2: %+v", host2.AddrInfo().String())

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
		start         = time.Now()
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

		t.Logf("Finished receiver goroutine")

		close(waitProcessing)
	}()

	t.Log("Sending messages...")

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

	messagesLost := sendSuccesses.Load() - receiveSuccesses.Load() - receiveFailures.Load()
	messageLossPercentage := float64(messagesLost) / float64(sendSuccesses.Load()+sendFailures.Load()) * 100

	t.Logf("Messages lost: %d (%.3f%%)", int64(messagesLost), messageLossPercentage)

	utils.LogDurationStats(t, "Receive latency:", receiveLatencies)
	utils.LogDurationStats(t, "Process latency:", processLatencies)
}
