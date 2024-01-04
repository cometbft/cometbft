package conn

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSecretConnKeySend  = errors.New("send invalid secret connection key")
	ErrInvalidSecreteConnKeyRecv = errors.New("received invalid secret connection key")
	ErrChallengeVerification     = errors.New("challenge verification failed")
)

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	source error
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet: %v", e.source)
}

func (e ErrPacketWrite) Unwrap() error {
	return e.source
}

type ErrUnexpectedPubKeyType struct {
	Expected any
	Got      any
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptConnection struct {
	source error
}

func (e ErrDecryptConnection) Error() string {
	return fmt.Sprintf("failed to decrypt SecretConnection: %v", e.source)
}

func (e ErrDecryptConnection) Unwrap() error {
	return e.source
}

type ErrPacketSize struct {
	Received int
	Max      int
}

func (e ErrPacketSize) Error() string {
	// return fmt.Sprintf("received message exceeds maximum capacity: %v < %v", e.Max, e.Received)
	return fmt.Sprintf("packet is too big (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkSize struct {
	Received int
	Max      int
}

func (e ErrChunkSize) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}
