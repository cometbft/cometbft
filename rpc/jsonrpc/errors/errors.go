package errors

import "fmt"

type ErrInvalidRemoteAddress struct {
	Addr string
}

func (e ErrInvalidRemoteAddress) Error() string {
	return fmt.Sprintf("invalid listening address %s (use fully formed addresses, including the tcp:// or unix:// prefix)", e.Addr)
}
