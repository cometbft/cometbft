package mempool

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		// round counter 1:
		0, // lane 7
		1, // lane 3
		3, // lane 1
		// round counter 2:
		11, // lane 7
		2,  // lane 3
		// round counter 3:
		22, // lane 7
		4,  // lane 3
		// round counter 4 - 7:
		33, 44, 55, 66, // lane 7
		// round counter 1:
		77, // lane 7
		5,  // lane 3
		6,  // lane 1
		// round counter 2:
		88, // lane 7
		7,  // lane 3
		// round counter 3:
		99, // lane 7
		8,  // lane 3
		// round counter 4- 7 have nothing
		// round counter 1:
		10, // lane 3
		9,  // lane 1
		// round counter 2:
		13, // lane 3
		// round counter 3:
		14, // lane 3
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

	mockClient.On("Info", mock.Anything, mock.Anything).Return(&abci.InfoResponse{LanePriorities: map[string]uint32{"1": 1, "2": 2, "3": 3}, DefaultLane: "1"}, nil)

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
				iter := NewBlockingIterator(context.Background(), mp, t.Name())
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
				iter := NewBlockingIterator(context.Background(), mp, t.Name())
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
			reqRes := newReqResWithLanes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK, strconv.Itoa(currLane))
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
		iter := NewBlockingIterator(context.Background(), mp, t.Name())
		require.Zero(t, mp.Size())
		entry := <-iter.WaitNextCh()
		require.NotNil(t, entry)
		require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(1))
	}()
	time.Sleep(100 * time.Millisecond)

	tx := kvstore.NewTxFromID(1)
	res := abci.ToCheckTxResponse(&abci.CheckTxResponse{Code: abci.CodeTypeOK})
	err := mp.handleCheckTxResponse(tx, "")(res)
	require.NoError(t, err)
	require.Equal(t, 1, mp.Size(), "pool size mismatch")
}

func TestBlockingIteratorsConsumeAllTxs(t *testing.T) {
	const numTxs = 1000
	const numIterators = 50

	tests := map[string]struct {
		app *kvstore.Application
	}{
		"lanes": {
			app: kvstore.NewInMemoryApplication(),
		},
		"no_lanes": {
			app: kvstore.NewInMemoryApplicationWithoutLanes(),
		},
	}

	for test, config := range tests {
		cc := proxy.NewLocalClientCreator(config.app)
		mp, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		wg := &sync.WaitGroup{}
		wg.Add(numIterators)

		// Start concurrent iterators.
		for i := 0; i < numIterators; i++ {
			go func(j int) {
				defer wg.Done()

				// Iterate until all txs added to the mempool are accessed.
				iter := NewBlockingIterator(context.Background(), mp, strconv.Itoa(j))
				counter := 0
				nilCounter := 0
				for counter < numTxs {
					entry := <-iter.WaitNextCh()
					if entry == nil {
						nilCounter++
						continue
					}
					if test == "no_lanes" {
						// Entries are accessed sequentially when there is only one lane.
						expectedTx := kvstore.NewTxFromID(counter)
						require.EqualValues(t, expectedTx, entry.Tx(), "i=%d, c=%d, tx=%v", i, counter, entry.Tx())
					}
					counter++
				}
				require.Equal(t, numTxs, counter)
				assert.Zero(t, nilCounter, "got nil entries")
				t.Logf("%s: iterator %d finished (nils=%d)\n", test, j, nilCounter)
			}(i)
		}

		// Add transactions with sequential ids.
		_ = addTxs(t, mp, 0, numTxs)
		require.Equal(t, numTxs, mp.Size())

		// Wait for all iterators to complete.
		waitTimeout(wg, 5*time.Second, func() {}, func() {
			t.Fatalf("Timed out waiting for all iterators to finish")
		})
	}
}

// Confirms that the transactions are returned in the same order.
// Note that for the cases with equal priorities the actual order
// will depend on the way we iterate over the map of lanes.
// With only two lanes of the same priority the order was predictable
// and matches the given order. In case these tests start to fail
// first thing to confirm is the order of lanes in mp.SortedLanes.
func TestIteratorExactOrder(t *testing.T) {
	tests := map[string]struct {
		lanePriorities         map[string]uint32
		expectedTxIDs          []int
		expectedTxIDsAlternate []int
	}{
		"unique_priority_lanes": {
			lanePriorities: map[string]uint32{"1": 1, "2": 2, "3": 3},
			expectedTxIDs:  []int{2, 1, 3, 5, 4, 8, 11, 7, 6, 10, 9},
		},
		"same_priority_lanes": {
			lanePriorities:         map[string]uint32{"1": 1, "2": 2, "3": 2},
			expectedTxIDs:          []int{1, 2, 3, 4, 5, 7, 8, 6, 10, 11, 9},
			expectedTxIDsAlternate: []int{2, 1, 3, 5, 4, 8, 7, 6, 11, 10, 9},
		},
		"one_lane": {
			lanePriorities: map[string]uint32{"1": 1},
			expectedTxIDs:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	for n, l := range tests {
		mockClient := new(abciclimocks.Client)
		mockClient.On("Start").Return(nil)
		mockClient.On("SetLogger", mock.Anything)
		mockClient.On("Error").Return(nil).Times(100)
		mockClient.On("Info", mock.Anything, mock.Anything).Return(&abci.InfoResponse{LanePriorities: l.lanePriorities, DefaultLane: "1"}, nil)
		mp, cleanup := newMempoolWithAppMock(mockClient)
		defer cleanup()

		// Disable rechecking to make sure the recheck logic is not interfering.
		mp.config.Recheck = false

		numLanes := len(l.lanePriorities)
		const numTxs = 11

		// Transactions are ordered into lanes by their IDs. This is the order in
		// which they should appear following WRR
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			waitForNumTxsInMempool(numTxs, mp)
			t.Log("Mempool full, starting to pick up transactions", mp.Size())
			alternate := false
			iter := NewBlockingIterator(context.Background(), mp, t.Name())
			for i := 0; i < numTxs; i++ {
				entry := <-iter.WaitNextCh()
				if entry == nil {
					continue
				}
				// When lanes have same priorities their order in the map of lanes
				// is arbitrary so we needv to check
				if n == "same_priority_lanes" {
					if mp.sortedLanes[1].id != "3" {
						alternate = true
					}
				}
				if alternate {
					require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(l.expectedTxIDsAlternate[i]), n)
				} else {
					require.EqualValues(t, entry.Tx(), kvstore.NewTxFromID(l.expectedTxIDs[i]), n)
				}
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
				reqRes := newReqResWithLanes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK, strconv.Itoa(currLane))
				require.NotNil(t, reqRes)

				mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(reqRes, nil).Once()
				_, err := mp.CheckTx(tx, "")
				require.NoError(t, err, err)
				reqRes.InvokeCallback()
			}
		}()

		wg.Wait()

		// Confirm also that the non blocking iterator works with lanes of same priorities
		iterNonBlocking := NewNonBlockingIterator(mp)
		reapedTx := mp.ReapMaxTxs(numTxs)
		alternate := false
		for i := 0; i < numTxs; i++ {
			tx := iterNonBlocking.Next().Tx()
			if n == "same_priority_lanes" {
				if mp.sortedLanes[1].id != "3" {
					alternate = true
				}
			}
			if !alternate {
				require.Equal(t, []byte(tx), kvstore.NewTxFromID(l.expectedTxIDs[i]), n)
				require.Equal(t, []byte(reapedTx[i]), kvstore.NewTxFromID(l.expectedTxIDs[i]), n)
			} else {
				require.Equal(t, []byte(tx), kvstore.NewTxFromID(l.expectedTxIDsAlternate[i]), n)
				require.Equal(t, []byte(reapedTx[i]), kvstore.NewTxFromID(l.expectedTxIDsAlternate[i]), n)
			}
		}
	}
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

		iter := NewBlockingIterator(context.Background(), mp, t.Name())
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
	const n = 100

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

		gossipIter := NewBlockingIterator(context.Background(), mp, t.Name())
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
