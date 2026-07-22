// Package secp256k1eth implements an Ethereum-compatible secp256k1 signature
// scheme for use as a CometBFT validator key type.
//
// Signing hashes the message with legacy Keccak-256 and emits a 65-byte
// signature [R || S || V] (V in {0,1}) in canonical lower-S form — byte-for-byte
// compatible with go-ethereum's pure-Go (non-CGO) signing path. Public keys are
// 33-byte compressed SEC1; addresses are the 20-byte Ethereum address
// Keccak256(uncompressedPubKey[1:])[12:].
package secp256k1eth

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/sha3"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

func init() {
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}

var _ crypto.PrivKey = PrivKey{}

// PrivKey is a 32-byte secp256k1 scalar.
type PrivKey []byte

// Bytes returns the raw 32-byte scalar.
func (privKey PrivKey) Bytes() []byte {
	return privKey
}

// PubKey returns the compressed SEC1 public key.
func (privKey PrivKey) PubKey() crypto.PubKey {
	priv := secp256k1.PrivKeyFromBytes(privKey)
	return PubKey(priv.PubKey().SerializeCompressed())
}

// Equals runs in constant time based on the length of the keys.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	if o, ok := other.(PrivKey); ok {
		return subtle.ConstantTimeCompare(privKey[:], o[:]) == 1
	}
	return false
}

// Type returns the key-type identifier.
func (PrivKey) Type() string {
	return KeyType
}

// GenPrivKey generates a new key using OS randomness.
func GenPrivKey() PrivKey {
	return genPrivKey(crypto.CReader())
}

func genPrivKey(rand io.Reader) PrivKey {
	var privKeyBytes [PrivKeySize]byte
	d := new(big.Int)
	for {
		privKeyBytes = [PrivKeySize]byte{}
		if _, err := io.ReadFull(rand, privKeyBytes[:]); err != nil {
			panic(err)
		}
		d.SetBytes(privKeyBytes[:])
		// Accept only valid field elements: 0 < d < curve order N.
		if 0 < d.Sign() && d.Cmp(secp256k1.Params().N) < 0 {
			break
		}
	}
	return privKeyBytes[:]
}

var one = new(big.Int).SetInt64(1)

// GenPrivKeySecp256k1Eth deterministically derives a key from secret bytes by
// hashing with SHA-256 and reducing into the valid scalar range. It mirrors
// secp256k1.GenPrivKeySecp256k1 and exists for reproducible e2e test keys — it
// is NOT HD/mnemonic ("seed phrase") derivation.
func GenPrivKeySecp256k1Eth(secret []byte) PrivKey {
	secHash := sha256.Sum256(secret)
	fe := new(big.Int).SetBytes(secHash[:])
	n := new(big.Int).Sub(secp256k1.Params().N, one)
	fe.Mod(fe, n)
	fe.Add(fe, one)

	feB := fe.Bytes()
	privKey32 := make([]byte, PrivKeySize)
	copy(privKey32[PrivKeySize-len(feB):PrivKeySize], feB)
	return privKey32
}

// keccak256 returns the legacy Keccak-256 digest used by go-ethereum. This is
// NOT FIPS-202 SHA3-256 (sha3.New256), which uses different padding and would
// break Ethereum compatibility.
func keccak256(msg []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write(msg) // hash.Hash never errors on Write
	return h.Sum(nil)
}

// Sign produces a 65-byte go-ethereum signature [R || S || V] (V in {0,1}) over
// the legacy Keccak-256 hash of msg, in canonical lower-S form. It returns an
// error if the private key is not a valid scalar in (0, N).
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	if len(privKey) != PrivKeySize {
		return nil, fmt.Errorf(
			"secp256k1eth: invalid private key size, expected %d bytes, got %d",
			PrivKeySize, len(privKey),
		)
	}
	var d secp256k1.ModNScalar
	if overflow := d.SetByteSlice(privKey); overflow || d.IsZero() {
		d.Zero()
		return nil, errors.New("secp256k1eth: private key scalar is not in the valid range (0, N)")
	}
	priv := secp256k1.NewPrivateKey(&d)
	defer priv.Zero()
	h := keccak256(msg)
	// decred returns compact form [recoveryByte || R || S] where the leading
	// byte is 27 + recoveryCode; go-ethereum's V is that recoveryCode (byte-27).
	sig := ecdsa.SignCompact(priv, h, false)
	// Convert to go-ethereum's [R || S || V] with V in {0,1}: drop the leading
	// recovery byte, append it (minus the 27 magic offset) at the end.
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[64] = v
	return sig, nil
}

// UnmarshalJSON unmarshals the private key from JSON, rejecting bytes of the
// wrong length so a malformed key errors at decode time instead of panicking
// later (e.g. during genesis validation).
func (privKey *PrivKey) UnmarshalJSON(bz []byte) error {
	var rawBytes []byte
	if err := json.Unmarshal(bz, &rawBytes); err != nil {
		return err
	}
	if len(rawBytes) != PrivKeySize {
		return fmt.Errorf(
			"secp256k1eth: invalid private key size, expected %d bytes, got %d",
			PrivKeySize, len(rawBytes),
		)
	}
	*privKey = rawBytes
	return nil
}

var _ crypto.PubKey = PubKey{}

// PubKey is the compressed SEC1 public key (33 bytes).
type PubKey []byte

// Address returns the 20-byte Ethereum address:
// Keccak256(uncompressedPubKey[1:])[12:].
func (pubKey PubKey) Address() crypto.Address {
	if len(pubKey) != PubKeySize {
		panic("length of pubkey is incorrect")
	}
	pub, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		panic(err)
	}
	return crypto.Address(addressFromPubKey(pub))
}

// Bytes returns the compressed public key bytes.
func (pubKey PubKey) Bytes() []byte {
	return pubKey
}

func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeySecp256k1eth{%X}", []byte(pubKey))
}

func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	if o, ok := other.(PubKey); ok {
		return bytes.Equal(pubKey[:], o[:])
	}
	return false
}

// Type returns the key-type identifier.
func (PubKey) Type() string {
	return KeyType
}

// VerifySignature verifies a go-ethereum signature over the legacy Keccak-256
// hash of msg. It accepts only the 65-byte [R || S || V] form produced by Sign,
// with V in {0,1}, and rejects malleable (non-lower-S) signatures.
func (pubKey PubKey) VerifySignature(msg []byte, sigStr []byte) bool {
	if len(sigStr) != SignatureSize || sigStr[64] > 1 {
		return false
	}

	var r secp256k1.ModNScalar
	var s secp256k1.ModNScalar
	if r.SetByteSlice(sigStr[:32]) || s.SetByteSlice(sigStr[32:64]) {
		return false
	}
	if r.IsZero() || s.IsZero() || s.IsOverHalfOrder() {
		return false
	}

	compact := make([]byte, SignatureSize)
	compact[0] = sigStr[64] + 27
	copy(compact[1:], sigStr[:64])
	recovered, _, err := ecdsa.RecoverCompact(compact, keccak256(msg))
	if err != nil {
		return false
	}
	return bytes.Equal(recovered.SerializeCompressed(), pubKey)
}

// NewPubKeyFromBytes validates the length and that the bytes are a valid
// compressed secp256k1 point, then returns a PubKey copy.
func NewPubKeyFromBytes(bz []byte) (PubKey, error) {
	if len(bz) != PubKeySize {
		return nil, fmt.Errorf(
			"secp256k1eth: invalid public key size, expected %d bytes, got %d",
			PubKeySize, len(bz),
		)
	}
	if _, err := secp256k1.ParsePubKey(bz); err != nil {
		return nil, fmt.Errorf("secp256k1eth: invalid public key: %w", err)
	}
	pk := make(PubKey, PubKeySize)
	copy(pk, bz)
	return pk, nil
}

// UnmarshalJSON unmarshals the public key from JSON, rejecting bytes that are
// not a valid compressed secp256k1 point so a malformed key errors at decode
// time instead of panicking later (e.g. during genesis validation).
func (pubKey *PubKey) UnmarshalJSON(bz []byte) error {
	var rawBytes []byte
	if err := json.Unmarshal(bz, &rawBytes); err != nil {
		return err
	}
	pk, err := NewPubKeyFromBytes(rawBytes)
	if err != nil {
		return err
	}
	*pubKey = pk
	return nil
}

func addressFromPubKey(pub *secp256k1.PublicKey) []byte {
	// SerializeUncompressed returns 65 bytes: 0x04 || X || Y. Drop the 0x04
	// prefix before hashing, matching go-ethereum's address derivation.
	hash := keccak256(pub.SerializeUncompressed()[1:])
	return hash[12:]
}
