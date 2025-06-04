//go:build bls12381

package bls12381_test

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	blst "github.com/supranational/blst/bindings/go"

	"github.com/cometbft/cometbft/v2/crypto"
	"github.com/cometbft/cometbft/v2/crypto/bls12381"
)

func TestNewPrivateKeyFromBytes(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()

	privKeyBytes := privKey.Bytes()
	privKey2, err := bls12381.NewPrivateKeyFromBytes(privKeyBytes)
	require.NoError(t, err)
	defer privKey2.Zeroize()

	assert.Equal(t, privKey, privKey2)

	_, err = bls12381.NewPrivateKeyFromBytes(crypto.CRandBytes(31))
	assert.Error(t, err)
}

func TestGenPrivateKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
	assert.NotNil(t, privKey)
}

func TestGenPrivKeyFromSecret(t *testing.T) {
	secret := []byte("this is my secret")
	privKey, err := bls12381.GenPrivKeyFromSecret(secret)
	require.NoError(t, err)
	assert.NotNil(t, privKey)
}

func TestGenPrivKeyFromSecret_SignVerify(t *testing.T) {
	secret := []byte("this is my secret for priv key")
	priv, err := bls12381.GenPrivKeyFromSecret(secret)
	require.NoError(t, err)
	msg := []byte("this is my message to sign")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	pub := priv.PubKey()
	assert.True(t, pub.VerifySignature(msg, sig), "Signature did not verify")
}

func TestPrivKeyBytes(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()

	privKeyBytes := privKey.Bytes()
	privKey2, err := bls12381.NewPrivateKeyFromBytes(privKeyBytes)
	require.NoError(t, err)
	defer privKey2.Zeroize()

	assert.Equal(t, privKey, privKey2)
}

func TestPrivKeyPubKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	pubKey := privKey.PubKey()
	assert.NotNil(t, pubKey)
}

func TestPrivKeyType(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()

	assert.Equal(t, "bls12_381", privKey.Type())
}

func TestPrivKeySignAndPubKeyVerifySignature(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
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

func TestPubKey(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
	pubKey := privKey.PubKey()
	assert.NotNil(t, pubKey)
}

func TestPubKeyType(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
	pubKey := privKey.PubKey()

	assert.Equal(t, "bls12_381", pubKey.Type())
}

func TestConst(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
	assert.Equal(t, bls12381.PrivKeySize, len(privKey.Bytes()))

	pubKey := privKey.PubKey()
	assert.Equal(t, bls12381.PubKeySize, len(pubKey.Bytes()))

	msg := crypto.CRandBytes(32)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)
	assert.Equal(t, bls12381.SignatureLength, len(sig))
}

func TestPrivKey_MarshalJSON(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()

	jsonBytes, err := privKey.MarshalJSON()
	require.NoError(t, err)

	privKey2 := new(bls12381.PrivKey)
	err = privKey2.UnmarshalJSON(jsonBytes)
	require.NoError(t, err)
}

func TestPubKey_MarshalJSON(t *testing.T) {
	privKey, err := bls12381.GenPrivKey()
	require.NoError(t, err)
	defer privKey.Zeroize()
	pubKey, _ := privKey.PubKey().(bls12381.PubKey)

	jsonBytes, err := pubKey.MarshalJSON()
	require.NoError(t, err)

	pubKey2 := new(bls12381.PubKey)
	err = pubKey2.UnmarshalJSON(jsonBytes)
	require.NoError(t, err)
}

func TestPubKey_NewPublicKeyFromInvalidBytes(t *testing.T) {
	unmarshal := func(s string) ([]byte, error) {
		type blstPublicKey = blst.P1Affine

		bz, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		pk := new(blstPublicKey).Uncompress(bz)
		if pk == nil {
			return nil, bls12381.ErrDeserialization
		}
		pkc := pk.Serialize()
		if pkc == nil {
			return nil, errors.New("could not serialize pubkey")
		}
		return pkc, nil
	}

	testCases := []struct {
		desc        string
		pkStr       string
		expectedErr error
	}{
		{"NotInG1", "8123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", bls12381.ErrInfinitePubKey},
		{"InfFalseB", "800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", bls12381.ErrDeserialization},
		{"InfTrueB", "c01000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", bls12381.ErrDeserialization},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			bz, err := unmarshal(tc.pkStr)
			if err != nil {
				t.Log(tc.desc, "unmarshal error", err)
				require.Equal(t, tc.expectedErr, err)
			}

			_, err = bls12381.NewPublicKeyFromBytes(bz)
			require.Equal(t, tc.expectedErr, err)
			t.Log(tc.desc, "NewPrivateKeyFromBytes error", err)
		})
	}
}
