package netaddress

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/p2p/nodekey"
)

var (
	ErrEmptyHost = errors.New("host is empty")
	ErrNoIP      = errors.New("no IP address found")
	ErrInvalidIP = errors.New("invalid IP address")
)

type ErrNetAddressNoID struct {
	Addr string
}

func (e ErrNetAddressNoID) Error() string {
	return fmt.Sprintf("address (%s) does not contain ID", e.Addr)
}

type ErrNetAddressInvalid struct {
	Addr string
	Err  error
}

func (e ErrNetAddressInvalid) Error() string {
	return fmt.Sprintf("invalid address (%s): %v", e.Addr, e.Err)
}

func (e ErrNetAddressInvalid) Unwrap() error { return e.Err }

type ErrNetAddressLookup struct {
	Addr string
	Err  error
}

func (e ErrNetAddressLookup) Error() string {
	return fmt.Sprintf("error looking up host (%s): %v", e.Addr, e.Err)
}

func (e ErrNetAddressLookup) Unwrap() error { return e.Err }

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
