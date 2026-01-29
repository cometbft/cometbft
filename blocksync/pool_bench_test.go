package blocksync

import (
	"testing"
	"time"
)

// benchmarkRetryTimer simulates the retry loop in bpRequester. The legacy version
// allocated a fresh timer per iteration, while the optimized version reuses a single
// timer and simply resets it. The difference in allocations mirrors what we expect
// when a requester repeatedly retries fetching the same block.
func benchmarkRetryTimer(b *testing.B, reuse bool) {
	const iterationsPerOp = 256
	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if reuse {
			timer := time.NewTimer(time.Second)
			for i := 0; i < iterationsPerOp; i++ {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(time.Nanosecond)
			}
			timer.Stop()
			select {
			case <-timer.C:
			default:
			}
		} else {
			for i := 0; i < iterationsPerOp; i++ {
				timer := time.NewTimer(time.Nanosecond)
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
		}
	}
}

func BenchmarkRetryTimerLegacy(b *testing.B) {
	benchmarkRetryTimer(b, false)
}

func BenchmarkRetryTimerReusable(b *testing.B) {
	benchmarkRetryTimer(b, true)
}
