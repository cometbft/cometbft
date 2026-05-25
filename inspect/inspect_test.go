package inspect_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/inspect"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/pubsub/query"
	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	indexermocks "github.com/cometbft/cometbft/state/indexer/mocks"
	statemocks "github.com/cometbft/cometbft/state/mocks"
	txindexmocks "github.com/cometbft/cometbft/state/txindex/mocks"
	"github.com/cometbft/cometbft/types"
	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func startInspector(t *testing.T, d *inspect.Inspector, listenAddr string) (stop func()) {
	t.Helper()
	parts := strings.SplitN(listenAddr, "://", 2)
	if len(parts) != 2 {
		t.Fatalf("malformed listen address: %s", listenAddr)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- d.Run(ctx) }()
	require.Eventually(t, func() bool {
		select {
		case err := <-errCh:
			cancel()
			require.NoError(t, err, "inspector exited before becoming ready")
			require.Fail(t, "inspector exited cleanly before becoming ready")
			return false
		default:
		}
		conn, err := net.Dial(parts[0], parts[1])
		if err == nil {
			conn.Close()
			return true
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)
	var once sync.Once
	stop = func() {
		once.Do(func() {
			cancel()
			require.NoError(t, <-errCh)
		})
	}
	t.Cleanup(stop)
	return stop
}

func TestInspectConstructor(t *testing.T) {
	cfg := test.ResetTestRoot("test")
	t.Cleanup(leaktest.Check(t))
	defer func() { _ = os.RemoveAll(cfg.RootDir) }()
	t.Run("from config", func(t *testing.T) {
		d, err := inspect.NewFromConfig(cfg)
		require.NoError(t, err)
		require.NotNil(t, d)
	})
}

func TestInspectRun(t *testing.T) {
	cfg := test.ResetTestRoot("test")
	t.Cleanup(leaktest.Check(t))
	defer func() { _ = os.RemoveAll(cfg.RootDir) }()
	t.Run("from config", func(t *testing.T) {
		d, err := inspect.NewFromConfig(cfg)
		require.NoError(t, err)
		stop := startInspector(t, d, cfg.RPC.ListenAddress)
		stop()
	})
}

func TestBlock(t *testing.T) {
	testHeight := int64(1)
	testBlock := new(types.Block)
	testBlock.Height = testHeight
	testBlock.LastCommitHash = []byte("test hash")
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)

	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Height").Return(testHeight)
	blockStoreMock.On("Base").Return(int64(0))
	blockStoreMock.On("LoadBlockMeta", testHeight).Return(&types.BlockMeta{})
	blockStoreMock.On("LoadBlock", testHeight).Return(testBlock)
	blockStoreMock.On("Close").Return(nil)

	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}

	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	resultBlock, err := cli.Block(context.Background(), &testHeight)
	require.NoError(t, err)
	require.Equal(t, testBlock.Height, resultBlock.Block.Height)
	require.Equal(t, testBlock.LastCommitHash, resultBlock.Block.LastCommitHash)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestTxSearch(t *testing.T) {
	testHash := []byte("test")
	testTx := []byte("tx")
	testQuery := fmt.Sprintf("tx.hash='%s'", string(testHash))
	testTxResult := &abcitypes.TxResult{
		Height: 1,
		Index:  100,
		Tx:     testTx,
	}

	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	txIndexerMock.On("Search", mock.Anything,
		mock.MatchedBy(func(q *query.Query) bool {
			return testQuery == strings.ReplaceAll(q.String(), " ", "")
		})).
		Return([]*abcitypes.TxResult{testTxResult}, nil)

	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)

	page := 1
	resultTxSearch, err := cli.TxSearch(context.Background(), testQuery, false, &page, &page, "")
	require.NoError(t, err)
	require.Len(t, resultTxSearch.Txs, 1)
	require.Equal(t, types.Tx(testTx), resultTxSearch.Txs[0].Tx)
	stop()

	txIndexerMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
	blockStoreMock.AssertExpectations(t)
}

func TestTx(t *testing.T) {
	testHash := []byte("test")
	testTx := []byte("tx")

	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blkIdxMock := &indexermocks.BlockIndexer{}
	txIndexerMock := &txindexmocks.TxIndexer{}
	txIndexerMock.On("Get", testHash).Return(&abcitypes.TxResult{
		Tx: testTx,
	}, nil)

	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)

	res, err := cli.Tx(context.Background(), testHash, false)
	require.NoError(t, err)
	require.Equal(t, types.Tx(testTx), res.Tx)
	stop()

	txIndexerMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
	blockStoreMock.AssertExpectations(t)
}

func TestConsensusParams(t *testing.T) {
	testHeight := int64(1)
	testMaxGas := int64(55)
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("Height").Return(testHeight)
	blockStoreMock.On("Base").Return(int64(0))
	stateStoreMock.On("LoadConsensusParams", testHeight).Return(types.ConsensusParams{
		Block: types.BlockParams{
			MaxGas: testMaxGas,
		},
	}, nil)
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	params, err := cli.ConsensusParams(context.Background(), &testHeight)
	require.NoError(t, err)
	require.Equal(t, params.ConsensusParams.Block.MaxGas, testMaxGas)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestBlockResults(t *testing.T) {
	testHeight := int64(1)
	testGasUsed := int64(100)
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	stateStoreMock.On("LoadFinalizeBlockResponse", testHeight).Return(&abcitypes.ResponseFinalizeBlock{
		TxResults: []*abcitypes.ExecTxResult{
			{
				GasUsed: testGasUsed,
			},
		},
	}, nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("Base").Return(int64(0))
	blockStoreMock.On("Height").Return(testHeight)
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	res, err := cli.BlockResults(context.Background(), &testHeight)
	require.NoError(t, err)
	require.Equal(t, res.TxsResults[0].GasUsed, testGasUsed)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestCommit(t *testing.T) {
	testHeight := int64(1)
	testRound := int32(101)
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("Base").Return(int64(0))
	blockStoreMock.On("Height").Return(testHeight)
	blockStoreMock.On("LoadBlockMeta", testHeight).Return(&types.BlockMeta{}, nil)
	blockStoreMock.On("LoadSeenCommit", testHeight).Return(&types.Commit{
		Height: testHeight,
		Round:  testRound,
	}, nil)
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	res, err := cli.Commit(context.Background(), &testHeight)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, res.Commit.Round, testRound)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestBlockByHash(t *testing.T) {
	testHeight := int64(1)
	testHash := []byte("test hash")
	testBlock := new(types.Block)
	testBlock.Height = testHeight
	testBlock.LastCommitHash = testHash
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("LoadBlockMeta", testHeight).Return(&types.BlockMeta{
		BlockID: types.BlockID{
			Hash: testHash,
		},
		Header: types.Header{
			Height: testHeight,
		},
	}, nil)
	blockStoreMock.On("LoadBlockByHash", testHash).Return(testBlock, nil)
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	res, err := cli.BlockByHash(context.Background(), testHash)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, []byte(res.BlockID.Hash), testHash)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestBlockchain(t *testing.T) {
	testHeight := int64(1)
	testBlock := new(types.Block)
	testBlockHash := []byte("test hash")
	testBlock.Height = testHeight
	testBlock.LastCommitHash = testBlockHash
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)

	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("Height").Return(testHeight)
	blockStoreMock.On("Base").Return(int64(0))
	blockStoreMock.On("LoadBlockMeta", testHeight).Return(&types.BlockMeta{
		BlockID: types.BlockID{
			Hash: testBlockHash,
		},
	})
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)
	res, err := cli.BlockchainInfo(context.Background(), 0, 100)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, testBlockHash, []byte(res.BlockMetas[0].BlockID.Hash))
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestValidators(t *testing.T) {
	testHeight := int64(1)
	testVotingPower := int64(100)
	testValidators := types.ValidatorSet{
		Validators: []*types.Validator{
			{
				VotingPower: testVotingPower,
			},
		},
	}
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)
	stateStoreMock.On("LoadValidators", testHeight).Return(&testValidators, nil)

	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)
	blockStoreMock.On("Height").Return(testHeight)
	blockStoreMock.On("Base").Return(int64(0))
	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)

	testPage := 1
	testPerPage := 100
	res, err := cli.Validators(context.Background(), &testHeight, &testPage, &testPerPage)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, testVotingPower, res.Validators[0].VotingPower)
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}

func TestBlockSearch(t *testing.T) {
	testHeight := int64(1)
	testBlockHash := []byte("test hash")
	testQuery := "block.height = 1"
	stateStoreMock := &statemocks.Store{}
	stateStoreMock.On("Close").Return(nil)

	blockStoreMock := &statemocks.BlockStore{}
	blockStoreMock.On("Close").Return(nil)

	txIndexerMock := &txindexmocks.TxIndexer{}
	blkIdxMock := &indexermocks.BlockIndexer{}
	blockStoreMock.On("LoadBlock", testHeight).Return(&types.Block{
		Header: types.Header{
			Height: testHeight,
		},
	}, nil)
	blockStoreMock.On("LoadBlockMeta", testHeight).Return(&types.BlockMeta{
		BlockID: types.BlockID{
			Hash: testBlockHash,
		},
	})
	blkIdxMock.On("Search", mock.Anything,
		mock.MatchedBy(func(q *query.Query) bool { return testQuery == q.String() })).
		Return([]int64{testHeight}, nil)
	rpcConfig := config.TestRPCConfig()
	d := inspect.New(rpcConfig, blockStoreMock, stateStoreMock, txIndexerMock, blkIdxMock)

	stop := startInspector(t, d, rpcConfig.ListenAddress)
	cli, err := httpclient.New(rpcConfig.ListenAddress, "/websocket")
	require.NoError(t, err)

	testPage := 1
	testPerPage := 100
	testOrderBy := "desc"
	res, err := cli.BlockSearch(context.Background(), testQuery, &testPage, &testPerPage, testOrderBy)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, testBlockHash, []byte(res.Blocks[0].BlockID.Hash))
	stop()

	blockStoreMock.AssertExpectations(t)
	stateStoreMock.AssertExpectations(t)
}
