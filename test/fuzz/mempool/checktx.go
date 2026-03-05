package reactor

import (
	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/config"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/proxy"
)

var mempool mempl.Mempool

func init() {
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	client, _ := cc.NewABCIClient()
	err := client.Start()
	if err != nil {
		panic(err)
	}

	cfg := config.DefaultMempoolConfig()
	cfg.Broadcast = false
	appConnMem := proxy.NewAppConnMempool(client, proxy.NopMetrics())
	mempool = mempl.NewCListMempool(cfg, appConnMem, 0)
}

func Fuzz(data []byte) int {
	err := mempool.CheckTx(data, nil, mempl.TxInfo{})
	if err != nil {
		return 0
	}

	return 1
}
