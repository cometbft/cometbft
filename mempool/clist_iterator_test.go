package mempool

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

func TestIteratorNonBlocking(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	// Add all txs with id up to n.
	n := 100
	for i := 0; i < n; i++ {
		tx := kvstore.NewTxFromID(i)
		rr, err := mp.CheckTx(tx, noSender)
		require.NoError(t, err)
		rr.Wait()
	}
	require.Equal(t, n, mp.Size())

	iter := mp.NewWRRIterator()
	expectedOrder := []int{
		0, 11, 22, 33, 44, 55, 66, // lane 7
		1, 2, 4, // lane 3
		3, // lane 1
		77, 88, 99,
		5, 7, 8,
		6,
		10, 13, 14,
		9,
		16, 17, 19,
		12,
		20, 23, 25,
		15,
	}

	var next Entry
	counter := 0

	// Check that txs are picked by the iterator in the expected order.
	for _, id := range expectedOrder {
		next = iter.Next()
		require.NotNil(t, next)
		require.Equal(t, types.Tx(kvstore.NewTxFromID(id)), next.Tx(), "id=%v", id)
		counter++
	}

	// Check that the rest of the entries are also consumed.
	for {
		if next = iter.Next(); next == nil {
			break
		}
		counter++
	}
	require.Equal(t, n, counter)
}

func TestIteratorNonBlockingOneLane(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	// Add all txs with id up to n to one lane.
	n := 100
	for i := 0; i < n; i++ {
		if i%11 != 0 {
			continue
		}
		tx := kvstore.NewTxFromID(i)
		rr, err := mp.CheckTx(tx, noSender)
		require.NoError(t, err)
		rr.Wait()
	}
	require.Equal(t, 10, mp.Size())

	iter := mp.NewWRRIterator()
	expectedOrder := []int{0, 11, 22, 33, 44, 55, 66, 77, 88, 99}

	var next Entry
	counter := 0

	// Check that txs are picked by the iterator in the expected order.
	for _, id := range expectedOrder {
		next = iter.Next()
		require.NotNil(t, next)
		require.Equal(t, types.Tx(kvstore.NewTxFromID(id)), next.Tx(), "id=%v", id)
		counter++
	}

	next = iter.Next()
	require.Nil(t, next)
}

func TestReapOrderMatchesGossipOrder(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	n := 10

	// Add a bunch of txs.
	for i := 1; i <= n; i++ {
		tx := kvstore.NewTxFromID(i)
		rr, err := mp.CheckTx(tx, "")
		require.NoError(t, err, err)
		rr.Wait()
	}
	require.Equal(t, n, mp.Size())

	gossipIter := mp.NewBlockingWRRIterator()
	reapIter := mp.NewWRRIterator()

	// Check that both iterators return the same entry as in the reaped txs.
	txs := make([]types.Tx, n)
	reapedTxs := mp.ReapMaxTxs(n)
	for i, reapedTx := range reapedTxs {
		entry := <-gossipIter.WaitNextCh()
		// entry can be nil only when an entry is removed concurrently.
		require.NotNil(t, entry)
		gossipTx := entry.Tx()

		reapTx := reapIter.Next().Tx()
		txs[i] = reapTx

		require.EqualValues(t, reapTx, gossipTx)
		require.EqualValues(t, reapTx, reapedTx)
	}
	require.EqualValues(t, txs, reapedTxs)

	err := mp.Update(1, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	require.Zero(t, mp.Size())
}
