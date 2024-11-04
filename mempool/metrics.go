package mempool

import (
	"github.com/cometbft/cometbft/libs/metrics"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "mempool"
)

//go:generate go run ../scripts/metricsgen -struct=Metrics

// Metrics contains metrics exposed by this package.
// see MetricsProvider for descriptions.
type Metrics struct {
	// Number of uncommitted transactions in the mempool.
	//
	// Deprecated: this value can be obtained as the sum of LaneSize.
	Size metrics.Gauge

	// Total size of the mempool in bytes.
	//
	// Deprecated: this value can be obtained as the sum of LaneBytes.
	SizeBytes metrics.Gauge

	// Number of uncommitted transactions per lane.
	LaneSize metrics.Gauge `metrics_labels:"lane"`

	// Number of used bytes per lane.
	LaneBytes metrics.Gauge `metrics_labels:"lane"`

	// TxLifeSpan measures the time each transaction has in the mempool, since
	// the time it enters until it is removed.
	// metrics:Duration in ms of a transaction in the mempool.
	TxLifeSpan metrics.Histogram `metrics_bucketsizes:"50,100,200,500,1000" metrics_labels:"lane"`

	// Histogram of transaction sizes in bytes.
	TxSizeBytes metrics.Histogram `metrics_bucketsizes:"1,3,7" metrics_buckettype:"exp"`

	// FailedTxs defines the number of failed transactions. These are
	// transactions that failed to make it into the mempool because they were
	// deemed invalid.
	// metrics:Number of failed transactions.
	FailedTxs metrics.Counter

	// RejectedTxs defines the number of rejected transactions. These are
	// transactions that failed to make it into the mempool due to resource
	// limits, e.g. mempool is full.
	// metrics:Number of rejected transactions.
	RejectedTxs metrics.Counter

	// EvictedTxs defines the number of evicted transactions. These are valid
	// transactions that passed CheckTx and make it into the mempool but later
	// became invalid.
	// metrics:Number of evicted transactions.
	EvictedTxs metrics.Counter

	// Number of times transactions are rechecked in the mempool.
	RecheckTimes metrics.Counter

	// Number of times transactions were received more than once.
	// metrics:Number of duplicate transaction reception.
	AlreadyReceivedTxs metrics.Counter

	// Number of connections being actively used for gossiping transactions
	// (experimental feature).
	ActiveOutboundConnections metrics.Gauge

	// Cumulative time spent rechecking transactions
	RecheckDurationSeconds metrics.Gauge
}
