//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && bls12381

package blst

import blst "github.com/supranational/blst/bindings/go"

// Internal types for blst.
type blstPublicKey = blst.P1Affine
type blstSignature = blst.P2Affine
