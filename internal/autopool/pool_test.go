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
			//nolint:errcheck // test
			go pool.PushPriority(value, priority)
		}

		expect := func() bool {
			mu.Lock()
			defer mu.Unlock()

			return len(results) == itemsCount
		}

		assert.Eventually(t, expect, 2*time.Second, 500*time.Millisecond)
	})

	// TestPriorityQueueWakeup verifies that the pool correctly processes all
	// items from the priority queue even when multiple items are pushed while
	// the consumer is blocked.
	//
	// This tests a race condition where:
	// 1. The pipePriorityQueue goroutine is blocked sending to inbound (full)
	// 2. Multiple items are pushed to the priority queue rapidly
	// 3. Only one signal is sent (valuesAvailable channel has capacity 1)
	// 4. Without the fix, the consumer would only process one item per wakeup
	//    and items would get stuck in the queue
	t.Run("PriorityQueueWakeup", func(t *testing.T) {
		scaler := NewThroughputLatencyScaler(
			1,  // min workers
			1,  // max workers (single worker to control flow)
			90.0,
			10*time.Millisecond,
			100*time.Millisecond, // long epoch to prevent scaling during test
			logger,
		)

		var (
			consumed       atomic.Int64
			blockConsumer  = make(chan struct{})
			consumerUnblocked atomic.Bool
		)

		consumer := func(_ int) {
			// Block the first message until we signal
			if !consumerUnblocked.Load() {
				<-blockConsumer
				consumerUnblocked.Store(true)
			}
			consumed.Add(1)
		}

		// Small inbound capacity (1) so pipePriorityQueue blocks quickly
		pool := New(
			scaler,
			consumer,
			1, // capacity of 1 - will block after one item
			WithLogger[int](logger),
			WithPriorityQueue[int](NewPriorityQueue(10)),
		)

		pool.Start()
		defer pool.Stop()

		const totalItems = 100

		// Push all items rapidly to the priority queue.
		// The pipePriorityQueue will:
		// 1. Pop first item, send to inbound (succeeds, inbound now has 1)
		// 2. Pop second item, block trying to send to inbound (full)
		// 3. Meanwhile, items 3-100 are pushed, but only ONE signal is sent
		//    (valuesAvailable channel is already full after item 3's push)
		for i := 1; i <= totalItems; i++ {
			err := pool.PushPriority(i, 1)
			require.NoError(t, err)
		}

		// Give pipePriorityQueue time to pop items and get blocked
		time.Sleep(50 * time.Millisecond)

		// Now unblock the consumer - this allows the worker to process messages
		// and frees up the inbound channel
		close(blockConsumer)

		// All items should eventually be processed
		require.Eventually(t, func() bool {
			return consumed.Load() == totalItems
		}, 5*time.Second, 10*time.Millisecond,
			"expected %d items to be consumed, got %d", totalItems, consumed.Load())
	})

	// TestPriorityQueueBurstAfterDrain verifies that items pushed after the
	// queue has been fully drained are still processed correctly.
	t.Run("PriorityQueueBurstAfterDrain", func(t *testing.T) {
		scaler := NewThroughputLatencyScaler(
			2,
			4,
			90.0,
			10*time.Millisecond,
			50*time.Millisecond,
			logger,
		)

		var consumed atomic.Int64

		consumer := func(_ int) {
			consumed.Add(1)
		}

		pool := New(
			scaler,
			consumer,
			10,
			WithLogger[int](logger),
			WithPriorityQueue[int](NewPriorityQueue(10)),
		)

		pool.Start()
		defer pool.Stop()

		// First burst
		for i := 0; i < 50; i++ {
			require.NoError(t, pool.PushPriority(i, 1))
		}

		// Wait for queue to drain completely
		require.Eventually(t, func() bool {
			return consumed.Load() == 50
		}, 2*time.Second, 10*time.Millisecond)

		// Second burst - the pipePriorityQueue should be waiting on
		// WaitForValues() now and must wake up for these new items
		for i := 50; i < 100; i++ {
			require.NoError(t, pool.PushPriority(i, 1))
		}

		// All items from second burst should also be processed
		require.Eventually(t, func() bool {
			return consumed.Load() == 100
		}, 2*time.Second, 10*time.Millisecond,
			"expected 100 items consumed after second burst, got %d", consumed.Load())
	})

}
