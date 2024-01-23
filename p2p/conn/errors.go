package conn

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSecretConnKeySend = errors.New("send invalid secret connection key")
	ErrInvalidSecretConnKeyRecv = errors.New("received invalid secret connection key")
	ErrChallengeVerification    = errors.New("challenge verification failed")
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

type ErrDecryptFrame struct {
	source error
}

func (e ErrDecryptFrame) Error() string {
	return fmt.Sprintf("SecretConnection: failed to decrypt the frame: %v", e.source)
}

func (e ErrDecryptFrame) Unwrap() error {
	return e.source
}

type ErrPacketTooBig struct {
	Received int
	Max      int
}

func (e ErrPacketTooBig) Error() string {
	// return fmt.Sprintf("received message exceeds maximum capacity: %v < %v", e.Max, e.Received)
	return fmt.Sprintf("packet is too big (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkTooBig struct {
	Received int
	Max      int
}

func (e ErrChunkTooBig) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}
