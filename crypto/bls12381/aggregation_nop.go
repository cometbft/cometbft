//go:build !bls12381

package bls12381

import "errors"

// AggregateSignatures is a nop.
func AggregateSignatures([][]byte) ([]byte, error) {
	return nil, errors.New("bls12381 is disabled")
}

// VerifyAggregateSignature is a nop.
func VerifyAggregateSignature([]byte, []*PubKey, []byte) bool {
	return false
}
