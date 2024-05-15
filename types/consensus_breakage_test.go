package types

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/stretchr/testify/require"
)

// Ensure validators_hash and next_validators_hash are deterministic.
func TestValidatorsHash(t *testing.T) {
	vset := deterministicValidatorSet(t)
	require.Equal(t, []byte{0x3a, 0x37, 0x2b, 0xdc, 0xb3, 0xb9, 0x41, 0x8f, 0x55, 0xe1, 0x32, 0x37, 0xc6, 0xf2, 0x80, 0x1a, 0x20, 0xf7, 0x9f, 0xbe, 0x5f, 0x46, 0xc7, 0xf3, 0xdb, 0x77, 0x80, 0x13, 0xd9, 0x3a, 0xe9, 0xd4}, vset.Hash())
}

// Ensure last_commit_hash is deterministic.
func TestLastCommitHash(t *testing.T) {
	lastCommit := deterministicLastCommit(t)
	require.Equal(t, []byte{0xf6, 0xd2, 0x8d, 0x12, 0x54, 0x35, 0x5c, 0x70, 0x74, 0x59, 0x82, 0x9d, 0x51, 0xcc, 0x81, 0xbf, 0x73, 0x7b, 0x34, 0x45, 0x1d, 0x1c, 0x15, 0xd3, 0x57, 0x3f, 0xcd, 0xe6, 0xaa, 0x64, 0x53, 0xaa}, lastCommit.Hash().Bytes())
}

// Ensure consensus_hash is deterministic.
func TestConsensusHash(t *testing.T) {
	params := DefaultConsensusParams()
	require.Equal(t, []byte{0x68, 0xec, 0xd6, 0xf3, 0x33, 0x11, 0x9c, 0xe4, 0x37, 0x51, 0xec, 0xe5, 0x83, 0xb9, 0x81, 0xf2, 0x35, 0x8, 0xae, 0xaf, 0x42, 0x21, 0xff, 0x58, 0x2b, 0x1b, 0xb3, 0x3b, 0xe4, 0x2b, 0xce, 0xfa}, params.Hash())
}

// Ensure data_hash is deterministic.
func TestDataHash(t *testing.T) {

}

// Ensure evidence_hash is deterministic.

// It's the responsibility of the ABCI developers to ensure that app_hash
// and last_results_hash are deterministic.

func deterministicValidatorSet(t *testing.T) *ValidatorSet {
	t.Helper()

	pkBytes, err := hex.DecodeString("D9838D11F68AE4679BD91BC2693CDF62FAABAEA7B4290A70ED5F200B4B67881C")
	require.NoError(t, err)
	pk := ed25519.PubKey(pkBytes)
	val := NewValidator(pk, 1)
	return NewValidatorSet([]*Validator{val})
}

func deterministicLastCommit(t *testing.T) *Commit {
	t.Helper()

	staticTmHash := make([]byte, tmhash.Size)
	staticValAddress := make([]byte, tmhash.TruncatedSize)
	return &Commit{
		Height:  1,
		Round:   0,
		BlockID: BlockID{staticTmHash, PartSetHeader{123, staticTmHash}},
		Signatures: []CommitSig{
			CommitSig{
				BlockIDFlag: BlockIDFlagAbsent,
			},
			CommitSig{
				BlockIDFlag:      BlockIDFlagCommit,
				ValidatorAddress: staticValAddress,
				Timestamp:        time.Unix(1515151515, 0),
				Signature:        make([]byte, ed25519.SignatureSize),
			}},
	}
}
