package abcicli

import "fmt"

type ErrUnknownAbciTransport struct {
	transport string
}

func (e ErrUnknownAbciTransport) Error() string {
	return fmt.Sprintf("Unknown abci transport: %s", e.transport)
}
