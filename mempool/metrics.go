package mempool

import (
	"github.com/go-kit/kit/metrics"
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
	Size metrics.Gauge

	// Total size of the mempool in bytes.
	SizeBytes metrics.Gauge

	// Histogram of transaction sizes in bytes.
	TxSizeBytes metrics.Histogram `metrics_buckettype:"exp" metrics_bucketsizes:"1,3,7"`

	// Number of failed transactions.
	FailedTxs metrics.Counter

	// RejectedTxs defines the number of rejected transactions. These are
	// transactions that passed CheckTx but failed to make it into the mempool
	// due to resource limits, e.g. mempool is full and no lower priority
	// transactions exist in the mempool.
	//metrics:Number of rejected transactions.
	RejectedTxs metrics.Counter

	// Number of times transactions are rechecked in the mempool.
	RecheckTimes metrics.Counter

	// Number of times transactions were received more than once.
	//metrics:Number of duplicate transaction reception.
	AlreadyReceivedTxs metrics.Counter
	
	// RequestedTxs defines the number of times that the node requested a
	// tx to a peer
	//metrics:Number of requested transactions (WantTx messages).
	RequestedTxs metrics.Counter

	// RerequestedTxs defines the number of times that a requested tx
	// never received a response in time and a new request was made.
	//metrics:Number of re-requested transactions.
	RerequestedTxs metrics.Counter

	// NoPeerForTx counts the number of times the reactor exhaust the list of
	// peers looking for a transaction for which it has received a SeenTx
	// message.
	//metrics:Number of times we cannot find a peer for a tx.
	NoPeerForTx metrics.Counter

	// Number of connections being actively used for gossiping transactions
	// (experimental feature).
	ActiveOutboundConnections metrics.Gauge
}
