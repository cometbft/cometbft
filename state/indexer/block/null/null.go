package null

import (
	"context"
	"errors"

	"github.com/cometbft/cometbft/internal/pubsub/query"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/types"
)

var _ indexer.BlockIndexer = (*BlockerIndexer)(nil)

// TxIndex implements a no-op block indexer.
type BlockerIndexer struct{}

func (*BlockerIndexer) SetRetainHeight(_ int64) error {
	return nil
}

func (*BlockerIndexer) GetRetainHeight() (int64, error) {
	return 0, nil
}

func (*BlockerIndexer) Prune(_ int64) (numPruned, newRetainHeight int64, err error) {
	return 0, 0, nil
}

func (*BlockerIndexer) Has(int64) (bool, error) {
	return false, errors.New(`indexing is disabled (set 'tx_index = "kv"' in config)`)
}

func (*BlockerIndexer) Index(types.EventDataNewBlockEvents) error {
	return nil
}

func (*BlockerIndexer) Search(context.Context, *query.Query) ([]int64, error) {
	return []int64{}, nil
}

func (*BlockerIndexer) SetLogger(log.Logger) {
}
