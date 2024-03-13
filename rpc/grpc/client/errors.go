package client

import "fmt"

type ErrBlockResults struct {
	Height int64
	Err    error
}

func (e ErrBlockResults) Error() string {
	return fmt.Sprintf("error fetching BlockResults for height %d: %s", e.Height, e.Err.Error())
}

type ErrStreamSetup struct {
	Err error
}

func (e ErrStreamSetup) Error() string {
	return "error getting a stream for the latest height: " + e.Err.Error()
}

func (e ErrStreamSetup) Unwrap() error {
	return e.Err
}

type ErrStreamReceive struct {
	Err error
}

func (e ErrStreamReceive) Error() string {
	return "error receiving the latest height from a stream: " + e.Err.Error()
}

func (e ErrStreamReceive) Unwrap() error {
	return e.Err
}

type ErrDail struct {
	Addr   string
	Source error
}

func (e ErrDail) Error() string {
	return fmt.Sprintf("failed to dial: address %s: %v", e.Addr, e.Source)
}

func (e ErrDail) Unwrap() error {
	return e.Source
}
