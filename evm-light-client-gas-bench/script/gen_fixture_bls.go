//go:build bls12381

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	cmtbls "github.com/cometbft/cometbft/crypto/bls12381"
	cmttypes "github.com/cometbft/cometbft/types"
	blst "github.com/supranational/blst/bindings/go"
)

const blsAvailable = true

// dstMinPk is the IETF hash-to-curve domain separation tag used by cometbft
// when signing with BLS12-381. It is duplicated here because cometbft does
// not export it. Any change here MUST mirror crypto/bls12381/key_bls12381.go.
var dstMinPk = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_NUL_")

func makeCanonicalBls(count int, powers []uint64, signerCount int, message []byte) canonicalBls {
	pubKeys := make([]string, count)
	pubKeysEip2537 := make([]string, count)
	leaves := make([][]byte, count)
	leavesHex := make([]string, count)
	int64Powers := make([]int64, count)
	cmtVals := make([]*cmttypes.Validator, count)

	privs := make([]*cmtbls.PrivKey, count)
	pubKeysG1Affine := make([]*blst.P1Affine, count)
	for i := 0; i < count; i++ {
		seed := sha256.Sum256([]byte(fmt.Sprintf("canonical-bls %d", i)))
		priv, err := cmtbls.GenPrivKeyFromSecret(seed[:])
		must(err)
		privs[i] = priv

		pub := priv.PubKey()
		int64Powers[i] = int64(powers[i])
		pubKeys[i] = "0x" + hex.EncodeToString(pub.Bytes())

		// cometbft's PubKey.Bytes() returns blst's uncompressed 96-byte G1 form
		// (Serialize, not Compress). Use Deserialize accordingly. KeyValidate
		// is used to confirm we got a valid affine point.
		pubAffine := new(blst.P1Affine).Deserialize(pub.Bytes())
		if pubAffine == nil || !pubAffine.KeyValidate() {
			panic(fmt.Sprintf("invalid BLS pubkey at %d", i))
		}
		pubKeysG1Affine[i] = pubAffine
		pubKeysEip2537[i] = "0x" + hex.EncodeToString(g1AffineToEip2537(pubAffine))

		v := cmttypes.NewValidator(pub, int64Powers[i])
		cmtVals[i] = v
		leaf := v.Bytes()
		leaves[i] = leaf
		leavesHex[i] = "0x" + hex.EncodeToString(leaf)
	}

	signatures := make([]string, signerCount)
	bitmap := new(big.Int)
	signerSigsAffine := make([]*blst.P2Affine, signerCount)
	signerPubsAffine := make([]*blst.P1Affine, signerCount)
	for i := 0; i < signerCount; i++ {
		bitmap.SetBit(bitmap, i, 1)
		sig, err := privs[i].Sign(message)
		must(err)
		signatures[i] = "0x" + hex.EncodeToString(sig)

		sigAffine := new(blst.P2Affine).Uncompress(sig)
		if sigAffine == nil || !sigAffine.SigValidate(false) {
			panic(fmt.Sprintf("invalid BLS signature at %d", i))
		}
		if !sigAffine.Verify(false, pubKeysG1Affine[i], true, message, dstMinPk) {
			panic(fmt.Sprintf("BLS signature verification failed at %d", i))
		}
		signerSigsAffine[i] = sigAffine
		signerPubsAffine[i] = pubKeysG1Affine[i]
	}

	hash := mustValidatorSetHash(cmtVals)

	// Aggregate signatures (in G2) and pubkeys (in G1) using blst directly.
	sigAgg := new(blst.P2Aggregate)
	if !sigAgg.Aggregate(signerSigsAffine, false) {
		panic("BLS signature aggregation failed")
	}
	aggSigP2 := sigAgg.ToAffine()
	if !aggSigP2.FastAggregateVerify(false, signerPubsAffine, message, dstMinPk) {
		panic("BLS aggregate verification failed")
	}

	pkAgg := new(blst.P1Aggregate)
	if !pkAgg.Aggregate(signerPubsAffine, false) {
		panic("BLS pubkey aggregation failed")
	}
	aggPubP1 := pkAgg.ToAffine()

	// Hash the message to G2 with the same DST cometbft uses to sign. The
	// hashed point is what the pairing check consumes on-chain.
	hashedP2 := blst.HashToG2(message, dstMinPk).ToAffine()
	wrongHashedP2 := blst.HashToG2(append(append([]byte{}, message...), byte(0x01)), dstMinPk).ToAffine()

	var missingAggSig string
	if signerCount > 1 {
		missingSigAgg := new(blst.P2Aggregate)
		if !missingSigAgg.Aggregate(signerSigsAffine[:signerCount-1], false) {
			panic("missing-signer BLS signature aggregation failed")
		}
		missingAggSig = "0x" + hex.EncodeToString(g2AffineToEip2537(missingSigAgg.ToAffine()))
	}

	return canonicalBls{
		Available:                        true,
		PubKeys:                          pubKeys,
		PubKeysEip2537:                   pubKeysEip2537,
		Powers:                           int64Powers,
		Leaves:                           leavesHex,
		Hash:                             "0x" + hex.EncodeToString(hash),
		Signatures:                       signatures,
		AggregateSigEip2537:              "0x" + hex.EncodeToString(g2AffineToEip2537(aggSigP2)),
		AggregatePubKeyEip2537:           "0x" + hex.EncodeToString(g1AffineToEip2537(aggPubP1)),
		HashedMessageEip2537:             "0x" + hex.EncodeToString(g2AffineToEip2537(hashedP2)),
		WrongMessageHashedEip2537:        "0x" + hex.EncodeToString(g2AffineToEip2537(wrongHashedP2)),
		MissingSignerAggregateSigEip2537: missingAggSig,
		SignerBitmap:                     bitmap,
		Message:                          "0x" + hex.EncodeToString(message),
	}
}

// g1AffineToEip2537 converts a blst-serialized G1 affine point (96 bytes:
// x_be(48) || y_be(48)) into the EIP-2537 input format (128 bytes:
// pad16 || x_be(48) || pad16 || y_be(48)).
func g1AffineToEip2537(p *blst.P1Affine) []byte {
	raw := p.Serialize()
	if len(raw) != 96 {
		panic(fmt.Sprintf("unexpected blst G1 serialize length %d", len(raw)))
	}
	out := make([]byte, 128)
	copy(out[16:64], raw[0:48])
	copy(out[80:128], raw[48:96])
	return out
}

// g2AffineToEip2537 converts a blst-serialized G2 affine point (192 bytes:
// x.c1(48) || x.c0(48) || y.c1(48) || y.c0(48), per IETF order) into the
// EIP-2537 input format (256 bytes: pad16 || x.c0 || pad16 || x.c1 ||
// pad16 || y.c0 || pad16 || y.c1).
//
// The byte ordering swap is the operative subtlety: blst serializes c1 first
// (the "imaginary" coefficient) per IETF, but EIP-2537 specifies (c0, c1)
// per FP2. Ed25519/secp256k1/EIP-2537 each pin their own byte order; this
// permutation is the only canonical bridge.
func g2AffineToEip2537(p *blst.P2Affine) []byte {
	raw := p.Serialize()
	if len(raw) != 192 {
		panic(fmt.Sprintf("unexpected blst G2 serialize length %d", len(raw)))
	}
	xC1 := raw[0:48]
	xC0 := raw[48:96]
	yC1 := raw[96:144]
	yC0 := raw[144:192]
	out := make([]byte, 256)
	copy(out[16:64], xC0)
	copy(out[80:128], xC1)
	copy(out[144:192], yC0)
	copy(out[208:256], yC1)
	return out
}
