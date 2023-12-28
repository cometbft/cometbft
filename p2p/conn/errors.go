package conn

import (
	"errors"
	"fmt"
)

var (
	ErrSendInvalidSecreteConnKey = errors.New("send invalid secret connection key")
	ErrRecvInvalidSecreteConnKey = errors.New("received invalid secret connection key")
	ErrChallengeVerification     = errors.New("challenge verification failed")
	ErrChunkLength               = errors.New("chunk length is greater than dataMaxSize")
)

// ErrPacketWrite Packet error when writing.
type ErrPacketWrite struct {
	Err any
}

func (e ErrPacketWrite) Error() string {
	return fmt.Sprintf("failed to write packet: %v", e.Err)
}

type ErrUnexpectedPubKeyType struct {
	Expected any
	Got      any
}

func (e ErrUnexpectedPubKeyType) Error() string {
	return fmt.Sprintf("expected pubkey type %s, got %s", e.Expected, e.Got)
}

type ErrDecryptConnection struct {
	Err any
}

func (e ErrDecryptConnection) Error() string {
	return fmt.Sprintf("failed to decrypt SecretConnection: %v", e.Err)
}

type ErrExceedsCapacity struct {
	Received int
	Capacity int
}

func (e ErrExceedsCapacity) Error() string {
	return fmt.Sprintf("received message exceeds available capacity: %v < %v", e.Capacity, e.Received)
}
