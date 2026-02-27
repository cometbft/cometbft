package blocksync

import (
	"testing"
	"time"
)

func TestRequestRetryTimerStopDrainsChannel(t *testing.T) {
	rt := newRequestRetryTimer(5 * time.Millisecond)
	t.Cleanup(rt.Stop)

	time.Sleep(10 * time.Millisecond) // let the timer fire
	rt.Stop()

	select {
	case <-rt.C():
		t.Fatal("timer channel should have been drained")
	default:
	}
}

func TestRequestRetryTimerResetRearms(t *testing.T) {
	rt := newRequestRetryTimer(5 * time.Millisecond)
	t.Cleanup(rt.Stop)

	start := time.Now()
	rt.Reset()

	select {
	case <-rt.C():
		if time.Since(start) < 5*time.Millisecond {
			t.Fatalf("timer fired too early: %v", time.Since(start))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timer never fired after reset")
	}
}

func TestRequestRetryTimerResetDropsStaleEvents(t *testing.T) {
	rt := newRequestRetryTimer(20 * time.Millisecond)
	t.Cleanup(rt.Stop)

	for i := 0; i < 5; i++ {
		rt.Reset()
		select {
		case <-rt.C():
			t.Fatalf("unexpected retry event on iteration %d", i)
		case <-time.After(2 * time.Millisecond):
			// retry signal should not fire before duration elapses
		}
	}
}
