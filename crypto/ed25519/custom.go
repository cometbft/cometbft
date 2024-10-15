package ed25519

import (
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/custom/indicator"
	cmtjson "github.com/cometbft/cometbft/libs/json"
)

// CustomOptions is a compatibility struct because the CometBFT ed25519 library exposes these variables.
// Consider these obsolete and at one point removed. Set them when you implement a new custom crypto library
// but do not use them in your codebase as they will be removed in the future.
type CustomOptions struct {
	PubKeyName     string
	PrivKeyName    string
	PrivateKeySize int
	PubKeySize     int
	KeyType        string
}

// CustomPrivKey is a PrivKey interface that supersedes the ed25519 library implementation.
type CustomPrivKey interface {
	crypto.PrivKey
	With(privKey PrivKey) CustomPrivKey
	GenPrivKey() PrivKey
	GenPrivKeyFromSecret(secret []byte) PrivKey
}

// CustomPubKey is a PubKey interface that supersedes the ed25519 library implementation.
type CustomPubKey interface {
	crypto.PubKey
	With(pubKey PubKey) CustomPubKey
	String() string
}

type CustomBatchVerifier interface {
	crypto.BatchVerifier
	With(batchVerifier BatchVerifier) CustomBatchVerifier
}

func RegisterCustomCrypto(privKey CustomPrivKey, pubKey CustomPubKey, batchVerifier CustomBatchVerifier, options CustomOptions) {
	if !indicator.IsCustomized() {
		panic("cannot register custom crypto library")
	}
	if options.PubKeyName == "" {
		panic("cannot register custom crypto library without PubKeyName")
	}
	PubKeyName = options.PubKeyName
	if options.PrivKeyName == "" {
		panic("cannot register custom crypto library without PrivKeyName")
	}
	PrivKeyName = options.PrivKeyName
	if options.PrivateKeySize == 0 {
		panic("cannot register custom crypto library without PrivateKeySize")
	}
	PrivateKeySize = options.PrivateKeySize
	if options.PubKeySize == 0 {
		panic("cannot register custom crypto library without PubKeySize")
	}
	PubKeySize = options.PubKeySize
	if options.KeyType == "" {
		KeyType = privKey.Type()
	}
	if KeyType != privKey.Type() {
		panic("cannot register custom crypto library with different KeyType")
	}
	KeyType = options.KeyType
	customPrivKey = privKey
	customPubKey = pubKey
	customBatchVerifier = batchVerifier

	// imitate the suppressed init function in ed25519 with the new parameters
	cmtjson.RegisterType(PubKey{}, PubKeyName)
	cmtjson.RegisterType(PrivKey{}, PrivKeyName)
}
