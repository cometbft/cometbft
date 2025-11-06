package priorityqueue

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/test/utils"
)

func TestQueue(t *testing.T) {
	t.Run("Push", func(t *testing.T) {
		// ARRANGE
		const iterations = 100_000

		queue := New(iterations)

		// Given random data
		inputs := genRandomData(iterations)

		// ACT
		durations := []time.Duration{}

		for _, item := range inputs {
			now := time.Now()
			queue.Push(item.value, item.priority)
			durations = append(durations, time.Since(now))
		}

		// ASSERT
		utils.LogDurationStats(t, "Push duration:", durations)

		t.Run("Consume", func(t *testing.T) {
			consumed := 0
			durations := []time.Duration{}
			valuesChan, cancel := queue.Consumer()

			defer cancel()

			lastConsumed := time.Now()

			for range valuesChan {
				durations = append(durations, time.Since(lastConsumed))

				consumed++
				if consumed == len(inputs) {
					break
				}

				lastConsumed = time.Now()
			}

			utils.LogDurationStats(t, "Consume duration:", durations)
		})
	})

	t.Run("PushAndConsume", func(t *testing.T) {
		// ARRANGE
		const iterations = 100_000

		queue := New(iterations)

		// Given random data
		inputs := genRandomData(iterations)

		// ACT
		pushDurations := []time.Duration{}
		consumeDurations := []time.Duration{}

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			for _, item := range inputs {
				now := time.Now()
				queue.Push(item.value, item.priority)
				pushDurations = append(pushDurations, time.Since(now))
			}
		}()

		go func() {
			defer wg.Done()

			valuesChan, cancel := queue.Consumer()
			defer cancel()

			consumed := 0
			lastConsumed := time.Now()

			for range valuesChan {
				consumeDurations = append(consumeDurations, time.Since(lastConsumed))

				consumed++
				if consumed == iterations {
					return
				}

				lastConsumed = time.Now()
			}
		}()

		wg.Wait()

		// ASSERT
		utils.LogDurationStats(t, "Push duration:", pushDurations)
		utils.LogDurationStats(t, "Consume duration:", consumeDurations)
	})
}

type testData struct {
	priority uint64
	value    string
}

func genRandomData(count int) []testData {
	out := []testData{}

	for i := 0; i < count; i++ {
		out = append(out, testData{
			priority: rand.Uint64() % 10,
			value:    fmt.Sprintf("value-%d", i),
		})
	}

	return out
}
