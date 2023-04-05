package types

import (
	"fmt"
	"os"
	"testing"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/tmhash"
	cmtrand "github.com/tendermint/tendermint/libs/rand"
	cmtproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	cmttime "github.com/tendermint/tendermint/types/time"
)

var config *cfg.Config // NOTE: must be reset for each _test.go file

func TestMain(m *testing.M) {
	config = cfg.ResetTestRoot("consensus_height_vote_set_test")
	code := m.Run()
	os.RemoveAll(config.RootDir)
	os.Exit(code)
}

func TestPeerCatchupRounds(t *testing.T) {
	valSet, privVals := types.RandValidatorSet(10, 1)

	hvs := NewHeightVoteSet(config.ChainID(), 1, valSet)

<<<<<<< HEAD
	vote999_0 := makeVoteHR(t, 1, 0, 999, privVals)
	added, err := hvs.AddVote(vote999_0, "peer1")
=======
	vote999_0 := makeVoteHR(1, 0, 999, privVals)
	added, err := hvs.AddVote(vote999_0, "peer1", true)
>>>>>>> 111d252d7 (Fix lints (#625))
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

<<<<<<< HEAD
	vote1000_0 := makeVoteHR(t, 1, 0, 1000, privVals)
	added, err = hvs.AddVote(vote1000_0, "peer1")
=======
	vote1000_0 := makeVoteHR(1, 0, 1000, privVals)
	added, err = hvs.AddVote(vote1000_0, "peer1", true)
>>>>>>> 111d252d7 (Fix lints (#625))
	if !added || err != nil {
		t.Error("Expected to successfully add vote from peer", added, err)
	}

<<<<<<< HEAD
	vote1001_0 := makeVoteHR(t, 1, 0, 1001, privVals)
	added, err = hvs.AddVote(vote1001_0, "peer1")
=======
	vote1001_0 := makeVoteHR(1, 0, 1001, privVals)
	added, err = hvs.AddVote(vote1001_0, "peer1", true)
>>>>>>> 111d252d7 (Fix lints (#625))
	if err != ErrGotVoteFromUnwantedRound {
		t.Errorf("expected GotVoteFromUnwantedRoundError, but got %v", err)
	}
	if added {
		t.Error("Expected to *not* add vote from peer, too many catchup rounds.")
	}

	added, err = hvs.AddVote(vote1001_0, "peer2")
	if !added || err != nil {
		t.Error("Expected to successfully add vote from another peer")
	}
<<<<<<< HEAD

}

func makeVoteHR(t *testing.T, height int64, valIndex, round int32, privVals []types.PrivValidator) *types.Vote {
=======
}

func TestInconsistentExtensionData(t *testing.T) {
	valSet, privVals := types.RandValidatorSet(10, 1)

	hvsE := NewExtendedHeightVoteSet(test.DefaultTestChainID, 1, valSet)
	voteNoExt := makeVoteHR(1, 0, 20, privVals)
	voteNoExt.Extension, voteNoExt.ExtensionSignature = nil, nil
	require.Panics(t, func() {
		_, _ = hvsE.AddVote(voteNoExt, "peer1", false)
	})

	hvsNoE := NewHeightVoteSet(test.DefaultTestChainID, 1, valSet)
	voteExt := makeVoteHR(1, 0, 20, privVals)
	require.Panics(t, func() {
		_, _ = hvsNoE.AddVote(voteExt, "peer1", true)
	})
}

func makeVoteHR(
	height int64,
	valIndex,
	round int32,
	privVals []types.PrivValidator,
) *types.Vote {
>>>>>>> 111d252d7 (Fix lints (#625))
	privVal := privVals[valIndex]
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		panic(err)
	}

	randBytes := cmtrand.Bytes(tmhash.Size)

	vote := &types.Vote{
		ValidatorAddress: pubKey.Address(),
		ValidatorIndex:   valIndex,
		Height:           height,
		Round:            round,
		Timestamp:        cmttime.Now(),
		Type:             cmtproto.PrecommitType,
		BlockID:          types.BlockID{Hash: randBytes, PartSetHeader: types.PartSetHeader{}},
	}
	chainID := config.ChainID()

	v := vote.ToProto()
	err = privVal.SignVote(chainID, v)
	if err != nil {
		panic(fmt.Sprintf("Error signing vote: %v", err))
	}

	vote.Signature = v.Signature

	return vote
}
