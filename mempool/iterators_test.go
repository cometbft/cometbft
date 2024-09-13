package mempool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	abciclimocks "github.com/cometbft/cometbft/abci/client/mocks"
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

	iter := NewNonBlockingIterator(mp)
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

	iter := NewNonBlockingIterator(mp)
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

// We have two iterators fetching transactions that
// then get removed.
func TestIteratorRace(t *testing.T) {
	mockClient := new(abciclimocks.Client)
	mockClient.On("Start").Return(nil)
	mockClient.On("SetLogger", mock.Anything)
	mockClient.On("Error").Return(nil).Times(100)
	mockClient.On("Info", mock.Anything, mock.Anything).Return(&abci.InfoResponse{LanePriorities: []uint32{1, 2, 3}, DefaultLanePriority: 1}, nil)

	mp, cleanup := newMempoolWithAppMock(mockClient)
	defer cleanup()

	// Disable rechecking to make sure the recheck logic is not interferint.
	mp.config.Recheck = false

	const numLanes = 3
	const numTxs = 100

	var wg sync.WaitGroup
	wg.Add(2)

	var counter atomic.Int64
	go func() {
		waitForNumTxsInMempool(numTxs, mp)

		go func() {
			defer wg.Done()

			for counter.Load() < int64(numTxs) {
				iter := NewBlockingIterator(mp)
				entry := <-iter.WaitNextCh()
				if entry == nil {
					continue
				}
				tx := entry.Tx()
				err := mp.Update(1, []types.Tx{tx}, abciResponses(1, 0), nil, nil)
				require.NoError(t, err, tx)
				counter.Add(1)
			}
		}()

		go func() {
			defer wg.Done()

			for counter.Load() < int64(numTxs) {
				iter := NewBlockingIterator(mp)
				entry := <-iter.WaitNextCh()
				if entry == nil {
					continue
				}
				tx := entry.Tx()
				err := mp.Update(1, []types.Tx{tx}, abciResponses(1, 0), nil, nil)
				require.NoError(t, err, tx)
				counter.Add(1)
			}
		}()
	}()

	// This was introduced because without a separate function
	// we have to sleep to wait for all txs to get into the mempool.
	// This way we loop in the function above until it is fool
	// without arbitrary timeouts.
	go func() {
		for i := 1; i <= int(numTxs); i++ {
			tx := kvstore.NewTxFromID(i)

			currLane := (i % numLanes) + 1
			reqRes := newReqResWithLanes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK, uint32(currLane))
			require.NotNil(t, reqRes)

			mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(reqRes, nil).Once()
			_, err := mp.CheckTx(tx, "")
			require.NoError(t, err, err)
			reqRes.InvokeCallback()
		}
	}()

	wg.Wait()

	require.Equal(t, counter.Load(), int64(numTxs+1))
}

func TestIteratorEmptyLanes(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_empty_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	go func() {
		iter := NewBlockingIterator(mp)
		require.Zero(t, mp.Size())
		entry := <-iter.WaitNextCh()
		require.NotNil(t, entry)
		require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(1))
	}()
	time.Sleep(100 * time.Millisecond)

	tx := kvstore.NewTxFromID(1)
	res := abci.ToCheckTxResponse(&abci.CheckTxResponse{Code: abci.CodeTypeOK})
	mp.handleCheckTxResponse(tx, "")(res)
	require.Equal(t, 1, mp.Size(), "pool size mismatch")
}

// Without lanes transactions should be returned as they were
// submitted - increasing tx IDs.
func TestIteratorNoLanes(t *testing.T) {
	app := kvstore.NewInMemoryApplicationWithoutLanes()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	const n = numTxs

	var wg sync.WaitGroup
	wg.Add(1)

	// Spawn a goroutine that iterates on the list until counting n entries.
	counter := 0
	go func() {
		defer wg.Done()

		iter := NewBlockingIterator(mp)
		for counter < n {
			entry := <-iter.WaitNextCh()
			if entry == nil {
				continue
			}
			require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(counter))
			counter++
		}
	}()

	// Add n transactions with sequential ids.
	for i := 0; i < n; i++ {
		tx := kvstore.NewTxFromID(i)
		rr, err := mp.CheckTx(tx, "")
		require.NoError(t, err)
		rr.Wait()
	}

	wg.Wait()
	require.Equal(t, n, counter)
}

// TODO automate the lane numbers so we can change the number of lanes
// and increase the number of transactions.
func TestIteratorExactOrder(t *testing.T) {
	mockClient := new(abciclimocks.Client)
	mockClient.On("Start").Return(nil)
	mockClient.On("SetLogger", mock.Anything)
	mockClient.On("Error").Return(nil).Times(100)
	mockClient.On("Info", mock.Anything, mock.Anything).Return(&abci.InfoResponse{LanePriorities: []uint32{1, 2, 3}, DefaultLanePriority: 1}, nil)

	mp, cleanup := newMempoolWithAppMock(mockClient)
	defer cleanup()

	// Disable rechecking to make sure the recheck logic is not interferint.
	mp.config.Recheck = false

	const numLanes = 3
	const numTxs = 11
	// Transactions are ordered into lanes by their IDs. This is the order in
	// which they should appear following WRR
	expectedTxIDs := []int{2, 5, 8, 1, 4, 3, 11, 7, 10, 6, 9}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		waitForNumTxsInMempool(numTxs, mp)
		t.Log("Mempool full, starting to pick up transactions", mp.Size())

		iter := NewBlockingIterator(mp)
		for i := 0; i < numTxs; i++ {
			entry := <-iter.WaitNextCh()
			if entry == nil {
				continue
			}
			require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(expectedTxIDs[i]))
		}
	}()

	// This was introduced because without a separate function
	// we have to sleep to wait for all txs to get into the mempool.
	// This way we loop in the function above until it is fool
	// without arbitrary timeouts.
	go func() {
		for i := 1; i <= numTxs; i++ {
			tx := kvstore.NewTxFromID(i)

			currLane := (i % numLanes) + 1
			reqRes := newReqResWithLanes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK, uint32(currLane))
			require.NotNil(t, reqRes)

			mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(reqRes, nil).Once()
			_, err := mp.CheckTx(tx, "")
			require.NoError(t, err, err)
			reqRes.InvokeCallback()
		}
	}()

	wg.Wait()
}

// This only tests that all transactions were submitted.
func TestIteratorCountOnly(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	var wg sync.WaitGroup
	wg.Add(1)

	const n = numTxs

	// Spawn a goroutine that iterates on the list until counting n entries.
	counter := 0
	go func() {
		defer wg.Done()

		iter := NewBlockingIterator(mp)
		for counter < n {
			entry := <-iter.WaitNextCh()
			if entry == nil {
				continue
			}
			counter++
		}
	}()

	// Add n transactions with sequential ids.
	for i := 0; i < n; i++ {
		tx := kvstore.NewTxFromID(i)
		rr, err := mp.CheckTx(tx, "")
		require.NoError(t, err)
		rr.Wait()
	}

	wg.Wait()
	require.Equal(t, n, counter)
}

func TestReapMatchesGossipOrder(t *testing.T) {
	const n = 10

	tests := map[string]struct {
		app *kvstore.Application
	}{
		"test_lanes": {
			app: kvstore.NewInMemoryApplication(),
		},
		"test_no_lanes": {
			app: kvstore.NewInMemoryApplicationWithoutLanes(),
		},
	}

	for test, config := range tests {
		cc := proxy.NewLocalClientCreator(config.app)
		mp, cleanup := newMempoolWithApp(cc)
		defer cleanup()
		// Add a bunch of txs.
		for i := 1; i <= n; i++ {
			tx := kvstore.NewTxFromID(i)
			rr, err := mp.CheckTx(tx, "")
			require.NoError(t, err, err)
			rr.Wait()
		}

		require.Equal(t, n, mp.Size())

		gossipIter := NewBlockingIterator(mp)
		reapIter := NewNonBlockingIterator(mp)

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
			if test == "test_no_lanes" {
				require.EqualValues(t, reapTx, kvstore.NewTxFromID(i+1))
			}
		}
		require.EqualValues(t, txs, reapedTxs)

		err := mp.Update(1, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
		require.NoError(t, err)
		require.Zero(t, mp.Size())
	}
}

func TestBlockingIteratorsConsumeAll(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	const numTxs = 1000
	const numIterators = 50

	mp.Flush()
	wg := sync.WaitGroup{}
	wg.Add(numIterators)

	// Concurrent iterators.
	for j := 0; j < numIterators; j++ {
		go func(j int) {
			defer wg.Done()
			iter := NewBlockingIterator(mp)
			// Iterate until all txs added to the mempool are accessed.
			for c := 0; c < numTxs; c++ {
				if entry := <-iter.WaitNextCh(); entry == nil {
					continue
				}
			}
			t.Logf("iterator %d finished\n", j)
		}(j)
	}

	// Add transactions.
	txs := addTxs(t, mp, 0, numTxs)
	require.Equal(t, numTxs, len(txs))
	require.Equal(t, numTxs, mp.Size())

	// Wait for all to complete.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("Timed out waiting for all iterators to finish")
	}
}
