//go:build bls12381

package bls12381

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	blst "github.com/supranational/blst/bindings/go"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

const (
	// Enabled indicates if this curve is enabled.
	Enabled = true
)

var (
	// ErrDeserialization is returned when deserialization fails.
	ErrDeserialization = errors.New("bls12381: deserialization error")

	dstMinSig = []byte("BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_")
)

// For minimal-signature-size operations.
type (
	blstPublicKey          = blst.P2Affine
	blstSignature          = blst.P1Affine
	blstAggregateSignature = blst.P1Aggregate
	blstAggregatePublicKey = blst.P2Aggregate
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

type PrivKey struct {
	sk *blst.SecretKey
}

// NewPrivateKeyFromBytes build a new key from the given bytes.
func NewPrivateKeyFromBytes(bz []byte) (*PrivKey, error) {
	sk := new(blst.SecretKey).Deserialize(bz)
	if sk == nil {
		return nil, ErrDeserialization
	}
	return &PrivKey{sk: sk}, nil
}

// GenPrivKey generates a new key.
func GenPrivKey() (*PrivKey, error) {
	var ikm [32]byte
	_, err := rand.Read(ikm[:])
	if err != nil {
		return nil, err
	}
	sk := blst.KeyGen(ikm[:])
	return &PrivKey{sk: sk}, nil
}

// Bytes returns the byte representation of the Key.
func (privKey *PrivKey) Bytes() []byte {
	return privKey.sk.Serialize()
}

// PubKey returns the private key's public key. If the privkey is not valid
// it returns a nil value.
func (privKey *PrivKey) PubKey() crypto.PubKey {
	return &PubKey{pk: new(blstPublicKey).From(privKey.sk)}
}

// Equals returns true if two keys are equal and false otherwise.
func (privKey *PrivKey) Equals(other crypto.PrivKey) bool {
	return privKey.Type() == other.Type() && bytes.Equal(privKey.Bytes(), other.Bytes())
}

// Type returns the type.
func (PrivKey) Type() string {
	return KeyType
}

// Sign signs the given byte array. If msg is larger than
// MaxMsgLen, SHA256 sum will be signed instead of the raw bytes.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		signature := new(blstSignature).Sign(privKey.sk, hash[:], dstMinSig)
		return signature.Compress(), nil
	}

	signature := new(blstSignature).Sign(privKey.sk, msg, dstMinSig)
	return signature.Compress(), nil
}

// Zeroize clears the private key.
func (privKey *PrivKey) Zeroize() {
	privKey.sk.Zeroize()
}

// ===============================================================================================
// Public Key
// ===============================================================================================

// Pubkey is a wrapper around the Ethereum BLS12-381 public key type. This
// wrapper conforms to crypto.Pubkey to allow for the use of the Ethereum
// BLS12-381 public key type.

var _ crypto.PubKey = &PubKey{}

type PubKey struct {
	pk *blstPublicKey
}

// Address returns the address of the key.
//
// The function will panic if the public key is invalid.
func (pubKey *PubKey) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey.pk.Serialize()))
}

// VerifySignature verifies the given signature.
func (pubKey *PubKey) VerifySignature(msg, sig []byte) bool {
	signature := new(blstSignature).Uncompress(sig)
	if signature == nil {
		return false
	}

	// Group check signature. Do not check for infinity since an aggregated signature
	// could be infinite.
	if !signature.SigValidate(false) {
		return false
	}

	if len(msg) > MaxMsgLen {
		hash := sha256.Sum256(msg)
		return signature.Verify(false, pubKey.pk, false, hash[:], dstMinSig)
	}

	return signature.Verify(false, pubKey.pk, false, msg, dstMinSig)
}

// Bytes returns the byte format.
func (pubKey *PubKey) Bytes() []byte {
	return pubKey.pk.Serialize()
}

// Type returns the key's type.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the other's type is the same and their bytes are deeply equal.
func (pubKey *PubKey) Equals(other crypto.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}
