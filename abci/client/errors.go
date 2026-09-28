package abcicli

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

// ErrClientStopped is returned when the ABCI client is stopped.
var ErrClientStopped = errors.New("ABCI client is stopped")

// ErrUnknownAbciTransport is returned when trying to create a client with an invalid transport option
type ErrUnknownAbciTransport struct {
	Transport string
}

func (e ErrUnknownAbciTransport) Error() string {
	return fmt.Sprintf("unknown abci transport: %s", e.Transport)
}

type ErrUnexpectedResponse struct {
	Response types.Response
	Reason   string
}

func (e ErrUnexpectedResponse) Error() string {
	return fmt.Sprintf("unexpected response %T: %s", e.Response.Value, e.Reason)
}
