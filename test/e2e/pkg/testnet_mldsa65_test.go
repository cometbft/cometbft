package e2e_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/crypto/mldsa65"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/types"
)

// TestMlDsa65Network spins up the in-memory representation of an end-to-end
// testnet pinned to the ML-DSA-65 validator key type, using the dedicated
// manifest under networks/mldsa65.toml.
//
// This is the same construction path the cometbft e2e harness drives in CI
// (manifest -> LoadTestnet -> per-node PrivvalKey generation); the test stops
// short of actually launching docker containers, but exercises everything the
// runner does prior to `runner setup`, including:
//
//   - Manifest parsing of `key_type = "ml_dsa_65"`.
//   - Deterministic ML-DSA-65 key generation through the e2e keyGenerator.
//   - Sign / verify against the generated validator keys.
//   - The encoding bridge (encoding.PubKeyToProto / PubKeyFromProto) used when
//     putting validator pubkeys into the genesis and ABCI messages.
//   - That the ML-DSA-65 key type is admitted by the validator pub-key-type
//     map used to populate `genesis.ConsensusParams.Validator.PubKeyTypes`.
//
// To run the full end-to-end (docker-backed) test suite against this network:
//
//	cd test/e2e
//	make runner
//	./build/runner -f networks/mldsa65.toml setup
//	./build/runner -f networks/mldsa65.toml start
//	./build/runner -f networks/mldsa65.toml test
func TestMlDsa65Network(t *testing.T) {
	manifestPath, err := filepath.Abs(filepath.Join("..", "networks", "mldsa65.toml"))
	require.NoError(t, err)

	manifest, err := e2e.LoadManifest(manifestPath)
	require.NoError(t, err)
	require.Equal(t, mldsa65.KeyType, manifest.KeyType,
		"manifest must pin all validators to ml_dsa_65")
	require.NotEmpty(t, manifest.Nodes, "manifest must declare at least one node")

	ifd, err := e2e.NewDockerInfrastructureData(manifest)
	require.NoError(t, err, "in-memory infrastructure data must build without docker")

	testnet, err := e2e.LoadTestnet(manifestPath, ifd)
	require.NoError(t, err, "testnet must load from the ml_dsa_65 manifest")
	require.Equal(t, mldsa65.KeyType, testnet.KeyType)
	require.NotEmpty(t, testnet.Validators, "testnet must contain validators")

	// Every validator must have an ML-DSA-65 priv key of the expected size,
	// produce signatures of the expected size, and verify those signatures
	// against the corresponding pub key.
	msg := []byte("ml-dsa-65 e2e network produces blocks")
	for node, power := range testnet.Validators {
		require.Positive(t, power, "validator %s has non-positive voting power", node.Name)

		priv := node.PrivvalKey
		require.NotNil(t, priv, "validator %s has no priv key", node.Name)
		require.Equal(t, mldsa65.KeyType, priv.Type(),
			"validator %s priv key type", node.Name)
		require.Len(t, priv.Bytes(), mldsa65.PrivKeySize,
			"validator %s priv key size", node.Name)

		pub := priv.PubKey()
		require.Equal(t, mldsa65.KeyType, pub.Type())
		require.Len(t, pub.Bytes(), mldsa65.PubKeySize)
		require.Len(t, pub.Address(), 20, "validator address must be SHA256-20")

		sig, err := priv.Sign(msg)
		require.NoError(t, err, "validator %s sign", node.Name)
		require.Len(t, sig, mldsa65.SignatureSize,
			"validator %s signature size", node.Name)
		require.True(t, pub.VerifySignature(msg, sig),
			"validator %s sig should verify", node.Name)

		// The encoding bridge is the path used when putting a validator
		// pubkey into the genesis doc and ABCI ValidatorUpdate messages.
		protoPub, err := encoding.PubKeyToProto(pub)
		require.NoError(t, err, "PubKeyToProto for %s", node.Name)
		roundTrip, err := encoding.PubKeyFromProto(protoPub)
		require.NoError(t, err, "PubKeyFromProto for %s", node.Name)
		require.True(t, pub.Equals(roundTrip),
			"validator %s pubkey must survive proto round-trip", node.Name)
	}

	// The validator pub-key-type map is what gates
	// `genesis.ConsensusParams.Validator.PubKeyTypes = []string{testnet.KeyType}`
	// in runner/setup.go; if ml_dsa_65 weren't registered, the chain would
	// reject its own validators at genesis.
	require.Contains(t, types.ABCIPubKeyTypesToNames, mldsa65.KeyType,
		"ml_dsa_65 must be a recognized validator pub-key type")

	// Sanity-check that two validators get distinct keys despite the
	// deterministic generator seed.
	var seen [][]byte
	for node := range testnet.Validators {
		for _, prev := range seen {
			require.NotEqual(t, prev, node.PrivvalKey.Bytes(),
				"validator key generator produced a duplicate key")
		}
		seen = append(seen, node.PrivvalKey.Bytes())
	}
}
