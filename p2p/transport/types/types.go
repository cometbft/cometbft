package types

import (
	"net"

	na "github.com/cometbft/cometbft/p2p/netaddr"
)

// Conn extends net.Conn with additional methods
type Conn interface {
	net.Conn
	CloseConn(reason string) error
}

type Transport interface {
	NetAddr() na.NetAddr
	Accept() (Conn, *na.NetAddr, error)
	Dial(addr na.NetAddr) (Conn, error)
	Listen(addr na.NetAddr) error
	Close() error
	Protocol() Protocol
}

type Protocol string
