package state_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	db "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/indexer"
	blockidxkv "github.com/cometbft/cometbft/state/indexer/block/kv"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestPruneIndexesToRetainHeight(t *testing.T) {
	pruner, txIndexer, _, eventBus := createTestSetup(t)

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

	pruner.PruneIndexesToRetainHeight(2)

	metaKeys := [][]byte{
		kv.LastTxIndexerRetainHeightKey,
		blockidxkv.LastBlockIndexerRetainHeightKey,
		sm.IndexerRetainHeightKey,
	}

	keysAfterPrune2 := setDiff(kv.GetKeys(txIndexer), metaKeys)
	require.True(t, isEqualSets(keysAfterPrune2, setDiff(keys[3], keys[0])))

	err := pruner.SetIndexerRetainHeight(int64(4))
	require.NoError(t, err)

	actual, err := pruner.GetIndexerRetainHeight()
	require.NoError(t, err)

	require.Equal(t, int64(4), actual)

	pruner.PruneIndexesToRetainHeight(4)

	keysAfterPrune4 := setDiff(kv.GetKeys(txIndexer), metaKeys)
	require.Equal(t, keysAfterPrune4, setDiff(keys[3], keys[2]))

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

	pruner.PruneIndexesToRetainHeight(4)
	keysSecondPrune4 := kv.GetKeys(txIndexer)

	require.Equal(t, keys14, keysSecondPrune4)
}

func createTestSetup(t *testing.T) (*sm.Pruner, *kv.TxIndex, indexer.BlockIndexer, *types.EventBus) {
	config := test.ResetTestRoot("pruner_test")
	// event bus
	eventBus := types.NewEventBus()
	eventBus.SetLogger(log.TestingLogger())
	err := eventBus.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := eventBus.Stop(); err != nil {
			t.Error(err)
		}
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			t.Error(err)
		}
	})

	// tx indexer
	memDB := db.NewMemDB()
	txIndexer := kv.NewTxIndex(memDB)
	blockIndexer := blockidxkv.New(db.NewPrefixDB(memDB, []byte("block_events")))

	blockDB := db.NewMemDB()
	stateDB := db.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	bs := store.NewBlockStore(blockDB)
	pruner := sm.NewPruner(stateStore, bs, blockIndexer, txIndexer, log.TestingLogger())

	return pruner, txIndexer, blockIndexer, eventBus
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

func isSubset(smaller [][]byte, bigger [][]byte) bool {
	for _, elem := range smaller {
		if !slices.ContainsFunc(bigger, func(i []byte) bool {
			return bytes.Equal(i, elem)
		}) {
			return false
		}
	}
	return true
}

func isEqualSets(x [][]byte, y [][]byte) bool {
	return isSubset(x, y) && isSubset(y, x)
}

func setDiff(bigger [][]byte, smaller [][]byte) [][]byte {
	var diff [][]byte
	for _, elem := range bigger {
		if !slices.ContainsFunc(smaller, func(i []byte) bool {
			return bytes.Equal(i, elem)
		}) {
			diff = append(diff, elem)
		}
	}
	return diff
}
