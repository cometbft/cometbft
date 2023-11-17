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
type NopMempool struct{}

var ErrNotAllowed = errors.New("not allowed with `nop` mempool")

var _ Mempool = &NopMempool{}

func (*NopMempool) CheckTx(tx types.Tx) (*abcicli.ReqRes, error) {
	return nil, ErrNotAllowed
}
func (*NopMempool) RemoveTxByKey(txKey types.TxKey) error               { return ErrNotAllowed }
func (*NopMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs { return nil }
func (*NopMempool) ReapMaxTxs(max int) types.Txs                        { return nil }
func (*NopMempool) Lock()                                               {}
func (*NopMempool) Unlock()                                             {}
func (*NopMempool) Update(
	height int64,
	txs types.Txs,
	txResults []*abci.ExecTxResult,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) error {
	return nil
}
func (*NopMempool) FlushAppConn() error { return nil }
func (*NopMempool) Flush()              {}
func (*NopMempool) TxsAvailable() <-chan struct{} {
	txsAvailable := make(chan struct{}, 1)
	close(txsAvailable)
	return txsAvailable
}
func (*NopMempool) EnableTxsAvailable()                             {}
func (*NopMempool) SetTxRemovedCallback(cb func(txKey types.TxKey)) {}
func (*NopMempool) Size() int                                       { return 0 }
func (*NopMempool) SizeBytes() int64                                { return 0 }

// NopMempoolReactor is a mempool reactor that does not wait for syncing.
type NopMempoolReactor struct {
	service.BaseService
}

func NewNopMempoolReactor() *NopMempoolReactor {
	return &NopMempoolReactor{*service.NewBaseService(nil, "NopMempoolReactor", nil)}
}

var _ p2p.Reactor = &NopMempoolReactor{}

func (*NopMempoolReactor) WaitSync() bool { return false }

func (*NopMempoolReactor) GetChannels() []*p2p.ChannelDescriptor        { return nil }
func (*NopMempoolReactor) AddPeer(peer p2p.Peer)                        {}
func (*NopMempoolReactor) InitPeer(peer p2p.Peer) p2p.Peer              { return nil }
func (*NopMempoolReactor) RemovePeer(peer p2p.Peer, reason interface{}) {}
func (*NopMempoolReactor) Receive(p2p.Envelope)                         {}
func (*NopMempoolReactor) SetSwitch(sw *p2p.Switch)                     {}
