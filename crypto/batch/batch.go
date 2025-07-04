package batch

import (
	"github.com/cometbft/cometbft/v2/crypto"
	"github.com/cometbft/cometbft/v2/crypto/ed25519"
)

// CreateBatchVerifier checks if a key type implements the batch verifier interface.
// Currently only ed25519 supports batch verification.
func CreateBatchVerifier(pk crypto.PubKey) (crypto.BatchVerifier, bool) {
	if pk == nil {
		return nil, false
	}
	switch pk.Type() {
	case ed25519.KeyType:
		return ed25519.NewBatchVerifier(), true
	default:
		return nil, false
	}
}

// SupportsBatchVerifier checks if a key type implements the batch verifier
// interface.
func SupportsBatchVerifier(pk crypto.PubKey) bool {
	if pk == nil {
		return false
	}

	switch pk.Type() {
	case ed25519.KeyType:
		return true
	default:
		return false
	}
}
