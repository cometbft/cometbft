package xrate

import (
	"time"

	legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"
)

type Monitor struct {
}

func New(sampleRate, windowSize time.Duration) *Monitor {
	return &Monitor{}
}

func (m *Monitor) Limit(want int, rate int64, block bool) (n int) {
	// TODO:
	return 0
}

func (m *Monitor) Update(n int) int {
	// TODO:
	return n
}

func (m *Monitor) SetREMA(rEMA float64) {
	// TODO:
}

func (m *Monitor) Status() legacyimpl.Status {
	// TODO:
	return legacyimpl.Status{}
}
