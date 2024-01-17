package proxy

import "fmt"

type ErrCreateHTTPClient struct {
	Addr string
	Err  error
}

func (e ErrCreateHTTPClient) Error() string {
	return fmt.Sprintf("failed to create http client for %s: %v", e.Addr, e.Err)
}

func (e ErrCreateHTTPClient) Unwrap() error {
	return e.Err
}

type ErrStartHTTPClient struct {
	Err error
}

func (e ErrStartHTTPClient) Error() string {
	return fmt.Sprintf("can't start client: %v", e.Err)
}

func (e ErrStartHTTPClient) Unwrap() error {
	return e.Err
}
