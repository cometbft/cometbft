//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package bls12381_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
)

func TestSignAndValidateBLS12381(t *testing.T) {
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
