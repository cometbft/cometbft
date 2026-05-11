//go:build !bls12381

package main

import "math/big"

const blsAvailable = false

func makeCanonicalBls(_ int, _ []uint64, _ int, _ []byte) canonicalBls {
	return canonicalBls{
		Available:      false,
		PubKeys:        []string{},
		PubKeysEip2537: []string{},
		Powers:         []int64{},
		Leaves:         []string{},
		Signatures:     []string{},
		SignerBitmap:   new(big.Int),
	}
}
