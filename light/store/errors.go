package store

import (
	"errors"
	"fmt"
)

// ErrLightBlockNotFound is returned when a store does not have the
// requested header.
var ErrLightBlockNotFound = errors.New("light block not found")

type ErrMarshalBlock struct {
	Err error
}

func (e ErrMarshalBlock) Error() string {
	return fmt.Sprintf("marshaling LightBlock: %v", e.Err)
}

func (e ErrMarshalBlock) Unwrap() error {
	return e.Err
}

type ErrUnmarshal struct {
	Err error
}

func (e ErrUnmarshal) Error() string {
	return fmt.Sprintf("unmarshal error: %v", e.Err)
}

func (e ErrUnmarshal) Unwrap() error {
	return e.Err
}

type ErrProtoConversion struct {
	Err error
}

func (e ErrProtoConversion) Error() string {
	return fmt.Sprintf("proto conversion error: %v", e.Err)
}

func (e ErrProtoConversion) Unwrap() error {
	return e.Err
}

type ErrStore struct {
	Err error
}

func (e ErrStore) Error() string {
	return e.Err.Error()
}

func (e ErrStore) Unwrap() error {
	return e.Err
}
