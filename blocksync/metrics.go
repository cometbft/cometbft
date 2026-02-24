package blocksync

import (
	"github.com/cometbft/cometbft/types"
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

	// AlreadyIncludedBlocks blocks that were already included in the chain
	AlreadyIncludedBlocks metrics.Counter `metrics_name:"already_included_blocks"`

	// IngestedBlocks blocks that were rejected by the consensus
	IngestedBlocks metrics.Counter `metrics_name:"ingested_blocks"`

	// RejectedBlocks blocks that were rejected by the consensus
	RejectedBlocks metrics.Counter `metrics_name:"rejected_blocks"`
}

func (m *Metrics) recordBlockMetrics(block *types.Block) {
	m.NumTxs.Set(float64(len(block.Txs)))
	m.TotalTxs.Add(float64(len(block.Txs)))
	m.BlockSizeBytes.Set(float64(block.Size()))
	m.LatestBlockHeight.Set(float64(block.Height))
}
