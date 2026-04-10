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

func PrintMatrixOnly() bool {
	return os.Getenv(EnvP2PBench) == "matrix-only"
}

func LogDurationStats(t *testing.T, title string, durations []time.Duration) map[string]any {
	t.Helper()

	require.NotEmpty(t, durations)
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	p50 := percentile(durations, 50)
	p90 := percentile(durations, 90)
	p95 := percentile(durations, 95)
	p99 := percentile(durations, 99)
	max := durations[len(durations)-1]

	t.Log(title)
	t.Logf(
		"  min: %s, p50: %s, p90: %s, p95: %s, p99: %s, max: %s",
		durations[0].String(),
		p50.String(),
		p90.String(),
		p95.String(),
		p99.String(),
		max.String(),
	)

	return map[string]any{
		"min_nanos": durations[0].Nanoseconds(),
		"p50_nanos": p50.Nanoseconds(),
		"p90_nanos": p90.Nanoseconds(),
		"p95_nanos": p95.Nanoseconds(),
		"p99_nanos": p99.Nanoseconds(),
		"max_nanos": max.Nanoseconds(),
	}
}

func LogBytesThroughputStats(t *testing.T, title string, bytes uint64, duration time.Duration) float64 {
	t.Helper()

	if duration == 0 {
		t.Logf("%s: N/A", title)
		return 0
	}

	bytesPerSec := float64(bytes) / duration.Seconds()
	t.Logf("%s: %s", title, formatBytesPerSecond(bytesPerSec))

	return bytesPerSec
}

func LogPerformanceStatsSend(
	t *testing.T,
	start time.Time,
	sendSuccess, sendFailed, receivedSuccess uint64,
	sendBytesTotal, receiveBytesTotal uint64,
	receiveLatencies []time.Duration,
) map[string]any {
	t.Helper()

	timeTaken := time.Since(start)

	sendTotal := sendSuccess + sendFailed
	rpsSent := float64(sendSuccess) / timeTaken.Seconds()

	t.Logf("Time taken: %s", timeTaken.String())
	t.Logf("Sent messages: %d", sendTotal)
	t.Logf("  success: %d, failure %d", sendSuccess, sendFailed)
	t.Logf("  send RPS: %.0f", rpsSent)

	// if sendFailed is low, then this diff indicates that messages are QUEUED in the priority queue
	// and NOT lost. Since we're benchmarking a concrete time frame, we don't wait for
	// all messages to be processed, so they'll lower the "throughput" value.
	inFlight := sendSuccess - receivedSuccess
	inFlightPercentage := 100 * float64(inFlight) / float64(sendSuccess+sendFailed)
	rpsReceived := float64(receivedSuccess) / timeTaken.Seconds()

	t.Logf("Received messages: %d", receivedSuccess)
	t.Logf("  success: %d, in-flight: %d", receivedSuccess, inFlight)
	t.Logf("  receive RPS: %.0f", rpsReceived)
	t.Logf("  still in-flight: %d (%.2f%%)", int64(inFlight), inFlightPercentage)

	sendThroughput := LogBytesThroughputStats(t, "Send throughput", sendBytesTotal, timeTaken)
	receiveThroughput := LogBytesThroughputStats(t, "Receive throughput", receiveBytesTotal, timeTaken)

	receiveLatenciesStats := LogDurationStats(t, "Receive latency", receiveLatencies)

	return map[string]any{
		"timeTaken": timeTaken,

		"messagesSentSuccess": sendSuccess,
		"messagesSentFailed":  sendFailed,
		"messagesSentTotal":   sendTotal,
		"messagesSentRPS":     rpsSent,

		"bytesSentTotal":         sendBytesTotal,
		"bytesSentThroughputSec": sendThroughput,

		"messagesReceived":           receivedSuccess,
		"messagesReceivedRPS":        rpsReceived,
		"messagesInFlight":           inFlight,
		"messagesInFlightPercentage": inFlightPercentage,

		"bytesReceivedTotal":         receiveBytesTotal,
		"bytesReceivedThroughputSec": receiveThroughput,

		"receiveLatenciesStats": receiveLatenciesStats,
	}
}

func LogPerformanceStatsBroadcast(
	t *testing.T,
	start time.Time,
	sentMessages int,
	receiveSuccess []atomic.Uint64,
	receiveBytes []atomic.Uint64,
	receiveLatencies [][]time.Duration,
	timeTaken []time.Duration,
) map[string]any {
	t.Helper()

	require.Equal(t, len(receiveSuccess), len(receiveBytes))
	require.Equal(t, len(receiveSuccess), len(receiveLatencies))
	require.Equal(t, len(receiveSuccess), len(timeTaken))

	var (
		totalReceived    uint64
		totalBytesPerSec float64
		peerStats        = make([]map[string]any, len(receiveSuccess))
	)

	for idx := range receiveSuccess {
		name := fmt.Sprintf("peer-%d", idx+1)

		peerReceived := receiveSuccess[idx].Load()
		peerBytes := receiveBytes[idx].Load()
		peerTimeTaken := timeTaken[idx]

		totalReceived += peerReceived
		totalBytesPerSec += float64(peerBytes) / peerTimeTaken.Seconds()

		inFlight := uint64(sentMessages) - peerReceived
		inFlightPercentage := 100 * float64(inFlight) / float64(sentMessages)
		rpsReceived := float64(peerReceived) / peerTimeTaken.Seconds()

		t.Logf("%s:", name)
		t.Logf("  received messages: %d", peerReceived)
		t.Logf("  receive RPS: %.0f", rpsReceived)
		t.Logf("  still in-flight: %d (%.2f%%)", int64(inFlight), inFlightPercentage)
		t.Logf("  time taken: %s", peerTimeTaken.String())

		receiveThroughput := LogBytesThroughputStats(t, "  receive throughput", peerBytes, peerTimeTaken)
		receiveLatenciesStats := LogDurationStats(t, "  receive latency", receiveLatencies[idx])

		peerStats[idx] = map[string]any{
			"name":      name,
			"timeTaken": peerTimeTaken,

			"messagesReceived":           peerReceived,
			"messagesReceivedRPS":        rpsReceived,
			"messagesInFlight":           inFlight,
			"messagesInFlightPercentage": inFlightPercentage,

			"bytesReceivedTotal":         peerBytes,
			"bytesReceivedThroughputSec": receiveThroughput,

			"receiveLatenciesStats": receiveLatenciesStats,
		}
	}

	t.Logf("Total received messages: %d", totalReceived)
	t.Logf("Total receive throughput: %s", formatBytesPerSecond(totalBytesPerSec))

	return map[string]any{
		"messagesSent": sentMessages,

		"totalReceived":            totalReceived,
		"totalReceivedBytesPerSec": totalBytesPerSec,

		"peerStats": peerStats,
	}
}

func WaitForProcessing(
	t *testing.T,
	ctx context.Context,
	name string,
	expected, actual *atomic.Uint64,
	maxIdleWait time.Duration,
) (completed bool) {
	t.Helper()

	const interval = 50 * time.Millisecond

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
			t.Logf("%s: Context canceled. Expected: %d, Actual: %d", name, expected.Load(), actual.Load())
			return false
		}

		// update last value and time if we've made progress
		if actualValue > lastValue {
			lastValue = actualValue
			lastProgressedAt = time.Now()
		}

		// idle for too long
		if time.Since(lastProgressedAt) > maxIdleWait {
			t.Logf("%s: Idle for too long. Expected: %d, Actual: %d", name, expectedValue, actualValue)
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
	Payload []byte

	SentAt      time.Time
	ReceivedAt  time.Time
	ProcessedAt time.Time
}

// message: [sent_at | payload]
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
