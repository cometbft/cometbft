package mocks

import (
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/state/indexer/block"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/types"
)

func CreateAndStartIndexerService(
	config *cfg.Config,
	chainID string,
	dbProvider cfg.DBProvider,
	logger log.Logger,
) (*txindex.IndexerService, error) {
	var (
		txIndexer    txindex.TxIndexer
		blockIndexer indexer.BlockIndexer
	)
	txIndexer, blockIndexer, indexerStore, err := block.IndexerFromConfig(config, dbProvider, chainID)
	if err != nil {
		return nil, err
	}

	eventBus, err := createAndStartEventBus(logger)
	if err != nil {
		panic(err)
	}

	txIndexer.SetLogger(logger.With("module", "txindex"))
	blockIndexer.SetLogger(logger.With("module", "txindex"))
	indexerService := txindex.NewIndexerService(txIndexer, blockIndexer, indexerStore, eventBus, false)
	indexerService.SetLogger(logger.With("module", "txindex"))

	if err := indexerService.Start(); err != nil {
		return nil, err
	}

	return indexerService, nil
}

func createAndStartEventBus(logger log.Logger) (*types.EventBus, error) {
	eventBus := types.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	if err := eventBus.Start(); err != nil {
		return nil, err
	}
	return eventBus, nil
}
