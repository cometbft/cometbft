package types

import (
	"bytes"
	"sort"
	"testing"
	"time"

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
	abciExtensionHeight int64
	precision           time.Duration
	messageDelay        time.Duration
}

func makeParams(args makeParamsArgs) ConsensusParams {
	if args.pubkeyTypes == nil {
		args.pubkeyTypes = valEd25519
	}
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
		ABCI: ABCIParams{
			VoteExtensionsEnableHeight: args.abciExtensionHeight,
		},
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

func TestConsensusParamsUpdate_VoteExtensionsEnableHeight(t *testing.T) {
	t.Run("set to height but initial height already run", func(*testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 1,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 10,
			},
		}
		require.Error(t, initialParams.ValidateUpdate(update, 1))
		require.Error(t, initialParams.ValidateUpdate(update, 5))
	})
	t.Run("reset to 0", func(t *testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 1,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 0,
			},
		}
		require.Error(t, initialParams.ValidateUpdate(update, 1))
	})
	t.Run("set to height before current height run", func(*testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 100,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 10,
			},
		}
		require.Error(t, initialParams.ValidateUpdate(update, 11))
		require.Error(t, initialParams.ValidateUpdate(update, 99))
	})
	t.Run("set to height after current height run", func(*testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 300,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 99,
			},
		}
		require.NoError(t, initialParams.ValidateUpdate(update, 11))
		require.NoError(t, initialParams.ValidateUpdate(update, 98))
	})
	t.Run("no error when unchanged", func(*testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 100,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 100,
			},
		}
		require.NoError(t, initialParams.ValidateUpdate(update, 500))
	})
	t.Run("updated from 0 to 0", func(t *testing.T) {
		initialParams := makeParams(makeParamsArgs{
			abciExtensionHeight: 0,
		})
		update := &cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 0,
			},
		}
		require.NoError(t, initialParams.ValidateUpdate(update, 100))
	})
}

func TestProto(t *testing.T) {
	params := []ConsensusParams{
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 2, evidenceAge: 3, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 4, evidenceAge: 3, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 2, evidenceAge: 4, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 2, blockGas: 5, evidenceAge: 7, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 1, blockGas: 7, evidenceAge: 6, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 9, blockGas: 5, evidenceAge: 4, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 7, blockGas: 8, evidenceAge: 9, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{blockBytes: 4, blockGas: 6, evidenceAge: 5, maxEvidenceBytes: 1, abciExtensionHeight: 1}),
		makeParams(makeParamsArgs{precision: time.Second, messageDelay: time.Minute}),
		makeParams(makeParamsArgs{precision: time.Nanosecond, messageDelay: time.Millisecond}),
		makeParams(makeParamsArgs{abciExtensionHeight: 100}),
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
