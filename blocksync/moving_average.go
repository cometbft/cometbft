package blocksync

import "time"

// MovingAverage maintains a sliding window average of time.Duration values.
type MovingAverage struct {
	buf   []time.Duration
	index int
	count int
	sum   time.Duration
}

// NewMovingAverage creates a moving average with the given window size.
func NewMovingAverage(windowSize int) *MovingAverage {
	return &MovingAverage{buf: make([]time.Duration, windowSize)}
}

// Add records a new value and returns the updated average.
// Always returns ok=true; use Avg to check whether any values exist.
func (ma *MovingAverage) Add(val time.Duration) (avg time.Duration, ok bool) {
	if ma.count < len(ma.buf) {
		ma.count++
	} else {
		ma.sum -= ma.buf[ma.index]
	}
	ma.buf[ma.index] = val
	ma.sum += val
	ma.index = (ma.index + 1) % len(ma.buf)
	return ma.sum / time.Duration(ma.count), true
}

// Avg returns the current moving average, or 0,false if empty.
func (ma *MovingAverage) Avg() (time.Duration, bool) {
	if ma.count == 0 {
		return 0, false
	}
	return ma.sum / time.Duration(ma.count), true
}

// Len returns the number of values in the window.
func (ma *MovingAverage) Len() int { return ma.count }

// Reset clears all values.
func (ma *MovingAverage) Reset() {
	ma.index, ma.count, ma.sum = 0, 0, 0
}
