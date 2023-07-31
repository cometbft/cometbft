package server

import (
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

type ErrUnknownServerType struct {
	serverType string
}

func (e ErrUnknownServerType) Error() string {
	return fmt.Sprintf("Unknown server type %s", e.serverType)
}

type ErrConnectionNotExists struct {
	connID int
}

func (e ErrConnectionNotExists) Error() string {
	return fmt.Sprintf("Connection %d does not exist", e.connID)
}

type ErrReadingMessage struct {
	err error
}

func (e ErrReadingMessage) Error() string {
	return fmt.Sprintf("Error reading message %e", e.err)
}

type ErrWritingMessage struct {
	err error
}

func (e ErrWritingMessage) Error() string {
	return fmt.Sprintf("Error writing message %e", e.err)
}

type ErrUnknownClientResponse struct {
	req *types.Request
}

func (e ErrUnknownClientResponse) Error() string {
	return fmt.Sprintf("Unknown response from client %T", e.req)
}
