package server

import (
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

// ErrUnknownServerType is returned when trying to create a server with invalid transport option.
type ErrUnknownServerType struct {
	ServerType string
}

func (e ErrUnknownServerType) Error() string {
	return fmt.Sprintf("Unknown server type %s", e.ServerType)
}

// ErrConnectionNotExists is returned when trying to access non-existent network connection
type ErrConnectionNotExists struct {
	ConnID int
}

func (e ErrConnectionNotExists) Error() string {
	return fmt.Sprintf("Connection %d does not exist", e.ConnID)
}

type ErrUnknownClientRequest struct {
	Req *types.Request
}

func (e ErrUnknownClientRequest) Error() string {
	return fmt.Sprintf("Unknown request from client %T", e.Req)
}
