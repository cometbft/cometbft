package types

import (
	"bytes"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
)

var (
	valEd25519   = []string{ABCIPubKeyTypeEd25519}
	valSecp256k1 = []string{ABCIPubKeyTypeSecp256k1}
)

func TestConsensusParamsValidation(t *testing.T) {
	testCases := []struct {
		name   string
		params ConsensusParams
		valid  bool
	}{
		// test block params
		{
			name: "normal values",
			params: makeParams(makeParamsArgs{
				blockBytes:   1,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: true,
		},
		{
			name: "blockBytes se to 0",
			params: makeParams(makeParamsArgs{
				blockBytes:   0,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "blockBytes set to a big valid value",
			params: makeParams(makeParamsArgs{
				blockBytes:   47 * 1024 * 1024,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: true,
		},
		{
			name: "blockBytes set to a small valid value",
			params: makeParams(makeParamsArgs{
				blockBytes:   10,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: true,
		},
		{
			name: "blockBytes set to the biggest valid value",
			params: makeParams(makeParamsArgs{
				blockBytes:   100 * 1024 * 1024,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: true,
		},
		{
			name: "blockBytes, biggest valid value, off-by-1",
			params: makeParams(makeParamsArgs{
				blockBytes:   100*1024*1024 + 1,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "blockBytes, biggest valid value, off-by-1MB",
			params: makeParams(makeParamsArgs{
				blockBytes:   101 * 1024 * 1024,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "blockBytes, value set to 1GB (too big)",
			params: makeParams(makeParamsArgs{
				blockBytes:   1024 * 1024 * 1024,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "blockBytes invalid, evidenceAge invalid",
			params: makeParams(makeParamsArgs{
				blockBytes:   1024 * 1024 * 1024,
				evidenceAge:  -1,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		// test evidence params
		{
			name: "evidenceAge 0",
			params: makeParams(makeParamsArgs{
				blockBytes:       1,
				evidenceAge:      0,
				maxEvidenceBytes: 0,
				precision:        1,
				messageDelay:     1,
			}),
			valid: false,
		},
		{
			name: "evidenceAge negative",
			params: makeParams(makeParamsArgs{
				blockBytes:   1 * 1024 * 1024,
				evidenceAge:  -1,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "maxEvidenceBytes not less than blockBytes",
			params: makeParams(makeParamsArgs{
				blockBytes:       1,
				evidenceAge:      2,
				maxEvidenceBytes: 2,
				precision:        1,
				messageDelay:     1,
			}),
			valid: false,
		},
		{
			name: "maxEvidenceBytes less than blockBytes",
			params: makeParams(makeParamsArgs{
				blockBytes:       1000,
				evidenceAge:      2,
				maxEvidenceBytes: 1,
				precision:        1,
				messageDelay:     1,
			}),
			valid: true,
		},
		{
			name: "maxEvidenceBytes 0",
			params: makeParams(makeParamsArgs{
				blockBytes:       1,
				evidenceAge:      1,
				maxEvidenceBytes: 0,
				precision:        1,
				messageDelay:     1,
			}),
			valid: true,
		},
		// test no pubkey type provided
		{
			name: "empty pubkeyTypes",
			params: makeParams(makeParamsArgs{
				blockBytes:   1,
				evidenceAge:  2,
				pubkeyTypes:  []string{},
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		// test invalid pubkey type provided
		{
			name: "bad pubkeyTypes",
			params: makeParams(makeParamsArgs{
				blockBytes:   1,
				evidenceAge:  2,
				pubkeyTypes:  []string{"potatoes make good pubkeys"},
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		{
			name: "blockBytes -1",
			params: makeParams(makeParamsArgs{
				blockBytes:   -1,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: true,
		},
		{
			name: "blockBytes -2",
			params: makeParams(makeParamsArgs{
				blockBytes:   -2,
				evidenceAge:  2,
				precision:    1,
				messageDelay: 1,
			}),
			valid: false,
		},
		// test invalid pubkey type provided
		{
			name: "messageDelay -1",
			params: makeParams(makeParamsArgs{
				evidenceAge:  2,
				precision:    1,
				messageDelay: -1,
			}),
			valid: false,
		},
		{
			name: "precision -1",
			params: makeParams(makeParamsArgs{
				evidenceAge:  2,
				precision:    -1,
				messageDelay: 1,
			}),
			valid: false,
		},
		// test pbts
		{
			name: "pbts height -1",
			params: makeParams(
				makeParamsArgs{
					blockBytes:   1,
					evidenceAge:  2,
					precision:    1,
					messageDelay: 1,
					pbtsHeight:   -1,
				}),
			valid: false,
		},
		{
			name: "pbts height 0",
			params: makeParams(
				makeParamsArgs{
					blockBytes:   1,
					evidenceAge:  2,
					precision:    1,
					messageDelay: 1,
					pbtsHeight:   0,
				}),
			valid: true,
		},
	}
	for i, tc := range testCases {
		if tc.valid {
			require.NoErrorf(t, tc.params.ValidateBasic(), "expected no error for valid params (#%d)", i)
		} else {
			require.Errorf(t, tc.params.ValidateBasic(), "expected error for non valid params (#%d)", i)
		}
	}
}

type makeParamsArgs struct {
	blockBytes          int64
	blockGas            int64
	evidenceAge         int64
	maxEvidenceBytes    int64
	pubkeyTypes         []string
	voteExtensionHeight int64
	pbtsHeight          int64
	precision           time.Duration
	messageDelay        time.Duration
}

func makeParams(args makeParamsArgs) ConsensusParams {
	if args.pubkeyTypes == nil {
		args.pubkeyTypes = valEd25519
	}
	p := DefaultFeatureParams()
	p.EnableVoteExtensions(args.voteExtensionHeight)
	p.EnablePbts(args.pbtsHeight)

	return ConsensusParams{
		Block: BlockParams{
			MaxBytes: args.blockBytes,
			MaxGas:   args.blockGas,
		},
		Evidence: EvidenceParams{
			MaxAgeNumBlocks: args.evidenceAge,
			MaxAgeDuration:  time.Duration(args.evidenceAge),
			MaxBytes:        args.maxEvidenceBytes,
		},
		Validator: ValidatorParams{
			PubKeyTypes: args.pubkeyTypes,
		},
		Synchrony: SynchronyParams{
			Precision:    args.precision,
			MessageDelay: args.messageDelay,
		},
		Feature: p,
	}
}

func TestConsensusParamsHash(t *testing.T) {
	params := []ConsensusParams{
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 2, evidenceAge: 3, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 4, evidenceAge: 3, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 4, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 2, blockGas: 5, evidenceAge: 7, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 7, evidenceAge: 6, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 9, blockGas: 5, evidenceAge: 4, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 7, blockGas: 8, evidenceAge: 9, maxEvidenceBytes: 1}),
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 6, evidenceAge: 5, maxEvidenceBytes: 1}),
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
		intialParams  ConsensusParams
		updates       *cmtproto.ConsensusParams
		updatedParams ConsensusParams
	}{
		// empty updates
		{
			intialParams:  makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
			updates:       &cmtproto.ConsensusParams{},
			updatedParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
		},
		{
			// update synchrony params
			intialParams: makeParams(makeParamsArgs{evidenceAge: 3, precision: time.Second, messageDelay: 3 * time.Second}),
			updates: &cmtproto.ConsensusParams{
				Synchrony: &cmtproto.SynchronyParams{
					Precision:    durationPtr(time.Second * 2),
					MessageDelay: durationPtr(time.Second * 4),
				},
			},
			updatedParams: makeParams(makeParamsArgs{evidenceAge: 3, precision: 2 * time.Second, messageDelay: 4 * time.Second}),
		},
		// update vote extensions
		{
			intialParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
			updates: &cmtproto.ConsensusParams{
				Feature: &cmtproto.FeatureParams{
					VoteExtensionsEnableHeight: &types.Int64Value{Value: 1},
				},
			},
			updatedParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3, voteExtensionHeight: 1}),
		},
		// update pbts
		{
			intialParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
			updates: &cmtproto.ConsensusParams{
				Feature: &cmtproto.FeatureParams{
					PbtsEnableHeight: &types.Int64Value{Value: 1},
				},
			},
			updatedParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3, pbtsHeight: 1}),
		},
		// update pbts and vote extensions
		{
			intialParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
			updates: &cmtproto.ConsensusParams{
				Feature: &cmtproto.FeatureParams{
					VoteExtensionsEnableHeight: &types.Int64Value{Value: 1},
					PbtsEnableHeight:           &types.Int64Value{Value: 1},
				},
			},
			updatedParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3, voteExtensionHeight: 1, pbtsHeight: 1}),
		},
		// fine updates
		{
			intialParams: makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3}),
			updates: &cmtproto.ConsensusParams{
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
			updatedParams: makeParams(makeParamsArgs{
				blockBytes: 100, blockGas: 200,
				evidenceAge:      300,
				maxEvidenceBytes: 50,
				pubkeyTypes:      valSecp256k1,
			}),
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.updatedParams, tc.intialParams.Update(tc.updates))
	}
}

func TestConsensusParamsUpdate_AppVersion(t *testing.T) {
	params := makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 3})

	assert.EqualValues(t, 0, params.Version.App)

	updated := params.Update(
		&cmtproto.ConsensusParams{Version: &cmtproto.VersionParams{App: 1}})

	assert.EqualValues(t, 1, updated.Version.App)
}

func TestConsensusParamsUpdate_EnableHeight(t *testing.T) {
	const nilTest = -10000000
	testCases := []struct {
		name        string
		current     int64
		from        int64
		to          int64
		expectedErr bool
	}{
		{"no change: 3, 0 -> 0", 3, 0, 0, false},
		{"no change: 3, 100 -> 100, ", 3, 100, 100, false},
		{"no change: 100, 100 -> 100, ", 100, 100, 100, false},
		{"no change: 300, 100 -> 100, ", 300, 100, 100, false},
		{"first time: 4, 0 -> 5, ", 4, 0, 5, false},
		{"first time: 3, 0 -> 5, ", 3, 0, 5, false},
		{"first time: 5, 0 -> 5, ", 5, 0, 5, true},
		{"first time: 6, 0 -> 5, ", 6, 0, 5, true},
		{"first time: 50, 0 -> 5, ", 50, 0, 5, true},
		{"reset to 0: 4, 5 -> 0, ", 4, 5, 0, false},
		{"reset to 0: 5, 5 -> 0, ", 5, 5, 0, true},
		{"reset to 0: 6, 5 -> 0, ", 6, 5, 0, true},
		{"reset to 0: 10, 5 -> 0, ", 10, 5, 0, true},
		{"modify backwards: 1, 10 -> 5, ", 1, 10, 5, false},
		{"modify backwards: 4, 10 -> 5, ", 4, 10, 5, false},
		{"modify backwards: 5, 10 -> 5, ", 5, 10, 5, true},
		{"modify backwards: 6, 10 -> 5, ", 6, 10, 5, true},
		{"modify backwards: 9, 10 -> 5, ", 9, 10, 5, true},
		{"modify backwards: 10, 10 -> 5, ", 10, 10, 5, true},
		{"modify backwards: 11, 10 -> 5, ", 11, 10, 5, true},
		{"modify backwards: 100, 10 -> 5, ", 100, 10, 5, true},
		{"modify forward: 3, 10 -> 15, ", 3, 10, 15, false},
		{"modify forward: 9, 10 -> 15, ", 9, 10, 15, false},
		{"modify forward: 10, 10 -> 15, ", 10, 10, 15, true},
		{"modify forward: 11, 10 -> 15, ", 11, 10, 15, true},
		{"modify forward: 14, 10 -> 15, ", 14, 10, 15, true},
		{"modify forward: 15, 10 -> 15, ", 15, 10, 15, true},
		{"modify forward: 16, 10 -> 15, ", 16, 10, 15, true},
		{"modify forward: 100, 10 -> 15, ", 100, 10, 15, true},
		{"set to negative value: 3, 0 -> -5", 3, 0, -5, true},
		{"set to negative value: 3, -5 -> 100, ", 3, -5, 100, false},
		{"set to negative value: 3, -10 -> 3, ", 3, -10, 3, true},
		{"set to negative value: 3, -3 -> -3", 3, -3, -3, true},
		{"set to negative value: 100, -8 -> -9, ", 100, -8, -9, true},
		{"set to negative value: 300, -10 -> -8, ", 300, -10, -8, true},
		{"nil: 300, 400 -> nil, ", 300, 400, nilTest, false},
		{"nil: 300, 200 -> nil, ", 300, 200, nilTest, false},
	}

	// Test VoteExtensions enabling
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			initialParams := makeParams(makeParamsArgs{
				voteExtensionHeight: tc.from,
			})
			update := &cmtproto.ConsensusParams{Feature: &cmtproto.FeatureParams{}}
			if tc.to == nilTest {
				update.Feature.VoteExtensionsEnableHeight = nil
			} else {
				update.Feature = &cmtproto.FeatureParams{
					VoteExtensionsEnableHeight: &types.Int64Value{Value: tc.to},
				}
			}
			if tc.expectedErr {
				require.Error(t, initialParams.ValidateUpdate(update, tc.current))
			} else {
				require.NoError(t, initialParams.ValidateUpdate(update, tc.current))
			}
		})
	}

	// Test PBTS enabling
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			initialParams := makeParams(makeParamsArgs{
				pbtsHeight: tc.from,
			})
			update := &cmtproto.ConsensusParams{Feature: &cmtproto.FeatureParams{}}
			if tc.to == nilTest {
				update.Feature.PbtsEnableHeight = nil
			} else {
				update.Feature = &cmtproto.FeatureParams{
					PbtsEnableHeight: &types.Int64Value{Value: tc.to},
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
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 2, evidenceAge: 3, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 4, evidenceAge: 3, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 4, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 2, blockGas: 5, evidenceAge: 7, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 7, evidenceAge: 6, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 9, blockGas: 5, evidenceAge: 4, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 7, blockGas: 8, evidenceAge: 9, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 6, evidenceAge: 5, maxEvidenceBytes: 1, voteExtensionHeight: 1}),
		makeParams(makeParamsArgs{precision: time.Second, messageDelay: time.Minute}),
		makeParams(makeParamsArgs{precision: time.Nanosecond, messageDelay: time.Millisecond}),
		makeParams(makeParamsArgs{voteExtensionHeight: 100}),
		makeParams(makeParamsArgs{pbtsHeight: 1}),
	}

	for i := range params {
		pbParams := params[i].ToProto()

		oriParams := ConsensusParamsFromProto(pbParams)

		assert.Equal(t, params[i], oriParams)
	}
}

func durationPtr(t time.Duration) *time.Duration {
	return &t
}
