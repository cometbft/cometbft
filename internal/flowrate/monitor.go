package flowrate

import (
	"time"

	legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"
	"github.com/cometbft/cometbft/internal/flowrate/xrate"
)

var DefaultImplementation = "xrate"

type Status = legacyimpl.Status

// FlowRate implements a rate limiter used by p2p connections.
type FlowRate interface {
	// Limit restricts the instantaneous (per-sample) data flow to rate bytes per
	// second. It returns the maximum number of bytes (0 <= n <= want) that may be
	// transferred immediately without exceeding the limit. If block == true, the
	// call blocks until n > 0. want is returned unmodified if want < 1, rate < 1,
	// or the transfer is inactive (after a call to Done).
	Limit(want int, rate int64, block bool) (n int)

	// Update records the transfer of n bytes and returns n. It should be called
	// after each Read/Write operation, even if n is 0.
	Update(n int) int

	// Status returns current transfer status information.
	Status() Status

	// Hack to set the current rEMA.
	// Used by block sync reactor.
	SetREMA(rEMA float64)
}

type Monitor struct {
	FlowRate
}

func New(sampleRate, windowSize time.Duration) *Monitor {
	return NewWithImpl(sampleRate, windowSize, DefaultImplementation)
}

func NewWithImpl(sampleRate, windowSize time.Duration, impl string) *Monitor {
	switch impl {
	case "legacy":
		return NewLegacy(sampleRate, windowSize)
	case "xrate":
		return NewRate(sampleRate, windowSize)
	default:
		return nil
	}
}

func NewLegacy(sampleRate, windowSize time.Duration) *Monitor {
	m := &Monitor{}
	m.FlowRate = legacyimpl.New(sampleRate, windowSize)
	return m
}

func NewRate(sampleRate, windowSize time.Duration) *Monitor {
	m := &Monitor{}
	m.FlowRate = xrate.New(sampleRate, windowSize)
	return m
}
