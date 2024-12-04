package mempool

import (
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abci "github.com/cometbft/cometbft/abci/types"
	memproto "github.com/cometbft/cometbft/api/cometbft/mempool/v1"
	cfg "github.com/cometbft/cometbft/config"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

const (
	numTxs  = 1000
	timeout = 120 * time.Second // ridiculously high because CircleCI is slow
)

type peerState struct {
	height int64
}

func (ps peerState) GetHeight() int64 {
	return ps.height
}

// Send a bunch of txs to the first reactor's mempool and wait for them all to
// be received in the others.
func TestReactorBroadcastTxsMessage(t *testing.T) {
	config := cfg.TestConfig()
	// if there were more than two reactors, the order of transactions could not be
	// asserted in waitForTxsOnReactors (due to transactions gossiping). If we
	// replace Connect2Switches (full mesh) with a func, which connects first
	// reactor to others and nothing else, this test should also pass with >2 reactors.
	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	txs := addRandomTxs(t, reactors[0].mempool, numTxs)
	waitForReactors(t, txs, reactors, checkTxsInMempool)
}

// regression test for https://github.com/tendermint/tendermint/issues/5408
func TestReactorConcurrency(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.Size = 5000
	config.Mempool.CacheSize = 5000
	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}
	var wg sync.WaitGroup

	const numTxs = 5

	for i := 0; i < 1000; i++ {
		wg.Add(2)

		// 1. submit a bunch of txs
		// 2. update the whole mempool
		txs := addRandomTxs(t, reactors[0].mempool, numTxs)
		go func() {
			defer wg.Done()

			reactors[0].mempool.PreUpdate()
			reactors[0].mempool.Lock()
			defer reactors[0].mempool.Unlock()

			err := reactors[0].mempool.Update(1, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
			require.NoError(t, err)
		}()

		// 1. submit a bunch of txs
		// 2. update none
		_ = addRandomTxs(t, reactors[1].mempool, numTxs)
		go func() {
			defer wg.Done()

			reactors[1].mempool.PreUpdate()
			reactors[1].mempool.Lock()
			defer reactors[1].mempool.Unlock()
			err := reactors[1].mempool.Update(1, []types.Tx{}, make([]*abci.ExecTxResult, 0), nil, nil)
			require.NoError(t, err)
		}()

		// 1. flush the mempool
		reactors[1].mempool.Flush()
	}

	wg.Wait()
}

// Send a bunch of txs to the first reactor's mempool, claiming it came from peer
// ensure peer gets no txs.
func TestReactorNoBroadcastToSender(t *testing.T) {
	config := cfg.TestConfig()
	const n = 2
	reactors, _ := makeAndConnectReactorsNoLanes(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// create random transactions
	txs := NewRandomTxs(numTxs, 20)

	// This subset should be broadcast
	var txsToBroadcast types.Txs
	const minToBroadcast = numTxs / 10

	// The second peer sends some transactions to the first peer
	secondNodeID := reactors[1].Switch.NodeInfo().ID()
	secondNode := reactors[0].Switch.Peers().Get(secondNodeID)
	for i, tx := range txs {
		shouldBroadcast := cmtrand.Bool() || // random choice
			// Force shouldBroadcast == true to ensure that
			// len(txsToBroadcast) >= minToBroadcast
			(len(txsToBroadcast) < minToBroadcast &&
				len(txs)-i <= minToBroadcast)

		t.Log(i, "adding", tx, "shouldBroadcast", shouldBroadcast)

		if !shouldBroadcast {
			// From the second peer => should not be broadcast
			_, err := reactors[0].TryAddTx(tx, secondNode)
			require.NoError(t, err)
		} else {
			// Emulate a tx received via RPC => should broadcast
			_, err := reactors[0].TryAddTx(tx, nil)
			require.NoError(t, err)
			txsToBroadcast = append(txsToBroadcast, tx)
		}
	}

	t.Log("Added", len(txs), "transactions, only", len(txsToBroadcast),
		"should be sent to the peer")

	// The second peer should receive only txsToBroadcast transactions
	waitForReactors(t, txsToBroadcast, reactors[1:], checkTxsInOrder)
}

// Test that a lagging peer does not receive txs.
func TestMempoolReactorSendLaggingPeer(t *testing.T) {
	config := cfg.TestConfig()
	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()

	// First reactor is at height 10 and knows that its peer is lagging at height 1.
	reactors[0].mempool.height.Store(10)
	peerID := reactors[1].Switch.NodeInfo().ID()
	reactors[0].Switch.Peers().Get(peerID).Set(types.PeerStateKey, peerState{1})

	// Add a bunch of txs to the first reactor. The second reactor should not receive any tx.
	txs1 := addTxs(t, reactors[0].mempool, 0, numTxs)
	ensureNoTxs(t, reactors[1], 5*PeerCatchupSleepIntervalMS*time.Millisecond)

	// Now we know that the second reactor has advanced to height 9, so it should receive all txs.
	reactors[0].Switch.Peers().Get(peerID).Set(types.PeerStateKey, peerState{9})
	waitForReactors(t, txs1, reactors, checkTxsInMempool)

	// Add a bunch of txs to first reactor. The second reactor should receive them all.
	txs2 := addTxs(t, reactors[0].mempool, numTxs, numTxs)
	waitForReactors(t, append(txs1, txs2...), reactors, checkTxsInMempool)
}

// Test the scenario where a tx selected for being sent to a peer is removed
// from the mempool before it is actually sent.
func TestMempoolReactorSendRemovedTx(t *testing.T) {
	config := cfg.TestConfig()
	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()

	// First reactor is at height 10 and knows that its peer is lagging at height 1.
	// We do this to hold sending transactions, giving us time to remove some of them.
	reactors[0].mempool.height.Store(10)
	peerID := reactors[1].Switch.NodeInfo().ID()
	reactors[0].Switch.Peers().Get(peerID).Set(types.PeerStateKey, peerState{1})

	// Add a bunch of txs to the first reactor. The second reactor should not receive any tx.
	txs := addRandomTxs(t, reactors[0].mempool, 20)
	ensureNoTxs(t, reactors[1], 5*PeerCatchupSleepIntervalMS*time.Millisecond)

	// Remove some txs from the mempool of the first reactor.
	txsToRemove := txs[:10]
	txsLeft := txs[10:]
	reactors[0].mempool.PreUpdate()
	reactors[0].mempool.Lock()
	err := reactors[0].mempool.Update(10, txsToRemove, abciResponses(len(txsToRemove), abci.CodeTypeOK), nil, nil)
	require.NoError(t, err)
	reactors[0].mempool.Unlock()
	require.Equal(t, len(txsLeft), reactors[0].mempool.Size())

	// Now we know that the second reactor is not lagging, so it should receive
	// all txs except those that were removed.
	reactors[0].Switch.Peers().Get(peerID).Set(types.PeerStateKey, peerState{9})
	waitForReactors(t, txsLeft, reactors, checkTxsInMempool)
}

func TestMempoolReactorMaxTxBytes(t *testing.T) {
	config := cfg.TestConfig()

	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, mempoolLogger("info"))
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Broadcast a tx, which has the max size
	// => ensure it's received by the second reactor.
	tx1 := kvstore.NewRandomTx(config.Mempool.MaxTxBytes)
	reqRes, err := reactors[0].TryAddTx(tx1, nil)
	require.NoError(t, err)
	require.False(t, reqRes.Response.GetCheckTx().IsErr())
	waitForReactors(t, []types.Tx{tx1}, reactors, checkTxsInOrder)

	reactors[0].mempool.Flush()
	reactors[1].mempool.Flush()

	// Broadcast a tx, which is beyond the max size
	// => ensure it's not sent
	tx2 := kvstore.NewRandomTx(config.Mempool.MaxTxBytes + 1)
	reqRes, err = reactors[0].TryAddTx(tx2, nil)
	require.Error(t, err)
	require.Nil(t, reqRes)
}

func TestBroadcastTxForPeerStopsWhenPeerStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	config := cfg.TestConfig()
	const n = 2
	reactors, _ := makeAndConnectReactors(config, n, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()

	// stop peer
	sw := reactors[1].Switch
	sw.StopPeerForError(sw.Peers().Copy()[0], errors.New("some reason"))

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when peer is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

func TestBroadcastTxForPeerStopsWhenReactorStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	config := cfg.TestConfig()
	const n = 2
	_, switches := makeAndConnectReactors(config, n, nil)

	// stop reactors
	for _, s := range switches {
		require.NoError(t, s.Stop())
	}

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when reactor is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

// Finding a solution for guaranteeing FIFO ordering is not easy; it would
// require changes at the p2p level. The order of messages is just best-effort,
// but this is not documented anywhere. If this is well understood and
// documented, we don't need this test. Until then, let's keep the test.
func TestMempoolFIFOWithParallelCheckTx(t *testing.T) {
	t.Skip("FIFO is not supposed to be guaranteed and this is just used to evidence one of the cases where it does not happen. Hence we skip this test.")

	config := cfg.TestConfig()
	reactors, _ := makeAndConnectReactors(config, 4, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Deliver the same sequence of transactions from multiple sources, in parallel.
	txs := newUniqueTxs(200)
	for i := 0; i < 3; i++ {
		go func() {
			for _, tx := range txs {
				_, _ = reactors[0].TryAddTx(tx, nil)
			}
		}()
	}

	// Confirm that FIFO order was respected.
	checkTxsInOrder(t, txs, reactors[0], 0)
}

// Test the experimental feature that limits the number of outgoing connections for gossiping
// transactions (only non-persistent peers).
// Note: in this test we know which gossip connections are active or not because of how the p2p
// functions are currently implemented, which affects the order in which peers are added to the
// mempool reactor.
func TestMempoolReactorMaxActiveOutboundConnections(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.ExperimentalMaxGossipConnectionsToNonPersistentPeers = 1
	reactors, _ := makeAndConnectReactors(config, 4, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Add a bunch transactions to the first reactor.
	txs := newUniqueTxs(100)
	tryAddTxs(t, reactors[0], txs)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInMempool(t, txs, reactors[1], 0)
	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	firstPeer := reactors[0].Switch.Peers().Copy()[0]
	reactors[0].Switch.StopPeerGracefully(firstPeer)

	// Now the third reactor should start receiving transactions from the first reactor; the fourth
	// reactor's mempool should still be empty.
	checkTxsInMempool(t, txs, reactors[2], 0)
	for _, r := range reactors[3:] {
		require.Zero(t, r.mempool.Size())
	}
}

// Test the experimental feature that limits the number of outgoing connections for gossiping
// transactions (only non-persistent peers).
// Given the disconnections, no transaction should be received in duplicate.
// Note: in this test we know which gossip connections are active or not because of how the p2p
// functions are currently implemented, which affects the order in which peers are added to the
// mempool reactor.
func TestMempoolReactorMaxActiveOutboundConnectionsNoDuplicate(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.ExperimentalMaxGossipConnectionsToNonPersistentPeers = 1
	reactors, _ := makeAndConnectReactors(config, 4, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Disconnect the second reactor from the third reactor.
	pCon1_2 := reactors[1].Switch.Peers().Copy()[1]
	reactors[1].Switch.StopPeerGracefully(pCon1_2)

	// Add a bunch transactions to the first reactor.
	txs := newUniqueTxs(100)
	tryAddTxs(t, reactors[0], txs)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInMempool(t, txs, reactors[1], 0)
	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	pCon0_1 := reactors[0].Switch.Peers().Copy()[0]
	reactors[0].Switch.StopPeerGracefully(pCon0_1)

	// Now the third reactor should start receiving transactions from the first reactor and
	// the fourth reactor from the second
	checkTxsInMempool(t, txs, reactors[2], 0)
	checkTxsInMempool(t, txs, reactors[3], 0)
}

// Test the experimental feature that limits the number of outgoing connections for gossiping
// transactions (only non-persistent peers) on a star shaped network.
// The star center will need to deliver the transactions to each point.
// Note: in this test we know which gossip connections are active or not because of how the p2p
// functions are currently implemented, which affects the order in which peers are added to the
// mempool reactor.
func TestMempoolReactorMaxActiveOutboundConnectionsStar(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.ExperimentalMaxGossipConnectionsToNonPersistentPeers = 1
	reactors, _ := makeAndConnectReactorsStar(config, 0, 4, nil)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().Copy() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}
	// Add a bunch transactions to the first reactor.
	txs := newUniqueTxs(5)
	tryAddTxs(t, reactors[0], txs)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInMempool(t, txs, reactors[0], 0)
	checkTxsInMempool(t, txs, reactors[1], 0)

	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	firstPeer := reactors[0].Switch.Peers().Copy()[0]
	reactors[0].Switch.StopPeerGracefully(firstPeer)

	// Now the third reactor should start receiving transactions from the first reactor; the fourth
	// reactor's mempool should still be empty.
	checkTxsInMempool(t, txs, reactors[0], 0)
	checkTxsInMempool(t, txs, reactors[1], 0)
	checkTxsInMempool(t, txs, reactors[2], 0)
	for _, r := range reactors[3:] {
		require.Zero(t, r.mempool.Size())
	}
}

// mempoolLogger is a TestingLogger which uses a different
// color for each validator ("validator" key must exist).
func mempoolLogger(level string) *log.Logger {
	logger := log.TestingLogger()

	// Customize log level
	option, err := log.AllowLevel(level)
	if err != nil {
		panic(err)
	}
	logger = log.NewFilter(logger, option)

	return &logger
}

// makeReactors creates n mempool reactors.
func makeReactors(config *cfg.Config, n int, logger *log.Logger, lanesEnabled bool) []*Reactor {
	if logger == nil {
		logger = mempoolLogger("info")
	}
	reactors := make([]*Reactor, n)
	for i := 0; i < n; i++ {
		var app *kvstore.Application
		if lanesEnabled {
			app = kvstore.NewInMemoryApplication()
		} else {
			app = kvstore.NewInMemoryApplicationWithoutLanes()
		}
		cc := proxy.NewLocalClientCreator(app)
		mempool, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		reactors[i] = NewReactor(config.Mempool, mempool, false) // so we dont start the consensus states
		reactors[i].SetLogger((*logger).With("validator", i))
	}
	return reactors
}

// connectReactors connects the list of N reactors through N switches.
func connectReactors(config *cfg.Config, reactors []*Reactor, connect func([]*p2p.Switch, int, int)) []*p2p.Switch {
	switches := p2p.MakeSwitches(config.P2P, len(reactors), func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", reactors[i])
		return s
	})
	for _, s := range switches {
		s.SetLogger(log.NewNopLogger())
	}
	return p2p.StartAndConnectSwitches(switches, connect)
}

func makeAndConnectReactorsNoLanes(config *cfg.Config, n int, logger *log.Logger) ([]*Reactor, []*p2p.Switch) {
	reactors := makeReactors(config, n, logger, false)
	switches := connectReactors(config, reactors, p2p.Connect2Switches)
	return reactors, switches
}

func makeAndConnectReactors(config *cfg.Config, n int, logger *log.Logger) ([]*Reactor, []*p2p.Switch) {
	reactors := makeReactors(config, n, logger, true)
	switches := connectReactors(config, reactors, p2p.Connect2Switches)
	return reactors, switches
}

// connect N mempool reactors through N switches as a star centered in c.
func makeAndConnectReactorsStar(config *cfg.Config, c, n int, logger *log.Logger) ([]*Reactor, []*p2p.Switch) {
	reactors := makeReactors(config, n, logger, true)
	switches := connectReactors(config, reactors, p2p.ConnectStarSwitches(c))
	return reactors, switches
}

func newUniqueTxs(n int) types.Txs {
	txs := make(types.Txs, n)
	for i := 0; i < n; i++ {
		txs[i] = kvstore.NewTxFromID(i)
	}
	return txs
}

// Wait for all reactors to finish applying a testing function to a list of
// transactions.
func waitForReactors(t *testing.T, txs types.Txs, reactors []*Reactor, testFunc func(*testing.T, types.Txs, *Reactor, int)) {
	t.Helper()
	wg := new(sync.WaitGroup)
	for i, reactor := range reactors {
		wg.Add(1)
		go func(r *Reactor, reactorIndex int) {
			defer wg.Done()
			testFunc(t, txs, r, reactorIndex)
		}(reactor, i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timer := time.After(timeout)
	select {
	case <-timer:
		t.Fatal("Timed out waiting for txs")
	case <-done:
	}
}

// Wait until the mempool has a certain number of transactions.
func waitForNumTxsInMempool(numTxs int, mempool Mempool) {
	for mempool.Size() < numTxs {
		time.Sleep(time.Millisecond * 100)
	}
}

// Wait until all txs are in the mempool and check that the number of txs in the
// mempool is as expected.
func checkTxsInMempool(t *testing.T, txs types.Txs, reactor *Reactor, _ int) {
	t.Helper()
	waitForNumTxsInMempool(len(txs), reactor.mempool)

	reapedTxs := reactor.mempool.ReapMaxTxs(len(txs))
	require.Len(t, txs, len(reapedTxs))
	require.Len(t, txs, reactor.mempool.Size())
}

// Wait until all txs are in the mempool and check that they are in the same
// order as given.
func checkTxsInOrder(t *testing.T, txs types.Txs, reactor *Reactor, reactorIndex int) {
	t.Helper()
	waitForNumTxsInMempool(len(txs), reactor.mempool)

	// Check that all transactions in the mempool are in the same order as txs.
	reapedTxs := reactor.mempool.ReapMaxTxs(len(txs))
	require.Equal(t, len(txs), len(reapedTxs))
	for i, tx := range txs {
		assert.Equalf(t, tx, reapedTxs[i],
			"txs at index %d on reactor %d don't match: %v vs %v", i, reactorIndex, tx, reapedTxs[i])
	}
}

// ensure no txs on reactor after some timeout.
func ensureNoTxs(t *testing.T, reactor *Reactor, timeout time.Duration) {
	t.Helper()
	time.Sleep(timeout) // wait for the txs in all mempools
	assert.Zero(t, reactor.mempool.Size())
}

// Try to add a list of transactions to the mempool of a given reactor.
func tryAddTxs(t *testing.T, reactor *Reactor, txs types.Txs) {
	t.Helper()
	for _, tx := range txs {
		rr, err := reactor.TryAddTx(tx, nil)
		require.Nil(t, err)
		rr.Wait()
	}
}

func TestMempoolVectors(t *testing.T) {
	testCases := []struct {
		testName string
		tx       []byte
		expBytes string
	}{
		{"tx 1", []byte{123}, "0a030a017b"},
		{"tx 2", []byte("proto encoding in mempool"), "0a1b0a1970726f746f20656e636f64696e6720696e206d656d706f6f6c"},
	}

	for _, tc := range testCases {
		msg := memproto.Message{
			Sum: &memproto.Message_Txs{
				Txs: &memproto.Txs{Txs: [][]byte{tc.tx}},
			},
		}
		bz, err := msg.Marshal()
		require.NoError(t, err, tc.testName)

		require.Equal(t, tc.expBytes, hex.EncodeToString(bz), tc.testName)
	}
}

// Verify that counting of duplicates and first time transactions work
// The test sends transactions from node2 to node1 twice.
// The second time they will get rejected.
func TestDOGTransactionCount(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.DOGProtocolEnabled = true

	// Put the interval to a higher value to make sure the values don't get reset
	config.Mempool.DOGAdjustInterval = 15 * time.Second
	reactors, _ := makeAndConnectReactors(config, 2, nil)

	// create random transactions
	txs := newUniqueTxs(numTxs)
	secondNodeID := reactors[1].Switch.NodeInfo().ID()
	secondNode := reactors[0].Switch.Peers().Get(secondNodeID)

	for _, tx := range txs {
		_, err := reactors[0].TryAddTx(tx, secondNode)
		require.NoError(t, err)
	}

	require.Equal(t, int64(len(txs)), reactors[0].redundancyControl.firstTimeTxs)
	for _, tx := range txs {
		_, err := reactors[0].TryAddTx(tx, secondNode)
		// The transaction is in cache, hence the Error
		require.Error(t, err)
	}
	require.Equal(t, int64(len(txs)), reactors[0].redundancyControl.duplicateTxs)

	reactors[0].redundancyControl.triggerAdjustment(reactors[0])
	time.Sleep(1 * time.Second)

	reactors[0].redundancyControl.mtx.RLock()
	dupTx := reactors[0].redundancyControl.duplicateTxs
	firstTimeTx := reactors[0].redundancyControl.firstTimeTxs
	reactors[0].redundancyControl.mtx.RUnlock()

	// Now the counters should be reset
	require.Equal(t, int64(0), dupTx)
	require.Equal(t, int64(0), firstTimeTx)
}

// Testing the disabled route between two nodes
// AS the description of DOG in the issue:
// The core idea of the protocol is the following.
// Consider a node A that receives from node B a transaction
// that it already has. Let's assume B itself had received
// the transaction from C. The fact that A received from B
// a transaction it already has means that there must exist
// a cycle in the network topology.
// Therefore, a tells B to stop sending transactions B would be getting from C
// (i.e. A tells B to disable route C → A → B).
// We then reduce the redundancy level forcing A to tell B to re-enable the routes.
func TestDOGDisabledRoute(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.DOGProtocolEnabled = true

	// Put the interval to a higher value to make sure the values don't get reset
	config.Mempool.DOGAdjustInterval = 35 * time.Second
	reactors, _ := makeAndConnectReactors(config, 3, nil)

	secondNodeID := reactors[1].Switch.NodeInfo().ID()
	secondNode := reactors[0].Switch.Peers().Get(secondNodeID)
	secondNodeFromThird := reactors[2].Switch.Peers().Get(secondNodeID)

	thirdNodeID := reactors[2].Switch.NodeInfo().ID()
	thirdNodeFromFirst := reactors[0].Switch.Peers().Get(thirdNodeID)

	firstNodeID := reactors[0].Switch.NodeInfo().ID()
	firstNodeFromThird := reactors[2].Switch.Peers().Get(firstNodeID)

	// create random transactions
	txs := newUniqueTxs(numTxs)
	// Add transactions to node 3 from node 2
	// node3.senders[tx] = node2
	for _, tx := range txs {
		_, err := reactors[2].TryAddTx(tx, secondNodeFromThird)
		require.NoError(t, err)
	}

	// Add the same transactions to node 1 from node 2
	for _, tx := range txs {
		_, err := reactors[0].TryAddTx(tx, secondNode)
		require.NoError(t, err)
	}

	// Trying to add the same transactions node 1 has received
	// from node 2, but this time from node 3
	// Node 1 should now ask node 3 to disable the route between
	// a node that has sent this tx to node 3(node 2) and node1
	for _, tx := range txs {
		_, err := reactors[0].TryAddTx(tx, thirdNodeFromFirst)
		// The transaction is in cache, hence the Error
		require.ErrorIs(t, err, ErrTxInCache)
	}

	reactors[0].redundancyControl.triggerAdjustment(reactors[0])
	// Wait for the redundancy adjustment to kick in
	time.Sleep(1 * time.Second)

	reactors[2].router.mtx.RLock()
	// Make sure that Node 3 has at least one disabled route
	require.Greater(t, len(reactors[2].router.disabledRoutes), 0)

	require.True(t, reactors[2].router.isRouteDisabled(secondNodeFromThird.ID(), firstNodeFromThird.ID()))
	reactors[2].router.mtx.RUnlock()

	// This will force Node 3 to delete all disabled routes
	reactors[2].Switch.StopPeerGracefully(secondNode)

	// The route should not be there
	require.False(t, reactors[2].router.isRouteDisabled(secondNodeFromThird.ID(), firstNodeFromThird.ID()))
}

// When a peer disconnects we want to remove all disabled route info
// for that peer only.
func TestDOGRemoveDisabledRoutesOnDisconnect(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.DOGProtocolEnabled = true

	reactors, _ := makeAndConnectReactors(config, 4, nil)

	fourthNodeID := reactors[3].Switch.NodeInfo().ID()

	secondNodeID := reactors[1].Switch.NodeInfo().ID()
	secondNode := reactors[0].Switch.Peers().Get(secondNodeID)

	thirdNodeID := reactors[2].Switch.NodeInfo().ID()

	reactors[0].router.disableRoute(secondNodeID, fourthNodeID)
	reactors[0].router.disableRoute(thirdNodeID, fourthNodeID)
	reactors[0].router.disableRoute(thirdNodeID, secondNodeID)

	require.True(t, reactors[0].router.isRouteDisabled(secondNodeID, fourthNodeID))
	require.True(t, reactors[0].router.isRouteDisabled(thirdNodeID, fourthNodeID))
	require.True(t, reactors[0].router.isRouteDisabled(thirdNodeID, secondNodeID))

	reactors[0].Switch.StopPeerGracefully(secondNode)

	require.False(t, reactors[0].router.isRouteDisabled(secondNodeID, fourthNodeID))
	require.False(t, reactors[0].router.isRouteDisabled(thirdNodeID, secondNodeID))
	require.True(t, reactors[0].router.isRouteDisabled(thirdNodeID, fourthNodeID))
}

// Test redundancy values depending on Number of transactions.
func TestDOGTestRedundancyCalculation(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.DOGProtocolEnabled = true
	config.Mempool.DOGTargetRedundancy = 0.5
	reactors, _ := makeAndConnectReactors(config, 1, nil)

	redundancy := reactors[0].redundancyControl.currentRedundancy()
	require.Equal(t, redundancy, float64(-1))

	reactors[0].redundancyControl.firstTimeTxs = 10
	reactors[0].redundancyControl.duplicateTxs = 0
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Equal(t, redundancy, float64(0))

	reactors[0].redundancyControl.duplicateTxs = 1000
	reactors[0].redundancyControl.firstTimeTxs = 10
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Greater(t, redundancy, config.Mempool.DOGTargetRedundancy)

	reactors[0].redundancyControl.duplicateTxs = 1000
	reactors[0].redundancyControl.firstTimeTxs = 0
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Equal(t, redundancy, reactors[0].redundancyControl.upperBound)
}
