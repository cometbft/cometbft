package proxy

import "fmt"

type ErrCreateHttpClient struct {
	Addr string
	Err  error
}

func (e ErrCreateHttpClient) Error() string {
	return fmt.Sprintf("failed to create http client for %s: %v", e.Addr, e.Err)
}

func (e ErrCreateHttpClient) Unwrap() error {
	return e.Err
}

type ErrStartClient struct {
	Err error
}

func (e ErrStartClient) Error() string {
	return fmt.Sprintf("can't start client: %v", e.Err)
}

func (e ErrStartClient) Unwrap() error {
	return e.Err
}
