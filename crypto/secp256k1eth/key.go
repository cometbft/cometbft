//go:build !secp256k1eth

package secp256k1eth

import (
	"errors"

	"github.com/cometbft/cometbft/v2/crypto"
)

const (
	// Enabled indicates if this curve is enabled.
	Enabled = false
)

// ErrDisabled is returned if the caller didn't use the `secp256k1eth` build tag.
var ErrDisabled = errors.New("secp256k1eth is disabled")

// ===============================================================================================
// Private Key
// ===============================================================================================

// PrivKey is a stub for the Ethereum secp256k1eth private key type.
// This stub conforms to crypto.Pubkey to allow the use of the Ethereum
// secp256k1eth private key type when `secp256k1eth` support is disabled`.
// Note: all operations on this type will result in a panic!

// Compile-time type assertion.
var _ crypto.PrivKey = &PrivKey{}

// PrivKey represents a secp256k1eth private key noop when secp256k1eth is not set as a build flag.
type PrivKey []byte

// GenPrivKey always panics.
func GenPrivKey() PrivKey {
	panic(ErrDisabled)
}

func GenPrivKeySecp256k1(_ []byte) PrivKey {
	panic(ErrDisabled)
}

// Bytes returns the byte representation of the Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey always panics.
func (PrivKey) PubKey() crypto.PubKey {
	panic(ErrDisabled)
}

// Type returns the key's type.
func (PrivKey) Type() string {
	return KeyType
}

// Sign always panics.
func (PrivKey) Sign([]byte) ([]byte, error) {
	panic(ErrDisabled)
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey represents a stub of the Ethereum secp256k1eth public key type.
// This stub conforms to crypto.Pubkey to allow the use of the Ethereum
// secp256k1eth public key type when build flag `secp256k1eth` is not set.
// Note: all operations on this type will result in a panic!

// Compile-time type assertion.
var _ crypto.PubKey = &PubKey{}

// PubKey represents a secp256k1eth private key noop
// when secp256k1eth is not set as a build flag.
type PubKey []byte

// Address always panics.
func (PubKey) Address() crypto.Address {
	panic(ErrDisabled)
}

// VerifySignature always panics.
func (PubKey) VerifySignature([]byte, []byte) bool {
	panic(ErrDisabled)
}

// Bytes always panics.
func (PubKey) Bytes() []byte {
	panic(ErrDisabled)
}

// Type returns the key's type.
func (PubKey) Type() string {
	return KeyType
}
