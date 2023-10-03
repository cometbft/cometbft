package cat

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"
)

func TestStoreSimple(t *testing.T) {
	store := newStore()

	tx := types.Tx("tx1")
	key := tx.Key()
	wtx := newWrappedTx(tx, key, 1, 1, 1, "")

	// asset zero state
	require.Nil(t, store.get(key))
	require.False(t, store.has(key))
	require.False(t, store.remove(key))
	require.Zero(t, store.size())
	require.Zero(t, store.totalBytes())
	require.Empty(t, store.getAllKeys())
	require.Empty(t, store.getAllTxs())

	// add a tx
	store.set(wtx)
	require.True(t, store.has(key))
	require.Equal(t, wtx, store.get(key))
	require.Equal(t, int(1), store.size())
	require.Equal(t, wtx.size(), store.totalBytes())

	// remove a tx
	store.remove(key)
	require.False(t, store.has(key))
	require.Nil(t, store.get(key))
	require.Zero(t, store.size())
	require.Zero(t, store.totalBytes())
}

func TestStoreReservingTxs(t *testing.T) {
	store := newStore()

	tx := types.Tx("tx1")
	key := tx.Key()
	wtx := newWrappedTx(tx, key, 1, 1, 1, "")

	// asset zero state
	store.release(key)

	// reserve a tx
	store.reserve(key)
	require.True(t, store.has(key))
	// should not update the total bytes
	require.Zero(t, store.totalBytes())

	// should be able to add a tx
	store.set(wtx)
	require.Equal(t, tx, store.get(key).tx)
	require.Equal(t, wtx.size(), store.totalBytes())

	// releasing should do nothing on a set tx
	store.release(key)
	require.True(t, store.has(key))
	require.Equal(t, tx, store.get(key).tx)

	store.remove(key)
	require.False(t, store.has(key))

	// reserve the tx again
	store.reserve(key)
	require.True(t, store.has(key))

	// release should remove the tx
	store.release(key)
	require.False(t, store.has(key))
}

func TestStoreConcurrentAccess(t *testing.T) {
	store := newStore()

	numTxs := 100

	wg := &sync.WaitGroup{}
	for i := 0; i < numTxs; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ticker := time.NewTicker(10 * time.Millisecond)
			for range ticker.C {
				tx := types.Tx(fmt.Sprintf("tx%d", i%(numTxs/10)))
				key := tx.Key()
				wtx := newWrappedTx(tx, key, 1, 1, 1, "")
				existingTx := store.get(key)
				if existingTx != nil && bytes.Equal(existingTx.tx, tx) {
					// tx has already been added
					return
				}
				if store.reserve(key) {
					// some fail
					if i%3 == 0 {
						store.release(key)
						return
					}
					store.set(wtx)
					// this should be a noop
					store.release(key)
					return
				}
				// already reserved so we retry in 10 milliseconds
			}
		}(i)
	}
	wg.Wait()

	require.Equal(t, numTxs/10, store.size())
}

func TestStoreGetTxs(t *testing.T) {
	store := newStore()

	numTxs := 100
	for i := 0; i < numTxs; i++ {
		tx := types.Tx(fmt.Sprintf("tx%d", i))
		key := tx.Key()
		wtx := newWrappedTx(tx, key, 1, 1, int64(i), "")
		store.set(wtx)
	}

	require.Equal(t, numTxs, store.size())

	// get all txs
	txs := store.getAllTxs()
	require.Equal(t, numTxs, len(txs))

	// get txs by keys
	keys := store.getAllKeys()
	require.Equal(t, numTxs, len(keys))

	// get txs below a certain priority
	txs, bz := store.getTxsBelowPriority(int64(numTxs / 2))
	require.Equal(t, numTxs/2, len(txs))
	var actualBz int64
	for _, tx := range txs {
		actualBz += tx.size()
	}
	require.Equal(t, actualBz, bz)
}

func TestStoreExpiredTxs(t *testing.T) {
	store := newStore()
	numTxs := 100
	for i := 0; i < numTxs; i++ {
		tx := types.Tx(fmt.Sprintf("tx%d", i))
		key := tx.Key()
		wtx := newWrappedTx(tx, key, int64(i), 1, 1, "")
		store.set(wtx)
	}

	// half of them should get purged
	store.purgeExpiredTxs(int64(numTxs/2), time.Time{})

	remainingTxs := store.getAllTxs()
	require.Equal(t, numTxs/2, len(remainingTxs))
	for _, tx := range remainingTxs {
		require.GreaterOrEqual(t, tx.height, int64(numTxs/2))
	}

	store.purgeExpiredTxs(int64(0), time.Now().Add(time.Second))
	require.Empty(t, store.getAllTxs())
}
