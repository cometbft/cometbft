package reactor

import (
	"context"

	"github.com/cometbft/cometbft/v2/abci/example/kvstore"
	"github.com/cometbft/cometbft/v2/config"
	mempl "github.com/cometbft/cometbft/v2/mempool"
	"github.com/cometbft/cometbft/v2/proxy"
)

var mempool mempl.Mempool

func init() {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	appConnMem, _ := cc.NewABCIMempoolClient()
	err := appConnMem.Start()
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
	mempool = mempl.NewCListMempool(cfg, appConnMem, lanesInfo, 0)
}

func Fuzz(data []byte) int {
	_, err := mempool.CheckTx(data, "")
	if err != nil {
		return 0
	}

	return 1
}
