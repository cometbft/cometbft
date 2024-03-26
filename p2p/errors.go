package p2p

import (
	"errors"
	"fmt"
	"net"

	"github.com/cometbft/cometbft/libs/bytes"
)

var (
	ErrEmptyHost  = errors.New("host is empty")
	ErrNoIP       = errors.New("no IP address found")
	ErrNoNodeInfo = errors.New("no node info found")
	ErrInvalidIP  = errors.New("invalid IP address")
)

// ErrFilterTimeout indicates that a filter operation timed out.
type ErrFilterTimeout struct{}

func (ErrFilterTimeout) Error() string {
	return "filter timed out"
}

// ErrRejected indicates that a Peer was rejected carrying additional
// information as to the reason.
type ErrRejected struct {
	addr              NetAddress
	conn              net.Conn
	err               error
	id                ID
	isAuthFailure     bool
	isDuplicate       bool
	isFiltered        bool
	isIncompatible    bool
	isNodeInfoInvalid bool
	isSelf            bool
}

// Addr returns the NetAddress for the rejected Peer.
func (e ErrRejected) Addr() NetAddress {
	return e.addr
}

func (e ErrRejected) Error() string {
	if e.isAuthFailure {
		return fmt.Sprintf("auth failure: %s", e.err)
	}

	if e.isDuplicate {
		if e.conn != nil {
			return fmt.Sprintf(
				"duplicate CONN<%s>",
				e.conn.RemoteAddr().String(),
			)
		}
		if e.id != "" {
			return fmt.Sprintf("duplicate ID<%v>", e.id)
		}
	}

	if e.isFiltered {
		if e.conn != nil {
			return fmt.Sprintf(
				"filtered CONN<%s>: %s",
				e.conn.RemoteAddr().String(),
				e.err,
			)
		}

		if e.id != "" {
			return fmt.Sprintf("filtered ID<%v>: %s", e.id, e.err)
		}
	}

	if e.isIncompatible {
		return fmt.Sprintf("incompatible: %s", e.err)
	}

	if e.isNodeInfoInvalid {
		return fmt.Sprintf("invalid NodeInfo: %s", e.err)
	}

	if e.isSelf {
		return fmt.Sprintf("self ID<%v>", e.id)
	}

	return e.err.Error()
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e ErrRejected) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e ErrRejected) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e ErrRejected) IsFiltered() bool { return e.isFiltered }

// IsIncompatible when Peer NodeInfo is not compatible with our own.
func (e ErrRejected) IsIncompatible() bool { return e.isIncompatible }

// IsNodeInfoInvalid when the sent NodeInfo is not valid.
func (e ErrRejected) IsNodeInfoInvalid() bool { return e.isNodeInfoInvalid }

// IsSelf when Peer is our own node.
func (e ErrRejected) IsSelf() bool { return e.isSelf }

// ErrSwitchDuplicatePeerID to be raised when a peer is connecting with a known
// ID.
type ErrSwitchDuplicatePeerID struct {
	ID ID
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
	Addr *NetAddress
}

func (e ErrSwitchConnectToSelf) Error() string {
	return fmt.Sprintf("connect to self: %v", e.Addr)
}

type ErrSwitchAuthenticationFailure struct {
	Dialed *NetAddress
	Got    ID
}

func (e ErrSwitchAuthenticationFailure) Error() string {
	return fmt.Sprintf(
		"failed to authenticate peer. Dialed %v, but got peer with ID %s",
		e.Dialed,
		e.Got,
	)
}

// ErrTransportClosed is raised when the Transport has been closed.
type ErrTransportClosed struct{}

func (ErrTransportClosed) Error() string {
	return "transport has been closed"
}

// ErrPeerRemoval is raised when attempting to remove a peer results in an error.
type ErrPeerRemoval struct{}

func (ErrPeerRemoval) Error() string {
	return "peer removal failed"
}

// -------------------------------------------------------------------

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

// ErrCurrentlyDialingOrExistingAddress indicates that we're currently
// dialing this address or it belongs to an existing peer.
type ErrCurrentlyDialingOrExistingAddress struct {
	Addr string
}

func (e ErrCurrentlyDialingOrExistingAddress) Error() string {
	return fmt.Sprintf("connection with %s has been established or dialed", e.Addr)
}

type ErrInvalidPort struct {
	Port uint32
}

func (e ErrInvalidPort) Error() string {
	return fmt.Sprintf("invalid port: %d", e.Port)
}

type ErrInvalidPeerID struct {
	ID     ID
	Source error
}

func (e ErrInvalidPeerID) Error() string {
	return fmt.Sprintf("invalid peer ID (%v): %v", e.ID, e.Source)
}

func (e ErrInvalidPeerID) Unwrap() error {
	return e.Source
}

type ErrInvalidNodeVersion struct {
	Version string
}

func (e ErrInvalidNodeVersion) Error() string {
	return fmt.Sprintf("invalid version %s: version must be valid ASCII text without tabs", e.Version)
}

type ErrDuplicateChannelID struct {
	ID byte
}

func (e ErrDuplicateChannelID) Error() string {
	return fmt.Sprintf("channels contains duplicate channel id %v", e.ID)
}

type ErrChannelsTooLong struct {
	Length int
	Max    int
}

func (e ErrChannelsTooLong) Error() string {
	return fmt.Sprintf("channels is too long (max: %d, got: %d)", e.Max, e.Length)
}

type ErrInvalidMoniker struct {
	Moniker string
}

func (e ErrInvalidMoniker) Error() string {
	return fmt.Sprintf("moniker must be valid non-empty ASCII text without tabs, but got %v", e.Moniker)
}

type ErrInvalidTxIndex struct {
	TxIndex string
}

func (e ErrInvalidTxIndex) Error() string {
	return fmt.Sprintf("tx index must be either 'on', 'off', or empty string, got '%v'", e.TxIndex)
}

type ErrInvalidRPCAddress struct {
	RPCAddress string
}

func (e ErrInvalidRPCAddress) Error() string {
	return fmt.Sprintf("rpc address must be valid ASCII text without tabs, but got %v", e.RPCAddress)
}

type ErrInvalidNodeInfoType struct {
	Type     string
	Expected string
}

func (e ErrInvalidNodeInfoType) Error() string {
	return fmt.Sprintf("invalid NodeInfo type, Expected %s but got %s", e.Expected, e.Type)
}

type ErrDifferentBlockVersion struct {
	Other uint64
	Our   uint64
}

func (e ErrDifferentBlockVersion) Error() string {
	return fmt.Sprintf("peer is on a different Block version. Got %d, expected %d",
		e.Other, e.Our)
}

type ErrDifferentNetwork struct {
	Other string
	Our   string
}

func (e ErrDifferentNetwork) Error() string {
	return fmt.Sprintf("peer is on a different network. Got %s, expected %s", e.Other, e.Our)
}

type ErrNoCommonChannels struct {
	OtherChannels bytes.HexBytes
	OurChannels   bytes.HexBytes
}

func (e ErrNoCommonChannels) Error() string {
	return fmt.Sprintf("no common channels between us (%v) and peer (%v)", e.OurChannels, e.OtherChannels)
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

type ErrInvalidPeerIDLength struct {
	Got      int
	Expected int
}

func (e ErrInvalidPeerIDLength) Error() string {
	return fmt.Sprintf("invalid peer ID length, got %d, expected %d", e.Expected, e.Got)
}
