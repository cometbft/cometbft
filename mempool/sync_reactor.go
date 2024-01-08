package mempool

import (
	"sync/atomic"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
)

// A mempool reactor must implement WaitSyncP2PReactor to allow the node to transition from block
// sync or state sync to consensus mode.
type WaitSyncReactor struct {
	p2p.BaseReactor
	Config *cfg.MempoolConfig

	waitSync   atomic.Bool
	waitSyncCh chan struct{} // for signaling when to start receiving and sending txs
}

func NewWaitSyncReactor(config *cfg.MempoolConfig, waitSync bool) *WaitSyncReactor {
	baseR := &WaitSyncReactor{Config: config}
	if waitSync {
		baseR.waitSync.Store(true)
		baseR.waitSyncCh = make(chan struct{})
	}
	baseR.BaseReactor = *p2p.NewBaseReactor("Mempool", baseR)
	return baseR
}

func (memR *WaitSyncReactor) EnableInOutTxs() {
	memR.Logger.Info("enabling inbound and outbound transactions")
	if !memR.waitSync.CompareAndSwap(true, false) {
		return
	}

	// Releases all the blocked broadcastTxRoutine instances.
	if memR.Config.Broadcast {
		close(memR.waitSyncCh)
	}
}

func (memR *WaitSyncReactor) WaitSync() bool {
	return memR.waitSync.Load()
}
