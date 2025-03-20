package store

import (
	"errors"
	"fmt"
)

var ErrNegativeHeight = errors.New("height must be greater than 0")

type ErrExceedLatestHeight struct {
	Height int64
}

func (e ErrExceedLatestHeight) Error() string {
	return fmt.Sprintf("cannot prune beyond the latest height %v", e.Height)
}

type ErrExceedBaseHeight struct {
	Height int64
	Base   int64
}

func (e ErrExceedBaseHeight) Error() string {
	return fmt.Sprintf("cannot prune to height %v, it is lower than base height %v", e.Height, e.Base)
}

type ErrMarshalCommit struct {
	Err error
}

func (e ErrMarshalCommit) Error() string {
	return fmt.Sprintf("unable to marshal commit: %v", e.Err)
}

func (e ErrMarshalCommit) Unwrap() error {
	return e.Err
}

type ErrDBOpt struct {
	Err error
}

func (e ErrDBOpt) Error() string {
	return e.Err.Error()
}

func (e ErrDBOpt) Unwrap() error {
	return e.Err
}
