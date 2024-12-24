package mempool

import (
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/go-kit/log/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/mock"
	memproto "github.com/cometbft/cometbft/proto/tendermint/mempool"
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
	const N = 2
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	txs := addRandomTxs(t, reactors[0].mempool, numTxs, UnknownPeerID)
	waitForTxsOnReactors(t, txs, reactors)
}

// regression test for https://github.com/cometbft/cometbft/issues/5408
func TestReactorConcurrency(t *testing.T) {
	config := cfg.TestConfig()
	const N = 2
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}
	var wg sync.WaitGroup

	const numTxs = 5

	for i := 0; i < 1000; i++ {
		wg.Add(2)

		// 1. submit a bunch of txs
		// 2. update the whole mempool
		txs := addRandomTxs(t, reactors[0].mempool, numTxs, UnknownPeerID)
		go func() {
			defer wg.Done()

			reactors[0].mempool.Lock()
			defer reactors[0].mempool.Unlock()

			txResponses := make([]*abci.ExecTxResult, len(txs))
			for i := range txs {
				txResponses[i] = &abci.ExecTxResult{Code: 0}
			}
			err := reactors[0].mempool.Update(1, txs, txResponses, nil, nil)
			assert.NoError(t, err)
		}()

		// 1. submit a bunch of txs
		// 2. update none
		_ = addRandomTxs(t, reactors[1].mempool, numTxs, UnknownPeerID)
		go func() {
			defer wg.Done()

			reactors[1].mempool.Lock()
			defer reactors[1].mempool.Unlock()
			err := reactors[1].mempool.Update(1, []types.Tx{}, make([]*abci.ExecTxResult, 0), nil, nil)
			assert.NoError(t, err)
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
	const N = 2
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	const peerID = 1
	addRandomTxs(t, reactors[0].mempool, numTxs, peerID)
	ensureNoTxs(t, reactors[peerID], 100*time.Millisecond)
}

func TestMempoolReactorMaxTxBytes(t *testing.T) {
	config := cfg.TestConfig()

	const N = 2
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Broadcast a tx, which has the max size
	// => ensure it's received by the second reactor.
	tx1 := kvstore.NewRandomTx(config.Mempool.MaxTxBytes)
	err := reactors[0].mempool.CheckTx(tx1, func(resp *abci.ResponseCheckTx) {
		require.False(t, resp.IsErr())
	}, TxInfo{SenderID: UnknownPeerID})
	require.NoError(t, err)
	waitForTxsOnReactors(t, []types.Tx{tx1}, reactors)

	reactors[0].mempool.Flush()
	reactors[1].mempool.Flush()

	// Broadcast a tx, which is beyond the max size
	// => ensure it's not sent
	tx2 := kvstore.NewRandomTx(config.Mempool.MaxTxBytes + 1)
	err = reactors[0].mempool.CheckTx(tx2, func(resp *abci.ResponseCheckTx) {
		require.False(t, resp.IsErr())
	}, TxInfo{SenderID: UnknownPeerID})
	require.Error(t, err)
}

func TestBroadcastTxForPeerStopsWhenPeerStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	config := cfg.TestConfig()
	const N = 2
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()

	// stop peer
	sw := reactors[1].Switch
	sw.StopPeerForError(sw.Peers().List()[0], errors.New("some reason"))

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when peer is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

func TestBroadcastTxForPeerStopsWhenReactorStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	config := cfg.TestConfig()
	const N = 2
	_, switches := makeAndConnectReactors(config, N)

	// stop reactors
	for _, s := range switches {
		assert.NoError(t, s.Stop())
	}

	// check that we are not leaking any go-routines
	// i.e. broadcastTxRoutine finishes when reactor is stopped
	leaktest.CheckTimeout(t, 10*time.Second)()
}

// TODO: This test tests that we don't panic and are able to generate new
// PeerIDs for each peer we add. It seems as though we should be able to test
// this in a much more direct way.
// https://github.com/cometbft/cometbft/issues/9639
func TestDontExhaustMaxActiveIDs(t *testing.T) {
	config := cfg.TestConfig()
	const N = 1
	reactors, _ := makeAndConnectReactors(config, N)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	reactor := reactors[0]

	for i := 0; i < MaxActiveIDs+1; i++ {
		peer := mock.NewPeer(nil)
		reactor.Receive(p2p.Envelope{
			ChannelID: MempoolChannel,
			Src:       peer,
			Message:   &memproto.Message{}, // This uses the wrong message type on purpose to stop the peer as in an error state in the reactor.
		},
		)
		reactor.AddPeer(peer)
	}
}

// Test the experimental feature that limits the number of outgoing connections for gossiping
// transactions (only non-persistent peers).
// Note: in this test we know which gossip connections are active or not because of how the p2p
// functions are currently implemented, which affects the order in which peers are added to the
// mempool reactor.
func TestMempoolReactorMaxActiveOutboundConnections(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.ExperimentalMaxGossipConnectionsToNonPersistentPeers = 1
	reactors, _ := makeAndConnectReactors(config, 4)
	defer func() {
		for _, r := range reactors {
			if err := r.Stop(); err != nil {
				assert.NoError(t, err)
			}
		}
	}()
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{1})
		}
	}

	// Add a bunch transactions to the first reactor.
	txs := newUniqueTxs(100)
	callCheckTx(t, reactors[0].mempool, txs, UnknownPeerID)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInMempool(t, txs, reactors[1], 0)
	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	firstPeer := reactors[0].Switch.Peers().List()[0]
	reactors[0].Switch.StopPeerGracefully(firstPeer)

	// Now the third reactor should start receiving transactions from the first reactor; the fourth
	// reactor's mempool should still be empty.
	checkTxsInMempool(t, txs, reactors[2], 0)
	for _, r := range reactors[3:] {
		require.Zero(t, r.mempool.Size())
	}
}

// mempoolLogger is a TestingLogger which uses a different
// color for each validator ("validator" key must exist).
func mempoolLogger() log.Logger {
	return log.TestingLoggerWithColorFn(func(keyvals ...interface{}) term.FgBgColor {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] == "validator" {
				return term.FgBgColor{Fg: term.Color(uint8(keyvals[i+1].(int) + 1))}
			}
		}
		return term.FgBgColor{}
	})
}

// connect N mempool reactors through N switches
func makeAndConnectReactors(config *cfg.Config, n int) ([]*Reactor, []*p2p.Switch) {
	reactors := make([]*Reactor, n)
	logger := mempoolLogger()
	for i := 0; i < n; i++ {
		app := kvstore.NewInMemoryApplication()
		cc := proxy.NewLocalClientCreator(app)
		mempool, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		reactors[i] = NewReactor(config.Mempool, mempool) // so we dont start the consensus states
		reactors[i].SetLogger(logger.With("validator", i))
	}

	switches := p2p.MakeConnectedSwitches(config.P2P, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", reactors[i])
		return s

	}, p2p.Connect2Switches)
	return reactors, switches
}

func newUniqueTxs(n int) types.Txs {
	txs := make(types.Txs, n)
	for i := 0; i < n; i++ {
		txs[i] = kvstore.NewTxFromID(i)
	}
	return txs
}

func waitForTxsOnReactors(t *testing.T, txs types.Txs, reactors []*Reactor) {
	// wait for the txs in all mempools
	wg := new(sync.WaitGroup)
	for i, reactor := range reactors {
		wg.Add(1)
		go func(r *Reactor, reactorIndex int) {
			defer wg.Done()
			checkTxsInOrder(t, txs, r, reactorIndex)
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
	waitForNumTxsInMempool(len(txs), reactor.mempool)

	reapedTxs := reactor.mempool.ReapMaxTxs(len(txs))
	require.Equal(t, len(txs), len(reapedTxs))
	require.Equal(t, len(txs), reactor.mempool.Size())
}

// Wait until all txs are in the mempool and check that they are in the same
// order as given.
func checkTxsInOrder(t *testing.T, txs types.Txs, reactor *Reactor, reactorIndex int) {
	waitForNumTxsInMempool(len(txs), reactor.mempool)

	// Check that all transactions in the mempool are in the same order as txs.
	reapedTxs := reactor.mempool.ReapMaxTxs(len(txs))
	for i, tx := range txs {
		assert.Equalf(t, tx, reapedTxs[i],
			"txs at index %d on reactor %d don't match: %v vs %v", i, reactorIndex, tx, reapedTxs[i])
	}
}

// ensure no txs on reactor after some timeout
func ensureNoTxs(t *testing.T, reactor *Reactor, timeout time.Duration) {
	time.Sleep(timeout) // wait for the txs in all mempools
	assert.Zero(t, reactor.mempool.Size())
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
		tc := tc

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
<<<<<<< HEAD
=======

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

	require.Equal(t, int64(len(txs)), reactors[0].redundancyControl.firstTimeTxs.Load())
	for _, tx := range txs {
		_, err := reactors[0].TryAddTx(tx, secondNode)
		// The transaction is in cache, hence the Error
		require.Error(t, err)
	}
	require.Equal(t, int64(len(txs)), reactors[0].redundancyControl.duplicateTxs.Load())

	reactors[0].redundancyControl.triggerAdjustment(reactors[0])
	// This is done to give enough time for the route changes to take effect
	// If the test starts failing, revisit this value
	time.Sleep(100 * time.Millisecond)

	dupTx := reactors[0].redundancyControl.duplicateTxs.Load()
	firstTimeTx := reactors[0].redundancyControl.firstTimeTxs.Load()

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
	// If the test starts failing, revisit this value
	time.Sleep(100 * time.Millisecond)

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

	reactors[0].redundancyControl.firstTimeTxs.Store(10)
	reactors[0].redundancyControl.duplicateTxs.Store(0)
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Equal(t, redundancy, float64(0))

	reactors[0].redundancyControl.duplicateTxs.Store(1000)
	reactors[0].redundancyControl.firstTimeTxs.Store(10)
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Greater(t, redundancy, config.Mempool.DOGTargetRedundancy)

	reactors[0].redundancyControl.duplicateTxs.Store(1000)
	reactors[0].redundancyControl.firstTimeTxs.Store(0)
	redundancy = reactors[0].redundancyControl.currentRedundancy()
	require.Equal(t, redundancy, reactors[0].redundancyControl.upperBound)
}

func BenchmarkCurrentRedundancy(b *testing.B) {
	config := &cfg.MempoolConfig{
		DOGTargetRedundancy: 1.0,
		DOGAdjustInterval:   1 * time.Second,
	}
	rc := newRedundancyControl(config)

	b.Run("CurrentRedundancy", func(b *testing.B) {
		// Pre-fill some data
		rc.incFirstTimeTxs()
		rc.incDuplicateTxs()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rc.currentRedundancy()
		}
	})

	b.Run("IncFirstTimeTxs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rc.incFirstTimeTxs()
		}
	})

	b.Run("IncDuplicateTxs", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rc.incDuplicateTxs()
		}
	})
}
>>>>>>> be894a0e (perf(mempool): use atomic instead of mutex to improve perf (#4701))
