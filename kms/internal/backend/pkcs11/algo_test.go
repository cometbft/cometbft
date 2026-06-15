package pkcs11

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

func TestEd25519DecodePub_DERWrapped(t *testing.T) {
	priv := ed25519.GenPrivKey()
	raw := priv.PubKey().Bytes() // 32 bytes
	// PKCS#11 v3.0 returns CKA_EC_POINT as a DER OCTET STRING: 0x04 0x20 <32 bytes>.
	der := append([]byte{0x04, 0x20}, raw...)

	pub, err := algos["ed25519"].decodePub(der)
	require.NoError(t, err)
	require.True(t, pub.Equals(priv.PubKey()))
}

func TestEd25519DecodePub_Raw(t *testing.T) {
	priv := ed25519.GenPrivKey()
	raw := priv.PubKey().Bytes() // some tokens return the raw 32-byte point

	pub, err := algos["ed25519"].decodePub(raw)
	require.NoError(t, err)
	require.True(t, pub.Equals(priv.PubKey()))
}

func TestEd25519DecodePub_BadLength(t *testing.T) {
	_, err := algos["ed25519"].decodePub([]byte{0x01, 0x02, 0x03})
	require.Error(t, err)
}

func TestEd25519FixSig_Identity(t *testing.T) {
	sig := []byte("a-64-byte-ed25519-signature-placeholder-value-for-testing-only!!")
	out, err := algos["ed25519"].fixSig(sig)
	require.NoError(t, err)
	require.Equal(t, sig, out)
}
