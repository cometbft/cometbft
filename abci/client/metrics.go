package abcicli

import (
	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "abci_client"
)

//go:generate go run ../../scripts/metricsgen -struct=Metrics

// Metrics contains the prometheus metrics exposed by the client package.
type Metrics struct {
	// Time waiting for a lock on each ABCI method.
	LockWaitSeconds metrics.Histogram `metrics_bucketsizes:".0001,.0004,.002,.009,.02,.1,.65,2,6,25" metrics_labels:"method"`
}
