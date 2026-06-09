// Package identity manages the Ed25519 key that authenticates the SecretConnection
// to a validator. It reuses CometBFT's node-key format and storage.
package identity

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/p2p"
)

// LoadOrGen loads the identity key at path, generating and persisting one if it
// does not exist.
func LoadOrGen(path string) (crypto.PrivKey, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("identity: mkdir: %w", err)
	}
	nk, err := p2p.LoadOrGenNodeKey(path)
	if err != nil {
		return nil, fmt.Errorf("identity: load/gen node key %q: %w", path, err)
	}
	return nk.PrivKey, nil
}
