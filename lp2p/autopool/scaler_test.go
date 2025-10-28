package autopool

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"
)

func TestThroughputLatencyScaler(t *testing.T) {
	// ARRANGE
	// Given scaler with the following parameters:
	const (
		min                 = 4
		max                 = 10
		thresholdPercentile = 90.0
		thresholdLatency    = 100 * time.Millisecond
		epochDuration       = time.Second
	)

	logger := log.TestingLogger()

	scaler := NewThroughputLatencyScaler(
		min,
		max,
		thresholdPercentile,
		thresholdLatency,
		epochDuration,
		logger,
	)

	numWorkers := min

	for index, tt := range []struct {
		latenciesMS        []int
		expectedDecision   uint8
		expectedNumWorkers int
	}{
		{
			latenciesMS:        []int{},
			expectedDecision:   ShouldStay,
			expectedNumWorkers: min,
		},
		{
			// one very slow req, but we can't shrink below min
			latenciesMS:        []int{200},
			expectedDecision:   ShouldStay,
			expectedNumWorkers: min,
		},
		{
			latenciesMS:        []int{50, 50, 50},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 5,
		},
		{
			latenciesMS:        []int{50, 50, 50},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 6,
		},
		{
			latenciesMS:        []int{50, 50, 50, 80},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 7,
		},
		{
			latenciesMS:        []int{50, 50, 50, 80},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 8,
		},
		{
			latenciesMS:        []int{50, 50, 50, 80, 90, 90},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 9,
		},
		{
			latenciesMS:        []int{50, 50, 50},
			expectedDecision:   ShouldShrink,
			expectedNumWorkers: 8,
		},
		{
			latenciesMS:        []int{50, 50, 50, 80, 90, 90, 95, 99},
			expectedDecision:   ShouldScale,
			expectedNumWorkers: 9,
		},
		{
			// here we processed a lot of message, but the latency became too high, so we should shrink
			latenciesMS:        []int{50, 50, 50, 80, 90, 90, 95, 99, 100, 120, 130, 150},
			expectedDecision:   ShouldShrink,
			expectedNumWorkers: 8,
		},
	} {
		// ACT
		// Simulate processing of tracking and deciding
		for _, latencyMS := range tt.latenciesMS {
			lt := time.Duration(latencyMS) * time.Millisecond
			scaler.Track(lt)
		}

		decision := scaler.Decide(numWorkers)
		switch decision {
		case ShouldScale:
			numWorkers++
		case ShouldShrink:
			numWorkers--
		case ShouldStay:
			// do nothing
		}

		// ASSERT
		// Check if the decision and number of workers are as expected
		require.Equal(t, tt.expectedDecision, decision, "expected decision at step %d", index)
		require.Equal(t, tt.expectedNumWorkers, numWorkers, "expected number of workers at step %d", index)
	}

}
