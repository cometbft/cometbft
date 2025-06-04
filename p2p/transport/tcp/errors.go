package tcp

import (
	"fmt"
	"net"

	"github.com/cometbft/cometbft/v2/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/v2/p2p/netaddr"
)

// ErrTransportClosed is raised when the Transport has been closed.
type ErrTransportClosed struct{}

func (ErrTransportClosed) Error() string {
	return "transport has been closed"
}

// ErrFilterTimeout indicates that a filter operation timed out.
type ErrFilterTimeout struct{}

func (ErrFilterTimeout) Error() string {
	return "filter timed out"
}

// ErrRejected indicates that a Peer was rejected carrying additional
// information as to the reason.
type ErrRejected struct {
	addr          na.NetAddr
	conn          net.Conn
	err           error
	id            nodekey.ID
	isAuthFailure bool
	isDuplicate   bool
	isFiltered    bool
}

// Addr returns the network address for the rejected Peer.
func (e ErrRejected) Addr() na.NetAddr {
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

	return e.err.Error()
}

// IsAuthFailure when Peer authentication was unsuccessful.
func (e ErrRejected) IsAuthFailure() bool { return e.isAuthFailure }

// IsDuplicate when Peer ID or IP are present already.
func (e ErrRejected) IsDuplicate() bool { return e.isDuplicate }

// IsFiltered when Peer ID or IP was filtered.
func (e ErrRejected) IsFiltered() bool { return e.isFiltered }
