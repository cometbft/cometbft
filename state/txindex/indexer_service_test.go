package txindex_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/state/indexer"
	"github.com/stretchr/testify/require"

	db "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	blockidxkv "github.com/cometbft/cometbft/state/indexer/block/kv"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/types"
)

func TestIndexerServiceIndexesBlocks(t *testing.T) {
	_, txIndexer, blockIndexer, eventBus := createTestSetup(t)

	height := int64(1)

	events, txResult1, txResult2 := getEventsAndResults(height)
	// publish block with events
	err := eventBus.PublishEventNewBlockEvents(events)
	require.NoError(t, err)

	err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult1})
	require.NoError(t, err)

	err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult2})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	res, err := txIndexer.Get(types.Tx(fmt.Sprintf("foo%d", height)).Hash())
	require.NoError(t, err)
	require.Equal(t, txResult1, res)

	ok, err := blockIndexer.Has(height)
	require.NoError(t, err)
	require.True(t, ok)

	res, err = txIndexer.Get(types.Tx(fmt.Sprintf("bar%d", height)).Hash())
	require.NoError(t, err)
	require.Equal(t, txResult2, res)
}

func TestIndexerService_Prune(t *testing.T) {
	service, txIndexer, _, eventBus := createTestSetup(t)

	var keys [][][]byte

	for height := int64(1); height <= 4; height++ {
		events, txResult1, txResult2 := getEventsAndResults(height)
		//publish block with events
		err := eventBus.PublishEventNewBlockEvents(events)
		require.NoError(t, err)

		err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult1})
		require.NoError(t, err)

		err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult2})
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		keys = append(keys, kv.GetKeys(txIndexer))
	}

	service.Prune(2)

	metaKeys := [][]byte{
		txindex.LastBlockIndexerRetainHeightKey,
		txindex.LastTxIndexerRetainHeightKey,
		txindex.IndexerRetainHeightKey,
	}

	keysAfterPrune2 := sliceDiff(kv.GetKeys(txIndexer), metaKeys)
	require.True(t, equalSlices(keysAfterPrune2, sliceDiff(keys[3], keys[0])))

	err := service.SaveIndexerRetainHeight(int64(4))
	require.NoError(t, err)

	actual, err := service.GetIndexerRetainHeight()
	require.NoError(t, err)

	require.Equal(t, int64(4), actual)

	service.Prune(4)

	keysAfterPrune4 := sliceDiff(kv.GetKeys(txIndexer), metaKeys)
	require.Equal(t, keysAfterPrune4, sliceDiff(keys[3], keys[2]))

	events, txResult1, txResult2 := getEventsAndResults(1)
	//publish block with events
	err = eventBus.PublishEventNewBlockEvents(events)
	require.NoError(t, err)

	err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult1})
	require.NoError(t, err)

	err = eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult2})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	keys14 := kv.GetKeys(txIndexer)

	service.Prune(4)
	keysSecondPrune4 := kv.GetKeys(txIndexer)

	require.Equal(t, keys14, keysSecondPrune4)
}

func createTestSetup(t *testing.T) (*txindex.IndexerService, *kv.TxIndex, indexer.BlockIndexer, *types.EventBus) {
	// event bus
	eventBus := types.NewEventBus()
	eventBus.SetLogger(log.TestingLogger())
	err := eventBus.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := eventBus.Stop(); err != nil {
			t.Error(err)
		}
	})

	// tx indexer
	store := db.NewMemDB()
	txIndexer := kv.NewTxIndex(store)
	blockIndexer := blockidxkv.New(db.NewPrefixDB(store, []byte("block_events")))

	service := txindex.NewIndexerService(txIndexer, blockIndexer, store, eventBus, false)
	service.SetLogger(log.TestingLogger())
	err = service.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := service.Stop(); err != nil {
			t.Error(err)
		}
	})
	return service, txIndexer, blockIndexer, eventBus
}

func getEventsAndResults(height int64) (types.EventDataNewBlockEvents, *abci.TxResult, *abci.TxResult) {
	events := types.EventDataNewBlockEvents{
		Height: height,
		Events: []abci.Event{
			{
				Type: "begin_event",
				Attributes: []abci.EventAttribute{
					{
						Key:   "proposer",
						Value: "FCAA001",
						Index: true,
					},
				},
			},
		},
		NumTxs: int64(2),
	}
	txResult1 := &abci.TxResult{
		Height: height,
		Index:  uint32(0),
		Tx:     types.Tx(fmt.Sprintf("foo%d", height)),
		Result: abci.ExecTxResult{Code: 0},
	}
	txResult2 := &abci.TxResult{
		Height: height,
		Index:  uint32(1),
		Tx:     types.Tx(fmt.Sprintf("bar%d", height)),
		Result: abci.ExecTxResult{Code: 0},
	}
	return events, txResult1, txResult2
}

func contains(slice [][]byte, target []byte) bool {
	for _, element := range slice {
		if bytes.Equal(element, target) {
			return true
		}
	}
	return false
}

func subslice(smaller [][]byte, bigger [][]byte) bool {
	for _, elem := range smaller {
		if !contains(bigger, elem) {
			return false
		}
	}
	return true
}

func equalSlices(x [][]byte, y [][]byte) bool {
	return subslice(x, y) && subslice(y, x)
}

func sliceDiff(bigger [][]byte, smaller [][]byte) [][]byte {
	var diff [][]byte
	for _, elem := range bigger {
		if !contains(smaller, elem) {
			diff = append(diff, elem)
		}
	}
	return diff
}
