package conn

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSecretConnKeySend = errors.New("send invalid secret connection key")
	ErrInvalidSecretConnKeyRecv = errors.New("invalid receive SecretConnection Key")
	ErrChallengeVerification    = errors.New("challenge verification failed")
)

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	Source error
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet message: %v", e.Source)
}

func (e ErrPacketWrite) Unwrap() error {
	return e.Source
}

type ErrUnexpectedPubKeyType struct {
	Expected string
	Got      string
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptFrame struct {
	Source error
}

func (e ErrDecryptFrame) Error() string {
	return fmt.Sprintf("SecretConnection: failed to decrypt the frame: %v", e.Source)
}

func (e ErrDecryptFrame) Unwrap() error {
	return e.Source
}

type ErrPacketTooBig struct {
	Received int
	Max      int
}

func (e ErrPacketTooBig) Error() string {
	return fmt.Sprintf("received message exceeds available capacity (max: %d, got: %d)", e.Max, e.Received)
}

type ErrChunkTooBig struct {
	Received int
	Max      int
}

func (e ErrChunkTooBig) Error() string {
	return fmt.Sprintf("chunk too big (max: %d, got %d)", e.Max, e.Received)
}
