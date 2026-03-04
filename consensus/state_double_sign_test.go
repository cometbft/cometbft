package consensus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	smmocks "github.com/cometbft/cometbft/state/mocks"
	"github.com/cometbft/cometbft/types"
)

// newStateForDoubleSignTest creates a minimal State configured for testing
// checkDoubleSigningRisk. It wires up a real privValidator (MockPV) and a
// mocked blockStore, avoiding the heavy full-consensus setup that is
// unnecessary for unit-testing this single method.
func newStateForDoubleSignTest(t *testing.T, doubleSignCheckHeight int64) (*State, *smmocks.BlockStore) {
	t.Helper()

	consConfig := cfg.DefaultConsensusConfig()
	consConfig.DoubleSignCheckHeight = doubleSignCheckHeight

	mockBS := &smmocks.BlockStore{}

	cs := &State{
		config:     consConfig,
		blockStore: mockBS,
	}
	// Set logger directly on the embedded BaseService to avoid nil-dereference
	// on timeoutTicker which is not needed for this unit test.
	cs.Logger = log.TestingLogger()

	// Wire a real private validator so privValidatorPubKey can be set.
	pv := types.NewMockPV()
	pubKey, err := pv.GetPubKey()
	require.NoError(t, err)

	cs.privValidator = pv
	cs.privValidatorPubKey = pubKey

	return cs, mockBS
}

// makeCommitWithValidator builds a *types.Commit whose single signature carries
// the provided validator address with BlockIDFlagCommit – the exact condition
// checkDoubleSigningRisk looks for.
func makeCommitWithValidator(height int64, validatorAddr types.Address) *types.Commit {
	return &types.Commit{
		Height: height,
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: validatorAddr,
				Timestamp:        time.Now(),
			},
		},
	}
}

// makeCommitWithDifferentValidator builds a *types.Commit containing a single
// signature that belongs to a *different* validator, so checkDoubleSigningRisk
// should NOT trigger.
func makeCommitWithDifferentValidator(height int64) *types.Commit {
	otherPV := types.NewMockPV()
	otherPub, _ := otherPV.GetPubKey()
	return &types.Commit{
		Height: height,
		Signatures: []types.CommitSig{
			{
				BlockIDFlag:      types.BlockIDFlagCommit,
				ValidatorAddress: otherPub.Address(),
				Timestamp:        time.Now(),
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Test 1 – core regression: doubleSignCheckHeight=1 must check height-1
// ---------------------------------------------------------------------------

// TestCheckDoubleSigningRisk_HeightOne_ChecksOnePreviousBlock demonstrates the
// off-by-one bug: with double_sign_check_height=1 the original loop
//
//	for i := int64(1); i < 1; i++ { … }
//
// never executes, so a signature present at height-1 is silently ignored.
// After the fix (i <= doubleSignCheckHeight) the loop runs once with i=1,
// loads the commit at height-1, and returns ErrSignatureFoundInPastBlocks.
func TestCheckDoubleSigningRisk_HeightOne_ChecksOnePreviousBlock(t *testing.T) {
	const targetHeight = int64(10)

	cs, mockBS := newStateForDoubleSignTest(t, 1)

	// The validator signed the commit at height-1.
	commit := makeCommitWithValidator(targetHeight-1, cs.privValidatorPubKey.Address())
	mockBS.On("LoadSeenCommit", targetHeight-int64(1)).Return(commit)

	err := cs.checkDoubleSigningRisk(targetHeight)

	require.ErrorIs(t, err, ErrSignatureFoundInPastBlocks,
		"expected ErrSignatureFoundInPastBlocks when doubleSignCheckHeight=1 "+
			"and a matching signature exists at height-1")

	mockBS.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Test 2 – doubleSignCheckHeight=0 must remain disabled (no regression)
// ---------------------------------------------------------------------------

// TestCheckDoubleSigningRisk_HeightZero_Disabled confirms that setting
// double_sign_check_height to 0 completely disables the check regardless of
// what the blockstore contains.
func TestCheckDoubleSigningRisk_HeightZero_Disabled(t *testing.T) {
	const targetHeight = int64(10)

	cs, mockBS := newStateForDoubleSignTest(t, 0)

	// No blockstore calls should be made when the feature is disabled.
	err := cs.checkDoubleSigningRisk(targetHeight)

	require.NoError(t, err, "expected nil when doubleSignCheckHeight=0 (feature disabled)")
	mockBS.AssertNotCalled(t, "LoadSeenCommit")
}

// ---------------------------------------------------------------------------
// Test 3 – doubleSignCheckHeight=2 must check both height-1 and height-2
// ---------------------------------------------------------------------------

// TestCheckDoubleSigningRisk_HeightTwo_ChecksTwoBlocks verifies that a value
// of 2 causes the function to look at both the immediately preceding block and
// the one before that.  A clean commit at height-1 followed by a matching
// signature at height-2 must still produce ErrSignatureFoundInPastBlocks.
func TestCheckDoubleSigningRisk_HeightTwo_ChecksTwoBlocks(t *testing.T) {
	const targetHeight = int64(10)

	cs, mockBS := newStateForDoubleSignTest(t, 2)
	valAddr := cs.privValidatorPubKey.Address()

	// height-1 has a commit from a *different* validator → no early return.
	cleanCommit := makeCommitWithDifferentValidator(targetHeight - 1)
	mockBS.On("LoadSeenCommit", targetHeight-int64(1)).Return(cleanCommit)

	// height-2 has the matching signature.
	matchingCommit := makeCommitWithValidator(targetHeight-2, valAddr)
	mockBS.On("LoadSeenCommit", targetHeight-int64(2)).Return(matchingCommit)

	err := cs.checkDoubleSigningRisk(targetHeight)

	require.ErrorIs(t, err, ErrSignatureFoundInPastBlocks,
		"expected ErrSignatureFoundInPastBlocks when matching signature is at height-2 "+
			"with doubleSignCheckHeight=2")

	mockBS.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Test 4 – doubleSignCheckHeight=2, no matching signature → nil (no false positive)
// ---------------------------------------------------------------------------

// TestCheckDoubleSigningRisk_HeightTwo_NoSignatureFound ensures that when none
// of the looked-up commits contain the local validator's address the function
// returns nil (no spurious errors).
func TestCheckDoubleSigningRisk_HeightTwo_NoSignatureFound(t *testing.T) {
	const targetHeight = int64(10)

	cs, mockBS := newStateForDoubleSignTest(t, 2)

	// Both recent blocks were signed by a different validator.
	mockBS.On("LoadSeenCommit", targetHeight-int64(1)).Return(makeCommitWithDifferentValidator(targetHeight - 1))
	mockBS.On("LoadSeenCommit", targetHeight-int64(2)).Return(makeCommitWithDifferentValidator(targetHeight - 2))

	err := cs.checkDoubleSigningRisk(targetHeight)

	require.NoError(t, err,
		"expected nil when no matching signature found with doubleSignCheckHeight=2")

	mockBS.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// Test 5 – edge case: chain shorter than doubleSignCheckHeight
// ---------------------------------------------------------------------------

// TestCheckDoubleSigningRisk_SmallChainHeight verifies that when the current
// height is smaller than doubleSignCheckHeight, the guard
//
//	if doubleSignCheckHeight > height { doubleSignCheckHeight = height }
//
// prevents any access below block 0, and the function returns nil without
// panicking.
func TestCheckDoubleSigningRisk_SmallChainHeight(t *testing.T) {
	// height=1 means there is exactly zero previous blocks to check (height-1 == 0
	// is the genesis, which has no seen commit).  With doubleSignCheckHeight=10
	// the clamp sets effectiveCheck = min(10, 1) = 1, so we look back 1 block.
	// LoadSeenCommit(0) legitimately returns nil.
	const targetHeight = int64(1)

	cs, mockBS := newStateForDoubleSignTest(t, 10)

	// The genesis block (height 0) has no seen commit.
	mockBS.On("LoadSeenCommit", int64(0)).Return((*types.Commit)(nil))

	require.NotPanics(t, func() {
		err := cs.checkDoubleSigningRisk(targetHeight)
		require.NoError(t, err,
			"expected nil when chain is shorter than doubleSignCheckHeight")
	})

	mockBS.AssertExpectations(t)
}
