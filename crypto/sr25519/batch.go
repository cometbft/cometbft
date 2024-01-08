package sr25519

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	"github.com/oasisprotocol/curve25519-voi/primitives/sr25519"
)

var _ crypto.BatchVerifier = &BatchVerifier{}

// ErrInvalidKey represents an error that could occur as a result of
// using an invalid private or public key. It wraps errors that could
// arise due to failures in serialization or the use of an incorrect
// key, i.e., uninitialised or not sr25519.
type ErrInvalidKey struct {
	Err error
}

func (e ErrInvalidKey) Error() string {
	return fmt.Sprintf("sr25519: invalid public key: %v", e.Err)
}

func (e ErrInvalidKey) Unwrap() error {
	return e.Err
}

// ErrInvalidSignature wraps an error that could occur as a result of
// generating an invalid signature.
type ErrInvalidSignature struct {
	Err error
}

func (e ErrInvalidSignature) Error() string {
	return fmt.Sprintf("sr25519: invalid signature: %v", e.Err)
}

func (e ErrInvalidSignature) Unwrap() error {
	return e.Err
}

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
		return ErrInvalidKey{Err: errors.New("sr25519: pubkey is not sr25519")}
	}

	var srpk sr25519.PublicKey
	if err := srpk.UnmarshalBinary(pk); err != nil {
		return ErrInvalidKey{Err: err}
	}

	var sig sr25519.Signature
	if err := sig.UnmarshalBinary(signature); err != nil {
		return ErrInvalidSignature{Err: err}
	}

	st := signingCtx.NewTranscriptBytes(msg)
	b.BatchVerifier.Add(&srpk, st, &sig)

	return nil
}

func (b *BatchVerifier) Verify() (bool, []bool) {
	return b.BatchVerifier.Verify(crypto.CReader())
}
