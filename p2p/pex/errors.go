package pex

import (
	"errors"
	"fmt"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
)

var (
	ErrEmptyAddressBook = errors.New("address book is empty and couldn't resolve any seed nodes")
	// ErrUnsolicitedList is thrown when a peer provides a list of addresses that have not been asked for.
	ErrUnsolicitedList = errors.New("unsolicited pexAddrsMessage")
)

type ErrAddrBookNonRoutable struct {
	Addr *na.NetAddr
}

func (err ErrAddrBookNonRoutable) Error() string {
	return fmt.Sprintf("Cannot add non-routable address %v", err.Addr)
}

type ErrAddrBookOldAddressNewBucket struct {
	Addr     *na.NetAddr
	BucketID int
}

func (err ErrAddrBookOldAddressNewBucket) Error() string {
	return fmt.Sprintf("failed consistency check!"+
		" Cannot add pre-existing address %v into new bucket %v",
		err.Addr, err.BucketID)
}

type ErrAddrBookSelf struct {
	Addr *na.NetAddr
}

func (err ErrAddrBookSelf) Error() string {
	return fmt.Sprintf("Cannot add ourselves with address %v", err.Addr)
}

type ErrAddrBookPrivate struct {
	Addr *na.NetAddr
}

func (err ErrAddrBookPrivate) Error() string {
	return fmt.Sprintf("Cannot add private peer with address %v", err.Addr)
}

func (ErrAddrBookPrivate) PrivateAddr() bool {
	return true
}

type ErrAddrBookPrivateSrc struct {
	Src *na.NetAddr
}

func (err ErrAddrBookPrivateSrc) Error() string {
	return fmt.Sprintf("Cannot add peer coming from private peer with address %v", err.Src)
}

func (ErrAddrBookPrivateSrc) PrivateAddr() bool {
	return true
}

type ErrAddrBookNilAddr struct {
	Addr *na.NetAddr
	Src  *na.NetAddr
}

func (err ErrAddrBookNilAddr) Error() string {
	return fmt.Sprintf("Cannot add a nil address. Got (addr, src) = (%v, %v)", err.Addr, err.Src)
}

type ErrAddrBookInvalidAddr struct {
	Addr    *na.NetAddr
	AddrErr error
}

func (err ErrAddrBookInvalidAddr) Error() string {
	return fmt.Sprintf("Cannot add invalid address %v: %v", err.Addr, err.AddrErr)
}

// ErrAddressBanned is thrown when the address has been banned and therefore cannot be used.
type ErrAddressBanned struct {
	Addr *na.NetAddr
}

func (err ErrAddressBanned) Error() string {
	return fmt.Sprintf("Address: %v is currently banned", err.Addr)
}

// ErrReceivedPEXRequestTooSoon is thrown when a peer sends a PEX request too soon after the last one.
type ErrReceivedPEXRequestTooSoon struct {
	Peer         nodekey.ID
	LastReceived time.Time
	Now          time.Time
	MinInterval  time.Duration
}

func (err ErrReceivedPEXRequestTooSoon) Error() string {
	return fmt.Sprintf("received PEX request from peer %v too soon (last received %v, now %v, min interval %v), Disconnecting peer",
		err.Peer, err.LastReceived, err.Now, err.MinInterval)
}

type ErrMaxAttemptsToDial struct {
	Max int
}

func (e ErrMaxAttemptsToDial) Error() string {
	return fmt.Sprintf("reached max attempts %d to dial", e.Max)
}

type ErrTooEarlyToDial struct {
	BackoffDuration time.Duration
	LastDialed      time.Time
}

func (e ErrTooEarlyToDial) Error() string {
	return fmt.Sprintf(
		"too early to dial (backoff duration: %d, last dialed: %v, time since: %v)",
		e.BackoffDuration, e.LastDialed, time.Since(e.LastDialed))
}

type ErrFailedToDial struct {
	TotalAttempts int
	Err           error
}

func (e ErrFailedToDial) Error() string {
	return fmt.Sprintf("failed to dial after %d attempts: %v", e.TotalAttempts, e.Err)
}

func (e ErrFailedToDial) Unwrap() error { return e.Err }

type ErrSeedNodeConfig struct {
	Err error
}

func (e ErrSeedNodeConfig) Error() string {
	return fmt.Sprintf("failed to parse seed node config: %v", e.Err)
}

func (e ErrSeedNodeConfig) Unwrap() error { return e.Err }
