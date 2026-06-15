package pkcs11

import (
	"fmt"

	"github.com/miekg/pkcs11"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
)

// ckmEDDSA is the standard PKCS#11 v3.0 EdDSA signing mechanism. miekg/pkcs11
// v1.1.x does not export it, so it is defined here against the spec value.
const ckmEDDSA = 0x00001057

// algoEd25519 is the config "algorithm" name for Ed25519 keys (the default).
const algoEd25519 = "ed25519"

// keyAlgo describes how one validator key algorithm maps onto PKCS#11: which
// signing mechanism to use, how to turn the token's public-key bytes into a
// crypto.PubKey, and how to normalize the raw signature the token returns.
//
// Adding a new key type (ml-dsa, secp256k1eth, ...) is a single new entry in
// algos: its mechanism, a decodePub, and (for ECDSA-family keys) a fixSig that
// converts the token's DER signature into the consensus wire format.
type keyAlgo struct {
	name      string
	mechanism func() []*pkcs11.Mechanism
	decodePub func(ckaECPoint []byte) (crypto.PubKey, error)
	fixSig    func(raw []byte) ([]byte, error)
}

// algos is the registry of supported key algorithms, keyed by the config
// "algorithm" string. Ed25519 is the only entry for now.
var algos = map[string]keyAlgo{
	algoEd25519: {
		name:      algoEd25519,
		mechanism: func() []*pkcs11.Mechanism { return []*pkcs11.Mechanism{pkcs11.NewMechanism(ckmEDDSA, nil)} },
		decodePub: decodeEd25519Pub,
		fixSig:    func(raw []byte) ([]byte, error) { return raw, nil },
	},
}

// decodeEd25519Pub turns a CKA_EC_POINT value into an ed25519 crypto.PubKey.
// PKCS#11 v3.0 encodes the point as a DER OCTET STRING wrapping the 32-byte key
// (0x04 0x20 <32 bytes>); some tokens return the raw 32 bytes. Both are accepted.
func decodeEd25519Pub(ckaECPoint []byte) (crypto.PubKey, error) {
	raw := ckaECPoint
	// DER OCTET STRING (tag 0x04) of length 0x20 (32) wrapping the key.
	if len(ckaECPoint) == ed25519.PubKeySize+2 && ckaECPoint[0] == 0x04 && ckaECPoint[1] == ed25519.PubKeySize {
		raw = ckaECPoint[2:]
	}
	if len(raw) != ed25519.PubKeySize {
		return nil, fmt.Errorf("ed25519 CKA_EC_POINT: expected %d-byte key, got %d bytes", ed25519.PubKeySize, len(raw))
	}
	pub := make(ed25519.PubKey, ed25519.PubKeySize)
	copy(pub, raw)
	return pub, nil
}
