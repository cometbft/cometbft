package mempool

import (
	"context"
	"sync"
	"testing"
	"time"

	abcimock "github.com/cometbft/cometbft/abci/client/mocks"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAppReactor(t *testing.T) {
	const (
		timeout  = 5 * time.Second
		interval = 200 * time.Millisecond
	)

	eventually := func(fn func() bool) {
		require.Eventually(t, fn, timeout, interval)
	}

	// ARRANGE
	// Given 3 nodes
	var (
		nodeA = newAppReactorNode(t, "A")
		nodeB = newAppReactorNode(t, "B")
		nodeC = newAppReactorNode(t, "C")
		nodes = []*appReactorNode{nodeA, nodeB, nodeC}
	)

	// With switches connected & mempool initialized
	onStart := func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("MEMPOOL", nodes[i].reactor)
		return s
	}

	switches := p2p.MakeConnectedSwitches(config.TestConfig().P2P, len(nodes), onStart, p2p.Connect2Switches)

	for i, node := range nodes {
		node.sw = switches[i]
		node.reactor.EnableInOutTxs()
	}

	defer func() {
		for _, node := range nodes {
			if err := node.sw.Stop(); err != nil {
				require.NoError(t, err)
			}
		}
	}()

	// ACT #1
	// Insert several txs into A
	txs1 := []types.Tx{
		types.Tx("from_a_to_b:1"),
		types.Tx("from_a_to_b:2"),
		types.Tx("from_a_to_b:3"),
		types.Tx("from_a_to_b:4"),
		types.Tx("from_a_to_b:5"),
	}
	for _, tx := range txs1 {
		err := nodeA.mempool.InsertTx(tx)
		require.NoError(t, err, "failed to insert tx %q into node A", tx)
	}

	// ASSERT #1
	// Wait for txs to arrive at B
	eventually(func() bool {
		received := nodeB.getReceivedTxs()
		return txsContain(received, txs1)
	})

	// ACT #2
	// Insert several txs into B
	txs2 := []types.Tx{
		types.Tx("from_b_to_a:1"),
		types.Tx("from_b_to_a:2"),
		types.Tx("from_b_to_a:3"),
		types.Tx("from_b_to_a:4"),
		types.Tx("from_b_to_a:5"),
	}
	for _, tx := range txs2 {
		err := nodeB.mempool.InsertTx(tx)
		require.NoError(t, err, "failed to insert tx %q into node B", tx)
	}

	// ASSERT #2
	// Wait for txs to arrive at A
	eventually(func() bool {
		received := nodeA.getReceivedTxs()
		return txsContain(received, txs2)
	})

	// ASSERT #3
	// Ensure all nodes (including C) have all txs
	allTxs := append(txs1, txs2...)
	eventually(func() bool {
		receivedA := nodeA.getReceivedTxs()
		receivedB := nodeB.getReceivedTxs()
		receivedC := nodeC.getReceivedTxs()
		return txsContain(receivedA, allTxs) &&
			txsContain(receivedB, allTxs) &&
			txsContain(receivedC, allTxs)
	})

	// Also make sure that all txs are unique in each node
	require.False(t, hasDuplicates(nodeA.getReceivedTxs()))
	require.False(t, hasDuplicates(nodeB.getReceivedTxs()))
	require.False(t, hasDuplicates(nodeC.getReceivedTxs()))
}

func TestChunkTxs(t *testing.T) {
	makeTx := func(size int) types.Tx {
		return types.Tx(rand.Bytes(size))
	}

	toTxs := func(sizes []int) types.Txs {
		txs := make([]types.Tx, 0, len(sizes))
		for _, size := range sizes {
			txs = append(txs, makeTx(size))
		}
		return txs
	}

	for _, tt := range []struct {
		name   string
		input  []int
		size   int
		output [][]int
	}{
		{
			name:   "single tx smaller than size",
			input:  []int{100},
			size:   200,
			output: [][]int{{100}},
		},
		{
			name:   "single tx bigger than size",
			input:  []int{100},
			size:   50,
			output: [][]int{{100}},
		},
		{
			name:   "basic",
			input:  []int{100, 100, 100},
			size:   200,
			output: [][]int{{100, 100}, {100}},
		},
		{
			name:   "txs equal size",
			input:  []int{100, 100, 100},
			size:   100,
			output: [][]int{{100}, {100}, {100}},
		},
		{
			name:   "edge-case",
			input:  []int{101, 20, 30, 50, 2, 102, 3},
			size:   100,
			output: [][]int{{101}, {20, 30, 50}, {2}, {102}, {3}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			input := toTxs(tt.input)

			expected := make([]types.Txs, 0, len(tt.output))
			for _, chunk := range tt.output {
				expected = append(expected, toTxs(chunk))
			}

			// ACT
			actual := chunkTxs(input, tt.size)

			// ASSERT
			require.Equal(t, len(expected), len(actual), "output length mismatch")

			for i, chunk := range actual {
				require.Equal(t, len(expected[i]), len(chunk), "chunk length mismatch (#%d)", i)
			}
		})
	}
}

type appReactorNode struct {
	t    *testing.T
	name string

	app     *abcimock.Client
	mempool *AppMempool
	reactor *AppReactor
	sw      *p2p.Switch

	mempoolTxs  types.Txs
	receivedTxs types.Txs
	mu          sync.Mutex

	logger log.Logger
}

func newAppReactorNode(t *testing.T, name string) *appReactorNode {
	config := config.TestConfig()
	logger := log.TestingLogger().With("name", name)
	app := abcimock.NewClient(t)

	mempool := NewAppMempool(
		config.Mempool,
		app,
		WithAMLogger(logger.With("module", "mempool")),
	)

	reactor := NewAppReactor(config.Mempool, mempool, true)
	reactor.SetLogger(logger.With("module", "reactor"))

	ts := &appReactorNode{
		t:       t,
		name:    name,
		app:     app,
		mempool: mempool,
		reactor: reactor,
		logger:  logger,
	}

	ts.setupAppMock()

	return ts
}

func (ts *appReactorNode) insertTx(tx types.Tx) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.logger.Info("inserting tx", "tx", string(tx))

	ts.mempoolTxs = append(ts.mempoolTxs, tx)
	ts.receivedTxs = append(ts.receivedTxs, tx)
}

func (ts *appReactorNode) reapTxs() types.Txs {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.logger.Info("reaping txs")

	out := make(types.Txs, 0, len(ts.mempoolTxs))
	out = append(out, ts.mempoolTxs...)

	ts.mempoolTxs = ts.mempoolTxs[:0]

	return out
}

func (ts *appReactorNode) getReceivedTxs() types.Txs {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	out := make(types.Txs, 0, len(ts.receivedTxs))
	out = append(out, ts.receivedTxs...)

	return out
}

func (ts *appReactorNode) setupAppMock() {
	mockGrpc := func(method string, fn any) *mock.Call {
		return ts.app.On(method, mock.Anything, mock.Anything).Return(fn).Maybe()
	}

	mockGrpc("InsertTx", func(_ context.Context, req *abci.RequestInsertTx) (*abci.ResponseInsertTx, error) {
		ts.insertTx(req.Tx)
		return &abci.ResponseInsertTx{
			Code: abci.CodeTypeOK,
		}, nil
	})

	mockGrpc("ReapTxs", func(_ context.Context, req *abci.RequestReapTxs) (*abci.ResponseReapTxs, error) {
		out := ts.reapTxs()

		return &abci.ResponseReapTxs{Txs: out.ToSliceOfBytes()}, nil
	})
}

func txsContain(set, subset types.Txs) bool {
	cache := make(map[string]struct{})

	for _, tx := range set {
		cache[tx.String()] = struct{}{}
	}

	for _, tx := range subset {
		if _, ok := cache[tx.String()]; !ok {
			return false
		}
	}

	return true
}

func hasDuplicates(txs types.Txs) bool {
	cache := make(map[string]struct{})

	for _, tx := range txs {
		if _, ok := cache[tx.String()]; ok {
			return true
		}

		cache[tx.String()] = struct{}{}
	}

	return false
}
