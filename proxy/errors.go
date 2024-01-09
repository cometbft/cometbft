package proxy

import (
	"fmt"
)

type ErrUnreachableProxy struct {
	Err error
}

func (e ErrUnreachableProxy) Error() string {
	return fmt.Sprintf("failed to connect to proxy: %v", e.Err)
}

func (e ErrUnreachableProxy) Unwrap() error {
	return e.Err
}

type ErrABCIClientCreate struct {
	ClientName string
	Err        error
}

func (e ErrABCIClientCreate) Error() string {
	return fmt.Sprintf("error creating ABCI client (%s client): %v", e.ClientName, e.Err)
}

func (e ErrABCIClientCreate) Unwrap() error {
	return e.Err
}

type ErrABCIClientStart struct {
	CliType string
	Err     error
}

func (e ErrABCIClientStart) Error() string {
	return fmt.Sprintf("error starting ABCI client (%s client): %v", e.CliType, e.Err)
}

func (e ErrABCIClientStart) Unwrap() error {
	return e.Err
}
