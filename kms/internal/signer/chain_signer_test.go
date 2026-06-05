package signer_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/signer"
)

type memBackend struct{ priv crypto.PrivKey }

func (m memBackend) PubKey(context.Context) (crypto.PubKey, error)    { return m.priv.PubKey(), nil }
func (m memBackend) Sign(_ context.Context, b []byte) ([]byte, error) { return m.priv.Sign(b) }

const chainID = "test-chain-1"

func newSigner(t *testing.T) (*signer.ChainSigner, crypto.PubKey, string) {
	t.Helper()
	priv := ed25519.GenPrivKey()
	state := filepath.Join(t.TempDir(), "state.json")
	cs, err := signer.NewChainSigner(chainID, memBackend{priv: priv}, state)
	require.NoError(t, err)
	return cs, priv.PubKey(), state
}

func precommit(h int64, r int32) *cmtproto.Vote {
	return &cmtproto.Vote{Type: cmtproto.PrecommitType, Height: h, Round: r}
}

func TestSignVoteProducesVerifiableSignature(t *testing.T) {
	cs, pub, _ := newSigner(t)
	v := precommit(10, 0)
	require.NoError(t, cs.SignVote(chainID, v))
	require.True(t, pub.VerifySignature(types.VoteSignBytes(chainID, v), v.Signature))
}

func TestDoubleSignRegressionRejected(t *testing.T) {
	cs, _, _ := newSigner(t)
	require.NoError(t, cs.SignVote(chainID, precommit(10, 0)))
	err := cs.SignVote(chainID, precommit(9, 0))
	require.Error(t, err)
}

func TestStatePersistsAcrossReload(t *testing.T) {
	priv := ed25519.GenPrivKey()
	state := filepath.Join(t.TempDir(), "state.json")

	cs1, err := signer.NewChainSigner(chainID, memBackend{priv: priv}, state)
	require.NoError(t, err)
	require.NoError(t, cs1.SignVote(chainID, precommit(100, 0)))

	cs2, err := signer.NewChainSigner(chainID, memBackend{priv: priv}, state)
	require.NoError(t, err)
	require.Error(t, cs2.SignVote(chainID, precommit(50, 0)))

	_, statErr := os.Stat(state)
	require.NoError(t, statErr)
}

func TestVoteExtensionSignedForNonNilPrecommit(t *testing.T) {
	cs, pub, _ := newSigner(t)
	v := &cmtproto.Vote{
		Type:    cmtproto.PrecommitType,
		Height:  10,
		Round:   0,
		BlockID: cmtproto.BlockID{Hash: []byte("01234567890123456789012345678901")},
	}
	require.NoError(t, cs.SignVote(chainID, v))
	require.NotEmpty(t, v.ExtensionSignature)
	require.True(t, pub.VerifySignature(types.VoteExtensionSignBytes(chainID, v), v.ExtensionSignature))
}

func TestConflictingBlockSameHRSRejected(t *testing.T) {
	cs, _, _ := newSigner(t)

	blockA := &cmtproto.Vote{
		Type:    cmtproto.PrecommitType,
		Height:  10,
		Round:   0,
		BlockID: cmtproto.BlockID{Hash: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")},
	}
	require.NoError(t, cs.SignVote(chainID, blockA))

	// Same H/R/step, DIFFERENT block -> must be refused (conflicting data).
	blockB := &cmtproto.Vote{
		Type:    cmtproto.PrecommitType,
		Height:  10,
		Round:   0,
		BlockID: cmtproto.BlockID{Hash: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")},
	}
	require.Error(t, cs.SignVote(chainID, blockB))
}

func TestSignProposalVerifiableAndRegressionRejected(t *testing.T) {
	cs, pub, _ := newSigner(t)

	prop := &cmtproto.Proposal{Type: cmtproto.ProposalType, Height: 20, Round: 0}
	require.NoError(t, cs.SignProposal(chainID, prop))
	require.True(t, pub.VerifySignature(types.ProposalSignBytes(chainID, prop), prop.Signature))

	// Lower height must be refused.
	lower := &cmtproto.Proposal{Type: cmtproto.ProposalType, Height: 19, Round: 0}
	require.Error(t, cs.SignProposal(chainID, lower))
}

func TestStateSaveFailureReturnsErrorNotPanic(t *testing.T) {
	priv := ed25519.GenPrivKey()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")

	cs, err := signer.NewChainSigner(chainID, memBackend{priv: priv}, statePath)
	require.NoError(t, err)

	// Sabotage persistence: remove any state file and create a DIRECTORY at the
	// state path, so the atomic write (temp file + rename onto statePath) fails.
	_ = os.Remove(statePath)
	require.NoError(t, os.Mkdir(statePath, 0o755))

	// SignVote must return an error (recovered panic), and must NOT panic the test.
	err = cs.SignVote(chainID, precommit(30, 0))
	require.Error(t, err)
}
