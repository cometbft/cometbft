//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && !bls12381

package bls

import (
	"github.com/cometbft/cometbft/crypto"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

const (
	// PrivKeySize defines the length of the PrivKey byte array.
	PrivKeySize = 32
	// PubKeySize defines the length of the PubKey byte array.
	PubKeySize = 48
	// SignatureLength defines the byte length of a BLSSignature.
	SignatureLength = 96
	// KeyType is the string constant for the bls12_381 algorithm.
	KeyType = "bls12_381"
)

// -------------------------------------.
const (
	PrivKeyName = "cometbft/PrivKeyBLS12_381"
	PubKeyName  = "cometbft/PubKeyBLS12_381"
)

func init() {
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}

// ===============================================================================================
// Private Key
// ===============================================================================================

// PrivKey is a wrapper around the Ethereum bls12_381 private key type. This wrapper conforms to
// crypotypes.Pubkey to allow for the use of the Ethereum bls12_381 private key type within the
// Cosmos SDK.

// Compile-time type assertion.
var _ cmcrypto.PrivKey = &PrivKey{}

// PrivKey represents a BLS private key noop when blst is not set as a build flag and cgo is disabled.
type PrivKey []byte

func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
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
func (privKey PrivKey) PubKey() cmcrypto.PubKey {
	panic("bls12_381 is disabled")
}

// Equals returns true if two ECDSA private keys are equal and false otherwise.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	panic("bls12_381 is disabled")
}

// Type returns eth_bls12_381.
func (privKey PrivKey) Type() string {
	return KeyType
}

func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	panic("bls12_381 is disabled")
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum bls12_381 public key type. This wrapper conforms to
// crypotypes.Pubkey to allow for the use of the Ethereum bls12_381 public key type within the
// Cosmos SDK.

// Compile-time type assertion.
var _ cmcrypto.PubKey = &PubKey{}

// PubKey represents a BLS private key noop when blst is not set as a build flag and cgo is disabled.
type PubKey []byte

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (pubKey PubKey) Address() cmcrypto.Address {
	panic("bls12_381 is disabled")
}

func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	panic("bls12_381 is disabled")
}

// Bytes returns the pubkey byte format.
func (pubKey PubKey) Bytes() []byte {
	panic("bls12_381 is disabled")
}

// Type returns eth_bls12_381.
func (pubKey PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cmcrypto.PubKey) bool {
	panic("bls12_381 is disabled")
}
