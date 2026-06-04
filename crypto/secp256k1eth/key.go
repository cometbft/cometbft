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
// the legacy Keccak-256 hash of msg, in canonical lower-S form.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	priv := secp256k1.PrivKeyFromBytes(privKey)
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
	// SerializeUncompressed returns 65 bytes: 0x04 || X || Y. Drop the 0x04
	// prefix before hashing, matching go-ethereum's address derivation.
	hash := keccak256(pub.SerializeUncompressed()[1:])
	return crypto.Address(hash[12:])
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
// hash of msg. It accepts the 65-byte [R || S || V] form produced by Sign or a
// 64-byte [R || S] form, and rejects malleable (non-lower-S) signatures. Only
// R and S are used for verification; the recovery byte V (when present) is not
// consulted, matching go-ethereum's crypto.VerifySignature which takes [R || S].
func (pubKey PubKey) VerifySignature(msg []byte, sigStr []byte) bool {
	if len(sigStr) != 64 && len(sigStr) != SignatureSize {
		return false
	}
	pub, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return false
	}

	sig := signatureFromBytes(sigStr[:64])
	// Reject malleable signatures: decred does not enforce low-S but
	// libsecp256k1 (and thus go-ethereum's Sign output) does. Reject directly
	// if S is in the upper half of the group order.
	var s secp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	if s.IsOverHalfOrder() {
		return false
	}

	return sig.Verify(keccak256(msg), pub)
}

// signatureFromBytes reads an ECDSA signature from R || S. The caller must
// ensure len(sigStr) == 64.
func signatureFromBytes(sigStr []byte) *ecdsa.Signature {
	var r secp256k1.ModNScalar
	r.SetByteSlice(sigStr[:32])
	var s secp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	return ecdsa.NewSignature(&r, &s)
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
