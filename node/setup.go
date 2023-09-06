package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	_ "net/http/pprof" //nolint: gosec // securely exposed on separate, optional port

	dbm "github.com/cometbft/cometbft-db"

	abci "github.com/cometbft/cometbft/abci/types"
	bc "github.com/cometbft/cometbft/blocksync"
	cfg "github.com/cometbft/cometbft/config"
	cs "github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/evidence"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/light"
	mempl "github.com/cometbft/cometbft/mempool"
	mempoolv0 "github.com/cometbft/cometbft/mempool/v0"
	mempoolv1 "github.com/cometbft/cometbft/mempool/v1" //nolint:staticcheck // SA1019 Priority mempool deprecated but still supported in this release.
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/pex"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/indexer"
	blockidxkv "github.com/cometbft/cometbft/state/indexer/block/kv"
	blockidxnull "github.com/cometbft/cometbft/state/indexer/block/null"
	"github.com/cometbft/cometbft/state/indexer/sink/psql"
	"github.com/cometbft/cometbft/state/txindex"
	"github.com/cometbft/cometbft/state/txindex/kv"
	"github.com/cometbft/cometbft/state/txindex/null"
	"github.com/cometbft/cometbft/statesync"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/version"

	_ "github.com/lib/pq" // provide the psql db driver
)

const readHeaderTimeout = 10 * time.Second

// ChecksummedGenesisDoc combines a GenesisDoc together with its
// SHA256 checksum.
type ChecksummedGenesisDoc struct {
	GenesisDoc     *types.GenesisDoc
	Sha256Checksum []byte
}

// GenesisDocProvider returns a GenesisDoc together with its SHA256 checksum.
// It allows the GenesisDoc to be pulled from sources other than the
// filesystem, for instance from a distributed key-value store cluster.
type GenesisDocProvider func() (ChecksummedGenesisDoc, error)

// DefaultGenesisDocProviderFunc returns a GenesisDocProvider that loads
// the GenesisDoc from the config.GenesisFile() on the filesystem.
func DefaultGenesisDocProviderFunc(config *cfg.Config) GenesisDocProvider {
	return func() (ChecksummedGenesisDoc, error) {
		// FIXME: find a way to stream the file incrementally,
		// for the JSON	parser and the checksum computation.
		// https://github.com/cometbft/cometbft/issues/1302
		jsonBlob, err := os.ReadFile(config.GenesisFile())
		if err != nil {
			return ChecksummedGenesisDoc{}, fmt.Errorf("couldn't read GenesisDoc file: %w", err)
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
type Provider func(*cfg.Config, log.Logger) (*Node, error)

// DefaultNewNode returns a CometBFT node with default settings for the
// PrivValidator, ClientCreator, GenesisDoc, and DBProvider.
// It implements NodeProvider.
func DefaultNewNode(config *cfg.Config, logger log.Logger) (*Node, error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, fmt.Errorf("failed to load or gen node key %s: %w", config.NodeKeyFile(), err)
	}

	return NewNode(config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

// MetricsProvider returns a consensus, p2p and mempool Metrics.
type MetricsProvider func(chainID string) (*cs.Metrics, *p2p.Metrics, *mempl.Metrics, *sm.Metrics, *proxy.Metrics)

// DefaultMetricsProvider returns Metrics build using Prometheus client library
// if Prometheus is enabled. Otherwise, it returns no-op Metrics.
func DefaultMetricsProvider(config *cfg.InstrumentationConfig) MetricsProvider {
	return func(chainID string) (*cs.Metrics, *p2p.Metrics, *mempl.Metrics, *sm.Metrics, *proxy.Metrics) {
		if config.Prometheus {
			return cs.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				p2p.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				mempl.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				sm.PrometheusMetrics(config.Namespace, "chain_id", chainID),
				proxy.PrometheusMetrics(config.Namespace, "chain_id", chainID)
		}
		return cs.NopMetrics(), p2p.NopMetrics(), mempl.NopMetrics(), sm.NopMetrics(), proxy.NopMetrics()
	}
}

// Temporary interface for switching to block sync, we should get rid of v0 and v1 reactors.
// See: https://github.com/tendermint/tendermint/issues/4595
type blockSyncReactor interface {
	SwitchToBlockSync(sm.State) error
}

//------------------------------------------------------------------------------

func initDBs(config *cfg.Config, dbProvider DBProvider) (blockStore *store.BlockStore, stateDB dbm.DB, err error) {
	var blockStoreDB dbm.DB
	blockStoreDB, err = dbProvider(&DBContext{"blockstore", config})
	if err != nil {
		return
	}
	blockStore = store.NewBlockStore(blockStoreDB)

	stateDB, err = dbProvider(&DBContext{"state", config})
	if err != nil {
		return
	}

	return
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
	dbProvider DBProvider,
	eventBus *types.EventBus,
	logger log.Logger,
) (*txindex.IndexerService, txindex.TxIndexer, indexer.BlockIndexer, error) {
	var (
		txIndexer    txindex.TxIndexer
		blockIndexer indexer.BlockIndexer
	)

	switch config.TxIndex.Indexer {
	case "kv":
		store, err := dbProvider(&DBContext{"tx_index", config})
		if err != nil {
			return nil, nil, nil, err
		}

		txIndexer = kv.NewTxIndex(store)
		blockIndexer = blockidxkv.New(dbm.NewPrefixDB(store, []byte("block_events")))

	case "psql":
		if config.TxIndex.PsqlConn == "" {
			return nil, nil, nil, errors.New(`no psql-conn is set for the "psql" indexer`)
		}
		es, err := psql.NewEventSink(config.TxIndex.PsqlConn, chainID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("creating psql indexer: %w", err)
		}
		txIndexer = es.TxIndexer()
		blockIndexer = es.BlockIndexer()

	default:
		txIndexer = &null.TxIndex{}
		blockIndexer = &blockidxnull.BlockerIndexer{}
	}

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
	proxyApp proxy.AppConns,
	consensusLogger log.Logger,
) error {
	handshaker := cs.NewHandshaker(stateStore, state, blockStore, genDoc)
	handshaker.SetLogger(consensusLogger)
	handshaker.SetEventBus(eventBus)
	if err := handshaker.HandshakeWithContext(ctx, proxyApp); err != nil {
		return fmt.Errorf("error during handshake: %v", err)
	}
	return nil
}

func logNodeStartupInfo(state sm.State, pubKey crypto.PubKey, logger, consensusLogger log.Logger) {
	// Log the version info.
	logger.Info("Version info",
		"tendermint_version", version.TMCoreSemVer,
		"abci", version.ABCISemVer,
		"block", version.BlockProtocol,
		"p2p", version.P2PProtocol,
		"commit_hash", version.TMGitCommitHash,
	)

	// If the state and software differ in block version, at least log it.
	if state.Version.Consensus.Block != version.BlockProtocol {
		logger.Info("Software and state have different block protocols",
			"software", version.BlockProtocol,
			"state", state.Version.Consensus.Block,
		)
	}

	addr := pubKey.Address()
	// Log whether this node is a validator or an observer
	if state.Validators.HasAddress(addr) {
		consensusLogger.Info("This node is a validator", "addr", addr, "pubKey", pubKey)
	} else {
		consensusLogger.Info("This node is not a validator", "addr", addr, "pubKey", pubKey)
	}
}

func onlyValidatorIsUs(state sm.State, pubKey crypto.PubKey) bool {
	if state.Validators.Size() > 1 {
		return false
	}
	addr, _ := state.Validators.GetByIndex(0)
	return bytes.Equal(pubKey.Address(), addr)
}

func createMempoolAndMempoolReactor(
	config *cfg.Config,
	proxyApp proxy.AppConns,
	state sm.State,
	memplMetrics *mempl.Metrics,
	logger log.Logger,
) (mempl.Mempool, p2p.Reactor) {
	switch config.Mempool.Type {
	// allow empty string for backward compatibility
	case cfg.MempoolTypeFlood, "":
		switch config.Mempool.Version {
		case cfg.MempoolV1:
			mp := mempoolv1.NewTxMempool(
				logger,
				config.Mempool,
				proxyApp.Mempool(),
				state.LastBlockHeight,
				mempoolv1.WithMetrics(memplMetrics),
				mempoolv1.WithPreCheck(sm.TxPreCheck(state)),
				mempoolv1.WithPostCheck(sm.TxPostCheck(state)),
			)

			reactor := mempoolv1.NewReactor(
				config.Mempool,
				mp,
			)
			if config.Consensus.WaitForTxs() {
				mp.EnableTxsAvailable()
			}

			return mp, reactor

		case cfg.MempoolV0:
			mp := mempoolv0.NewCListMempool(
				config.Mempool,
				proxyApp.Mempool(),
				state.LastBlockHeight,
				mempoolv0.WithMetrics(memplMetrics),
				mempoolv0.WithPreCheck(sm.TxPreCheck(state)),
				mempoolv0.WithPostCheck(sm.TxPostCheck(state)),
			)

			mp.SetLogger(logger)

			reactor := mempoolv0.NewReactor(
				config.Mempool,
				mp,
			)
			if config.Consensus.WaitForTxs() {
				mp.EnableTxsAvailable()
			}

			return mp, reactor

		default:
			return nil, nil
		}
	case cfg.MempoolTypeNop:
		// Strictly speaking, there's no need to have a `mempl.NopMempoolReactor`, but
		// adding it leads to a cleaner code.
		return &mempl.NopMempool{}, mempl.NewNopMempoolReactor()
	default:
		panic(fmt.Sprintf("unknown mempool type: %q", config.Mempool.Type))
	}
}

func createEvidenceReactor(config *cfg.Config, dbProvider DBProvider,
	stateDB dbm.DB, blockStore *store.BlockStore, logger log.Logger,
) (*evidence.Reactor, *evidence.Pool, error) {
	evidenceDB, err := dbProvider(&DBContext{"evidence", config})
	if err != nil {
		return nil, nil, err
	}
	evidenceLogger := logger.With("module", "evidence")
	evidencePool, err := evidence.NewPool(evidenceDB, sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: config.Storage.DiscardABCIResponses,
	}), blockStore)
	if err != nil {
		return nil, nil, err
	}
	evidenceReactor := evidence.NewReactor(evidencePool)
	evidenceReactor.SetLogger(evidenceLogger)
	return evidenceReactor, evidencePool, nil
}

func createBlockchainReactor(config *cfg.Config,
	state sm.State,
	blockExec *sm.BlockExecutor,
	blockStore *store.BlockStore,
	blockSync bool,
	logger log.Logger,
	offlineStateSyncHeight int64,
) (bcReactor p2p.Reactor, err error) {
	switch config.BlockSync.Version {
	case "v0":
		bcReactor = bc.NewReactorWithOfflineStateSync(state.Copy(), blockExec, blockStore, blockSync, offlineStateSyncHeight)
	case "v1", "v2":
		return nil, fmt.Errorf("block sync version %s has been deprecated. Please use v0", config.BlockSync.Version)
	default:
		return nil, fmt.Errorf("unknown block sync version %s", config.BlockSync.Version)
	}

	bcReactor.SetLogger(logger.With("module", "blockchain"))
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
) (*cs.Reactor, *cs.State) {
	consensusState := cs.NewState(
		config.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		mempool,
		evidencePool,
		cs.StateMetrics(csMetrics),
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
	nodeInfo p2p.NodeInfo,
	nodeKey *p2p.NodeKey,
	proxyApp proxy.AppConns,
) (
	*p2p.MultiplexTransport,
	[]p2p.PeerFilterFunc,
) {
	var (
		mConnConfig = p2p.MConnConfig(config.P2P)
		transport   = p2p.NewMultiplexTransport(nodeInfo, *nodeKey, mConnConfig)
		connFilters = []p2p.ConnFilterFunc{}
		peerFilters = []p2p.PeerFilterFunc{}
	)

	if !config.P2P.AllowDuplicateIP {
		connFilters = append(connFilters, p2p.ConnDuplicateIPFilter())
	}

	// Filter peers by addr or pubkey with an ABCI query.
	// If the query return code is OK, add peer.
	if config.FilterPeers {
		connFilters = append(
			connFilters,
			// ABCI query for address filtering.
			func(_ p2p.ConnSet, c net.Conn, _ []net.IP) error {
				res, err := proxyApp.Query().QuerySync(abci.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/addr/%s", c.RemoteAddr().String()),
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
				res, err := proxyApp.Query().QuerySync(abci.RequestQuery{
					Path: fmt.Sprintf("/p2p/filter/id/%s", p.ID()),
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

	p2p.MultiplexTransportConnFilters(connFilters...)(transport)

	// Limit the number of incoming connections.
	max := config.P2P.MaxNumInboundPeers + len(splitAndTrimEmpty(config.P2P.UnconditionalPeerIDs, ",", " "))
	p2p.MultiplexTransportMaxIncomingConnections(max)(transport)

	return transport, peerFilters
}

func createSwitch(config *cfg.Config,
	transport p2p.Transport,
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
	sw.AddReactor("BLOCKCHAIN", bcReactor)
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
		addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.ID(), config.P2P.ExternalAddress))
		if err != nil {
			return nil, fmt.Errorf("p2p.external_address is incorrect: %w", err)
		}
		addrBook.AddOurAddress(addr)
	}
	if config.P2P.ListenAddress != "" {
		addr, err := p2p.NewNetAddressString(p2p.IDAddressString(nodeKey.ID(), config.P2P.ListenAddress))
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
func startStateSync(ssR *statesync.Reactor, bcR blockSyncReactor, conR *cs.Reactor,
	stateProvider statesync.StateProvider, config *cfg.StateSyncConfig, blockSync bool,
	stateStore sm.Store, blockStore *store.BlockStore, state sm.State,
) error {
	ssR.Logger.Info("Starting state sync")

	if stateProvider == nil {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		stateProvider, err = statesync.NewLightClientStateProvider(
			ctx,
			state.ChainID, state.Version, state.InitialHeight,
			config.RPCServers, light.TrustOptions{
				Period: config.TrustPeriod,
				Height: config.TrustHeight,
				Hash:   config.TrustHashBytes(),
			}, ssR.Logger.With("module", "light"))
		if err != nil {
			return fmt.Errorf("failed to set up light client state provider: %w", err)
		}
	}

	go func() {
		state, commit, err := ssR.Sync(stateProvider, config.DiscoveryTime)
		if err != nil {
			ssR.Logger.Error("State sync failed", "err", err)
			return
		}
		err = stateStore.Bootstrap(state)
		if err != nil {
			ssR.Logger.Error("Failed to bootstrap node with new state", "err", err)
			return
		}
		err = blockStore.SaveSeenCommit(state.LastBlockHeight, commit)
		if err != nil {
			ssR.Logger.Error("Failed to store last seen commit", "err", err)
			return
		}

		if blockSync {
			// FIXME Very ugly to have these metrics bleed through here.
			conR.Metrics.StateSyncing.Set(0)
			conR.Metrics.BlockSyncing.Set(1)
			err = bcR.SwitchToBlockSync(state)
			if err != nil {
				ssR.Logger.Error("Failed to switch to block sync", "err", err)
				return
			}
		} else {
			conR.SwitchToConsensus(state, true)
		}
	}()
	return nil
}

//------------------------------------------------------------------------------

// var genesisDocKey = []byte("genesisDoc")
var genesisDocHashKey = []byte("genesisDocHash")

// LoadStateFromDBOrGenesisDocProvider attempts to load the state from the
// database, or creates one using the given genesisDocProvider. On success this also
// returns the genesis doc loaded through the given provider.
func LoadStateFromDBOrGenesisDocProvider(
	stateDB dbm.DB,
	genesisDocProvider GenesisDocProvider,
) (sm.State, *types.GenesisDoc, error) {
	// Get genesis doc hash
	genDocHash, err := stateDB.Get(genesisDocHashKey)
	if err != nil {
		return sm.State{}, nil, fmt.Errorf("error retrieving genesis doc hash: %w", err)
	}
	csGenDoc, err := genesisDocProvider()
	if err != nil {
		return sm.State{}, nil, err
	}

	if err := csGenDoc.GenesisDoc.ValidateAndComplete(); err != nil {
		return sm.State{}, nil, fmt.Errorf("error in genesis doc: %w", err)
	}

	if len(genDocHash) == 0 {
		// Save the genDoc hash in the store if it doesn't already exist for future verification
		if err := stateDB.SetSync(genesisDocHashKey, csGenDoc.Sha256Checksum); err != nil {
			return sm.State{}, nil, fmt.Errorf("failed to save genesis doc hash to db: %w", err)
		}
	} else {
		if !bytes.Equal(genDocHash, csGenDoc.Sha256Checksum) {
			return sm.State{}, nil, fmt.Errorf("genesis doc hash in db does not match loaded genesis doc")
		}
	}
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisDoc(csGenDoc.GenesisDoc)
	if err != nil {
		return sm.State{}, nil, err
	}
	return state, csGenDoc.GenesisDoc, nil
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
