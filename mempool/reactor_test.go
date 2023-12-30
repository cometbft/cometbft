package mempool

import (
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/go-kit/log/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	abci "github.com/cometbft/cometbft/abci/types"
	memproto "github.com/cometbft/cometbft/api/cometbft/mempool/v1"
	cfg "github.com/cometbft/cometbft/config"
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

	txs := checkTxs(t, reactors[0].mempool, numTxs)
	waitForReactors(t, txs, reactors, checkTxsInOrder)
}

// regression test for https://github.com/tendermint/tendermint/issues/5408
func TestReactorConcurrency(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.Size = 5000
	config.Mempool.CacheSize = 5000
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
		txs := checkTxs(t, reactors[0].mempool, numTxs)
		go func() {
			defer wg.Done()

			reactors[0].mempool.Lock()
			defer reactors[0].mempool.Unlock()

			err := reactors[0].mempool.Update(1, txs, abciResponses(len(txs), abci.CodeTypeOK), nil, nil)
			assert.NoError(t, err)
		}()

		// 1. submit a bunch of txs
		// 2. update none
		_ = checkTxs(t, reactors[1].mempool, numTxs)
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

	// create random transactions
	txs := NewRandomTxs(numTxs, 20)

	// the second peer sends all the transactions to the first peer
	secondNodeID := reactors[1].Switch.NodeInfo().ID()
	for _, tx := range txs {
		reactors[0].addSender(tx.Key(), secondNodeID)
		_, err := reactors[0].mempool.CheckTx(tx)
		require.NoError(t, err)
	}

	// the second peer should not receive any transaction
	ensureNoTxs(t, reactors[1], 100*time.Millisecond)
}

func TestReactor_MaxTxBytes(t *testing.T) {
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
	reqRes, err := reactors[0].mempool.CheckTx(tx1)
	require.NoError(t, err)
	require.False(t, reqRes.Response.GetCheckTx().IsErr())
	waitForReactors(t, []types.Tx{tx1}, reactors, checkTxsInOrder)

	reactors[0].mempool.Flush()
	reactors[1].mempool.Flush()

	// Broadcast a tx, which is beyond the max size
	// => ensure it's not sent
	tx2 := kvstore.NewRandomTx(config.Mempool.MaxTxBytes + 1)
	reqRes, err = reactors[0].mempool.CheckTx(tx2)
	require.Error(t, err)
	require.Nil(t, reqRes)
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

func TestReactorTxSendersLocal(t *testing.T) {
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

	tx1 := kvstore.NewTxFromID(1)
	tx2 := kvstore.NewTxFromID(2)
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), "peer1"))

	reactor.addSender(types.Tx(tx1).Key(), "peer1")
	reactor.addSender(types.Tx(tx1).Key(), "peer2")
	reactor.addSender(types.Tx(tx2).Key(), "peer1")
	require.True(t, reactor.isSender(types.Tx(tx1).Key(), "peer1"))
	require.True(t, reactor.isSender(types.Tx(tx1).Key(), "peer2"))
	require.True(t, reactor.isSender(types.Tx(tx2).Key(), "peer1"))

	reactor.removeSenders(types.Tx(tx1).Key())
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), "peer1"))
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), "peer2"))
	require.True(t, reactor.isSender(types.Tx(tx2).Key(), "peer1"))
}

// Test that:
// - If a transaction came from a peer AND if the transaction is added to the
// mempool, it must have a non-empty list of senders in the reactor.
// - If a transaction is removed from the mempool, it must also be removed from
// the list of senders in the reactor.
func TestReactorTxSendersMultiNode(t *testing.T) {
	config := cfg.TestConfig()
	config.Mempool.Size = 1000
	config.Mempool.CacheSize = 1000
	const N = 3
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
	firstReactor := reactors[0]

	numTxs := config.Mempool.Size
	txs := newUniqueTxs(numTxs)

	// Initially, there are no transactions (and no senders).
	for _, r := range reactors {
		require.Zero(t, len(r.txSenders))
	}

	// Add transactions to the first reactor.
	callCheckTx(t, firstReactor.mempool, txs)

	// Wait for all txs to be in the mempool of each reactor.
	waitForReactors(t, txs, reactors, checkTxsInMempool)
	for i, r := range reactors {
		checkTxsInMempoolAndSenders(t, r, txs, i)
	}

	// Split the transactions in three groups of different sizes.
	splitIndex := numTxs / 6
	validTxs := txs[:splitIndex]                 // will be used to update the mempool, as valid txs
	invalidTxs := txs[splitIndex : 3*splitIndex] // will be used to update the mempool, as invalid txs
	ignoredTxs := txs[3*splitIndex:]             // will remain in the mempool

	// Update the mempools with a list of valid and invalid transactions.
	for i, r := range reactors {
		updateMempool(t, r.mempool, validTxs, invalidTxs)

		// Txs included in a block should have been removed from the mempool and
		// have no senders.
		for _, tx := range append(validTxs, invalidTxs...) {
			require.False(t, r.mempool.InMempool(tx.Key()))
			_, hasSenders := r.txSenders[tx.Key()]
			require.False(t, hasSenders)
		}

		// Ignored txs should still be in the mempool.
		checkTxsInMempoolAndSenders(t, r, ignoredTxs, i)
	}

	// The first reactor should not receive transactions from other peers.
	require.Zero(t, len(firstReactor.txSenders))
}

func TestMempoolFIFOWithParallelCheckTx(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("FIFO is not supposed to be guaranteed and this this is just used to evidence one of the cases where it does not happen. Hence we skip this test during CI.")
	}

	config := cfg.TestConfig()
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

	// Deliver the same sequence of transactions from multiple sources, in parallel.
	txs := newUniqueTxs(200)
	mp := reactors[0].mempool
	for i := 0; i < 3; i++ {
		go func() {
			for _, tx := range txs {
				mp.CheckTx(tx) //nolint:errcheck
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
	callCheckTx(t, reactors[0].mempool, txs)

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

// Test the experimental feature that limits the number of outgoing connections for gossiping
// transactions (only non-persistent peers).
// Given the disconnections, no transaction should be received in duplicate.
// Note: in this test we know which gossip connections are active or not because of how the p2p
// functions are currently implemented, which affects the order in which peers are added to the
// mempool reactor.
func TestMempoolReactorMaxActiveOutboundConnectionsNoDuplicate(t *testing.T) {
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

	// Disconnect the second reactor from the third reactor.
	pCon1_2 := reactors[1].Switch.Peers().List()[1]
	reactors[1].Switch.StopPeerGracefully(pCon1_2)

	// Add a bunch transactions to the first reactor.
	txs := newUniqueTxs(100)
	callCheckTx(t, reactors[0].mempool, txs)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInOrder(t, txs, reactors[1], 0)
	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	pCon0_1 := reactors[0].Switch.Peers().List()[0]
	reactors[0].Switch.StopPeerGracefully(pCon0_1)

	// Now the third reactor should start receiving transactions from the first reactor and
	// the fourth reactor from the second
	checkTxsInOrder(t, txs, reactors[2], 0)
	checkTxsInOrder(t, txs, reactors[3], 0)
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
	reactors, _ := makeAndConnectReactorsStar(config, 0, 4)
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
	txs := newUniqueTxs(5)
	callCheckTx(t, reactors[0].mempool, txs)

	// Wait for all txs to be in the mempool of the second reactor; the other reactors should not
	// receive any tx. (The second reactor only sends transactions to the first reactor.)
	checkTxsInOrder(t, txs, reactors[0], 0)
	checkTxsInOrder(t, txs, reactors[1], 0)

	for _, r := range reactors[2:] {
		require.Zero(t, r.mempool.Size())
	}

	// Disconnect the second reactor from the first reactor.
	firstPeer := reactors[0].Switch.Peers().List()[0]
	reactors[0].Switch.StopPeerGracefully(firstPeer)

	// Now the third reactor should start receiving transactions from the first reactor; the fourth
	// reactor's mempool should still be empty.
	checkTxsInOrder(t, txs, reactors[0], 0)
	checkTxsInOrder(t, txs, reactors[1], 0)
	checkTxsInOrder(t, txs, reactors[2], 0)
	for _, r := range reactors[3:] {
		require.Zero(t, r.mempool.Size())
	}
}

// Check that the mempool has exactly the given list of txs and, if it's not the
// first reactor (reactorIndex == 0), then each tx has a non-empty list of senders.
func checkTxsInMempoolAndSenders(t *testing.T, r *Reactor, txs types.Txs, reactorIndex int) {
	r.txSendersMtx.Lock()
	defer r.txSendersMtx.Unlock()

	require.Equal(t, len(txs), r.mempool.Size())
	if reactorIndex == 0 {
		require.Zero(t, len(r.txSenders))
	} else {
		require.Equal(t, len(txs), len(r.txSenders))
	}

	// Each transaction is in the mempool and, if it's not the first reactor, it
	// has a non-empty list of senders.
	for _, tx := range txs {
		assert.True(t, r.mempool.InMempool(tx.Key()))
		senders, hasSenders := r.txSenders[tx.Key()]
		if reactorIndex == 0 {
			require.False(t, hasSenders)
		} else {
			require.True(t, hasSenders && len(senders) > 0)
		}
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

		reactors[i] = NewReactor(config.Mempool, mempool, false) // so we dont start the consensus states
		reactors[i].SetLogger(logger.With("validator", i))
	}

	switches := p2p.MakeConnectedSwitches(config.P2P, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", reactors[i])
		return s
	}, p2p.Connect2Switches)
	return reactors, switches
}

// connect N mempool reactors through N switches as a star centered in c
func makeAndConnectReactorsStar(config *cfg.Config, c, n int) ([]*Reactor, []*p2p.Switch) {
	reactors := make([]*Reactor, n)
	logger := mempoolLogger()
	for i := 0; i < n; i++ {
		app := kvstore.NewInMemoryApplication()
		cc := proxy.NewLocalClientCreator(app)
		mempool, cleanup := newMempoolWithApp(cc)
		defer cleanup()

		reactors[i] = NewReactor(config.Mempool, mempool, false) // so we dont start the consensus states
		reactors[i].SetLogger(logger.With("validator", i))
	}

	switches := p2p.MakeConnectedSwitches(config.P2P, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", reactors[i])
		return s
	}, p2p.ConnectStarSwitches(c))
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

func updateMempool(t *testing.T, mp Mempool, validTxs types.Txs, invalidTxs types.Txs) {
	allTxs := append(validTxs, invalidTxs...)

	validTxResponses := abciResponses(len(validTxs), abci.CodeTypeOK)
	invalidTxResponses := abciResponses(len(invalidTxs), 1)
	allResponses := append(validTxResponses, invalidTxResponses...)

	mp.Lock()
	err := mp.Update(1, allTxs, allResponses, nil, nil)
	mp.Unlock()

	require.NoError(t, err)
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
