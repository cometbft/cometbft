package utils

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
)

const EnvP2PBench = "P2P_BENCH_TEST"

func GuardP2PBenchTest(t *testing.T) {
	if os.Getenv(EnvP2PBench) == "" {
		t.Skip(EnvP2PBench + " is not set")
	}
}

func LogDurationStats(t *testing.T, title string, durations []time.Duration) {
	require.NotEmpty(t, durations)
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	t.Log(title)
	t.Logf(
		"  min: %s, p50: %s, p90: %s, p95: %s, p99: %s, max: %s",
		durations[0].String(),
		percentile(durations, 50).String(),
		percentile(durations, 90).String(),
		percentile(durations, 95).String(),
		percentile(durations, 99).String(),
		durations[len(durations)-1].String(),
	)
}

func LogBytesThroughputStats(t *testing.T, title string, bytes uint64, duration time.Duration) {
	require.NotEmpty(t, bytes)
	require.Greater(t, duration, time.Duration(0))

	bytesPerSec := float64(bytes) / duration.Seconds()
	t.Logf("%s: %s", title, formatBytesPerSecond(bytesPerSec))
}

func LogPerformanceStatsSendTimeframe(
	t *testing.T,
	start time.Time,
	sendSuccess, sendFailed, receivedSuccess uint64,
	sendBytesTotal, receiveBytesTotal uint64,
	receiveLatencies []time.Duration,
) {
	timeTaken := time.Since(start)

	// if sendFailed is low, then this diff indicates that messages are QUEUED in the priority queue
	// and NOT lost. Since we're benchmarking a concrete time frame, we don't wait for
	// all messages to be processed, so they'll lower the "throughput" value.
	inFlight := sendSuccess - receivedSuccess
	inFlightPercentage := 100 * float64(inFlight) / float64(sendSuccess+sendFailed)

	t.Logf("Time taken: %s", timeTaken.String())
	t.Logf("Sent messages: %d", sendSuccess+sendFailed)
	t.Logf("  success: %d, failure %d", sendSuccess, sendFailed)
	t.Logf("  send RPS: %.0f", float64(sendSuccess)/timeTaken.Seconds())

	t.Logf("Received messages: %d", receivedSuccess)
	t.Logf("  success: %d, in-flight: %d", receivedSuccess, inFlight)
	t.Logf("  receive RPS: %.0f", float64(receivedSuccess)/timeTaken.Seconds())
	t.Logf("  still in-flight: %d (%.3f%%)", int64(inFlight), inFlightPercentage)

	LogBytesThroughputStats(t, "Send throughput", sendBytesTotal, timeTaken)
	LogBytesThroughputStats(t, "Receive throughput", receiveBytesTotal, timeTaken)
	LogDurationStats(t, "Receive latency:", receiveLatencies)
}

func WaitForProcessing(t *testing.T, ctx context.Context, expected, actual *atomic.Uint64) (completed bool) {
	t.Helper()

	const (
		interval    = 50 * time.Millisecond
		maxIdleWait = 1 * time.Second
	)

	var (
		lastValue        = actual.Load()
		lastProgressedAt = time.Now()
	)

	for {
		// load values once per loop
		expectedValue := expected.Load()
		actualValue := actual.Load()

		// check if we've completed
		if actualValue >= expectedValue {
			return true
		}

		// check if the context is done
		if ctx.Err() != nil {
			t.Logf("Context canceled. Expected: %d, Actual: %d", expected.Load(), actual.Load())
			return false
		}

		// update last value and time if we've made progress
		if actualValue > lastValue {
			lastValue = actualValue
			lastProgressedAt = time.Now()
		}

		// idle for too long
		if time.Since(lastProgressedAt) > maxIdleWait {
			t.Logf("Idle for too long. Expected: %d, Actual: %d", expectedValue, actualValue)
			return false
		}

		time.Sleep(interval)
	}

}

func formatBytesPerSecond(bps float64) string {
	const unit = 1024

	if bps < unit {
		return fmt.Sprintf("%.2f B/s", bps)
	}

	div, exp := float64(unit), 0
	for n := bps / unit; n >= unit && exp < len("KMG")-1; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB/s", bps/div, "KMG"[exp])
}

func percentile(durations []time.Duration, p float64) time.Duration {
	switch {
	case len(durations) == 0:
		return 0
	case p <= 0:
		return durations[0]
	case p >= 100:
		return durations[len(durations)-1]
	}

	rank := (p / 100) * float64(len(durations)-1)
	low := int(math.Floor(rank))
	high := int(math.Ceil(rank))

	if low == high {
		return durations[low]
	}

	// linear interpolation between low and high
	weight := rank - float64(low)
	dLow := float64(durations[low])
	dHigh := float64(durations[high])

	return time.Duration(dLow + (dHigh-dLow)*weight)
}

// PerfRecord dummy payload just to measure various performance metrics in benchmarks.
type PerfRecord struct {
	Payload     []byte
	SentAt      time.Time
	ReceivedAt  time.Time
	ProcessedAt time.Time
}

func (r *PerfRecord) AsEcho() *types.RequestEcho {
	msg := make([]byte, 8+len(r.Payload))
	binary.BigEndian.PutUint64(msg[:8], uint64(r.SentAt.UnixMicro()))
	copy(msg[8:], r.Payload)

	return &types.RequestEcho{Message: string(msg)}
}

func (r *PerfRecord) FromEcho(echo *types.RequestEcho) error {
	raw := []byte(echo.Message)
	if len(raw) < 8 {
		return fmt.Errorf("invalid perf record: got %d bytes", len(raw))
	}

	tsMicros := int64(binary.BigEndian.Uint64(raw[:8]))
	r.SentAt = time.UnixMicro(tsMicros)
	r.Payload = append(make([]byte, 0, len(raw)-8), raw[8:]...)

	return nil
}
