package types

import (
	"github.com/cometbft/cometbft/crypto/bls12381"
	cmtmath "github.com/cometbft/cometbft/libs/math"
)

// MaxSignatureSize is a maximum allowed signature size for the Proposal
// and Vote.
// XXX: secp256k1 does not have Size nor MaxSize defined.
var MaxSignatureSize = cmtmath.MaxInt(bls12381.SignatureLength, 64)

// Signable is an interface for all signable things.
// It typically removes signatures before serializing.
// SignBytes returns the bytes to be signed
// NOTE: chainIDs are part of the SignBytes but not
// necessarily the object themselves.
// NOTE: Expected to panic if there is an error marshaling.
type Signable interface {
	SignBytes(chainID string) []byte
}
