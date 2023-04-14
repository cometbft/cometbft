package mempool

import (
	"sync"

	"github.com/cometbft/cometbft/types"
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
	AlreadyReceivedTxs metrics.Counter

	// Histogram of times a transaction was received.
	TimesTxsWereReceived metrics.Histogram `metrics_buckettype:"exp" metrics_bucketsizes:"1,2,5"`
	// For keeping track of the number of times each transaction in the mempool
	// was received and whether that value was observed.
	txsReceived sync.Map
}

type txReceivedCounter struct {
	count       uint64
	wasObserved bool
}

func (m *Metrics) countOneTimeTxWasReceived(tx types.TxKey) {
	value, _ := m.txsReceived.LoadOrStore(tx, txReceivedCounter{0, false})
	counter := value.(txReceivedCounter)
	counter.count += 1
	m.txsReceived.Store(tx, counter)
}

func (m *Metrics) resetTimesTxWasReceived(tx types.TxKey) {
	m.txsReceived.Delete(tx)
}

func (m *Metrics) observeTimesTxsWereReceived() {
	m.txsReceived.Range(func(key, value interface{}) bool {
		counter := value.(txReceivedCounter)
		if !counter.wasObserved {
			m.TimesTxsWereReceived.Observe(float64(counter.count))
			counter.wasObserved = true
			m.txsReceived.Store(key.(types.TxKey), counter)
		}
		return true
	})
}
