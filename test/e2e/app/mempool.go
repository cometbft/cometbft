package app

import (
	"bytes"
	"fmt"
	"slices"
	"sync"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/types"
)

// AppMempool is a mock for app-side mempool
// (think of evm-mempool in cosmos-evm)
type AppMempool struct {
	txs    map[string]types.Tx
	logger log.Logger
	mu     sync.RWMutex
}

// NewAppMempool creates a new AppMempool
func NewAppMempool(logger log.Logger) *AppMempool {
	return &AppMempool{
		txs:    make(map[string]types.Tx),
		logger: logger,
	}
}

func (m *AppMempool) InsertTx(bz []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx := types.Tx(bz)
	txHash := fmt.Sprintf("%x", tx.Hash())

	if _, ok := m.txs[txHash]; ok {
		m.logger.Info("Tx already exists in app-side mempool", "tx", txHash)
	} else {
		m.txs[txHash] = tx
		m.logger.Info("Inserted tx into app-side mempool", "tx", txHash)
	}
}

func (m *AppMempool) ReapTxs(flush bool) types.Txs {
	m.mu.Lock()
	defer m.mu.Unlock()

	txs := make([]types.Tx, 0, len(m.txs))
	for _, tx := range m.txs {
		txs = append(txs, tx)
	}

	slices.SortFunc(txs, func(a, b types.Tx) int {
		return bytes.Compare(a, b)
	})

	if flush {
		m.txs = make(map[string]types.Tx)
	}

	return txs
}
