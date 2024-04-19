package types

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
)

func Ed25519ValidatorUpdate(pk []byte, power int64) ValidatorUpdate {
	return ValidatorUpdate{
		Power:       power,
		PubKeyBytes: pk,
		PubKeyType:  ed25519.KeyType,
	}
}

func UpdateValidator(pk []byte, power int64, keyType string) ValidatorUpdate {
	switch keyType {
	case "", ed25519.KeyType:
		return Ed25519ValidatorUpdate(pk, power)
	case secp256k1.KeyType:
		return ValidatorUpdate{
			Power:       power,
			PubKeyBytes: pk,
			PubKeyType:  keyType,
		}
	case bls12381.KeyType:
		return ValidatorUpdate{
			Power:       power,
			PubKeyBytes: pk,
			PubKeyType:  keyType,
		}
	default:
		panic(fmt.Sprintf("key type %s not supported", keyType))
	}
}

func PubKeyFromValidatorUpdate(v ValidatorUpdate) (crypto.PubKey, error) {
	var pubKey crypto.PubKey
	switch v.PubKeyType {
	case ed25519.KeyType:
		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, v.PubKeyBytes)
		pubKey = pk
	case secp256k1.KeyType:
		pk := make(secp256k1.PubKey, secp256k1.PubKeySize)
		copy(pk, v.PubKeyBytes)
		pubKey = pk
	case bls12381.KeyType:
		pk := make(bls12381.PubKey, bls12381.PubKeySize)
		copy(pk, v.PubKeyBytes)
		pubKey = pk
	default:
		return nil, fmt.Errorf("unknown pubkey type: %s", v.PubKeyType)
	}
	return pubKey, nil
}
