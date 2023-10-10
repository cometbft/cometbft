package sr25519

import (
	"errors"
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/primitives/sr25519"

	"github.com/cometbft/cometbft/crypto"
)

var _ crypto.BatchVerifier = &BatchVerifier{}

// InvalidKeyError represents an error that could occur as a result of
// using an invalid private or public key. It wraps errors that could
// arise due to failures in serialization or the use of an incorrect
// key, i.e., uninitialised or not sr25519.
type InvalidKeyError struct{ Err error }

func (e *InvalidKeyError) Error() string {
	return fmt.Sprintf("sr25519: invalid public key: %v", e.Err)
}

func (e *InvalidKeyError) Unwrap() error { return e.Err }

// InvalidSignatureError wraps an error that could occur as a result of
// generating an invalid signature.
type InvalidSignatureError struct{ Err error }

func (e *InvalidSignatureError) Error() string {
	return fmt.Sprintf("sr25519: invalid signature: %v", e.Err)
}

func (e *InvalidSignatureError) Unwrap() error { return e.Err }

// BatchVerifier implements batch verification for sr25519.
type BatchVerifier struct {
	*sr25519.BatchVerifier
}

func NewBatchVerifier() crypto.BatchVerifier {
	return &BatchVerifier{sr25519.NewBatchVerifier()}
}

func (b *BatchVerifier) Add(key crypto.PubKey, msg, signature []byte) error {
	pk, ok := key.(PubKey)
	if !ok {
		return &InvalidKeyError{Err: errors.New("sr25519: pubkey is not sr25519")}
	}

	var srpk sr25519.PublicKey
	if err := srpk.UnmarshalBinary(pk); err != nil {
		return &InvalidKeyError{Err: err}
	}

	var sig sr25519.Signature
	if err := sig.UnmarshalBinary(signature); err != nil {
		return &InvalidSignatureError{Err: err}
	}

	st := signingCtx.NewTranscriptBytes(msg)
	b.BatchVerifier.Add(&srpk, st, &sig)

	return nil
}

func (b *BatchVerifier) Verify() (bool, []bool) {
	return b.BatchVerifier.Verify(crypto.CReader())
}
