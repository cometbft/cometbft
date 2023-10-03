package cat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/example/kvstore"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
)

// application extends the KV store application by overriding CheckTx to provide
// transaction priority based on the value in the key/value pair.
type application struct {
	*kvstore.Application
}

type testTx struct {
	tx       types.Tx
	priority int64
}

func newTx(i int, peerID uint16, msg []byte, priority int64) []byte {
	return []byte(fmt.Sprintf("sender-%d-%d=%X=%d", i, peerID, msg, priority))
}

func newDefaultTx(msg string) types.Tx {
	return types.Tx(newTx(0, 0, []byte(msg), 1))
}

func (app *application) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	var (
		priority int64
		sender   string
	)

	// infer the priority from the raw transaction value (sender=key=value)
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) == 3 {
		v, err := strconv.ParseInt(string(parts[2]), 10, 64)
		if err != nil {
			return abci.ResponseCheckTx{
				Priority:  priority,
				Code:      100,
				GasWanted: 1,
			}
		}

		priority = v
		sender = string(parts[0])
	} else {
		return abci.ResponseCheckTx{
			Priority:  priority,
			Code:      101,
			GasWanted: 1,
		}
	}

	return abci.ResponseCheckTx{
		Priority:  priority,
		Sender:    sender,
		Code:      code.CodeTypeOK,
		GasWanted: 1,
	}
}

func setup(t testing.TB, cacheSize int, options ...TxPoolOption) *TxPool {
	t.Helper()

	app := &application{kvstore.NewApplication()}
	cc := proxy.NewLocalClientCreator(app)

	cfg := config.TestMempoolConfig()
	cfg.CacheSize = cacheSize

	appConnMem, err := cc.NewABCIClient()
	require.NoError(t, err)
	require.NoError(t, appConnMem.Start())

	t.Cleanup(func() {
		os.RemoveAll(cfg.RootDir)
		require.NoError(t, appConnMem.Stop())
	})

	return NewTxPool(log.TestingLogger().With("test", t.Name()), cfg, appConnMem, 1, options...)
}

// mustCheckTx invokes txmp.CheckTx for the given transaction and waits until
// its callback has finished executing. It fails t if CheckTx fails.
func mustCheckTx(t *testing.T, txmp *TxPool, spec string) {
	require.NoError(t, txmp.CheckTx([]byte(spec), nil, mempool.TxInfo{}))
}

func checkTxs(t *testing.T, txmp *TxPool, numTxs int, peerID uint16) []testTx {
	txs := make([]testTx, numTxs)
	txInfo := mempool.TxInfo{SenderID: peerID}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	current := txmp.Size()
	for i := 0; i < numTxs; i++ {
		prefix := make([]byte, 20)
		_, err := rng.Read(prefix)
		require.NoError(t, err)

		priority := int64(rng.Intn(9999-1000) + 1000)

		txs[i] = testTx{
			tx:       newTx(i, peerID, prefix, priority),
			priority: priority,
		}
		require.NoError(t, txmp.CheckTx(txs[i].tx, nil, txInfo))
		// assert that none of them get silently evicted
		require.Equal(t, current+i+1, txmp.Size())
	}

	return txs
}

func TestTxPool_TxsAvailable(t *testing.T) {
	txmp := setup(t, 0)
	txmp.EnableTxsAvailable()

	ensureNoTxFire := func() {
		timer := time.NewTimer(500 * time.Millisecond)
		select {
		case <-txmp.TxsAvailable():
			require.Fail(t, "unexpected transactions event")
		case <-timer.C:
		}
	}

	ensureTxFire := func() {
		timer := time.NewTimer(500 * time.Millisecond)
		select {
		case <-txmp.TxsAvailable():
		case <-timer.C:
			require.Fail(t, "expected transactions event")
		}
	}

	// ensure no event as we have not executed any transactions yet
	ensureNoTxFire()

	// Execute CheckTx for some transactions and ensure TxsAvailable only fires
	// once.
	txs := checkTxs(t, txmp, 100, 0)
	ensureTxFire()
	ensureNoTxFire()

	rawTxs := make([]types.Tx, len(txs))
	for i, tx := range txs {
		rawTxs[i] = tx.tx
	}

	responses := make([]*abci.ResponseDeliverTx, len(rawTxs[:50]))
	for i := 0; i < len(responses); i++ {
		responses[i] = &abci.ResponseDeliverTx{Code: abci.CodeTypeOK}
	}

	require.Equal(t, 100, txmp.Size())

	// commit half the transactions and ensure we fire an event
	txmp.Lock()
	require.NoError(t, txmp.Update(1, rawTxs[:50], responses, nil, nil))
	txmp.Unlock()
	ensureTxFire()
	ensureNoTxFire()

	// Execute CheckTx for more transactions and ensure we do not fire another
	// event as we're still on the same height (1).
	_ = checkTxs(t, txmp, 100, 0)
	ensureNoTxFire()
}

func TestTxPool_Size(t *testing.T) {
	txmp := setup(t, 0)
	txs := checkTxs(t, txmp, 100, 0)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())

	rawTxs := make([]types.Tx, len(txs))
	for i, tx := range txs {
		rawTxs[i] = tx.tx
	}

	responses := make([]*abci.ResponseDeliverTx, len(rawTxs[:50]))
	for i := 0; i < len(responses); i++ {
		responses[i] = &abci.ResponseDeliverTx{Code: abci.CodeTypeOK}
	}

	txmp.Lock()
	require.NoError(t, txmp.Update(1, rawTxs[:50], responses, nil, nil))
	txmp.Unlock()

	require.Equal(t, len(rawTxs)/2, txmp.Size())
	require.Equal(t, int64(2850), txmp.SizeBytes())
}

func TestTxPool_Eviction(t *testing.T) {
	txmp := setup(t, 1000)
	txmp.config.Size = 5
	txmp.config.MaxTxsBytes = 60
	txExists := func(spec string) bool {
		return txmp.Has(types.Tx(spec).Key())
	}

	// A transaction bigger than the mempool should be rejected even when there
	// are slots available.
	err := txmp.CheckTx(types.Tx("big=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef=1"), nil, mempool.TxInfo{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "mempool is full")
	require.Equal(t, 0, txmp.Size())

	// Nearly-fill the mempool with a low-priority transaction, to show that it
	// is evicted even when slots are available for a higher-priority tx.
	const bigTx = "big=0123456789abcdef0123456789abcdef0123456789abcdef01234=2"
	mustCheckTx(t, txmp, bigTx)
	require.Equal(t, 1, txmp.Size()) // bigTx is the only element
	require.True(t, txExists(bigTx))
	require.Equal(t, int64(len(bigTx)), txmp.SizeBytes())

	// The next transaction should evict bigTx, because it is higher priority
	// but does not fit on size.
	mustCheckTx(t, txmp, "key1=0000=25")
	require.True(t, txExists("key1=0000=25"))
	require.False(t, txExists(bigTx))
	require.Equal(t, int64(len("key1=0000=25")), txmp.SizeBytes())

	// Now fill up the rest of the slots with other transactions.
	mustCheckTx(t, txmp, "key2=0001=5")
	mustCheckTx(t, txmp, "key3=0002=10")
	mustCheckTx(t, txmp, "key4=0003=3")
	mustCheckTx(t, txmp, "key5=0004=3")

	// A new transaction with low priority should be discarded.
	err = txmp.CheckTx(types.Tx("key6=0005=1"), nil, mempool.TxInfo{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "mempool is full")
	require.False(t, txExists("key6=0005=1"))

	// A new transaction with higher priority should evict key5, which is the
	// newest of the two transactions with lowest priority.
	mustCheckTx(t, txmp, "key7=0006=7")
	require.True(t, txExists("key7=0006=7"))  // new transaction added
	require.False(t, txExists("key5=0004=3")) // newest low-priority tx evicted
	require.True(t, txExists("key4=0003=3"))  // older low-priority tx retained

	// Another new transaction evicts the other low-priority element.
	mustCheckTx(t, txmp, "key8=0007=20")
	require.True(t, txExists("key8=0007=20"))
	require.False(t, txExists("key4=0003=3"))

	// Now the lowest-priority tx is 5, so that should be the next to go.
	mustCheckTx(t, txmp, "key9=0008=9")
	require.True(t, txExists("key9=0008=9"))
	require.False(t, txExists("key2=0001=5"))

	// Add a transaction that requires eviction of multiple lower-priority
	// entries, in order to fit the size of the element.
	mustCheckTx(t, txmp, "key10=0123456789abcdef=11") // evict 10, 9, 7; keep 25, 20, 11
	require.True(t, txExists("key1=0000=25"))
	require.True(t, txExists("key8=0007=20"))
	require.True(t, txExists("key10=0123456789abcdef=11"))
	require.False(t, txExists("key3=0002=10"))
	require.False(t, txExists("key9=0008=9"))
	require.False(t, txExists("key7=0006=7"))

	// Free up some space so we can add back previously evicted txs
	err = txmp.Update(1, types.Txs{types.Tx("key10=0123456789abcdef=11")}, []*abci.ResponseDeliverTx{{Code: abci.CodeTypeOK}}, nil, nil)
	require.NoError(t, err)
	require.False(t, txExists("key10=0123456789abcdef=11"))
	mustCheckTx(t, txmp, "key3=0002=10")
	require.True(t, txExists("key3=0002=10"))

	// remove a high priority tx and check if there is
	// space for the previously evicted tx
	require.NoError(t, txmp.RemoveTxByKey(types.Tx("key8=0007=20").Key()))
	require.False(t, txExists("key8=0007=20"))
}

func TestTxPool_Flush(t *testing.T) {
	txmp := setup(t, 0)
	txs := checkTxs(t, txmp, 100, 0)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())

	rawTxs := make([]types.Tx, len(txs))
	for i, tx := range txs {
		rawTxs[i] = tx.tx
	}

	responses := make([]*abci.ResponseDeliverTx, len(rawTxs[:50]))
	for i := 0; i < len(responses); i++ {
		responses[i] = &abci.ResponseDeliverTx{Code: abci.CodeTypeOK}
	}

	txmp.Lock()
	require.NoError(t, txmp.Update(1, rawTxs[:50], responses, nil, nil))
	txmp.Unlock()

	txmp.Flush()
	require.Zero(t, txmp.Size())
	require.Equal(t, int64(0), txmp.SizeBytes())
}

func TestTxPool_ReapMaxBytesMaxGas(t *testing.T) {
	txmp := setup(t, 0)
	tTxs := checkTxs(t, txmp, 100, 0) // all txs request 1 gas unit
	require.Equal(t, len(tTxs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())

	txMap := make(map[types.TxKey]testTx)
	priorities := make([]int64, len(tTxs))
	for i, tTx := range tTxs {
		txMap[tTx.tx.Key()] = tTx
		priorities[i] = tTx.priority
	}

	sort.Slice(priorities, func(i, j int) bool {
		// sort by priority, i.e. decreasing order
		return priorities[i] > priorities[j]
	})

	ensurePrioritized := func(reapedTxs types.Txs) {
		reapedPriorities := make([]int64, len(reapedTxs))
		for i, rTx := range reapedTxs {
			reapedPriorities[i] = txMap[rTx.Key()].priority
		}

		require.Equal(t, priorities[:len(reapedPriorities)], reapedPriorities)
	}

	// reap by gas capacity only
	reapedTxs := txmp.ReapMaxBytesMaxGas(-1, 50)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(tTxs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	require.Len(t, reapedTxs, 50)

	// reap by transaction bytes only
	reapedTxs = txmp.ReapMaxBytesMaxGas(1200, -1)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(tTxs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	// each tx is 57 bytes, 20 * 57 = 1140 + overhead for proto encoding
	require.Equal(t, len(reapedTxs), 20)

	// Reap by both transaction bytes and gas, where the size yields 31 reaped
	// transactions and the gas limit reaps 25 transactions.
	reapedTxs = txmp.ReapMaxBytesMaxGas(2000, 25)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(tTxs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	require.Len(t, reapedTxs, 25)
}

func TestTxPool_ReapMaxTxs(t *testing.T) {
	txmp := setup(t, 0)
	txs := checkTxs(t, txmp, 100, 0)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())

	txMap := make(map[types.TxKey]int64)
	for _, tx := range txs {
		txMap[tx.tx.Key()] = tx.priority
	}

	ensurePrioritized := func(reapedTxs types.Txs) {
		for i := 0; i < len(reapedTxs)-1; i++ {
			currPriority := txMap[reapedTxs[i].Key()]
			nextPriority := txMap[reapedTxs[i+1].Key()]
			require.GreaterOrEqual(t, currPriority, nextPriority)
		}
	}

	// reap all transactions
	reapedTxs := txmp.ReapMaxTxs(-1)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	require.Len(t, reapedTxs, len(txs))

	// reap a single transaction
	reapedTxs = txmp.ReapMaxTxs(1)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	require.Len(t, reapedTxs, 1)

	// reap half of the transactions
	reapedTxs = txmp.ReapMaxTxs(len(txs) / 2)
	ensurePrioritized(reapedTxs)
	require.Equal(t, len(txs), txmp.Size())
	require.Equal(t, int64(5690), txmp.SizeBytes())
	require.Len(t, reapedTxs, len(txs)/2)
}

func TestTxPool_CheckTxExceedsMaxSize(t *testing.T) {
	txmp := setup(t, 0)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx := make([]byte, txmp.config.MaxTxBytes+1)
	_, err := rng.Read(tx)
	require.NoError(t, err)

	err = txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: 0})
	require.Equal(t, mempool.ErrTxTooLarge{Max: txmp.config.MaxTxBytes, Actual: len(tx)}, err)

	tx = make([]byte, txmp.config.MaxTxBytes-1)
	_, err = rng.Read(tx)
	require.NoError(t, err)

	err = txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: 0})
	require.NotEqual(t, mempool.ErrTxTooLarge{Max: txmp.config.MaxTxBytes, Actual: len(tx)}, err)
}

func TestTxPool_CheckTxSamePeer(t *testing.T) {
	txmp := setup(t, 100)
	peerID := uint16(1)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	prefix := make([]byte, 20)
	_, err := rng.Read(prefix)
	require.NoError(t, err)

	tx := []byte(fmt.Sprintf("sender-0=%X=%d", prefix, 50))

	require.NoError(t, txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: peerID}))
	require.Error(t, txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: peerID}))
}

func TestTxPool_ConcurrentTxs(t *testing.T) {
	txmp := setup(t, 100)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	checkTxDone := make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for i := 0; i < 20; i++ {
			_ = checkTxs(t, txmp, 100, 0)
			dur := rng.Intn(1000-500) + 500
			time.Sleep(time.Duration(dur) * time.Millisecond)
		}

		wg.Done()
		close(checkTxDone)
	}()

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		defer wg.Done()

		var height int64 = 1

		for range ticker.C {
			reapedTxs := txmp.ReapMaxTxs(200)
			if len(reapedTxs) > 0 {
				responses := make([]*abci.ResponseDeliverTx, len(reapedTxs))
				for i := 0; i < len(responses); i++ {
					var code uint32

					if i%10 == 0 {
						code = 100
					} else {
						code = abci.CodeTypeOK
					}

					responses[i] = &abci.ResponseDeliverTx{Code: code}
				}

				txmp.Lock()
				require.NoError(t, txmp.Update(height, reapedTxs, responses, nil, nil))
				txmp.Unlock()

				height++
			} else {
				// only return once we know we finished the CheckTx loop
				select {
				case <-checkTxDone:
					return
				default:
				}
			}
		}
	}()

	wg.Wait()
	require.Zero(t, txmp.Size())
	require.Zero(t, txmp.SizeBytes())
}

func TestTxPool_ExpiredTxs_Timestamp(t *testing.T) {
	txmp := setup(t, 5000)
	txmp.config.TTLDuration = 5 * time.Millisecond

	added1 := checkTxs(t, txmp, 10, 0)
	require.Equal(t, len(added1), txmp.Size())

	// Wait a while, then add some more transactions that should not be expired
	// when the first batch TTLs out.
	//
	// ms: 0   1   2   3   4   5   6
	//     ^           ^       ^   ^
	//     |           |       |   +-- Update (triggers pruning)
	//     |           |       +------ first batch expires
	//     |           +-------------- second batch added
	//     +-------------------------- first batch added
	//
	// The exact intervals are not important except that the delta should be
	// large relative to the cost of CheckTx (ms vs. ns is fine here).
	time.Sleep(3 * time.Millisecond)
	added2 := checkTxs(t, txmp, 10, 1)

	// Wait a while longer, so that the first batch will expire.
	time.Sleep(3 * time.Millisecond)

	// Trigger an update so that pruning will occur.
	txmp.Lock()
	defer txmp.Unlock()
	require.NoError(t, txmp.Update(txmp.height+1, nil, nil, nil, nil))

	// All the transactions in the original set should have been purged.
	for _, tx := range added1 {
		if txmp.store.has(tx.tx.Key()) {
			t.Errorf("Transaction %X should have been purged for TTL", tx.tx.Key())
		}
		if txmp.rejectedTxCache.Has(tx.tx.Key()) {
			t.Errorf("Transaction %X should have been removed from the cache", tx.tx.Key())
		}
	}

	// All the transactions added later should still be around.
	for _, tx := range added2 {
		if !txmp.store.has(tx.tx.Key()) {
			t.Errorf("Transaction %X should still be in the mempool, but is not", tx.tx.Key())
		}
	}
}

func TestTxPool_ExpiredTxs_NumBlocks(t *testing.T) {
	txmp := setup(t, 500)
	txmp.height = 100
	txmp.config.TTLNumBlocks = 10

	tTxs := checkTxs(t, txmp, 100, 0)
	require.Equal(t, len(tTxs), txmp.Size())

	// reap 5 txs at the next height -- no txs should expire
	reapedTxs := txmp.ReapMaxTxs(5)
	responses := make([]*abci.ResponseDeliverTx, len(reapedTxs))
	for i := 0; i < len(responses); i++ {
		responses[i] = &abci.ResponseDeliverTx{Code: abci.CodeTypeOK}
	}

	txmp.Lock()
	require.NoError(t, txmp.Update(txmp.height+1, reapedTxs, responses, nil, nil))
	txmp.Unlock()

	require.Equal(t, 95, txmp.Size())

	// check more txs at height 101
	_ = checkTxs(t, txmp, 50, 1)
	require.Equal(t, 145, txmp.Size())

	// Reap 5 txs at a height that would expire all the transactions from before
	// the previous Update (height 100).
	//
	// NOTE: When we reap txs below, we do not know if we're picking txs from the
	// initial CheckTx calls or from the second round of CheckTx calls. Thus, we
	// cannot guarantee that all 95 txs are remaining that should be expired and
	// removed. However, we do know that that at most 95 txs can be expired and
	// removed.
	reapedTxs = txmp.ReapMaxTxs(5)
	responses = make([]*abci.ResponseDeliverTx, len(reapedTxs))
	for i := 0; i < len(responses); i++ {
		responses[i] = &abci.ResponseDeliverTx{Code: abci.CodeTypeOK}
	}

	txmp.Lock()
	require.NoError(t, txmp.Update(txmp.height+10, reapedTxs, responses, nil, nil))
	txmp.Unlock()

	require.GreaterOrEqual(t, txmp.Size(), 45)
}

func TestTxPool_CheckTxPostCheckError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{
			name: "error",
			err:  errors.New("test error"),
		},
		{
			name: "no error",
			err:  nil,
		},
	}
	for _, tc := range cases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			postCheckFn := func(_ types.Tx, _ *abci.ResponseCheckTx) error {
				return testCase.err
			}
			txmp := setup(t, 0, WithPostCheck(postCheckFn))
			tx := []byte("sender=0000=1")
			err := txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: 0})
			require.True(t, errors.Is(err, testCase.err))
		})
	}
}

// func TestTxPool_RemoveBlobTx(t *testing.T) {
// 	app := kvstore.NewApplication()
// 	cc := proxy.NewLocalClientCreator(app)

// 	cfg := config.TestMempoolConfig()
// 	cfg.CacheSize = 100

// 	appConnMem, err := cc.NewABCIClient()
// 	require.NoError(t, err)
// 	require.NoError(t, appConnMem.Start())

// 	t.Cleanup(func() {
// 		os.RemoveAll(cfg.RootDir)
// 		require.NoError(t, appConnMem.Stop())
// 	})

// 	txmp := NewTxPool(log.TestingLogger(), cfg, appConnMem, 1)

// 	originalTx := []byte{1, 2, 3, 4}
// 	indexWrapper, err := types.MarshalIndexWrapper(originalTx, 100)
// 	require.NoError(t, err)
// 	namespaceOne := bytes.Repeat([]byte{1}, consts.NamespaceIDSize)

// 	// create the blobTx
// 	b := tmproto.Blob{
// 		NamespaceId:      namespaceOne,
// 		Data:             []byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 9},
// 		ShareVersion:     0,
// 		NamespaceVersion: 0,
// 	}
// 	bTx, err := types.MarshalBlobTx(originalTx, &b)
// 	require.NoError(t, err)

// 	err = txmp.CheckTx(bTx, nil, mempool.TxInfo{})
// 	require.NoError(t, err)

// 	err = txmp.Update(1, []types.Tx{indexWrapper}, abciResponses(1, abci.CodeTypeOK), nil, nil)
// 	require.NoError(t, err)
// 	require.EqualValues(t, 0, txmp.Size())
// 	require.EqualValues(t, 0, txmp.SizeBytes())
// }

func abciResponses(n int, code uint32) []*abci.ResponseDeliverTx {
	responses := make([]*abci.ResponseDeliverTx, 0, n)
	for i := 0; i < n; i++ {
		responses = append(responses, &abci.ResponseDeliverTx{Code: code})
	}
	return responses
}

func TestTxPool_ConcurrentlyAddingTx(t *testing.T) {
	txmp := setup(t, 500)
	tx := types.Tx("sender=0000=1")

	numTxs := 10
	errCh := make(chan error, numTxs)
	wg := &sync.WaitGroup{}
	for i := 0; i < numTxs; i++ {
		wg.Add(1)
		go func(sender uint16) {
			defer wg.Done()
			_, err := txmp.TryAddNewTx(tx, tx.Key(), mempool.TxInfo{SenderID: sender})
			errCh <- err
		}(uint16(i + 1))
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()

	errCount := 0
	for err := range errCh {
		if err != nil {
			require.Equal(t, ErrTxInMempool, err)
			errCount++
		}
	}
	require.Equal(t, numTxs-1, errCount)
}

func TestTxPool_BroadcastQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	txmp := setup(t, 1)
	txs := 10

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < txs; i++ {
			select {
			case <-ctx.Done():
				assert.FailNowf(t, "failed to receive all txs (got %d/%d)", "", i+1, txs)
			case wtx := <-txmp.next():
				require.Equal(t, wtx.tx, newDefaultTx(fmt.Sprintf("%d", i)))
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for i := 0; i < txs; i++ {
		tx := newDefaultTx(fmt.Sprintf("%d", i))
		require.NoError(t, txmp.CheckTx(tx, nil, mempool.TxInfo{SenderID: 0}))
	}

	wg.Wait()
}
