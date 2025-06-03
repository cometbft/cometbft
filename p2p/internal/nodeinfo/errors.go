package nodeinfo

import (
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/v2/libs/bytes"
)

var ErrNoNodeInfo = errors.New("no node info found")

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
