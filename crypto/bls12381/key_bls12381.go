//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package bls12381

import (
	"bytes"

	sha256 "github.com/minio/sha256-simd"

	"github.com/cometbft/cometbft/crypto"
	bls12381 "github.com/cosmos/crypto/curves/bls12381"

	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

const (
	// Enabled indicates if this curve is enabled.
	Enabled = true
)

// -------------------------------------.

func init() {
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}

// ===============================================================================================
// Private Key
// ===============================================================================================

// PrivKey is a wrapper around the Ethereum BLS12-381 private key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// BLS12-381 private key type.

var _ crypto.PrivKey = &PrivKey{}

type PrivKey []byte

// NewPrivateKeyFromBytes build a new key from the given bytes.
func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}
	return secretKey.Marshal(), nil
}

// GenPrivKey generates a new key.
func GenPrivKey() (PrivKey, error) {
	secretKey, err := bls12381.RandKey()
	return PrivKey(secretKey.Marshal()), err
}

// Bytes returns the byte representation of the Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() crypto.PubKey {
	secretKey, err := bls12381.SecretKeyFromBytes(privKey)
	if err != nil {
		return nil
	}

	return PubKey(secretKey.PublicKey().Marshal())
}

// Equals returns true if two keys are equal and false otherwise.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the type.
func (PrivKey) Type() string {
	return KeyType
}

// Sign signs the given byte array. If msg is larger than
// MaxMsgLen, SHA256 sum will be signed instead of the raw bytes.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(privKey)
	if err != nil {
		return nil, err
	}

	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		sig := secretKey.Sign(hash[:])
		return sig.Marshal(), nil
	}
	sig := secretKey.Sign(msg)
	return sig.Marshal(), nil
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum BLS12-381 public key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// BLS12-381 public key type.

var _ crypto.PubKey = &PubKey{}

type PubKey []byte

// Address returns the address of the key.
//
// The function will panic if the public key is invalid.
func (pubKey PubKey) Address() crypto.Address {
	pk, _ := bls12381.PublicKeyFromBytes(pubKey)
	if len(pk.Marshal()) != PubKeySize {
		panic("pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKey))
}

// VerifySignature verifies the given signature.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureLength {
		return false
	}

	pubK, err := bls12381.PublicKeyFromBytes(pubKey)
	if err != nil { // invalid pubkey
		return false
	}

	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		msg = hash[:]
	}

	ok, err := bls12381.VerifySignature(sig, [MaxMsgLen]byte(msg[:MaxMsgLen]), pubK)
	if err != nil { // bad signature
		return false
	}

	return ok
}

// Bytes returns the byte format.
func (pubKey PubKey) Bytes() []byte {
	return pubKey
}

// Type returns the key's type.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the other's type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}
