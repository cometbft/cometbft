package consensus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/internal/test"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	"github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"
)

// TestDoubleSignCheckHeightOne tests that when double_sign_check_height = 1,
// the checkDoubleSigningRisk function properly checks at least one block
// (the current height - 1) for double signing risks.
func TestDoubleSignCheckHeightOne(t *testing.T) {
	// Create validators
	state, privVals := randGenesisState(1, false, 10, test.ConsensusParams())

	// Get private validator
	pv := privVals[0]
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)

	// Create block store and state store
	blockDB := dbm.NewMemDB()
	blockStore := store.NewBlockStore(blockDB)
	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})

	// Save initial state
	err = stateStore.Save(state)
	require.NoError(t, err)

	// Create a commit at height 1 with our validator's signature
	height := int64(1)
	blockHash := make([]byte, 32)
	copy(blockHash, []byte("test_block_hash"))
	partHash := make([]byte, 32)
	copy(partHash, []byte("test_part_set_hash"))

	blockID := types.BlockID{
		Hash: blockHash,
		PartSetHeader: types.PartSetHeader{
			Total: 1,
			Hash:  partHash,
		},
	}

	// Create vote
	vote := &types.Vote{
		Type:             cmtproto.PrecommitType,
		Height:           height,
		Round:            0,
		BlockID:          blockID,
		Timestamp:        tmtime.Now(),
		ValidatorAddress: pubKey.Address(),
		ValidatorIndex:   0,
	}

	// Sign the vote
	v := vote.ToProto()
	err = pv.SignVote(state.ChainID, v)
	require.NoError(t, err)
	vote.Signature = v.Signature

	// Create commit with the signature
	commit := &types.Commit{
		Height:  height,
		Round:   0,
		BlockID: blockID,
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: pubKey.Address(),
				Timestamp:        vote.Timestamp,
				Signature:        vote.Signature,
			},
		},
	}

	// Store the commit as a seen commit
	blockStore.SaveSeenCommit(height, commit)

	// Create consensus state
	app := kvstore.NewInMemoryApplication()
	cs := newState(state, pv, app)
	cs.SetLogger(log.TestingLogger())

	// Set private validator and blockstore
	cs.SetPrivValidator(pv)
	cs.blockStore = blockStore

	// Override config
	cs.config.DoubleSignCheckHeight = 1

	// Test 1: checkDoubleSigningRisk should detect the existing signature at height 1
	// when we try to join consensus at height 2
	err = cs.checkDoubleSigningRisk(2)
	assert.Error(t, err, "Should detect double signing risk when double_sign_check_height = 1")
	assert.Equal(t, ErrSignatureFoundInPastBlocks, err, "Should return ErrSignatureFoundInPastBlocks")

	// Test 2: No error when checking at height 1 (no previous blocks to check)
	err = cs.checkDoubleSigningRisk(1)
	assert.NoError(t, err, "Should not error when no previous blocks exist")

	// Test 3: Verify that with DoubleSignCheckHeight = 0, no checks are performed
	cs.config.DoubleSignCheckHeight = 0
	err = cs.checkDoubleSigningRisk(2)
	assert.NoError(t, err, "Should not check when DoubleSignCheckHeight = 0")
}

// TestDoubleSignCheckHeightMultiple tests various values of double_sign_check_height
// to ensure the function checks the correct number of blocks
func TestDoubleSignCheckHeightMultiple(t *testing.T) {
	testCases := []struct {
		name                  string
		doubleSignCheckHeight int
		currentHeight         int64
		signedHeights         []int64
		expectError           bool
	}{
		{
			name:                  "check_height_1_with_signature_at_height_minus_1",
			doubleSignCheckHeight: 1,
			currentHeight:         5,
			signedHeights:         []int64{4},
			expectError:           true, // Should detect the signature at height 4
		},
		{
			name:                  "check_height_1_no_signature_at_height_minus_1",
			doubleSignCheckHeight: 1,
			currentHeight:         5,
			signedHeights:         []int64{2, 3}, // No signature at height 4
			expectError:           false,
		},
		{
			name:                  "check_height_2_with_signatures",
			doubleSignCheckHeight: 2,
			currentHeight:         5,
			signedHeights:         []int64{3, 4},
			expectError:           true, // Should detect signatures at heights 3 or 4
		},
		{
			name:                  "check_height_3_with_old_signature",
			doubleSignCheckHeight: 3,
			currentHeight:         5,
			signedHeights:         []int64{2},
			expectError:           true, // Should detect signature at height 2
		},
		{
			name:                  "check_height_exceeds_current_height",
			doubleSignCheckHeight: 10,
			currentHeight:         3,
			signedHeights:         []int64{1, 2},
			expectError:           true, // Should check all available blocks
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create validators
			state, privVals := randGenesisState(1, false, 10, test.ConsensusParams())
			pv := privVals[0]
			pubKey, err := pv.GetPubKey()
			require.NoError(t, err)

			blockDB := dbm.NewMemDB()
			blockStore := store.NewBlockStore(blockDB)

			// Create and store commits at specified heights
			for _, height := range tc.signedHeights {
				bHash := make([]byte, 32)
				bHash[0] = byte(height)
				pHash := make([]byte, 32)
				pHash[0] = byte(height)

				blockID := types.BlockID{
					Hash: bHash,
					PartSetHeader: types.PartSetHeader{
						Total: 1,
						Hash:  pHash,
					},
				}

				vote := &types.Vote{
					Type:             cmtproto.PrecommitType,
					Height:           height,
					Round:            0,
					BlockID:          blockID,
					Timestamp:        tmtime.Now(),
					ValidatorAddress: pubKey.Address(),
					ValidatorIndex:   0,
				}

				v := vote.ToProto()
				err := pv.SignVote(state.ChainID, v)
				require.NoError(t, err)
				vote.Signature = v.Signature

				commit := &types.Commit{
					Height:  height,
					Round:   0,
					BlockID: blockID,
					Signatures: []types.CommitSig{
						{
							BlockIDFlag:      types.BlockIDFlagCommit,
							ValidatorAddress: pubKey.Address(),
							Timestamp:        vote.Timestamp,
							Signature:        vote.Signature,
						},
					},
				}

				blockStore.SaveSeenCommit(height, commit)
			}

			// Create consensus state
			app := kvstore.NewInMemoryApplication()
			cs := newState(state, pv, app)
			cs.SetLogger(log.TestingLogger())
			cs.SetPrivValidator(pv)
			cs.blockStore = blockStore
			cs.config.DoubleSignCheckHeight = int64(tc.doubleSignCheckHeight)

			// Test
			err = cs.checkDoubleSigningRisk(tc.currentHeight)
			if tc.expectError {
				assert.Error(t, err, "Expected double signing risk to be detected")
				assert.Equal(t, ErrSignatureFoundInPastBlocks, err)
			} else {
				assert.NoError(t, err, "Expected no double signing risk")
			}
		})
	}
}

// TestDoubleSignCheckWithRestart simulates a validator restart scenario
// where the validator has already signed blocks and attempts to rejoin consensus
func TestDoubleSignCheckWithRestart(t *testing.T) {
	// Create validators - using simpler approach
	state, privVals := randGenesisState(1, false, 10, test.ConsensusParams())
	pv := privVals[0]
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)

	blockDB := dbm.NewMemDB()
	blockStore := store.NewBlockStore(blockDB)

	// Simulate validator signing at height 10
	currentHeight := int64(10)
	blockHash10 := make([]byte, 32)
	copy(blockHash10, []byte("block_10_hash"))
	partHash10 := make([]byte, 32)
	copy(partHash10, []byte("part_10_hash"))

	blockID := types.BlockID{
		Hash: blockHash10,
		PartSetHeader: types.PartSetHeader{
			Total: 1,
			Hash:  partHash10,
		},
	}

	vote := &types.Vote{
		Type:             cmtproto.PrecommitType,
		Height:           currentHeight,
		Round:            0,
		BlockID:          blockID,
		Timestamp:        tmtime.Now(),
		ValidatorAddress: pubKey.Address(),
		ValidatorIndex:   0,
	}

	v := vote.ToProto()
	err = pv.SignVote(state.ChainID, v)
	require.NoError(t, err)
	vote.Signature = v.Signature

	commit := &types.Commit{
		Height:  currentHeight,
		Round:   0,
		BlockID: blockID,
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: pubKey.Address(),
				Timestamp:        vote.Timestamp,
				Signature:        vote.Signature,
			},
		},
	}

	blockStore.SaveSeenCommit(currentHeight, commit)

	// Create a simple mock consensus state for testing (not using newState)
	testCfg := test.ResetTestRoot("consensus_double_sign_restart")
	cs := &State{
		config:              testCfg.Consensus,
		privValidator:       pv,
		privValidatorPubKey: pubKey,
		blockStore:          blockStore,
	}
	cs.BaseService = *service.NewBaseService(nil, "State", cs)
	cs.Logger = log.TestingLogger()
	cs.config.DoubleSignCheckHeight = 1

	// Test: When validator restarts and tries to participate at height 11,
	// it should detect the previous signature at height 10
	err = cs.checkDoubleSigningRisk(11)
	assert.Error(t, err, "Should detect previous signature after restart with double_sign_check_height = 1")
	assert.Equal(t, ErrSignatureFoundInPastBlocks, err, "Should return ErrSignatureFoundInPastBlocks")

	t.Logf("Double sign check with height=1 correctly detected previous signature at height %d when joining at height %d", currentHeight, currentHeight+1)
}

// TestDoubleSignCheckProtectsAgainstDoubleSign verifies that the fix prevents actual double signing
// This test demonstrates that with double_sign_check_height = 1, the validator is protected
// from double signing even without FilePrivVal state
func TestDoubleSignCheckProtectsAgainstDoubleSign(t *testing.T) {
	// Create validators
	state, privVals := randGenesisState(1, false, 10, test.ConsensusParams())
	pv := privVals[0]
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)

	// Instance 1: Has signed at height 5
	blockDB1 := dbm.NewMemDB()
	blockStore1 := store.NewBlockStore(blockDB1)

	height := int64(5)
	blockHash1 := make([]byte, 32)
	copy(blockHash1, []byte("block_hash_instance_1"))
	partHash1 := make([]byte, 32)
	copy(partHash1, []byte("part_hash_instance_1"))

	blockID1 := types.BlockID{
		Hash: blockHash1,
		PartSetHeader: types.PartSetHeader{
			Total: 1,
			Hash:  partHash1,
		},
	}

	vote1 := &types.Vote{
		Type:             cmtproto.PrecommitType,
		Height:           height,
		Round:            0,
		BlockID:          blockID1,
		Timestamp:        tmtime.Now(),
		ValidatorAddress: pubKey.Address(),
		ValidatorIndex:   0,
	}

	v1 := vote1.ToProto()
	err = pv.SignVote(state.ChainID, v1)
	require.NoError(t, err)
	vote1.Signature = v1.Signature

	commit1 := &types.Commit{
		Height:  height,
		Round:   0,
		BlockID: blockID1,
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: pubKey.Address(),
				Timestamp:        vote1.Timestamp,
				Signature:        vote1.Signature,
			},
		}}

	blockStore1.SaveSeenCommit(height, commit1)

	// Instance 2: Trying to sign different block at same height (potential double sign)
	blockDB2 := dbm.NewMemDB()
	blockStore2 := store.NewBlockStore(blockDB2)

	// Copy the commit from instance 1 to instance 2's blockstore
	// (simulating network sync or shared storage)
	blockStore2.SaveSeenCommit(height, commit1)

	// Create consensus state for instance 2
	app := kvstore.NewInMemoryApplication()
	cs2 := newState(state, pv, app)
	cs2.SetLogger(log.TestingLogger())
	cs2.SetPrivValidator(pv)
	cs2.blockStore = blockStore2
	cs2.config.DoubleSignCheckHeight = 1

	// Test: Instance 2 should detect the existing signature and refuse to participate
	// This prevents double signing at height 6 (checking height 5)
	err = cs2.checkDoubleSigningRisk(6)
	assert.Error(t, err, "Should prevent double signing by detecting existing signature")
	assert.Equal(t, ErrSignatureFoundInPastBlocks, err)

	t.Log("Successfully prevented potential double signing scenario with double_sign_check_height = 1")
}
