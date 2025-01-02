//go:build bls12381

package bls12381

import (
	"errors"

	blst "github.com/supranational/blst/bindings/go"
)

// ErrAggregation is returned when aggregation fails.
var ErrAggregation = errors.New("bls12381: failed to aggregate signatures")

// For minimal-pubkey-size operations.
//
// Changing this to 'minimal-signature-size' would render CometBFT not Ethereum
// compatible.
type (
	blstAggregateSignature = blst.P2Aggregate
)

// AggregateSignatures aggregates the given compressed signatures.
//
// Does not group-check the signatures.
func AggregateSignatures(sigsToAgg [][]byte) ([]byte, error) {
	var agProj blstAggregateSignature
	if !agProj.AggregateCompressed(sigsToAgg, false) {
		return nil, ErrAggregation
	}
	agSig := agProj.ToAffine()
	return agSig.Compress(), nil
}

// VerifyAggregateSignature verifies the given compressed aggregate signature.
//
// Group-checks the signature.
func VerifyAggregateSignature(agSigCompressed []byte, pubks []*PubKey, msg []byte) bool {
	agSig := new(blstSignature).Uncompress(agSigCompressed)
	if agSig == nil {
		return false
	}
	blsPubKeys := make([]*blstPublicKey, len(pubks))
	for i, pubk := range pubks {
		blsPubKeys[i] = pubk.pk
	}
	return agSig.FastAggregateVerify(true, blsPubKeys, msg, dstMinPk)
}
