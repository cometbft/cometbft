package config

import (
	"errors"
	"fmt"
)

var (
	ErrMaxOpenConnectionsNegative = errors.New("max_open_connections can't be negative")
	ErrMaxSubscriptionClients     = errors.New("max_subscription_clients can't be negative")
)

// ErrInSection is returned if validate basic does not pass for any underlying config service.
type ErrInSection struct {
	Err     error
	Section string
}

func (e ErrInSection) Error() string {
	return fmt.Sprintf("error in [%s] section: %s", e.Section, e.Err.Error())
}

func (e ErrInSection) Unwrap() error {
	return e.Err
}
