package node

import (
	"github.com/cometbft/cometbft/config"
)

func ExampleMetrics() {
	config := &config.InstrumentationConfig{
		Namespace:  "cometbft",
		Prometheus: true,
	}

	// Register the metrics once
	shared := PrometheusMetrics(config, "chain_id")

	// Create a new metrics object from the shared one
	_ = MetricsProvider(func(chainID string) *Metrics {
		return shared.With("chain_id", chainID)
	})
}
