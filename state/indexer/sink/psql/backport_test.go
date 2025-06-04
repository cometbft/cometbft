package psql

import (
	"github.com/cometbft/cometbft/v2/state/indexer"
	"github.com/cometbft/cometbft/v2/state/txindex"
)

var (
	_ indexer.BlockIndexer = BackportBlockIndexer{}
	_ txindex.TxIndexer    = BackportTxIndexer{}
)
