package abcicli

import (
	"fmt"

	"github.com/cometbft/cometbft/abci/types"
)

type ErrUnknownAbciTransport struct {
	transport string
}

func (e ErrUnknownAbciTransport) Error() string {
	return fmt.Sprintf("Unknown abci transport: %s", e.transport)
}

type ErrUnexpectedResponse struct {
	response types.Response
	comment  string
}

func (e ErrUnexpectedResponse) Error() string {
	return fmt.Sprintf("Unexpected response %T. %s", e.response.Value, e.comment)
}
