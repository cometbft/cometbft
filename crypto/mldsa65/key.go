// Package mldsa65 implements the NIST ML-DSA-65 post-quantum signature
// scheme (FIPS 204) for use as a CometBFT validator key type.
//
// Keys and signatures are encoded as the packed byte form produced by
// github.com/cloudflare/circl/sign/mldsa/mldsa65. Signing is deterministic
// (randomized=false) and signs over an empty context string, matching the
// "pure" mode of ML-DSA.
package mldsa65

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

var (
	// ErrInvalidPrivKeySize is returned when private key bytes are the wrong length.
	ErrInvalidPrivKeySize = fmt.Errorf("mldsa65: invalid private key size, expected %d bytes", PrivKeySize)
	// ErrInvalidPubKeySize is returned when public key bytes are the wrong length.
	ErrInvalidPubKeySize = fmt.Errorf("mldsa65: invalid public key size, expected %d bytes", PubKeySize)
	// ErrDeserialization is returned when key bytes fail circl's UnmarshalBinary.
	ErrDeserialization = errors.New("mldsa65: deserialization error")
)

func init() {
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}

// ===============================================================================================
// Private Key
// ===============================================================================================

var _ crypto.PrivKey = &PrivKey{}

// PrivKey wraps a parsed ML-DSA-65 private key. The parsed form is retained so
// Sign and PubKey avoid re-deserializing the 4032-byte packed key on every
// call.
type PrivKey struct {
	sk *mldsa65.PrivateKey
}

// GenPrivKey generates a fresh ML-DSA-65 key using OS randomness.
func GenPrivKey() (PrivKey, error) {
	return genPrivKey(crypto.CReader())
}

func genPrivKey(r io.Reader) (PrivKey, error) {
	_, sk, err := mldsa65.GenerateKey(r)
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{sk: sk}, nil
}

// GenPrivKeyFromSeed deterministically derives a key from a 32-byte seed.
func GenPrivKeyFromSeed(seed []byte) (PrivKey, error) {
	if len(seed) != SeedSize {
		return PrivKey{}, fmt.Errorf("mldsa65: seed must be %d bytes, got %d", SeedSize, len(seed))
	}
	var s [SeedSize]byte
	copy(s[:], seed)
	_, sk := mldsa65.NewKeyFromSeed(&s)
	return PrivKey{sk: sk}, nil
}

// NewPrivKeyFromBytes validates and returns a PrivKey from the packed bytes.
func NewPrivKeyFromBytes(bz []byte) (PrivKey, error) {
	if len(bz) != PrivKeySize {
		return PrivKey{}, ErrInvalidPrivKeySize
	}
	sk := new(mldsa65.PrivateKey)
	if err := sk.UnmarshalBinary(bz); err != nil {
		return PrivKey{}, ErrDeserialization
	}
	return PrivKey{sk: sk}, nil
}

// Bytes returns the packed private key bytes.
func (privKey PrivKey) Bytes() []byte {
	if privKey.sk == nil {
		return nil
	}
	bz, err := privKey.sk.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("mldsa65: marshal privkey: %v", err))
	}
	return bz
}

// PubKey returns the corresponding public key.
func (privKey PrivKey) PubKey() crypto.PubKey {
	pk, ok := privKey.sk.Public().(*mldsa65.PublicKey)
	if !ok {
		panic("mldsa65: unexpected public key type")
	}
	return PubKey{pk: pk}
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	otherKey, ok := other.(PrivKey)
	if !ok {
		return false
	}
	return bytes.Equal(privKey.Bytes(), otherKey.Bytes())
}

// Type returns the algorithm identifier.
func (PrivKey) Type() string {
	return KeyType
}

// Sign produces a deterministic ML-DSA-65 signature with an empty context.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	sig := make([]byte, SignatureSize)
	if err := mldsa65.SignTo(privKey.sk, msg, nil, false, sig); err != nil {
		return nil, err
	}
	return sig, nil
}

// MarshalJSON marshals the private key to JSON as its packed bytes.
//
// XXX: Not a pointer because our JSON encoder (libs/json) does not correctly
// handle pointers.
func (privKey PrivKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(privKey.Bytes())
}

// UnmarshalJSON unmarshals the private key from JSON.
func (privKey *PrivKey) UnmarshalJSON(bz []byte) error {
	var rawBytes []byte
	if err := json.Unmarshal(bz, &rawBytes); err != nil {
		return err
	}
	pk, err := NewPrivKeyFromBytes(rawBytes)
	if err != nil {
		return err
	}
	privKey.sk = pk.sk
	return nil
}

// ===============================================================================================
// Public Key
// ===============================================================================================

var _ crypto.PubKey = &PubKey{}

// PubKey wraps a parsed ML-DSA-65 public key. The parsed form is retained so
// VerifySignature avoids re-deserializing the 1952-byte packed key on every
// call.
type PubKey struct {
	pk *mldsa65.PublicKey
}

// NewPubKeyFromBytes validates and returns a PubKey from the packed bytes.
func NewPubKeyFromBytes(bz []byte) (PubKey, error) {
	if len(bz) != PubKeySize {
		return PubKey{}, ErrInvalidPubKeySize
	}
	pk := new(mldsa65.PublicKey)
	if err := pk.UnmarshalBinary(bz); err != nil {
		return PubKey{}, ErrDeserialization
	}
	return PubKey{pk: pk}, nil
}

// Address is SHA256(pubkey) truncated to 20 bytes, matching the convention used
// by ed25519 and bls12381 validator keys.
func (pubKey PubKey) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey.Bytes()))
}

// VerifySignature verifies the signature against msg with an empty context.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureSize || pubKey.pk == nil {
		return false
	}
	return mldsa65.Verify(pubKey.pk, msg, nil, sig)
}

// Bytes returns the packed public key bytes.
func (pubKey PubKey) Bytes() []byte {
	if pubKey.pk == nil {
		return nil
	}
	bz, err := pubKey.pk.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("mldsa65: marshal pubkey: %v", err))
	}
	return bz
}

// Type returns the algorithm identifier.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	otherKey, ok := other.(PubKey)
	if !ok {
		return false
	}
	return bytes.Equal(pubKey.Bytes(), otherKey.Bytes())
}

// String returns a hex-formatted representation of the public key.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyMlDsa65{%X}", pubKey.Bytes())
}

// MarshalJSON marshals the public key to JSON as its packed bytes.
//
// XXX: Not a pointer because our JSON encoder (libs/json) does not correctly
// handle pointers.
func (pubKey PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(pubKey.Bytes())
}

// UnmarshalJSON unmarshals the public key from JSON.
func (pubKey *PubKey) UnmarshalJSON(bz []byte) error {
	var rawBytes []byte
	if err := json.Unmarshal(bz, &rawBytes); err != nil {
		return err
	}
	pk, err := NewPubKeyFromBytes(rawBytes)
	if err != nil {
		return err
	}
	pubKey.pk = pk.pk
	return nil
}
