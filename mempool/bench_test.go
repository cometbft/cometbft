package mempool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abciserver "github.com/cometbft/cometbft/abci/server"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/proxy"
)

func BenchmarkReap(b *testing.B) {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	mp, cleanup := newMempoolWithApp(cc)
	defer cleanup()

	mp.config.Size = 100_000_000 // so that the nmempool never saturates

	size := 10000
	for i := 0; i < size; i++ {
		tx := kvstore.NewTxFromID(i)
		if _, err := mp.CheckTx(tx, ""); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.ReapMaxBytesMaxGas(100000000, 10000000)
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

		if _, err := mp.CheckTx(tx, ""); err != nil {
			b.Fatal(err)
		}
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
			if _, err := mp.CheckTx(tx, ""); err != nil {
				b.Fatal(err)
			}
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
	require.NotErrorIs(b, nil, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mp.CheckTx(tx, ""); err == nil {
			b.Fatal("tx should be duplicate")
		}
	}
}

func BenchmarkUpdateRemoteClient(b *testing.B) {
	sockPath := fmt.Sprintf("unix:///tmp/echo_%v.sock", cmtrand.Str(6))
	app := kvstore.NewInMemoryApplication()

	// Start server
	server := abciserver.NewSocketServer(sockPath, app)
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	if err := server.Start(); err != nil {
		b.Fatalf("Error starting socket server: %v", err.Error())
	}

	b.Cleanup(func() {
		if err := server.Stop(); err != nil {
			b.Error(err)
		}
	})
	cfg := test.ResetTestRoot("mempool_test")
	mp, cleanup := newMempoolWithAppAndConfig(proxy.NewRemoteClientCreator(sockPath, "socket", true), cfg)
	defer cleanup()

	b.ResetTimer()
	for i := 1; i <= b.N; i++ {
		tx := kvstore.NewTxFromID(i)

		_, err := mp.CheckTx(tx, "")
		require.NoError(b, err)

		err = mp.FlushAppConn()
		require.NoError(b, err)

		require.Equal(b, 1, mp.Size())

		txs := mp.ReapMaxTxs(mp.Size())
		doCommit(b, mp, app, txs, int64(i))
		assert.True(b, true)
	}
}

// Test adding transactions while a concurrent routine reaps txs and updates the mempool, simulating
// the consensus module, when using an async ABCI client.
func BenchmarkUpdateWithConcurrentCheckTx(b *testing.B) {
	mp, cleanup := newMempoolWithAsyncConnection(b)
	defer cleanup()

	numTxs := 1000
	maxHeight := 1000
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Add some txs to mempool.
	for i := 1; i <= numTxs; i++ {
		rr, err := mp.CheckTx(kvstore.NewTxFromID(i), "")
		require.NoError(b, err)
		rr.Wait()
	}

	// A process that continuously reaps and updates the mempool, simulating creation and committing
	// of blocks by the consensus module.
	go func() {
		defer wg.Done()

		b.ResetTimer()
		for h := 1; h <= maxHeight; h++ {
			if mp.Size() == 0 {
				break
			}
			b.StartTimer()
			txs := mp.ReapMaxBytesMaxGas(1000, -1)
			mp.PreUpdate()
			mp.Lock()
			err := mp.FlushAppConn() // needed to process the pending CheckTx requests and their callbacks
			require.NoError(b, err)
			err = mp.Update(int64(h), txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
			require.NoError(b, err)
			mp.Unlock()
			b.StopTimer()
		}
	}()

	// Concurrently, add more transactions.
	for i := numTxs + 1; i <= numTxs; i++ {
		_, err := mp.CheckTx(kvstore.NewTxFromID(i), "")
		require.NoError(b, err)
	}

	wg.Wait()

	// All added transactions should have been removed from the mempool.
	require.Zero(b, mp.Size())
}
