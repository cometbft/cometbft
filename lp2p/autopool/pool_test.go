package autopool

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	logger := log.TestingLogger()

	t.Run("Push", func(t *testing.T) {
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

		// ARRANGE
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
	})

	t.Run("PushPriority", func(t *testing.T) {
		// ARRANGE
		// Given scaler
		scaler := NewThroughputLatencyScaler(
			4,
			8,
			90.0,
			10*time.Millisecond,
			20*time.Millisecond,
			logger,
		)

		logger.Info(scaler.String())

		// Given pool with decision counters
		mu := sync.Mutex{}
		results := []string{}

		consumer := func(msg string) {
			mu.Lock()
			defer mu.Unlock()
			results = append(results, msg)
		}

		pool := New(
			scaler,
			consumer,
			1024,
			WithLogger[string](logger),
			WithPriorityQueue[string](NewPriorityQueue(10)),
		)

		pool.Start()
		defer pool.Stop()

		const itemsCount = 1000

		for i := 0; i < itemsCount; i++ {
			var (
				// [1-10]
				priority = 1 + i%10
				value    = fmt.Sprintf("message-%.2d", i)
			)

			// simulate async push
			go pool.PushPriority(value, priority)
		}

		expect := func() bool {
			mu.Lock()
			defer mu.Unlock()

			return len(results) == itemsCount
		}

		assert.Eventually(t, expect, 2*time.Second, 500*time.Millisecond)
	})

}
