package server

import (
	"errors"
	"fmt"
)

var ErrConnectionStopped = errors.New("connection was stopped")

type ErrMarshalResponse struct {
	Source error
}

func (e ErrMarshalResponse) Error() string {
	return fmt.Sprintf("failed to marshal response: %v", e.Source)
}

func (e ErrMarshalResponse) Unwrap() error {
	return e.Source
}

type ErrListening struct {
	Addr   string
	Source error
}

func (e ErrListening) Error() string {
	return fmt.Sprintf("failed to listen on: %s :%v", e.Addr, e.Source)
}

func (e ErrListening) Unwrap() error {
	return e.Source
}
