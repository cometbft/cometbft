package mldsa65_test

import (
	"fmt"
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

// TestVerifyManySignatures verifies many signatures under the same pubkey
// and confirms that repeated verification on a single PubKey instance yields
// correct results.
func TestVerifyManySignatures(t *testing.T) {
	priv, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	pub := priv.PubKey().(mldsa65.PubKey)

	const iterations = 256
	for i := 0; i < iterations; i++ {
		msg := []byte(fmt.Sprintf("message %d", i))
		sig, err := priv.Sign(msg)
		require.NoError(t, err)
		require.True(t, pub.VerifySignature(msg, sig), "iteration %d", i)

		bad := append([]byte(nil), sig...)
		bad[0] ^= 0xff
		require.False(t, pub.VerifySignature(msg, bad), "tamper iteration %d", i)
	}
}

// BenchmarkSign measures Sign on a single PrivKey. With the parsed key cached
// in the struct, this should not include UnmarshalBinary cost.
func BenchmarkSign(b *testing.B) {
	priv, err := mldsa65.GenPrivKey()
	if err != nil {
		b.Fatal(err)
	}
	msg := []byte("benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := priv.Sign(msg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVerifySignature measures verification under a single pubkey.
func BenchmarkVerifySignature(b *testing.B) {
	priv, err := mldsa65.GenPrivKey()
	if err != nil {
		b.Fatal(err)
	}
	pub := priv.PubKey().(mldsa65.PubKey)
	msg := []byte("benchmark message")
	sig, err := priv.Sign(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !pub.VerifySignature(msg, sig) {
			b.Fatal("verify failed")
		}
	}
}

// BenchmarkNewPubKeyFromBytes measures the cost of parsing a packed pubkey,
// which is now the dominant cost when constructing PubKeys from wire bytes.
func BenchmarkNewPubKeyFromBytes(b *testing.B) {
	priv, err := mldsa65.GenPrivKey()
	if err != nil {
		b.Fatal(err)
	}
	bz := priv.PubKey().Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mldsa65.NewPubKeyFromBytes(bz); err != nil {
			b.Fatal(err)
		}
	}
}
