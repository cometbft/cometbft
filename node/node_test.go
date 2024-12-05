package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/internal/evidence"
	kt "github.com/cometbft/cometbft/internal/keytypes"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/internal/test"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	p2pmock "github.com/cometbft/cometbft/p2p/mock"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

func TestNodeStartStop(t *testing.T) {
	config := test.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)
	err = n.Start()
	require.NoError(t, err)

	t.Logf("Started node %v", n.sw.NodeInfo())

	// wait for the node to produce a block
	blocksSub, err := n.EventBus().Subscribe(context.Background(), "node_test", types.EventQueryNewBlock)
	require.NoError(t, err)
	select {
	case <-blocksSub.Out():
	case <-blocksSub.Canceled():
		t.Fatal("blocksSub was canceled")
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for the node to produce a block")
	}

	// stop the node
	go func() {
		err = n.Stop()
		require.NoError(t, err)
	}()

	select {
	case <-n.Quit():
	case <-time.After(5 * time.Second):
		pid := os.Getpid()
		p, err := os.FindProcess(pid)
		if err != nil {
			panic(err)
		}
		err = p.Signal(syscall.SIGABRT)
		fmt.Println(err)
		t.Fatal("timed out waiting for shutdown")
	}
}

func TestSplitAndTrimEmpty(t *testing.T) {
	testCases := []struct {
		s        string
		sep      string
		cutset   string
		expected []string
	}{
		{"a,b,c", ",", " ", []string{"a", "b", "c"}},
		{" a , b , c ", ",", " ", []string{"a", "b", "c"}},
		{" a, b, c ", ",", " ", []string{"a", "b", "c"}},
		{" a, ", ",", " ", []string{"a"}},
		{"   ", ",", " ", []string{}},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, splitAndTrimEmpty(tc.s, tc.sep, tc.cutset), "%s", tc.s)
	}
}

func TestCompanionInitialHeightSetup(t *testing.T) {
	config := test.ResetTestRoot("companion_initial_height")
	defer os.RemoveAll(config.RootDir)
	config.Storage.Pruning.DataCompanion.Enabled = true
	config.Storage.Pruning.DataCompanion.InitialBlockRetainHeight = 1
	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)

	companionRetainHeight, err := n.stateStore.GetCompanionBlockRetainHeight()
	require.NoError(t, err)
	require.Equal(t, int64(1), companionRetainHeight)
}

func TestNodeDelayedStart(t *testing.T) {
	config := test.ResetTestRoot("node_delayed_start_test")
	defer os.RemoveAll(config.RootDir)
	now := cmttime.Now()

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	n.genesisTime = now.Add(2 * time.Second)
	require.NoError(t, err)
	n.genesisTime = now.Add(2 * time.Second)

	err = n.Start()
	require.NoError(t, err)
	defer n.Stop() //nolint:errcheck // ignore for tests

	startTime := cmttime.Now()
	assert.True(t, true, startTime.After(n.genesisTime))
}

func TestNodeSetAppVersion(t *testing.T) {
	config := test.ResetTestRoot("node_app_version_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)

	// default config uses the kvstore app
	appVersion := kvstore.AppVersion

	// check version is set in state
	state, err := n.stateStore.Load()
	require.NoError(t, err)
	assert.Equal(t, state.Version.Consensus.App, appVersion)

	// check version is set in node info
	assert.Equal(t, n.nodeInfo.(p2p.NodeInfoDefault).ProtocolVersion.App, appVersion)
}

func TestPprofServer(t *testing.T) {
	config := test.ResetTestRoot("node_pprof_test")
	defer os.RemoveAll(config.RootDir)
	config.RPC.PprofListenAddress = testFreeAddr(t)

	// should not work yet
	_, err := http.Get("http://" + config.RPC.PprofListenAddress) //nolint: bodyclose
	require.Error(t, err)

	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)
	require.NoError(t, n.Start())
	defer func() {
		require.NoError(t, n.Stop())
	}()
	assert.NotNil(t, n.pprofSrv)

	resp, err := http.Get("http://" + config.RPC.PprofListenAddress + "/debug/pprof")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNodeSetPrivValTCP(t *testing.T) {
	addr := "tcp://" + testFreeAddr(t)

	config := test.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.PrivValidatorListenAddr = addr

	dialer := privval.DialTCPFn(addr, 100*time.Millisecond, ed25519.GenPrivKey())
	dialerEndpoint := privval.NewSignerDialerEndpoint(
		log.TestingLogger(),
		dialer,
	)
	privval.SignerDialerEndpointTimeoutReadWrite(100 * time.Millisecond)(dialerEndpoint)

	signerServer := privval.NewSignerServer(
		dialerEndpoint,
		test.DefaultTestChainID,
		types.NewMockPV(),
	)

	go func() {
		err := signerServer.Start()
		if err != nil {
			panic(err)
		}
	}()
	defer signerServer.Stop() //nolint:errcheck // ignore for tests

	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)
	assert.IsType(t, &privval.RetrySignerClient{}, n.PrivValidator())
}

// address without a protocol must result in error.
func TestPrivValidatorListenAddrNoProtocol(t *testing.T) {
	addrNoPrefix := testFreeAddr(t)

	config := test.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.PrivValidatorListenAddr = addrNoPrefix

	_, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.ErrorAs(t, err, &ErrPrivValidatorSocketClient{})
}

func TestNodeSetPrivValIPC(t *testing.T) {
	tmpfile := "/tmp/kms." + cmtrand.Str(6) + ".sock"
	defer os.Remove(tmpfile) // clean up

	config := test.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.PrivValidatorListenAddr = "unix://" + tmpfile

	dialer := privval.DialUnixFn(tmpfile)
	dialerEndpoint := privval.NewSignerDialerEndpoint(
		log.TestingLogger(),
		dialer,
	)
	privval.SignerDialerEndpointTimeoutReadWrite(100 * time.Millisecond)(dialerEndpoint)

	pvsc := privval.NewSignerServer(
		dialerEndpoint,
		test.DefaultTestChainID,
		types.NewMockPV(),
	)

	go func() {
		err := pvsc.Start()
		require.NoError(t, err)
	}()
	defer pvsc.Stop() //nolint:errcheck // ignore for tests

	n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, nil)
	require.NoError(t, err)
	assert.IsType(t, &privval.RetrySignerClient{}, n.PrivValidator())
}

func TestNodeSetFilePrivVal(t *testing.T) {
	for _, keyType := range kt.ListSupportedKeyTypes() {
		t.Run(keyType, func(t *testing.T) {
			config := test.ResetTestRootWithChainIDNoOverwritePrivval("node_priv_val_file_test_"+keyType, "test_chain_"+keyType)
			defer os.RemoveAll(config.RootDir)

			keyGenF := func() (crypto.PrivKey, error) {
				return kt.GenPrivKey(keyType)
			}
			n, err := DefaultNewNode(config, log.TestingLogger(), CliParams{}, keyGenF)
			require.NoError(t, err)
			assert.IsType(t, &privval.FilePV{}, n.PrivValidator())
		})
	}
}

// testFreeAddr claims a free port so we don't block on listener being ready.
func testFreeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	return fmt.Sprintf("127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
}

// create a proposal block using real and full
// mempool and evidence pool and validate it.
func TestCreateProposalBlock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := test.ResetTestRoot("node_create_proposal")
	defer os.RemoveAll(config.RootDir)
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := proxy.NewAppConns(cc, proxy.NopMetrics())
	err := proxyApp.Start()
	require.NoError(t, err)
	defer proxyApp.Stop() //nolint:errcheck // ignore for tests

	logger := log.TestingLogger()

	var height int64 = 1
	state, stateDB, privVals := state(1, height)
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	var (
		partSize uint32 = 256
		maxBytes int64  = 16384
	)
	maxEvidenceBytes := maxBytes / 2
	state.ConsensusParams.Block.MaxBytes = maxBytes
	state.ConsensusParams.Evidence.MaxBytes = maxEvidenceBytes
	proposerAddr, _ := state.Validators.GetByIndex(0)

	// Make Mempool
	resp, err := app.Info(context.Background(), proxy.InfoRequest)
	require.NoError(t, err)
	lanesInfo, err := mempl.BuildLanesInfo(resp.LanePriorities, resp.DefaultLane)
	require.NoError(t, err)
	memplMetrics := mempl.NopMetrics()
	mempool := mempl.NewCListMempool(config.Mempool,
		proxyApp.Mempool(),
		lanesInfo,
		state.LastBlockHeight,
		mempl.WithMetrics(memplMetrics),
		mempl.WithPreCheck(sm.TxPreCheck(state)),
		mempl.WithPostCheck(sm.TxPostCheck(state)))

	// Make EvidencePool
	evidenceDB := dbm.NewMemDB()
	blockStore := store.NewBlockStore(dbm.NewMemDB())
	evidencePool, err := evidence.NewPool(evidenceDB, stateStore, blockStore)
	require.NoError(t, err)
	evidencePool.SetLogger(logger)

	// fill the evidence pool with more evidence
	// than can fit in a block
	var currentBytes int64
	for currentBytes <= maxEvidenceBytes {
		ev, err := types.NewMockDuplicateVoteEvidenceWithValidator(height, cmttime.Now(), privVals[0], "test-chain")
		require.NoError(t, err)
		currentBytes += int64(len(ev.Bytes()))
		evidencePool.ReportConflictingVotes(ev.VoteA, ev.VoteB)
	}

	evList, size := evidencePool.PendingEvidence(state.ConsensusParams.Evidence.MaxBytes)
	require.Less(t, size, state.ConsensusParams.Evidence.MaxBytes+1)
	evData := &types.EvidenceData{Evidence: evList}
	require.EqualValues(t, size, evData.ByteSize())

	// fill the mempool with more txs
	// than can fit in a block
	txLength := 100
	for i := 0; i <= int(maxBytes)/txLength; i++ {
		tx := cmtrand.Bytes(txLength)
		_, err := mempool.CheckTx(tx, "")
		require.NoError(t, err)
	}

	blockExec := sm.NewBlockExecutor(
		stateStore,
		logger,
		proxyApp.Consensus(),
		mempool,
		evidencePool,
		blockStore,
	)

	extCommit := &types.ExtendedCommit{Height: height - 1}
	block, err := blockExec.CreateProposalBlock(
		ctx,
		height,
		state,
		extCommit,
		proposerAddr,
	)
	require.NoError(t, err)

	// check that the part set does not exceed the maximum block size
	partSet, err := block.MakePartSet(partSize)
	require.NoError(t, err)
	assert.Less(t, partSet.ByteSize(), maxBytes)

	partSetFromHeader := types.NewPartSetFromHeader(partSet.Header())
	for partSetFromHeader.Count() < partSetFromHeader.Total() {
		added, err := partSetFromHeader.AddPart(partSet.GetPart(int(partSetFromHeader.Count())))
		require.NoError(t, err)
		require.True(t, added)
	}
	assert.EqualValues(t, partSetFromHeader.ByteSize(), partSet.ByteSize())

	err = blockExec.ValidateBlock(state, block)
	require.NoError(t, err)
}

func TestMaxProposalBlockSize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := test.ResetTestRoot("node_create_proposal")
	defer os.RemoveAll(config.RootDir)
	app := kvstore.NewInMemoryApplication()
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := proxy.NewAppConns(cc, proxy.NopMetrics())
	err := proxyApp.Start()
	require.NoError(t, err)
	defer proxyApp.Stop() //nolint:errcheck // ignore for tests

	logger := log.TestingLogger()

	var height int64 = 1
	state, stateDB, _ := state(1, height)
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	var maxBytes int64 = 16384
	var partSize uint32 = 256
	state.ConsensusParams.Block.MaxBytes = maxBytes
	proposerAddr, _ := state.Validators.GetByIndex(0)

	// Make Mempool
	resp, err := app.Info(context.Background(), proxy.InfoRequest)
	require.NoError(t, err)
	lanesInfo, err := mempl.BuildLanesInfo(resp.LanePriorities, resp.DefaultLane)
	require.NoError(t, err)
	memplMetrics := mempl.NopMetrics()
	mempool := mempl.NewCListMempool(config.Mempool,
		proxyApp.Mempool(),
		lanesInfo,
		state.LastBlockHeight,
		mempl.WithMetrics(memplMetrics),
		mempl.WithPreCheck(sm.TxPreCheck(state)),
		mempl.WithPostCheck(sm.TxPostCheck(state)))

	blockStore := store.NewBlockStore(dbm.NewMemDB())

	// fill the mempool with one txs just below the maximum size
	txLength := int(types.MaxDataBytesNoEvidence(maxBytes, 1))
	tx := cmtrand.Bytes(txLength - 4) // to account for the varint
	_, err = mempool.CheckTx(tx, "")
	require.NoError(t, err)

	blockExec := sm.NewBlockExecutor(
		stateStore,
		logger,
		proxyApp.Consensus(),
		mempool,
		sm.EmptyEvidencePool{},
		blockStore,
	)

	extCommit := &types.ExtendedCommit{Height: height - 1}
	block, err := blockExec.CreateProposalBlock(
		ctx,
		height,
		state,
		extCommit,
		proposerAddr,
	)
	require.NoError(t, err)

	pb, err := block.ToProto()
	require.NoError(t, err)
	assert.Less(t, int64(pb.Size()), maxBytes)

	// check that the part set does not exceed the maximum block size
	partSet, err := block.MakePartSet(partSize)
	require.NoError(t, err)
	assert.EqualValues(t, partSet.ByteSize(), int64(pb.Size()))
}

func TestNodeNewNodeCustomReactors(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_custom_reactors_test")
	defer os.RemoveAll(config.RootDir)

	cr := p2pmock.NewReactor()
	cr.Channels = []transport.StreamDescriptor{
		conn.StreamDescriptor{
			ID:                  byte(0x31),
			Priority:            5,
			SendQueueCapacity:   100,
			RecvMessageCapacity: 100,
		},
	}
	customBlocksyncReactor := p2pmock.NewReactor()

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	pv, err := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)
	n, err := NewNode(context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
		CustomReactors(map[string]p2p.Reactor{"FOO": cr, "BLOCKSYNC": customBlocksyncReactor}),
	)
	require.NoError(t, err)

	err = n.Start()
	require.NoError(t, err)
	defer n.Stop() //nolint:errcheck // ignore for tests

	assert.True(t, cr.IsRunning())
	assert.Equal(t, cr, n.Switch().Reactor("FOO"))

	assert.True(t, customBlocksyncReactor.IsRunning())
	assert.Equal(t, customBlocksyncReactor, n.Switch().Reactor("BLOCKSYNC"))

	channels := n.NodeInfo().(p2p.NodeInfoDefault).Channels
	assert.Contains(t, channels, mempl.MempoolChannel)
	assert.Contains(t, channels, cr.Channels[0].StreamID())
}

// Simple test to confirm that an existing genesis file will be deleted from the DB
// TODO Confirm that the deletion of a very big file does not crash the machine.
func TestNodeNewNodeDeleteGenesisFileFromDB(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_delete_genesis_from_db")
	defer os.RemoveAll(config.RootDir)
	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.PebbleDBBackend)
	// Ensure the genesis doc hash is saved to db
	stateDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "state", Config: config})
	require.NoError(t, err)

	err = stateDB.SetSync(genesisDocKey, []byte("genFile"))
	require.NoError(t, err)

	genDocFromDB, err := stateDB.Get(genesisDocKey)
	require.NoError(t, err)
	require.Equal(t, genDocFromDB, []byte("genFile"))

	stateDB.Close()

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	pv, err := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.NoError(t, err)

	// Start and stop to close the db for later reading
	err = n.Start()
	require.NoError(t, err)

	err = n.Stop()
	require.NoError(t, err)

	stateDB, err = cfg.DefaultDBProvider(&cfg.DBContext{ID: "state", Config: config})
	require.NoError(t, err)
	genDocHash, err := stateDB.Get(genesisDocHashKey)
	require.NoError(t, err)
	require.NotNil(t, genDocHash, "genesis doc hash should be saved in db")
	require.Len(t, genDocHash, tmhash.Size)

	err = stateDB.Close()
	require.NoError(t, err)
}

func TestNodeNewNodeGenesisHashMismatch(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_genesis_hash")
	defer os.RemoveAll(config.RootDir)

	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.PebbleDBBackend)

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	pv, err := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)
	n, err := NewNode(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.NoError(t, err)

	// Start and stop to close the db for later reading
	err = n.Start()
	require.NoError(t, err)

	err = n.Stop()
	require.NoError(t, err)

	// Ensure the genesis doc hash is saved to db
	stateDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "state", Config: config})
	require.NoError(t, err)

	genDocHash, err := stateDB.Get(genesisDocHashKey)
	require.NoError(t, err)
	require.NotNil(t, genDocHash, "genesis doc hash should be saved in db")
	require.Len(t, genDocHash, tmhash.Size)

	err = stateDB.Close()
	require.NoError(t, err)

	// Modify the genesis file chain ID to get a different hash
	genBytes := cmtos.MustReadFile(config.GenesisFile())
	var genesisDoc types.GenesisDoc
	err = cmtjson.Unmarshal(genBytes, &genesisDoc)
	require.NoError(t, err)

	genesisDoc.ChainID = "different-chain-id"
	err = genesisDoc.SaveAs(config.GenesisFile())
	require.NoError(t, err)

	pv, err = privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)
	_, err = NewNode(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.ErrorIs(t, err, ErrLoadedGenesisDocHashMismatch, "NewNode should error when genesisDoc is changed")
}

func TestNodeGenesisHashFlagMatch(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_genesis_hash_flag_match")
	defer os.RemoveAll(config.RootDir)

	config.DBBackend = string(dbm.PebbleDBBackend)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)
	// Get correct hash of correct genesis file
	jsonBlob, err := os.ReadFile(config.GenesisFile())
	require.NoError(t, err)

	// Set the cli params variable to the correct hash
	incomingChecksum := tmhash.Sum(jsonBlob)
	cliParams := CliParams{GenesisHash: incomingChecksum}
	pv, err := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)

	_, err = NewNodeWithCliParams(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
		cliParams,
	)
	require.NoError(t, err)
}

func TestNodeGenesisHashFlagMismatch(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_genesis_hash_flag_mismatch")
	defer os.RemoveAll(config.RootDir)

	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.PebbleDBBackend)

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	// Generate hash of wrong file
	f, err := os.ReadFile(config.PrivValidatorKeyFile())
	require.NoError(t, err)
	flagHash := tmhash.Sum(f)

	// Set genesis flag value to incorrect hash
	cliParams := CliParams{GenesisHash: flagHash}

	pv, err := privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile(), nil)
	require.NoError(t, err)
	_, err = NewNodeWithCliParams(
		context.Background(),
		config,
		pv,
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
		cliParams,
	)
	require.ErrorIs(t, err, ErrPassedGenesisHashMismatch, "NewNode should error when genesis flag value is incorrectly set")

	f, err = os.ReadFile(config.GenesisFile())
	require.NoError(t, err)

	genHash := tmhash.Sum(f)

	genHashMismatch := bytes.Equal(genHash, flagHash)
	require.False(t, genHashMismatch)
}

func TestLoadStateFromDBOrGenesisDocProviderWithConfig(t *testing.T) {
	config := test.ResetTestRoot(t.Name())
	config.DBBackend = string(dbm.PebbleDBBackend)

	_, stateDB, err := initDBs(config, cfg.DefaultDBProvider)
	require.NoErrorf(t, err, "state DB setup: %s", err)

	genDocProviderFunc := func(sha256Checksum []byte) GenesisDocProvider {
		return func() (ChecksummedGenesisDoc, error) {
			genDocJSON, err := os.ReadFile(config.GenesisFile())
			if err != nil {
				formatStr := "reading genesis file: %s"
				return ChecksummedGenesisDoc{}, fmt.Errorf(formatStr, err)
			}

			genDoc, err := types.GenesisDocFromJSON(genDocJSON)
			if err != nil {
				formatStr := "parsing genesis file: %s"
				return ChecksummedGenesisDoc{}, fmt.Errorf(formatStr, err)
			}

			checksummedGenesisDoc := ChecksummedGenesisDoc{
				GenesisDoc:     genDoc,
				Sha256Checksum: sha256Checksum,
			}

			return checksummedGenesisDoc, nil
		}
	}

	t.Run("NilGenesisChecksum", func(t *testing.T) {
		genDocProvider := genDocProviderFunc(nil)

		_, _, err = LoadStateFromDBOrGenesisDocProviderWithConfig(
			stateDB,
			genDocProvider,
			"",
			nil,
		)

		wantErr := "invalid genesis doc SHA256 checksum: expected 64 characters, but have 0"
		assert.EqualError(t, err, wantErr)
	})

	t.Run("ShorterGenesisChecksum", func(t *testing.T) {
		genDocProvider := genDocProviderFunc([]byte("shorter"))

		_, _, err = LoadStateFromDBOrGenesisDocProviderWithConfig(
			stateDB,
			genDocProvider,
			"",
			nil,
		)

		wantErr := "invalid genesis doc SHA256 checksum: expected 64 characters, but have 14"
		assert.EqualError(t, err, wantErr)
	})
}

func TestGenesisDoc(t *testing.T) {
	var (
		config = test.ResetTestRoot(t.Name())
		n      = &Node{config: config}
	)

	// In the following tests we always overwrite the genesis file with a dummy.
	// We can do so because the method under test's sole responsibility is
	// retrieving and returning the GenesisDoc from disk. Therefore, we test only
	// whether the retrieval process goes as expected; we don't check if the
	// GenesisDoc is valid.

	t.Run("NoError", func(t *testing.T) {
		// A trivial, incomplete genesis to test correct behavior.
		gDocStr := `{
"genesis_time": "2018-10-10T08:20:13.695936996Z",
"chain_id": "test-chain",
"initial_height": "1",
"app_hash": ""
}`

		err := os.WriteFile(config.GenesisFile(), []byte(gDocStr), 0o644)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		wantgDoc := &types.GenesisDoc{
			GenesisTime:   time.Date(2018, 10, 10, 8, 20, 13, 695936996, time.UTC),
			ChainID:       "test-chain",
			InitialHeight: 1,
			AppHash:       []byte{},
		}

		gDoc, err := n.GenesisDoc()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !reflect.DeepEqual(gDoc, wantgDoc) {
			t.Errorf("\nwant: %#v\ngot: %#v\n", wantgDoc, gDoc)
		}
	})

	t.Run("ErrGenesisFilePath", func(t *testing.T) {
		n.config.Genesis = "foo.json"
		_, err := n.GenesisDoc()
		if err == nil {
			t.Fatal("expected error but got none")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected os.ErrNotExist, got %s", err)
		}
	})

	t.Run("ErrGenesisUnmarshal", func(t *testing.T) {
		// A trivial, incomplete genesis where initial_height is set to an invalid
		// value.
		// We don't need anything more complex to test this error.
		gDocStr := `{
"genesis_time": "2018-10-10T08:20:13.695936996Z",
"chain_id": "test-chain",
"initial_height": "hello world",
"app_hash": ""
}`

		// note: Recall that in the previous test we set the path n.config.Genesis to
		// foo.json. Therefore, config.GenesisFile() returns the path to foo.json.
		err := os.WriteFile(config.GenesisFile(), []byte(gDocStr), 0o644)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		_, err = n.GenesisDoc()
		if err == nil {
			t.Fatal("expected error but got none")
		}

		var errUnmarshal *json.SyntaxError
		if !errors.As(err, &errUnmarshal) {
			t.Errorf("expected json.SyntaxError, got %s", err)
		}
	})
}

func state(nVals int, height int64) (sm.State, dbm.DB, []types.PrivValidator) {
	privVals := make([]types.PrivValidator, nVals)
	vals := make([]types.GenesisValidator, nVals)
	for i := 0; i < nVals; i++ {
		privVal := types.NewMockPV()
		privVals[i] = privVal
		vals[i] = types.GenesisValidator{
			Address: privVal.PrivKey.PubKey().Address(),
			PubKey:  privVal.PrivKey.PubKey(),
			Power:   1000,
			Name:    fmt.Sprintf("test%d", i),
		}
	}
	s, _ := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:    "test-chain",
		Validators: vals,
		AppHash:    nil,
	})

	// save validators to db for 2 heights
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	if err := stateStore.Save(s); err != nil {
		panic(err)
	}

	for i := 1; i < int(height); i++ {
		s.LastBlockHeight++
		s.LastValidators = s.Validators.Copy()
		if err := stateStore.Save(s); err != nil {
			panic(err)
		}
	}
	return s, stateDB, privVals
}
