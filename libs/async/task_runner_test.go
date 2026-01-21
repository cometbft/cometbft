package async_test

import (
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

		for i := 0; i < 5; i++ {
			tr.Enqueue(func() { results = append(results, i) })
		}
		tr.Stop()

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
		var panicCaught bool
		tr := async.NewTaskRunner(10, func(r any, stack []byte) {
			panicCaught = true
		})

		tr.Enqueue(func() { panic("test panic") })
		tr.Enqueue(func() { executed = true })

		tr.Stop()
		require.True(t, executed, "task after panic should still execute")
		require.True(t, panicCaught, "panic should be caught")
	})

	t.Run("enqueue returns false after stop", func(t *testing.T) {
		tr := async.NewTaskRunner(1, nil)
		tr.Stop()

		ok := tr.Enqueue(func() {})
		require.False(t, ok, "Enqueue should return false after Stop")
	})

	t.Run("drains remaining tasks on stop", func(t *testing.T) {
		var count int
		tr := async.NewTaskRunner(10, nil)

		// Fill the buffer
		for i := 0; i < 5; i++ {
			tr.Enqueue(func() { count++ })
		}

		tr.Stop()
		require.Equal(t, 5, count, "all enqueued tasks should execute")
	})
}
