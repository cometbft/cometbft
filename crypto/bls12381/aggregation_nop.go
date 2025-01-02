//go:build !bls12381

package bls12381

import "errors"

// AggregateSignatures aggregates the given compressed signatures.
func AggregateSignatures([][]byte) ([]byte, error) {
	return nil, errors.New("bls12381 is disabled")
}

// // VerifyAggregateSignature verifies the given compressed aggregate signature.
// func VerifyAggregateSignature([]byte, []*blstPublicKey, []byte) bool {
// 	return false
// }
