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
	LastIndexerRetainHeightKey = []byte("lastIndexerRetainHeight")
	IndexerRetainHeightKey     = []byte("indexerRetainHeight")

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
	lastRetainHeight, err := is.getLastIndexerRetainHeight()
	if err != nil {
		panic(err)
	}
	if retainHeight <= lastRetainHeight {
		return
	}
	is.txIdxr.Prune(lastRetainHeight, retainHeight)
	is.blockIdxr.Prune(lastRetainHeight, retainHeight)
	is.setLastIndexerRetainHeight(retainHeight)
}

func (is *IndexerService) SaveIndexerRetainHeight(height int64) error {
	return is.indexerStore.SetSync(IndexerRetainHeightKey, int64ToBytes(height))
}

func (is *IndexerService) GetIndexerRetainHeight() (int64, error) {
	buf, err := is.getValue(IndexerRetainHeightKey)
	if err != nil {
		return 0, err
	}
	height := int64FromBytes(buf)

	if height < 0 {
		return 0, ErrInvalidHeightValue
	}

	return height, nil
}

func (is *IndexerService) getLastIndexerRetainHeight() (int64, error) {
	bz, err := is.getValue(LastIndexerRetainHeightKey)
	if errors.Is(err, ErrKeyNotFound) {
		return 0, nil
	}
	height := int64FromBytes(bz)
	if height < 0 {
		return 0, ErrInvalidHeightValue
	}
	return height, nil
}

func (is *IndexerService) setLastIndexerRetainHeight(height int64) {
	if err := is.indexerStore.SetSync(LastIndexerRetainHeightKey, int64ToBytes(height)); err != nil {
		panic(err)
	}
}

func (is *IndexerService) getValue(key []byte) ([]byte, error) {
	bz, err := is.indexerStore.Get(key)
	if err != nil {
		return nil, err
	}

	if len(bz) == 0 {
		return nil, ErrKeyNotFound
	}
	return bz, nil
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
