package blocksync

import (
	"reflect"
	"unsafe"

	"github.com/cometbft/cometbft/types"
	"github.com/go-kit/kit/metrics"
	prometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
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

func (m *Metrics) recordBlockMetrics(block *types.Block) {
	m.NumTxs.Set(float64(len(block.Data.Txs)))
	m.TotalTxs.Add(float64(len(block.Data.Txs)))
	m.BlockSizeBytes.Set(float64(block.Size()))
	m.LatestBlockHeight.Set(float64(block.Height))
}

// UnregisterGauge removes a gauge metric from the Prometheus registry. No new
// data will be generated in the http /metrics endpoint for this metric.
func UnregisterGauge(g metrics.Gauge) bool {
	metric, ok := g.(*prometheus.Gauge)
	if !ok {
		return false
	}

	// stdprometheus.Unregister takes an stdprometheus.Collector. Here we access
	// prometheus.Gauge's unexported field "gv", which implements the interface
	// stdprometheus.Collector.
	gv := reflect.ValueOf(metric).Elem().FieldByName("gv")
	gaugeVec, ok := getUnexportedField(gv).(*stdprometheus.GaugeVec)
	if !ok {
		return false
	}

	return stdprometheus.Unregister(gaugeVec)
}

// See https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}
