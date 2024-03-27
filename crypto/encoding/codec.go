package encoding

import (
	"fmt"

	pc "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/libs/json"
)

// ErrUnsupportedKey describes an error resulting from the use of an
// unsupported key in [PubKeyToProto] or [PubKeyFromProto].
type ErrUnsupportedKey struct {
	Key any
}

func (e ErrUnsupportedKey) Error() string {
	return fmt.Sprintf("encoding: unsupported key %v", e.Key)
}

// InvalidKeyLen describes an error resulting from the use of a key with
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
}

// PubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey.
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
	default:
		return kp, ErrUnsupportedKey{Key: k}
	}
	return kp, nil
}

// PubKeyFromProto takes a protobuf Pubkey and transforms it to a crypto.Pubkey.
func PubKeyFromProto(k pc.PublicKey) (crypto.PubKey, error) {
	switch k := k.Sum.(type) {
	case *pc.PublicKey_Ed25519:
		if len(k.Ed25519) != ed25519.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  k,
				Got:  len(k.Ed25519),
				Want: ed25519.PubKeySize,
			}
		}
		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, k.Ed25519)
		return pk, nil
	case *pc.PublicKey_Secp256K1:
		if len(k.Secp256K1) != secp256k1.PubKeySize {
			return nil, ErrInvalidKeyLen{
				Key:  k,
				Got:  len(k.Secp256K1),
				Want: secp256k1.PubKeySize,
			}
		}
		pk := make(secp256k1.PubKey, secp256k1.PubKeySize)
		copy(pk, k.Secp256K1)
		return pk, nil
	default:
		return nil, ErrUnsupportedKey{Key: k}
	}
}
