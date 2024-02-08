package mempool

import (
	"context"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	abciclient "github.com/cometbft/cometbft/abci/client"
	abciclimocks "github.com/cometbft/cometbft/abci/client/mocks"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	abciserver "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/internal/service"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

// A cleanupFunc cleans up any config / test files created for a particular
// test.
type cleanupFunc func()

func newMempoolWithAppMock(client abciclient.Client) (*CListMempool, cleanupFunc) {
	conf := test.ResetTestRoot("mempool_test")

	mp, cu := newMempoolWithAppAndConfigMock(conf, client)
	return mp, cu
}

func newMempoolWithAppAndConfigMock(
	cfg *config.Config,
	client abciclient.Client,
) (*CListMempool, cleanupFunc) {
	appConnMem := client
	appConnMem.SetLogger(log.TestingLogger().With("module", "abci-client", "connection", "mempool"))
	err := appConnMem.Start()
	if err != nil {
		panic(err)
	}

	mp := NewCListMempool(cfg.Mempool, appConnMem, 0)
	mp.SetLogger(log.TestingLogger())

	return mp, func() { os.RemoveAll(cfg.RootDir) }
}

func newMempoolWithApp(cc proxy.ClientCreator) (*CListMempool, cleanupFunc) {
	conf := test.ResetTestRoot("mempool_test")

	mp, cu := newMempoolWithAppAndConfig(cc, conf)
	return mp, cu
}

func newMempoolWithAppAndConfig(cc proxy.ClientCreator, cfg *config.Config) (*CListMempool, cleanupFunc) {
	appConnMem, _ := cc.NewABCIMempoolClient()
	appConnMem.SetLogger(log.TestingLogger().With("module", "abci-client", "connection", "mempool"))
	err := appConnMem.Start()
	if err != nil {
		panic(err)
	}

	mp := NewCListMempool(cfg.Mempool, appConnMem, 0)
	mp.SetLogger(log.TestingLogger())

	return mp, func() { os.RemoveAll(cfg.RootDir) }
}

func ensureNoFire(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	timer := time.NewTimer(100 * time.Millisecond)
	select {
	case <-ch:
		t.Fatal("Expected not to fire")
	case <-timer.C:
	}
}

func ensureFire(t *testing.T, ch <-chan struct{}, timeoutMS int) {
	t.Helper()
	timer := time.NewTimer(time.Duration(timeoutMS) * time.Millisecond)
	select {
	case <-ch:
	case <-timer.C:
		t.Fatal("Expected to fire")
	}
}

// Call CheckTx on a given mempool on each transaction in the list.
func callCheckTx(t *testing.T, mp Mempool, txs types.Txs) {
	t.Helper()
	for i, tx := range txs {
		if _, err := mp.CheckTx(tx); err != nil {
			// Skip invalid txs.
			// TestMempoolFilters will fail otherwise. It asserts a number of txs
			// returned.
			if IsPreCheckError(err) {
				continue
			}
			t.Fatalf("CheckTx failed: %v while checking #%d tx", err, i)
		}
	}
}

// Generate a list of random transactions.
func NewRandomTxs(numTxs int, txLen int) types.Txs {
	txs := make(types.Txs, numTxs)
	for i := 0; i < numTxs; i++ {
		txBytes := kvstore.NewRandomTx(txLen)
		txs[i] = txBytes
	}
	return txs
}

// Generate a list of random transactions of a given size and call CheckTx on
// each of them.
func checkTxs(t *testing.T, mp Mempool, count int) types.Txs {
	t.Helper()
	txs := NewRandomTxs(count, 20)
	callCheckTx(t, mp, txs)
	return txs
}

func TestReapMaxBytesMaxGas(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	// Ensure gas calculation behaves as expected
	checkTxs(t, mp, 1)
	tx0 := mp.TxsFront().Value.(*mempoolTx)
	require.Equal(t, tx0.gasWanted, int64(1), "transactions gas was set incorrectly")
	// ensure each tx is 20 bytes long
	require.Len(t, tx0.tx, 20, "Tx is longer than 20 bytes")
	mp.Flush()

	// each table driven test creates numTxsToCreate txs with checkTx, and at the end clears all remaining txs.
	// each tx has 20 bytes
	tests := []struct {
		numTxsToCreate int
		maxBytes       int64
		maxGas         int64
		expectedNumTxs int
	}{
		{20, -1, -1, 20},
		{20, -1, 0, 0},
		{20, -1, 10, 10},
		{20, -1, 30, 20},
		{20, 0, -1, 0},
		{20, 0, 10, 0},
		{20, 10, 10, 0},
		{20, 24, 10, 1},
		{20, 240, 5, 5},
		{20, 240, -1, 10},
		{20, 240, 10, 10},
		{20, 240, 15, 10},
		{20, 20000, -1, 20},
		{20, 20000, 5, 5},
		{20, 20000, 30, 20},
	}
	for tcIndex, tt := range tests {
		checkTxs(t, mp, tt.numTxsToCreate)
		got := mp.ReapMaxBytesMaxGas(tt.maxBytes, tt.maxGas)
		require.Len(t, got, tt.expectedNumTxs, "Got %d txs, expected %d, tc #%d",
			len(got), tt.expectedNumTxs, tcIndex)
		mp.Flush()
	}
}

func TestMempoolFilters(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()
	emptyTxArr := []types.Tx{[]byte{}}

	nopPreFilter := func(tx types.Tx) error { return nil }
	nopPostFilter := func(tx types.Tx, res *abci.CheckTxResponse) error { return nil }

	// each table driven test creates numTxsToCreate txs with checkTx, and at the end clears all remaining txs.
	// each tx has 20 bytes
	tests := []struct {
		numTxsToCreate int
		preFilter      PreCheckFunc
		postFilter     PostCheckFunc
		expectedNumTxs int
	}{
		{10, nopPreFilter, nopPostFilter, 10},
		{10, PreCheckMaxBytes(10), nopPostFilter, 0},
		{10, PreCheckMaxBytes(22), nopPostFilter, 10},
		{10, nopPreFilter, PostCheckMaxGas(-1), 10},
		{10, nopPreFilter, PostCheckMaxGas(0), 0},
		{10, nopPreFilter, PostCheckMaxGas(1), 10},
		{10, nopPreFilter, PostCheckMaxGas(3000), 10},
		{10, PreCheckMaxBytes(10), PostCheckMaxGas(20), 0},
		{10, PreCheckMaxBytes(30), PostCheckMaxGas(20), 10},
		{10, PreCheckMaxBytes(22), PostCheckMaxGas(1), 10},
		{10, PreCheckMaxBytes(22), PostCheckMaxGas(0), 0},
	}
	for tcIndex, tt := range tests {
		err := mp.Update(1, emptyTxArr, abciResponses(len(emptyTxArr), abci.CodeTypeOK), tt.preFilter, tt.postFilter)
		require.NoError(t, err)
		checkTxs(t, mp, tt.numTxsToCreate)
		require.Equal(t, tt.expectedNumTxs, mp.Size(), "mempool had the incorrect size, on test case %d", tcIndex)
		mp.Flush()
	}
}

func TestMempoolUpdate(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	// 1. Adds valid txs to the cache
	{
		tx1 := kvstore.NewTxFromID(1)
		err := mp.Update(1, []types.Tx{tx1}, abciResponses(1, abci.CodeTypeOK), nil, nil)
		require.NoError(t, err)
		_, err = mp.CheckTx(tx1)
		if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
			assert.Equal(t, ErrTxInCache, err)
		}
	}

	// 2. Removes valid txs from the mempool
	{
		tx2 := kvstore.NewTxFromID(2)
		_, err := mp.CheckTx(tx2)
		require.NoError(t, err)
		err = mp.Update(1, []types.Tx{tx2}, abciResponses(1, abci.CodeTypeOK), nil, nil)
		require.NoError(t, err)
		assert.Zero(t, mp.Size())
	}

	// 3. Removes invalid transactions from the cache and the mempool (if present)
	{
		tx3 := kvstore.NewTxFromID(3)
		_, err := mp.CheckTx(tx3)
		require.NoError(t, err)
		err = mp.Update(1, []types.Tx{tx3}, abciResponses(1, 1), nil, nil)
		require.NoError(t, err)
		assert.Zero(t, mp.Size())

		_, err = mp.CheckTx(tx3)
		require.NoError(t, err)
	}
}

// Test dropping CheckTx requests when rechecking transactions. It mocks an asynchronous connection
// to the app.
func TestMempoolUpdateDoesNotPanicWhenApplicationMissedTx(t *testing.T) {
	var callback abciclient.Callback
	mockClient := new(abciclimocks.Client)
	mockClient.On("Start").Return(nil)
	mockClient.On("SetLogger", mock.Anything)
	mockClient.On("Error").Return(nil).Times(4)
	mockClient.On("SetResponseCallback", mock.MatchedBy(func(cb abciclient.Callback) bool { callback = cb; return true }))
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(nil, nil)
	mockClient.On("Flush", mock.Anything).Return(nil)

	mp, cleanup := newMempoolWithAppMock(mockClient)
	defer cleanup()

	// Add 4 transactions to the mempool by calling the mempool's `CheckTx` on each of them.
	txs := []types.Tx{[]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}}
	for _, tx := range txs {
		_, err := mp.CheckTx(tx)
		require.NoError(t, err)
	}
	require.Zero(t, mp.Size())

	// Invoke CheckTx callbacks asynchronously.
	for _, tx := range txs {
		reqRes := newReqRes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK)
		callback(reqRes.Request, reqRes.Response)
	}
	require.Len(t, txs, mp.Size())
	require.True(t, mp.recheck.done())

	// Calling update to remove the first transaction from the mempool.
	// This call also triggers the mempool to recheck its remaining transactions.
	err := mp.Update(0, []types.Tx{txs[0]}, abciResponses(1, abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)

	// The mempool has now sent its requests off to the client to be rechecked
	// and is waiting for the corresponding callbacks to be called.
	// We now call the mempool-supplied callback on the first and third transaction.
	// This simulates the client dropping the second request.
	// Previous versions of this code panicked when the ABCI application missed
	// a recheck-tx request.
	reqRes := newReqRes(txs[1], abci.CodeTypeOK, abci.CHECK_TX_TYPE_RECHECK)
	callback(reqRes.Request, reqRes.Response)

	reqRes = newReqRes(txs[3], abci.CodeTypeOK, abci.CHECK_TX_TYPE_RECHECK)
	callback(reqRes.Request, reqRes.Response)

	mockClient.AssertExpectations(t)
}

func TestMempool_KeepInvalidTxsInCache(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	wcfg := config.DefaultConfig()
	wcfg.Mempool.KeepInvalidTxsInCache = true
	mp, cleanup := newMempoolWithAppAndConfig(cc, wcfg)
	defer cleanup()

	// 1. An invalid transaction must remain in the cache after Update
	{
		a := make([]byte, 8)
		binary.BigEndian.PutUint64(a, 0)

		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, 1)

		_, err := mp.CheckTx(b)
		require.NoError(t, err)

		// simulate new block
		_, err = app.FinalizeBlock(context.Background(), &abci.FinalizeBlockRequest{
			Txs: [][]byte{a, b},
		})
		require.NoError(t, err)
		err = mp.Update(1, []types.Tx{a, b},
			[]*abci.ExecTxResult{{Code: abci.CodeTypeOK}, {Code: 2}}, nil, nil)
		require.NoError(t, err)

		// a must be added to the cache
		_, err = mp.CheckTx(a)
		if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
			assert.Equal(t, ErrTxInCache, err)
		}

		// b must remain in the cache
		_, err = mp.CheckTx(b)
		if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
			assert.Equal(t, ErrTxInCache, err)
		}
	}

	// 2. An invalid transaction must remain in the cache
	{
		a := make([]byte, 8)
		binary.BigEndian.PutUint64(a, 0)

		// remove a from the cache to test (2)
		mp.cache.Remove(a)

		_, err := mp.CheckTx(a)
		require.NoError(t, err)
	}
}

func TestTxsAvailable(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()
	mp.EnableTxsAvailable()

	timeoutMS := 100

	// with no txs, it shouldn't fire
	ensureNoFire(t, mp.TxsAvailable())

	// send a bunch of txs, it should only fire once
	txs := checkTxs(t, mp, 100)
	ensureFire(t, mp.TxsAvailable(), timeoutMS)
	ensureNoFire(t, mp.TxsAvailable())

	// call update with half the txs.
	// it should fire once now for the new height
	// since there are still txs left
	committedTxs, remainingTxs := txs[:50], txs[50:]
	if err := mp.Update(1, committedTxs, abciResponses(len(committedTxs), abci.CodeTypeOK), nil, nil); err != nil {
		t.Error(err)
	}
	ensureFire(t, mp.TxsAvailable(), timeoutMS)
	ensureNoFire(t, mp.TxsAvailable())

	// send a bunch more txs. we already fired for this height so it shouldn't fire again
	moreTxs := checkTxs(t, mp, 50)
	ensureNoFire(t, mp.TxsAvailable())

	// now call update with all the txs. it should not fire as there are no txs left
	committedTxs = append(remainingTxs, moreTxs...)
	if err := mp.Update(2, committedTxs, abciResponses(len(committedTxs), abci.CodeTypeOK), nil, nil); err != nil {
		t.Error(err)
	}
	ensureNoFire(t, mp.TxsAvailable())

	// send a bunch more txs, it should only fire once
	checkTxs(t, mp, 100)
	ensureFire(t, mp.TxsAvailable(), timeoutMS)
	ensureNoFire(t, mp.TxsAvailable())
}

func TestSerialReap(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	appConnCon, _ := cc.NewABCIConsensusClient()
	appConnCon.SetLogger(log.TestingLogger().With("module", "abci-client", "connection", "consensus"))
	err := appConnCon.Start()
	require.NoError(t, err)

	cacheMap := make(map[string]struct{})
	deliverTxsRange := func(start, end int) {
		// Deliver some txs.
		for i := start; i < end; i++ {
			txBytes := kvstore.NewTx(strconv.Itoa(i), "true")
			_, err := mp.CheckTx(txBytes)
			_, cached := cacheMap[string(txBytes)]
			if cached {
				require.Error(t, err, "expected error for cached tx")
			} else {
				require.NoError(t, err, "expected no err for uncached tx")
			}
			cacheMap[string(txBytes)] = struct{}{}

			// Duplicates are cached and should return error
			_, err = mp.CheckTx(txBytes)
			require.Error(t, err, "Expected error after CheckTx on duplicated tx")
		}
	}

	reapCheck := func(exp int) {
		txs := mp.ReapMaxBytesMaxGas(-1, -1)
		require.Len(t, txs, exp)
	}

	updateRange := func(start, end int) {
		txs := make(types.Txs, end-start)
		for i := start; i < end; i++ {
			txs[i-start] = kvstore.NewTx(strconv.Itoa(i), "true")
		}
		if err := mp.Update(0, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil); err != nil {
			t.Error(err)
		}
	}

	commitRange := func(start, end int) {
		// Deliver some txs in a block
		txs := make([][]byte, end-start)
		for i := start; i < end; i++ {
			txs[i-start] = kvstore.NewTx(strconv.Itoa(i), "true")
		}

		res, err := appConnCon.FinalizeBlock(context.Background(), &abci.FinalizeBlockRequest{Txs: txs})
		if err != nil {
			t.Errorf("client error committing tx: %v", err)
		}
		for _, txResult := range res.TxResults {
			if txResult.IsErr() {
				t.Errorf("error committing tx. Code:%v result:%X log:%v",
					txResult.Code, txResult.Data, txResult.Log)
			}
		}
		if len(res.AppHash) != 8 {
			t.Errorf("error committing. Hash:%X", res.AppHash)
		}

		_, err = appConnCon.Commit(context.Background(), &abci.CommitRequest{})
		if err != nil {
			t.Errorf("client error committing: %v", err)
		}
	}

	//----------------------------------------

	// Deliver some txs.
	deliverTxsRange(0, 100)

	// Reap the txs.
	reapCheck(100)

	// Reap again.  We should get the same amount
	reapCheck(100)

	// Deliver 0 to 999, we should reap 900 new txs
	// because 100 were already counted.
	deliverTxsRange(0, 1000)

	// Reap the txs.
	reapCheck(1000)

	// Reap again.  We should get the same amount
	reapCheck(1000)

	// Commit from the consensus AppConn
	commitRange(0, 500)
	updateRange(0, 500)

	// We should have 500 left.
	reapCheck(500)

	// Deliver 100 invalid txs and 100 valid txs
	deliverTxsRange(900, 1100)

	// We should have 600 now.
	reapCheck(600)
}

func TestMempool_CheckTxChecksTxSize(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	mempl, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	maxTxSize := mempl.config.MaxTxBytes

	testCases := []struct {
		len int
		err bool
	}{
		// check small txs. no error
		0: {10, false},
		1: {1000, false},
		2: {1000000, false},

		// check around maxTxSize
		3: {maxTxSize - 1, false},
		4: {maxTxSize, false},
		5: {maxTxSize + 1, true},
	}

	for i, testCase := range testCases {
		caseString := fmt.Sprintf("case %d, len %d", i, testCase.len)

		tx := cmtrand.Bytes(testCase.len)

		_, err := mempl.CheckTx(tx)
		bv := gogotypes.BytesValue{Value: tx}
		bz, err2 := bv.Marshal()
		require.NoError(t, err2)
		require.Len(t, bz, proto.Size(&bv), caseString)

		if !testCase.err {
			require.NoError(t, err, caseString)
		} else {
			require.Equal(t, ErrTxTooLarge{
				Max:    maxTxSize,
				Actual: testCase.len,
			}, err, caseString)
		}
	}
}

func TestMempoolTxsBytes(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_test")

	cfg.Mempool.MaxTxsBytes = 100
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	// 1. zero by default
	assert.EqualValues(t, 0, mp.SizeBytes())

	// 2. len(tx) after CheckTx
	tx1 := kvstore.NewRandomTx(10)
	_, err := mp.CheckTx(tx1)
	require.NoError(t, err)
	assert.EqualValues(t, 10, mp.SizeBytes())

	// 3. zero again after tx is removed by Update
	err = mp.Update(1, []types.Tx{tx1}, abciResponses(1, abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 0, mp.SizeBytes())

	// 4. zero after Flush
	tx2 := kvstore.NewRandomTx(20)
	_, err = mp.CheckTx(tx2)
	require.NoError(t, err)
	assert.EqualValues(t, 20, mp.SizeBytes())

	mp.Flush()
	assert.EqualValues(t, 0, mp.SizeBytes())

	// 5. ErrMempoolIsFull is returned when/if MaxTxsBytes limit is reached.
	tx3 := kvstore.NewRandomTx(100)
	_, err = mp.CheckTx(tx3)
	require.NoError(t, err)

	tx4 := kvstore.NewRandomTx(10)
	_, err = mp.CheckTx(tx4)
	if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
		assert.IsType(t, ErrMempoolIsFull{}, err)
	}

	// 6. zero after tx is rechecked and removed due to not being valid anymore
	app2 := kvstore.NewInMemoryApplication()
	cc = proxy.NewLocalClientCreator(app2)

	mp, cleanup = newMempoolWithApp(cc)
	defer cleanup()

	txBytes := kvstore.NewRandomTx(10)

	_, err = mp.CheckTx(txBytes)
	require.NoError(t, err)
	assert.EqualValues(t, 10, mp.SizeBytes())

	appConnCon, _ := cc.NewABCIConsensusClient()
	appConnCon.SetLogger(log.TestingLogger().With("module", "abci-client", "connection", "consensus"))
	err = appConnCon.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := appConnCon.Stop(); err != nil {
			t.Error(err)
		}
	})

	res, err := appConnCon.FinalizeBlock(context.Background(), &abci.FinalizeBlockRequest{Txs: [][]byte{txBytes}})
	require.NoError(t, err)
	require.EqualValues(t, 0, res.TxResults[0].Code)
	require.NotEmpty(t, res.AppHash)

	_, err = appConnCon.Commit(context.Background(), &abci.CommitRequest{})
	require.NoError(t, err)

	// Pretend like we committed nothing so txBytes gets rechecked and removed.
	err = mp.Update(1, []types.Tx{}, abciResponses(0, abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 10, mp.SizeBytes())

	// 7. Test RemoveTxByKey function
	_, err = mp.CheckTx(tx1)
	require.NoError(t, err)
	assert.EqualValues(t, 20, mp.SizeBytes())
	require.Error(t, mp.RemoveTxByKey(types.Tx([]byte{0x07}).Key()))
	assert.EqualValues(t, 20, mp.SizeBytes())
	require.NoError(t, mp.RemoveTxByKey(types.Tx(tx1).Key()))
	assert.EqualValues(t, 10, mp.SizeBytes())
}

func TestMempoolNoCacheOverflow(t *testing.T) {
	mp, cleanup := newMempoolWithAsyncConnection(t)
	defer cleanup()

	// add tx0
	tx0 := kvstore.NewTxFromID(0)
	_, err := mp.CheckTx(tx0)
	require.NoError(t, err)
	err = mp.FlushAppConn()
	require.NoError(t, err)

	// saturate the cache to remove tx0
	for i := 1; i <= mp.config.CacheSize; i++ {
		_, err = mp.CheckTx(kvstore.NewTxFromID(i))
		require.NoError(t, err)
	}
	err = mp.FlushAppConn()
	require.NoError(t, err)
	assert.False(t, mp.cache.Has(kvstore.NewTxFromID(0)))

	// add again tx0
	_, err = mp.CheckTx(tx0)
	require.NoError(t, err)
	err = mp.FlushAppConn()
	require.NoError(t, err)

	// tx0 should appear only once in mp.txs
	found := 0
	for e := mp.txs.Front(); e != nil; e = e.Next() {
		if types.Tx.Key(e.Value.(*mempoolTx).tx) == types.Tx.Key(tx0) {
			found++
		}
	}
	assert.Equal(t, 1, found)
}

// This will non-deterministically catch some concurrency failures like
// https://github.com/tendermint/tendermint/issues/3509
// TODO: all of the tests should probably also run using the remote proxy app
// since otherwise we're not actually testing the concurrency of the mempool here!
func TestMempoolRemoteAppConcurrency(t *testing.T) {
	mp, cleanup := newMempoolWithAsyncConnection(t)
	defer cleanup()

	// generate small number of txs
	nTxs := 10
	txLen := 200
	txs := NewRandomTxs(nTxs, txLen)

	// simulate a group of peers sending them over and over
	N := mp.config.Size
	for i := 0; i < N; i++ {
		txNum := mrand.Intn(nTxs)
		tx := txs[txNum]

		// this will err with ErrTxInCache many times ...
		mp.CheckTx(tx) //nolint: errcheck // will error
	}

	require.NoError(t, mp.FlushAppConn())
}

func TestMempoolConcurrentUpdateAndReceiveCheckTxResponse(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	for h := 1; h <= 100; h++ {
		// Two concurrent threads for each height. One updates the mempool with one valid tx,
		// writing the pool's height; the other, receives a CheckTx response, reading the height.
		var wg sync.WaitGroup
		wg.Add(2)

		go func(h int) {
			defer wg.Done()

			err := mp.Update(int64(h), []types.Tx{tx}, abciResponses(1, abci.CodeTypeOK), nil, nil)
			require.NoError(t, err)
			require.Equal(t, int64(h), mp.height.Load(), "height mismatch")
		}(h)

		go func(h int) {
			defer wg.Done()

			tx := kvstore.NewTxFromID(h)
			mp.resCbFirstTime(tx, &abci.CheckTxResponse{Code: abci.CodeTypeOK})
			require.Equal(t, h, mp.Size(), "pool size mismatch")
		}(h)

		wg.Wait()
	}
}

func TestMempoolNotifyTxsAvailable(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)

	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(cc, cfg)
	defer cleanup()

	mp.EnableTxsAvailable()
	assert.NotNil(t, mp.txsAvailable)
	require.False(t, mp.notifiedTxsAvailable.Load())

	// Adding a new valid tx to the pool will notify a tx is available
	tx := kvstore.NewTxFromID(1)
	mp.resCbFirstTime(tx, &abci.CheckTxResponse{Code: abci.CodeTypeOK})
	require.Equal(t, 1, mp.Size(), "pool size mismatch")
	require.True(t, mp.notifiedTxsAvailable.Load())
	require.Len(t, mp.TxsAvailable(), 1)
	<-mp.TxsAvailable()

	// Receiving CheckTx response for a tx already in the pool should not notify of available txs
	mp.resCbFirstTime(tx, &abci.CheckTxResponse{Code: abci.CodeTypeOK})
	require.Equal(t, 1, mp.Size())
	require.True(t, mp.notifiedTxsAvailable.Load())
	require.Empty(t, mp.TxsAvailable())

	// Updating the pool will remove the tx and set the variable to false
	err := mp.Update(1, []types.Tx{tx}, abciResponses(1, abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	require.Zero(t, mp.Size())
	require.False(t, mp.notifiedTxsAvailable.Load())
}

func TestMempoolSyncRecheckTxReturnError(t *testing.T) {
	var callback abciclient.Callback
	mockClient := new(abciclimocks.Client)
	mockClient.On("Start").Return(nil)
	mockClient.On("SetLogger", mock.Anything)
	mockClient.On("Flush", mock.Anything).Return(nil)
	mockClient.On("Error").Return(nil).Times(4)
	mockClient.On("SetResponseCallback", mock.MatchedBy(func(cb abciclient.Callback) bool { callback = cb; return true }))

	mp, cleanup := newMempoolWithAppMock(mockClient)
	defer cleanup()
	mp.EnableTxsAvailable()

	// First we add a few txs to the mempool.
	txs := []types.Tx{[]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}}
	for _, tx := range txs {
		// We are mocking a sync client, which will call the callback just after the response from
		// the app.
		mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			req := args.Get(1).(*abci.CheckTxRequest)
			reqRes := newReqRes(req.Tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK)
			callback(reqRes.Request, reqRes.Response)
		}).Return(nil, nil).Once()
		_, err := mp.CheckTx(tx)
		require.NoError(t, err)
	}
	require.Len(t, txs, mp.Size())

	// The first tx is valid when rechecking, and the client will call the callback right after the
	// response from the app and before returning.
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		reqRes := newReqRes(txs[0], abci.CodeTypeOK, abci.CHECK_TX_TYPE_RECHECK)
		callback(reqRes.Request, reqRes.Response)
	}).Return(nil, nil).Once()

	// On the second tx the app returns an error, so the callback will not be called by the client.
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(nil, ErrCheckTxAsync{}).Once()

	// The third tx is invalid, so it will be removed from the mempool.
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		reqRes := newReqRes(txs[2], 1, abci.CHECK_TX_TYPE_RECHECK)
		callback(reqRes.Request, reqRes.Response)
	}).Return(nil, nil).Once()

	// On the fourth tx the app returns an error, so the callback will not be called by the client.
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(nil, ErrCheckTxAsync{}).Once()

	// Check that recheck has not started yet.
	require.True(t, mp.recheck.done())
	require.Nil(t, mp.recheck.cursor)
	require.Nil(t, mp.recheck.end)

	// mp.recheck.done() should be true only before and after calling recheckTxs.
	mp.recheckTxs()
	require.True(t, mp.recheck.done())
	require.Nil(t, mp.recheck.cursor)
	require.NotNil(t, mp.recheck.end)
	require.Equal(t, mp.recheck.end, mp.txs.Back())
	require.Equal(t, len(txs)-1, mp.Size()) // one invalid tx was removed
	require.Equal(t, int32(2), mp.recheck.numPendingTxs.Load())

	mockClient.AssertExpectations(t)
}

func TestMempoolAsyncRecheckTxReturnError(t *testing.T) {
	var callback abciclient.Callback
	mockClient := new(abciclimocks.Client)
	mockClient.On("Start").Return(nil)
	mockClient.On("SetLogger", mock.Anything)
	mockClient.On("Error").Return(nil).Times(4)
	mockClient.On("SetResponseCallback", mock.MatchedBy(func(cb abciclient.Callback) bool { callback = cb; return true }))

	// Mocking the async client will just send the request and not call the callback, which we do
	// manually later. The first 4 times CheckTxAsync is called are for adding the transactions to
	// the mempool. The other 4 are for rechecking, and in particular, it fails on the second and
	// last transactions.
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(nil, nil).Times(6)
	mockClient.On("CheckTxAsync", mock.Anything, mock.Anything).Return(nil, ErrCheckTxAsync{}).Times(2)

	mp, cleanup := newMempoolWithAppMock(mockClient)
	defer cleanup()
	mp.EnableTxsAvailable()

	// First send CheckTx requests for a few txs.
	txs := []types.Tx{[]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}}
	for _, tx := range txs {
		_, err := mp.CheckTx(tx)
		require.NoError(t, err)
	}
	require.Zero(t, mp.Size())

	// Invoke CheckTx callbacks asynchronously to add txs to mempool.
	for _, tx := range txs {
		reqRes := newReqRes(tx, abci.CodeTypeOK, abci.CHECK_TX_TYPE_CHECK)
		callback(reqRes.Request, reqRes.Response)
	}
	require.Len(t, txs, mp.Size())

	// Check that recheck has not started yet.
	require.True(t, mp.recheck.done())
	require.Nil(t, mp.recheck.cursor)
	require.Nil(t, mp.recheck.end)

	// The first tx is valid, the third is invalid, and the request for the second and fourth tx
	// were not sent, so the callback is not called.
	mockClient.On("Flush", mock.Anything).Run(func(args mock.Arguments) {
		reqRes1 := newReqRes(txs[0], abci.CodeTypeOK, abci.CHECK_TX_TYPE_RECHECK)
		callback(reqRes1.Request, reqRes1.Response)
		reqRes2 := newReqRes(txs[2], 1, abci.CHECK_TX_TYPE_RECHECK)
		callback(reqRes2.Request, reqRes2.Response)
	}).Return(nil)

	// mp.recheck.done() should be true only before and after calling recheckTxs.
	mp.recheckTxs()
	require.True(t, mp.recheck.done())
	require.Nil(t, mp.recheck.cursor)
	require.NotNil(t, mp.recheck.end)
	require.Equal(t, mp.recheck.end, mp.txs.Back())
	require.Equal(t, len(txs)-1, mp.Size()) // one invalid tx was removed
	require.Equal(t, int32(2), mp.recheck.numPendingTxs.Load())

	mockClient.AssertExpectations(t)
}

func TestMempoolRecheckRace(t *testing.T) {
	mp, cleanup := newMempoolWithAsyncConnection(t)
	defer cleanup()

	// Add a bunch of transactions to the mempool.
	var err error
	txs := newUniqueTxs(10)
	for _, tx := range txs {
		_, err = mp.CheckTx(tx)
		require.NoError(t, err)
	}

	// Update one transaction to force rechecking the rest.
	mp.Lock()
	err = mp.FlushAppConn()
	require.NoError(t, err)
	err = mp.Update(1, txs[:1], abciResponses(1, abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	mp.Unlock()

	// Add again the same transaction that was updated. Checking this
	// transaction and rechecking the others simultaneously should not result
	// in a data race for the variable recheckCursor.
	_, err = mp.CheckTx(txs[:1][0])
	require.Equal(t, err, ErrTxInCache)
	require.Zero(t, mp.recheck.numPendingTxs.Load())
}

// While connected to the app with an async client, this test adds txs with CheckTx while a
// concurrent routine updates the mempool.
func TestMempoolConcurrentCheckTxAndUpdate(t *testing.T) {
	mp, cleanup := newMempoolWithAsyncConnection(t)
	defer cleanup()

	maxHeight := 100
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		time.Sleep(50 * time.Millisecond) // wait a bit to have some txs in mempool before starting updating
		for h := 1; h <= maxHeight; h++ {
			if mp.Size() == 0 {
				break
			}
			txs := mp.ReapMaxBytesMaxGas(100, -1)
			mp.Lock()
			mp.FlushAppConn() // needed to process the pending CheckTx requests and their callbacks
			err := mp.Update(int64(h), txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
			mp.Unlock()
			require.NoError(t, err)
		}
	}()

	for h := 1; h <= maxHeight; h++ {
		_, err := mp.CheckTx(kvstore.NewTxFromID(h))
		require.NoError(t, err)
	}

	wg.Wait()
}

func newMempoolWithAsyncConnection(t *testing.T) (*CListMempool, cleanupFunc) {
	sockPath := fmt.Sprintf("unix:///tmp/echo_%v.sock", cmtrand.Str(6))
	app := kvstore.NewInMemoryApplication()
	server := newRemoteApp(t, sockPath, app)
	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Error(err)
		}
	})
	cfg := test.ResetTestRoot("mempool_test")
	return newMempoolWithAppAndConfig(proxy.NewRemoteClientCreator(sockPath, "socket", true), cfg)
}

// caller must close server.
func newRemoteApp(t *testing.T, addr string, app abci.Application) service.Service {
	t.Helper()
	_, err := abciclient.NewClient(addr, "socket", true)
	require.NoError(t, err)

	// Start server
	server := abciserver.NewSocketServer(addr, app)
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	if err := server.Start(); err != nil {
		t.Fatalf("Error starting socket server: %v", err.Error())
	}

	return server
}

func newReqRes(tx types.Tx, code uint32, requestType abci.CheckTxType) *abciclient.ReqRes {
	reqRes := abciclient.NewReqRes(abci.ToCheckTxRequest(&abci.CheckTxRequest{Tx: tx, Type: requestType}))
	reqRes.Response = abci.ToCheckTxResponse(&abci.CheckTxResponse{Code: code})
	return reqRes
}

func abciResponses(n int, code uint32) []*abci.ExecTxResult {
	responses := make([]*abci.ExecTxResult, 0, n)
	for i := 0; i < n; i++ {
		responses = append(responses, &abci.ExecTxResult{Code: code})
	}
	return responses
}

func doCommit(t require.TestingT, mp Mempool, app abci.Application, txs types.Txs, height int64) {
	rfb := &abci.FinalizeBlockRequest{Txs: make([][]byte, len(txs))}
	for i, tx := range txs {
		rfb.Txs[i] = tx
	}
	_, e := app.FinalizeBlock(context.Background(), rfb)
	require.NoError(t, e)
	mp.Lock()
	e = mp.FlushAppConn()
	require.NoError(t, e)
	_, e = app.Commit(context.Background(), &abci.CommitRequest{})
	require.NoError(t, e)
	e = mp.Update(height, txs, abciResponses(txs.Len(), abci.CodeTypeOK), nil, nil)
	require.NoError(t, e)
	mp.Unlock()
}
