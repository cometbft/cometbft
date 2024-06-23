package null

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/pubsub/query"
)

// TxIndex acts as a /dev/null.
type TxIndex[BatchT, PaginationT any] struct{}

func (*TxIndex[_, _]) SetRetainHeight(int64) error {
	return nil
}

func (*TxIndex[_, _]) GetRetainHeight() (int64, error) {
	return 0, nil
}

func (*TxIndex[_, _]) Prune(int64) (numPruned, newRetainHeight int64, err error) {
	return 0, 0, nil
}

// Get on a TxIndex is disabled and panics when invoked.
func (*TxIndex[_, _]) Get([]byte) (*abci.TxResult, error) {
	return nil, errors.New(`indexing is disabled (set 'tx_index = "kv"' in config)`)
}

// AddBatch is a noop and always returns nil.
func (*TxIndex[BatchT, _]) AddBatch(BatchT) error {
	return nil
}

// Index is a noop and always returns nil.
func (*TxIndex[_, _]) Index(*abci.TxResult) error {
	return nil
}

func (*TxIndex[_, PaginationT]) Search(context.Context, *query.Query, PaginationT) ([]*abci.TxResult, int, error) {
	return []*abci.TxResult{}, 0, nil
}

func (*TxIndex[_, _]) SetLogger(log.Logger) {
}
