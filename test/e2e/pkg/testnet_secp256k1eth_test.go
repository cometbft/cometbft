package e2e_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/crypto/secp256k1eth"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/types"
)

// TestSecp256k1EthNetwork builds the in-memory representation of an e2e testnet
// pinned to the secp256k1eth validator key type (networks/secp256k1eth.toml),
// exercising manifest parsing, deterministic key generation through the e2e
// keyGenerator, sign/verify, and the encoding bridge — the same path the CI
// runner drives prior to `runner setup`.
func TestSecp256k1EthNetwork(t *testing.T) {
	testSecp256k1EthNetwork(t, "secp256k1eth.toml", false)
}

// TestSecp256k1EthLibp2pNetwork is the same as TestSecp256k1EthNetwork but for
// the variant manifest with libp2p networking enabled on every node
// (networks/secp256k1eth-libp2p.toml).
func TestSecp256k1EthLibp2pNetwork(t *testing.T) {
	testSecp256k1EthNetwork(t, "secp256k1eth-libp2p.toml", true)
}

func testSecp256k1EthNetwork(t *testing.T, manifestFile string, wantLibp2p bool) {
	t.Helper()

	manifestPath, err := filepath.Abs(filepath.Join("..", "networks", manifestFile))
	require.NoError(t, err)

	manifest, err := e2e.LoadManifest(manifestPath)
	require.NoError(t, err)
	require.Equal(t, secp256k1eth.KeyType, manifest.KeyType,
		"manifest must pin all validators to secp256k1eth")
	require.NotEmpty(t, manifest.Nodes, "manifest must declare at least one node")

	ifd, err := e2e.NewDockerInfrastructureData(manifest)
	require.NoError(t, err, "in-memory infrastructure data must build without docker")

	testnet, err := e2e.LoadTestnet(manifestPath, ifd)
	require.NoError(t, err, "testnet must load from the secp256k1eth manifest")
	require.Equal(t, secp256k1eth.KeyType, testnet.KeyType)
	require.NotEmpty(t, testnet.Validators, "testnet must contain validators")

	// Every validator must have a secp256k1eth priv key of the expected size,
	// produce signatures of the expected size, and verify those signatures
	// against the corresponding pub key.
	msg := []byte("secp256k1eth e2e network produces blocks")
	for node, power := range testnet.Validators {
		require.Positive(t, power, "validator %s has non-positive voting power", node.Name)
		require.Equal(t, wantLibp2p, node.UseLibp2p,
			"validator %s libp2p setting", node.Name)

		priv := node.PrivvalKey
		require.NotNil(t, priv, "validator %s has no priv key", node.Name)
		require.Equal(t, secp256k1eth.KeyType, priv.Type(),
			"validator %s priv key type", node.Name)
		require.Len(t, priv.Bytes(), secp256k1eth.PrivKeySize,
			"validator %s priv key size", node.Name)

		pub := priv.PubKey()
		require.Equal(t, secp256k1eth.KeyType, pub.Type())
		require.Len(t, pub.Bytes(), secp256k1eth.PubKeySize)
		require.Len(t, pub.Address(), 20, "validator address must be 20 bytes")

		sig, err := priv.Sign(msg)
		require.NoError(t, err, "validator %s sign", node.Name)
		require.Len(t, sig, secp256k1eth.SignatureSize,
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
	// in runner/setup.go; if secp256k1eth weren't registered, the chain would
	// reject its own validators at genesis.
	require.Contains(t, types.ABCIPubKeyTypesToNames, secp256k1eth.KeyType,
		"secp256k1eth must be a recognized validator pub-key type")

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
