package mempool

import (
	"sync/atomic"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
)

type SyncReactor interface {
	p2p.Reactor

	WaitSync() bool

	EnableInOutTxs()
}

type BaseSyncReactor struct {
	p2p.BaseReactor
	Config *cfg.MempoolConfig

	waitSync   atomic.Bool
	waitSyncCh chan struct{} // for signaling when to start receiving and sending txs
}

func NewBaseSyncReactor(config *cfg.MempoolConfig, waitSync bool) *BaseSyncReactor {
	baseR := &BaseSyncReactor{Config: config}
	if waitSync {
		baseR.waitSync.Store(true)
		baseR.waitSyncCh = make(chan struct{})
	}
	baseR.BaseReactor = *p2p.NewBaseReactor("Mempool", baseR)
	return baseR
}

func (memR *BaseSyncReactor) EnableInOutTxs() {
	memR.Logger.Info("enabling inbound and outbound transactions")
	if !memR.waitSync.CompareAndSwap(true, false) {
		return
	}

	// Releases all the blocked broadcastTxRoutine instances.
	if memR.Config.Broadcast {
		close(memR.waitSyncCh)
	}
}

func (memR *BaseSyncReactor) WaitSync() bool {
	return memR.waitSync.Load()
}