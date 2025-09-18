package mempool

import (
	"encoding/binary"
	"testing"
)

func BenchmarkCacheInsertTime(b *testing.B) {
	cache := NewLRUTxCache(b.N)

	txs := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		txs[i] = make([]byte, 8)
		binary.BigEndian.PutUint64(txs[i], uint64(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Push(txs[i])
	}
}

func BenchmarkCacheRemoveTime(b *testing.B) {
	cache := NewLRUTxCache(b.N)

	txs := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		txs[i] = make([]byte, 8)
		binary.BigEndian.PutUint64(txs[i], uint64(i))
		cache.Push(txs[i])
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			cache.Remove(txs[idx%b.N])
			idx++
		}
	})
}
