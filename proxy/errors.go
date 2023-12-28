package proxy

import (
	"fmt"
)

type ErrNewProxyClient struct {
	Err error
}

func (e ErrNewProxyClient) Error() string {
	return fmt.Sprintf("failed to connect to proxy: %v", e.Err)
}

func (e ErrNewProxyClient) Unwrap() error {
	return e.Err
}

type ErrABCIClient struct {
	Action  string
	CliType string
	Err     error
}

func (e ErrABCIClient) Error() string {
	return fmt.Sprintf("error %s ABCI client (%s client): %v", e.Action, e.CliType, e.Err)
}

func (e ErrABCIClient) Unwrap() error {
	return e.Err
}
