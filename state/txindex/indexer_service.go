package txindex

import (
	"context"
	"encoding/binary"
	"errors"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/types"
)

// XXX/TODO: These types should be moved to the indexer package.

const (
	subscriber = "IndexerService"
)

var (
	LastTxIndexerRetainHeightKey    = []byte("lastTxIndexerRetainHeight")
	LastBlockIndexerRetainHeightKey = []byte("lastBlockIndexerRetainHeight")
	IndexerRetainHeightKey          = []byte("indexerRetainHeight")

	ErrKeyNotFound        = errors.New("key not found")
	ErrInvalidHeightValue = errors.New("invalid height value")
)

// IndexerService connects event bus, transaction and block indexers together in
// order to index transactions and blocks coming from the event bus.
type IndexerService struct {
	service.BaseService

	txIdxr           TxIndexer
	blockIdxr        indexer.BlockIndexer
	indexerStore     dbm.DB
	eventBus         *types.EventBus
	terminateOnError bool
}

// NewIndexerService returns a new service instance.
func NewIndexerService(
	txIdxr TxIndexer,
	blockIdxr indexer.BlockIndexer,
	indexerStore dbm.DB,
	eventBus *types.EventBus,
	terminateOnError bool,
) *IndexerService {

	is := &IndexerService{
		txIdxr:           txIdxr,
		blockIdxr:        blockIdxr,
		indexerStore:     indexerStore,
		eventBus:         eventBus,
		terminateOnError: terminateOnError}
	is.BaseService = *service.NewBaseService(nil, "IndexerService", is)
	return is
}

// OnStart implements service.Service by subscribing for all transactions
// and indexing them by events.
func (is *IndexerService) OnStart() error {
	// Use SubscribeUnbuffered here to ensure both subscriptions does not get
	// canceled due to not pulling messages fast enough. Cause this might
	// sometimes happen when there are no other subscribers.
	blockSub, err := is.eventBus.SubscribeUnbuffered(
		context.Background(),
		subscriber,
		types.EventQueryNewBlockEvents)
	if err != nil {
		return err
	}

	txsSub, err := is.eventBus.SubscribeUnbuffered(context.Background(), subscriber, types.EventQueryTx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-blockSub.Canceled():
				return
			case msg := <-blockSub.Out():
				eventNewBlockEvents := msg.Data().(types.EventDataNewBlockEvents)
				height := eventNewBlockEvents.Height
				numTxs := eventNewBlockEvents.NumTxs

				batch := NewBatch(numTxs)

				for i := int64(0); i < numTxs; i++ {
					msg2 := <-txsSub.Out()
					txResult := msg2.Data().(types.EventDataTx).TxResult

					if err = batch.Add(&txResult); err != nil {
						is.Logger.Error(
							"failed to add tx to batch",
							"height", height,
							"index", txResult.Index,
							"err", err,
						)

						if is.terminateOnError {
							if err := is.Stop(); err != nil {
								is.Logger.Error("failed to stop", "err", err)
							}
							return
						}
					}
				}

				if err := is.blockIdxr.Index(eventNewBlockEvents); err != nil {
					is.Logger.Error("failed to index block", "height", height, "err", err)
					if is.terminateOnError {
						if err := is.Stop(); err != nil {
							is.Logger.Error("failed to stop", "err", err)
						}
						return
					}
				} else {
					is.Logger.Info("indexed block events", "height", height)
				}

				if err = is.txIdxr.AddBatch(batch); err != nil {
					is.Logger.Error("failed to index block txs", "height", height, "err", err)
					if is.terminateOnError {
						if err := is.Stop(); err != nil {
							is.Logger.Error("failed to stop", "err", err)
						}
						return
					}
				} else {
					is.Logger.Debug("indexed transactions", "height", height, "num_txs", numTxs)
				}
			}
		}
	}()
	return nil
}

func (is *IndexerService) Prune(retainHeight int64) {
	lastTxIndexerRetainHeight, err := is.getLastTxIndexerRetainHeight()
	if err != nil {
		panic(err)
	}

	lastBlockIndexerRetainHeight, err := is.getLastBlockIndexerRetainHeight()
	if err != nil {
		panic(err)
	}

	if retainHeight <= lastTxIndexerRetainHeight && retainHeight <= lastBlockIndexerRetainHeight {
		return
	}

	txPrunedHeight, err := is.txIdxr.Prune(lastTxIndexerRetainHeight, retainHeight)
	is.setLastTxIndexerRetainHeight(txPrunedHeight)
	if err != nil {
		panic(err)
	}

	blockPrunedHeight, err := is.blockIdxr.Prune(lastBlockIndexerRetainHeight, retainHeight)
	is.setLastBlockIndexerRetainHeight(blockPrunedHeight)
	if err != nil {
		panic(err)
	}
}

func (is *IndexerService) SaveIndexerRetainHeight(height int64) error {
	currentValue, err := is.GetIndexerRetainHeight()
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return err
	}
	if height <= currentValue {
		return nil
	}
	return is.indexerStore.SetSync(IndexerRetainHeightKey, int64ToBytes(height))
}

func (is *IndexerService) GetIndexerRetainHeight() (int64, error) {
	buf, err := is.indexerStore.Get(IndexerRetainHeightKey)
	if err != nil {
		return 0, err
	}
	height := int64FromBytes(buf)

	if height < 0 {
		return 0, ErrInvalidHeightValue
	}

	return height, nil
}

func (is *IndexerService) getLastTxIndexerRetainHeight() (int64, error) {
	bz, err := is.indexerStore.Get(LastTxIndexerRetainHeightKey)
	if errors.Is(err, ErrKeyNotFound) {
		return 0, nil
	}
	height := int64FromBytes(bz)
	if height < 0 {
		return 0, ErrInvalidHeightValue
	}
	return height, nil
}

func (is *IndexerService) setLastTxIndexerRetainHeight(height int64) {
	currentHeight, err := is.getLastTxIndexerRetainHeight()
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		panic(err)
	}
	if height < currentHeight {
		return
	}
	if err := is.indexerStore.SetSync(LastTxIndexerRetainHeightKey, int64ToBytes(height)); err != nil {
		panic(err)
	}
}

func (is *IndexerService) getLastBlockIndexerRetainHeight() (int64, error) {
	bz, err := is.indexerStore.Get(LastBlockIndexerRetainHeightKey)
	if errors.Is(err, ErrKeyNotFound) {
		return 0, nil
	}
	height := int64FromBytes(bz)
	if height < 0 {
		return 0, ErrInvalidHeightValue
	}
	return height, nil
}

func (is *IndexerService) setLastBlockIndexerRetainHeight(height int64) {
	currentHeight, err := is.getLastBlockIndexerRetainHeight()
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		panic(err)
	}
	if height < currentHeight {
		return
	}
	if err := is.indexerStore.SetSync(LastBlockIndexerRetainHeightKey, int64ToBytes(height)); err != nil {
		panic(err)
	}
}

// OnStop implements service.Service by unsubscribing from all transactions.
func (is *IndexerService) OnStop() {
	if is.eventBus.IsRunning() {
		_ = is.eventBus.UnsubscribeAll(context.Background(), subscriber)
	}
}

// ----- Util
func int64FromBytes(bz []byte) int64 {
	v, _ := binary.Varint(bz)
	return v
}

func int64ToBytes(i int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, i)
	return buf[:n]
}
