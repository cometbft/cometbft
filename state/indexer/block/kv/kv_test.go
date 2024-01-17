package kv_test

import (
	"context"
	"fmt"
	"testing"

	db "github.com/cometbft/cometbft-db"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	blockidxkv "github.com/tendermint/tendermint/state/indexer/block/kv"
	"github.com/tendermint/tendermint/types"
)

func TestBlockIndexer(t *testing.T) {
	store := db.NewPrefixDB(db.NewMemDB(), []byte("block_events"))
	indexer := blockidxkv.New(store)

	require.NoError(t, indexer.Index(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 1},
		ResultBeginBlock: abci.ResponseBeginBlock{
			Events: []abci.Event{
				{
					Type: "begin_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("proposer"),
							Value: []byte("FCAA001"),
							Index: true,
						},
					},
				},
			},
		},
		ResultEndBlock: abci.ResponseEndBlock{
			Events: []abci.Event{
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("100"),
							Index: true,
						},
					},
				},
			},
		},
	}))

	for i := 2; i < 12; i++ {
		var index bool
		if i%2 == 0 {
			index = true
		}

		require.NoError(t, indexer.Index(types.EventDataNewBlockHeader{
			Header: types.Header{Height: int64(i)},
			ResultBeginBlock: abci.ResponseBeginBlock{
				Events: []abci.Event{
					{
						Type: "begin_event",
						Attributes: []abci.EventAttribute{
							{
								Key:   []byte("proposer"),
								Value: []byte("FCAA001"),
								Index: true,
							},
						},
					},
				},
			},
			ResultEndBlock: abci.ResponseEndBlock{
				Events: []abci.Event{
					{
						Type: "end_event",
						Attributes: []abci.EventAttribute{
							{
								Key:   []byte("foo"),
								Value: []byte(fmt.Sprintf("%d", i)),
								Index: index,
							},
						},
					},
				},
			},
		}))
	}

	testCases := map[string]struct {
		q       *query.Query
		results []int64
	}{
		"block.height = 100": {
			q:       query.MustParse("block.height = 100"),
			results: []int64{},
		},
		"block.height = 5": {
			q:       query.MustParse("block.height = 5"),
			results: []int64{5},
		},
		"begin_event.key1 = 'value1'": {
			q:       query.MustParse("begin_event.key1 = 'value1'"),
			results: []int64{},
		},
		"begin_event.proposer = 'FCAA001'": {
			q:       query.MustParse("begin_event.proposer = 'FCAA001'"),
			results: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		"end_event.foo <= 5": {
			q:       query.MustParse("end_event.foo <= 5"),
			results: []int64{2, 4},
		},
		"end_event.foo >= 100": {
			q:       query.MustParse("end_event.foo >= 100"),
			results: []int64{1},
		},
		"end_event.foo > 100": {
			q:       query.MustParse("end_event.foo > 100"),
			results: []int64{},
		},
		"block.height > 2 AND end_event.foo <= 8": {
			q:       query.MustParse("block.height > 2 AND end_event.foo <= 8"),
			results: []int64{4, 6, 8},
		},
		"block.height >= 2 AND end_event.foo < 8": {
			q:       query.MustParse("block.height >= 2 AND end_event.foo < 8"),
			results: []int64{2, 4, 6},
		},
		"begin_event.proposer CONTAINS 'FFFFFFF'": {
			q:       query.MustParse("begin_event.proposer CONTAINS 'FFFFFFF'"),
			results: []int64{},
		},
		"begin_event.proposer CONTAINS 'FCAA001'": {
			q:       query.MustParse("begin_event.proposer CONTAINS 'FCAA001'"),
			results: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
		"end_event.foo CONTAINS '1'": {
			q:       query.MustParse("end_event.foo CONTAINS '1'"),
			results: []int64{1, 10},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			results, err := indexer.Search(context.Background(), tc.q)
			require.NoError(t, err)
			require.Equal(t, tc.results, results)
		})
	}
}

func TestBlockIndexerMulti(t *testing.T) {
	store := db.NewPrefixDB(db.NewMemDB(), []byte("block_events"))
	indexer := blockidxkv.New(store)

	require.NoError(t, indexer.Index(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 1},
		ResultBeginBlock: abci.ResponseBeginBlock{
			Events: []abci.Event{},
		},
		ResultEndBlock: abci.ResponseEndBlock{
			Events: []abci.Event{
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("100"),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("200"),
							Index: true,
						},
					},
				},
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("300"),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("500"),
							Index: true,
						},
					},
				},
			},
		},
	}))

	require.NoError(t, indexer.Index(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 2},
		ResultBeginBlock: abci.ResponseBeginBlock{
			Events: []abci.Event{},
		},
		ResultEndBlock: abci.ResponseEndBlock{
			Events: []abci.Event{
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("100"),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("200"),
							Index: true,
						},
					},
				},
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("300"),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("400"),
							Index: true,
						},
					},
				},
			},
		},
	}))

	testCases := map[string]struct {
		q       *query.Query
		results []int64
	}{
		"query return all events from a height - exact": {
			q:       query.MustParse("match.events = 1 AND block.height = 1"),
			results: []int64{1},
		},
		"query return all events from a height - exact - no match.events": {
			q:       query.MustParse("block.height = 1"),
			results: []int64{1},
		},
		"query return all events from a height - exact (deduplicate height)": {
			q:       query.MustParse("match.events = 1 AND block.height = 1 AND block.height = 2"),
			results: []int64{1},
		},
		"query return all events from a height - exact (deduplicate height) - no match.events": {
			q:       query.MustParse("block.height = 1 AND block.height = 2"),
			results: []int64{1},
		},
		"query return all events from a height - range": {
			q:       query.MustParse("match.events = 1 AND block.height < 2 AND block.height > 0 AND block.height > 0"),
			results: []int64{1},
		},
		"query return all events from a height - range - no match.events": {
			q:       query.MustParse("block.height < 2 AND block.height > 0 AND block.height > 0"),
			results: []int64{1},
		},
		"query return all events from a height - range 2": {
			q:       query.MustParse("match.events = 1 AND block.height < 3 AND block.height < 2 AND block.height > 0 AND block.height > 0"),
			results: []int64{1},
		},
		"query return all events from a height - range 3": {
			q:       query.MustParse("match.events = 1 AND block.height < 1 AND block.height > 1"),
			results: []int64{},
		},
		"query matches fields from same event": {
			q:       query.MustParse("match.events = 1 AND end_event.bar < 300 AND end_event.foo = 100 AND block.height > 0 AND block.height <= 2"),
			results: []int64{1, 2},
		},
		"query matches fields from same event - no match.events": {
			q:       query.MustParse("end_event.bar < 300 AND end_event.foo = 100 AND block.height > 0 AND block.height <= 2"),
			results: []int64{1, 2},
		},
		"query matches fields from multiple events": {
			q:       query.MustParse("match.events = 1 AND end_event.foo = 100 AND end_event.bar = 400 AND block.height = 2"),
			results: []int64{},
		},
		"query matches fields from multiple events 2": {
			q:       query.MustParse("match.events = 1 AND end_event.foo = 100 AND end_event.bar > 200 AND block.height > 0 AND block.height < 3"),
			results: []int64{},
		},
		"query matches fields from multiple events 2 - match.events set to 0": {
			q:       query.MustParse("match.events = 0 AND end_event.foo = 100 AND end_event.bar > 200 AND block.height > 0 AND block.height < 3"),
			results: []int64{1, 2},
		},
		"deduplication test - match.events only at beginning": {
			q:       query.MustParse("end_event.foo = 100 AND end_event.bar = 400 AND block.height = 2 AND match.events = 1"),
			results: []int64{2},
		},
		"deduplication test - match.events only at beginning 2": {
			q:       query.MustParse("end_event.foo = 100 AND match.events = 1 AND end_event.bar = 400 AND block.height = 2"),
			results: []int64{2},
		},
		"deduplication test - match.events multiple": {
			q:       query.MustParse("match.events = 1 AND end_event.foo = 100 AND end_event.bar = 400 AND block.height = 2 AND match.events = 1"),
			results: []int64{},
		},
		"deduplication test - match.events multiple 2": {
			q:       query.MustParse("match.events = 1 AND end_event.foo = 100 AND match.events = 1 AND end_event.bar = 400 AND block.height = 2"),
			results: []int64{},
		},
		"query matches fields from multiple events allowed": {
			q:       query.MustParse("end_event.foo = 100 AND end_event.bar = 400"),
			results: []int64{2},
		},
		"query matches all fields from multiple events": {
			q:       query.MustParse("match.events = 1 AND end_event.bar > 100 AND end_event.bar <= 500"),
			results: []int64{1, 2},
		},
		"query matches all fields from multiple events - no match.events": {
			q:       query.MustParse("end_event.bar > 100 AND end_event.bar <= 500"),
			results: []int64{1, 2},
		},
		"query matches fields from all events whose attribute is within range": {
			q:       query.MustParse("match.events = 1 AND block.height  = 2 AND end_event.foo < 300"),
			results: []int64{2},
		},
		"query using CONTAINS matches fields from all events whose attribute is within range": {
			q:       query.MustParse("match.events = 1 AND block.height  = 2 AND end_event.foo CONTAINS '30'"),
			results: []int64{2},
		},
		"query with height range and height equality - should ignore equality": {
			q:       query.MustParse("match.events = 1 AND block.height = 2 AND end_event.foo >= 100 AND block.height < 2"),
			results: []int64{1},
		},
		"query with non-existent field": {
			q:       query.MustParse("match.events = 1 AND end_event.baz = 100"),
			results: []int64{},
		},
		"query with non-existent field - no match.events": {
			q:       query.MustParse("end_event.baz = 100"),
			results: []int64{},
		},
		"query with non-existent type": {
			q:       query.MustParse("match.events = 1 AND end_event_xyz.foo = 100"),
			results: []int64{},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			results, err := indexer.Search(context.Background(), tc.q)
			require.NoError(t, err)
			require.Equal(t, tc.results, results)
		})
	}
}

func TestBigInt(t *testing.T) {

	bigInt := "10000000000000000000"
	store := db.NewPrefixDB(db.NewMemDB(), []byte("block_events"))
	indexer := blockidxkv.New(store)

	require.NoError(t, indexer.Index(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 1},
		ResultBeginBlock: abci.ResponseBeginBlock{
			Events: []abci.Event{},
		},
		ResultEndBlock: abci.ResponseEndBlock{
			Events: []abci.Event{
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte("100"),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("10000000000000000000.76"),
							Index: true,
						},
						{
							Key:   []byte("bar_lower"),
							Value: []byte("10000000000000000000.1"),
							Index: true,
						},
					},
				},
				{
					Type: "end_event",
					Attributes: []abci.EventAttribute{
						{
							Key:   []byte("foo"),
							Value: []byte(bigInt),
							Index: true,
						},
						{
							Key:   []byte("bar"),
							Value: []byte("500"),
							Index: true,
						},
						{
							Key:   []byte("bla"),
							Value: []byte("500.5"),
							Index: true,
						},
					},
				},
			},
		},
	}))

	testCases := map[string]struct {
		q       *query.Query
		results []int64
	}{

		"query return all events from a height - exact": {
			q:       query.MustParse("block.height = 1"),
			results: []int64{1},
		},
		"query return all events from a height - exact (deduplicate height)": {
			q:       query.MustParse("block.height = 1 AND block.height = 2"),
			results: []int64{1},
		},
		"query return all events from a height - range": {
			q:       query.MustParse("block.height < 2 AND block.height > 0 AND block.height > 0"),
			results: []int64{1},
		},
		"query matches fields with big int and height - no match": {
			q:       query.MustParse("end_event.foo = " + bigInt + " AND end_event.bar = 500 AND block.height = 2"),
			results: []int64{},
		},
		"query matches fields with big int with less and height - no match": {
			q:       query.MustParse("end_event.foo <= " + bigInt + " AND end_event.bar = 500 AND block.height = 2"),
			results: []int64{},
		},
		"query matches fields with big int and height - match": {
			q:       query.MustParse("end_event.foo = " + bigInt + " AND end_event.bar = 500 AND block.height = 1"),
			results: []int64{1},
		},
		"query matches big int in range": {
			q:       query.MustParse("end_event.foo = " + bigInt),
			results: []int64{1},
		},
		"query matches big int in range with float - does not pass as float is not converted to int": {
			q:       query.MustParse("end_event.bar >= " + bigInt),
			results: []int64{},
		},
		"query matches big int in range with float - fails because float is converted to int": {
			q:       query.MustParse("end_event.bar > " + bigInt),
			results: []int64{},
		},
		"query matches big int in range with float lower dec point - fails because float is converted to int": {
			q:       query.MustParse("end_event.bar_lower > " + bigInt),
			results: []int64{},
		},
		"query matches big int in range with float with less - found": {
			q:       query.MustParse("end_event.foo <= " + bigInt),
			results: []int64{1},
		},
		"query matches big int in range with float with less with height range - found": {
			q:       query.MustParse("end_event.foo <= " + bigInt + " AND block.height > 0"),
			results: []int64{1},
		},
		"query matches big int in range with float with less - not found": {
			q:       query.MustParse("end_event.foo < " + bigInt + " AND end_event.foo > 100"),
			results: []int64{},
		},
		"query does not parse float": {
			q:       query.MustParse("end_event.bla >= 500"),
			results: []int64{},
		},
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			results, err := indexer.Search(context.Background(), tc.q)
			require.NoError(t, err)
			require.Equal(t, tc.results, results)
		})
	}
}
