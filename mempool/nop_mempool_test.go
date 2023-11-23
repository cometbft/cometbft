package mempool

import (
	"testing"

	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/assert"
)

var tx = types.Tx([]byte{0x01})

func TestNopMempool_Basic(t *testing.T) {
	mem := &NopMempool{}

	assert.Equal(t, 0, mem.Size())
	assert.Equal(t, int64(0), mem.SizeBytes())

	_, err := mem.CheckTx(tx)
	assert.Equal(t, errNotAllowed, err)

	err = mem.RemoveTxByKey(tx.Key())
	assert.Equal(t, errNotAllowed, err)

	txs := mem.ReapMaxBytesMaxGas(0, 0)
	assert.Nil(t, txs)

	txs = mem.ReapMaxTxs(0)
	assert.Nil(t, txs)

	err = mem.FlushAppConn()
	assert.NoError(t, err)

	err = mem.Update(0, nil, nil, nil, nil)
	assert.NoError(t, err)

	txsAvailable := mem.TxsAvailable()
	assert.Nil(t, txsAvailable)
}
