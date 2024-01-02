package node

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/internal/evidence"
	cmtos "github.com/cometbft/cometbft/internal/os"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	sm "github.com/cometbft/cometbft/internal/state"
	"github.com/cometbft/cometbft/internal/store"
	"github.com/cometbft/cometbft/internal/test"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	mempl "github.com/cometbft/cometbft/mempool"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	p2pmock "github.com/cometbft/cometbft/p2p/mock"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

func TestNodeStartStop(t *testing.T) {
	config := test.ResetTestRoot("node_node_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
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
	testCases := map[string]struct {
		s        string
		sep      string
		cutset   string
		expected []string
	}{
		"Simple case":                {"a,b,c", ",", " ", []string{"a", "b", "c"}},
		"Spaces around separators":   {" a , b , c ", ",", " ", []string{"a", "b", "c"}},
		"Spaces after separators":    {" a, b, c ", ",", " ", []string{"a", "b", "c"}},
		"Only one non-empty element": {" a, ", ",", " ", []string{"a"}},
		"All spaces":                 {"   ", ",", " ", []string{}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, splitAndTrimEmpty(tc.s, tc.sep, tc.cutset))
		})
	}
}

func TestCompanionInitialHeightSetup(t *testing.T) {
	config := test.ResetTestRoot("companion_initial_height")
	defer os.RemoveAll(config.RootDir)
	config.Storage.Pruning.DataCompanion.Enabled = true
	config.Storage.Pruning.DataCompanion.InitialBlockRetainHeight = 1
	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)

	companionRetainHeight, err := n.stateStore.GetCompanionBlockRetainHeight()
	require.NoError(t, err)
	require.Equal(t, companionRetainHeight, int64(1))
}

func TestNodeDelayedStart(t *testing.T) {
	// Setup the test environment
	testRoot := "node_delayed_start_test"
	config := test.ResetTestRoot(testRoot)
	defer os.RemoveAll(config.RootDir)

	// Define the delay duration
	delayDuration := 2 * time.Second

	// Create a new node
	node, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err, "Failed to create a new node")

	// Set the genesis time to be 2 seconds from now
	node.GenesisDoc().GenesisTime = cmttime.Now().Add(delayDuration)

	// Start the node
	err = node.Start()
	require.NoError(t, err, "Failed to start the node")

	// Ensure the node is stopped after the test
	defer func() {
		err := node.Stop()
		require.NoError(t, err, "Failed to stop the node")
	}()

	// Get the time after the node start
	startTime := cmttime.Now()

	// Assert that the start time is after the genesis time
	assert.True(t, startTime.After(node.GenesisDoc().GenesisTime), "Start time should be after the genesis time")

	// Assert that the difference between the start time and the genesis time is within a reasonable range
	// We use a large delta to account for potential discrepancies in timing
	assert.InDelta(t, delayDuration.Seconds(), startTime.Sub(node.GenesisDoc().GenesisTime).Seconds(), 2.0, "The difference between start time and genesis time should be within a reasonable range")
}

func TestNodeSetAppVersion(t *testing.T) {
	config := test.ResetTestRoot("node_app_version_test")
	defer os.RemoveAll(config.RootDir)

	// create & start node
	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)

	// default config uses the kvstore app
	appVersion := kvstore.AppVersion

	// check version is set in state
	state, err := n.stateStore.Load()
	require.NoError(t, err)
	assert.Equal(t, state.Version.Consensus.App, appVersion)

	// check version is set in node info
	assert.Equal(t, n.nodeInfo.(p2p.DefaultNodeInfo).ProtocolVersion.App, appVersion)
}

func TestPprofServer(t *testing.T) {
	config := test.ResetTestRoot("node_pprof_test")
	defer os.RemoveAll(config.RootDir)
	config.RPC.PprofListenAddress = testFreeAddr(t)

	// should not work yet
	_, err := http.Get("http://" + config.RPC.PprofListenAddress) //nolint: bodyclose
	assert.Error(t, err)

	n, err := DefaultNewNode(config, log.TestingLogger())
	assert.NoError(t, err)
	assert.NoError(t, n.Start())
	defer func() {
		require.NoError(t, n.Stop())
	}()
	assert.NotNil(t, n.pprofSrv)

	// Retry mechanism
	for i := 0; i < 5; i++ {
		resp, err := http.Get("http://" + config.RPC.PprofListenAddress + "/debug/pprof")
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, 200, resp.StatusCode)
			return
		}
		time.Sleep(time.Second)
	}

	t.Fatal("pprof server not accessible after 5 retries")
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

	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)
	assert.IsType(t, &privval.RetrySignerClient{}, n.PrivValidator())
}

// address without a protocol must result in error
func TestPrivValidatorListenAddrNoProtocol(t *testing.T) {
	addrNoPrefix := testFreeAddr(t)

	config := test.ResetTestRoot("node_priv_val_tcp_test")
	defer os.RemoveAll(config.RootDir)
	config.BaseConfig.PrivValidatorListenAddr = addrNoPrefix

	_, err := DefaultNewNode(config, log.TestingLogger())
	assert.Error(t, err)
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

	n, err := DefaultNewNode(config, log.TestingLogger())
	require.NoError(t, err)
	assert.IsType(t, &privval.RetrySignerClient{}, n.PrivValidator())
}

// testFreeAddr claims a free port so we don't block on listener being ready.
func testFreeAddr(t *testing.T) string {
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
	cc := proxy.NewLocalClientCreator(kvstore.NewInMemoryApplication())
	proxyApp := proxy.NewAppConns(cc, proxy.NopMetrics())
	err := proxyApp.Start()
	require.Nil(t, err)
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
	memplMetrics := mempl.NopMetrics()
	mempool := mempl.NewCListMempool(config.Mempool,
		proxyApp.Mempool(),
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
		ev, err := types.NewMockDuplicateVoteEvidenceWithValidator(height, time.Now(), privVals[0], "test-chain")
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
		_, err := mempool.CheckTx(tx)
		assert.NoError(t, err)
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
	assert.NoError(t, err)
}

func TestMaxProposalBlockSize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := test.ResetTestRoot("node_create_proposal")
	defer os.RemoveAll(config.RootDir)
	cc := proxy.NewLocalClientCreator(kvstore.NewInMemoryApplication())
	proxyApp := proxy.NewAppConns(cc, proxy.NopMetrics())
	err := proxyApp.Start()
	require.Nil(t, err)
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
	memplMetrics := mempl.NopMetrics()
	mempool := mempl.NewCListMempool(config.Mempool,
		proxyApp.Mempool(),
		state.LastBlockHeight,
		mempl.WithMetrics(memplMetrics),
		mempl.WithPreCheck(sm.TxPreCheck(state)),
		mempl.WithPostCheck(sm.TxPostCheck(state)))

	blockStore := store.NewBlockStore(dbm.NewMemDB())

	// fill the mempool with one txs just below the maximum size
	txLength := int(types.MaxDataBytesNoEvidence(maxBytes, 1))
	tx := cmtrand.Bytes(txLength - 4) // to account for the varint
	_, err = mempool.CheckTx(tx)
	assert.NoError(t, err)

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

// TestNodeNewNodeCustomReactors tests that custom reactors can be added to a new node
// and that they are running after the node is started. It also checks that the custom
// reactors' channels are included in the node's channels.
func TestNodeNewNodeCustomReactors(t *testing.T) {
	// Reset the test root and ensure it's cleaned up after the test
	config := test.ResetTestRoot("node_new_node_custom_reactors_test")
	defer os.RemoveAll(config.RootDir)

	// Create a new custom reactor
	cr := p2pmock.NewReactor()
	cr.Channels = []*conn.ChannelDescriptor{
		{
			ID:                  byte(0x31),
			Priority:            5,
			SendQueueCapacity:   100,
			RecvMessageCapacity: 100,
		},
	}
	// Create another custom reactor
	customBlocksyncReactor := p2pmock.NewReactor()

	// Load or generate a new node key
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	// Create a new node with the custom reactors
	n, err := NewNode(context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
		CustomReactors(map[string]p2p.Reactor{"FOO": cr, "BLOCKSYNC": customBlocksyncReactor}),
	)
	require.NoError(t, err)

	// Start the node
	err = n.Start()
	require.NoError(t, err)
	defer n.Stop() //nolint:errcheck // ignore for tests

	// Check that the custom reactors are running
	assert.True(t, cr.IsRunning())
	assert.True(t, customBlocksyncReactor.IsRunning())

	// Check that the custom reactors are part of the node's reactors
	assert.Equal(t, cr, n.Switch().Reactor("FOO"))
	assert.Equal(t, customBlocksyncReactor, n.Switch().Reactor("BLOCKSYNC"))

	// Check that the custom reactors' channels are part of the node's channels
	channels := n.NodeInfo().(p2p.DefaultNodeInfo).Channels
	assert.Contains(t, channels, mempl.MempoolChannel)
	assert.Contains(t, channels, cr.Channels[0].ID)
}

// Simple test to confirm that an existing genesis file will be deleted from the DB
// TODO Confirm that the deletion of a very big file does not crash the machine
func TestNodeNewNodeDeleteGenesisFileFromDB(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_delete_genesis_from_db")
	defer os.RemoveAll(config.RootDir)
	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.GoLevelDBBackend)
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

	n, err := NewNode(
		context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.NoError(t, err)

	_, err = stateDB.Get(genesisDocKey)
	require.Error(t, err)

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

// TestNodeNewNodeGenesisHashMismatch tests that creating a new node with a modified genesis file
// that results in a different hash from the one saved in the database will result in an error.
func TestNodeNewNodeGenesisHashMismatch(t *testing.T) {
	// Reset the test root and ensure it's cleaned up after the test
	config := test.ResetTestRoot("node_new_node_genesis_hash")
	defer os.RemoveAll(config.RootDir)

	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.GoLevelDBBackend)

	// Load or generate a new node key
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err, "Failed to load or generate node key")

	// Create and start a new node
	_, err = createAndStartNode(config, nodeKey)
	require.NoError(t, err, "Failed to create and start a new node")

	// Ensure the genesis doc hash is saved to db
	genDocHash, err := getGenesisDocHashFromDB(config)
	require.NoError(t, err, "Failed to get genesis doc hash from DB")
	require.NotNil(t, genDocHash, "genesis doc hash should be saved in db")
	require.Len(t, genDocHash, tmhash.Size, "Invalid genesis doc hash size")

	// Modify the genesis file chain ID to get a different hash
	err = modifyGenesisFileChainID(config)
	require.NoError(t, err, "Failed to modify genesis file chain ID")

	// Attempt to create a new node with the modified genesis file
	_, err = NewNode(
		context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	// Expect an error because the genesis doc hash in the db does not match the loaded genesis doc
	require.Error(t, err, "NewNode should error when genesisDoc is changed")
	require.Equal(t, "genesis doc hash in db does not match loaded genesis doc", err.Error())
}

func createAndStartNode(config *cfg.Config, nodeKey *p2p.NodeKey) (*Node, error) {
	n, err := NewNode(
		context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)

	if err != nil {
		return nil, err
	}

	err = n.Start()
	if err != nil {
		return nil, err
	}

	err = n.Stop()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func getGenesisDocHashFromDB(config *cfg.Config) ([]byte, error) {
	stateDB, err := cfg.DefaultDBProvider(&cfg.DBContext{ID: "state", Config: config})
	if err != nil {
		return nil, err
	}
	defer stateDB.Close()

	genDocHash, err := stateDB.Get(genesisDocHashKey)
	if err != nil {
		return nil, err
	}

	return genDocHash, nil
}

func modifyGenesisFileChainID(config *cfg.Config) error {
	genBytes := cmtos.MustReadFile(config.GenesisFile())
	var genesisDoc types.GenesisDoc
	err := cmtjson.Unmarshal(genBytes, &genesisDoc)
	if err != nil {
		return err
	}

	genesisDoc.ChainID = "different-chain-id"
	err = genesisDoc.SaveAs(config.GenesisFile())
	if err != nil {
		return err
	}

	return nil
}

func TestNodeGenesisHashFlagMatch(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_genesis_hash_flag_match")
	defer os.RemoveAll(config.RootDir)

	config.DBBackend = string(dbm.GoLevelDBBackend)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)
	// Get correct hash of correct genesis file
	jsonBlob, err := os.ReadFile(config.GenesisFile())
	require.NoError(t, err)

	incomingChecksum := tmhash.Sum(jsonBlob)
	// Set genesis flag value to incorrect hash
	config.Storage.GenesisHash = hex.EncodeToString(incomingChecksum)
	_, err = NewNode(
		context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.NoError(t, err)
}

func TestNodeGenesisHashFlagMismatch(t *testing.T) {
	config := test.ResetTestRoot("node_new_node_genesis_hash_flag_mismatch")
	defer os.RemoveAll(config.RootDir)

	// Use goleveldb so we can reuse the same db for the second NewNode()
	config.DBBackend = string(dbm.GoLevelDBBackend)

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	require.NoError(t, err)

	// Generate hash of wrong file
	f, err := os.ReadFile(config.PrivValidatorKeyFile())
	require.NoError(t, err)
	flagHash := tmhash.Sum(f)

	// Set genesis flag value to incorrect hash
	config.Storage.GenesisHash = hex.EncodeToString(flagHash)

	_, err = NewNode(
		context.Background(),
		config,
		privval.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()),
		nodeKey,
		proxy.DefaultClientCreator(config.ProxyApp, config.ABCI, config.DBDir()),
		DefaultGenesisDocProviderFunc(config),
		cfg.DefaultDBProvider,
		DefaultMetricsProvider(config.Instrumentation),
		log.TestingLogger(),
	)
	require.Error(t, err)

	f, err = os.ReadFile(config.GenesisFile())
	require.NoError(t, err)

	genHash := tmhash.Sum(f)

	genHashMismatch := bytes.Equal(genHash, flagHash)
	require.False(t, genHashMismatch)
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
