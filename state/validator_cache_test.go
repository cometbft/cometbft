package state_test

import (
	"context"
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

// oneValCommit returns a last-commit with a single absent signature.
// It satisfies BuildLastCommitInfo's len(Signatures)==len(Validators) invariant
// when using a single-validator state.
func oneValCommit(height int64) *types.Commit {
	return &types.Commit{
		Height:     height,
		Signatures: []types.CommitSig{types.NewCommitSigAbsent()},
	}
}

// TestValidatorCacheHitWithinBlockCycle verifies that ProcessProposal and
// ExtendVote share the same cached validator set within one block cycle.
// Only the first call should reach the DB; the second must be a cache hit.
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

// TestValidatorCacheInvalidatesOnNewHeight verifies that the cache is not
// carried over between block cycles. When the block height advances, the
// executor must make a fresh DB load rather than reusing the stale entry.
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
