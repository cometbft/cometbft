package async_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/async"
)

func TestTaskRunner(t *testing.T) {
	t.Run("executes tasks in order", func(t *testing.T) {
		var results []int
		tr := async.NewTaskRunner(10, nil)
		defer tr.Stop()

		var wg sync.WaitGroup
		wg.Add(5)
		for i := 0; i < 5; i++ {
			i := i
			tr.Enqueue(func() {
				results = append(results, i)
				wg.Done()
			})
		}
		waitDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(waitDone)
		}()
		select {
		case <-waitDone:
		case <-time.After(time.Second):
			t.Fatal("tasks did not finish in time")
		}

		require.Equal(t, []int{0, 1, 2, 3, 4}, results)
	})

	t.Run("stop waits for in-flight task", func(t *testing.T) {
		started := make(chan struct{})
		done := make(chan struct{})
		tr := async.NewTaskRunner(1, nil)

		tr.Enqueue(func() {
			close(started)
			<-done // block until signaled
		})

		<-started // wait for task to start

		stopDone := make(chan struct{})
		go func() {
			tr.Stop()
			close(stopDone)
		}()

		// Stop should be blocked waiting for the task
		select {
		case <-stopDone:
			t.Fatal("Stop returned before task completed")
		case <-time.After(50 * time.Millisecond):
			// expected: stop is still waiting
		}

		close(done) // unblock the task

		select {
		case <-stopDone:
			// expected: stop completed after task finished
		case <-time.After(time.Second):
			t.Fatal("Stop did not return after task completed")
		}
	})

	t.Run("handles panic without crashing", func(t *testing.T) {
		var executed bool
		var panicCaught atomic.Bool
		done := make(chan struct{})
		tr := async.NewTaskRunner(10, func(r any, stack []byte) {
			panicCaught.Store(true)
		})

		tr.Enqueue(func() { panic("test panic") })
		tr.Enqueue(func() {
			executed = true
			close(done)
		})

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("task after panic did not finish in time")
		}
		tr.Stop()
		require.True(t, executed, "task after panic should still execute")
		require.True(t, panicCaught.Load(), "panic should be caught")
	})

	t.Run("enqueue returns false after stop", func(t *testing.T) {
		tr := async.NewTaskRunner(1, nil)
		tr.Stop()

		ok := tr.Enqueue(func() {})
		require.False(t, ok, "Enqueue should return false after Stop")
	})

}
