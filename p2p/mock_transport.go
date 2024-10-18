package p2p

import (
	"net"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
)

type mockTransport struct {
	ln   net.Listener
	addr na.NetAddr
}

func (t *mockTransport) Listen(addr na.NetAddr) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return err
	}
	t.addr = addr
	t.ln = ln
	return nil
}

func (t *mockTransport) NetAddr() na.NetAddr {
	return t.addr
}

func (t *mockTransport) Accept() (net.Conn, *na.NetAddr, error) {
	c, err := t.ln.Accept()
	return c, nil, err
}

func (*mockTransport) Dial(addr na.NetAddr) (net.Conn, error) {
	return addr.DialTimeout(time.Second)
}

func (*mockTransport) Cleanup(net.Conn) error {
	return nil
}
