package state

import (
	"github.com/go-kit/kit/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "state"
)

//go:generate go run ../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
type Metrics struct {
	// Time between BeginBlock and EndBlock in ms.
	BlockProcessingTime metrics.Histogram `metrics_buckettype:"lin" metrics_bucketsizes:"1, 10, 10"`

	// ConsensusParamUpdates is the total number of times the application has
	// udated the consensus params since process start.
	ConsensusParamUpdates metrics.Counter

	// ValidatorSetUpdates is the total number of times the application has
	// udated the validator set since process start.
	ValidatorSetUpdates metrics.Counter

	// PruningServiceBlockRetainHeight is the accepted block
	// retain height set by the data companion
	PruningServiceBlockRetainHeight metrics.Gauge

	// PruningServiceBlockResultsRetainHeight is the accepted block results
	// retain height set by the data companion
	PruningServiceBlockResultsRetainHeight metrics.Gauge

	// ApplicationBlockRetainHeight is the accepted block
	// retain height set by the application
	ApplicationBlockRetainHeight metrics.Gauge

	// BlockStoreBaseHeight shows the first height at which
	// a block is available
	BlockStoreBaseHeight metrics.Gauge

	// ABCIResultsBaseHeight shows the first height at which
	// abci results are available
	ABCIResultsBaseHeight metrics.Gauge
}
