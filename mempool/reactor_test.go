package mempool

import (
	"encoding/hex"
	"errors"
	"runtime/debug"
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

	txs := checkTxs(t, reactors[0].mempool, numTxs)
	waitForReactors(t, txs, reactors, checkTxsInMempoolInOrder)
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

	const peerID0 = 0
	const peerID1 = 1
	// the second peer sends all the transactions to the first peer
	for _, tx := range txs {
		reactors[0].addSender(tx.Key(), peerID1)
		_, err := reactors[peerID0].mempool.CheckTx(tx)
		require.NoError(t, err)
	}

	// the second peer should not receive any transaction
	ensureNoTxs(t, reactors[peerID1], 100*time.Millisecond)
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
	waitForReactors(t, []types.Tx{tx1}, reactors, checkTxsInMempoolInOrder)

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
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), 1))

	reactor.addSender(types.Tx(tx1).Key(), 1)
	reactor.addSender(types.Tx(tx1).Key(), 2)
	reactor.addSender(types.Tx(tx2).Key(), 1)
	require.True(t, reactor.isSender(types.Tx(tx1).Key(), 1))
	require.True(t, reactor.isSender(types.Tx(tx1).Key(), 2))
	require.True(t, reactor.isSender(types.Tx(tx2).Key(), 1))

	reactor.removeSenders(types.Tx(tx1).Key())
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), 1))
	require.False(t, reactor.isSender(types.Tx(tx1).Key(), 2))
	require.True(t, reactor.isSender(types.Tx(tx2).Key(), 1))
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

	numTxs := config.Mempool.Size
	txs := newUniqueTxs(numTxs)

	// Initially, there are no transactions (and no senders).
	for _, r := range reactors {
		require.Zero(t, r.mempool.Size())
		require.Zero(t, len(r.txSenders))
	}

	// Add transactions to the first reactor.
	callCheckTx(t, reactors[0].mempool, txs)

	// Wait for all txs to be in the mempool of each reactor.
	waitForReactors(t, txs, reactors, checkTxsInMempoolInOrder)
	for i, r := range reactors {
		checkTxsInMempoolAndSenders(t, r, txs, i)
	}

	// Split the transactions in three groups of different sizes.
	splitIndex := numTxs / 6
	validTxs := txs[:splitIndex]                 // 1/6 will be used to update the mempool, as valid txs
	invalidTxs := txs[splitIndex : 3*splitIndex] // 2/6 will be used to update the mempool, as invalid txs
	ignoredTxs := txs[3*splitIndex:]             // 3/6 will remain in the mempool

	// Update the mempools with a list of valid and invalid transactions.
	for i, r := range reactors {
		updateMempool(t, r.mempool, validTxs, invalidTxs)

		// Txs included in a block should have been removed from the mempool and
		// have no senders.
		for _, tx := range append(validTxs, invalidTxs...) {
			require.False(t, r.mempool.Contains(tx.Key()))
			_, hasSenders := r.txSenders[tx.Key()]
			require.False(t, hasSenders)
		}

		// Ignored txs should still be in the mempool.
		checkTxsInMempoolAndSenders(t, r, ignoredTxs, i)
	}

	// The first reactor should not receive transactions from other peers.
	require.Zero(t, len(reactors[0].txSenders))
}

// Test that, even if the sender sleeps because the receiver is late, the
// receiver will eventually get the transactions.
func TestReactorPeerLagging(t *testing.T) {
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

	numTxs := config.Mempool.Size
	txs := newUniqueTxs(numTxs)

	// Update the first reactor to set its height ahead of all its peers.
	err := reactors[0].mempool.Update(3, types.Txs{}, make([]*abci.ExecTxResult, 0), nil, nil)
	require.NoError(t, err)

	// No reactor has transactions (or senders).
	for _, r := range reactors {
		require.Zero(t, r.mempool.Size())
		require.Zero(t, len(r.txSenders))
	}

	// Add transactions to the first reactor.
	callCheckTx(t, reactors[0].mempool, txs)
	// Ensure the transactions were added in the right order.
	waitForReactors(t, txs, reactors[:0], checkTxsInMempoolInOrder)

	// Because the peers are lagging, the first reactor should keep sleeping and
	// not broadcast the transactions even if it has had plenty of time.
	time.Sleep(PeerCatchupSleepIntervalMS * time.Millisecond * 100)
	waitForReactors(t, txs, reactors[1:], checkNoTxsInMempool)

	// peerState catches up.
	for _, r := range reactors {
		for _, peer := range r.Switch.Peers().List() {
			peer.Set(types.PeerStateKey, peerState{3})
		}
	}

	// Now the txs should be propagated and received in the right order.
	// This test may be flaky.
	waitForReactors(t, txs, reactors, checkTxsInMempoolInOrder)
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
		assert.True(t, r.mempool.Contains(tx.Key()))
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

		reactors[i] = NewReactor(config.Mempool, mempool) // so we don't start the consensus states
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
		t.Fatal("Timed out waiting for txs ", string(debug.Stack()))
	case <-done:
	}
}

func checkNoTxsInMempool(t *testing.T, _ types.Txs, reactor *Reactor, _ int) {
	require.Equal(t, 0, reactor.mempool.Size())
}

// Wait until the mempool has a certain number of transactions.
func waitForNumTxsInMempool(numTxs int, mempool Mempool) {
	for mempool.Size() < numTxs {
		time.Sleep(time.Millisecond * 100)
	}
}

// Wait until all txs are in the mempool and check that they are in the same
// order as given.
func checkTxsInMempoolInOrder(t *testing.T, txs types.Txs, reactor *Reactor, reactorIndex int) {
	waitForNumTxsInMempool(len(txs), reactor.mempool)

	// Check that all transactions in the mempool are in the same order as txs.
	reapedTxs := reactor.mempool.ReapMaxTxs(len(txs))
	require.Equal(t, len(txs), len(reapedTxs))
	require.Equal(t, len(txs), reactor.mempool.Size())
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
