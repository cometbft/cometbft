package encoding

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/libs/json"
	pc "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

// ErrUnsupportedKey describes an error resulting from the use of an
// unsupported key in [PubKeyToProto] or [PubKeyFromProto].
type ErrUnsupportedKey struct {
	Key any
}

func (e ErrUnsupportedKey) Error() string {
	return fmt.Sprintf("encoding: unsupported key %v", e.Key)
}

// ErrInvalidKeyLen describes an error resulting from the use of a key with
// an invalid length in [PubKeyFromProto].
type ErrInvalidKeyLen struct {
	Key       any
	Got, Want int
}

func (e ErrInvalidKeyLen) Error() string {
	return fmt.Sprintf("encoding: invalid key length for %v, got %d, want %d", e.Key, e.Got, e.Want)
}

func init() {
	json.RegisterType((*pc.PublicKey)(nil), "tendermint.crypto.PublicKey")
	json.RegisterType((*pc.PublicKey_Ed25519)(nil), "tendermint.crypto.PublicKey_Ed25519")
	json.RegisterType((*pc.PublicKey_Secp256K1)(nil), "tendermint.crypto.PublicKey_Secp256K1")
	if bls12381.Enabled {
		json.RegisterType((*pc.PublicKey_Bls12381)(nil), "tendermint.crypto.PublicKey_Bls12381")
	}
}

// PubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey
func PubKeyToProto(k crypto.PubKey) (pc.PublicKey, error) {
	var kp pc.PublicKey
	switch k := k.(type) {
	case ed25519.PubKey:
		kp = pc.PublicKey{
			Sum: &pc.PublicKey_Ed25519{
				Ed25519: k,
			},
		}
	case secp256k1.PubKey:
		kp = pc.PublicKey{
			Sum: &pc.PublicKey_Secp256K1{
				Secp256K1: k,
			},
		}
	case bls12381.PubKey:
		if !bls12381.Enabled {
			return kp, ErrUnsupportedKey{Key: k}
		}

		kp = pc.PublicKey{
			Sum: &pc.PublicKey_Bls12381{
				Bls12381: k.Bytes(),
			},
		}
	default:
		return kp, ErrUnsupportedKey{Key: k}
	}
	return kp, nil
}

// PubKeyFromProto takes a protobuf Pubkey and transforms it to a crypto.Pubkey
func PubKeyFromProto(k pc.PublicKey) (crypto.PubKey, error) {
	switch k := k.Sum.(type) {
	case *pc.PublicKey_Ed25519:
		if len(k.Ed25519) != ed25519.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeyEd25519. Got %d, expected %d",
				len(k.Ed25519), ed25519.PubKeySize)
		}
		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, k.Ed25519)
		return pk, nil
	case *pc.PublicKey_Secp256K1:
		if len(k.Secp256K1) != secp256k1.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeySecp256k1. Got %d, expected %d",
				len(k.Secp256K1), secp256k1.PubKeySize)
		}
		pk := make(secp256k1.PubKey, secp256k1.PubKeySize)
		copy(pk, k.Secp256K1)
		return pk, nil
	case *pc.PublicKey_Bls12381:
		if !bls12381.Enabled {
			return nil, ErrUnsupportedKey{Key: k}
		}

		if len(k.Bls12381) != bls12381.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  k,
				Got:  len(k.Bls12381),
				Want: bls12381.PubKeySize,
			}
		}
		return bls12381.NewPublicKeyFromBytes(k.Bls12381)
	default:
		return nil, ErrUnsupportedKey{Key: k}
	}
}

// PubKeyFromTypeAndBytes builds a crypto.PubKey from the given type
// and bytes. It returns ErrUnsupportedKey if the pubkey type is
// unsupported.
func PubKeyFromTypeAndBytes(pkType string, bytes []byte) (crypto.PubKey, error) {
	var pubKey crypto.PubKey
	switch pkType {
	case ed25519.KeyType:
		if len(bytes) != ed25519.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  pkType,
				Got:  len(bytes),
				Want: ed25519.PubKeySize,
			}
		}

		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, bytes)
		pubKey = pk
	case secp256k1.KeyType:
		if len(bytes) != secp256k1.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  pkType,
				Got:  len(bytes),
				Want: secp256k1.PubKeySize,
			}
		}

		pk := make(secp256k1.PubKey, secp256k1.PubKeySize)
		copy(pk, bytes)
		pubKey = pk
	case bls12381.KeyType:
		if !bls12381.Enabled {
			return nil, ErrUnsupportedKey{Key: pkType}
		}

		if len(bytes) != bls12381.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  pkType,
				Got:  len(bytes),
				Want: bls12381.PubKeySize,
			}
		}

		return bls12381.NewPublicKeyFromBytes(bytes)
	default:
		return nil, ErrUnsupportedKey{Key: pkType}
	}
	return pubKey, nil
}
