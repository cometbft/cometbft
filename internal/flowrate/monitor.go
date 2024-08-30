package flowrate

import (
	"time"

	legacyimpl "github.com/cometbft/cometbft/internal/flowrate/legacy"
)

type Monitor struct {
	*legacyimpl.Monitor
}

type Status = legacyimpl.Status

func New(sampleRate, windowSize time.Duration) *Monitor {
	return NewLegacy(sampleRate, windowSize)
}

func NewLegacy(sampleRate, windowSize time.Duration) *Monitor {
	m := &Monitor{}
	m.Monitor = legacyimpl.New(sampleRate, windowSize)
	return m
}
