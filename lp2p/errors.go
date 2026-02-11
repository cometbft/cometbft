package lp2p

import (
	"errors"
	"fmt"
)

// ErrorTransient is an error that is transient and can be retried.
type ErrorTransient struct {
	Err error
}

func (e *ErrorTransient) Error() string {
	return fmt.Sprintf("transient error: %s", e.Err.Error())
}

func (e *ErrorTransient) Unwrap() error {
	return e.Err
}

func TransientErrorFromAny(v any) (*ErrorTransient, bool) {
	err, ok := v.(error)
	if !ok {
		return nil, false
	}

	var te *ErrorTransient
	if !errors.As(err, &te) {
		return nil, false
	}

	return te, true
}
