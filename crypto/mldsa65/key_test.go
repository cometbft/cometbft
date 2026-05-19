package mldsa65_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/mldsa65"
)

func TestSignAndVerify(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	require.Len(t, priv.Bytes(), mldsa65.PrivKeySize)

	pub := priv.PubKey()
	require.Equal(t, mldsa65.KeyType, pub.Type())
	require.Len(t, pub.Bytes(), mldsa65.PubKeySize)

	msg := []byte("the quick brown fox jumps over the lazy dog")
	sig, err := priv.Sign(msg)
	require.NoError(t, err)
	require.Len(t, sig, mldsa65.SignatureSize)

	require.True(t, pub.VerifySignature(msg, sig))

	// Tamper with the message.
	tampered := append([]byte(nil), msg...)
	tampered[0] ^= 0xff
	require.False(t, pub.VerifySignature(tampered, sig))

	// Tamper with the signature.
	badSig := append([]byte(nil), sig...)
	badSig[0] ^= 0xff
	require.False(t, pub.VerifySignature(msg, badSig))
}

func TestDeterministicFromSeed(t *testing.T) {
	seed := make([]byte, mldsa65.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}

	a, err := mldsa65.GenPrivKeyFromSeed(seed)
	require.NoError(t, err)
	b, err := mldsa65.GenPrivKeyFromSeed(seed)
	require.NoError(t, err)
	require.True(t, a.Equals(b))
	require.True(t, a.PubKey().Equals(b.PubKey()))
}

func TestRoundTripBytes(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)

	got, err := mldsa65.NewPrivKeyFromBytes(priv.Bytes())
	require.NoError(t, err)
	require.True(t, priv.Equals(got))

	pub := priv.PubKey().(mldsa65.PubKey)
	gotPub, err := mldsa65.NewPubKeyFromBytes(pub.Bytes())
	require.NoError(t, err)
	require.True(t, pub.Equals(gotPub))

	// Wrong sizes rejected.
	_, err = mldsa65.NewPrivKeyFromBytes(priv.Bytes()[:10])
	require.ErrorIs(t, err, mldsa65.ErrInvalidPrivKeySize)
	_, err = mldsa65.NewPubKeyFromBytes(pub.Bytes()[:10])
	require.ErrorIs(t, err, mldsa65.ErrInvalidPubKeySize)
}

func TestAddressIs20Bytes(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	addr := priv.PubKey().Address()
	require.Len(t, addr, 20)
}
