package client

import "fmt"

var ErrEventTimeout = fmt.Errorf("event timeout")

type ErrWaitThreshold struct {
	Got      int64
	Expected int64
}

func (e ErrWaitThreshold) Error() string {
	return fmt.Sprintf("waiting for %d blocks exceeded the threshold %d", e.Got, e.Expected)
}

type ErrSubscribe struct {
	Source error
}

func (e ErrSubscribe) Error() string {
	return fmt.Sprintf("failed to subscribe: %v", e.Source)
}

func (e ErrSubscribe) Unwrap() error {
	return e.Source
}
