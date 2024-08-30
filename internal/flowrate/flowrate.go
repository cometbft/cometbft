package flowrate

import legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"

var DefaultImplementation = "legacy"

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
