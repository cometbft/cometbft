package server

import (
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

type ErrUnknownServerType struct {
	ServerType string
}

func (e ErrUnknownServerType) Error() string {
	return fmt.Sprintf("Unknown server type %s", e.ServerType)
}

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
