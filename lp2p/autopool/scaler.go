package autopool

import (
	"slices"
	"time"

	"github.com/cometbft/cometbft/libs/log"
)

// ThroughputLatencyScaler is a scaler that scales the number of workers based on throughput
// The more messages are processed, the more workers are scaled up.
// But if latency percentile is too high, the scaler will shrink the number of workers.
type ThroughputLatencyScaler struct {
	minWorkers int
	maxWorkers int

	thresholdPercentile float64
	thresholdLatency    time.Duration
	epochDuration       time.Duration

	// latencies of the current epoch
	epochLatencies []time.Duration

	// throughput of the last epoch
	lastThroughput uint

	logger log.Logger
}

const (
	ShouldStay uint8 = iota
	ShouldScale
	ShouldShrink
)

func NewThroughputLatencyScaler(
	min, max int,
	thresholdPercentile float64,
	thresholdLatency time.Duration,
	epochDuration time.Duration,
	logger log.Logger,
) *ThroughputLatencyScaler {
	if min <= 0 {
		min = 4
	}

	if max <= 0 {
		max = min * 2
	}

	if thresholdPercentile < 0.0 || thresholdPercentile > 100.0 {
		thresholdPercentile = 90.0
	}

	if thresholdLatency <= 0 {
		thresholdLatency = 100 * time.Millisecond
	}

	return &ThroughputLatencyScaler{
		minWorkers:          min,
		maxWorkers:          max,
		thresholdPercentile: thresholdPercentile,
		thresholdLatency:    thresholdLatency,
		epochDuration:       epochDuration,
		epochLatencies:      []time.Duration{},
		lastThroughput:      0,
		logger:              logger.With("component", "scaler"),
	}
}

func (s *ThroughputLatencyScaler) EpochDuration() time.Duration {
	return s.epochDuration
}

func (s *ThroughputLatencyScaler) Min() int {
	return s.minWorkers
}

func (s *ThroughputLatencyScaler) Max() int {
	return s.maxWorkers
}

func (s *ThroughputLatencyScaler) Track(duration time.Duration) {
	s.epochLatencies = append(s.epochLatencies, duration)
}

func (s *ThroughputLatencyScaler) Decide(currentNumWorkers int) uint8 {
	var (
		epochThroughput    = uint(len(s.epochLatencies))
		epochDurPercentile = calculatePercentile(s.epochLatencies, s.thresholdPercentile)
		logger             = s.logger.With(
			"current_workers", currentNumWorkers,
			"throughput", epochThroughput,
			"prev_throughput", s.lastThroughput,
			"epoch_dur_percentile_ms", epochDurPercentile.Milliseconds(),
		)
	)

	logger.Debug("Deciding")

	// handle inactivity
	if epochThroughput == 0 {
		s.lastThroughput = 0
		s.epochLatencies = []time.Duration{}

		if currentNumWorkers == s.minWorkers {
			logger.Debug("Inactivity detected, at min workers")
			return ShouldStay
		}

		logger.Debug("Inactivity detected, recommending shrink")
		return ShouldShrink
	}

	decision := ShouldStay

	if s.lastThroughput <= epochThroughput {
		logger.Debug("Scaling")
		decision = ShouldScale
	} else {
		logger.Debug("Shrinking")
		decision = ShouldShrink
	}

	if decision == ShouldScale && epochDurPercentile >= s.thresholdLatency {
		logger.Debug("Wanted to scale, but latency is too high. Shrinking")
		decision = ShouldShrink
	}

	if decision == ShouldScale && currentNumWorkers >= s.maxWorkers {
		logger.Debug("Wanted to scale, but at max workers. Staying")
		decision = ShouldStay
	}

	if decision == ShouldShrink && currentNumWorkers <= s.minWorkers {
		logger.Debug("Wanted to shrink, but at min workers. Staying")
		decision = ShouldStay
	}

	// update state
	s.lastThroughput = epochThroughput
	s.epochLatencies = make([]time.Duration, 0, len(s.epochLatencies))

	return decision
}

func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	// should not happen
	if percentile < 0.0 || percentile > 100.0 {
		panic("percentile must be between 0.0 and 100.0")
	}

	if len(durations) == 0 {
		return 0
	}

	slices.Sort(durations)

	idx := int(float64(len(durations)) * percentile / 100.0)
	if idx >= len(durations) {
		idx = len(durations) - 1
	}

	return durations[idx]
}
