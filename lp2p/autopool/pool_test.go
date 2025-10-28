package autopool

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
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

	// Given fake producer that produces messages
	var (
		messagesPublished = atomic.Int64{}
		messagesConsumed  = atomic.Int64{}
		queue             = make(chan time.Duration, 1024)
		closeTest         = make(chan struct{})
	)

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

				queue <- consumerDelay
				messagesPublished.Add(1)

				// also pause producer for shorter amount of time
				time.Sleep(consumerDelay / 8)
			}
		}
	}()

	// Given consumer that consumes messages
	consumer := func(latency time.Duration) {
		time.Sleep(latency)

		total := messagesConsumed.Add(1)

		if total%100 == 0 {
			qs := len(queue)
			if qs > 0 {
				logger.Info("Queue size", "size", qs)
			}
		}
	}

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

	// Given pool
	pool := New(scaler, queue, consumer, logger)

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
	close(queue)

	// ASSERT
	t.Logf("Messages published: %d", messagesPublished.Load())
	t.Logf("Messages consumed: %d", messagesConsumed.Load())
}
