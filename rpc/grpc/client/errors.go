package client

import "fmt"

type ErrBlockResults struct {
	Height int64
	Err    error
	latest bool
}

func (e ErrBlockResults) Error() string {
	if e.latest {
		return fmt.Sprintf("error fetching BlockResults for latest height :: %s", e.Err.Error())
	}

	return fmt.Sprintf("error fetching BlockResults for height %d:: %s", e.Height, e.Err.Error())
}

type ErrStreamSetup struct {
	Err error
}

func (e ErrStreamSetup) Error() string {
	return fmt.Sprintf("error getting a stream for the latest height: %s", e.Err.Error())
}

func (e ErrStreamSetup) Unwrap() error {
	return e.Err
}

type ErrStreamReceive struct {
	Err error
}

func (e ErrStreamReceive) Error() string {
	return fmt.Sprintf("error receiving the latest height from a stream: %s", e.Err.Error())
}

func (e ErrStreamReceive) Unwrap() error {
	return e.Err
}

type ErrDail struct {
	Addr   string
	Source error
}

func (e ErrDail) Error() string {
	return fmt.Sprintf("failed to dail: address %s: %v", e.Addr, e.Source)
}

func (e ErrDail) Unwrap() error {
	return e.Source
}
