package cat

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/types"
)

func TestSeenTxSet(t *testing.T) {
	var (
		tx1Key        = types.Tx("tx1").Key()
		tx2Key        = types.Tx("tx2").Key()
		tx3Key        = types.Tx("tx3").Key()
		peer1  uint16 = 1
		peer2  uint16 = 2
	)

	seenSet := NewSeenTxSet()
	require.Zero(t, seenSet.Pop(tx1Key))

	seenSet.Add(tx1Key, peer1)
	seenSet.Add(tx1Key, peer1)
	require.Equal(t, 1, seenSet.Len())
	seenSet.Add(tx1Key, peer2)
	peers := seenSet.Get(tx1Key)
	require.NotNil(t, peers)
	require.Equal(t, map[uint16]struct{}{peer1: {}, peer2: {}}, peers)
	seenSet.Add(tx2Key, peer1)
	seenSet.Add(tx3Key, peer1)
	require.Equal(t, 3, seenSet.Len())
	seenSet.RemoveKey(tx2Key)
	require.Equal(t, 2, seenSet.Len())
	require.Zero(t, seenSet.Pop(tx2Key))
	require.Equal(t, peer1, seenSet.Pop(tx3Key))
}

func TestLRUTxCacheRemove(t *testing.T) {
	cache := NewLRUTxCache(100)
	numTxs := 10

	txs := make([][32]byte, numTxs)
	for i := 0; i < numTxs; i++ {
		// probability of collision is 2**-256
		txBytes := make([]byte, 32)
		_, err := rand.Read(txBytes)
		require.NoError(t, err)

		copy(txs[i][:], txBytes)
		cache.Push(txs[i])

		// make sure its added to both the linked list and the map
		require.Equal(t, i+1, cache.list.Len())
	}

	for i := 0; i < numTxs; i++ {
		cache.Remove(txs[i])
		// make sure its removed from both the map and the linked list
		require.Equal(t, numTxs-(i+1), cache.list.Len())
	}
}

func TestLRUTxCacheSize(t *testing.T) {
	const size = 10
	cache := NewLRUTxCache(size)

	for i := 0; i < size*2; i++ {
		tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
		cache.Push(tx.Key())
		require.Less(t, cache.list.Len(), size+1)
	}
}

func TestSeenTxSetConcurrency(t *testing.T) {
	seenSet := NewSeenTxSet()

	const (
		concurrency = 10
		numTx       = 100
	)

	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(peer uint16) {
			defer wg.Done()
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.Add(tx.Key(), peer)
			}
		}(uint16(i % 2))
	}
	time.Sleep(time.Millisecond)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(peer uint16) {
			defer wg.Done()
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.Has(tx.Key(), peer)
			}
		}(uint16(i % 2))
	}
	time.Sleep(time.Millisecond)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(peer uint16) {
			defer wg.Done()
			for i := numTx - 1; i >= 0; i-- {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				seenSet.RemoveKey(tx.Key())
			}
		}(uint16(i % 2))
	}
	wg.Wait()
}

func TestLRUTxCacheConcurrency(t *testing.T) {
	cache := NewLRUTxCache(100)

	const (
		concurrency = 10
		numTx       = 100
	)

	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				cache.Push(tx.Key())
			}
			for i := 0; i < numTx; i++ {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				cache.Has(tx.Key())
			}
			for i := numTx - 1; i >= 0; i-- {
				tx := types.Tx([]byte(fmt.Sprintf("tx%d", i)))
				cache.Remove(tx.Key())
			}
		}()
	}
	wg.Wait()
}
