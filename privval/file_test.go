package privval

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtrand "github.com/cometbft/cometbft/internal/rand"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
)

func TestGenLoadValidator(t *testing.T) {
	privVal, tempKeyFileName, tempStateFileName := newTestFilePV(t, nil) // TODO test with all key types

	height := int64(100)
	privVal.LastSignState.Height = height
	privVal.Save()
	addr := privVal.GetAddress()

	privVal = LoadFilePV(tempKeyFileName, tempStateFileName)
	assert.Equal(t, addr, privVal.GetAddress(), "expected privval addr to be the same")
	assert.Equal(t, height, privVal.LastSignState.Height, "expected privval.LastHeight to have been saved")
}

func TestResetValidator(t *testing.T) {
	privVal, _, tempStateFileName := newTestFilePV(t, nil) // TODO test with all key types
	emptyState := FilePVLastSignState{filePath: tempStateFileName}

	// new priv val has empty state
	assert.Equal(t, privVal.LastSignState, emptyState)

	// test vote
	height, round := int64(10), int32(1)
	voteType := types.PrevoteType
	randBytes := cmtrand.Bytes(tmhash.Size)
	blockID := types.BlockID{Hash: randBytes, PartSetHeader: types.PartSetHeader{}}
	vote := newVote(privVal.Key.Address, height, round, voteType, blockID)
	err := privVal.SignVote("mychainid", vote.ToProto(), false)
	require.NoError(t, err, "expected no error signing vote")

	// priv val after signing is not same as empty
	assert.NotEqual(t, privVal.LastSignState, emptyState)

	// priv val after AcceptNewConnection is same as empty
	privVal.Reset()
	assert.Equal(t, privVal.LastSignState, emptyState)
}

func TestLoadOrGenValidator(t *testing.T) {
	assert := assert.New(t)

	tempKeyFile, err := os.CreateTemp("", "priv_validator_key_")
	require.NoError(t, err)
	tempStateFile, err := os.CreateTemp("", "priv_validator_state_")
	require.NoError(t, err)

	tempKeyFilePath := tempKeyFile.Name()
	err = os.Remove(tempKeyFilePath)
	require.NoError(t, err)
	tempStateFilePath := tempStateFile.Name()
	err = os.Remove(tempStateFilePath)
	require.NoError(t, err)

	privVal, err := LoadOrGenFilePV(tempKeyFilePath, tempStateFilePath, nil) // TODO extend for all key types
	require.NoError(t, err)
	addr := privVal.GetAddress()
	privVal, err = LoadOrGenFilePV(tempKeyFilePath, tempStateFilePath, nil)
	require.NoError(t, err)
	assert.Equal(addr, privVal.GetAddress(), "expected privval addr to be the same")
}

func TestUnmarshalValidatorState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// create some fixed values
	serialized := `{
		"height": "1",
		"round": 1,
		"step": 1
	}`

	val := FilePVLastSignState{}
	err := cmtjson.Unmarshal([]byte(serialized), &val)
	require.NoError(err, "%+v", err)

	// make sure the values match
	assert.EqualValues(val.Height, 1)
	assert.EqualValues(val.Round, 1)
	assert.EqualValues(val.Step, 1)

	// export it and make sure it is the same
	out, err := cmtjson.Marshal(val)
	require.NoError(err, "%+v", err)
	assert.JSONEq(serialized, string(out))
}

func TestUnmarshalValidatorKey(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// create some fixed values
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	pubBytes := pubKey.Bytes()
	privBytes := privKey.Bytes()
	pubB64 := base64.StdEncoding.EncodeToString(pubBytes)
	privB64 := base64.StdEncoding.EncodeToString(privBytes)

	serialized := fmt.Sprintf(`{
  "address": "%s",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "%s"
  },
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "%s"
  }
}`, addr, pubB64, privB64)

	val := FilePVKey{}
	err := cmtjson.Unmarshal([]byte(serialized), &val)
	require.NoError(err, "%+v", err)

	// make sure the values match
	assert.EqualValues(addr, val.Address)
	assert.EqualValues(pubKey, val.PubKey)
	assert.EqualValues(privKey, val.PrivKey)

	// export it and make sure it is the same
	out, err := cmtjson.Marshal(val)
	require.NoError(err, "%+v", err)
	assert.JSONEq(serialized, string(out))
}

func TestSignVote(t *testing.T) {
	assert := assert.New(t)

	privVal, _, _ := newTestFilePV(t, nil) // TODO test with all key types

	randbytes := cmtrand.Bytes(tmhash.Size)
	randbytes2 := cmtrand.Bytes(tmhash.Size)

	block1 := types.BlockID{
		Hash:          randbytes,
		PartSetHeader: types.PartSetHeader{Total: 5, Hash: randbytes},
	}
	block2 := types.BlockID{
		Hash:          randbytes2,
		PartSetHeader: types.PartSetHeader{Total: 10, Hash: randbytes2},
	}

	height, round := int64(10), int32(1)
	voteType := types.PrevoteType

	// sign a vote for first time
	vote := newVote(privVal.Key.Address, height, round, voteType, block1)
	v := vote.ToProto()
	err := privVal.SignVote("mychainid", v, false)
	require.NoError(t, err, "expected no error signing vote")

	// try to sign the same vote again; should be fine
	err = privVal.SignVote("mychainid", v, false)
	require.NoError(t, err, "expected no error on signing same vote")

	// now try some bad votes
	cases := []*types.Vote{
		newVote(privVal.Key.Address, height, round-1, voteType, block1),   // round regression
		newVote(privVal.Key.Address, height-1, round, voteType, block1),   // height regression
		newVote(privVal.Key.Address, height-2, round+4, voteType, block1), // height regression and different round
		newVote(privVal.Key.Address, height, round, voteType, block2),     // different block
	}

	for _, c := range cases {
		cpb := c.ToProto()
		err = privVal.SignVote("mychainid", cpb, false)
		require.Error(t, err, "expected error on signing conflicting vote")
	}

	// try signing a vote with a different time stamp
	sig := vote.Signature
	vote.Timestamp = vote.Timestamp.Add(time.Duration(1000))
	err = privVal.SignVote("mychainid", v, false)
	require.NoError(t, err)
	assert.Equal(sig, vote.Signature)
}

func TestSignProposal(t *testing.T) {
	assert := assert.New(t)

	privVal, _, _ := newTestFilePV(t, nil) // TODO test with all key types

	randbytes := cmtrand.Bytes(tmhash.Size)
	randbytes2 := cmtrand.Bytes(tmhash.Size)

	block1 := types.BlockID{
		Hash:          randbytes,
		PartSetHeader: types.PartSetHeader{Total: 5, Hash: randbytes},
	}
	block2 := types.BlockID{
		Hash:          randbytes2,
		PartSetHeader: types.PartSetHeader{Total: 10, Hash: randbytes2},
	}
	height, round := int64(10), int32(1)

	// sign a proposal for first time
	proposal := newProposal(height, round, block1)
	pbp := proposal.ToProto()
	err := privVal.SignProposal("mychainid", pbp)
	require.NoError(t, err, "expected no error signing proposal")

	// try to sign the same proposal again; should be fine
	err = privVal.SignProposal("mychainid", pbp)
	require.NoError(t, err, "expected no error on signing same proposal")

	// now try some bad Proposals
	cases := []*types.Proposal{
		newProposal(height, round-1, block1),   // round regression
		newProposal(height-1, round, block1),   // height regression
		newProposal(height-2, round+4, block1), // height regression and different round
		newProposal(height, round, block2),     // different block
	}

	for _, c := range cases {
		err = privVal.SignProposal("mychainid", c.ToProto())
		require.Error(t, err, "expected error on signing conflicting proposal")
	}

	// try signing a proposal with a different time stamp
	sig := proposal.Signature
	proposal.Timestamp = proposal.Timestamp.Add(time.Duration(1000))
	err = privVal.SignProposal("mychainid", pbp)
	require.NoError(t, err)
	assert.Equal(sig, proposal.Signature)
}

func TestSignBytes(t *testing.T) {
	privVal, _, _ := newTestFilePV(t, nil) // TODO test with all key types
	testBytes := []byte("test bytes for signing")

	// Sign the test bytes
	sig, err := privVal.SignBytes(testBytes)
	require.NoError(t, err, "expected no error signing bytes")

	// Verify the signature
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err, "expected no error getting public key")
	assert.True(t, pubKey.VerifySignature(testBytes, sig), "signature verification failed")
}

func TestDifferByTimestamp(t *testing.T) {
	tempKeyFile, err := os.CreateTemp("", "priv_validator_key_")
	require.NoError(t, err)
	tempStateFile, err := os.CreateTemp("", "priv_validator_state_")
	require.NoError(t, err)

	privVal, err := GenFilePV(tempKeyFile.Name(), tempStateFile.Name(), nil)
	require.NoError(t, err)
	randbytes := cmtrand.Bytes(tmhash.Size)
	block1 := types.BlockID{Hash: randbytes, PartSetHeader: types.PartSetHeader{Total: 5, Hash: randbytes}}
	height, round := int64(10), int32(1)
	chainID := "mychainid"

	// test proposal
	{
		proposal := newProposal(height, round, block1)
		pb := proposal.ToProto()
		err := privVal.SignProposal(chainID, pb)
		require.NoError(t, err, "expected no error signing proposal")
		signBytes := types.ProposalSignBytes(chainID, pb)

		sig := proposal.Signature
		timeStamp := proposal.Timestamp

		// manipulate the timestamp. should get changed back
		pb.Timestamp = pb.Timestamp.Add(time.Millisecond)
		var emptySig []byte
		proposal.Signature = emptySig
		err = privVal.SignProposal("mychainid", pb)
		require.NoError(t, err, "expected no error on signing same proposal")

		assert.Equal(t, timeStamp, pb.Timestamp)
		assert.Equal(t, signBytes, types.ProposalSignBytes(chainID, pb))
		assert.Equal(t, sig, proposal.Signature)
	}

	// test vote
	{
		voteType := types.PrevoteType
		blockID := types.BlockID{Hash: randbytes, PartSetHeader: types.PartSetHeader{}}
		vote := newVote(privVal.Key.Address, height, round, voteType, blockID)
		v := vote.ToProto()
		err := privVal.SignVote("mychainid", v, false)
		require.NoError(t, err, "expected no error signing vote")

		signBytes := types.VoteSignBytes(chainID, v)
		sig := v.Signature
		extSig := v.ExtensionSignature
		timeStamp := vote.Timestamp

		// manipulate the timestamp. should get changed back
		v.Timestamp = v.Timestamp.Add(time.Millisecond)
		var emptySig []byte
		v.Signature = emptySig
		v.ExtensionSignature = emptySig
		err = privVal.SignVote("mychainid", v, false)
		require.NoError(t, err, "expected no error on signing same vote")

		assert.Equal(t, timeStamp, v.Timestamp)
		assert.Equal(t, signBytes, types.VoteSignBytes(chainID, v))
		assert.Equal(t, sig, v.Signature)
		assert.Equal(t, extSig, v.ExtensionSignature)
	}
}

func TestVoteExtensionsAreSignedIfSignExtensionIsTrue(t *testing.T) {
	privVal, _, _ := newTestFilePV(t, nil)
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	block := types.BlockID{
		Hash:          cmtrand.Bytes(tmhash.Size),
		PartSetHeader: types.PartSetHeader{Total: 5, Hash: cmtrand.Bytes(tmhash.Size)},
	}

	height, round := int64(10), int32(1)
	voteType := types.PrecommitType

	// We initially sign this vote without an extension
	vote1 := newVote(privVal.Key.Address, height, round, voteType, block)
	vpb1 := vote1.ToProto()

	err = privVal.SignVote("mychainid", vpb1, true)
	require.NoError(t, err, "expected no error signing vote")
	assert.NotNil(t, vpb1.ExtensionSignature)

	vesb1 := types.VoteExtensionSignBytes("mychainid", vpb1)
	assert.True(t, pubKey.VerifySignature(vesb1, vpb1.ExtensionSignature))

	// We duplicate this vote precisely, including its timestamp, but change
	// its extension
	vote2 := vote1.Copy()
	vote2.Extension = []byte("new extension")
	vpb2 := vote2.ToProto()

	err = privVal.SignVote("mychainid", vpb2, true)
	require.NoError(t, err, "expected no error signing same vote with manipulated vote extension")

	// We need to ensure that a valid new extension signature has been created
	// that validates against the vote extension sign bytes with the new
	// extension, and does not validate against the vote extension sign bytes
	// with the old extension.
	vesb2 := types.VoteExtensionSignBytes("mychainid", vpb2)
	assert.True(t, pubKey.VerifySignature(vesb2, vpb2.ExtensionSignature))
	assert.False(t, pubKey.VerifySignature(vesb1, vpb2.ExtensionSignature))

	// We now manipulate the timestamp of the vote with the extension, as per
	// TestDifferByTimestamp
	expectedTimestamp := vpb2.Timestamp

	vpb2.Timestamp = vpb2.Timestamp.Add(time.Millisecond)
	vpb2.Signature = nil
	vpb2.ExtensionSignature = nil

	err = privVal.SignVote("mychainid", vpb2, true)
	require.NoError(t, err, "expected no error signing same vote with manipulated timestamp and vote extension")
	assert.Equal(t, expectedTimestamp, vpb2.Timestamp)

	vesb3 := types.VoteExtensionSignBytes("mychainid", vpb2)
	assert.True(t, pubKey.VerifySignature(vesb3, vpb2.ExtensionSignature))
	assert.False(t, pubKey.VerifySignature(vesb1, vpb2.ExtensionSignature))
}

func TestVoteExtensionsAreNotSignedIfSignExtensionIsFalse(t *testing.T) {
	privVal, _, _ := newTestFilePV(t, nil)

	block := types.BlockID{
		Hash:          cmtrand.Bytes(tmhash.Size),
		PartSetHeader: types.PartSetHeader{Total: 5, Hash: cmtrand.Bytes(tmhash.Size)},
	}

	height, round := int64(10), int32(1)
	voteType := types.PrecommitType

	// We initially sign this vote without an extension
	vote1 := newVote(privVal.Key.Address, height, round, voteType, block)
	vpb1 := vote1.ToProto()

	err := privVal.SignVote("mychainid", vpb1, false)
	require.NoError(t, err, "expected no error signing vote")
	assert.Nil(t, vpb1.ExtensionSignature)
}

func newVote(addr types.Address, height int64, round int32,
	typ types.SignedMsgType, blockID types.BlockID,
) *types.Vote {
	return &types.Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   0,
		Height:           height,
		Round:            round,
		Type:             typ,
		Timestamp:        cmttime.Now(),
		BlockID:          blockID,
	}
}

func newProposal(height int64, round int32, blockID types.BlockID) *types.Proposal {
	return &types.Proposal{
		Height:    height,
		Round:     round,
		BlockID:   blockID,
		Timestamp: cmttime.Now(),
	}
}

func newTestFilePV(t *testing.T, keyGenF func() (crypto.PrivKey, error)) (*FilePV, string, string) {
	t.Helper()
	tempKeyFile, err := os.CreateTemp(t.TempDir(), "priv_validator_key_")
	require.NoError(t, err)
	tempStateFile, err := os.CreateTemp(t.TempDir(), "priv_validator_state_")
	require.NoError(t, err)

	privVal, err := GenFilePV(tempKeyFile.Name(), tempStateFile.Name(), keyGenF)
	require.NoError(t, err)

	return privVal, tempKeyFile.Name(), tempStateFile.Name()
}
