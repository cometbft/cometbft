package quic

import (
	"net"

	quic "github.com/quic-go/quic-go"

	"github.com/cometbft/cometbft/p2p/transport"
)

type Conn struct {
	quic.Connection
}

var _ transport.Conn = (*Conn)(nil)

func (c *Conn) OpenStream(byte, any) (transport.Stream, error) {
	panic("implement me")
}

func (c *Conn) LocalAddr() net.Addr {
	panic("implement me")
}

func (c *Conn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c *Conn) Close(string) error {
	panic("implement me")
}

func (c *Conn) FlushAndClose(string) error {
	panic("implement me")
}

func (c *Conn) ConnState() transport.ConnState {
	panic("implement me")
}

func (c *Conn) ErrorCh() <-chan error {
	panic("implement me")
}

func (c *Conn) HandshakeStream() transport.HandshakeStream {
	panic("implement me")
}
