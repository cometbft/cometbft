package node

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "net/http/pprof" //nolint: gosec,gci // securely exposed on separate, optional port

	_ "github.com/lib/pq" //nolint: gci // provide the psql db driver.

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	cfg "github.com/cometbft/cometbft/v2/config"
	"github.com/cometbft/cometbft/v2/crypto"
	"github.com/cometbft/cometbft/v2/crypto/tmhash"
	"github.com/cometbft/cometbft/v2/internal/blocksync"
	cs "github.com/cometbft/cometbft/v2/internal/consensus"
	"github.com/cometbft/cometbft/v2/internal/evidence"
	"github.com/cometbft/cometbft/v2/libs/log"
	"github.com/cometbft/cometbft/v2/light"
	mempl "github.com/cometbft/cometbft/v2/mempool"
	"github.com/cometbft/cometbft/v2/p2p"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
	"github.com/cometbft/cometbft/v2/p2p/pex"
	"github.com/cometbft/cometbft/v2/p2p/transport"
	"github.com/cometbft/cometbft/v2/p2p/transport/tcp"
	tcpconn "github.com/cometbft/cometbft/v2/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/v2/privval"
	"github.com/cometbft/cometbft/v2/proxy"
	sm "github.com/cometbft/cometbft/v2/state"
	"github.com/cometbft/cometbft/v2/state/indexer"
	"github.com/cometbft/cometbft/v2/state/indexer/block"
	"github.com/cometbft/cometbft/v2/state/txindex"
	"github.com/cometbft/cometbft/v2/statesync"
	"github.com/cometbft/cometbft/v2/store"
	"github.com/cometbft/cometbft/v2/types"
	"github.com/cometbft/cometbft/v2/version"
)

const readHeaderTimeout = 10 * time.Second

// ChecksummedGenesisDoc combines a GenesisDoc together with its
// SHA256 checksum.
type ChecksummedGenesisDoc struct {
	GenesisDoc     *types.GenesisDoc
	Sha256Checksum []byte
}

// Introduced to store parameters passed via cli and needed to start the node.
// This parameters should not be stored or persisted in the config file.
// This can then be further extended to include additional flags without further
// API breaking changes.
type CliParams struct {
	// SHA-256 hash of the genesis file provided via the command line.
	// This hash is used is compared against the computed hash of the
	// actual genesis file or the hash stored in the database.
	// If there is a mismatch between the hash provided via cli and the
	// hash of the genesis file or the hash in the DB, the node will not boot.
	GenesisHash []byte
}

// GenesisDocProvider returns a GenesisDoc together with its SHA256 checksum.
// It allows the GenesisDoc to be pulled from sources other than the
// filesystem, for instance from a distributed key-value store cluster.
// It is the responsibility of the GenesisDocProvider to ensure that the SHA256
// checksum correctly matches the GenesisDoc, that is:
// sha256(GenesisDoc) == Sha256Checksum.
type GenesisDocProvider func() (ChecksummedGenesisDoc, error)

// DefaultGenesisDocProviderFunc returns a GenesisDocProvider that loads
// the GenesisDoc from the config.GenesisFile() on the filesystem.
func DefaultGenesisDocProviderFunc(config *cfg.Config) GenesisDocProvider {
	return func() (ChecksummedGenesisDoc, error) {
		// FIXME: find a way to stream the file incrementally,
		// for the JSON	parser and the checksum computation.
		// https://github.com/cometbft/cometbft/v2/issues/1302
		jsonBlob, err := os.ReadFile(config.GenesisFile())
		if err != nil {
			return ChecksummedGenesisDoc{}, ErrorReadingGenesisDoc{Err: err}
		}
		incomingChecksum := tmhash.Sum(jsonBlob)
		genDoc, err := types.GenesisDocFromJSON(jsonBlob)
		if err != nil {
			return ChecksummedGenesisDoc{}, err
		}
		return ChecksummedGenesisDoc{GenesisDoc: genDoc, Sha256Checksum: incomingChecksum}, nil
	}
}

// Provider takes a config and a logger and returns a ready to go Node.
type Provider func(*cfg.Config, log.Logger, CliParams, func() (crypto.PrivKey, error)) (*Node, error)

// DefaultNewNode returns a CometBFT node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements Provider.
func DefaultNewNode(
	config *cfg.Config,
	logger log.Logger,
	cliParams CliParams,
	keyGenF func() (crypto.PrivKey, error),
) (*Node, error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, ErrorLoadOrGenNodeKey{Err: err, NodeKeyFile: config.NodeKeyFile()}
	}

	pv, err := privval.LoadOrGenFilePV(
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
		keyGenF,
	)
	if err != nil {
		return nil, ErrorLoadOrGenFilePV{
			Err:       err,
			KeyFile:   config.PrivValidatorKeyFile(),
			StateFile: config.PrivValidatorStateFile(),
		}
	}

	return NewNodeWithCliParams(context.Background(), config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		logger,
		cliParams,
	)
}

// MetricsProvider returns a consensus, p2p and mempool Metrics.
type MetricsProvider func(chainID string) (*cs.Metrics, *p2p.Metrics, *mempl.Metrics, *sm.Metrics, *store.Metrics, *proxy.Metrics, *blocksync.Metrics, *statesync.Metrics)

// DefaultMetricsProvider returns Metrics build using Prometheus client library
// if Prometheus is enabled. Otherwise, it returns no-op Metrics.
func DefaultMetricsProvider(config *cfg.InstrumentationConfig) MetricsProvider {
	return func(chainID string) (*cs.Metrics, *p2p.Metrics, *mempl.Metrics, *sm.Metrics, *store.Metrics, *proxy.Metrics, *blocksync.Metrics, *statesync.Metrics) {
		if config.Prometheus {
			return cs.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				p2p.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				mempl.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				sm.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				store.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				proxy.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				blocksync.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				statesync.PrometheusMetrics(config.Namespace, "chain_id", chainID)
		}
		return cs.NopMetrics(), p2p.NopMetrics(), mempl.NopMetrics(), sm.NopMetrics(), store.NopMetrics(), proxy.NopMetrics(), blocksync.NopMetrics(), statesync.NopMetrics()
	}
}

type blockSyncReactor interface {
	SwitchToBlockSync(state sm.State) error
}

// ------------------------------------------------------------------------------

func initDBs(config *cfg.Config, dbProvider cfg.DBProvider) (bsDB dbm.DB, stateDB dbm.DB, err error) {
	bsDB, err = dbProvider(&cfg.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return nil, nil, err
	}

	stateDB, err = dbProvider(&cfg.DBContext{ID: "state", Config: config})
	if err != nil {
		return nil, nil, err
	}

	return bsDB, stateDB, nil
}

func createAndStartProxyAppConns(clientCreator proxy.ClientCreator, logger log.Logger, metrics *proxy.Metrics) (proxy.AppConns, error) {
	proxyApp := proxy.NewAppConns(clientCreator, metrics)
	proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return nil, fmt.Errorf("error starting proxy app connections: %v", err)
	}
	return proxyApp, nil
}

func createAndStartEventBus(logger log.Logger) (*types.EventBus, error) {
	eventBus := types.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	if err := eventBus.Start(); err != nil {
		return nil, err
	}
	return eventBus, nil
}

func createAndStartIndexerService(
	config *cfg.Config,
	chainID string,
	dbProvider cfg.DBProvider,
	eventBus *types.EventBus,
	logger log.Logger,
) (*txindex.IndexerService, txindex.TxIndexer, indexer.BlockIndexer, error) {
	var (
		txIndexer    txindex.TxIndexer
		blockIndexer indexer.BlockIndexer
	)

	txIndexer, blockIndexer, allIndexersDisabled, err := block.IndexerFromConfig(config, dbProvider, chainID)
	if err != nil {
		return nil, nil, nil, err
	}
	if allIndexersDisabled {
		return nil, txIndexer, blockIndexer, nil
	}

	txIndexer.SetLogger(logger.With("module", "txindex"))
	blockIndexer.SetLogger(logger.With("module", "txindex"))

	indexerService := txindex.NewIndexerService(txIndexer, blockIndexer, eventBus, false)
	indexerService.SetLogger(logger.With("module", "txindex"))
	if err := indexerService.Start(); err != nil {
		return nil, nil, nil, err
	}

	return indexerService, txIndexer, blockIndexer, nil
}

func doHandshake(
	ctx context.Context,
	stateStore sm.Store,
	state sm.State,
	blockStore sm.BlockStore,
	genDoc *types.GenesisDoc,
	eventBus types.BlockEventPublisher,
	appInfoResponse *abci.InfoResponse,
	proxyApp proxy.AppConns,
	consensusLogger log.Logger,
) error {
	handshaker := cs.NewHandshaker(stateStore, state, blockStore, genDoc)
	handshaker.SetLogger(consensusLogger)
	handshaker.SetEventBus(eventBus)
	if err := handshaker.Handshake(ctx, appInfoResponse, proxyApp); err != nil {
		return fmt.Errorf("error during handshake: %v", err)
	}
	return nil
}

// isAddressValidator returns whether the given pubkey is associated with a validator in the state.
func isAddressValidator(state sm.State, pubKey crypto.PubKey) bool {
	return state.Validators.HasAddress(pubKey.Address())
}

func logNodeStartupInfo(
	state sm.State,
	pubKey crypto.PubKey,
	logger, consensusLogger log.Logger,
) {
	// Log the version info.
	logger.Info("Version info",
		"tendermint_version", version.CMTSemVer,
		"abci", version.ABCISemVer,
		"block", version.BlockProtocol,
		"p2p", version.P2PProtocol,
		"commit_hash", version.CMTGitCommitHash,
	)

	// If the state and software differ in block version, at least log it.
	if state.Version.Consensus.Block != version.BlockProtocol {
		logger.Info("Software and state have different block protocols",
			"software", version.BlockProtocol,
			"state", state.Version.Consensus.Block,
		)
	}

	// Log whether this node is a validator or an observer
	if isAddressValidator(state, pubKey) {
		consensusLogger.Info("This node is a validator", "addr", pubKey.Address(), "pubKey", pubKey)
	} else {
		consensusLogger.Info("This node is not a validator", "addr", pubKey.Address(), "pubKey", pubKey)
	}
}

// createMempoolAndMempoolReactor creates a mempool and a mempool reactor based on the config.
func createMempoolAndMempoolReactor(
	config *cfg.Config,
	proxyApp proxy.AppConns,
	state sm.State,
	eventBus *types.EventBus,
	waitSync bool,
	memplMetrics *mempl.Metrics,
	logger log.Logger,
	appInfoResponse *abci.InfoResponse,
) (mempl.Mempool, mempoolReactor) {
	switch config.Mempool.Type {
	// allow empty string for backward compatibility
	case cfg.MempoolTypeFlood, "":
		lanesInfo, err := mempl.BuildLanesInfo(appInfoResponse.LanePriorities, appInfoResponse.DefaultLane)
		if err != nil {
			panic(fmt.Sprintf("could not get lanes info from app: %s", err))
		}

		logger = logger.With("module", "mempool")
		options := []mempl.CListMempoolOption{
			mempl.WithMetrics(memplMetrics),
			mempl.WithPreCheck(sm.TxPreCheck(state)),
			mempl.WithPostCheck(sm.TxPostCheck(state)),
		}
		if config.Mempool.ExperimentalPublishEventPendingTx {
			options = append(options, mempl.WithNewTxCallback(func(tx types.Tx) {
				_ = eventBus.PublishEventPendingTx(types.EventDataPendingTx{
					Tx: tx,
				})
			}))
		}
		mp := mempl.NewCListMempool(
			config.Mempool,
			proxyApp.Mempool(),
			lanesInfo,
			state.LastBlockHeight,
			options...,
		)
		mp.SetLogger(logger)
		reactor := mempl.NewReactor(
			config.Mempool,
			mp,
			waitSync,
		)
		if config.Consensus.WaitForTxs() {
			mp.EnableTxsAvailable()
		}
		reactor.SetLogger(logger)

		return mp, reactor
	case cfg.MempoolTypeNop:
		// Strictly speaking, there's no need to have a `mempl.NopMempoolReactor`, but
		// adding it leads to a cleaner code.
		return &mempl.NopMempool{}, mempl.NewNopMempoolReactor()
	default:
		panic(fmt.Sprintf("unknown mempool type: %q", config.Mempool.Type))
	}
}

func createEvidenceReactor(config *cfg.Config, dbProvider cfg.DBProvider,
	stateStore sm.Store, blockStore *store.BlockStore, logger log.Logger,
) (*evidence.Reactor, *evidence.Pool, error) {
	evidenceDB, err := dbProvider(&cfg.DBContext{ID: "evidence", Config: config})
	if err != nil {
		return nil, nil, err
	}
	evidenceLogger := logger.With("module", "evidence")
	evidencePool, err := evidence.NewPool(evidenceDB, stateStore, blockStore, evidence.WithDBKeyLayout(config.Storage.ExperimentalKeyLayout))
	if err != nil {
		return nil, nil, err
	}
	evidenceReactor := evidence.NewReactor(evidencePool)
	evidenceReactor.SetLogger(evidenceLogger)
	return evidenceReactor, evidencePool, nil
}

func createBlocksyncReactor(config *cfg.Config,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore *store.BlockStore,
	blockSync bool,
	localAddr crypto.Address,
	logger log.Logger,
	metrics *blocksync.Metrics,
	offlineStateSyncHeight int64,
) (bcReactor p2p.Reactor, err error) {
	switch config.BlockSync.Version {
	case "v0":
		bcReactor = blocksync.NewReactor(state.Copy(), blockExec, blockStore, blockSync, localAddr, metrics, offlineStateSyncHeight)
	case "v1", "v2":
		return nil, fmt.Errorf("block sync version %s has been deprecated. Please use v0", config.BlockSync.Version)
	default:
		return nil, fmt.Errorf("unknown block sync version %s", config.BlockSync.Version)
	}

	bcReactor.SetLogger(logger.With("module", "blocksync"))
	return bcReactor, nil
}

func createConsensusReactor(config *cfg.Config,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore sm.BlockStore,
	mempool mempl.Mempool,
	evidencePool *evidence.Pool,
	privValidator types.PrivValidator,
	csMetrics *cs.Metrics,
	waitSync bool,
	eventBus *types.EventBus,
	consensusLogger log.Logger,
	offlineStateSyncHeight int64,
) (*cs.Reactor, *cs.State) {
	consensusState := cs.NewState(
		config.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		mempool,
		evidencePool,
		cs.StateMetrics(csMetrics),
		cs.OfflineStateSyncHeight(offlineStateSyncHeight),
	)
	consensusState.SetLogger(consensusLogger)
	if privValidator != nil {
		consensusState.SetPrivValidator(privValidator)
	}
	consensusReactor := cs.NewReactor(consensusState, waitSync, cs.ReactorMetrics(csMetrics))
	consensusReactor.SetLogger(consensusLogger)
	// services which will be publishing and/or subscribing for messages (events)
	// consensusReactor will set it on consensusState and blockExecutor
	consensusReactor.SetEventBus(eventBus)
	return consensusReactor, consensusState
}

func createTransport(
	config *cfg.Config,
	nodeKey *p2p.NodeKey,
	proxyApp proxy.AppConns,
) (
	*tcp.MultiplexTransport,
	[]p2p.PeerFilterFunc,
) {
	tcpConfig := tcpconn.DefaultMConnConfig()
	tcpConfig.FlushThrottle = config.P2P.FlushThrottleTimeout
	tcpConfig.SendRate = config.P2P.SendRate
	tcpConfig.RecvRate = config.P2P.RecvRate
	tcpConfig.MaxPacketMsgPayloadSize = config.P2P.MaxPacketMsgPayloadSize
	tcpConfig.TestFuzz = config.P2P.TestFuzz
	tcpConfig.TestFuzzConfig = config.P2P.TestFuzzConfig
	var (
		transport   = tcp.NewMultiplexTransport(*nodeKey, tcpConfig)
		connFilters = []tcp.ConnFilterFunc{}
		peerFilters = []p2p.PeerFilterFunc{}
	)

	if !config.P2P.AllowDuplicateIP {
		connFilters = append(connFilters, tcp.ConnDuplicateIPFilter())
	}

	// Filter peers by addr or pubkey with an ABCI query.
	// If the query return code is OK, add peer.
	if config.FilterPeers {
		connFilters = append(
			connFilters,
			// ABCI query for address filtering.
			func(_ tcp.ConnSet, c net.Conn, _ []net.IP) error {
				res, err := proxyApp.Query().Query(context.TODO(), &abci.QueryRequest{
					Path: "/p2p/filter/addr/" + c.RemoteAddr().String(),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("error querying abci app: %v", res)
				}

				return nil
			},
		)

		peerFilters = append(
			peerFilters,
			// ABCI query for ID filtering.
			func(_ p2p.IPeerSet, p p2p.Peer) error {
				res, err := proxyApp.Query().Query(context.TODO(), &abci.QueryRequest{
					Path: "/p2p/filter/id/" + p.ID(),
				})
				if err != nil {
					return err
				}
				if res.IsErr() {
					return fmt.Errorf("error querying abci app: %v", res)
				}

				return nil
			},
		)
	}

	tcp.MultiplexTransportConnFilters(connFilters...)(transport)

	// Limit the number of incoming connections.
	max := config.P2P.MaxNumInboundPeers + len(splitAndTrimEmpty(config.P2P.UnconditionalPeerIDs, ",", " "))
	tcp.MultiplexTransportMaxIncomingConnections(max)(transport)

	return transport, peerFilters
}

func createSwitch(config *cfg.Config,
	transport transport.Transport,
	p2pMetrics *p2p.Metrics,
	peerFilters []p2p.PeerFilterFunc,
	mempoolReactor p2p.Reactor,
	bcReactor p2p.Reactor,
	stateSyncReactor *statesync.Reactor,
	consensusReactor *cs.Reactor,
	evidenceReactor *evidence.Reactor,
	nodeInfo p2p.NodeInfo,
	nodeKey *p2p.NodeKey,
	p2pLogger log.Logger,
) *p2p.Switch {
	sw := p2p.NewSwitch(
		config.P2P,
		transport,
		p2p.WithMetrics(p2pMetrics),
		p2p.SwitchPeerFilters(peerFilters...),
	)
	sw.SetLogger(p2pLogger)
	if config.Mempool.Type != cfg.MempoolTypeNop {
		sw.AddReactor("MEMPOOL", mempoolReactor)
	}
	sw.AddReactor("BLOCKSYNC", bcReactor)
	sw.AddReactor("CONSENSUS", consensusReactor)
	sw.AddReactor("EVIDENCE", evidenceReactor)
	sw.AddReactor("STATESYNC", stateSyncReactor)

	sw.SetNodeInfo(nodeInfo)
	sw.SetNodeKey(nodeKey)

	p2pLogger.Info("P2P Node ID", "ID", nodeKey.ID(), "file", config.NodeKeyFile())
	return sw
}

func createAddrBookAndSetOnSwitch(config *cfg.Config, sw *p2p.Switch,
	p2pLogger log.Logger, nodeKey *p2p.NodeKey,
) (pex.AddrBook, error) {
	addrBook := pex.NewAddrBook(config.P2P.AddrBookFile(), config.P2P.AddrBookStrict)
	addrBook.SetLogger(p2pLogger.With("book", config.P2P.AddrBookFile()))

	// Add ourselves to addrbook to prevent dialing ourselves
	if config.P2P.ExternalAddress != "" {
		addr, err := na.NewFromString(na.IDAddrString(nodeKey.ID(), config.P2P.ExternalAddress))
		if err != nil {
			return nil, fmt.Errorf("p2p.external_address is incorrect: %w", err)
		}
		addrBook.AddOurAddress(addr)
	}
	if config.P2P.ListenAddress != "" {
		addr, err := na.NewFromString(na.IDAddrString(nodeKey.ID(), config.P2P.ListenAddress))
		if err != nil {
			return nil, fmt.Errorf("p2p.laddr is incorrect: %w", err)
		}
		addrBook.AddOurAddress(addr)
	}

	sw.SetAddrBook(addrBook)

	return addrBook, nil
}

func createPEXReactorAndAddToSwitch(addrBook pex.AddrBook, config *cfg.Config,
	sw *p2p.Switch, logger log.Logger,
) *pex.Reactor {
	// TODO persistent peers ? so we can have their DNS addrs saved
	pexReactor := pex.NewReactor(addrBook,
		&pex.ReactorConfig{
			Seeds:    splitAndTrimEmpty(config.P2P.Seeds, ",", " "),
			SeedMode: config.P2P.SeedMode,
			// See consensus/reactor.go: blocksToContributeToBecomeGoodPeer 10000
			// blocks assuming 10s blocks ~ 28 hours.
			// TODO (melekes): make it dynamic based on the actual block latencies
			// from the live network.
			// https://github.com/tendermint/tendermint/issues/3523
			SeedDisconnectWaitPeriod:     28 * time.Hour,
			PersistentPeersMaxDialPeriod: config.P2P.PersistentPeersMaxDialPeriod,
		})
	pexReactor.SetLogger(logger.With("module", "pex"))
	sw.AddReactor("PEX", pexReactor)
	return pexReactor
}

// startStateSync starts an asynchronous state sync process, then switches to block sync mode.
func startStateSync(
	ssR *statesync.Reactor,
	stateProvider statesync.StateProvider,
	config *cfg.StateSyncConfig,
	stateStore sm.Store,
	blockStore *store.BlockStore,
	state sm.State,
	dbKeyLayoutVersion string,
) (chan sm.State, chan error, error) {
	ssR.Logger.Info("Starting state sync")

	if stateProvider == nil {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		stateProvider, err = statesync.NewLightClientStateProviderWithDBKeyVersion(
			ctx,
			state.ChainID, state.Version, state.InitialHeight,
			config.RPCServers, light.TrustOptions{
				Period: config.TrustPeriod,
				Height: config.TrustHeight,
				Hash:   config.TrustHashBytes(),
			}, ssR.Logger.With("module", "light"),
			dbKeyLayoutVersion)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to set up light client state provider: %w", err)
		}
	}

	// Run statesync in a separate goroutine in order not to block `Node.Start`
	stateCh := make(chan sm.State, 1)
	errc := make(chan error, 1)
	go func() {
		newState, commit, err := ssR.Sync(stateProvider, config.MaxDiscoveryTime)
		if err != nil {
			errc <- fmt.Errorf("statesync: %w", err)
			return
		}

		err = stateStore.Bootstrap(newState)
		if err != nil {
			errc <- fmt.Errorf("bootstrap: %w", err)
			return
		}

		err = blockStore.SaveSeenCommit(newState.LastBlockHeight, commit)
		if err != nil {
			errc <- fmt.Errorf("save seen commit: %w", err)
			return
		}

		stateCh <- newState
	}()

	return stateCh, errc, nil
}

// ------------------------------------------------------------------------------

var (
	genesisDocKey     = []byte("genesisDoc")
	genesisDocHashKey = []byte("genesisDocHash")
)

func LoadStateFromDBOrGenesisDocProviderWithConfig(
	stateDB dbm.DB,
	genesisDocProvider GenesisDocProvider,
	operatorGenesisHashHex string,
	config *cfg.Config,
) (sm.State, *types.GenesisDoc, error) {
	// Get genesis doc hash
	genDocHash, err := stateDB.Get(genesisDocHashKey)
	if err != nil {
		return sm.State{}, nil, ErrRetrieveGenesisDocHash{Err: err}
	}
	csGenDoc, err := genesisDocProvider()
	if err != nil {
		return sm.State{}, nil, err
	}

	if err = csGenDoc.GenesisDoc.ValidateAndComplete(); err != nil {
		return sm.State{}, nil, ErrGenesisDoc{Err: err}
	}

	checkSumStr := hex.EncodeToString(csGenDoc.Sha256Checksum)
	if err := tmhash.ValidateSHA256(checkSumStr); err != nil {
		const formatStr = "invalid genesis doc SHA256 checksum: %s"
		return sm.State{}, nil, fmt.Errorf(formatStr, err)
	}

	// Validate that existing or recently saved genesis file hash matches optional --genesis_hash passed by operator
	if operatorGenesisHashHex != "" {
		decodedOperatorGenesisHash, err := hex.DecodeString(operatorGenesisHashHex)
		if err != nil {
			return sm.State{}, nil, ErrGenesisHashDecode
		}
		if !bytes.Equal(csGenDoc.Sha256Checksum, decodedOperatorGenesisHash) {
			return sm.State{}, nil, ErrPassedGenesisHashMismatch
		}
	}

	if len(genDocHash) == 0 {
		// Save the genDoc hash in the store if it doesn't already exist for future verification
		if err = stateDB.SetSync(genesisDocHashKey, csGenDoc.Sha256Checksum); err != nil {
			return sm.State{}, nil, ErrSaveGenesisDocHash{Err: err}
		}
	} else {
		if !bytes.Equal(genDocHash, csGenDoc.Sha256Checksum) {
			return sm.State{}, nil, ErrLoadedGenesisDocHashMismatch
		}
	}

	dbKeyLayoutVersion := ""
	if config != nil {
		dbKeyLayoutVersion = config.Storage.ExperimentalKeyLayout
	}
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
		DBKeyLayout:          dbKeyLayoutVersion,
	})

	state, err := stateStore.LoadFromDBOrGenesisDoc(csGenDoc.GenesisDoc)
	if err != nil {
		return sm.State{}, nil, err
	}
	return state, csGenDoc.GenesisDoc, nil
}

// LoadStateFromDBOrGenesisDocProvider attempts to load the state from the
// database, or creates one using the given genesisDocProvider. On success this also
// returns the genesis doc loaded through the given provider.

// Note that if you don't have a version of the key layout set in your DB already,
// and no config is passed, it will default to v1.
func LoadStateFromDBOrGenesisDocProvider(
	stateDB dbm.DB,
	genesisDocProvider GenesisDocProvider,
	operatorGenesisHashHex string,
) (sm.State, *types.GenesisDoc, error) {
	return LoadStateFromDBOrGenesisDocProviderWithConfig(stateDB, genesisDocProvider, operatorGenesisHashHex, nil)
}

func createAndStartPrivValidatorSocketClient(
	listenAddr,
	chainID string,
	logger log.Logger,
) (types.PrivValidator, error) {
	pve, err := privval.NewSignerListener(listenAddr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to start private validator: %w", err)
	}

	pvsc, err := privval.NewSignerClient(pve, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to start private validator: %w", err)
	}

	// try to get a pubkey from private validate first time
	_, err = pvsc.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("can't get pubkey: %w", err)
	}

	const (
		retries = 50 // 50 * 100ms = 5s total
		timeout = 100 * time.Millisecond
	)
	pvscWithRetries := privval.NewRetrySignerClient(pvsc, retries, timeout)

	return pvscWithRetries, nil
}

// splitAndTrimEmpty slices s into all subslices separated by sep and returns a
// slice of the string s with all leading and trailing Unicode code points
// contained in cutset removed. If sep is empty, SplitAndTrim splits after each
// UTF-8 sequence. First part is equivalent to strings.SplitN with a count of
// -1.  also filter out empty strings, only return non-empty strings.
func splitAndTrimEmpty(s, sep, cutset string) []string {
	if s == "" {
		return []string{}
	}

	spl := strings.Split(s, sep)
	nonEmptyStrings := make([]string, 0, len(spl))
	for i := 0; i < len(spl); i++ {
		element := strings.Trim(spl[i], cutset)
		if element != "" {
			nonEmptyStrings = append(nonEmptyStrings, element)
		}
	}
	return nonEmptyStrings
}
