package client

import "fmt"

type ErrBlockResults struct {
	Height int64
	Source error
}

func (e ErrBlockResults) Error() string {
	return fmt.Sprintf("error fetching BlockResults for height %d: %s", e.Height, e.Source.Error())
}

type ErrStreamSetup struct {
	Source error
}

func (e ErrStreamSetup) Error() string {
	return "error getting a stream for the latest height: " + e.Source.Error()
}

func (e ErrStreamSetup) Unwrap() error {
	return e.Source
}

type ErrStreamReceive struct {
	Source error
}

func (e ErrStreamReceive) Error() string {
	return "error receiving the latest height from a stream: " + e.Source.Error()
}

func (e ErrStreamReceive) Unwrap() error {
	return e.Source
}

type ErrDial struct {
	Addr   string
	Source error
}

func (e ErrDial) Error() string {
	return fmt.Sprintf("failed to dial: address %s: %v", e.Addr, e.Source)
}

func (e ErrDial) Unwrap() error {
	return e.Source
}
