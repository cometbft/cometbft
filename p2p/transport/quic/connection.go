package quic

import (
	"net"

	"github.com/cometbft/cometbft/p2p/transport"
)

type Connection struct{}

var _ transport.Conn = (*Connection)(nil)

func (c *Connection) OpenStream(byte, any) (transport.Stream, error) {
	panic("implement me")
}

func (c *Connection) LocalAddr() net.Addr {
	panic("implement me")
}

func (c *Connection) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c *Connection) Close(string) error {
	panic("implement me")
}

func (c *Connection) FlushAndClose(string) error {
	panic("implement me")
}

func (c *Connection) ConnState() transport.ConnState {
	panic("implement me")
}

func (c *Connection) ErrorCh() <-chan error {
	panic("implement me")
}

func (c *Connection) HandshakeStream() transport.HandshakeStream {
	panic("implement me")
}
