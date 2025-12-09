package utils

import (
	"math"
	"os"
	"sort"
	"testing"
	"time"

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
