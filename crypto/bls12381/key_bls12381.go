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
	// PrivKeySize defines the length of the PrivKey byte array.
	PrivKeySize = 32
	// PubKeySize defines the length of the PubKey byte array.
	PubKeySize = 48
	// SignatureLength defines the byte length of a BLS signature.
	SignatureLength = 96
	// KeyType is the string constant for the bls12_381 algorithm.
	KeyType = "bls12_381"
	// Enabled indicates if this curve is enabled.
	Enabled = true
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

// PrivKey is a wrapper around the Ethereum bls12_381 private key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// bls12_381 private key type.

var _ crypto.PrivKey = &PrivKey{}

type PrivKey []byte

func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}
	return secretKey.Marshal(), nil
}

func GenPrivKey() (PrivKey, error) {
	secretKey, err := bls12381.RandKey()
	return PrivKey(secretKey.Marshal()), err
}

// Bytes returns the byte representation of the Private Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() crypto.PubKey {
	secretKey, _ := bls12381.SecretKeyFromBytes(privKey)

	return PubKey(secretKey.PublicKey().Marshal())
}

// Equals returns true if two private keys are equal and false otherwise.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns eth_bls12_381.
func (PrivKey) Type() string {
	return KeyType
}

func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	secretKey, err := bls12381.SecretKeyFromBytes(privKey)
	if err != nil {
		return nil, err
	}

	bz := digestBz
	if len(bz) > 32 {
		hash := sha256.Sum256(bz)
		bz = hash[:]
	}

	sig := secretKey.Sign(bz)
	return sig.Marshal(), nil
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum bls12_381 public key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// bls12_381 public key type.

var _ crypto.PubKey = &PubKey{}

type PubKey []byte

// Address returns the address of the public key.
// The function will panic if the public key is invalid.
func (pubKey PubKey) Address() crypto.Address {
	pk, _ := bls12381.PublicKeyFromBytes(pubKey)
	if len(pk.Marshal()) != PubKeySize {
		panic("pubkey is incorrect size")
	}
	// TODO: do we want to keep this address format?
	return crypto.Address(tmhash.SumTruncated(pubKey))
}

func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureLength {
		return false
	}
	bz := msg
	if len(msg) > 32 {
		hash := sha256.Sum256(msg)
		bz = hash[:]
	}

	pubK, _ := bls12381.PublicKeyFromBytes(pubKey)
	ok, err := bls12381.VerifySignature(sig, [32]byte(bz[:32]), pubK)
	if err != nil {
		return false
	}
	return ok
}

// Bytes returns the pubkey byte format.
func (pubKey PubKey) Bytes() []byte {
	return pubKey
}

// Type returns eth_bls12_381.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}
