package pex

import (
	"errors"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/p2p"
)

var (
	ErrEmptyAddressBook = errors.New("address book is empty and couldn't resolve any seed nodes")
	// ErrUnsolicitedList is thrown when a peer provides a list of addresses that have not been asked for.
	ErrUnsolicitedList = errors.New("unsolicited pexAddrsMessage")
)

type ErrAddrBookNonRoutable struct {
	Addr *p2p.NetAddress
}

func (err ErrAddrBookNonRoutable) Error() string {
	return fmt.Sprintf("Cannot add non-routable address %v", err.Addr)
}

type errAddrBookOldAddressNewBucket struct {
	Addr     *p2p.NetAddress
	BucketID int
}

func (err errAddrBookOldAddressNewBucket) Error() string {
	return fmt.Sprintf("failed consistency check!"+
		" Cannot add pre-existing address %v into new bucket %v",
		err.Addr, err.BucketID)
}

type ErrAddrBookSelf struct {
	Addr *p2p.NetAddress
}

func (err ErrAddrBookSelf) Error() string {
	return fmt.Sprintf("Cannot add ourselves with address %v", err.Addr)
}

type ErrAddrBookPrivate struct {
	Addr *p2p.NetAddress
}

func (err ErrAddrBookPrivate) Error() string {
	return fmt.Sprintf("Cannot add private peer with address %v", err.Addr)
}

func (err ErrAddrBookPrivate) PrivateAddr() bool {
	return true
}

type ErrAddrBookPrivateSrc struct {
	Src *p2p.NetAddress
}

func (err ErrAddrBookPrivateSrc) Error() string {
	return fmt.Sprintf("Cannot add peer coming from private peer with address %v", err.Src)
}

func (err ErrAddrBookPrivateSrc) PrivateAddr() bool {
	return true
}

type ErrAddrBookNilAddr struct {
	Addr *p2p.NetAddress
	Src  *p2p.NetAddress
}

func (err ErrAddrBookNilAddr) Error() string {
	return fmt.Sprintf("Cannot add a nil address. Got (addr, src) = (%v, %v)", err.Addr, err.Src)
}

type ErrAddrBookInvalidAddr struct {
	Addr    *p2p.NetAddress
	AddrErr error
}

func (err ErrAddrBookInvalidAddr) Error() string {
	return fmt.Sprintf("Cannot add invalid address %v: %v", err.Addr, err.AddrErr)
}

// ErrAddressBanned is thrown when the address has been banned and therefore cannot be used.
type ErrAddressBanned struct {
	Addr *p2p.NetAddress
}

func (err ErrAddressBanned) Error() string {
	return fmt.Sprintf("Address: %v is currently banned", err.Addr)
}

// ErrReceivedPEXRequestTooSoon is thrown when a peer sends a PEX request too soon after the last one.
type ErrReceivedPEXRequestTooSoon struct {
	Peer         p2p.ID
	LastReceived time.Time
	Now          time.Time
	MinInterval  time.Duration
}

func (err ErrReceivedPEXRequestTooSoon) Error() string {
	return fmt.Sprintf("received PEX request from peer %v too soon (last received %v, now %v, min interval %v), Disconnecting peer",
		err.Peer, err.LastReceived, err.Now, err.MinInterval)
}

type errMaxAttemptsToDial struct{}

func (e errMaxAttemptsToDial) Error() string {
	return fmt.Sprintf("reached max attempts %d to dial", maxAttemptsToDial)
}

type errTooEarlyToDial struct {
	backoffDuration time.Duration
	lastDialed      time.Time
}

func (e errTooEarlyToDial) Error() string {
	return fmt.Sprintf(
		"too early to dial (backoff duration: %d, last dialed: %v, time since: %v)",
		e.backoffDuration, e.lastDialed, time.Since(e.lastDialed))
}

type errFailedToDial struct {
	totalAttempts int
	err           error
}

func (e errFailedToDial) Error() string {
	return fmt.Sprintf("failed to dial after %d attempts: %v", e.totalAttempts, e.err)
}

type errSeedNodeConfig struct {
	err any
}

func (e errSeedNodeConfig) Error() string {
	return fmt.Sprintf("failed to parse seed node config: %v", e.err)
}
