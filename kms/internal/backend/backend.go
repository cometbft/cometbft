// Package backend defines the signing-key abstraction used by cometkms.
// Implementations wrap a concrete key custodian (softsign file, PKCS#11 HSM,
// AWS KMS, ...). The interface is algorithm-agnostic: the public key carries its
// own algorithm via crypto.PubKey.Type(), and Sign produces a valid signature
// over the canonical consensus sign-bytes using whatever scheme the key requires.
package backend

import (
	"context"

	"github.com/cometbft/cometbft/crypto"
)

// Signer is implemented by every key backend.
type Signer interface {
	// PubKey returns the validator public key.
	PubKey(ctx context.Context) (crypto.PubKey, error)
	// Sign signs the canonical consensus sign-bytes and returns the signature.
	Sign(ctx context.Context, signBytes []byte) ([]byte, error)
}
