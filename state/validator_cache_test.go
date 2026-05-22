package state_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	abcimocks "github.com/cometbft/cometbft/abci/types/mocks"
	"github.com/cometbft/cometbft/libs/log"
	mpmocks "github.com/cometbft/cometbft/mempool/mocks"
	"github.com/cometbft/cometbft/proxy"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/state/mocks"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

func oneValCommit(height int64) *types.Commit {
	return &types.Commit{
		Height:     height,
		Signatures: []types.CommitSig{types.NewCommitSigAbsent()},
	}
}

func TestValidatorCacheHitWithinBlockCycle(t *testing.T) {
	state, stateDB, _ := makeState(1, 3) // 1 validator, LastBlockHeight=2
	realStore := sm.NewStore(stateDB, sm.StoreOptions{})
	valSet, err := realStore.LoadValidators(state.LastBlockHeight)
	require.NoError(t, err)

	storeMock := mocks.NewStore(t)
	storeMock.On("LoadValidators", state.LastBlockHeight).Return(valSet, nil)

	app := abcimocks.NewApplication(t)
	app.On("ProcessProposal", mock.Anything, mock.Anything).
		Return(&abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil)
	app.On("ExtendVote", mock.Anything, mock.Anything).
		Return(&abci.ResponseExtendVote{}, nil)

	proxyApp := proxy.NewAppConns(proxy.NewLocalClientCreator(app), proxy.NopMetrics())
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { _ = proxyApp.Stop() })

	blockExec := sm.NewBlockExecutor(
		storeMock, log.NewNopLogger(), proxyApp.Consensus(),
		new(mpmocks.Mempool), sm.EmptyEvidencePool{},
		store.NewBlockStore(dbm.NewMemDB()),
	)

	block, err := makeBlock(state, state.LastBlockHeight+1, oneValCommit(state.LastBlockHeight))
	require.NoError(t, err)

	// ProcessProposal: first operation in the block cycle (cache miss → 1 DB load).
	_, err = blockExec.ProcessProposal(block, state)
	require.NoError(t, err)

	// ExtendVote: second operation in the same cycle (cache hit → no additional DB load).
	vote := &types.Vote{
		Height:  block.Height,
		BlockID: types.BlockID{Hash: block.Hash()},
	}
	_, err = blockExec.ExtendVote(context.Background(), vote, block, state)
	require.NoError(t, err)

	storeMock.AssertNumberOfCalls(t, "LoadValidators", 1)
}

func TestValidatorCacheInvalidatesOnNewHeight(t *testing.T) {
	state, stateDB, _ := makeState(1, 3) // LastBlockHeight=2; validators saved at 1–4
	realStore := sm.NewStore(stateDB, sm.StoreOptions{})
	valSet2, err := realStore.LoadValidators(2)
	require.NoError(t, err)
	valSet3, err := realStore.LoadValidators(3)
	require.NoError(t, err)

	storeMock := mocks.NewStore(t)
	storeMock.On("LoadValidators", int64(2)).Return(valSet2, nil)
	storeMock.On("LoadValidators", int64(3)).Return(valSet3, nil)

	app := abcimocks.NewApplication(t)
	app.On("ProcessProposal", mock.Anything, mock.Anything).
		Return(&abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil)

	proxyApp := proxy.NewAppConns(proxy.NewLocalClientCreator(app), proxy.NopMetrics())
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { _ = proxyApp.Stop() })

	blockExec := sm.NewBlockExecutor(
		storeMock, log.NewNopLogger(), proxyApp.Consensus(),
		new(mpmocks.Mempool), sm.EmptyEvidencePool{},
		store.NewBlockStore(dbm.NewMemDB()),
	)

	// Block cycle 1: block at height 3, loads validators at height 2 (cache miss).
	block1, err := makeBlock(state, 3, oneValCommit(state.LastBlockHeight))
	require.NoError(t, err)
	_, err = blockExec.ProcessProposal(block1, state)
	require.NoError(t, err)

	// Block cycle 2: block at height 4, must load validators at height 3 (cache miss).
	// The height-2 entry must not be reused.
	state2 := state
	state2.LastBlockHeight = 3
	block2, err := makeBlock(state2, 4, oneValCommit(3))
	require.NoError(t, err)
	_, err = blockExec.ProcessProposal(block2, state2)
	require.NoError(t, err)

	storeMock.AssertNumberOfCalls(t, "LoadValidators", 2)
}

func TestValidatorCacheHitAcrossProposalAndProcess(t *testing.T) {
	state, stateDB, privVals := makeState(1, 3)
	realStore := sm.NewStore(stateDB, sm.StoreOptions{})
	valSet, err := realStore.LoadValidators(state.LastBlockHeight)
	require.NoError(t, err)

	storeMock := mocks.NewStore(t)
	storeMock.On("LoadValidators", state.LastBlockHeight).Return(valSet, nil)

	app := abcimocks.NewApplication(t)
	app.On("PrepareProposal", mock.Anything, mock.Anything).
		Return(&abci.ResponsePrepareProposal{}, nil)
	app.On("ProcessProposal", mock.Anything, mock.Anything).
		Return(&abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil)

	proxyApp := proxy.NewAppConns(proxy.NewLocalClientCreator(app), proxy.NopMetrics())
	require.NoError(t, proxyApp.Start())
	t.Cleanup(func() { _ = proxyApp.Stop() })

	mp := new(mpmocks.Mempool)
	mp.On("ReapMaxBytesMaxGas", mock.Anything, mock.Anything).Return(types.Txs{})

	blockExec := sm.NewBlockExecutor(
		storeMock, log.NewNopLogger(), proxyApp.Consensus(),
		mp, sm.EmptyEvidencePool{},
		store.NewBlockStore(dbm.NewMemDB()),
	)

	proposerAddr, _ := state.Validators.GetByIndex(0)
	lastExtCommit, _, err := makeValidCommit(state.LastBlockHeight, types.BlockID{}, state.Validators, privVals)
	require.NoError(t, err)

	// CreateProposalBlock: cache miss → 1 DB load.
	block, err := blockExec.CreateProposalBlock(t.Context(), state.LastBlockHeight+1, state, lastExtCommit, proposerAddr)
	require.NoError(t, err)

	// ProcessProposal: same height validators → cache hit, no additional DB load.
	_, err = blockExec.ProcessProposal(block, state)
	require.NoError(t, err)

	storeMock.AssertNumberOfCalls(t, "LoadValidators", 1)
}

func BenchmarkLoadValidatorsNoCache(b *testing.B) {
	for _, nVals := range []int{10, 100} {
		b.Run(fmt.Sprintf("%dvals", nVals), func(b *testing.B) {
			state, stateDB, _ := makeState(nVals, 10)
			stateStore := sm.NewStore(stateDB, sm.StoreOptions{})
			loadHeight := state.LastBlockHeight

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkLoadValidatorsWithCache(b *testing.B) {
	for _, nVals := range []int{10, 100} {
		b.Run(fmt.Sprintf("%dvals", nVals), func(b *testing.B) {
			state, stateDB, _ := makeState(nVals, 10)
			stateStore := sm.NewStore(stateDB, sm.StoreOptions{})
			loadHeight := state.LastBlockHeight

			var cache atomic.Pointer[types.ValidatorSet]

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				v, err := stateStore.LoadValidators(loadHeight)
				if err != nil {
					b.Fatal(err)
				}
				cache.Store(v)
				_ = cache.Load()
				_ = cache.Load()
			}
		})
	}
}
