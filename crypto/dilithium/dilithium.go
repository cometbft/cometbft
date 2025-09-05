package dilithium

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"

	dilithium2 "github.com/cloudflare/circl/sign/dilithium/mode2"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmjson "github.com/cometbft/cometbft/libs/json"
)

//-------------------------------------

var _ crypto.PrivKey = PrivKey{}

const (
	PrivKeyName = "cometbft/PrivKeyDilithium"
	PubKeyName  = "cometbft/PubKeyDilithium"

	// PubKeySize is is the size, in bytes, of public keys as used in this package.
	PubKeySize = 1312
	// PrivateKeySize is the size, in bytes, of private keys as used in this package.
	PrivateKeySize = 2528
	// Size of an Edwards25519 signature. Namely the size of a compressed
	// Edwards25519 point, and a field element. Both of which are 32 bytes.
	SignatureSize = 2420
	// SeedSize is the size, in bytes, of private key seeds. These are the
	// private key representations used by RFC 8032.
	SeedSize = 32

	KeyType = "dilithium2"
)

func init() {
	tmjson.RegisterType(PubKey{}, PubKeyName)
	tmjson.RegisterType(PrivKey{}, PrivKeyName)
}

// PrivKey implements crypto.PrivKey.
type PrivKey []byte

// Bytes returns the privkey byte format.
func (privKey PrivKey) Bytes() []byte {
	return []byte(privKey)
}

// Sign produces a signature on the provided message.
func (privKey PrivKey) Sign(msg []byte) ([]byte, error) {
	signatureBytes := make([]byte, SignatureSize)

	var dil2PrivKey dilithium2.PrivateKey
	dil2PrivKey.Unpack((*[PrivateKeySize]byte)(privKey))

	dilithium2.SignTo(&dil2PrivKey, msg, signatureBytes)

	return signatureBytes, nil
}

// PubKey gets the corresponding public key from the private key.
//
// Panics if the private key is not initialized.
func (privKey PrivKey) PubKey() crypto.PubKey {
	var dil2PrivKey dilithium2.PrivateKey
	dil2PrivKey.Unpack((*[PrivateKeySize]byte)(privKey))

	return PubKey(dil2PrivKey.Public().(*dilithium2.PublicKey).Bytes())
}

func (privKey PrivKey) Type() string {
	return KeyType
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey PrivKey) Equals(other crypto.PrivKey) bool {
	if otherDil, ok := other.(PrivKey); ok {
		return subtle.ConstantTimeCompare(privKey[:], otherDil[:]) == 1
	}

	return false
}

func GenPrivKey() PrivKey {
	return genPrivKey(crypto.CReader())
}

// genPrivKey generates a new Dilithium private key using the provided reader.
func genPrivKey(rand io.Reader) PrivKey {
	_, sk, err := dilithium2.GenerateKey(rand)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate Dilithium key pair: %v", err))
	}
	return PrivKey(sk.Bytes())
}

// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSeed(secret []byte) PrivKey {
	seed := crypto.Sha256(secret) // Not Ripemd160 because we want 32 bytes.
	if len(seed) != 32 {
		panic("seed must be exactly 32 bytes")
	}

	var seedArray [32]byte
	copy(seedArray[:], seed)

	_, sk := dilithium2.NewKeyFromSeed(&seedArray)
	return PrivKey(sk.Bytes())
}

var _ crypto.PubKey = PubKey{}

type PubKey []byte

// Address is the SHA256-20 of the raw pubkey bytes.
func (pubKey PubKey) Address() crypto.Address {
	if len(pubKey) != PubKeySize {
		panic("pubkey is incorrect size")
	}
	return crypto.Address(tmhash.SumTruncated(pubKey))
}

// Bytes returns the PubKey byte format.
func (pubKey PubKey) Bytes() []byte {
	return []byte(pubKey)
}

func (pubKey PubKey) VerifySignature(msg []byte, sig []byte) bool {
	// make sure we use the same algorithm to sign
	if len(sig) != SignatureSize {
		return false
	}

	var dil2PubKey dilithium2.PublicKey
	dil2PubKey.Unpack((*[PubKeySize]byte)(pubKey))

	return dilithium2.Verify(&dil2PubKey, msg, sig)
}

func (pubKey PubKey) String() string {
	return fmt.Sprintf("PubKeyDil2{%X}", []byte(pubKey))
}

func (pubKey PubKey) Type() string {
	return KeyType
}

func (pubKey PubKey) Equals(other crypto.PubKey) bool {
	if otherDil, ok := other.(PubKey); ok {
		return bytes.Equal(pubKey[:], otherDil[:])
	}

	return false
}
