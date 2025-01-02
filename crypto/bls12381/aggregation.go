//go:build bls12381

package bls12381

import (
	"errors"

	blst "github.com/supranational/blst/bindings/go"
)

// ErrAggregation is returned when aggregation fails.
var ErrorAggregation = errors.New("bls12381: failed to aggregate signatures")

// For minimal-signature-size operations.
type (
	blstAggregateSignature = blst.P1Aggregate
	blstAggregatePublicKey = blst.P2Aggregate
)

// AggregateSignatures aggregates the given compressed signatures.
func AggregateSignatures(sigs [][]byte) ([]byte, error) {
	var agProj blstAggregateSignature
	if !agProj.AggregateCompressed(sigsToAgg, false) {
		return nil, ErrAggregation
	}
	agSig := agProj.ToAffine()
	return agSig.Compress(), nil
}

// VerifyAggregateSignature verifies the given compressed aggregate signature.
func VerifyAggregateSignature(agSigCompressed []byte, pubks []*blstPublicKey, msg []byte) bool {
	agSig := new(blstAggregateSignature).Deserialize(agSigCompressed)
	if agSig == nil {
		return false
	}
	return agSig.FastAggregateVerify(false, pubks, msg, dstMinSig)
}
