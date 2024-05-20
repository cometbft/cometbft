package types

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtversion "github.com/cometbft/cometbft/api/cometbft/version/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
)

// Ensure validators_hash and next_validators_hash are deterministic.
func TestValidatorsHash(t *testing.T) {
	vset := deterministicValidatorSet(t)
	require.Equal(t, []byte{0x3a, 0x37, 0x2b, 0xdc, 0xb3, 0xb9, 0x41, 0x8f, 0x55, 0xe1, 0x32, 0x37, 0xc6, 0xf2, 0x80, 0x1a, 0x20, 0xf7, 0x9f, 0xbe, 0x5f, 0x46, 0xc7, 0xf3, 0xdb, 0x77, 0x80, 0x13, 0xd9, 0x3a, 0xe9, 0xd4}, vset.Hash())
}

// Ensure last_commit_hash is deterministic.
func TestLastCommitHash(t *testing.T) {
	lastCommit := deterministicLastCommit()
	require.Equal(t, []byte{0x8, 0xba, 0xdc, 0xd5, 0x36, 0x3f, 0x2e, 0xb5, 0x47, 0x91, 0x0, 0xc0, 0xa, 0xea, 0x5c, 0x20, 0xb, 0x5b, 0x81, 0x2, 0x6, 0x27, 0xe9, 0x22, 0x77, 0xff, 0x82, 0xc3, 0x1, 0x1e, 0xba, 0xb5}, lastCommit.Hash().Bytes())
}

// Ensure consensus_hash is deterministic.
func TestConsensusHash(t *testing.T) {
	params := DefaultConsensusParams()
	require.Equal(t, []byte{0x68, 0xec, 0xd6, 0xf3, 0x33, 0x11, 0x9c, 0xe4, 0x37, 0x51, 0xec, 0xe5, 0x83, 0xb9, 0x81, 0xf2, 0x35, 0x8, 0xae, 0xaf, 0x42, 0x21, 0xff, 0x58, 0x2b, 0x1b, 0xb3, 0x3b, 0xe4, 0x2b, 0xce, 0xfa}, params.Hash())
}

// Ensure data_hash is deterministic.
func TestDataHash(t *testing.T) {
	// hash from byte slices
	data := Data{
		Txs: Txs{
			[]byte{0x01, 0x02, 0x03},
		},
	}
	require.Equal(t, []byte{0x17, 0xfd, 0x4, 0x25, 0xd0, 0x2b, 0xac, 0x41, 0x1c, 0x75, 0x83, 0xd6, 0xa9, 0xfa, 0x75, 0x80, 0x37, 0x9a, 0x26, 0x91, 0x62, 0x9e, 0x9c, 0x1c, 0xe6, 0xc6, 0x7f, 0x89, 0x53, 0x19, 0xb, 0x99}, data.Hash().Bytes())
}

// Ensure evidence_hash is deterministic.
func TestEvidenceHash(t *testing.T) {
	valSet := deterministicValidatorSet(t)

	// DuplicateVoteEvidence
	valAddress := valSet.Validators[0].Address
	dp, err := NewDuplicateVoteEvidence(
		deterministicVote(1, valAddress),
		deterministicVote(2, valAddress),
		time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		valSet,
	)
	require.NoError(t, err)

	require.Equal(t, []byte{0x92, 0xa7, 0x6b, 0x39, 0x43, 0x37, 0xf0, 0xc0, 0x4c, 0x95, 0x15, 0x46, 0xad, 0xc7, 0x5a, 0x59, 0xcb, 0x7c, 0xae, 0x7b, 0xca, 0x7, 0xe, 0x49, 0xfc, 0x93, 0xc1, 0x11, 0x14, 0x9, 0xb5, 0xe2}, dp.Hash())

	// LightClientAttackEvidence
	lcE := LightClientAttackEvidence{
		ConflictingBlock: &LightBlock{
			SignedHeader: &SignedHeader{},
			ValidatorSet: valSet,
		},
		CommonHeight: 1,

		ByzantineValidators: valSet.Validators,
		TotalVotingPower:    100,
		Timestamp:           time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	require.Equal(t, []byte{0x58, 0xcc, 0x2f, 0x44, 0xd3, 0xa2, 0x78, 0x66, 0x87, 0x47, 0x1, 0xfb, 0xad, 0x57, 0x3d, 0xa9, 0xad, 0x1c, 0xfd, 0x88, 0xfa, 0x31, 0x45, 0x53, 0x1c, 0x82, 0x2f, 0x20, 0xa5, 0x8b, 0xee, 0xa1}, lcE.Hash())
}

// Ensure last_block_id is deterministic.
func TestLastBlockID(t *testing.T) {
	lbi := BlockID{
		Hash: tmhash.Sum([]byte("blockID_hash")),
		PartSetHeader: PartSetHeader{
			Total: 1000000,
			Hash:  tmhash.Sum([]byte("blockID_part_set_header_hash")),
		},
	}
	lbip := lbi.ToProto()
	v, err := lbip.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0xa, 0x20, 0x8b, 0x1, 0x2, 0x33, 0x86, 0xc3, 0x71, 0x77, 0x8e, 0xcb, 0x63, 0x68, 0x57, 0x3e, 0x53, 0x9a, 0xfc, 0x3c, 0xc8, 0x60, 0xec, 0x3a, 0x2f, 0x61, 0x4e, 0x54, 0xfe, 0x56, 0x52, 0xf4, 0xfc, 0x80, 0x12, 0x26, 0x8, 0xc0, 0x84, 0x3d, 0x12, 0x20, 0x72, 0xdb, 0x3d, 0x95, 0x96, 0x35, 0xdf, 0xf1, 0xbb, 0x56, 0x7b, 0xed, 0xaa, 0x70, 0x57, 0x33, 0x92, 0xc5, 0x15, 0x96, 0x66, 0xa3, 0xf8, 0xca, 0xf1, 0x1e, 0x41, 0x3a, 0xac, 0x52, 0x20, 0x7a}, v)
}

// Ensure version is deterministic.
func TestVersion(t *testing.T) {
	cV := cmtversion.Consensus{
		Block: 1,
		App:   2,
	}
	v, err := cV.Marshal()
	require.NoError(t, err)
	require.Equal(t, []byte{0x8, 0x1, 0x10, 0x2}, v)
}

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

func deterministicLastCommit() *Commit {
	return &Commit{
		Height: 1,
		Round:  0,
		BlockID: BlockID{
			Hash: tmhash.Sum([]byte("blockID_hash")),
			PartSetHeader: PartSetHeader{
				Total: 1000000,
				Hash:  tmhash.Sum([]byte("blockID_part_set_header_hash")),
			},
		},
		Signatures: []CommitSig{
			{
				BlockIDFlag: BlockIDFlagAbsent,
			},
			{
				BlockIDFlag:      BlockIDFlagCommit,
				ValidatorAddress: crypto.AddressHash([]byte("validator_address")),
				Timestamp:        time.Unix(1515151515, 0),
				Signature:        make([]byte, ed25519.SignatureSize),
			},
		},
	}
}

func deterministicVote(t byte, valAddress crypto.Address) *Vote {
	stamp, err := time.Parse(TimeFormat, "2017-12-25T03:00:01.234Z")
	if err != nil {
		panic(err)
	}

	return &Vote{
		Type:      SignedMsgType(t),
		Height:    3,
		Round:     2,
		Timestamp: stamp,
		BlockID: BlockID{
			Hash: tmhash.Sum([]byte("blockID_hash")),
			PartSetHeader: PartSetHeader{
				Total: 1000000,
				Hash:  tmhash.Sum([]byte("blockID_part_set_header_hash")),
			},
		},
		ValidatorAddress: valAddress,
		ValidatorIndex:   56789,
	}
}
