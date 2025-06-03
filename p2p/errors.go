package p2p

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/v2/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
)

// ErrSwitchDuplicatePeerID to be raised when a peer is connecting with a known
// ID.
type ErrSwitchDuplicatePeerID struct {
	ID nodekey.ID
}

func (e ErrSwitchDuplicatePeerID) Error() string {
	return fmt.Sprintf("duplicate peer ID %v", e.ID)
}

// ErrSwitchDuplicatePeerIP to be raised whena a peer is connecting with a known
// IP.
type ErrSwitchDuplicatePeerIP struct {
	IP net.IP
}

func (e ErrSwitchDuplicatePeerIP) Error() string {
	return fmt.Sprintf("duplicate peer IP %v", e.IP.String())
}

// ErrSwitchConnectToSelf to be raised when trying to connect to itself.
type ErrSwitchConnectToSelf struct {
	Addr *na.NetAddr
}

func (e ErrSwitchConnectToSelf) Error() string {
	return fmt.Sprintf("connect to self: %v", e.Addr)
}

type ErrSwitchAuthenticationFailure struct {
	Dialed *na.NetAddr
	Got    nodekey.ID
}

func (e ErrSwitchAuthenticationFailure) Error() string {
	return fmt.Sprintf(
		"failed to authenticate peer. Dialed %v, but got peer with ID %s",
		e.Dialed,
		e.Got,
	)
}

// ErrPeerRemoval is raised when attempting to remove a peer results in an error.
type ErrPeerRemoval struct{}

func (ErrPeerRemoval) Error() string {
	return "peer removal failed"
}

// -------------------------------------------------------------------

// ErrCurrentlyDialingOrExistingAddress indicates that we're currently
// dialing this address or it belongs to an existing peer.
type ErrCurrentlyDialingOrExistingAddress struct {
	Addr string
}

func (e ErrCurrentlyDialingOrExistingAddress) Error() string {
	return fmt.Sprintf("connection with %s has been established or dialed", e.Addr)
}

type ErrStart struct {
	Service any
	Err     error
}

func (e ErrStart) Error() string {
	return fmt.Sprintf("failed to start %v: %v", e.Service, e.Err)
}

func (e ErrStart) Unwrap() error {
	return e.Err
}
