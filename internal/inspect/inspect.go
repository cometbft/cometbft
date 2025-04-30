package inspect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/internal/inspect/rpc"
	cmtstrings "github.com/cometbft/cometbft/internal/strings"
	"github.com/cometbft/cometbft/libs/log"
	rpccore "github.com/cometbft/cometbft/rpc/core"
	"github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/indexer"
	"github.com/cometbft/cometbft/state/indexer/block"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

var logger = log.NewLogger(os.Stdout)

// Inspector manages an RPC service that exports methods to debug a failed node.
// After a node shuts down due to a consensus failure, it will no longer start
// up its state cannot easily be inspected. An Inspector value provides a similar interface
// to the node, using the underlying CometBFT data stores, without bringing up
// any other components. A caller can query the Inspector service to inspect the
// persisted state and debug the failure.
type Inspector struct {
	routes rpccore.RoutesMap

	config *config.RPCConfig

	logger log.Logger

	// References to the state store and block store are maintained to enable
	// the Inspector to safely close them on shutdown.
	ss    state.Store
	bs    state.BlockStore
	txIdx txindex.TxIndexer
}

// New returns an Inspector that serves RPC on the specified BlockStore and StateStore.
// The Inspector type does not modify the state or block stores.
// The sinks are used to enable block and transaction querying via the RPC server.
// The caller is responsible for starting and stopping the Inspector service.
//

func New(
	cfg *config.RPCConfig,
	bs state.BlockStore,
	ss state.Store,
	txidx txindex.TxIndexer,
	blkidx indexer.BlockIndexer,
) *Inspector {
	routes := rpc.Routes(*cfg, ss, bs, txidx, blkidx, logger)
	eb := types.NewEventBus()
	eb.SetLogger(logger.With("module", "events"))
	return &Inspector{
		routes: routes,
		config: cfg,
		logger: logger,
		ss:     ss,
		bs:     bs,
		txIdx:  txidx,
	}
}

// NewFromConfig constructs an Inspector using the values defined in the passed in config.
func NewFromConfig(cfg *config.Config) (*Inspector, error) {
	bsDB, err := config.DefaultDBProvider(&config.DBContext{ID: "blockstore", Config: cfg})
	if err != nil {
		return nil, err
	}
	bs := store.NewBlockStore(bsDB, store.WithDBKeyLayout(cfg.Storage.ExperimentalKeyLayout))
	sDB, err := config.DefaultDBProvider(&config.DBContext{ID: "state", Config: cfg})
	if err != nil {
		return nil, err
	}
	genDoc, err := types.GenesisDocFromFile(cfg.GenesisFile())
	if err != nil {
		return nil, err
	}
	txidx, blkidx, _, err := block.IndexerFromConfig(cfg, config.DefaultDBProvider, genDoc.ChainID)
	if err != nil {
		return nil, err
	}
	ss := state.NewStore(sDB, state.StoreOptions{})
	return New(cfg.RPC, bs, ss, txidx, blkidx), nil
}

// Run starts the Inspector servers and blocks until the servers shut down. The passed
// in context is used to control the lifecycle of the servers.
func (ins *Inspector) Run(ctx context.Context) error {
	defer ins.Close()

	return startRPCServers(ctx, ins.config, ins.logger, ins.routes)
}

// Close closes all of the databases that the Inspector uses.
func (ins *Inspector) Close() error {
	errs := make([]string, 0, 3)

	if err := ins.txIdx.Close(); err != nil {
		errs = append(errs, "txIdx: "+err.Error())
	}
	if err := ins.ss.Close(); err != nil {
		errs = append(errs, "ss: "+err.Error())
	}
	if err := ins.bs.Close(); err != nil {
		errs = append(errs, "bs: "+err.Error())
	}

	if len(errs) == 0 {
		return nil
	}

	return fmt.Errorf("closing inspector's databases: %s", strings.Join(errs, "; "))
}

func startRPCServers(
	ctx context.Context,
	cfg *config.RPCConfig,
	logger log.Logger,
	routes rpccore.RoutesMap,
) error {
	var (
		g, tctx     = errgroup.WithContext(ctx)
		listenAddrs = cmtstrings.SplitAndTrimEmpty(cfg.ListenAddress, ",", " ")
		rh          = rpc.Handler(cfg, routes, logger)
	)
	for _, listenerAddr := range listenAddrs {
		server := rpc.Server{
			Logger:  logger,
			Config:  cfg,
			Handler: rh,
			Addr:    listenerAddr,
		}
		if cfg.IsTLSEnabled() {
			var (
				keyFile      = cfg.KeyFile()
				certFile     = cfg.CertFile()
				listenerAddr = listenerAddr
			)
			g.Go(func() error {
				logger.Info(
					"RPC HTTPS ironbird server starting",
					"address", listenerAddr,
					"certfile", certFile,
					"keyfile", keyFile,
				)

				err := server.ListenAndServeTLS(tctx, certFile, keyFile)
				if !errors.Is(err, net.ErrClosed) {
					return err
				}

				logger.Info("RPC HTTPS server stopped", "address", listenerAddr)
				return nil
			})
		} else {
			listenerAddr := listenerAddr
			g.Go(func() error {
				logger.Info("RPC HTTP server starting", "address", listenerAddr)

				err := server.ListenAndServe(tctx)
				if !errors.Is(err, net.ErrClosed) {
					return err
				}

				logger.Info("RPC HTTP server stopped", "address", listenerAddr)
				return nil
			})
		}
	}
	return g.Wait()
}
