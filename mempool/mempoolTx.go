package mempool

import (
	"sync/atomic"

	"github.com/cometbft/cometbft/types"
)

// mempoolTx is an entry in the mempool.
type mempoolTx struct {
	tx        types.Tx
	height    int64
	gasWanted int64
}

// Height returns the height for this transaction.
func (memTx *mempoolTx) Height() int64 {
	return atomic.LoadInt64(&memTx.height)
}
