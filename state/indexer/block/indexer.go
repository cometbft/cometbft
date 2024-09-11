package block

import (
	"errors"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/state/indexer"
	blockidxkv "github.com/cometbft/cometbft/state/indexer/block/kv"
	blockidxnull "github.com/cometbft/cometbft/state/indexer/block/null"
	"github.com/cometbft/cometbft/state/indexer/sink/psql"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/state/txindex/null"
)

// IndexerFromConfig constructs a slice of indexer.EventSink using the provided
// configuration.
func IndexerFromConfig(cfg *config.Config, dbProvider config.DBProvider, chainID string) (
	txIdx txindex.TxIndexer, blockIdx indexer.BlockIndexer, err error,
) {
	txidx, blkidx, _, err := IndexerFromConfigWithDisabledIndexers(cfg, dbProvider, chainID)
	return txidx, blkidx, err
}

// IndexerFromConfigWithDisabledIndexers constructs a slice of indexer.EventSink using the provided
// configuration. If all indexers are disabled in the configuration, it returns null indexers.
// Otherwise, it creates the appropriate indexers based on the configuration.
func IndexerFromConfigWithDisabledIndexers(cfg *config.Config, dbProvider config.DBProvider, chainID string) (
	txIdx txindex.TxIndexer, blockIdx indexer.BlockIndexer, allIndexersDisabled bool, err error,
) {
	switch cfg.TxIndex.Indexer {
	case "kv":
		store, err := dbProvider(&config.DBContext{ID: "tx_index", Config: cfg})
		if err != nil {
			return nil, nil, false, err
		}

		return kv.NewTxIndex(store), blockidxkv.New(dbm.NewPrefixDB(store, []byte("block_events"))), false, nil

	case "psql":
		conn := cfg.TxIndex.PsqlConn
		if conn == "" {
			return nil, nil, false, errors.New("the psql connection settings cannot be empty")
		}
		es, err := psql.NewEventSink(cfg.TxIndex.PsqlConn, chainID)
		if err != nil {
			return nil, nil, false, fmt.Errorf("creating psql indexer: %w", err)
		}
		return es.TxIndexer(), es.BlockIndexer(), false, nil

	default:
		return &null.TxIndex{}, &blockidxnull.BlockerIndexer{}, true, nil
	}
}
