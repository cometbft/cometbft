package state_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	db "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	sm "github.com/cometbft/cometbft/state"
	blockidxkv "github.com/cometbft/cometbft/state/indexer/block/kv"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestPruneBlockIndexerToRetainHeight(t *testing.T) {
	pruner, _, blockIndexer, _ := createTestSetup(t)

	var keys [][][]byte

	for height := int64(1); height <= 4; height++ {
		events, _, _ := getEventsAndResults(height)
		err := blockIndexer.Index(events)
		require.NoError(t, err)
		keys = append(keys, blockidxkv.GetKeys(blockIndexer))
	}
	err := pruner.SetBlockIndexerRetainHeight(2)
	require.NoError(t, err)
	actual, err := pruner.GetBlockIndexerRetainHeight()
	require.NoError(t, err)
	require.Equal(t, int64(2), actual)

	newRetainHeight := pruner.PruneBlockIndexerToRetainHeight(0)
	require.Equal(t, int64(2), newRetainHeight)

	metaKeys := [][]byte{
		kv.LastTxIndexerRetainHeightKey,
		blockidxkv.LastBlockIndexerRetainHeightKey,
		kv.TxIndexerRetainHeightKey,
		blockidxkv.BlockIndexerRetainHeightKey,
	}

	keysAfterPrune2 := setDiff(blockidxkv.GetKeys(blockIndexer), metaKeys)
	require.True(t, isEqualSets(keysAfterPrune2, setDiff(keys[3], keys[0])))

	err = pruner.SetBlockIndexerRetainHeight(int64(4))
	require.NoError(t, err)
	actual, err = pruner.GetBlockIndexerRetainHeight()
	require.NoError(t, err)
	require.Equal(t, int64(4), actual)

	pruner.PruneBlockIndexerToRetainHeight(2)

	keysAfterPrune4 := setDiff(blockidxkv.GetKeys(blockIndexer), metaKeys)
	require.Equal(t, keysAfterPrune4, setDiff(keys[3], keys[2]))

	events, _, _ := getEventsAndResults(1)

	err = blockIndexer.Index(events)
	require.NoError(t, err)

	keys14 := blockidxkv.GetKeys(blockIndexer)

	pruner.PruneBlockIndexerToRetainHeight(4)
	keysSecondPrune4 := blockidxkv.GetKeys(blockIndexer)

	require.Equal(t, keys14, keysSecondPrune4)
}

func TestPruneTxIndexerToRetainHeight(t *testing.T) {
	pruner, txIndexer, _, _ := createTestSetup(t)

	var keys [][][]byte

	for height := int64(1); height <= 4; height++ {
		_, txResult1, txResult2 := getEventsAndResults(height)
		err := txIndexer.Index(txResult1)
		require.NoError(t, err)
		err = txIndexer.Index(txResult2)
		require.NoError(t, err)
		keys = append(keys, kv.GetKeys(txIndexer))
	}

	err := pruner.SetTxIndexerRetainHeight(2)
	require.NoError(t, err)
	actual, err := pruner.GetTxIndexerRetainHeight()
	require.NoError(t, err)
	require.Equal(t, int64(2), actual)

	newRetainHeight := pruner.PruneTxIndexerToRetainHeight(0)
	require.Equal(t, int64(2), newRetainHeight)

	metaKeys := [][]byte{
		kv.LastTxIndexerRetainHeightKey,
		blockidxkv.LastBlockIndexerRetainHeightKey,
		kv.TxIndexerRetainHeightKey,
		blockidxkv.BlockIndexerRetainHeightKey,
	}

	keysAfterPrune2 := setDiff(kv.GetKeys(txIndexer), metaKeys)
	require.True(t, isEqualSets(keysAfterPrune2, setDiff(keys[3], keys[0])))

	err = pruner.SetTxIndexerRetainHeight(int64(4))
	require.NoError(t, err)

	actual, err = pruner.GetTxIndexerRetainHeight()
	require.NoError(t, err)

	require.Equal(t, int64(4), actual)

	pruner.PruneTxIndexerToRetainHeight(2)

	keysAfterPrune4 := setDiff(kv.GetKeys(txIndexer), metaKeys)
	require.Equal(t, keysAfterPrune4, setDiff(keys[3], keys[2]))

	_, txResult1, txResult2 := getEventsAndResults(1)

	err = txIndexer.Index(txResult1)
	require.NoError(t, err)
	err = txIndexer.Index(txResult2)
	require.NoError(t, err)

	keys14 := kv.GetKeys(txIndexer)

	pruner.PruneTxIndexerToRetainHeight(4)
	keysSecondPrune4 := kv.GetKeys(txIndexer)

	require.Equal(t, keys14, keysSecondPrune4)
}

func createTestSetup(t *testing.T) (*sm.Pruner, *kv.TxIndex, blockidxkv.BlockerIndexer, *types.EventBus) {
	config := test.ResetTestRoot("pruner_test")
	t.Cleanup(func() {
		err := os.RemoveAll(config.RootDir)
		if err != nil {
			t.Error(err)
		}
	})
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

	return pruner, txIndexer, *blockIndexer, eventBus
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
