package null

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/pubsub/query"
	"github.com/cometbft/cometbft/internal/state/txindex"
	"github.com/cometbft/cometbft/libs/log"
)

var _ txindex.TxIndexer = (*TxIndex)(nil)

// TxIndex acts as a /dev/null.
type TxIndex struct{}

func (txi *TxIndex) SetRetainHeight(_ int64) error {
	return nil
}

func (txi *TxIndex) GetRetainHeight() (int64, error) {
	return 0, nil
}

func (txi *TxIndex) Prune(_ int64) (int64, int64, error) {
	return 0, 0, nil
}

// Get on a TxIndex is disabled and panics when invoked.
func (txi *TxIndex) Get(_ []byte) (*abci.TxResult, error) {
	return nil, errors.New(`indexing is disabled (set 'tx_index = "kv"' in config)`)
}

// AddBatch is a noop and always returns nil.
func (txi *TxIndex) AddBatch(_ *txindex.Batch) error {
	return nil
}

// Index is a noop and always returns nil.
func (txi *TxIndex) Index(_ *abci.TxResult) error {
	return nil
}

func (txi *TxIndex) Search(_ context.Context, _ *query.Query) ([]*abci.TxResult, error) {
	return []*abci.TxResult{}, nil
}

func (txi *TxIndex) SetLogger(log.Logger) {
}
