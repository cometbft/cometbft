package kv

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/types"
)

func TestBigInt(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	bigInt := "10000000000000000000"
	bigIntPlus1 := "10000000000000000001"
	bigFloat := bigInt + ".76"
	bigFloatLower := bigInt + ".1"

	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: bigInt, Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: bigIntPlus1, Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: bigFloatLower, Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "owner", Value: "/Ivan/", Index: true}}},
		{Type: "", Attributes: []abci.EventAttribute{{Key: "not_allowed", Value: "Vlad", Index: true}}},
	})
	hash := types.Tx(txResult.Tx).Hash()

	err := indexer.Index(txResult)

	require.NoError(t, err)

	txResult2 := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: bigFloat, Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: bigFloat, Index: true}, {Key: "amount", Value: "5", Index: true}}},
	})

	txResult2.Tx = types.Tx("NEW TX")
	txResult2.Height = 2
	txResult2.Index = 2

	hash2 := types.Tx(txResult2.Tx).Hash()

	err = indexer.Index(txResult2)
	require.NoError(t, err)
	testCases := []struct {
		q             string
		txRes         *abci.TxResult
		resultsLength int
	}{
		//	search by hash
		{fmt.Sprintf("tx.hash = '%X'", hash), txResult, 1},
		// search by hash (lower)
		{fmt.Sprintf("tx.hash = '%x'", hash), txResult, 1},
		{fmt.Sprintf("tx.hash = '%x'", hash2), txResult2, 1},
		// search by exact match (one key) - bigint
		{"account.number >= " + bigInt, nil, 1},
		// search by exact match (one key) - bigint range
		{"account.number >= " + bigInt + " AND tx.height > 0", nil, 1},
		{"account.number >= " + bigInt + " AND tx.height > 0 AND account.owner = '/Ivan/'", nil, 0},
		// Floats are not parsed
		{"account.number >= " + bigInt + " AND tx.height > 0 AND account.amount > 4", txResult2, 0},
		{"account.number >= " + bigInt + " AND tx.height > 0 AND account.amount = 5", txResult2, 0},
		{"account.number >= " + bigInt + " AND account.amount <= 5", txResult2, 0},
		{"account.number < " + bigInt + " AND tx.height = 1", nil, 0},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.q, func(t *testing.T) {
			results, err := indexer.Search(ctx, query.MustParse(tc.q))
			assert.NoError(t, err)
			assert.Len(t, results, tc.resultsLength)
			if tc.resultsLength > 0 && tc.txRes != nil {
				assert.True(t, proto.Equal(results[0], tc.txRes))
			}
		})
	}
}

func TestTxIndex(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	tx := types.Tx("HELLO WORLD")
	txResult := &abci.TxResult{
		Height: 1,
		Index:  0,
		Tx:     tx,
		Result: abci.ResponseDeliverTx{
			Data: []byte{0},
			Code: abci.CodeTypeOK, Log: "", Events: nil,
		},
	}
	hash := tx.Hash()

	batch := txindex.NewBatch(1)
	if err := batch.Add(txResult); err != nil {
		t.Error(err)
	}
	err := indexer.AddBatch(batch)
	require.NoError(t, err)

	loadedTxResult, err := indexer.Get(hash)
	require.NoError(t, err)
	assert.True(t, proto.Equal(txResult, loadedTxResult))

	tx2 := types.Tx("BYE BYE WORLD")
	txResult2 := &abci.TxResult{
		Height: 1,
		Index:  0,
		Tx:     tx2,
		Result: abci.ResponseDeliverTx{
			Data: []byte{0},
			Code: abci.CodeTypeOK, Log: "", Events: nil,
		},
	}
	hash2 := tx2.Hash()

	err = indexer.Index(txResult2)
	require.NoError(t, err)

	loadedTxResult2, err := indexer.Get(hash2)
	require.NoError(t, err)
	assert.True(t, proto.Equal(txResult2, loadedTxResult2))
}

func TestTxSearch(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "owner", Value: "/Ivan/", Index: true}}},
		{Type: "", Attributes: []abci.EventAttribute{{Key: "not_allowed", Value: "Vlad", Index: true}}},
	})
	hash := types.Tx(txResult.Tx).Hash()

	err := indexer.Index(txResult)
	require.NoError(t, err)

	testCases := []struct {
		q             string
		resultsLength int
	}{
		//	search by hash
		{fmt.Sprintf("tx.hash = '%X'", hash), 1},
		// search by hash (lower)
		{fmt.Sprintf("tx.hash = '%x'", hash), 1},
		// search by exact match (one key)
		{"account.number = 1", 1},
		// search by exact match (two keys)
		{"account.number = 1 AND account.owner = 'Ivan'", 0},
		{"account.owner = 'Ivan' AND account.number = 1", 0},
		{"account.owner = '/Ivan/'", 1},
		// search by exact match (two keys)
		{"account.number = 1 AND account.owner = 'Vlad'", 0},
		{"account.owner = 'Vlad' AND account.number = 1", 0},
		{"account.number >= 1 AND account.owner = 'Vlad'", 0},
		{"account.owner = 'Vlad' AND account.number >= 1", 0},
		{"account.number <= 0", 0},
		{"account.number <= 0 AND account.owner = 'Ivan'", 0},
		{"account.number < 10000 AND account.owner = 'Ivan'", 0},
		// search using a prefix of the stored value
		{"account.owner = 'Iv'", 0},
		// search by range
		{"account.number >= 1 AND account.number <= 5", 1},
		// search by range and another key
		{"account.number >= 1 AND account.owner = 'Ivan' AND account.number <= 5", 0},
		// search by range (lower bound)
		{"account.number >= 1", 1},
		// search by range (upper bound)
		{"account.number <= 5", 1},
		{"account.number <= 1", 1},
		// search using not allowed key
		{"not_allowed = 'boom'", 0},
		{"not_allowed = 'Vlad'", 0},
		// search for not existing tx result
		{"account.number >= 2 AND account.number <= 5 AND tx.height > 0", 0},
		// search using not existing key
		{"account.date >= TIME 2013-05-03T14:45:00Z", 0},
		// search using CONTAINS
		{"account.owner CONTAINS 'an'", 1},
		//	search for non existing value using CONTAINS
		{"account.owner CONTAINS 'Vlad'", 0},
		{"account.owner CONTAINS 'Ivann'", 0},
		{"account.owner CONTAINS 'IIvan'", 0},
		{"account.owner CONTAINS 'Iva n'", 0},
		{"account.owner CONTAINS ' Ivan'", 0},
		{"account.owner CONTAINS 'Ivan '", 0},
		// search using the wrong key (of numeric type) using CONTAINS
		{"account.number CONTAINS 'Iv'", 0},
		// search using EXISTS
		{"account.number EXISTS", 1},
		// search using EXISTS for non existing key
		{"account.date EXISTS", 0},
		{"not_allowed EXISTS", 0},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.q, func(t *testing.T) {
			results, err := indexer.Search(ctx, query.MustParse(tc.q))
			assert.NoError(t, err)

			assert.Len(t, results, tc.resultsLength)
			if tc.resultsLength > 0 {
				for _, txr := range results {
					assert.True(t, proto.Equal(txResult, txr))
				}
			}
		})
	}
}

func TestTxSearchEventMatch(t *testing.T) {

	indexer := NewTxIndex(db.NewMemDB())

	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}, {Key: "owner", Value: "Ana", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "2", Index: true}, {Key: "owner", Value: "/Ivan/.test", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "3", Index: false}, {Key: "owner", Value: "Mickey", Index: false}}},
		{Type: "", Attributes: []abci.EventAttribute{{Key: "not_allowed", Value: "Vlad", Index: true}}},
	})

	err := indexer.Index(txResult)
	require.NoError(t, err)

	testCases := map[string]struct {
		q             string
		resultsLength int
	}{
		"Return all events from a height": {
			q:             "tx.height = 1",
			resultsLength: 1,
		},
		"Don't match non-indexed events": {
			q:             "account.number = 3 AND account.owner = 'Mickey'",
			resultsLength: 0,
		},
		"Return all events from a height with range": {
			q:             "tx.height > 0",
			resultsLength: 1,
		},
		"Return all events from a height with range 2": {
			q:             "tx.height <= 1",
			resultsLength: 1,
		},
		"Return all events from a height (deduplicate height)": {
			q:             "tx.height = 1 AND tx.height = 1",
			resultsLength: 1,
		},
		"Match attributes with height range and event": {
			q:             "tx.height < 2 AND tx.height > 0 AND account.number > 0 AND account.number <= 1 AND account.owner CONTAINS 'Ana'",
			resultsLength: 1,
		},
		"Match attributes with multiple CONTAIN and height range": {
			q:             "tx.height < 2 AND tx.height > 0 AND account.number = 1 AND account.owner CONTAINS 'Ana' AND account.owner CONTAINS 'An'",
			resultsLength: 1,
		},
		"Match attributes with height range and event - no match": {
			q:             "tx.height < 2 AND tx.height > 0 AND account.number = 2 AND account.owner = 'Ana'",
			resultsLength: 0,
		},
		"Match attributes with event": {
			q:             "account.number = 2 AND account.owner = 'Ana' AND tx.height = 1",
			resultsLength: 0,
		},
		"Deduplication test - should return nothing if attribute repeats multiple times": {
			q:             "tx.height < 2 AND account.number = 3 AND account.number = 2 AND account.number = 5",
			resultsLength: 0,
		},
		"Deduplication test - should return nothing if attribute repeats multiple times with match events": {
			q:             "tx.height < 2 AND account.number = 3 AND account.number = 2 AND account.number = 5",
			resultsLength: 0,
		},
		" Match range with match events": {
			q:             "account.number < 2 AND account.owner = '/Ivan/.test'",
			resultsLength: 0,
		},
		" Match range with match events 2": {
			q:             "account.number <= 2 AND account.owner = '/Ivan/.test' AND tx.height > 0",
			resultsLength: 1,
		},
		" Match range with match events contains with multiple items": {
			q:             "account.number <= 2 AND account.owner CONTAINS '/Iv' AND account.owner CONTAINS 'an' AND tx.height = 1",
			resultsLength: 1,
		},
		" Match range with match events contains": {
			q:             "account.number <= 2 AND account.owner CONTAINS 'an' AND tx.height > 0",
			resultsLength: 1,
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.q, func(t *testing.T) {
			results, err := indexer.Search(ctx, query.MustParse(tc.q))
			assert.NoError(t, err)

			assert.Len(t, results, tc.resultsLength)
			if tc.resultsLength > 0 {
				for _, txr := range results {
					assert.True(t, proto.Equal(txResult, txr))
				}
			}
		})
	}
}
func TestTxSearchWithCancelation(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "owner", Value: "Ivan", Index: true}}},
		{Type: "", Attributes: []abci.EventAttribute{{Key: "not_allowed", Value: "Vlad", Index: true}}},
	})
	err := indexer.Index(txResult)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := indexer.Search(ctx, query.MustParse("account.number = 1"))
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestTxSearchDeprecatedIndexing(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	// index tx using events indexing (composite key)
	txResult1 := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}}},
	})
	hash1 := types.Tx(txResult1.Tx).Hash()

	err := indexer.Index(txResult1)
	require.NoError(t, err)

	// index tx also using deprecated indexing (event as key)
	txResult2 := txResultWithEvents(nil)
	txResult2.Tx = types.Tx("HELLO WORLD 2")

	hash2 := types.Tx(txResult2.Tx).Hash()
	b := indexer.store.NewBatch()

	rawBytes, err := proto.Marshal(txResult2)
	require.NoError(t, err)

	depKey := []byte(fmt.Sprintf("%s/%s/%d/%d",
		"sender",
		"addr1",
		txResult2.Height,
		txResult2.Index,
	))

	err = b.Set(depKey, hash2)
	require.NoError(t, err)
	err = b.Set(keyForHeight(txResult2), hash2)
	require.NoError(t, err)
	err = b.Set(hash2, rawBytes)
	require.NoError(t, err)
	err = b.Write()
	require.NoError(t, err)

	testCases := []struct {
		q       string
		results []*abci.TxResult
	}{
		// search by hash
		{fmt.Sprintf("tx.hash = '%X'", hash1), []*abci.TxResult{txResult1}},
		// search by hash
		{fmt.Sprintf("tx.hash = '%X'", hash2), []*abci.TxResult{txResult2}},
		// search by exact match (one key)
		{"account.number = 1", []*abci.TxResult{txResult1}},
		{"account.number >= 1 AND account.number <= 5", []*abci.TxResult{txResult1}},
		// search by range (lower bound)
		{"account.number >= 1", []*abci.TxResult{txResult1}},
		// search by range (upper bound)
		{"account.number <= 5", []*abci.TxResult{txResult1}},
		// search using not allowed key
		{"not_allowed = 'boom'", []*abci.TxResult{}},
		// search for not existing tx result
		{"account.number >= 2 AND account.number <= 5", []*abci.TxResult{}},
		// search using not existing key
		{"account.date >= TIME 2013-05-03T14:45:00Z", []*abci.TxResult{}},
		// search by deprecated key
		{"sender = 'addr1'", []*abci.TxResult{txResult2}},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.q, func(t *testing.T) {
			results, err := indexer.Search(ctx, query.MustParse(tc.q))
			require.NoError(t, err)
			for _, txr := range results {
				for _, tr := range tc.results {
					assert.True(t, proto.Equal(tr, txr))
				}
			}
		})
	}
}

func TestTxSearchOneTxWithMultipleSameTagsButDifferentValues(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "2", Index: true}}},
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "3", Index: false}}},
	})

	err := indexer.Index(txResult)
	require.NoError(t, err)

	testCases := []struct {
		q     string
		found bool
	}{
		{
			q:     "account.number >= 1",
			found: true,
		},
		{
			q:     "account.number > 2",
			found: false,
		},
		{
			q:     "account.number >= 1 AND tx.height = 3 AND tx.height > 0",
			found: true,
		},
		{
			q:     "account.number >= 1 AND tx.height > 0 AND tx.height = 3",
			found: true,
		},

		{
			q:     "account.number >= 1 AND tx.height = 1  AND tx.height = 2 AND tx.height = 3",
			found: true,
		},

		{
			q:     "account.number >= 1 AND tx.height = 3  AND tx.height = 2 AND tx.height = 1",
			found: false,
		},
		{
			q:     "account.number >= 1 AND tx.height = 3",
			found: false,
		},
		{
			q:     "account.number > 1 AND tx.height < 2",
			found: true,
		},
		{
			q:     "account.number >= 2",
			found: true,
		},
		{
			q:     "account.number <= 1",
			found: true,
		},
		{
			q:     "account.number = 'something'",
			found: false,
		},
		{
			q:     "account.number CONTAINS 'bla'",
			found: false,
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		results, err := indexer.Search(ctx, query.MustParse(tc.q))
		assert.NoError(t, err)
		len := 0
		if tc.found {
			len = 1
		}
		assert.Len(t, results, len)
		assert.True(t, !tc.found || proto.Equal(txResult, results[0]))

	}
}

func TestTxIndexDuplicatePreviouslySuccessful(t *testing.T) {
	var mockTx = types.Tx("MOCK_TX_HASH")

	testCases := []struct {
		name         string
		tx1          *abci.TxResult
		tx2          *abci.TxResult
		expOverwrite bool // do we expect the second tx to overwrite the first tx
	}{
		{
			"don't overwrite as a non-zero code was returned and the previous tx was successful",
			&abci.TxResult{
				Height: 1,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK,
				},
			},
			&abci.TxResult{
				Height: 2,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK + 1,
				},
			},
			false,
		},
		{
			"overwrite as the previous tx was also unsuccessful",
			&abci.TxResult{
				Height: 1,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK + 1,
				},
			},
			&abci.TxResult{
				Height: 2,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK + 1,
				},
			},
			true,
		},
		{
			"overwrite as the most recent tx was successful",
			&abci.TxResult{
				Height: 1,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK,
				},
			},
			&abci.TxResult{
				Height: 2,
				Index:  0,
				Tx:     mockTx,
				Result: abci.ResponseDeliverTx{
					Code: abci.CodeTypeOK,
				},
			},
			true,
		},
	}

	hash := mockTx.Hash()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			indexer := NewTxIndex(db.NewMemDB())

			// index the first tx
			err := indexer.Index(tc.tx1)
			require.NoError(t, err)

			// index the same tx with different results
			err = indexer.Index(tc.tx2)
			require.NoError(t, err)

			res, err := indexer.Get(hash)
			require.NoError(t, err)

			if tc.expOverwrite {
				require.Equal(t, tc.tx2, res)
			} else {
				require.Equal(t, tc.tx1, res)
			}
		})
	}
}

func TestTxSearchMultipleTxs(t *testing.T) {
	indexer := NewTxIndex(db.NewMemDB())

	// indexed first, but bigger height (to test the order of transactions)
	txResult := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "1", Index: true}}},
	})

	txResult.Tx = types.Tx("Bob's account")
	txResult.Height = 2
	txResult.Index = 1
	err := indexer.Index(txResult)
	require.NoError(t, err)

	// indexed second, but smaller height (to test the order of transactions)
	txResult2 := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "2", Index: true}}},
	})
	txResult2.Tx = types.Tx("Alice's account")
	txResult2.Height = 1
	txResult2.Index = 2

	err = indexer.Index(txResult2)
	require.NoError(t, err)

	// indexed third (to test the order of transactions)
	txResult3 := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number", Value: "3", Index: true}}},
	})
	txResult3.Tx = types.Tx("Jack's account")
	txResult3.Height = 1
	txResult3.Index = 1
	err = indexer.Index(txResult3)
	require.NoError(t, err)

	// indexed fourth (to test we don't include txs with similar events)
	// https://github.com/tendermint/tendermint/issues/2908
	txResult4 := txResultWithEvents([]abci.Event{
		{Type: "account", Attributes: []abci.EventAttribute{{Key: "number.id", Value: "1", Index: true}}},
	})
	txResult4.Tx = types.Tx("Mike's account")
	txResult4.Height = 2
	txResult4.Index = 2
	err = indexer.Index(txResult4)
	require.NoError(t, err)

	ctx := context.Background()

	results, err := indexer.Search(ctx, query.MustParse("account.number >= 1"))
	assert.NoError(t, err)

	require.Len(t, results, 3)
}

func txResultWithEvents(events []abci.Event) *abci.TxResult {
	tx := types.Tx("HELLO WORLD")
	return &abci.TxResult{
		Height: 1,
		Index:  0,
		Tx:     tx,
		Result: abci.ResponseDeliverTx{
			Data:   []byte{0},
			Code:   abci.CodeTypeOK,
			Log:    "",
			Events: events,
		},
	}
}

func benchmarkTxIndex(txsCount int64, b *testing.B) {
	dir, err := os.MkdirTemp("", "tx_index_db")
	require.NoError(b, err)
	defer os.RemoveAll(dir)

	store, err := db.NewDB("tx_index", "goleveldb", dir)
	require.NoError(b, err)
	indexer := NewTxIndex(store)

	batch := txindex.NewBatch(txsCount)
	txIndex := uint32(0)
	for i := int64(0); i < txsCount; i++ {
		tx := cmtrand.Bytes(250)
		txResult := &abci.TxResult{
			Height: 1,
			Index:  txIndex,
			Tx:     tx,
			Result: abci.ResponseDeliverTx{
				Data:   []byte{0},
				Code:   abci.CodeTypeOK,
				Log:    "",
				Events: []abci.Event{},
			},
		}
		if err := batch.Add(txResult); err != nil {
			b.Fatal(err)
		}
		txIndex++
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		err = indexer.AddBatch(batch)
	}
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkTxIndex1(b *testing.B)     { benchmarkTxIndex(1, b) }
func BenchmarkTxIndex500(b *testing.B)   { benchmarkTxIndex(500, b) }
func BenchmarkTxIndex1000(b *testing.B)  { benchmarkTxIndex(1000, b) }
func BenchmarkTxIndex2000(b *testing.B)  { benchmarkTxIndex(2000, b) }
func BenchmarkTxIndex10000(b *testing.B) { benchmarkTxIndex(10000, b) }
