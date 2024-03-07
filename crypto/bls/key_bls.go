//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && blst

package bls

import (
	"bytes"

	sha256 "github.com/minio/sha256-simd"

	"github.com/cometbft/cometbft/crypto"
	cmcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls/blst"

	"github.com/cometbft/cometbft/crypto/tmhash"
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

type PrivKey []byte

func NewPrivateKeyFromBytes(bz []byte) (PrivKey, error) {
	secretKey, err := blst.SecretKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}
	return secretKey.Marshal(), nil
}

func GenPrivKey() (PrivKey, error) {
	secretKey, err := blst.RandKey()
	return PrivKey(secretKey.Marshal()), err
}

// Bytes returns the byte representation of the ECDSA Private Key.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the ECDSA private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey PrivKey) PubKey() cmcrypto.PubKey {
	secretKey, _ := blst.SecretKeyFromBytes(privKey)

	return PubKey(secretKey.PublicKey().Marshal())
}

// Equals returns true if two ECDSA private keys are equal and false otherwise.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns eth_bls12_381.
func (privKey PrivKey) Type() string {
	return KeyType
}

func (privKey PrivKey) Sign(digestBz []byte) ([]byte, error) {
	secretKey, err := blst.SecretKeyFromBytes(privKey)
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

// Pubkey is a wrapper around the Ethereum bls12_381 public key type. This wrapper conforms to
// crypotypes.Pubkey to allow for the use of the Ethereum bls12_381 public key type within the
// Cosmos SDK.

// Compile-time type assertion.
var _ cmcrypto.PubKey = &PubKey{}

type PubKey []byte

// Address returns the address of the ECDSA public key.
// The function will return an empty address if the public key is invalid.
func (pubKey PubKey) Address() cmcrypto.Address {
	pk, _ := blst.PublicKeyFromBytes(pubKey)
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

	pubK, _ := blst.PublicKeyFromBytes(pubKey)
	ok, err := blst.VerifySignature(sig, [32]byte(bz[:32]), pubK)
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
func (pubKey PubKey) Type() string {
	return KeyType
}

// Equals returns true if the pubkey type is the same and their bytes are deeply equal.
func (pubKey PubKey) Equals(other cmcrypto.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}
