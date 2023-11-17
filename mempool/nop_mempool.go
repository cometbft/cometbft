package mempool

import (
	"errors"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/types"
)

type NopMempool struct{}

var ErrNotAllowed = errors.New("not allowed with `nop` mempool")

var _ Mempool = &NopMempool{}

func (mem *NopMempool) CheckTx(tx types.Tx) (*abcicli.ReqRes, error) {
	return nil, ErrNotAllowed
}
func (mem *NopMempool) RemoveTxByKey(txKey types.TxKey) error               { return ErrNotAllowed }
func (mem *NopMempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) types.Txs { return nil }
func (mem *NopMempool) ReapMaxTxs(max int) types.Txs                        { return nil }
func (mem *NopMempool) Lock()                                               {}
func (mem *NopMempool) Unlock()                                             {}
func (mem *NopMempool) Update(
	height int64,
	txs types.Txs,
	txResults []*abci.ExecTxResult,
	preCheck PreCheckFunc,
	postCheck PostCheckFunc,
) error {
	return nil
}
func (mem *NopMempool) FlushAppConn() error { return nil }
func (mem *NopMempool) Flush()              {}
func (mem *NopMempool) TxsAvailable() <-chan struct{} {
	txsAvailable := make(chan struct{}, 1)
	close(txsAvailable)
	return txsAvailable
}
func (mem *NopMempool) EnableTxsAvailable()                             {}
func (mem *NopMempool) SetTxRemovedCallback(cb func(txKey types.TxKey)) {}
func (mem *NopMempool) Size() int                                       { return 0 }
func (mem *NopMempool) SizeBytes() int64                                { return 0 }
