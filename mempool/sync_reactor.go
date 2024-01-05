// Package mempool provides the implementation for transaction memory pool.
package mempool

import (
	"sync/atomic"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
)

// SyncReactor is an interface that wraps the p2p.Reactor interface
// with additional methods for synchronization.
type SyncReactor interface {
	p2p.Reactor

	// WaitSync returns true if the reactor is waiting for synchronization.
	WaitSync() bool

	// EnableInOutTxs enables the reactor to start receiving and sending transactions.
	EnableInOutTxs()
}

// BaseSyncReactor is the base implementation of the SyncReactor interface.
// It embeds p2p.BaseReactor and includes additional fields and methods for synchronization.
type BaseSyncReactor struct {
	p2p.BaseReactor
	Config *cfg.MempoolConfig

	waitSync   atomic.Bool   // atomic boolean for synchronization status
	waitSyncCh chan struct{} // channel used for signaling when to start receiving and sending txs
}

// NewBaseSyncReactor creates a new BaseSyncReactor with the given configuration and synchronization status.
// If waitSync is true, it initializes the waitSyncCh channel.
func NewBaseSyncReactor(config *cfg.MempoolConfig, waitSync bool) *BaseSyncReactor {
	baseR := &BaseSyncReactor{Config: config}
	if waitSync {
		baseR.waitSync.Store(true)
		baseR.waitSyncCh = make(chan struct{})
	}
	baseR.BaseReactor = *p2p.NewBaseReactor("Mempool", baseR)
	return baseR
}

// EnableInOutTxs enables the reactor to start receiving and sending transactions.
// If the reactor is not waiting for synchronization, it returns immediately.
// If the reactor is configured to broadcast transactions, it closes the waitSyncCh channel
// to signal all blocked broadcastTxRoutine instances to start.
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

// WaitSync returns the current synchronization status of the reactor.
func (memR *BaseSyncReactor) WaitSync() bool {
	return memR.waitSync.Load()
}
