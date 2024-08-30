package flowrate

import (
	"time"

	legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"
	"github.com/cometbft/cometbft/internal/flowrate/xrate"
)

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
