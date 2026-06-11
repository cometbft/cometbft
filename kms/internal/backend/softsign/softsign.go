// Package softsign implements an in-memory, file-backed Ed25519 backend.Signer.
// NOT for production custody: the private key is held in process memory.
package softsign

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtjson "github.com/cometbft/cometbft/libs/json"

	"github.com/cometbft/cometbft/kms/internal/backend"
)

// Signer is a softsign backend holding an Ed25519 private key in memory.
type Signer struct {
	priv crypto.PrivKey
	pub  crypto.PubKey
}

var _ backend.Signer = (*Signer)(nil)

// Load reads a key file. It accepts either a CometBFT priv_validator_key.json
// (typed JSON with a "priv_key" field) or a file containing the base64-encoded
// 64-byte Ed25519 private key.
func Load(path string) (*Signer, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("softsign: read key file %q: %w", path, err)
	}

	priv, err := parseKey(raw)
	if err != nil {
		return nil, fmt.Errorf("softsign: parse key file %q: %w", path, err)
	}
	return &Signer{priv: priv, pub: priv.PubKey()}, nil
}

func parseKey(raw []byte) (crypto.PrivKey, error) {
	// Try priv_validator_key.json shape first (both concrete and interface-typed variants).
	if bytes.Contains(raw, []byte("priv_key")) {
		// Try interface-typed JSON first ({"type":"...","value":"..."} envelope).
		var kfIface struct {
			PrivKey crypto.PrivKey `json:"priv_key"`
		}
		if err := cmtjson.Unmarshal(raw, &kfIface); err == nil && kfIface.PrivKey != nil {
			return kfIface.PrivKey, nil
		}
		// Try concrete ed25519 JSON (plain base64 string value).
		var kfConcrete struct {
			PrivKey ed25519.PrivKey `json:"priv_key"`
		}
		if err := cmtjson.Unmarshal(raw, &kfConcrete); err == nil && len(kfConcrete.PrivKey) == ed25519.PrivateKeySize {
			return kfConcrete.PrivKey, nil
		}
	}
	// Fall back to base64 raw 64-byte ed25519 key.
	dec, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(raw)))
	if err != nil {
		return nil, fmt.Errorf("not priv_validator_key.json and not base64: %w", err)
	}
	if len(dec) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("expected %d-byte ed25519 key, got %d", ed25519.PrivateKeySize, len(dec))
	}
	return ed25519.PrivKey(dec), nil
}

// PubKey implements backend.Signer.
func (s *Signer) PubKey(context.Context) (crypto.PubKey, error) { return s.pub, nil }

// Sign implements backend.Signer.
func (s *Signer) Sign(_ context.Context, signBytes []byte) ([]byte, error) {
	return s.priv.Sign(signBytes)
}
