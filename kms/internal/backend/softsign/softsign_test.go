package softsign_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/backend/softsign"
)

func TestLoadBase64AndSign(t *testing.T) {
	priv := ed25519.GenPrivKey()
	dir := t.TempDir()
	path := filepath.Join(dir, "key.b64")
	require.NoError(t, os.WriteFile(path, []byte(base64.StdEncoding.EncodeToString(priv.Bytes())), 0o600))

	s, err := softsign.Load(path)
	require.NoError(t, err)

	pub, err := s.PubKey(context.Background())
	require.NoError(t, err)
	require.True(t, pub.Equals(priv.PubKey()))

	msg := []byte("canonical-sign-bytes")
	sig, err := s.Sign(context.Background(), msg)
	require.NoError(t, err)
	require.True(t, pub.VerifySignature(msg, sig))
}

func TestLoadPrivValidatorKeyJSON(t *testing.T) {
	priv := ed25519.GenPrivKey()
	raw, err := cmtjson.MarshalIndent(struct {
		PrivKey ed25519.PrivKey `json:"priv_key"`
	}{PrivKey: priv}, "", "  ")
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "priv_validator_key.json")
	require.NoError(t, os.WriteFile(path, raw, 0o600))

	s, err := softsign.Load(path)
	require.NoError(t, err)
	pub, err := s.PubKey(context.Background())
	require.NoError(t, err)
	require.True(t, pub.Equals(priv.PubKey()))
}

func TestLoadRejectsMissingFile(t *testing.T) {
	_, err := softsign.Load(filepath.Join(t.TempDir(), "nope"))
	require.Error(t, err)
}
