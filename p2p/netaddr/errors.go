package netaddr

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/p2p/internal/nodekey"
)

var (
	ErrEmptyHost = errors.New("host is empty")
	ErrNoIP      = errors.New("no IP address found")
	ErrInvalidIP = errors.New("invalid IP address")
)

type ErrNoID struct {
	Addr string
}

func (e ErrNoID) Error() string {
	return fmt.Sprintf("address (%s) does not contain ID", e.Addr)
}

type ErrInvalid struct {
	Err error
}

func (e ErrInvalid) Error() string {
	return "invalid address: " + e.Err.Error()
}

func (e ErrInvalid) Unwrap() error { return e.Err }

type ErrLookup struct {
	Addr string
	Err  error
}

func (e ErrLookup) Error() string {
	return fmt.Sprintf("error looking up host (%s): %v", e.Addr, e.Err)
}

func (e ErrLookup) Unwrap() error { return e.Err }

type ErrInvalidPort struct {
	Port uint32
}

func (e ErrInvalidPort) Error() string {
	return fmt.Sprintf("invalid port: %d", e.Port)
}

type ErrInvalidPeerID struct {
	ID     nodekey.ID
	Source error
}

func (e ErrInvalidPeerID) Error() string {
	return fmt.Sprintf("invalid peer ID (%v): %v", e.ID, e.Source)
}

func (e ErrInvalidPeerID) Unwrap() error {
	return e.Source
}

type ErrInvalidPeerIDLength struct {
	Got      int
	Expected int
}

func (e ErrInvalidPeerIDLength) Error() string {
	return fmt.Sprintf("invalid peer ID length, got %d, expected %d", e.Expected, e.Got)
}
