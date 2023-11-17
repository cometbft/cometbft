package mempool

import (
	"errors"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
)

// NopMempool is a mempool that does nothing.
//
// The ABCI app is responsible for storing, disseminating, and proposing transactions.
// See [ADR-111](../docs/architecture/adr-111-nop-mempool.md).
type NopMempool struct{}

// ErrNotAllowed indicates that the operation is not allowed with `nop` mempool.
var ErrNotAllowed = errors.New("not allowed with `nop` mempool")

var _ Mempool = &NopMempool{}

// CheckTx always returns ErrNotAllowed.
func (*NopMempool) CheckTx(tx types.Tx) (*abcicli.ReqRes, error) {
	return nil, ErrNotAllowed
}

// RemoveTxByKey always returns ErrNotAllowed.
func (*NopMempool) RemoveTxByKey(txKey types.TxKey) error { return ErrNotAllowed }

// ReapMaxBytesMaxGas always returns nil.
func (*NopMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs { return nil }

// ReapMaxTxs always returns nil.
func (*NopMempool) ReapMaxTxs(max int) types.Txs { return nil }

// Lock does nothing.
func (*NopMempool) Lock() {}

// Unlock does nothing.
func (*NopMempool) Unlock() {}

// Update does nothing.
func (*NopMempool) Update(
	height int64,
	txs types.Txs,
	txResults []*abci.ExecTxResult,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) error {
	return nil
}

// FlushAppConn does nothing.
func (*NopMempool) FlushAppConn() error { return nil }

// Flush does nothing.
func (*NopMempool) Flush() {}

// TxsAvailable always returns a closed channel.
func (*NopMempool) TxsAvailable() <-chan struct{} {
	txsAvailable := make(chan struct{}, 1)
	close(txsAvailable)
	return txsAvailable
}

// EnableTxsAvailable does nothing.
func (*NopMempool) EnableTxsAvailable() {}

// SetTxRemovedCallback does nothing.
func (*NopMempool) SetTxRemovedCallback(cb func(txKey types.TxKey)) {}

// Size always returns 0.
func (*NopMempool) Size() int { return 0 }

// SizeBytes always returns 0.
func (*NopMempool) SizeBytes() int64 { return 0 }

// NopMempoolReactor is a mempool reactor that does nothing.
type NopMempoolReactor struct {
	service.BaseService
}

// NewNopMempoolReactor returns a new `nop` reactor.
//
// To be used only in RPC.
func NewNopMempoolReactor() *NopMempoolReactor {
	return &NopMempoolReactor{*service.NewBaseService(nil, "NopMempoolReactor", nil)}
}

var _ p2p.Reactor = &NopMempoolReactor{}

// WaitSync always returns false.
func (*NopMempoolReactor) WaitSync() bool { return false }

// GetChannels always returns nil.
func (*NopMempoolReactor) GetChannels() []*p2p.ChannelDescriptor { return nil }

// AddPeer does nothing.
func (*NopMempoolReactor) AddPeer(peer p2p.Peer) {}

// InitPeer always returns nil.
func (*NopMempoolReactor) InitPeer(peer p2p.Peer) p2p.Peer { return nil }

// RemovePeer does nothing.
func (*NopMempoolReactor) RemovePeer(peer p2p.Peer, reason interface{}) {}

// Receive does nothing.
func (*NopMempoolReactor) Receive(p2p.Envelope) {}

// SetSwitch does nothing.
func (*NopMempoolReactor) SetSwitch(sw *p2p.Switch) {}
