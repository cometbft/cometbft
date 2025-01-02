//go:build bls12381

package bls12381_test

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/bls12381"
	"github.com/stretchr/testify/require"
)

func TestAggregateAndVerify(t *testing.T) {
	// Generate private keys.
	genPrivKeyFn := func() *bls12381.PrivKey {
		k, err := bls12381.GenPrivKey()
		if err != nil {
			panic(err)
		}
		return k
	}

	privateKeys := []*bls12381.PrivKey{
		genPrivKeyFn(),
		genPrivKeyFn(),
		genPrivKeyFn(),
	}

	msg := []byte("hello world")

	// Generate signatures.
	signatures := make([][]byte, len(privateKeys))
	for i, privKey := range privateKeys {
		sig, err := privKey.Sign(msg)
		require.NoError(t, err)
		signatures[i] = sig

		valid := privKey.PubKey().VerifySignature(msg, sig)
		require.True(t, valid)
	}

	// Aggregate signatures.
	aggregatedSignature, err := bls12381.AggregateSignatures(signatures)
	require.NoError(t, err)
	require.NotNil(t, aggregatedSignature)

	pubKeys := make([]*bls12381.PubKey, len(privateKeys))
	for i, privKey := range privateKeys {
		pubKeys[i] = privKey.PubKey().(*bls12381.PubKey)
	}

	// Verify aggregated signature
	valid := bls12381.VerifyAggregateSignature(aggregatedSignature, pubKeys, msg)
	require.True(t, valid)

	// Test with invalid aggregated signature
	invalidSignature := []byte("Invalid")
	valid = bls12381.VerifyAggregateSignature(invalidSignature, pubKeys, msg)
	require.False(t, valid)
}
