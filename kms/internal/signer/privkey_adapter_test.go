package signer

import (
	"context"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"
)

// stubBackend is a minimal backend.Signer for tests.
type stubBackend struct{ priv crypto.PrivKey }

func (s stubBackend) PubKey(context.Context) (crypto.PubKey, error)    { return s.priv.PubKey(), nil }
func (s stubBackend) Sign(_ context.Context, b []byte) ([]byte, error) { return s.priv.Sign(b) }

func TestAdapterSatisfiesPrivKeyAndSigns(t *testing.T) {
	priv := ed25519.GenPrivKey()
	be := stubBackend{priv: priv}

	pk, err := newBackendPrivKey(context.Background(), be)
	require.NoError(t, err)

	// The compile-time interface assertion lives in privkey_adapter.go
	// (var _ crypto.PrivKey = (*backendPrivKey)(nil)); here we exercise behavior.
	require.True(t, pk.PubKey().Equals(priv.PubKey()))
	require.Equal(t, "ed25519", pk.Type())

	msg := []byte("hello")
	sig, err := pk.Sign(msg)
	require.NoError(t, err)
	require.True(t, pk.PubKey().VerifySignature(msg, sig))
}
