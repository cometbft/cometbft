package pkcs11_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pk "github.com/cometbft/cometbft/kms/internal/backend/pkcs11"
	"github.com/cometbft/cometbft/kms/internal/backend/pkcs11/pkcs11test"
)

func TestOpenPubKeySignVerify(t *testing.T) {
	module := pkcs11test.FindModule(t)
	want := pkcs11test.SetupToken(t, module)

	s, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   pkcs11test.KeyLabel,
		PIN:        pkcs11test.UserPIN,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	pub, err := s.PubKey(context.Background())
	require.NoError(t, err)
	require.True(t, pub.Equals(want), "backend pubkey must match the on-token key")

	msg := []byte("canonical-consensus-sign-bytes")
	sig, err := s.Sign(context.Background(), msg)
	require.NoError(t, err)
	require.True(t, pub.VerifySignature(msg, sig), "signature must verify under the on-token pubkey")
}

func TestOpenByKeyID(t *testing.T) {
	module := pkcs11test.FindModule(t)
	want := pkcs11test.SetupToken(t, module)

	s, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyID:      pkcs11test.KeyID,
		PIN:        pkcs11test.UserPIN,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	pub, err := s.PubKey(context.Background())
	require.NoError(t, err)
	require.True(t, pub.Equals(want))
}

func TestOpenWrongPIN(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)

	_, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   pkcs11test.KeyLabel,
		PIN:        "9999",
	})
	require.Error(t, err)
}

func TestOpenUnknownToken(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)

	_, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: "no-such-token",
		KeyLabel:   pkcs11test.KeyLabel,
		PIN:        pkcs11test.UserPIN,
	})
	require.Error(t, err)
}

func TestOpenUnknownKey(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)

	_, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   "no-such-key",
		PIN:        pkcs11test.UserPIN,
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "no object matching")
}

func TestOpenAmbiguousKeyLabel(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)
	// A second key pair with the same label but a different id makes a
	// label-only lookup ambiguous.
	pkcs11test.AddKeyPair(t, module, pkcs11test.KeyLabel, []byte{0x02})

	_, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   pkcs11test.KeyLabel,
		PIN:        pkcs11test.UserPIN,
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "multiple objects match")

	// Adding the unique key_id disambiguates.
	s, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   pkcs11test.KeyLabel,
		KeyID:      pkcs11test.KeyID,
		PIN:        pkcs11test.UserPIN,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
}

func TestDoubleCloseSafe(t *testing.T) {
	module := pkcs11test.FindModule(t)
	pkcs11test.SetupToken(t, module)

	s, err := pk.Open(pk.Config{
		Module:     module,
		TokenLabel: pkcs11test.TokenLabel,
		KeyLabel:   pkcs11test.KeyLabel,
		PIN:        pkcs11test.UserPIN,
	})
	require.NoError(t, err)
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
}
