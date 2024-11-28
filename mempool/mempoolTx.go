package mempool

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/types"
)

// mempoolTx is an entry in the mempool.
type mempoolTx struct {
	height    int64    // height that this tx had been validated in
	gasWanted int64    // amount of gas this tx states it will require
	tx        types.Tx // validated by the application
	lane      LaneID
	seq       int64
	timestamp time.Time // time when entry was created

	// ids of peers who've sent us this tx (as a map for quick lookups).
	// senders: PeerID -> struct{}
	senders sync.Map
}

func (memTx *mempoolTx) Tx() types.Tx {
	return memTx.tx
}

func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}

func (memTx *mempoolTx) GasWanted() int64 {
	return memTx.gasWanted
}

func (memTx *mempoolTx) IsSender(peerID nodekey.ID) bool {
	_, ok := memTx.senders.Load(peerID)
	return ok
}

// Add the peer ID to the list of senders. Return true iff it exists already in the list.
func (memTx *mempoolTx) addSender(peerID nodekey.ID) bool {
	if len(peerID) == 0 {
		return false
	}
	if _, loaded := memTx.senders.LoadOrStore(peerID, struct{}{}); loaded {
		return true
	}
	return false
}

func (memTx *mempoolTx) Senders() []nodekey.ID {
	senders := make([]nodekey.ID, 0)
	memTx.senders.Range(func(key, _ any) bool {
		senders = append(senders, key.(nodekey.ID))
		return true
	})
	return senders
}
