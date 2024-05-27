package mempool

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abciserver "github.com/cometbft/cometbft/abci/server"
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
		if err := mp.CheckTx(tx, nil, TxInfo{}); err != nil {
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

		if err := mp.CheckTx(tx, nil, TxInfo{}); err != nil {
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
			if err := mp.CheckTx(tx, nil, TxInfo{}); err != nil {
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
	if err := mp.CheckTx(tx, nil, TxInfo{}); err != nil {
		b.Fatal(err)
	}
	err := mp.FlushAppConn()
	require.NotErrorIs(b, nil, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := mp.CheckTx(tx, nil, TxInfo{}); err == nil {
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

		err := mp.CheckTx(tx, nil, TxInfo{})
		require.NoError(b, err)

		err = mp.FlushAppConn()
		require.NoError(b, err)

		require.Equal(b, 1, mp.Size())

		txs := mp.ReapMaxTxs(mp.Size())
		doCommit(b, mp, app, txs, int64(i))
		assert.True(b, true)
	}
}
