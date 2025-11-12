package queue

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/cometbft/cometbft/test/utils"
	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	q := New()

	q.Push(1)
	q.Push(2)
	q.Push(3)

	require.Equal(t, 3, q.Len())

	pop := func(expected int) {
		v, ok := q.Pop()
		require.True(t, ok)
		require.Equal(t, expected, v)
	}

	pop(1)
	pop(2)

	q.Push(4)

	pop(3)
	pop(4)

	require.Equal(t, 0, q.Len())

	_, ok := q.Pop()
	require.False(t, ok)
}

func TestPriorityQueue(t *testing.T) {
	const (
		iterations = 100_000
		priorities = 10
	)

	t.Run("Push", func(t *testing.T) {
		// ARRANGE
		queue := NewPriorityQueue(priorities)

		// Given random data
		inputs := genRandomData(iterations, priorities)

		// ACT
		durations := []time.Duration{}

		for _, item := range inputs {
			now := time.Now()

			err := queue.Push(item.value, int(item.priority))
			if err != nil {
				t.Fatalf("failed to push item: %v", err)
			}

			durations = append(durations, time.Since(now))
		}

		// ASSERT
		utils.LogDurationStats(t, "Push duration:", durations)

		t.Run("Consume", func(t *testing.T) {
			consumed := 0
			durations := []time.Duration{}

			lastConsumed := time.Now()

			for {
				_, ok := queue.Pop()
				if !ok {
					break
				}

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
		queue := NewPriorityQueue(priorities)

		// Given random data
		inputs := genRandomData(iterations, priorities)

		// ACT
		pushDurations := make([]time.Duration, 0, iterations)
		consumeDurations := make([]time.Duration, 0, iterations)
		consumedValues := make([]string, 0, iterations)

		wg := sync.WaitGroup{}
		wg.Add(2)

		start := time.Now()

		go func() {
			defer wg.Done()
			for _, item := range inputs {
				now := time.Now()
				err := queue.Push(item.value, int(item.priority))
				if err != nil {
					// should not happen
					panic(err)
				}

				pushDurations = append(pushDurations, time.Since(now))
			}
		}()

		go func() {
			defer wg.Done()

			consumed := 0
			lastConsumed := time.Now()

			for {
				value, ok := queue.Pop()
				if !ok {
					time.Sleep(10 * time.Millisecond)
					lastConsumed = time.Now()
					continue
				}

				consumeDurations = append(consumeDurations, time.Since(lastConsumed))
				consumedValues = append(consumedValues, value.(string))
				consumed++

				if consumed == iterations {
					return
				}

				lastConsumed = time.Now()
			}
		}()

		wg.Wait()

		// ASSERT
		t.Logf("Time taken: %s", time.Since(start))
		utils.LogDurationStats(t, "Push duration:", pushDurations)
		utils.LogDurationStats(t, "Consume duration:", consumeDurations)

		// check that all values were consumed
		actualValues := make(map[string]struct{}, len(consumedValues))
		for _, value := range consumedValues {
			actualValues[value] = struct{}{}
		}

		for _, item := range inputs {
			if _, ok := actualValues[item.value]; !ok {
				t.Fatalf("value %s not consumed", item.value)
			}
		}
	})
}

type testData struct {
	priority uint64
	value    string
}

func genRandomData(count int, priorities uint64) []testData {
	out := []testData{}

	for i := 0; i < count; i++ {
		out = append(out, testData{
			priority: 1 + (rand.Uint64() % priorities),
			value:    fmt.Sprintf("value-%d", i),
		})
	}

	return out
}
