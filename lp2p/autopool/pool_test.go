package autopool

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	// ARRANGE
	const (
		testDuration = 10 * time.Second

		minWorkers = 4
		maxWorkers = 10

		// we want to test 90th percentile
		// because 90th rand() would most likely generate at least one value above the threshold
		latencyPercentile = 80.0
		latencyThreshold  = 20 * time.Millisecond

		// additional param to randomizer that generates values
		// between 0 and 110% of the threshold
		latencyMaxDiff = latencyThreshold / 10

		// how frequently we should autoscale
		epochDuration = latencyThreshold * 10
	)

	logger := log.TestingLogger()

	// Given scaler
	scaler := NewThroughputLatencyScaler(
		minWorkers,
		maxWorkers,
		latencyPercentile,
		latencyThreshold,
		epochDuration,
		logger,
	)

	logger.Info(scaler.String())

	// Given pool with decision counters
	var (
		scaled, shrunk, stayed atomic.Int64
		messagesPublished      = atomic.Int64{}
		messagesConsumed       = atomic.Int64{}
		queueCapacity          = 1024
		closeTest              = make(chan struct{})
		pool                   *Pool[time.Duration]
	)

	consumer := func(latency time.Duration) {
		time.Sleep(latency)

		total := messagesConsumed.Add(1)

		if total%200 == 0 {
			qs := pool.Len()
			if qs > 0 {
				logger.Info("Queue size", "size", qs)
			}
		}
	}

	pool = New(
		scaler,
		consumer,
		queueCapacity,
		WithLogger[time.Duration](logger),
		WithOnScale[time.Duration](func() { scaled.Add(1) }),
		WithOnShrink[time.Duration](func() { shrunk.Add(1) }),
		WithOnStay[time.Duration](func() { stayed.Add(1) }),
	)

	// Given fake producer that produces messages
	timer := time.NewTimer(testDuration)
	defer timer.Stop()

	go func() {
		defer close(closeTest)

		for {
			select {
			case <-timer.C:
				return
			default:
				// sleep for [0...latencyThreshold)
				consumerDelay := time.Duration(
					rand.Uint64() % uint64(latencyThreshold+latencyMaxDiff),
				)

				pool.Push(consumerDelay)
				messagesPublished.Add(1)

				// also pause producer for shorter amount of time
				time.Sleep(consumerDelay / 8)
			}
		}
	}()

	// ACT
	// start pool
	logger.Info("Running", "duration", testDuration)
	pool.Start()

	// wait for test to finish
	logger.Info("Waiting for test to finish")
	<-closeTest

	// stop pool and close queue
	logger.Info("Stopping pool")
	pool.Stop()

	// ASSERT
	t.Logf("Messages published: %d", messagesPublished.Load())
	t.Logf("Messages consumed: %d", messagesConsumed.Load())
	t.Logf("Scaled %d times", scaled.Load())
	t.Logf("Shrunk %d times", shrunk.Load())
	t.Logf("Stayed %d times", stayed.Load())

	// 20%
	delta := float64(messagesPublished.Load()) * 0.2

	require.InDelta(t, messagesConsumed.Load(), messagesPublished.Load(), delta, "Consumer is too slow")
}
