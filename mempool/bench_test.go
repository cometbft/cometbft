package mempool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/proxy"
)

func BenchmarkReap(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mp.config.Size = 100_000_000 // so that the mempool never saturates
	addTxs(b, mp, 0, 10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.ReapMaxBytesMaxGas(100_000_000, -1)
	}
}

func BenchmarkCheckTx(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mp.config.Size = 100_000_000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tx := kvstore.NewTxFromID(i)
		b.StartTimer()

		rr, err := mp.CheckTx(tx, "")
		require.NoError(b, err, i)
		rr.Wait()
	}
}

func BenchmarkParallelCheckTx(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mp.config.Size = 100_000_000
	var txcnt uint64
	next := func() uint64 {
		return atomic.AddUint64(&txcnt, 1)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tx := kvstore.NewTxFromID(int(next()))
			rr, err := mp.CheckTx(tx, "")
			require.NoError(b, err, tx)
			rr.Wait()
		}
	})
}

func BenchmarkCheckDuplicateTx(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mp.config.Size = 2

	tx := kvstore.NewTxFromID(1)
	if _, err := mp.CheckTx(tx, ""); err != nil {
		b.Fatal(err)
	}
	err := mp.FlushAppConn()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mp.CheckTx(tx, "")
		require.ErrorAs(b, err, &ErrTxInCache, "tx should be duplicate")
	}
}

func BenchmarkUpdate(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	numTxs := 1000
	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		b.StopTimer()
		txs := addTxs(b, mp, i*numTxs, numTxs)
		require.Equal(b, numTxs, len(txs))
		require.Equal(b, numTxs, mp.Size())
		b.StartTimer()

		doUpdate(b, mp, int64(i), txs)
		require.Zero(b, mp.Size())
	}
}

func BenchmarkUpdateAndRecheck(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	numTxs := 1000
	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		b.StopTimer()
		mp.Flush()
		txs := addTxs(b, mp, 0, numTxs)
		require.Equal(b, numTxs, len(txs))
		require.Equal(b, numTxs, mp.Size())
		b.StartTimer()

		// Update a part of txs and recheck the rest.
		doUpdate(b, mp, int64(i), txs[:numTxs/2])
	}
}

func BenchmarkUpdateRemoteClient(b *testing.B) {
	mp, cleanup := newMempoolWithAsyncConnection(b)
	defer cleanup()

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		b.StopTimer()
		tx := kvstore.NewTxFromID(i)
		_, err := mp.CheckTx(tx, "")
		require.NoError(b, err)
		err = mp.FlushAppConn()
		require.NoError(b, err)
		require.Equal(b, 1, mp.Size())
		b.StartTimer()

		txs := mp.ReapMaxTxs(mp.Size())
		doUpdate(b, mp, int64(i), txs)
	}
}

// Benchmarks the time it takes an iterator to access all transactions in the
// mempool.
func BenchmarkIterator(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	const numTxs = 1000
	txs := addTxs(b, mp, 0, numTxs)
	require.Equal(b, numTxs, len(txs))
	require.Equal(b, numTxs, mp.Size())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		iter := mp.NewIterator(context.TODO())
		b.StartTimer()

		// Iterate until all txs in the mempool are accessed.
		for c := 0; c < numTxs; c++ {
			if entry := <-iter.WaitNextCh(); entry == nil {
				continue
			}
		}
	}
}

// Benchmarks the time it takes multiple concurrent iterators to access all
// transactions in the mempool.
func BenchmarkConcurrentkIterators(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	const numTxs = 1000
	const numIterators = 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Flush mempool and add a bunch of txs.
		mp.Flush()
		txs := addTxs(b, mp, 0, numTxs)
		require.Equal(b, numTxs, len(txs))
		require.Equal(b, numTxs, mp.Size())
		// Create concurrent iterators.
		iters := make([]Iterator, numIterators)
		for j := 0; j < numIterators; j++ {
			iters[j] = mp.NewIterator(context.TODO())
		}
		wg := sync.WaitGroup{}
		wg.Add(numIterators)
		b.StartTimer()

		for j := 0; j < numIterators; j++ {
			go func(iter Iterator) {
				defer wg.Done()
				// Iterate until all txs in the mempool are accessed.
				for c := 0; c < numTxs; c++ {
					if entry := <-iter.WaitNextCh(); entry == nil {
						continue
					}
				}
			}(iters[j])
		}

		wg.Wait()
	}
}
