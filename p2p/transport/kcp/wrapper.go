package kcp

import (
	"net"
	"time"

	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/xtaci/kcp-go"
)

// Conn wraps a KCP session to implement transport.Conn
type Conn struct {
	session *kcp.UDPSession
	errorCh chan error
	created time.Time
}

func NewConn(session *kcp.UDPSession) *Conn {
	return &Conn{
		session: session,
		errorCh: make(chan error, 1),
		created: time.Now(),
	}
}

// OpenStream implements transport.Conn
func (c *Conn) OpenStream(streamID byte, desc any) (transport.Stream, error) {
	// KCP doesn't support multiple streams, so we just return a wrapper around the session
	return &Stream{session: c.session}, nil
}

// LocalAddr implements net.Conn
func (c *Conn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

// RemoteAddr implements net.Conn
func (c *Conn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

// Close implements transport.Conn
func (c *Conn) Close(reason string) error {
	return c.session.Close()
}

// FlushAndClose implements transport.Conn
func (c *Conn) FlushAndClose(reason string) error {
	return c.Close(reason)
}

// ConnState implements transport.Conn
func (c *Conn) ConnState() transport.ConnState {
	return transport.ConnState{
		ConnectedFor: time.Since(c.created),
		StreamStates: make(map[byte]transport.StreamState),
	}
}

// ErrorCh implements transport.Conn
func (c *Conn) ErrorCh() <-chan error {
	return c.errorCh
}

// HandshakeStream implements transport.Conn
func (c *Conn) HandshakeStream() transport.HandshakeStream {
	return c.session
}

// Stream implements transport.Stream
type Stream struct {
	session *kcp.UDPSession
}

func (s *Stream) Write(b []byte) (n int, err error) {
	return s.session.Write(b)
}

func (s *Stream) Close() error {
	return s.session.Close()
}

func (s *Stream) TryWrite(b []byte) (n int, err error) {
	return s.Write(b)
}
