package p2p

import (
	"net"

	cmtsync "github.com/cometbft/cometbft/internal/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(conn net.Conn) bool
	HasIP(ip net.IP) bool
	Set(conn net.Conn, ip []net.IP)
	Remove(conn net.Conn)
	RemoveAddr(addr net.Addr)
}

type connSetItem struct {
	conn net.Conn
	ips  []net.IP
}

type connSet struct {
	cmtsync.RWMutex

	conns map[string]connSetItem
}

// NewConnSet returns a ConnSet implementation.
func NewConnSet() ConnSet {
	return &connSet{
		conns: map[string]connSetItem{},
	}
}

func (cs *connSet) Has(c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[c.RemoteAddr().String()]

	return ok
}

func (cs *connSet) HasIP(ip net.IP) bool {
	cs.RLock()
	defer cs.RUnlock()

	for _, c := range cs.conns {
		for _, known := range c.ips {
			if known.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func (cs *connSet) Remove(c net.Conn) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, c.RemoteAddr().String())
}

func (cs *connSet) RemoveAddr(addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, addr.String())
}

func (cs *connSet) Set(c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[c.RemoteAddr().String()] = connSetItem{
		conn: c,
		ips:  ips,
	}
}
