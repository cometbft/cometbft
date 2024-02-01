// Package metrics exports metrics types so that applications can use them.
package node

import (
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/blocksync"
	"github.com/cometbft/cometbft/internal/consensus"
	sm "github.com/cometbft/cometbft/internal/state"
	"github.com/cometbft/cometbft/internal/statesync"
	"github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
)

type Metrics struct {
	Consensus *ConsensusMetrics
	P2P       *P2PMetrics
	Mempool   *MempoolMetrics
	State     *StateMetrics
	Proxy     *ProxyMetrics
	Blocksync *BlocksyncMetrics
	Statesync *StatesyncMetrics
}

func PrometheusMetrics(config *config.InstrumentationConfig, labels ...string) *Metrics {
	return &Metrics{consensus.PrometheusMetrics(config.Namespace, labels...),
		p2p.PrometheusMetrics(config.Namespace, labels...),
		mempool.PrometheusMetrics(config.Namespace, labels...),
		sm.PrometheusMetrics(config.Namespace, labels...),
		proxy.PrometheusMetrics(config.Namespace, labels...),
		blocksync.PrometheusMetrics(config.Namespace, labels...),
		statesync.PrometheusMetrics(config.Namespace, labels...)}
}

func NopMetrics() *Metrics {
	return &Metrics{consensus.NopMetrics(), p2p.NopMetrics(), mempool.NopMetrics(), sm.NopMetrics(), proxy.NopMetrics(), blocksync.NopMetrics(), statesync.NopMetrics()}
}

type ConsensusMetrics = consensus.Metrics
type P2PMetrics = p2p.Metrics
type MempoolMetrics = mempool.Metrics
type StateMetrics = sm.Metrics
type ProxyMetrics = proxy.Metrics
type BlocksyncMetrics = blocksync.Metrics
type StatesyncMetrics = statesync.Metrics

func (a *Metrics) With(labelsAndValues ...string) *Metrics {
	return &Metrics{
		Consensus: a.Consensus.With(labelsAndValues...),
		P2P:       a.P2P.With(labelsAndValues...),
		Mempool:   a.Mempool.With(labelsAndValues...),
		State:     a.State.With(labelsAndValues...),
		Proxy:     a.Proxy.With(labelsAndValues...),
		Blocksync: a.Blocksync.With(labelsAndValues...),
		Statesync: a.Statesync.With(labelsAndValues...),
	}
}
