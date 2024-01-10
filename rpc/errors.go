package rpc

import "fmt"

type ErrParseQuery struct {
	Source error
}

func (e ErrParseQuery) Error() string {
	return fmt.Sprintf("failed to parse query: %v", e.Source)
}

func (e ErrParseQuery) Unwrap() error {
	return e.Source
}
