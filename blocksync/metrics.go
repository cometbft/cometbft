package blocksync

import (
	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "blocksync"
)

//go:generate go run ../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Whether or not a node is block syncing. 1 if yes, 0 if no.
	Syncing metrics.Gauge

	// Number of transactions in the latest block.
	NumTxs metrics.Gauge
	// Total number of transactions.
	TotalTxs metrics.Gauge
	// Size of the latest block.
	BlockSizeBytes metrics.Gauge
	// The height of the latest block.
	LatestBlockHeight metrics.Gauge
}
