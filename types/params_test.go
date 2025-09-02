package types

import (
	"bytes"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

var (
	valEd25519   = []string{ABCIPubKeyTypeEd25519}
	valSecp256k1 = []string{ABCIPubKeyTypeSecp256k1}
)

func TestConsensusParamsValidation(t *testing.T) {
	testCases := []struct {
		params ConsensusParams
		valid  bool
	}{
		// test block params
		0: {makeParams(1, 0, 2, 0, valEd25519, 0), true},
		1: {makeParams(0, 0, 2, 0, valEd25519, 0), false},
		2: {makeParams(47*1024*1024, 0, 2, 0, valEd25519, 0), true},
		3: {makeParams(10, 0, 2, 0, valEd25519, 0), true},
		4: {makeParams(100*1024*1024, 0, 2, 0, valEd25519, 0), true},
		5: {makeParams(101*1024*1024, 0, 2, 0, valEd25519, 0), false},
		6: {makeParams(1024*1024*1024, 0, 2, 0, valEd25519, 0), false},
		// test evidence params
		7:  {makeParams(1, 0, 0, 0, valEd25519, 0), false},
		8:  {makeParams(1, 0, 2, 2, valEd25519, 0), false},
		9:  {makeParams(1000, 0, 2, 1, valEd25519, 0), true},
		10: {makeParams(1, 0, -1, 0, valEd25519, 0), false},
		// test no pubkey type provided
		11: {makeParams(1, 0, 2, 0, []string{}, 0), false},
		// test invalid pubkey type provided
		12: {makeParams(1, 0, 2, 0, []string{"potatoes make good pubkeys"}, 0), false},
		13: {makeParams(-1, 0, 2, 0, valEd25519, 0), true},
		14: {makeParams(-2, 0, 2, 0, valEd25519, 0), false},
	}
	for i, tc := range testCases {
		if tc.valid {
			assert.NoErrorf(t, tc.params.ValidateBasic(), "expected no error for valid params (#%d)", i)
		} else {
			assert.Errorf(t, tc.params.ValidateBasic(), "expected error for non valid params (#%d)", i)
		}
	}
}

func makeParams(
	blockBytes, blockGas int64,
	evidenceAge int64,
	maxEvidenceBytes int64,
	pubkeyTypes []string,
	abciExtensionHeight int64,
) ConsensusParams {
	return ConsensusParams{
		Block: BlockParams{
			MaxBytes: blockBytes,
			MaxGas:   blockGas,
		},
		Evidence: EvidenceParams{
			MaxAgeNumBlocks: evidenceAge,
			MaxAgeDuration:  time.Duration(evidenceAge),
			MaxBytes:        maxEvidenceBytes,
		},
		Validator: ValidatorParams{
			PubKeyTypes: pubkeyTypes,
		},
		ABCI: ABCIParams{
			VoteExtensionsEnableHeight: abciExtensionHeight,
		},
	}
}

func TestConsensusParamsHash(t *testing.T) {
	params := []ConsensusParams{
		makeParams(4, 2, 3, 1, valEd25519, 0),
		makeParams(1, 4, 3, 1, valEd25519, 0),
		makeParams(1, 2, 4, 1, valEd25519, 0),
		makeParams(2, 5, 7, 1, valEd25519, 0),
		makeParams(1, 7, 6, 1, valEd25519, 0),
		makeParams(9, 5, 4, 1, valEd25519, 0),
		makeParams(7, 8, 9, 1, valEd25519, 0),
		makeParams(4, 6, 5, 1, valEd25519, 0),
	}

	hashes := make([][]byte, len(params))
	for i := range params {
		hashes[i] = params[i].Hash()
	}

	// make sure there are no duplicates...
	// sort, then check in order for matches
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) < 0
	})
	for i := 0; i < len(hashes)-1; i++ {
		assert.NotEqual(t, hashes[i], hashes[i+1])
	}
}

func TestConsensusParamsUpdate(t *testing.T) {
	testCases := []struct {
		params        ConsensusParams
		updates       *cmtproto.ConsensusParams
		updatedParams ConsensusParams
	}{
		// empty updates
		{
			makeParams(1, 2, 3, 0, valEd25519, 0),
			&cmtproto.ConsensusParams{},
			makeParams(1, 2, 3, 0, valEd25519, 0),
		},
		// fine updates
		{
			makeParams(1, 2, 3, 0, valEd25519, 0),
			&cmtproto.ConsensusParams{
				Block: &cmtproto.BlockParams{
					MaxBytes: 100,
					MaxGas:   200,
				},
				Evidence: &cmtproto.EvidenceParams{
					MaxAgeNumBlocks: 300,
					MaxAgeDuration:  time.Duration(300),
					MaxBytes:        50,
				},
				Validator: &cmtproto.ValidatorParams{
					PubKeyTypes: valSecp256k1,
				},
			},
			makeParams(100, 200, 300, 50, valSecp256k1, 0),
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.updatedParams, tc.params.Update(tc.updates))
	}
}

func TestConsensusParamsUpdate_AppVersion(t *testing.T) {
	params := makeParams(1, 2, 3, 0, valEd25519, 0)

	assert.EqualValues(t, 0, params.Version.App)

	updated := params.Update(
		&cmtproto.ConsensusParams{Version: &cmtproto.VersionParams{App: 1}})

	assert.EqualValues(t, 1, updated.Version.App)
}

func TestConsensusParamsUpdate_VoteExtensionsEnableHeight(t *testing.T) {
	const nilTest = -10000000
	testCases := []struct {
		name        string
		current     int64
		from        int64
		to          int64
		expectedErr bool
	}{
		// no change
		{"current: 3, 0 -> 0", 3, 0, 0, false},
		{"current: 3, 100 -> 100, ", 3, 100, 100, false},
		{"current: 100, 100 -> 100, ", 100, 100, 100, false},
		{"current: 300, 100 -> 100, ", 300, 100, 100, false},
		// set for the first time
		{"current: 3, 0 -> 5, ", 3, 0, 5, false},
		{"current: 4, 0 -> 5, ", 4, 0, 5, false},
		{"current: 5, 0 -> 5, ", 5, 0, 5, true},
		{"current: 6, 0 -> 5, ", 6, 0, 5, true},
		{"current: 50, 0 -> 5, ", 50, 0, 5, true},
		// reset to 0
		{"current: 4, 5 -> 0, ", 4, 5, 0, false},
		{"current: 5, 5 -> 0, ", 5, 5, 0, true},
		{"current: 6, 5 -> 0, ", 6, 5, 0, true},
		{"current: 10, 5 -> 0, ", 10, 5, 0, true},
		// modify backwards
		{"current: 1, 10 -> 5, ", 1, 10, 5, false},
		{"current: 4, 10 -> 5, ", 4, 10, 5, false},
		{"current: 5, 10 -> 5, ", 5, 10, 5, true},
		{"current: 6, 10 -> 5, ", 6, 10, 5, true},
		{"current: 9, 10 -> 5, ", 9, 10, 5, true},
		{"current: 10, 10 -> 5, ", 10, 10, 5, true},
		{"current: 11, 10 -> 5, ", 11, 10, 5, true},
		{"current: 100, 10 -> 5, ", 100, 10, 5, true},
		// modify forward
		{"current: 3, 10 -> 15, ", 3, 10, 15, false},
		{"current: 9, 10 -> 15, ", 9, 10, 15, false},
		{"current: 10, 10 -> 15, ", 10, 10, 15, true},
		{"current: 11, 10 -> 15, ", 11, 10, 15, true},
		{"current: 14, 10 -> 15, ", 14, 10, 15, true},
		{"current: 15, 10 -> 15, ", 15, 10, 15, true},
		{"current: 16, 10 -> 15, ", 16, 10, 15, true},
		{"current: 100, 10 -> 15, ", 100, 10, 15, true},
		// negative values
		{"current: 3, 0 -> -5", 3, 0, -5, true},
		{"current: 3, -5 -> 100, ", 3, -5, 100, false},
		{"current: 3, -10 -> 3, ", 3, -10, 3, true},
		{"current: 3, -3 -> -3", 3, -3, -3, true},
		{"current: 100, -8 -> -9, ", 100, -8, -9, true},
		{"current: 300, -10 -> -8, ", 300, -10, -8, true},
		// test for nil
		{"current: 300, 400 -> nil, ", 300, 400, nilTest, false},
		{"current: 300, 200 -> nil, ", 300, 200, nilTest, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			initialParams := makeParams(1, 0, 2, 0, valEd25519, tc.from)
			update := &cmtproto.ConsensusParams{}
			if tc.to == nilTest {
				update.Abci = nil
			} else {
				update.Abci = &cmtproto.ABCIParams{
					VoteExtensionsEnableHeight: tc.to,
				}
			}
			if tc.expectedErr {
				require.Error(t, initialParams.ValidateUpdate(update, tc.current))
			} else {
				require.NoError(t, initialParams.ValidateUpdate(update, tc.current))
			}
		})
	}
}

func TestProto(t *testing.T) {
	params := []ConsensusParams{
		makeParams(4, 2, 3, 1, valEd25519, 1),
		makeParams(1, 4, 3, 1, valEd25519, 1),
		makeParams(1, 2, 4, 1, valEd25519, 1),
		makeParams(2, 5, 7, 1, valEd25519, 1),
		makeParams(1, 7, 6, 1, valEd25519, 1),
		makeParams(9, 5, 4, 1, valEd25519, 1),
		makeParams(7, 8, 9, 1, valEd25519, 1),
		makeParams(4, 6, 5, 1, valEd25519, 1),
	}

	for i := range params {
		pbParams := params[i].ToProto()

		oriParams := ConsensusParamsFromProto(pbParams)

		assert.Equal(t, params[i], oriParams)

	}
}
