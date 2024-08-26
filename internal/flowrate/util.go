//
// Written by Maxim Khitrov (November 2012)
//

package flowrate

import (
	"math"
	"sync/atomic"
	"time"
)

// clockRate is the resolution and precision of clock().
const clockRate = 20 * time.Millisecond

var (
	hasInitializedClock = atomic.Bool{}
	currentClockValue   = atomic.Int64{}
	clockStartTime      = time.Time{}
)

// checks if the clock update timer is running. If not, sets clockStartTime and starts it.
func ensureClockRunning() {
	firstRun := hasInitializedClock.CompareAndSwap(false, true)
	if !firstRun {
		return
	}
	clockStartTime = time.Now().Round(clockRate)
	go runClockUpdates()
}

// increments the current clock value every clockRate interval.
func runClockUpdates() {
	// Create a ticker that sends the current time on the channel every clockRate interval
	ticker := time.Tick(clockRate)

	// First tick happens after clockrate, therefore initial value is 0.
	for t := range ticker {
		delta := t.Sub(clockStartTime)
		rounded := clockRound(delta)
		currentClockValue.Store(int64(rounded))
	}
}

// clock returns a low resolution timestamp relative to the process start time.
func clock() time.Duration {
	return time.Duration(currentClockValue.Load())
}

// clockToTime converts a clock() timestamp to an absolute time.Time value.
func clockToTime(c time.Duration) time.Time {
	return clockStartTime.Add(c)
}

// clockRound returns d rounded to the nearest clockRate increment.
func clockRound(d time.Duration) time.Duration {
	//nolint:durationcheck
	return ((d + clockRate>>1) / clockRate) * clockRate
}

// round returns x rounded to the nearest int64 (non-negative values only).
func round(x float64) int64 {
	if _, frac := math.Modf(x); frac >= 0.5 {
		return int64(math.Ceil(x))
	}
	return int64(math.Floor(x))
}
