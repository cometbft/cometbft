// Package crypto defines common cryptographic interfaces and utilities used by
// CometBFT and its subpackages (e.g. ed25519, secp256k1, bls12381).

// Keys

// Key generation functions in subpackages return values that implement the
// PrivKey interface, which defines:

//     Bytes() []byte
//     Sign(msg []byte) ([]byte, error)
//     PubKey() PubKey
//     Equals(PrivKey) bool
//     Type() string

// The corresponding public key implements the PubKey interface:

//     Address() Address
//     Bytes() []byte
//     VerifySignature(msg []byte, sig []byte) bool
//     Equals(PubKey) bool
//     Type() string

// Basic usage

//     // Generate a key (ed25519 does not return an error).
//     //   other subpackages exist, e.g. secp256k1.GenPrivKey().
//     privKey := ed25519.GenPrivKey()

//     pubKey := privKey.PubKey()
//     msg := []byte("hello")
//     sig, _ := privKey.Sign(msg)
//     ok := pubKey.VerifySignature(msg, sig)
//     _ = ok

// Note: some algorithms (e.g. BLS12-381) may expose generators that return an
// error; consult the respective subpackage docs.

// Addresses

// Package crypto defines type Address (a hex-encoded byte slice) and helpers
// like AddressHash. Address size is tied to tmhash.TruncatedSize. For secp256k1
// specifically, addresses are RIPEMD160(SHA256(pubkey)).

// Hashing utilities

// A thin wrapper is provided for SHA-256:

//     sum := crypto.Sha256([]byte("This is CometBFT"))
//     _ = sum

package crypto
