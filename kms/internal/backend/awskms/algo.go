package awskms

import (
	"crypto/ed25519"
	"crypto/x509"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"

	"github.com/cometbft/cometbft/crypto"
	cometed25519 "github.com/cometbft/cometbft/crypto/ed25519"
)

// algoEd25519 is the config "algorithm" name for Ed25519 keys (the default).
const algoEd25519 = "ed25519"

// keyAlgo describes how one validator key algorithm maps onto AWS KMS: which key
// spec the KMS key must have, which signing algorithm to request, how to turn
// the DER SubjectPublicKeyInfo that GetPublicKey returns into a crypto.PubKey,
// and how to normalize the signature KMS returns.
//
// Adding a new key type (secp256k1, ml-dsa, ...) is a single new entry in algos:
// its key spec, its signing algorithm, a decodePub, and — for ECDSA-family keys
// — a fixSig that DER-decodes the (r,s) signature, normalizes s to low-S, and
// emits the 64-byte r||s consensus wire form.
type keyAlgo struct {
	name      string
	keySpec   types.KeySpec
	signAlgo  types.SigningAlgorithmSpec
	decodePub func(spki []byte) (crypto.PubKey, error)
	// fixSig converts the raw signature KMS returns into the consensus wire
	// format. Ed25519 KMS signatures are already raw 64-byte R||S, so it is the
	// identity; ECDSA-family keys will DER-decode (r,s), apply low-S, and emit
	// 64-byte r||s here.
	fixSig func(raw []byte) ([]byte, error)
}

// algos is the registry of supported key algorithms, keyed by the config
// "algorithm" string. Ed25519 is the only entry for now.
//
// Ed25519 uses the ECC_NIST_EDWARDS25519 key spec with the ED25519_SHA_512
// signing algorithm and MessageType=RAW, which is standard RFC 8032 PureEd25519
// over the raw message — identical to the softsign/pkcs11 backends. The
// signature is a fixed raw 64 bytes, so fixSig is the identity.
var algos = map[string]keyAlgo{
	algoEd25519: {
		name:      algoEd25519,
		keySpec:   types.KeySpecEccNistEdwards25519,
		signAlgo:  types.SigningAlgorithmSpecEd25519Sha512,
		decodePub: decodeEd25519Pub,
		fixSig:    func(raw []byte) ([]byte, error) { return raw, nil },
	},
}

// decodeEd25519Pub turns the DER SubjectPublicKeyInfo returned by KMS
// GetPublicKey into an ed25519 crypto.PubKey.
func decodeEd25519Pub(spki []byte) (crypto.PubKey, error) {
	parsed, err := x509.ParsePKIXPublicKey(spki)
	if err != nil {
		return nil, fmt.Errorf("parse SubjectPublicKeyInfo: %w", err)
	}
	edPub, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected ed25519 public key, got %T", parsed)
	}
	if len(edPub) != cometed25519.PubKeySize {
		return nil, fmt.Errorf("ed25519 public key: expected %d bytes, got %d", cometed25519.PubKeySize, len(edPub))
	}
	pub := make(cometed25519.PubKey, cometed25519.PubKeySize)
	copy(pub, edPub)
	return pub, nil
}
