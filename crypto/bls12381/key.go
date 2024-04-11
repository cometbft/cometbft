//go:build !bls12381 || !((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64))

package bls12381

import (
	"github.com/cometbft/cometbft/crypto"
)

const (
	// PrivKeySize defines the length of the PrivKey byte array.
	PrivKeySize = 32
	// PubKeySize defines the length of the PubKey byte array.
	PubKeySize = 48
	// SignatureLength defines the byte length of a BLS signature.
	SignatureLength = 96
	// KeyType is the string constant for the bls12_381 algorithm.
	KeyType = "bls12_381"
	// Enabled indicates if this curve is enabled.
	Enabled = false
)

// -------------------------------------.
const (
	PubKeyName = "cometbft/PubKeyBLS12_381"
)

// ===============================================================================================
// Private Key
// ===============================================================================================

// PrivKey is a wrapper around the Ethereum bls12_381 private key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// bls12_381 private key type.

// Compile-time type assertion.
var _ crypto.PrivKey = &PrivKey{}

// PrivKey represents a BLS private key noop when blst is not set as a build flag and cgo is disabled.
type PrivKey []byte

func NewPrivateKeyFromBytes([]byte) (PrivKey, error) {
	panic("bls12_381 is disabled")
}

func GenPrivKey() (PrivKey, error) {
	panic("bls12_381 is disabled")
}

// Bytes returns the byte representation of the ECDSA Private Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the ECDSA private key's public key. If the privkey is not valid
// it returns a nil value.
func (PrivKey) PubKey() crypto.PubKey {
	panic("bls12_381 is disabled")
}

// Equals returns true if two ECDSA private keys are equal and false otherwise.
func (PrivKey) Equals(crypto.PrivKey) bool {
	panic("bls12_381 is disabled")
}

// Type returns eth_bls12_381.
func (PrivKey) Type() string {
	return KeyType
}

func (PrivKey) Sign([]byte) ([]byte, error) {
	panic("bls12_381 is disabled")
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum bls12_381 public key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// bls12_381 public key type.

// Compile-time type assertion.
var _ crypto.PubKey = &PubKey{}

// PubKey represents a BLS private key noop when blst is not set as a build flag and cgo is disabled.
type PubKey []byte

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (PubKey) Address() crypto.Address {
	panic("bls12_381 is disabled")
}

func (PubKey) VerifySignature([]byte, []byte) bool {
	panic("bls12_381 is disabled")
}

// Bytes returns the pubkey byte format.
func (PubKey) Bytes() []byte {
	panic("bls12_381 is disabled")
}

// Type returns eth_bls12_381.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (PubKey) Equals(crypto.PubKey) bool {
	panic("bls12_381 is disabled")
}
