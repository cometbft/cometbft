//go:build gofuzz || go1.20

package tests

import (
	"context"
	"testing"

	abciclient "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/config"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/proxy"
)

func FuzzMempool(f *testing.F) {
	app := kvstore.NewInMemoryApplication()
	mtx := new(cmtsync.Mutex)
	conn := abciclient.NewLocalClient(mtx, app)
	err := conn.Start()
	if err != nil {
		panic(err)
	}

	cfg := config.DefaultMempoolConfig()
	cfg.Broadcast = false

	resp, err := app.Info(context.Background(), proxy.InfoRequest)
	if err != nil {
		panic(err)
	}
	lanesInfo, err := mempl.BuildLanesInfo(resp.LanePriorities, resp.DefaultLane)
	if err != nil {
		panic(err)
	}
	mp := mempl.NewCListMempool(cfg, conn, lanesInfo, 0)

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = mp.CheckTx(data, "")
	})
}
