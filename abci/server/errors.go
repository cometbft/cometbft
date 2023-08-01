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

type ErrUnknownClientRequest struct {
	req *types.Request
}

func (e ErrUnknownClientRequest) Error() string {
	return fmt.Sprintf("Unknown request from client %T", e.req)
}
