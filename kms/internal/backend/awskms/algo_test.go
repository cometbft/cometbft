package awskms

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	cometed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

func TestDecodeEd25519PubFromSPKI(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_ = priv

	spki, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)

	got, err := decodeEd25519Pub(spki)
	require.NoError(t, err)
	require.Equal(t, algoEd25519, got.Type())
	require.Len(t, got.Bytes(), cometed25519.PubKeySize)
	require.Equal(t, []byte(pub), got.Bytes())
}

func TestDecodeEd25519PubRejectsGarbage(t *testing.T) {
	_, err := decodeEd25519Pub([]byte("not-a-spki"))
	require.Error(t, err)
}

func TestDecodeEd25519PubRejectsNonEd25519(t *testing.T) {
	// An RSA SPKI parses fine but is the wrong key type.
	der := rsaSPKIForTest(t)
	_, err := decodeEd25519Pub(der)
	require.ErrorContains(t, err, "expected ed25519")
}

func TestEd25519AlgoRegistered(t *testing.T) {
	a, ok := algos[algoEd25519]
	require.True(t, ok)
	require.Equal(t, "ECC_NIST_EDWARDS25519", string(a.keySpec))
	require.Equal(t, "ED25519_SHA_512", string(a.signAlgo))
	out, err := a.fixSig([]byte{1, 2, 3})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2, 3}, out) // identity for ed25519
}

func rsaSPKIForTest(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return der
}
