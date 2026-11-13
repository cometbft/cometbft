package node

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAsyncEventRunnerRunsTasksFIFO(t *testing.T) {
	runner := newAsyncEventRunner(3)
	defer runner.Stop()

	done := make(chan int, 3)
	for i := 0; i < 3; i++ {
		runner.Submit(func() { done <- i })
	}

	for i := 0; i < 3; i++ {
		select {
		case got := <-done:
			require.Equal(t, i, got)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for task %d", i)
		}
	}
}

func TestAsyncEventRunnerStopDrainsAcceptedTasks(t *testing.T) {
	runner := newAsyncEventRunner(2)

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	completed := make(chan int, 3)
	runner.Submit(func() {
		close(firstStarted)
		<-releaseFirst
		completed <- 0
	})
	<-firstStarted
	runner.Submit(func() { completed <- 1 })
	runner.Submit(func() { completed <- 2 })

	stopDone := make(chan struct{})
	go func() {
		runner.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
		t.Fatal("Stop returned while a task was still running")
	case <-time.After(10 * time.Millisecond):
	}
	close(releaseFirst)

	select {
	case <-stopDone:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return after queued tasks completed")
	}
	for i := 0; i < 3; i++ {
		require.Equal(t, i, <-completed)
	}
}

func TestAsyncEventRunnerSubmitReturnsAfterStop(t *testing.T) {
	runner := newAsyncEventRunner(0)
	runner.Stop()

	done := make(chan struct{})
	go func() {
		runner.Submit(func() {
			t.Error("task submitted after Stop must not run")
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Submit blocked after runner stopped")
	}
}

func TestAsyncEventRunnerStartsOnFirstSubmit(t *testing.T) {
	runner := newAsyncEventRunner(1)
	defer runner.Stop()

	done := make(chan struct{})
	go func() {
		for i := 0; i < defaultAsyncEventQueueSize+1; i++ {
			runner.Submit(func() {})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("submissions blocked because the runner did not start")
	}
}
