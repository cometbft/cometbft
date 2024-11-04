package types

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptomocks "github.com/cometbft/cometbft/crypto/mocks"

	cmtmath "github.com/cometbft/cometbft/libs/math"
	cmttime "github.com/cometbft/cometbft/types/time"
)

// Check VerifyCommit, VerifyCommitLight and VerifyCommitLightTrusting basic
// verification.
func TestValidatorSet_VerifyCommit_All(t *testing.T) {
	var (
		round  = int32(0)
		height = int64(100)

		blockID    = makeBlockID([]byte("blockhash"), 1000, []byte("partshash"))
		chainID    = "Lalande21185"
		trustLevel = cmtmath.Fraction{Numerator: 2, Denominator: 3}
	)

	testCases := []struct {
		description, description2 string // description2, if not empty, is checked against VerifyCommitLightTrusting
		// vote chainID
		chainID string
		// vote blockID
		blockID BlockID
		valSize int

		// height of the commit
		height int64

		// votes
		blockVotes  int
		nilVotes    int
		absentVotes int

		expErr bool
	}{
		{"good (batch verification)", "", chainID, blockID, 3, height, 3, 0, 0, false},
		{"good (single verification)", "", chainID, blockID, 1, height, 1, 0, 0, false},

		{"wrong signature (#0)", "", "EpsilonEridani", blockID, 2, height, 2, 0, 0, true},
		{"wrong block ID", "", chainID, makeBlockIDRandom(), 2, height, 2, 0, 0, true},
		{"wrong height", "", chainID, blockID, 1, height - 1, 1, 0, 0, true},

		{"wrong set size: 4 vs 3", "", chainID, blockID, 4, height, 3, 0, 0, true},
		{"wrong set size: 1 vs 2", "double vote from Validator", chainID, blockID, 1, height, 2, 0, 0, true},

		{"insufficient voting power: got 30, needed more than 66", "", chainID, blockID, 10, height, 3, 2, 5, true},
		{"insufficient voting power: got 0, needed more than 6", "", chainID, blockID, 1, height, 0, 0, 1, true}, // absent
		{"insufficient voting power: got 0, needed more than 6", "", chainID, blockID, 1, height, 0, 1, 0, true}, // nil
		{"insufficient voting power: got 60, needed more than 60", "", chainID, blockID, 9, height, 6, 3, 0, true},
	}

	for _, tc := range testCases {
		countAllSignatures := false
		f := func(t *testing.T) {
			t.Helper()
			_, valSet, vals := randVoteSet(tc.height, round, PrecommitType, tc.valSize, 10, false)
			totalVotes := tc.blockVotes + tc.absentVotes + tc.nilVotes
			sigs := make([]CommitSig, totalVotes)
			vi := 0
			// add absent sigs first
			for i := 0; i < tc.absentVotes; i++ {
				sigs[vi] = NewCommitSigAbsent()
				vi++
			}
			for i := 0; i < tc.blockVotes+tc.nilVotes; i++ {
				pubKey, err := vals[vi%len(vals)].GetPubKey()
				require.NoError(t, err)
				vote := &Vote{
					ValidatorAddress: pubKey.Address(),
					ValidatorIndex:   int32(vi),
					Height:           tc.height,
					Round:            round,
					Type:             PrecommitType,
					BlockID:          tc.blockID,
					Timestamp:        cmttime.Now(),
				}
				if i >= tc.blockVotes {
					vote.BlockID = BlockID{}
				}

				v := vote.ToProto()

				require.NoError(t, vals[vi%len(vals)].SignVote(tc.chainID, v, false))
				vote.Signature = v.Signature

				sigs[vi] = vote.CommitSig()

				vi++
			}
			commit := &Commit{
				Height:     tc.height,
				Round:      round,
				BlockID:    tc.blockID,
				Signatures: sigs,
			}

			err := valSet.VerifyCommit(chainID, blockID, height, commit)
			if tc.expErr {
				if assert.Error(t, err, "VerifyCommit") { //nolint:testifylint // require.Error doesn't work with the conditional here
					assert.Contains(t, err.Error(), tc.description, "VerifyCommit")
				}
			} else {
				require.NoError(t, err, "VerifyCommit")
			}

			if countAllSignatures {
				err = valSet.VerifyCommitLightAllSignatures(chainID, blockID, height, commit)
			} else {
				err = valSet.VerifyCommitLight(chainID, blockID, height, commit)
			}
			if tc.expErr {
				if assert.Error(t, err, "VerifyCommitLight") { //nolint:testifylint // require.Error doesn't work with the conditional here
					assert.Contains(t, err.Error(), tc.description, "VerifyCommitLight")
				}
			} else {
				require.NoError(t, err, "VerifyCommitLight")
			}

			// only a subsection of the tests apply to VerifyCommitLightTrusting
			expErr := tc.expErr
			if (!countAllSignatures && totalVotes != tc.valSize) || totalVotes < tc.valSize || !tc.blockID.Equals(blockID) || tc.height != height {
				expErr = false
			}
			if countAllSignatures {
				err = valSet.VerifyCommitLightTrustingAllSignatures(chainID, commit, trustLevel)
			} else {
				err = valSet.VerifyCommitLightTrusting(chainID, commit, trustLevel)
			}
			if expErr {
				if assert.Error(t, err, "VerifyCommitLightTrusting") { //nolint:testifylint // require.Error doesn't work with the conditional here
					errStr := tc.description2
					if len(errStr) == 0 {
						errStr = tc.description
					}
					assert.Contains(t, err.Error(), errStr, "VerifyCommitLightTrusting")
				}
			} else {
				require.NoError(t, err, "VerifyCommitLightTrusting")
			}
		}
		t.Run(tc.description+"/"+strconv.FormatBool(countAllSignatures), f)
		countAllSignatures = true
		t.Run(tc.description+"/"+strconv.FormatBool(countAllSignatures), f)
	}
}

func TestValidatorSet_VerifyCommit_CheckAllSignatures(t *testing.T) {
	var (
		chainID = "test_chain_id"
		h       = int64(3)
		blockID = makeBlockIDRandom()
	)

	voteSet, valSet, vals := randVoteSet(h, 0, PrecommitType, 4, 10, false)
	extCommit, err := MakeExtCommit(blockID, h, 0, voteSet, vals, cmttime.Now(), false)
	require.NoError(t, err)
	commit := extCommit.ToCommit()
	require.NoError(t, valSet.VerifyCommit(chainID, blockID, h, commit))

	// malleate 4th signature
	vote := voteSet.GetByIndex(3)
	v := vote.ToProto()
	err = vals[3].SignVote("CentaurusA", v, true)
	require.NoError(t, err)
	vote.Signature = v.Signature
	vote.ExtensionSignature = v.ExtensionSignature
	commit.Signatures[3] = vote.CommitSig()

	err = valSet.VerifyCommit(chainID, blockID, h, commit)
	if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
		assert.Contains(t, err.Error(), "wrong signature (#3)")
	}
}

func TestValidatorSet_VerifyCommitLight_ReturnsAsSoonAsMajOfVotingPowerSignedIffNotAllSigs(t *testing.T) {
	var (
		chainID = "test_chain_id"
		h       = int64(3)
		blockID = makeBlockIDRandom()
	)

	voteSet, valSet, vals := randVoteSet(h, 0, PrecommitType, 4, 10, false)
	extCommit, err := MakeExtCommit(blockID, h, 0, voteSet, vals, cmttime.Now(), false)
	require.NoError(t, err)
	commit := extCommit.ToCommit()
	require.NoError(t, valSet.VerifyCommit(chainID, blockID, h, commit))

	err = valSet.VerifyCommitLightAllSignatures(chainID, blockID, h, commit)
	require.NoError(t, err)

	// malleate 4th signature (3 signatures are enough for 2/3+)
	vote := voteSet.GetByIndex(3)
	v := vote.ToProto()
	err = vals[3].SignVote("CentaurusA", v, true)
	require.NoError(t, err)
	vote.Signature = v.Signature
	vote.ExtensionSignature = v.ExtensionSignature
	commit.Signatures[3] = vote.CommitSig()

	err = valSet.VerifyCommitLight(chainID, blockID, h, commit)
	require.NoError(t, err)
	err = valSet.VerifyCommitLightAllSignatures(chainID, blockID, h, commit)
	require.Error(t, err) // counting all signatures detects the malleated signature
}

func TestValidatorSet_VerifyCommitLightTrusting_ReturnsAsSoonAsTrustLevelSignedIffNotAllSigs(t *testing.T) {
	var (
		chainID = "test_chain_id"
		h       = int64(3)
		blockID = makeBlockIDRandom()
	)

	voteSet, valSet, vals := randVoteSet(h, 0, PrecommitType, 4, 10, false)
	extCommit, err := MakeExtCommit(blockID, h, 0, voteSet, vals, cmttime.Now(), false)
	require.NoError(t, err)
	commit := extCommit.ToCommit()
	require.NoError(t, valSet.VerifyCommit(chainID, blockID, h, commit))

	err = valSet.VerifyCommitLightTrustingAllSignatures(
		chainID,
		commit,
		cmtmath.Fraction{Numerator: 1, Denominator: 3},
	)
	require.NoError(t, err)

	// malleate 3rd signature (2 signatures are enough for 1/3+ trust level)
	vote := voteSet.GetByIndex(2)
	v := vote.ToProto()
	err = vals[2].SignVote("CentaurusA", v, true)
	require.NoError(t, err)
	vote.Signature = v.Signature
	vote.ExtensionSignature = v.ExtensionSignature
	commit.Signatures[2] = vote.CommitSig()

	err = valSet.VerifyCommitLightTrusting(chainID, commit, cmtmath.Fraction{Numerator: 1, Denominator: 3})
	require.NoError(t, err)
	err = valSet.VerifyCommitLightTrustingAllSignatures(
		chainID,
		commit,
		cmtmath.Fraction{Numerator: 1, Denominator: 3},
	)
	require.Error(t, err) // counting all signatures detects the malleated signature
}

func TestValidatorSet_VerifyCommitLightTrusting(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
		newValSet, _                  = RandValidatorSet(2, 1)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	testCases := []struct {
		valSet *ValidatorSet
		err    bool
	}{
		// good
		0: {
			valSet: originalValset,
			err:    false,
		},
		// bad - no overlap between validator sets
		1: {
			valSet: newValSet,
			err:    true,
		},
		// good - first two are different but the rest of the same -> >1/3
		2: {
			valSet: NewValidatorSet(append(newValSet.Validators, originalValset.Validators...)),
			err:    false,
		},
	}

	for _, tc := range testCases {
		err = tc.valSet.VerifyCommitLightTrusting("test_chain_id", commit,
			cmtmath.Fraction{Numerator: 1, Denominator: 3})
		if tc.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestValidatorSet_VerifyCommitLightTrustingWithCache_UpdatesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
		newValSet, _                  = RandValidatorSet(2, 1)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	valSet := NewValidatorSet(append(originalValset.Validators, newValSet.Validators...))
	cache := NewSignatureCache()
	err = valSet.VerifyCommitLightTrustingWithCache("test_chain_id", commit, cmtmath.Fraction{Numerator: 1, Denominator: 3}, cache)
	require.NoError(t, err)
	require.Equal(t, 3, cache.Len()) // 8 validators, getting to 1/3 takes 3 signatures

	cacheVal, ok := cache.Get(string(commit.Signatures[0].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[0].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 0), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[1].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[1].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 1), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[2].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[2].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 2), cacheVal.VoteSignBytes)
}

func TestValidatorSet_VerifyCommitLightTrustingWithCache_UsesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
		newValSet, _                  = RandValidatorSet(2, 1)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	valSet := NewValidatorSet(append(newValSet.Validators, originalValset.Validators...))

	cache := NewSignatureCache()
	cache.Add(string(commit.Signatures[0].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: valSet.Validators[0].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 0),
	})
	cache.Add(string(commit.Signatures[1].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: valSet.Validators[1].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 1),
	})
	cache.Add(string(commit.Signatures[2].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: valSet.Validators[2].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 2),
	})

	err = valSet.VerifyCommitLightTrustingWithCache("test_chain_id", commit, cmtmath.Fraction{Numerator: 1, Denominator: 3}, cache)
	require.NoError(t, err)
	require.Equal(t, 3, cache.Len()) // no new signature checks, so no new cache entries
}

func TestValidatorSet_VerifyCommitLightWithCache_UpdatesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	cache := NewSignatureCache()
	err = originalValset.VerifyCommitLightWithCache("test_chain_id", blockID, 1, commit, cache)
	require.NoError(t, err)

	require.Equal(t, 5, cache.Len()) // 6 validators, getting to 2/3 takes 5 signatures

	cacheVal, ok := cache.Get(string(commit.Signatures[0].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[0].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 0), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[1].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[1].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 1), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[2].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[2].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 2), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[3].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[3].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 3), cacheVal.VoteSignBytes)

	cacheVal, ok = cache.Get(string(commit.Signatures[4].Signature))
	require.True(t, ok)
	require.Equal(t, originalValset.Validators[4].PubKey.Bytes(), cacheVal.ValidatorPubKeyBytes)
	require.Equal(t, commit.VoteSignBytes("test_chain_id", 4), cacheVal.VoteSignBytes)
}

func TestValidatorSet_VerifyCommitLightWithCache_UsesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	cache := NewSignatureCache()
	cache.Add(string(commit.Signatures[0].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[0].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 0),
	})
	cache.Add(string(commit.Signatures[1].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[1].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 1),
	})
	cache.Add(string(commit.Signatures[2].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[2].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 2),
	})
	cache.Add(string(commit.Signatures[3].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[3].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 3),
	})
	cache.Add(string(commit.Signatures[4].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[4].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 4),
	})

	err = originalValset.VerifyCommitLightWithCache("test_chain_id", blockID, 1, commit, cache)
	require.NoError(t, err)
	require.Equal(t, 5, cache.Len()) // no new signature checks, so no new cache entries
}

func TestValidatorSet_VerifyCommitLightTrustingErrorsOnOverflow(t *testing.T) {
	var (
		blockID               = makeBlockIDRandom()
		voteSet, valSet, vals = randVoteSet(1, 1, PrecommitType, 1, MaxTotalVotingPower, false)
		extCommit, err        = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
	)
	require.NoError(t, err)

	err = valSet.VerifyCommitLightTrusting("test_chain_id", extCommit.ToCommit(),
		cmtmath.Fraction{Numerator: 25, Denominator: 55})
	if assert.Error(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
		assert.Contains(t, err.Error(), "int64 overflow")
	}
}

func TestValidation_verifyCommitBatch_UsesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	cache := NewSignatureCache()
	cache.Add(string(commit.Signatures[0].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[0].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 0),
	})
	cache.Add(string(commit.Signatures[1].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[1].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 1),
	})
	cache.Add(string(commit.Signatures[2].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[2].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 2),
	})
	cache.Add(string(commit.Signatures[3].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[3].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 3),
	})
	cache.Add(string(commit.Signatures[4].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[4].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 4),
	})

	// ignore all commit signatures that are not for the block
	ignore := func(c CommitSig) bool { return c.BlockIDFlag != BlockIDFlagCommit }

	// count all the remaining signatures
	count := func(_ CommitSig) bool { return true }

	bv := cryptomocks.NewBatchVerifier(t)

	err = verifyCommitBatch("test_chain_id", originalValset, commit, 4, ignore, count, false, true, bv, cache)
	require.NoError(t, err)
	bv.AssertNotCalled(t, "Add")
	bv.AssertNotCalled(t, "Verify")
}
func TestValidation_verifyCommitSingle_UsesCache(t *testing.T) {
	var (
		blockID                       = makeBlockIDRandom()
		voteSet, originalValset, vals = randVoteSet(1, 1, PrecommitType, 6, 1, false)
		extCommit, err                = MakeExtCommit(blockID, 1, 1, voteSet, vals, cmttime.Now(), false)
	)
	require.NoError(t, err)
	commit := extCommit.ToCommit()

	cache := NewSignatureCache()
	cache.Add(string(commit.Signatures[0].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[0].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 0),
	})
	cache.Add(string(commit.Signatures[1].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[1].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 1),
	})
	cache.Add(string(commit.Signatures[2].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[2].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 2),
	})
	cache.Add(string(commit.Signatures[3].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[3].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 3),
	})
	cache.Add(string(commit.Signatures[4].Signature), SignatureCacheValue{
		ValidatorPubKeyBytes: originalValset.Validators[4].PubKey.Bytes(),
		VoteSignBytes:        commit.VoteSignBytes("test_chain_id", 4),
	})

	// ignore all commit signatures that are not for the block
	ignore := func(c CommitSig) bool { return c.BlockIDFlag != BlockIDFlagCommit }

	// count all the remaining signatures
	count := func(_ CommitSig) bool { return true }

	mockValPubkeys := []*cryptomocks.PubKey{
		cryptomocks.NewPubKey(t),
		cryptomocks.NewPubKey(t),
		cryptomocks.NewPubKey(t),
		cryptomocks.NewPubKey(t),
		cryptomocks.NewPubKey(t),
	}

	mockValPubkeys[0].On("Bytes").Return(originalValset.Validators[0].PubKey.Bytes())
	mockValPubkeys[1].On("Bytes").Return(originalValset.Validators[1].PubKey.Bytes())
	mockValPubkeys[2].On("Bytes").Return(originalValset.Validators[2].PubKey.Bytes())
	mockValPubkeys[3].On("Bytes").Return(originalValset.Validators[3].PubKey.Bytes())
	mockValPubkeys[4].On("Bytes").Return(originalValset.Validators[4].PubKey.Bytes())

	originalValset.Validators[0].PubKey = mockValPubkeys[0]
	originalValset.Validators[1].PubKey = mockValPubkeys[1]
	originalValset.Validators[2].PubKey = mockValPubkeys[2]
	originalValset.Validators[3].PubKey = mockValPubkeys[3]
	originalValset.Validators[4].PubKey = mockValPubkeys[4]

	err = verifyCommitSingle("test_chain_id", originalValset, commit, 4, ignore, count, false, true, cache)
	require.NoError(t, err)

	mockValPubkeys[0].AssertCalled(t, "Bytes")
	mockValPubkeys[1].AssertCalled(t, "Bytes")
	mockValPubkeys[2].AssertCalled(t, "Bytes")
	mockValPubkeys[3].AssertCalled(t, "Bytes")
	mockValPubkeys[4].AssertCalled(t, "Bytes")

	mockValPubkeys[0].AssertNotCalled(t, "VerifySignature")
	mockValPubkeys[1].AssertNotCalled(t, "VerifySignature")
	mockValPubkeys[2].AssertNotCalled(t, "VerifySignature")
	mockValPubkeys[3].AssertNotCalled(t, "VerifySignature")
	mockValPubkeys[4].AssertNotCalled(t, "VerifySignature")
}
