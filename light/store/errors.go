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

type ErrStoreDel struct {
	Err error
}

func (e ErrStoreDel) Error() string {
	return e.Err.Error()
}

func (e ErrStoreDel) Unwrap() error {
	return e.Err
}

type ErrStoreSet struct {
	Err error
}

func (e ErrStoreSet) Error() string {
	return e.Err.Error()
}

func (e ErrStoreSet) Unwrap() error {
	return e.Err
}

type ErrStoreWriteSync struct {
	Err error
}

func (e ErrStoreWriteSync) Error() string {
	return e.Err.Error()
}

func (e ErrStoreWriteSync) Unwrap() error {
	return e.Err
}

type ErrStoreIterError struct {
	Err error
}

func (e ErrStoreIterError) Error() string {
	return e.Err.Error()
}

func (e ErrStoreIterError) Unwrap() error {
	return e.Err
}

type ErrStoreIter struct {
	Err error
}

func (e ErrStoreIter) Error() string {
	return e.Err.Error()
}

func (e ErrStoreIter) Unwrap() error {
	return e.Err
}

type ErrStorePersistSize struct {
	Err error
}

func (e ErrStorePersistSize) Error() string {
	return fmt.Sprintf("failed to persist size: %v", e.Err)
}

func (e ErrStorePersistSize) Unwrap() error {
	return e.Err
}
