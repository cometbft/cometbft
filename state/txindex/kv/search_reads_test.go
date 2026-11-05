package kv

import (
	"context"
	"fmt"
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/types"
)

type searchReadDB struct {
	dbm.DB
	reads     int
	readBytes int
	afterRead func()
}

func (db *searchReadDB) Get(key []byte) ([]byte, error) {
	value, err := db.DB.Get(key)
	db.reads++
	db.readBytes += len(value)
	if db.afterRead != nil {
		db.afterRead()
	}
	return value, err
}

func repeatedEventIndex(tb testing.TB, backend string, txCount, eventCount int) (*TxIndex, *searchReadDB, []*abci.TxResult) {
	tb.Helper()
	var store dbm.DB
	if backend == "goleveldb" {
		var err error
		store, err = dbm.NewGoLevelDB("search", tb.TempDir())
		require.NoError(tb, err)
	} else {
		store = dbm.NewMemDB()
	}
	tb.Cleanup(func() { require.NoError(tb, store.Close()) })
	counted := &searchReadDB{DB: store}
	indexer := NewTxIndex(counted)
	batch := txindex.NewBatch(int64(txCount))
	results := make([]*abci.TxResult, txCount)
	for i := range results {
		events := make([]abci.Event, eventCount)
		for j := range events {
			events[j] = abci.Event{Type: "transfer", Attributes: []abci.EventAttribute{
				{Key: "recipient", Value: "alice", Index: true},
				{Key: "sender", Value: "bob", Index: true},
				{Key: "amount", Value: "10", Index: true},
			}}
		}
		results[i] = &abci.TxResult{
			Height: 1, Index: uint32(i), Tx: []byte(fmt.Sprintf("batched-transfer-%d", i)),
			Result: abci.ExecTxResult{Events: events},
		}
		require.NoError(tb, batch.Add(results[i]))
	}
	require.NoError(tb, indexer.AddBatch(batch))
	counted.reads, counted.readBytes = 0, 0
	return indexer, counted, results
}

func TestTxSearchReadsEachResultOnce(t *testing.T) {
	queries := []string{
		"transfer.recipient = 'alice'",
		"transfer.recipient EXISTS",
		"transfer.recipient CONTAINS 'lic'",
		"transfer.amount >= 10 AND transfer.amount <= 10",
		"transfer.recipient = 'alice' AND transfer.sender = 'bob'",
		"tx.height = 1",
	}
	for _, backend := range []string{"memdb", "goleveldb"} {
		for _, n := range []int{1, 10, 100} {
			t.Run(fmt.Sprintf("%s/N=%d", backend, n), func(t *testing.T) {
				indexer, counted, want := repeatedEventIndex(t, backend, 2, n)
				for _, text := range queries {
					t.Run(text, func(t *testing.T) {
						counted.reads = 0
						got, err := indexer.Search(context.Background(), query.MustCompile(text))
						require.NoError(t, err)
						require.ElementsMatch(t, want, got)
						require.Equal(t, len(want), counted.reads, "one result read per distinct transaction")
					})
				}
			})
		}
	}
}

func TestTxSearchRepeatedEventsCancellation(t *testing.T) {
	for _, cancelBefore := range []bool{true, false} {
		t.Run(fmt.Sprintf("before=%t", cancelBefore), func(t *testing.T) {
			indexer, counted, want := repeatedEventIndex(t, "memdb", 2, 100)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if cancelBefore {
				cancel()
			} else {
				counted.afterRead = cancel
			}
			got, err := indexer.Search(ctx, query.MustCompile("transfer.recipient = 'alice'"))
			require.NoError(t, err)
			if cancelBefore {
				require.Empty(t, got)
				require.Zero(t, counted.reads)
			} else {
				require.Len(t, got, 1)
				require.True(t, proto.Equal(want[got[0].Index], got[0]))
				require.Equal(t, 1, counted.reads)
			}
		})
	}
}

func TestTxSearchRepeatedEventsDecodeError(t *testing.T) {
	indexer, counted, want := repeatedEventIndex(t, "memdb", 1, 100)
	require.NoError(t, counted.Set(types.Tx(want[0].Tx).Hash(), []byte{0xff}))
	got, err := indexer.Search(context.Background(), query.MustCompile("transfer.recipient = 'alice'"))
	require.ErrorContains(t, err, "failed to get Tx")
	require.Nil(t, got)
	require.Equal(t, 1, counted.reads)
}

func BenchmarkTxSearchRepeatedEvents(b *testing.B) {
	for _, backend := range []string{"memdb", "goleveldb"} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%s/N=%d", backend, n), func(b *testing.B) {
				benchmarkTxSearchEvents(b, backend, 1, n)
			})
		}
	}
}

func BenchmarkTxSearchDistinctTransactions(b *testing.B) {
	for _, backend := range []string{"memdb", "goleveldb"} {
		for _, n := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%s/N=%d", backend, n), func(b *testing.B) {
				benchmarkTxSearchEvents(b, backend, n, 1)
			})
		}
	}
}

func benchmarkTxSearchEvents(b *testing.B, backend string, txCount, eventCount int) {
	indexer, counted, _ := repeatedEventIndex(b, backend, txCount, eventCount)
	q := query.MustCompile("transfer.recipient = 'alice'")
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := indexer.Search(ctx, q)
		if err != nil || len(results) != txCount {
			b.Fatalf("unexpected result: %v, %v", results, err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(counted.reads)/float64(b.N), "gets/op")
	b.ReportMetric(float64(counted.readBytes)/float64(b.N), "decoded-bytes/op")
}
