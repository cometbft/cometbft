package state_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
)

func TestStoreLoadValidators(t *testing.T) {
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	val, _ := types.RandValidator(true, 10)
	vals := types.NewValidatorSet([]*types.Validator{val})

	// 1) LoadValidators loads validators using a height where they were last changed
	err := sm.SaveValidatorsInfo(stateDB, 1, 1, vals)
	require.NoError(t, err)
	err = sm.SaveValidatorsInfo(stateDB, 2, 1, vals)
	require.NoError(t, err)
	loadedVals, err := stateStore.LoadValidators(2)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())

	// 2) LoadValidators loads validators using a checkpoint height

	err = sm.SaveValidatorsInfo(stateDB, sm.ValSetCheckpointInterval, 1, vals)
	require.NoError(t, err)

	loadedVals, err = stateStore.LoadValidators(sm.ValSetCheckpointInterval)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())
}

func BenchmarkLoadValidators(b *testing.B) {
	const valSetSize = 100

	config := cfg.ResetTestRoot("state_")
	defer os.RemoveAll(config.RootDir)
	dbType := dbm.BackendType(config.DBBackend)
	stateDB, err := dbm.NewDB("state", dbType, config.DBDir())
	require.NoError(b, err)
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisFile(config.GenesisFile())
	if err != nil {
		b.Fatal(err)
	}

	state.Validators = genValSet(valSetSize)
	state.NextValidators = state.Validators.CopyIncrementProposerPriority(1)
	err = stateStore.Save(state)
	require.NoError(b, err)

	for i := 10; i < 10000000000; i *= 10 { // 10, 100, 1000, ...
		i := i
		if err := sm.SaveValidatorsInfo(stateDB,
			int64(i), state.LastHeightValidatorsChanged, state.NextValidators); err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("height=%d", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err := stateStore.LoadValidators(int64(i))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestPruneStates(t *testing.T) {
	testcases := map[string]struct {
		makeHeights  int64
		pruneFrom    int64
		pruneTo      int64
		expectErr    bool
		expectVals   []int64
		expectParams []int64
		expectABCI   []int64
	}{
		"error on pruning from 0":      {100, 0, 5, true, nil, nil, nil},
		"error when from > to":         {100, 3, 2, true, nil, nil, nil},
		"error when from == to":        {100, 3, 3, true, nil, nil, nil},
		"error when to does not exist": {100, 1, 101, true, nil, nil, nil},
		"prune all":                    {100, 1, 100, false, []int64{93, 100}, []int64{95, 100}, []int64{100}},
		"prune some": {10, 2, 8, false, []int64{1, 3, 8, 9, 10},
			[]int64{1, 5, 8, 9, 10}, []int64{1, 8, 9, 10}},
		"prune across checkpoint": {100001, 1, 100001, false, []int64{99993, 100000, 100001},
			[]int64{99995, 100001}, []int64{100001}},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			db := dbm.NewMemDB()
			stateStore := sm.NewStore(db, sm.StoreOptions{
				DiscardABCIResponses: false,
			})
			pk := ed25519.GenPrivKey().PubKey()

			// Generate a bunch of state data. Validators change for heights ending with 3, and
			// parameters when ending with 5.
			validator := &types.Validator{Address: pk.Address(), VotingPower: 100, PubKey: pk}
			validatorSet := &types.ValidatorSet{
				Validators: []*types.Validator{validator},
				Proposer:   validator,
			}
			valsChanged := int64(0)
			paramsChanged := int64(0)

			for h := int64(1); h <= tc.makeHeights; h++ {
				if valsChanged == 0 || h%10 == 2 {
					valsChanged = h + 1 // Have to add 1, since NextValidators is what's stored
				}
				if paramsChanged == 0 || h%10 == 5 {
					paramsChanged = h
				}

				state := sm.State{
					InitialHeight:   1,
					LastBlockHeight: h - 1,
					Validators:      validatorSet,
					NextValidators:  validatorSet,
					ConsensusParams: types.ConsensusParams{
						Block: types.BlockParams{MaxBytes: 10e6},
					},
					LastHeightValidatorsChanged:      valsChanged,
					LastHeightConsensusParamsChanged: paramsChanged,
				}

				if state.LastBlockHeight >= 1 {
					state.LastValidators = state.Validators
				}

				err := stateStore.Save(state)
				require.NoError(t, err)

				err = stateStore.SaveABCIResponses(h, &cmtstate.ABCIResponses{
					DeliverTxs: []*abci.ResponseDeliverTx{
						{Data: []byte{1}},
						{Data: []byte{2}},
						{Data: []byte{3}},
					},
				})
				require.NoError(t, err)
			}

			// Test assertions
			err := stateStore.PruneStates(tc.pruneFrom, tc.pruneTo)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			expectVals := sliceToMap(tc.expectVals)
			expectParams := sliceToMap(tc.expectParams)
			expectABCI := sliceToMap(tc.expectABCI)

			for h := int64(1); h <= tc.makeHeights; h++ {
				vals, err := stateStore.LoadValidators(h)
				if expectVals[h] {
					require.NoError(t, err, "validators height %v", h)
					require.NotNil(t, vals)
				} else {
					require.Error(t, err, "validators height %v", h)
					require.Equal(t, sm.ErrNoValSetForHeight{Height: h}, err)
				}

				params, err := stateStore.LoadConsensusParams(h)
				if expectParams[h] {
					require.NoError(t, err, "params height %v", h)
					require.NotEmpty(t, params)
				} else {
					require.Error(t, err, "params height %v", h)
					require.Empty(t, params)
				}

				abci, err := stateStore.LoadABCIResponses(h)
				if expectABCI[h] {
					require.NoError(t, err, "abci height %v", h)
					require.NotNil(t, abci)
				} else {
					require.Error(t, err, "abci height %v", h)
					require.Equal(t, sm.ErrNoABCIResponsesForHeight{Height: h}, err)
				}
			}
		})
	}
}

func TestABCIResponsesResultsHash(t *testing.T) {
	responses := &cmtstate.ABCIResponses{
		BeginBlock: &abci.ResponseBeginBlock{},
		DeliverTxs: []*abci.ResponseDeliverTx{
			{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
		},
		EndBlock: &abci.ResponseEndBlock{},
	}

	root := sm.ABCIResponsesResultsHash(responses)

	// root should be Merkle tree root of DeliverTxs responses
	results := types.NewResults(responses.DeliverTxs)
	assert.Equal(t, root, results.Hash())

	// test we can prove first DeliverTx
	proof := results.ProveResult(0)
	bz, err := results[0].Marshal()
	require.NoError(t, err)
	assert.NoError(t, proof.Verify(root, bz))
}

func sliceToMap(s []int64) map[int64]bool {
	m := make(map[int64]bool, len(s))
	for _, i := range s {
		m[i] = true
	}
	return m
}

func makeStateAndBlockStore() (sm.State, *store.BlockStore, func(), sm.Store) {
	config := cfg.ResetTestRoot("blockchain_reactor_test")
	blockDB := dbm.NewMemDB()
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	state, err := stateStore.LoadFromDBOrGenesisFile(config.GenesisFile())
	if err != nil {
		panic(fmt.Errorf("error constructing state from genesis file: %w", err))
	}
	return state, store.NewBlockStore(blockDB), func() { os.RemoveAll(config.RootDir) }, stateStore
}

func fillStore(t *testing.T, height int64, stateStore sm.Store, bs *store.BlockStore, state sm.State, response1 *cmtstate.ABCIResponses) {
	if response1 != nil {
		for h := int64(1); h <= height; h++ {
			err := stateStore.SaveABCIResponses(h, response1)
			require.NoError(t, err)
		}
		// search for the last finalize block response and check if it has saved.
		lastResponse, err := stateStore.LoadLastABCIResponse(height)
		require.NoError(t, err)
		// check to see if the saved response height is the same as the loaded height.
		assert.Equal(t, lastResponse, response1)
		// check if the abci response didnt save in the abciresponses.
		responses, err := stateStore.LoadABCIResponses(height)
		require.NoError(t, err, responses)
		require.Equal(t, response1, responses)
	}
	b1 := state.MakeBlock(state.LastBlockHeight+1, test.MakeNTxs(state.LastBlockHeight+1, 10), new(types.Commit), nil, nil)
	partSet, err := b1.MakePartSet(2)
	require.NoError(t, err)
	bs.SaveBlock(b1, partSet, &types.Commit{Height: state.LastBlockHeight + 1})
}

func TestSaveRetainHeight(t *testing.T) {
	state, bs, callbackF, stateStore := makeStateAndBlockStore()
	defer callbackF()
	height := int64(10)
	state.LastBlockHeight = height - 1

	fillStore(t, height, stateStore, bs, state, nil)

	pruner := sm.NewPruner(stateStore, bs, log.TestingLogger())

	// We should not save a height that is 0
	err := pruner.SetApplicationRetainHeight(0)
	require.Error(t, err)

	// We should not save a height above the blockstore's height
	err = pruner.SetApplicationRetainHeight(11)
	require.Error(t, err)

	err = pruner.SetApplicationRetainHeight(10)
	require.NoError(t, err)

	err = pruner.SetCompanionRetainHeight(10)
	require.NoError(t, err)
}

func TestMinRetainHeight(t *testing.T) {
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})
	pruner := sm.NewPruner(stateStore, nil, log.TestingLogger())
	minHeight := pruner.FindMinRetainHeight()
	require.Equal(t, minHeight, int64(0))

	err := stateStore.SaveApplicationRetainHeight(10)
	require.NoError(t, err)
	minHeight = pruner.FindMinRetainHeight()
	require.Equal(t, minHeight, int64(10))

	err = stateStore.SaveCompanionBlockRetainHeight(11)
	require.NoError(t, err)
	minHeight = pruner.FindMinRetainHeight()
	require.Equal(t, minHeight, int64(10))
}

func TestABCIResPruningStandalone(t *testing.T) {
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})

	responses, err := stateStore.LoadABCIResponses(1)
	require.Error(t, err)
	require.Nil(t, responses)
	// stub the abciresponses.
	response1 := &cmtstate.ABCIResponses{
		DeliverTxs: []*abci.ResponseDeliverTx{
			{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
		},
	}
	_, bs, callbackF, stateStore := makeStateAndBlockStore()
	defer callbackF()

	for height := int64(1); height <= 10; height++ {
		err := stateStore.SaveABCIResponses(height, response1)
		require.NoError(t, err)
	}
	pruner := sm.NewPruner(stateStore, bs, log.TestingLogger())

	retainHeight := int64(2)
	err = stateStore.SaveABCIResRetainHeight(retainHeight)
	require.NoError(t, err)
	abciResRetainHeight, err := stateStore.GetABCIResRetainHeight()
	require.NoError(t, err)
	require.Equal(t, retainHeight, abciResRetainHeight)
	newRetainHeight := pruner.PruneABCIResToRetainHeight(0)
	require.Equal(t, retainHeight, newRetainHeight)

	_, err = stateStore.LoadABCIResponses(1)
	require.Error(t, err)

	for h := retainHeight; h <= 10; h++ {
		_, err = stateStore.LoadABCIResponses(h)
		require.NoError(t, err)
	}

	// This should not have any impact because the retain height is still 2 and we will not prune blocks to 3
	newRetainHeight = pruner.PruneABCIResToRetainHeight(3)
	require.Equal(t, retainHeight, newRetainHeight)

	retainHeight = 3
	err = stateStore.SaveABCIResRetainHeight(retainHeight)
	require.NoError(t, err)
	newRetainHeight = pruner.PruneABCIResToRetainHeight(2)
	require.Equal(t, retainHeight, newRetainHeight)

	_, err = stateStore.LoadABCIResponses(2)
	require.Error(t, err)
	for h := retainHeight; h <= 10; h++ {
		_, err = stateStore.LoadABCIResponses(h)
		require.NoError(t, err)
	}

	retainHeight = 10
	err = stateStore.SaveABCIResRetainHeight(retainHeight)
	require.NoError(t, err)
	newRetainHeight = pruner.PruneABCIResToRetainHeight(2)
	require.Equal(t, retainHeight, newRetainHeight)

	for h := int64(0); h < 10; h++ {
		_, err = stateStore.LoadABCIResponses(h)
		require.Error(t, err)
	}
	_, err = stateStore.LoadABCIResponses(10)
	require.NoError(t, err)
}

type prunerObserver struct {
	sm.NoopPrunerObserver
	prunedABCIResInfoCh   chan *sm.ABCIResponsesPrunedInfo
	prunedBlocksResInfoCh chan *sm.BlocksPrunedInfo
}

func newPrunerObserver(infoChCap int) *prunerObserver {
	return &prunerObserver{
		prunedABCIResInfoCh:   make(chan *sm.ABCIResponsesPrunedInfo, infoChCap),
		prunedBlocksResInfoCh: make(chan *sm.BlocksPrunedInfo, infoChCap),
	}
}

func (o *prunerObserver) PrunerPrunedABCIRes(info *sm.ABCIResponsesPrunedInfo) {
	o.prunedABCIResInfoCh <- info
}
func (o *prunerObserver) PrunerPrunedBlocks(info *sm.BlocksPrunedInfo) {
	o.prunedBlocksResInfoCh <- info
}

func TestLastABCIResponses(t *testing.T) {
	t.Run("Persisting responses", func(t *testing.T) {
		stateDB := dbm.NewMemDB()
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: false,
		})
		responses, err := stateStore.LoadABCIResponses(1)
		require.Error(t, err)
		require.Nil(t, responses)

		// stub the abciresponses.
		response1 := &cmtstate.ABCIResponses{
			DeliverTxs: []*abci.ResponseDeliverTx{
				{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
			},
		}
		state, bs, callbackF, stateStore := makeStateAndBlockStore()
		defer callbackF()
		height := int64(10)
		state.LastBlockHeight = height - 1

		fillStore(t, height, stateStore, bs, state, response1)

		obs := newPrunerObserver(1)
		pruner := sm.NewPruner(
			stateStore,
			bs,
			log.TestingLogger(),
			sm.WithPrunerInterval(1*time.Second),
			sm.WithPrunerObserver(obs),
		)

		// Check that we have written a finalize block result at height 'height - 1'
		_, err = stateStore.LoadABCIResponses(height - 1)
		require.NoError(t, err)
		require.NoError(t, pruner.SetABCIResRetainHeight(height))
		require.NoError(t, pruner.Start())

		select {
		case info := <-obs.prunedABCIResInfoCh:
			require.Equal(t, height-1, info.ToHeight)
			t.Log("Done pruning ABCI results ")
		case <-time.After(5 * time.Second):
			require.Fail(t, "timed out waiting for pruning run to complete")
		}

		// Check that the response at height h - 1 has been deleted
		_, err = stateStore.LoadABCIResponses(height - 1)
		require.Error(t, err)
		_, err = stateStore.LoadABCIResponses(height)
		require.NoError(t, err)

	})

}

func TestLastFinalizeBlockResponses(t *testing.T) {
	// create an empty state store.
	t.Run("Not persisting responses", func(t *testing.T) {
		stateDB := dbm.NewMemDB()
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: false,
		})
		responses, err := stateStore.LoadABCIResponses(1)
		require.Error(t, err)
		require.Nil(t, responses)
		// stub the abciresponses.
		response1 := &cmtstate.ABCIResponses{
			BeginBlock: &abci.ResponseBeginBlock{},
			DeliverTxs: []*abci.ResponseDeliverTx{
				{Code: 32, Data: []byte("Hello"), Log: "Huh?"},
			},
			EndBlock: &abci.ResponseEndBlock{},
		}
		// create new db and state store and set discard abciresponses to false.
		stateDB = dbm.NewMemDB()
		stateStore = sm.NewStore(stateDB, sm.StoreOptions{DiscardABCIResponses: false})
		height := int64(10)
		// save the last abci response.
		err = stateStore.SaveABCIResponses(height, response1)
		require.NoError(t, err)
		// search for the last abciresponse and check if it has saved.
		lastResponse, err := stateStore.LoadLastABCIResponse(height)
		require.NoError(t, err)
		// check to see if the saved response height is the same as the loaded height.
		assert.Equal(t, lastResponse, response1)
		// use an incorret height to make sure the state store errors.
		_, err = stateStore.LoadLastABCIResponse(height + 1)
		assert.Error(t, err)
		// check if the abci response didnt save in the abciresponses.
		responses, err = stateStore.LoadABCIResponses(height)
		require.NoError(t, err, responses)
		require.Equal(t, response1, responses)
	})

	t.Run("persisting responses", func(t *testing.T) {
		stateDB := dbm.NewMemDB()
		height := int64(10)
		// stub the second abciresponse.
		response2 := &cmtstate.ABCIResponses{
			BeginBlock: &abci.ResponseBeginBlock{},
			DeliverTxs: []*abci.ResponseDeliverTx{
				{Code: 44, Data: []byte("Hello again"), Log: "????"},
			},
			EndBlock: &abci.ResponseEndBlock{},
		}
		// create a new statestore with the responses on.
		stateStore := sm.NewStore(stateDB, sm.StoreOptions{
			DiscardABCIResponses: true,
		})
		// save an additional response.
		err := stateStore.SaveABCIResponses(height+1, response2)
		require.NoError(t, err)
		// check to see if the response saved by calling the last response.
		lastResponse2, err := stateStore.LoadLastABCIResponse(height + 1)
		require.NoError(t, err)
		// check to see if the saved response height is the same as the loaded height.
		assert.Equal(t, response2, lastResponse2)
		// should error as we are no longer saving the response.
		_, err = stateStore.LoadABCIResponses(height + 1)
		assert.Equal(t, sm.ErrABCIResponsesNotPersisted, err)
	})
}
func TestIntConversion(t *testing.T) {
	x := int64(10)
	b := sm.Int64ToBytes(x)
	require.Equal(t, x, sm.Int64FromBytes(b))
}
