package blocksync

import (
	"errors"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
)

var (
	// ErrNilMessage is returned when provided message is empty
	ErrNilMessage = errors.New("message cannot be nil")
)

type ErrInvalidHeight struct {
	Height int64
	Reason string
}

func (e ErrInvalidHeight) Error() string {
	return fmt.Sprintf("Invalid height %v. %s", e.Height, e.Reason)
}

type ErrInvalidBase struct {
	Base   int64
	Reason string
}

func (e ErrInvalidBase) Error() string {
	return fmt.Sprintf("Invalid base %v. %s", e.Base, e.Reason)
}

type ErrUnknownMessageType struct {
	Msg proto.Message
}

func (e ErrUnknownMessageType) Error() string {
	return fmt.Sprintf("unknown message type %T", e.Msg)
}

type ErrReactorValidation struct {
	Err error
}

func (e ErrReactorValidation) Error() string {
	return fmt.Sprintf("Reactor validation error %v", e.Err)
}
