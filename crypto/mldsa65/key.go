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
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

// cacheSize is the number of parsed public keys retained in cachingVerifier.
// Matches crypto/ed25519's cacheSize so a validator set of typical size fits
// comfortably with room for incoming peers and recent rotations.
const cacheSize = 4096

// cachingVerifier holds parsed *mldsa65.PublicKey instances keyed by their
// 1952-byte packed form. Signature verification under a previously-seen
// validator pubkey skips the expensive UnmarshalBinary step, mirroring the
// crypto/ed25519 package's cachingVerifier pattern.
var cachingVerifier = newPubKeyCache(cacheSize)

type pubKeyCache struct {
	cache *lru.Cache[string, *mldsa65.PublicKey]
}

func newPubKeyCache(size int) *pubKeyCache {
	c, err := lru.New[string, *mldsa65.PublicKey](size)
	if err != nil {
		// lru.New only errors on size <= 0, which is a programmer error.
		panic(fmt.Sprintf("mldsa65: cannot create pubkey cache: %v", err))
	}
	return &pubKeyCache{cache: c}
}

// publicKey returns the parsed *mldsa65.PublicKey for the given packed bytes,
// populating the cache on first sight. Returns (nil, false) if the bytes are
// not a valid ML-DSA-65 public key.
func (c *pubKeyCache) publicKey(packed []byte) (*mldsa65.PublicKey, bool) {
	key := string(packed)
	if pk, ok := c.cache.Get(key); ok {
		return pk, true
	}
	pk := new(mldsa65.PublicKey)
	if err := pk.UnmarshalBinary(packed); err != nil {
		return nil, false
	}
	c.cache.Add(key, pk)
	return pk, true
}

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

var _ crypto.PrivKey = PrivKey{}

// PrivKey is the packed ML-DSA-65 private key (FIPS 204 form).
type PrivKey []byte

// GenPrivKey generates a fresh ML-DSA-65 key using OS randomness.
func GenPrivKey() (PrivKey, error) {
	return genPrivKey(rand.Reader)
}

func genPrivKey(r io.Reader) (PrivKey, error) {
	_, sk, err := mldsa65.GenerateKey(r)
	if err != nil {
		return nil, err
	}
	bz, err := sk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return PrivKey(bz), nil
}

// GenPrivKeyFromSeed deterministically derives a key from a 32-byte seed.
func GenPrivKeyFromSeed(seed []byte) (PrivKey, error) {
	if len(seed) != SeedSize {
		return nil, fmt.Errorf("mldsa65: seed must be %d bytes, got %d", SeedSize, len(seed))
	}
	var s [SeedSize]byte
	copy(s[:], seed)
	_, sk := mldsa65.NewKeyFromSeed(&s)
	bz, err := sk.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return PrivKey(bz), nil
}

// NewPrivKeyFromBytes validates and returns a PrivKey from the packed bytes.
func NewPrivKeyFromBytes(bz []byte) (PrivKey, error) {
	if len(bz) != PrivKeySize {
		return nil, ErrInvalidPrivKeySize
	}
	sk := new(mldsa65.PrivateKey)
	if err := sk.UnmarshalBinary(bz); err != nil {
		return nil, ErrDeserialization
	}
	out := make(PrivKey, PrivKeySize)
	copy(out, bz)
	return out, nil
}

// Bytes returns the packed private key bytes.
func (privKey PrivKey) Bytes() []byte {
	return []byte(privKey)
}

// PubKey returns the corresponding public key.
//
// Panics if the underlying private key bytes cannot be parsed; callers
// constructing a PrivKey via NewPrivKeyFromBytes or GenPrivKey will never hit
// this path.
func (privKey PrivKey) PubKey() crypto.PubKey {
	sk := new(mldsa65.PrivateKey)
	if err := sk.UnmarshalBinary(privKey); err != nil {
		panic(fmt.Sprintf("mldsa65: invalid private key: %v", err))
	}
	pk, ok := sk.Public().(*mldsa65.PublicKey)
	if !ok {
		panic("mldsa65: unexpected public key type")
	}
	pkBytes, err := pk.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("mldsa65: marshal pubkey: %v", err))
	}
	return PubKey(pkBytes)
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	if otherKey, ok := other.(PrivKey); ok {
		return bytes.Equal(privKey, otherKey)
	}
	return false
}

// Type returns the algorithm identifier.
func (PrivKey) Type() string {
	return KeyType
}

// Sign produces a deterministic ML-DSA-65 signature with an empty context.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	sk := new(mldsa65.PrivateKey)
	if err := sk.UnmarshalBinary(privKey); err != nil {
		return nil, fmt.Errorf("mldsa65: invalid private key: %w", err)
	}
	sig := make([]byte, SignatureSize)
	if err := mldsa65.SignTo(sk, msg, nil, false, sig); err != nil {
		return nil, err
	}
	return sig, nil
}

// ===============================================================================================
// Public Key
// ===============================================================================================

var _ crypto.PubKey = PubKey{}

// PubKey is the packed ML-DSA-65 public key (FIPS 204 form).
type PubKey []byte

// NewPubKeyFromBytes validates and returns a PubKey from the packed bytes.
// As a side effect the parsed *mldsa65.PublicKey is inserted into
// cachingVerifier, so subsequent VerifySignature calls skip the parse step.
func NewPubKeyFromBytes(bz []byte) (PubKey, error) {
	if len(bz) != PubKeySize {
		return nil, ErrInvalidPubKeySize
	}
	if _, ok := cachingVerifier.publicKey(bz); !ok {
		return nil, ErrDeserialization
	}
	out := make(PubKey, PubKeySize)
	copy(out, bz)
	return out, nil
}

// Address is SHA256(pubkey) truncated to 20 bytes, matching the convention used
// by ed25519 and bls12381 validator keys.
func (pubKey PubKey) Address() crypto.Address {
	if len(pubKey) != PubKeySize {
		panic("mldsa65: pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKey))
}

// VerifySignature verifies the signature against msg with an empty context.
// The parsed *mldsa65.PublicKey is fetched from cachingVerifier (or parsed
// and inserted on miss), so verification under a recently-seen pubkey skips
// the UnmarshalBinary step.
func (pubKey PubKey) VerifySignature(msg, sig []byte) bool {
	if len(sig) != SignatureSize || len(pubKey) != PubKeySize {
		return false
	}
	pk, ok := cachingVerifier.publicKey(pubKey)
	if !ok {
		return false
	}
	return mldsa65.Verify(pk, msg, nil, sig)
}

// Bytes returns the packed public key bytes.
func (pubKey PubKey) Bytes() []byte {
	return []byte(pubKey)
}

// Type returns the algorithm identifier.
func (PubKey) Type() string {
	return KeyType
}

// Equals returns true if the other key is also ML-DSA-65 and the bytes match.
func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	if otherKey, ok := other.(PubKey); ok {
		return bytes.Equal(pubKey, otherKey)
	}
	return false
}

// String returns a hex-formatted representation of the public key.
func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyMlDsa65{%X}", []byte(pubKey))
}
