package basic

import "github.com/cometbft/cometbft/internal/metrics"

//go:generate go run ../../../../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// simple metric that tracks the height of the chain.
	Height metrics.Gauge
}
