package mempool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

func TestCacheRemove(t *testing.T) {
	cache := NewLRUTxCache(100)
	tx := types.Tx([]byte{0x01})
	txKey := tx.Key()

	cache.Push(tx)
	assert.True(t, cache.Has(tx))

	cache.Remove(tx)
	assert.False(t, cache.Has(tx))
	assert.Nil(t, cache.cacheMap[txKey])
}

func TestCacheAfterUpdate(t *testing.T) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	// reAddIndices & txsInCache can have elements > numTxsToCreate
	// also assumes max index is 255 for convenience
	// txs in cache also checks order of elements
	tests := []struct {
		name          string
		numTxsToAdd   int
		updateIndices []int
		txsInCache    []int
	}{
		{"empty mempool", 0, nil, nil},
		{"remove from middle", 5, []int{1, 2, 3}, []int{0, 4}},
		{"remove and readd", 5, []int{1, 2, 1, 2}, []int{0, 1, 2, 3, 4}},
		{"update all", 5, []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.numTxsToAdd; i++ {
				tx := types.Tx{byte(i)}
				_, err := mp.CheckTx(tx, "")
				require.NoError(t, err)
			}

			txs := make(types.Txs, len(tt.updateIndices))
			for i, idx := range tt.updateIndices {
				txs[i] = types.Tx{byte(idx)}
			}
			mp.Lock()
			err := mp.Update(1, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
			require.NoError(t, err)
			mp.Unlock()

			for _, idx := range tt.txsInCache {
				require.True(t, mp.cache.Has(types.Tx{byte(idx)}), "Tx %d expected to be in cache", idx)
			}
		})
	}
}

func TestStatsLRUTxCache(t *testing.T) {
	cache := NewStatsLRUTxCache(10)

	// Test initial state
	assert.Equal(t, uint64(0), cache.Hits())
	assert.Equal(t, uint64(0), cache.Misses())
	assert.Equal(t, uint64(0), cache.Evictions())
	assert.Equal(t, 0, cache.Size())

	// Test adding a transaction
	tx1 := types.Tx([]byte{0x01})
	added := cache.Push(tx1)
	assert.True(t, added)
	assert.Equal(t, uint64(0), cache.Hits())
	assert.Equal(t, uint64(1), cache.Misses())
	assert.Equal(t, uint64(0), cache.Evictions())
	assert.Equal(t, 1, cache.Size())

	// Test cache hit
	added = cache.Push(tx1)
	assert.False(t, added)
	assert.Equal(t, uint64(1), cache.Hits())
	assert.Equal(t, uint64(1), cache.Misses())

	// Test Has with cache hit
	has := cache.Has(tx1)
	assert.True(t, has)
	assert.Equal(t, uint64(2), cache.Hits())
	assert.Equal(t, uint64(1), cache.Misses())

	// Test Has with cache miss
	tx2 := types.Tx([]byte{0x02})
	has = cache.Has(tx2)
	assert.False(t, has)
	assert.Equal(t, uint64(2), cache.Hits())
	assert.Equal(t, uint64(2), cache.Misses())

	// Test eviction
	cache = NewStatsLRUTxCache(2)
	tx1 = types.Tx([]byte{0x01})
	tx2 = types.Tx([]byte{0x02})
	tx3 := types.Tx([]byte{0x03})

	cache.Push(tx1)
	cache.Push(tx2)
	assert.Equal(t, uint64(0), cache.Evictions())

	// This should evict tx1
	cache.Push(tx3)
	assert.Equal(t, uint64(1), cache.Evictions())
	assert.False(t, cache.Has(tx1))
	assert.True(t, cache.Has(tx2))
	assert.True(t, cache.Has(tx3))

	// Test Reset
	cache.Reset()
	assert.Equal(t, uint64(0), cache.Hits())
	assert.Equal(t, uint64(0), cache.Misses())
	assert.Equal(t, uint64(0), cache.Evictions())
	assert.Equal(t, 0, cache.Size())
	assert.False(t, cache.Has(tx2))
	assert.False(t, cache.Has(tx3))
}

func TestStatsLRUTxCacheRemove(t *testing.T) {
	cache := NewStatsLRUTxCache(100)
	tx := types.Tx([]byte{0x01})
	txKey := tx.Key()

	cache.Push(tx)
	assert.True(t, cache.Has(tx))

	cache.Remove(tx)
	assert.False(t, cache.Has(tx))
	assert.Nil(t, cache.cacheMap[txKey])
}
