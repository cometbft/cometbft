package abcicli

import (
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

type ErrUnknownAbciTransport struct {
	Transport string
}

func (e ErrUnknownAbciTransport) Error() string {
	return fmt.Sprintf("Unknown abci transport: %s", e.Transport)
}

type ErrUnexpectedResponse struct {
	Response types.Response
	Comment  string
}

func (e ErrUnexpectedResponse) Error() string {
	return fmt.Sprintf("Unexpected response %T. %s", e.Response.Value, e.Comment)
}
