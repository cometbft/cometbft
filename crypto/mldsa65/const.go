package mldsa65

import "github.com/cloudflare/circl/sign/mldsa/mldsa65"

const (
	// PrivKeySize is the size, in bytes, of an ML-DSA-65 private key as
	// serialized by circl (FIPS 204 packed form).
	PrivKeySize = mldsa65.PrivateKeySize
	// PubKeySize is the size, in bytes, of an ML-DSA-65 public key.
	PubKeySize = mldsa65.PublicKeySize
	// SignatureSize is the size, in bytes, of an ML-DSA-65 signature.
	SignatureSize = mldsa65.SignatureSize
	// SeedSize is the size, in bytes, of the seed used to derive a key pair.
	SeedSize = mldsa65.SeedSize

	// KeyType is the string identifier for the ML-DSA-65 algorithm.
	KeyType = "ml_dsa_65"
	// PrivKeyName is the amino route for the private key.
	PrivKeyName = "cometbft/PrivKeyMlDsa65"
	// PubKeyName is the amino route for the public key.
	PubKeyName = "cometbft/PubKeyMlDsa65"
)
