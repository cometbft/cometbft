package p2p

import (
	"net"
	"time"

	"github.com/cometbft/cometbft/p2p/abstract"
)

type mockStream struct {
	net.Conn
}

func (s mockStream) Read(b []byte) (n int, err error) {
	return s.Conn.Read(b)
}

func (s mockStream) Write(b []byte) (n int, err error) {
	return s.Conn.Write(b)
}

func (mockStream) Close() error {
	return nil
}
func (s mockStream) SetDeadline(t time.Time) error      { return s.Conn.SetReadDeadline(t) }
func (s mockStream) SetReadDeadline(t time.Time) error  { return s.Conn.SetReadDeadline(t) }
func (s mockStream) SetWriteDeadline(t time.Time) error { return s.Conn.SetWriteDeadline(t) }

type mockConnection struct {
	net.Conn
	connectedAt time.Time
}

func newMockConnection(c net.Conn) *mockConnection {
	return &mockConnection{
		Conn:        c,
		connectedAt: time.Now(),
	}
}

func (c mockConnection) OpenStream(byte, any) (abstract.Stream, error) {
	return &mockStream{
		Conn: c.Conn,
	}, nil
}

func (c mockConnection) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c mockConnection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c mockConnection) Close(string) error         { return c.Conn.Close() }
func (c mockConnection) FlushAndClose(string) error { return c.Conn.Close() }
func (mockConnection) ErrorCh() <-chan error        { return nil }

type mockStatus struct {
	connectedFor time.Duration
}

func (s mockStatus) ConnectedFor() time.Duration { return s.connectedFor }
func (c mockConnection) ConnectionState() any {
	return &mockStatus{
		connectedFor: time.Since(c.connectedAt),
	}
}
