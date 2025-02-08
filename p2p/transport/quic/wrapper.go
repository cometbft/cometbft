package quic

import (
	"context"
	"net"
	"time"

	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/quic-go/quic-go"
)

// Conn wraps a QUIC connection and stream to implement transport.Conn
type Conn struct {
	conn    quic.Connection
	stream  quic.Stream
	errorCh chan error
	created time.Time
}

func NewConn(conn quic.Connection, stream quic.Stream) *Conn {
	return &Conn{
		conn:    conn,
		stream:  stream,
		errorCh: make(chan error, 1),
		created: time.Now(),
	}
}

// Read implements io.Reader
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write implements io.Writer
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close implements transport.Conn
func (c *Conn) Close(reason string) error {
	if err := c.stream.Close(); err != nil {
		return err
	}
	return c.conn.CloseWithError(0, reason)
}

// LocalAddr implements net.Conn
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr implements net.Conn
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements net.Conn
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline implements net.Conn
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline implements net.Conn
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// HandshakeStream returns the underlying stream for the handshake
func (c *Conn) HandshakeStream() transport.HandshakeStream {
	return c.stream
}

// ErrorCh returns a channel that will receive errors from the connection
func (c *Conn) ErrorCh() <-chan error {
	return c.errorCh
}

// OpenStream implements transport.Conn
func (c *Conn) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	stream, err := c.conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}
	return &Stream{stream: stream}, nil
}

// ConnState implements transport.Conn
func (c *Conn) ConnState() transport.ConnState {
	return transport.ConnState{
		ConnectedFor: time.Since(c.created),
		StreamStates: make(map[byte]transport.StreamState),
	}
}

// FlushAndClose implements transport.Conn
func (c *Conn) FlushAndClose(reason string) error {
	return c.Close(reason)
}
