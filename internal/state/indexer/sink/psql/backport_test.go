package psql

import (
	"github.com/cometbft/cometbft/internal/state/indexer"
	"github.com/cometbft/cometbft/internal/state/txindex"
)

var (
	_ indexer.BlockIndexer = BackportBlockIndexer{}
	_ txindex.TxIndexer    = BackportTxIndexer{}
)
