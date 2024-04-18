//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package bls12381_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
)

func TestNewPrivateKeyFromBytes(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)

	privKeyBytes := privKey.Bytes()
	privKey2, err := bls12381.NewPrivateKeyFromBytes(privKeyBytes)
	require.NoError(t, err)

	assert.True(t, privKey.Equals(privKey2))

	_, err = bls12381.NewPrivateKeyFromBytes(crypto.CRandBytes(31))
	assert.Error(t, err)
}

func TestGenPrivateKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	assert.NotNil(t, privKey)
}

func TestPrivKeyBytes(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)

	privKeyBytes := privKey.Bytes()
	privKey2, err := bls12381.NewPrivateKeyFromBytes(privKeyBytes)
	require.NoError(t, err)

	assert.True(t, privKey.Equals(privKey2))
}

func TestPrivKeyPubKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()
	assert.NotNil(t, pubKey)
}

func TestPrivKeyEquals(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	privKey2, err := bls12381.GenPrivKey()
	require.NoError(t, err)

	assert.True(t, privKey.Equals(privKey))
	assert.False(t, privKey.Equals(privKey2))
}

func TestPrivKeyType(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)

	assert.Equal(t, "bls12_381", privKey.Type())
}

func TestPrivKeySignAndPubKeyVerifySignature(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(32)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifySignature(msg, sig))

	msg = crypto.CRandBytes(192)
	sig, err = privKey.Sign(msg)
	require.NoError(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))
}

func TestPrivKeyPubKey_CorruptPrivKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	privKey = privKey[1:] // corrupt key
	pubKey := privKey.PubKey()
	require.Nil(t, pubKey)
}

func TestPubKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()
	assert.NotNil(t, pubKey)
}

func TestPubKeyEquals(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()
	pubKey2 := privKey.PubKey()

	assert.True(t, pubKey.Equals(pubKey))
	assert.True(t, pubKey.Equals(pubKey2))
}

func TestPubKeyType(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()

	assert.Equal(t, "bls12_381", pubKey.Type())
}
